package main

import (
	"net/http"
	"strings"
	"testing"
)

func Test_application_AllBooks(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	mux := app.routes()

	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedBody       string
	}{
		{"all books page renders", "/all-books", http.StatusOK, "All Books"},
		{"all books has table", "/all-books", http.StatusOK, `id="books-table"`},
		{"all books has paginator", "/all-books", http.StatusOK, `id="paginator"`},
		{"all books has csrf token", "/all-books", http.StatusOK, `csrfToken`},
		{"all books has name column", "/all-books", http.StatusOK, `data-sort="name"`},
		{"all books has description column", "/all-books", http.StatusOK, `data-sort="description"`},
		{"all books shows sign in link", "/all-books", http.StatusOK, `Sign in/up`},
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

func Test_application_AllBooks_admin(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	adminUser := testAdminUser()

	var tests = []struct {
		name         string
		expectedBody string
	}{
		{"admin sees add book button", `id="add-book-btn"`},
		{"admin sees add book modal", `id="add-book-modal"`},
		{"admin sees add name input", `id="add-name-input"`},
		{"admin sees add description input", `id="add-description-input"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testHandlerWithUser(t, app, app.AllBooks, adminUser, "/all-books")

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

func Test_application_AllBooks_nonAdmin(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	regularUser := testRegularUser()

	rr := testHandlerWithUser(t, app, app.AllBooks, regularUser, "/all-books")

	if rr.Code != http.StatusOK {
		t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if strings.Contains(body, `id="add-book-btn"`) {
		t.Error("want non-admin to NOT see add book button")
	}
	if strings.Contains(body, `id="add-book-modal"`) {
		t.Error("want non-admin to NOT see add book modal")
	}
}

func Test_application_ShowBook(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	mux := app.routes()

	var tests = []struct {
		name               string
		url                string
		expectedStatusCode int
		expectedBody       string
	}{
		{"book page renders", "/books/some-uuid", http.StatusOK, "Book"},
		{"book page has name field", "/books/some-uuid", http.StatusOK, `id="name"`},
		{"book page has description field", "/books/some-uuid", http.StatusOK, `id="description"`},
		{"book page has back link", "/books/some-uuid", http.StatusOK, `/all-books`},
		{"book page has csrf token", "/books/some-uuid", http.StatusOK, `csrfToken`},
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

func Test_application_ShowBook_admin(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	adminUser := testAdminUser()

	var tests = []struct {
		name         string
		expectedBody string
	}{
		{"admin sees edit button", `id="edit-btn"`},
		{"admin sees delete button", `id="delete-cancel-btn"`},
		{"admin sees save modal", `id="save-modal"`},
		{"admin sees delete modal", `id="delete-modal"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := testHandlerWithUser(t, app, app.ShowBook, adminUser, "/books/some-uuid")

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

func Test_application_ShowBook_nonAdmin(t *testing.T) {
	t.Parallel()

	app := newTestApplication(t)
	regularUser := testRegularUser()

	rr := testHandlerWithUser(t, app, app.ShowBook, regularUser, "/books/some-uuid")

	if rr.Code != http.StatusOK {
		t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if strings.Contains(body, `id="edit-btn"`) {
		t.Error("want non-admin to NOT see edit button")
	}
	if strings.Contains(body, `id="delete-cancel-btn"`) {
		t.Error("want non-admin to NOT see delete button")
	}
}
