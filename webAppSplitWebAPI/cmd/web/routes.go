package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {

	mux := chi.NewRouter()
	mux.Use(app.sessionLoad)
	mux.Use(middleware.Recoverer)
	mux.Use(app.authenticate)
	mux.Use(app.verifyFormCSRF)

	mux.Get("/", app.Home)
	mux.Get("/ws", app.WsEndpoint)

	mux.Get("/all-books", app.AllBooks)
	mux.Get("/books/{uuid}", app.ShowBook)

	// auth-protected routes
	mux.Group(func(r chi.Router) {
		r.Use(app.requireAuth)
	})

	// admin-only routes
	mux.Group(func(r chi.Router) {
		r.Use(app.requireAuth)
		r.Use(app.requireAdmin)
		r.Get("/admin/all-users", app.AllUsers)
		r.Get("/admin/users/{uuid}", app.ShowUser)
		r.Get("/admin/add-user", app.AddUserPage)
	})

	//auth
	mux.Get("/login", app.LoginPage)
	mux.Get("/forgot-password", app.ForgotPassword)
	mux.Get("/reset-password", app.ShowResetPassword)
	mux.Get("/register", app.RegisterPage)
	mux.Get("/resend-activation", app.ResendActivationPage)
	mux.Get("/activate-account", app.ActivateAccount)

	fileServer := http.FileServer(http.Dir(app.staticPath))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
