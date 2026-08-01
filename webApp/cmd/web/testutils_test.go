package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"webProj/internal/models"
	"webProj/internal/service"
	"webProj/internal/testutil"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

// newTestApplication creates a minimal application for testing web page handlers
func newTestApplication(t *testing.T) *application {

	t.Helper()

	return &application{
		ctx:           t.Context(),
		config:        config{csrfSecret: "test-secret"},
		infoLog:       log.New(io.Discard, "", 0),
		errorLog:      log.New(io.Discard, "", 0),
		templateCache: make(map[string]*template.Template),
		Session:       scs.New(),
		staticPath:    t.TempDir(),
	}
}

// newTestApplicationWithDB creates an application with in-memory SQLite database
func newTestApplicationWithDB(t *testing.T) *application {

	t.Helper()

	app := newTestApplication(t)
	app.db = testutil.NewTestDB(t)
	app.services = service.NewServices(app.db)
	return app
}

// createTestUser inserts a user with the given credentials and returns the full record (with UUID)
func createTestUser(t *testing.T, app *application, email, password, role string, isActive bool) *models.User {

	t.Helper()

	return testutil.CreateUser(t, app.db, email, password, role, isActive)
}

func testAdminUser() *models.User {

	return &models.User{
		ID:       1,
		UUID:     "admin-uuid",
		Email:    "admin@test.com",
		Role:     models.RoleAdmin,
		IsActive: true,
	}
}

func testRegularUser() *models.User {

	return &models.User{
		ID:       2,
		UUID:     "user-uuid",
		Email:    "user@test.com",
		Role:     models.RoleUser,
		IsActive: true,
	}
}

// testGet sends a GET request through the handler and returns the recorder
func testGet(t *testing.T, handler http.Handler, url string) *httptest.ResponseRecorder {

	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", url, nil)
	handler.ServeHTTP(rr, req)
	return rr
}

// testPostJSON sends a POST request with a JSON body
func testPostJSON(t *testing.T, handler http.Handler, url string, body any) *httptest.ResponseRecorder {

	t.Helper()

	jsonBody, err := json.Marshal(body)
	if err != nil {

		t.Fatalf("failed to marshal request body: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", url, bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rr, req)
	return rr
}

// testRawRequest sends a request with a raw (non-marshalled) body string, used for testing invalid JSON input
func testRawRequest(t *testing.T, handler http.Handler, method, url, rawBody string) *httptest.ResponseRecorder {

	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, strings.NewReader(rawBody))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(rr, req)
	return rr
}

// testHandlerWithUser calls a handler wrapped in session middleware with optional user context
// used for web page handler tests that need session + authenticated user
func testHandlerWithUser(t *testing.T, app *application, handler http.HandlerFunc, user *models.User, url string) *httptest.ResponseRecorder {

	t.Helper()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", url, nil)
	if user != nil {

		req = app.contextSetUser(req, user)
	}
	app.Session.LoadAndSave(handler).ServeHTTP(rr, req)
	return rr
}

// testJSONRequest sends a JSON API request with optional user context and chi URL params
// used for API handler tests that need auth context + JSON body + URL params
func testJSONRequest(t *testing.T, app *application, handler http.Handler, method, url string, body any, user *models.User, params ...map[string]string) *httptest.ResponseRecorder {

	t.Helper()

	var bodyReader io.Reader
	if body != nil {

		jsonBody, err := json.Marshal(body)
		if err != nil {

			t.Fatalf("failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, bodyReader)
	if body != nil {

		req.Header.Set("Content-Type", "application/json")
	}
	if user != nil {

		req = app.contextSetUser(req, user)
	}
	if len(params) > 0 && params[0] != nil {

		req = withChiURLParams(req, params[0])
	}
	handler.ServeHTTP(rr, req)
	return rr
}

// withChiURLParams sets chi URL params on a request
func withChiURLParams(r *http.Request, params map[string]string) *http.Request {

	rctx := chi.NewRouteContext()
	for k, v := range params {

		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// createMultipartRequest builds a multipart/form-data POST request with a single file field
func createMultipartRequest(t *testing.T, url, fieldName, filename string, fileContent []byte, contentType string) *http.Request {

	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	h := make(map[string][]string)
	h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename)}
	h["Content-Type"] = []string{contentType}
	part, err := writer.CreatePart(h)
	if err != nil {

		t.Fatal(err)
	}

	if _, err := io.Copy(part, bytes.NewReader(fileContent)); err != nil {

		t.Fatal(err)
	}
	writer.Close()

	req := httptest.NewRequest("POST", url, &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// readResponseJSON decodes the recorder's body into dst
func readResponseJSON(t *testing.T, rr *httptest.ResponseRecorder, dst any) {

	t.Helper()

	if err := json.NewDecoder(rr.Body).Decode(dst); err != nil {

		t.Fatalf("failed to decode response JSON: %v", err)
	}
}
