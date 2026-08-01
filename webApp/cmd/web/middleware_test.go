package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"webProj/internal/models"
	"webProj/internal/tokens"
	"webProj/internal/validator"
)

func Test_application_apiRequireAuth(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.apiRequireAuth(ok)

	t.Run("no user returns 401", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("want status %d; got %d", http.StatusUnauthorized, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["message"] != validator.AuthRequiredErr {
			t.Errorf("want message %q; got %v", validator.AuthRequiredErr, resp["message"])
		}
	})

	t.Run("with user passes through", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		req = app.contextSetUser(req, testRegularUser())
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
	})
}

func Test_application_apiRequireAdmin(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.apiRequireAdmin(ok)

	t.Run("no user returns 403", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/admin/test", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["message"] != validator.AdminRequiredErr {
			t.Errorf("want message %q; got %v", validator.AdminRequiredErr, resp["message"])
		}
	})

	t.Run("regular user returns 403", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/admin/test", nil)
		req = app.contextSetUser(req, testRegularUser())
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
		}
	})

	t.Run("admin passes through", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/admin/test", nil)
		req = app.contextSetUser(req, testAdminUser())
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
	})
}

func Test_application_verifyCSRF(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.verifyCSRF(ok)

	t.Run("GET requests bypass CSRF", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("POST without token returns 403", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/test", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
		}
	})

	t.Run("POST with invalid token returns 403", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/test", nil)
		req.Header.Set("X-CSRF-Token", "invalid-token")
		req.AddCookie(&http.Cookie{Name: "device_id", Value: tokens.GenerateSecureToken(tokens.DeviceTokenBytes)})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
		}
	})

	t.Run("POST with valid token passes through", func(t *testing.T) {
		deviceID := tokens.GenerateSecureToken(tokens.DeviceTokenBytes)
		csrfToken := tokens.Generate(app.config.csrfSecret, deviceID)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/test", nil)
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&http.Cookie{Name: "device_id", Value: deviceID})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("POST with mismatched device_id returns 403", func(t *testing.T) {
		deviceID := tokens.GenerateSecureToken(tokens.DeviceTokenBytes)
		csrfToken := tokens.Generate(app.config.csrfSecret, deviceID)
		otherDevice := tokens.GenerateSecureToken(tokens.DeviceTokenBytes)

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/test", nil)
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&http.Cookie{Name: "device_id", Value: otherDevice})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
		}
	})
}

func Test_application_requireAuth(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.requireAuth(ok)

	t.Run("no user redirects to login", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/admin/users", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusSeeOther {
			t.Errorf("want status %d; got %d", http.StatusSeeOther, rr.Code)
		}
		if rr.Header().Get("Location") != "/login" {
			t.Errorf("want redirect to /login; got %s", rr.Header().Get("Location"))
		}
	})

	t.Run("with user passes through", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/admin/users", nil)
		req = app.contextSetUser(req, testRegularUser())
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
	})
}

func Test_application_requireAdmin(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.requireAdmin(ok)

	t.Run("no user returns 403", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/admin/users", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
		}
	})

	t.Run("regular user returns 403", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/admin/users", nil)
		req = app.contextSetUser(req, testRegularUser())
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Errorf("want status %d; got %d", http.StatusForbidden, rr.Code)
		}
	})

	t.Run("admin passes through", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/admin/users", nil)
		req = app.contextSetUser(req, testAdminUser())
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
	})
}

func Test_application_rateLimitByIP(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a rate limiter that allows 2 requests per second with burst of 2
	rl := newIPRateLimiter(t.Context(), 2, 2)
	handler := app.rateLimitByIP(rl)(ok)

	t.Run("allows requests within limit", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		// Exhaust the burst for this IP
		for range 3 {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/test", nil)
			req.RemoteAddr = "10.0.0.1:12345"
			handler.ServeHTTP(rr, req)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusTooManyRequests {
			t.Errorf("want status %d; got %d", http.StatusTooManyRequests, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["message"] != validator.RateLimitErr {
			t.Errorf("want message %q; got %v", validator.RateLimitErr, resp["message"])
		}
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		// Even though 10.0.0.1 is rate-limited, a different IP should work
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "172.16.0.1:12345"
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
	})
}

func Test_application_apiAuthenticate(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	user := createTestUser(t, app, "auth@test.com", "password123", models.RoleUser, true)

	// Create a user as active, log in, then deactivate to get a valid token for an inactive user
	willDeactivate := createTestUser(t, app, "inactive@test.com", "password123", models.RoleUser, true)

	// Login to get valid auth tokens
	_, tok, err := app.services.User.Login(t.Context(), user.Email, "password123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	token := tok.PlainText

	_, inactiveTok, err := app.services.User.Login(t.Context(), willDeactivate.Email, "password123")
	if err != nil {
		t.Fatalf("login inactive: %v", err)
	}
	inactiveToken := inactiveTok.PlainText

	// Now deactivate the user
	willDeactivate.IsActive = false
	if err := app.db.User.Update(t.Context(), willDeactivate); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}

	var capturedUser *models.User
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = app.contextGetUser(r)
		w.WriteHeader(http.StatusOK)
	})

	handler := app.apiAuthenticate(ok)

	t.Run("no cookie passes through without user", func(t *testing.T) {
		capturedUser = nil
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
		if capturedUser != nil {
			t.Error("want no user in context when no cookie")
		}
	})

	t.Run("valid cookie sets user in context", func(t *testing.T) {
		capturedUser = nil
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
		if capturedUser == nil {
			t.Fatal("want user in context")
		}
		if capturedUser.Email != "auth@test.com" {
			t.Errorf("want email auth@test.com; got %s", capturedUser.Email)
		}
	})

	t.Run("invalid cookie clears cookie and passes through", func(t *testing.T) {
		capturedUser = nil
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: "invalid-token"})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
		if capturedUser != nil {
			t.Error("want no user for invalid token")
		}

		// Check cookie was cleared
		for _, c := range rr.Result().Cookies() {
			if c.Name == "auth_token" && c.MaxAge != -1 {
				t.Errorf("want auth_token MaxAge=-1; got %d", c.MaxAge)
			}
		}
	})

	t.Run("inactive user clears cookie and passes through", func(t *testing.T) {
		capturedUser = nil
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: inactiveToken})
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
		if capturedUser != nil {
			t.Error("want no user for inactive account")
		}
	})
}
