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

type ipRateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen int64
}

func newIPRateLimiter(ctx context.Context, r rate.Limit, burst int, wg *sync.WaitGroup) *ipRateLimiter {

	rl := &ipRateLimiter{
		rate:  r,
		burst: burst,
	}

	ticker := time.NewTicker(3 * time.Minute)

	// clean up stale entries every 3 minutes
	wg.Add(1)
	go func() {

		defer wg.Done()
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

func (app *application) authenticate(next http.Handler) http.Handler {

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

func (app *application) requireAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)
		if user == nil {

			app.errorJSON(w, http.StatusUnauthorized, validator.AuthRequiredErr)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) isAdmin(r *http.Request) bool {

	user := app.contextGetUser(r)
	return user != nil && user.Role == models.RoleAdmin
}

func (app *application) requireAdmin(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !app.isAdmin(r) {

			app.errorJSON(w, http.StatusForbidden, validator.AdminRequiredErr)
			return
		}

		next.ServeHTTP(w, r)
	})
}
