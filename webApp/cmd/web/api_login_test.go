package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webProj/internal/models"
	"webProj/internal/validator"
)

func Test_application_CreateAuthToken(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 background users
	for i := range 100 {
		createTestUser(t, app, fmt.Sprintf("bulk%d@test.com", i), "password123", models.RoleUser, true)
	}

	createTestUser(t, app, "active@test.com", "password123", models.RoleUser, true)
	createTestUser(t, app, "inactive@test.com", "password123", models.RoleUser, false)

	handler := http.HandlerFunc(app.CreateAuthToken)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "POST", "/api/authenticate", "{bad json")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	var tests = []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedMsg    string
	}{
		{
			"missing email",
			map[string]string{"password": "pass"},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"missing password",
			map[string]string{"email": "test@test.com"},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"wrong email",
			map[string]string{"email": "wrong@test.com", "password": "password123"},
			http.StatusUnauthorized, validator.InvalidCredentialsErr,
		},
		{
			"wrong password",
			map[string]string{"email": "active@test.com", "password": "wrongpass"},
			http.StatusUnauthorized, validator.InvalidCredentialsErr,
		},
		{
			"inactive account",
			map[string]string{"email": "inactive@test.com", "password": "password123"},
			http.StatusForbidden, validator.AccountNotActiveResendErr,
		},
		{
			"successful login among 102 users",
			map[string]string{"email": "active@test.com", "password": "password123"},
			http.StatusOK, "Authenticated successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testPostJSON(t, handler, "/api/authenticate", tt.body)

			if rr.Code != tt.expectedStatus {
				t.Errorf("want status %d; got %d", tt.expectedStatus, rr.Code)
			}

			var resp map[string]any
			readResponseJSON(t, rr, &resp)
			msg, _ := resp["message"].(string)
			if msg != tt.expectedMsg {
				t.Errorf("want message %q; got %q", tt.expectedMsg, msg)
			}
		})
	}

	t.Run("successful login sets auth cookie", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/authenticate", map[string]string{
			"email": "active@test.com", "password": "password123",
		})

		found := false
		for _, c := range rr.Result().Cookies() {
			if c.Name == "auth_token" && c.Value != "" {
				found = true
				break
			}
		}
		if !found {
			t.Error("want auth_token cookie to be set")
		}
	})
}

func Test_application_RegisterUser(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 background users
	for i := range 100 {
		createTestUser(t, app, fmt.Sprintf("bulk%d@test.com", i), "password123", models.RoleUser, true)
	}

	createTestUser(t, app, "existing@test.com", "password123", models.RoleUser, true)

	handler := http.HandlerFunc(app.RegisterUser)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "POST", "/api/register", "not json")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	var tests = []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedMsg    string
	}{
		{
			"missing first name",
			map[string]string{
				"last_name": "User", "email": "new@test.com", "confirm_email": "new@test.com",
				"password": "password123", "confirm_password": "password123",
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"missing last name",
			map[string]string{
				"first_name": "New", "email": "new@test.com", "confirm_email": "new@test.com",
				"password": "password123", "confirm_password": "password123",
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"missing email",
			map[string]string{
				"first_name": "New", "last_name": "User", "confirm_email": "new@test.com",
				"password": "password123", "confirm_password": "password123",
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"emails do not match",
			map[string]string{
				"first_name": "New", "last_name": "User",
				"email": "new@test.com", "confirm_email": "different@test.com",
				"password": "password123", "confirm_password": "password123",
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"password too short",
			map[string]string{
				"first_name": "New", "last_name": "User",
				"email": "new@test.com", "confirm_email": "new@test.com",
				"password": "short", "confirm_password": "short",
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"passwords do not match",
			map[string]string{
				"first_name": "New", "last_name": "User",
				"email": "new@test.com", "confirm_email": "new@test.com",
				"password": "password123", "confirm_password": "different123",
			},
			http.StatusUnprocessableEntity, validator.ValidationFailedErr,
		},
		{
			"email already registered among 101",
			map[string]string{
				"first_name": "New", "last_name": "User",
				"email": "existing@test.com", "confirm_email": "existing@test.com",
				"password": "password123", "confirm_password": "password123",
			},
			http.StatusBadRequest, validator.EmailAlreadyRegisteredErr,
		},
		{
			"successful registration among 101",
			map[string]string{
				"first_name": "New", "last_name": "User",
				"email": "newuser@test.com", "confirm_email": "newuser@test.com",
				"password": "password123", "confirm_password": "password123",
			},
			http.StatusCreated, "Registration successful, your account will be activated by an administrator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testPostJSON(t, handler, "/api/register", tt.body)

			if rr.Code != tt.expectedStatus {
				t.Errorf("want status %d; got %d", tt.expectedStatus, rr.Code)
			}

			var resp map[string]any
			readResponseJSON(t, rr, &resp)
			msg, _ := resp["message"].(string)
			if msg != tt.expectedMsg {
				t.Errorf("want message %q; got %q", tt.expectedMsg, msg)
			}
		})
	}
}

func Test_application_GetCurrentUser(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	handler := http.HandlerFunc(app.GetCurrentUser)

	user := &models.User{
		ID: 1, UUID: "test-uuid", FirstName: "John", LastName: "Doe",
		Email: "john@test.com", Role: models.RoleUser, IsActive: true,
	}

	rr := testJSONRequest(t, app, handler, "GET", "/api/me", nil, user)

	if rr.Code != http.StatusOK {
		t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
	}

	var resp map[string]any
	readResponseJSON(t, rr, &resp)
	if resp["error"] != false {
		t.Error("want error=false")
	}

	userData, ok := resp["user"].(map[string]any)
	if !ok {
		t.Fatal("want user object in response")
	}
	if userData["email"] != "john@test.com" {
		t.Errorf("want email john@test.com; got %v", userData["email"])
	}
	if userData["uuid"] != "test-uuid" {
		t.Errorf("want uuid test-uuid; got %v", userData["uuid"])
	}
}

func Test_application_APILogout(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)
	handler := http.HandlerFunc(app.APILogout)

	t.Run("logout without cookie", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/logout", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["message"] != "logged out" {
			t.Errorf("want message 'logged out'; got %v", resp["message"])
		}
	})

	t.Run("logout with cookie clears it", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/logout", nil)
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: "some-token"})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		for _, c := range rr.Result().Cookies() {
			if c.Name == "auth_token" {
				if c.MaxAge != -1 {
					t.Errorf("want auth_token MaxAge=-1; got %d", c.MaxAge)
				}
				if c.Value != "" {
					t.Errorf("want auth_token value empty; got %q", c.Value)
				}
			}
		}
	})
}
