package main

import (
	"net/http"
	"os"
	"strings"
	"webProj/internal/validator"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/time/rate"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(allowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Device-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	mux.Use(middleware.Recoverer)

	globalLimiter := rate.NewLimiter(rate.Limit(500.0/60.0), 1000)
	mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if !globalLimiter.Allow() {

				app.errorJSON(w, http.StatusTooManyRequests, validator.ServerBusyErr)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	writeLimiter := newIPRateLimiter(app.ctx, rate.Limit(10.0/60.0), 10, app.wg)
	readLimiter := newIPRateLimiter(app.ctx, rate.Limit(30.0/60.0), 10, app.wg)
	// Stricter rate limit for expensive auth operations (bcrypt, emails)
	authLimiter := newIPRateLimiter(app.ctx, rate.Limit(5.0/60.0), 5, app.wg)

	mux.Route("/api", func(r chi.Router) {
		// public — login
		r.Group(func(r chi.Router) {
			r.Use(app.rateLimitByIP(authLimiter))
			r.Post("/authenticate", app.CreateAuthToken)
			r.Post("/register", app.RegisterUser)
			r.Post("/forgot-password", app.SendPasswordResetEmail)
			r.Post("/reset-password", app.ResetPassword)
			r.Post("/activate-account", app.ActivateAccount)
			r.Post("/resend-activation", app.ResendActivation)
		})

		// public — browse
		r.Group(func(r chi.Router) {
			r.Use(app.rateLimitByIP(readLimiter))
			r.Get("/book/{uuid}", app.GetBookById)
			r.Post("/all-books", app.GetAllBooks)
			r.Get("/search", app.Search)
		})

		// auth-aware — needs cookie check
		r.Group(func(r chi.Router) {
			r.Use(app.authenticate)
			r.Get("/me", app.GetCurrentUser)
			r.Post("/logout", app.Logout)
		})

		// public — buying books (CSRF protected, no auth required)
		r.Group(func(r chi.Router) {
			r.Use(app.verifyCSRF)

			r.Group(func(r chi.Router) {
				r.Use(app.rateLimitByIP(readLimiter))
			})

			r.Group(func(r chi.Router) {
				r.Use(app.rateLimitByIP(writeLimiter))
			})
		})

		// auth + CSRF protected
		r.Group(func(r chi.Router) {
			r.Use(app.authenticate)
			r.Use(app.requireAuth)
			r.Use(app.verifyCSRF)
			r.Use(app.rateLimitByIP(writeLimiter))
		})

		// admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(app.authenticate)
			r.Use(app.requireAuth)
			r.Use(app.requireAdmin)

			r.Group(func(r chi.Router) {
				r.Use(app.rateLimitByIP(readLimiter))
				r.Post("/admin/all-users", app.GetAllUsers)
				r.Get("/admin/get-user/{uuid}", app.GetUser)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.rateLimitByIP(writeLimiter))
				r.Use(app.verifyCSRF)
				r.Post("/admin/edit-user", app.EditUser)
				r.Delete("/admin/delete-user/{uuid}", app.DeleteUser)
				r.Post("/admin/add-user", app.AddUser)
				r.Post("/admin/reset-user-password", app.ResetUserPassword)
				r.Post("/admin/edit-book", app.EditBook)
				r.Delete("/admin/delete-book/{uuid}", app.DeleteBook)
				r.Post("/admin/add-book", app.AddBook)
			})
		})
	})

	return mux
}
