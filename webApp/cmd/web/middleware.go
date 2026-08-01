package main

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"webProj/internal/httputil"
	"webProj/internal/models"
	"webProj/internal/tokens"
	"webProj/internal/validator"

	"golang.org/x/time/rate"
)

// Shared

func (app *application) sessionLoad(next http.Handler) http.Handler {

	return app.Session.LoadAndSave(next)
}

// Web middleware (HTML responses)

func (app *application) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		c, err := r.Cookie("auth_token")
		if err != nil {

			next.ServeHTTP(w, r)
			return
		}

		user, err := app.db.Token.GetUserByToken(r.Context(), c.Value)
		if err != nil {

			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   app.config.env == "production",
				SameSite: http.SameSiteLaxMode,
			})
			next.ServeHTTP(w, r)
			return
		}

		if !user.IsActive {
			// deactivated user — clear cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   app.config.env == "production",
				SameSite: http.SameSiteLaxMode,
			})
			next.ServeHTTP(w, r)
			return
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if app.contextGetUser(r) == nil {

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAdmin(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !app.isAdmin(r) {

			http.Error(w, validator.ForbiddenErr, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// USE: currently dormant — every web route is a GET and this returns early on GET,
// and all state changes today go through JS fetch → /api guarded by verifyCSRF (X-CSRF-Token header).
// It earns its keep the moment a web route serves a server-rendered HTML form that POSTs
// straight to that route (a no-JS / progressive-enhancement fallback): drop the hidden
// csrf_token field ({{index .StringMap "csrf_token"}}) into the form, register the POST
// handler in the web group, and this validates it — same device-bound HMAC token as
// verifyCSRF, just read from r.FormValue("csrf_token") instead of the header.
func (app *application) verifyFormCSRF(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {

			next.ServeHTTP(w, r)
			return
		}

		token := r.FormValue("csrf_token")

		deviceID := app.getOrCreateDeviceID(w, r)

		if !tokens.Verify(app.config.csrfSecret, token, deviceID, 1*time.Hour) {

			app.errorLog.Println("CSRF: invalid or expired form token")
			http.Error(w, validator.ForbiddenErr, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// API middleware (JSON responses)

func (app *application) apiAuthenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		c, err := r.Cookie("auth_token")
		if err != nil {

			next.ServeHTTP(w, r)
			return
		}

		user, err := app.db.Token.GetUserByToken(r.Context(), c.Value)
		if err != nil {
			// invalid/expired token — clear the cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   app.config.env == "production",
				SameSite: http.SameSiteLaxMode,
			})
			next.ServeHTTP(w, r)
			return
		}

		if !user.IsActive {
			// deactivated user — revoke token and clear cookie
			app.db.Token.DeleteByToken(r.Context(), c.Value)
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				Secure:   app.config.env == "production",
				SameSite: http.SameSiteLaxMode,
			})
			next.ServeHTTP(w, r)
			return
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (app *application) apiRequireAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)
		if user == nil {

			app.errorJSON(w, http.StatusUnauthorized, validator.AuthRequiredErr)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) apiRequireAdmin(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !app.isAdmin(r) {

			app.errorJSON(w, http.StatusForbidden, validator.AdminRequiredErr)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) verifyCSRF(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {

			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("X-CSRF-Token")
		deviceID := app.getOrCreateDeviceID(w, r)

		if !tokens.Verify(app.config.csrfSecret, token, deviceID, 1*time.Hour) {

			app.errorJSON(w, http.StatusForbidden, validator.InvalidOrExpiredTokenErr)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Rate limiting

func (app *application) isAdmin(r *http.Request) bool {

	user := app.contextGetUser(r)
	return user != nil && user.Role == models.RoleAdmin
}

type ipRateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen int64
}

func newIPRateLimiter(ctx context.Context, r rate.Limit, burst int) *ipRateLimiter {

	rl := &ipRateLimiter{
		rate:  r,
		burst: burst,
	}

	ticker := time.NewTicker(3 * time.Minute)

	// clean up stale entries every 3 minutes
	go func() {

		defer ticker.Stop()
		for {

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rl.limiters.Range(func(key, value any) bool {

					entry := value.(*rateLimiterEntry)
					if time.Since(time.Unix(atomic.LoadInt64(&entry.lastSeen), 0)) > 5*time.Minute {

						rl.limiters.Delete(key)
					}
					return true
				})
			}
		}
	}()

	return rl
}

func (rl *ipRateLimiter) getLimiter(ip string) *rate.Limiter {

	v, exists := rl.limiters.Load(ip)
	if !exists {

		limiter := rate.NewLimiter(rl.rate, rl.burst)
		newEntry := &rateLimiterEntry{limiter: limiter, lastSeen: time.Now().Unix()}
		v, _ = rl.limiters.LoadOrStore(ip, newEntry)
	}

	entry := v.(*rateLimiterEntry)
	atomic.StoreInt64(&entry.lastSeen, time.Now().Unix())
	return entry.limiter
}

func (app *application) rateLimitByIP(rl *ipRateLimiter) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ip := httputil.NetworkID(httputil.ClientIP(r))
			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {

				app.errorJSON(w, http.StatusTooManyRequests, validator.RateLimitErr)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
