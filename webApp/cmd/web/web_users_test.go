package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_application_adminPages_unauthorized(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	mux := app.routes()

	var tests = []struct {
		name string
		url  string
	}{
		{"all users requires auth", "/admin/all-users"},
		{"show user requires auth", "/admin/users/some-uuid"},
		{"add user requires auth", "/admin/add-user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testGet(t, mux, tt.url)

			if rr.Code != http.StatusSeeOther {
				t.Errorf("want status %d; got %d", http.StatusSeeOther, rr.Code)
			}

			location := rr.Header().Get("Location")
			if location != "/login" {
				t.Errorf("want redirect to /login; got %s", location)
			}
		})
	}
}

func Test_application_adminPages_nonAdmin(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	regularUser := testRegularUser()

	var tests = []struct {
		name    string
		handler http.HandlerFunc
		url     string
	}{
		{"all users forbidden for non-admin", app.AllUsers, "/admin/all-users"},
		{"show user forbidden for non-admin", app.ShowUser, "/admin/users/some-uuid"},
		{"add user forbidden for non-admin", app.AddUserPage, "/admin/add-user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.url, nil)
			req = app.contextSetUser(req, regularUser)

			chain := app.Session.LoadAndSave(
				app.requireAuth(
					app.requireAdmin(
						tt.handler,
					),
				),
			)
			chain.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
			}
		})
	}
}

func Test_application_adminPages_admin(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	adminUser := testAdminUser()

	var tests = []struct {
		name         string
		handler      http.HandlerFunc
		url          string
		expectedBody string
	}{
		{"all users page renders", app.AllUsers, "/admin/all-users", "Users"},
		{"all users has table", app.AllUsers, "/admin/all-users", `id="users-table"`},
		{"all users has add user link", app.AllUsers, "/admin/all-users", `/admin/add-user`},
		{"show user page renders", app.ShowUser, "/admin/users/admin-uuid", "User"},
		{"show user has edit button", app.ShowUser, "/admin/users/admin-uuid", `id="edit-btn"`},
		{"show user has delete button", app.ShowUser, "/admin/users/admin-uuid", `id="delete-cancel-btn"`},
		{"show user has back link", app.ShowUser, "/admin/users/admin-uuid", `/admin/all-users`},
		{"add user page renders", app.AddUserPage, "/admin/add-user", "Add User"},
		{"add user has email field", app.AddUserPage, "/admin/add-user", `id="email"`},
		{"add user has password field", app.AddUserPage, "/admin/add-user", `id="password"`},
		{"add user has role select", app.AddUserPage, "/admin/add-user", `id="role"`},
		{"add user has status select", app.AddUserPage, "/admin/add-user", `id="status"`},
		{"admin sees logout", app.AllUsers, "/admin/all-users", `Logout`},
		{"admin sees admin nav", app.AllUsers, "/admin/all-users", `Admin`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testHandlerWithUser(t, app, tt.handler, adminUser, tt.url)

			if rr.Code != http.StatusOK {
				t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("want body to contain %q", tt.expectedBody)
			}
		})
	}
}
