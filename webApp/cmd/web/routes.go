package main

import (
	"net/http"
	"webProj/internal/validator"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
)

func (app *application) routes() http.Handler {

	mux := chi.NewRouter()

	// FIX: add a security-headers middleware here before production — CSP, HSTS, X-Frame-Options, X-Content-Type-Options
	// none are set today; bootstrap is served locally now so a strict CSP is feasible
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

	// FIX: these limiters live in memory per process — behind more than one instance the limit is per instance, not global
	// use a shared store (redis) if you scale past a single instance
	// rate.Limit is per-second, so rate.Limit(N/60.0) allows N requests per minute;
	// the second argument is the burst (how many are allowed to arrive at once).
	writeLimiter := newIPRateLimiter(app.ctx, rate.Limit(10.0/60.0), 10)
	readLimiter := newIPRateLimiter(app.ctx, rate.Limit(30.0/60.0), 10)
	// Stricter rate limit for expensive auth operations (bcrypt)
	authLimiter := newIPRateLimiter(app.ctx, rate.Limit(5.0/60.0), 5)

	// API routes (JSON)
	mux.Route("/api", func(r chi.Router) {
		// public — login
		r.Group(func(r chi.Router) {
			r.Use(app.rateLimitByIP(authLimiter))
			r.Post("/authenticate", app.CreateAuthToken)
			r.Post("/register", app.RegisterUser)
		})

		// Public browse
		r.Group(func(r chi.Router) {
			r.Use(app.rateLimitByIP(readLimiter))
			r.Get("/book/{uuid}", app.GetBookById)
			r.Post("/all-books", app.GetAllBooks)
			r.Get("/search", app.Search)
		})

		// Auth — needs cookie check
		r.Group(func(r chi.Router) {
			r.Use(app.apiAuthenticate)
			r.Get("/me", app.GetCurrentUser)
			r.Post("/logout", app.APILogout)
		})

		// USE: empty for now — a slot for public, state-changing endpoints that need CSRF
		// but not login (e.g. a contact form or signup that POSTs). verifyCSRF covers the whole
		// group; register each route in the read- or write-limited sub-group so it also gets a
		// rate limit. Nothing here runs until a route is added.
		r.Group(func(r chi.Router) {
			r.Use(app.verifyCSRF)

			r.Group(func(r chi.Router) {
				r.Use(app.rateLimitByIP(readLimiter))
			})

			r.Group(func(r chi.Router) {
				r.Use(app.rateLimitByIP(writeLimiter))
			})
		})

		// USE: empty for now — a slot for any logged-in user (not just admin) making writes
		// (e.g. editing their own profile or password). Every write route today is admin-only
		// in the group below, so this stays empty until a self-service route exists.
		r.Group(func(r chi.Router) {
			r.Use(app.apiAuthenticate)
			r.Use(app.apiRequireAuth)
			r.Use(app.verifyCSRF)
			r.Use(app.rateLimitByIP(writeLimiter))
		})

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(app.apiAuthenticate)
			r.Use(app.apiRequireAuth)
			r.Use(app.apiRequireAdmin)

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
				r.Post("/admin/upload-profile-image/{uuid}", app.UploadProfileImage)
				r.Post("/admin/edit-book", app.EditBook)
				r.Delete("/admin/delete-book/{uuid}", app.DeleteBook)
				r.Post("/admin/add-book", app.AddBook)
			})
		})
	})

	// Web routes (HTML pages)
	mux.Group(func(r chi.Router) {
		r.Use(app.sessionLoad)
		r.Use(app.authenticate)
		// dormant until a web route accepts POSTs — see the USE: note on verifyFormCSRF
		r.Use(app.verifyFormCSRF)

		r.Get("/", app.Home)

		r.Get("/all-books", app.AllBooks)
		r.Get("/books/{uuid}", app.ShowBook)

		// Auth-protected routes
		r.Group(func(r chi.Router) {
			r.Use(app.requireAuth)
		})

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(app.requireAuth)
			r.Use(app.requireAdmin)
			r.Get("/admin/all-users", app.AllUsers)
			r.Get("/admin/users/{uuid}", app.ShowUser)
			r.Get("/admin/add-user", app.AddUserPage)
		})

		// Auth pages
		r.Get("/login", app.LoginPage)
		r.Get("/register", app.RegisterPage)
	})

	fileServer := http.FileServer(http.Dir(app.staticPath))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
