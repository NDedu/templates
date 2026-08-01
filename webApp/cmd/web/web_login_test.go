package main

import (
	"net/http"
	"strings"
	"testing"
)

func Test_application_LoginPage(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	mux := app.routes()

	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedBody       string
	}{
		{"login page renders", "/login", http.StatusOK, "Login"},
		{"login page has email field", "/login", http.StatusOK, `id="email"`},
		{"login page has password field", "/login", http.StatusOK, `id="password"`},
		{"login page has register link", "/login", http.StatusOK, `/register`},
		{"login page has csrf token", "/login", http.StatusOK, `csrfToken`},
		{"login page shows sign in link", "/login", http.StatusOK, `Sign in/up`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testGet(t, mux, tt.url)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("want status %d; got %d", tt.expectedStatusCode, rr.Code)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("want body to contain %q", tt.expectedBody)
			}
		})
	}
}

func Test_application_RegisterPage(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	mux := app.routes()

	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedBody       string
	}{
		{"register page renders", "/register", http.StatusOK, "Register"},
		{"register page has first name", "/register", http.StatusOK, `id="first-name"`},
		{"register page has last name", "/register", http.StatusOK, `id="last-name"`},
		{"register page has email", "/register", http.StatusOK, `id="email"`},
		{"register page has confirm email", "/register", http.StatusOK, `id="confirm-email"`},
		{"register page has password", "/register", http.StatusOK, `id="password"`},
		{"register page has confirm password", "/register", http.StatusOK, `id="confirm-password"`},
		{"register page has csrf token", "/register", http.StatusOK, `csrfToken`},
		{"register page shows sign in link", "/register", http.StatusOK, `Sign in/up`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testGet(t, mux, tt.url)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("want status %d; got %d", tt.expectedStatusCode, rr.Code)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("want body to contain %q", tt.expectedBody)
			}
		})
	}
}
