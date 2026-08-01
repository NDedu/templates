package main

import (
	"net/http"
	"strings"
	"testing"
)

func Test_application_home(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	mux := app.routes()

	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedBody       string
	}{
		{"home page renders", "/", http.StatusOK, "Home"},
		{"home page has hero image", "/", http.StatusOK, `fantasy3.jpg`},
		{"home page shows sign in link", "/", http.StatusOK, `Sign in/up`},
		{"404 for unknown route", "/test", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testGet(t, mux, tt.url)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("want status %d; got %d", tt.expectedStatusCode, rr.Code)
			}

			if tt.expectedBody != "" {
				body := rr.Body.String()
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("want body to contain %q", tt.expectedBody)
				}
			}
		})
	}
}

func Test_application_home_authenticated(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	adminUser := testAdminUser()

	rr := testHandlerWithUser(t, app, app.Home, adminUser, "/")

	if rr.Code != http.StatusOK {
		t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Logout") {
		t.Error("want authenticated user to see Logout link")
	}
	if !strings.Contains(body, "Admin") {
		t.Error("want admin to see Admin nav")
	}
}

func Test_application_home_regularUser(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	regularUser := testRegularUser()

	rr := testHandlerWithUser(t, app, app.Home, regularUser, "/")

	if rr.Code != http.StatusOK {
		t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Logout") {
		t.Error("want authenticated user to see Logout link")
	}
	if strings.Contains(body, `>Admin<`) {
		t.Error("want non-admin to NOT see Admin nav")
	}
}
