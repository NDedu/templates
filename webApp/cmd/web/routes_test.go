package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func Test_application_routes(t *testing.T) {
	t.Parallel()

	var registered = []struct {
		route  string
		method string
	}{
		// Web routes
		{"/", "GET"},
		{"/all-books", "GET"},
		{"/books/{uuid}", "GET"},
		{"/login", "GET"},
		{"/register", "GET"},
		{"/admin/all-users", "GET"},
		{"/admin/users/{uuid}", "GET"},
		{"/admin/add-user", "GET"},
		{"/static/*", "GET"},

		// API — public auth
		{"/api/authenticate", "POST"},
		{"/api/register", "POST"},

		// API — public browse
		{"/api/book/{uuid}", "GET"},
		{"/api/all-books", "POST"},
		{"/api/search", "GET"},

		// API — auth
		{"/api/me", "GET"},
		{"/api/logout", "POST"},

		// API — admin read
		{"/api/admin/all-users", "POST"},
		{"/api/admin/get-user/{uuid}", "GET"},

		// API — admin write
		{"/api/admin/edit-user", "POST"},
		{"/api/admin/delete-user/{uuid}", "DELETE"},
		{"/api/admin/add-user", "POST"},
		{"/api/admin/reset-user-password", "POST"},
		{"/api/admin/edit-book", "POST"},
		{"/api/admin/delete-book/{uuid}", "DELETE"},
		{"/api/admin/add-book", "POST"},
	}

	var notRegistered = []struct {
		route  string
		method string
	}{
		{"/test", "GET"},
		{"/api/nonexistent", "GET"},
		{"/admin/delete-user/{uuid}", "GET"},
	}

	app := application{ctx: t.Context()}
	mux := app.routes()

	chiRoutes := mux.(chi.Routes)

	for _, route := range registered {

		if !routeExists(route.route, route.method, chiRoutes) {

			t.Errorf("route %s is not registered", route.route)
		}
	}

	for _, route := range notRegistered {

		if routeExists(route.route, route.method, chiRoutes) {

			t.Errorf("route %s should not registered", route.route)
		}
	}
}

func routeExists(testRoute, testMethod string, chiRoutes chi.Routes) bool {

	found := false

	_ = chi.Walk(chiRoutes, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {

		if strings.EqualFold(method, testMethod) && strings.EqualFold(route, testRoute) {

			found = true
		}

		return nil
	})

	return found
}
