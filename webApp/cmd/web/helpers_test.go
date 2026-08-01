package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"webProj/internal/tokens"
	"webProj/internal/validator"
)

func TestReadJSON(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	t.Run("decodes valid JSON", func(t *testing.T) {

		var dst struct {
			Name string `json:"name"`
		}
		rr := testRawRequest(t, app.routes(), "POST", "/", `{"name":"test"}`)
		// readJSON is used inside handlers, test it directly
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"test"}`))
		r.Header.Set("Content-Type", "application/json")

		err := app.readJSON(w, r, &dst)
		if err != nil {
			t.Fatalf("want nil error; got %v", err)
		}
		if dst.Name != "test" {
			t.Errorf("want name test; got %s", dst.Name)
		}
		_ = rr
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {

		var dst struct{ Name string }
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{invalid`))
		r.Header.Set("Content-Type", "application/json")

		err := app.readJSON(w, r, &dst)
		if err == nil {
			t.Error("want error for invalid JSON; got nil")
		}
	})

	t.Run("rejects unknown fields", func(t *testing.T) {

		var dst struct {
			Name string `json:"name"`
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"test","unknown":"field"}`))
		r.Header.Set("Content-Type", "application/json")

		err := app.readJSON(w, r, &dst)
		if err == nil {
			t.Error("want error for unknown field; got nil")
		}
	})

	t.Run("rejects multiple JSON values", func(t *testing.T) {

		var dst struct{ Name string }
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"a"}{"name":"b"}`))
		r.Header.Set("Content-Type", "application/json")

		err := app.readJSON(w, r, &dst)
		if err == nil {
			t.Error("want error for multiple JSON values; got nil")
		}
		if err != nil && err.Error() != validator.InvalidJSONErr {
			t.Errorf("want %q; got %q", validator.InvalidJSONErr, err.Error())
		}
	})
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	t.Run("writes JSON with correct content type and status", func(t *testing.T) {

		rr := httptest.NewRecorder()
		app.writeJSON(rr, http.StatusOK, map[string]string{"msg": "ok"})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("want Content-Type application/json; got %s", ct)
		}
		body := rr.Body.String()
		if !strings.Contains(body, `"msg":"ok"`) {
			t.Errorf("want body to contain msg:ok; got %s", body)
		}
	})

	t.Run("writes custom status code", func(t *testing.T) {

		rr := httptest.NewRecorder()
		app.writeJSON(rr, http.StatusCreated, map[string]string{"id": "1"})

		if rr.Code != http.StatusCreated {
			t.Errorf("want status %d; got %d", http.StatusCreated, rr.Code)
		}
	})

	t.Run("applies extra headers", func(t *testing.T) {

		rr := httptest.NewRecorder()
		extra := http.Header{"X-Custom": []string{"value"}}
		app.writeJSON(rr, http.StatusOK, map[string]string{}, extra)

		if rr.Header().Get("X-Custom") != "value" {
			t.Errorf("want X-Custom header; got %s", rr.Header().Get("X-Custom"))
		}
	})
}

func TestErrorJSON(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	t.Run("returns error JSON with correct status", func(t *testing.T) {

		rr := httptest.NewRecorder()
		app.errorJSON(rr, http.StatusBadRequest, "bad input")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}

		body := rr.Body.String()
		if !strings.Contains(body, `"error":true`) {
			t.Error("want error:true in response")
		}
		if !strings.Contains(body, `"message":"bad input"`) {
			t.Errorf("want message 'bad input'; got %s", body)
		}
	})

	t.Run("works without optional error param", func(t *testing.T) {

		rr := httptest.NewRecorder()
		app.errorJSON(rr, http.StatusInternalServerError, "fail")

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("want status %d; got %d", http.StatusInternalServerError, rr.Code)
		}
	})
}

func TestFailedValidation(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	rr := httptest.NewRecorder()
	errs := map[string]string{"email": "required", "password": "too short"}
	app.failedValidation(rr, errs)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("want status %d; got %d", http.StatusUnprocessableEntity, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, `"error":true`) {
		t.Error("want error:true in response")
	}
	if !strings.Contains(body, validator.ValidationFailedErr) {
		t.Errorf("want message %q in response", validator.ValidationFailedErr)
	}
	if !strings.Contains(body, `"email":"required"`) {
		t.Error("want email error in response")
	}
	if !strings.Contains(body, `"password":"too short"`) {
		t.Error("want password error in response")
	}
}

func TestGetOrCreateDeviceID(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)

	t.Run("generates new device ID when no cookie", func(t *testing.T) {

		rr := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		id := app.getOrCreateDeviceID(rr, r)

		if id == "" {
			t.Fatal("want non-empty device ID")
		}
		if len(id) != tokens.DeviceTokenLen {
			t.Errorf("want length %d; got %d", tokens.DeviceTokenLen, len(id))
		}

		// should set a cookie
		cookies := rr.Result().Cookies()
		found := false
		for _, c := range cookies {
			if c.Name == "device_id" {
				found = true
				if c.Value != id {
					t.Errorf("want cookie value %s; got %s", id, c.Value)
				}
			}
		}
		if !found {
			t.Error("want device_id cookie to be set")
		}
	})

	t.Run("returns existing device ID from cookie", func(t *testing.T) {

		existing := tokens.GenerateSecureToken(tokens.DeviceTokenBytes)

		rr := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "device_id", Value: existing})

		id := app.getOrCreateDeviceID(rr, r)

		if id != existing {
			t.Errorf("want existing ID %s; got %s", existing, id)
		}

		// should NOT set a new cookie
		cookies := rr.Result().Cookies()
		for _, c := range cookies {
			if c.Name == "device_id" {
				t.Error("want no new device_id cookie when existing is valid")
			}
		}
	})

	t.Run("regenerates when cookie has wrong length", func(t *testing.T) {

		rr := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "device_id", Value: "too-short"})

		id := app.getOrCreateDeviceID(rr, r)

		if len(id) != tokens.DeviceTokenLen {
			t.Errorf("want new token with length %d; got %d", tokens.DeviceTokenLen, len(id))
		}
	})

	t.Run("generates unique IDs", func(t *testing.T) {

		seen := make(map[string]bool)
		for i := range 50 {

			rr := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			id := app.getOrCreateDeviceID(rr, r)
			if seen[id] {
				t.Fatalf("duplicate device ID at iteration %d", i)
			}
			seen[id] = true
		}
	})
}
