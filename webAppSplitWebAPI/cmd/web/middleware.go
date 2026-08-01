package main

import (
	"net/http"
	"time"
	"webProj/internal/models"
	"webProj/internal/tokens"
	"webProj/internal/validator"
)

func (app *application) sessionLoad(next http.Handler) http.Handler {

	return app.Session.LoadAndSave(next)
}

func (app *application) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		c, err := r.Cookie("auth_token")
		if err != nil {

			next.ServeHTTP(w, r)
			return
		}

		user, err := app.DB.Token.GetUserByToken(r.Context(), c.Value)
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

func (app *application) isAdmin(r *http.Request) bool {

	user := app.contextGetUser(r)
	return user != nil && user.Role == models.RoleAdmin
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
