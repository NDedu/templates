package main

import (
	"fmt"
	"net/http"
	"testing"
	"webProj/internal/testutil"
	"webProj/internal/validator"
)

func Test_application_Search(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 books
	for i := range 100 {
		testutil.CreateBook(t, app.db, fmt.Sprintf("Book %d", i), fmt.Sprintf("Description %d", i), "")
	}
	testutil.CreateBook(t, app.db, "Golang Handbook", "A guide to Go programming", "")
	testutil.CreateBook(t, app.db, "Advanced Golang", "Deep dive into Go", "")

	handler := http.HandlerFunc(app.Search)

	t.Run("query too short returns empty", func(t *testing.T) {
		rr := testGet(t, handler, "/api/search?q=ab")

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp []any
		readResponseJSON(t, rr, &resp)
		if len(resp) != 0 {
			t.Errorf("want 0 results; got %d", len(resp))
		}
	})

	t.Run("empty query returns empty", func(t *testing.T) {
		rr := testGet(t, handler, "/api/search?q=")

		var resp []any
		readResponseJSON(t, rr, &resp)
		if len(resp) != 0 {
			t.Errorf("want 0 results; got %d", len(resp))
		}
	})

	t.Run("matching query returns results", func(t *testing.T) {
		rr := testGet(t, handler, "/api/search?q=Golang")

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp []map[string]any
		readResponseJSON(t, rr, &resp)
		if len(resp) != 2 {
			t.Errorf("want 2 results; got %d", len(resp))
		}
	})

	t.Run("limits to 10 results from 100+ books", func(t *testing.T) {
		rr := testGet(t, handler, "/api/search?q=Book")

		var resp []map[string]any
		readResponseJSON(t, rr, &resp)
		if len(resp) > 10 {
			t.Errorf("want at most 10 results; got %d", len(resp))
		}
	})

	t.Run("no match returns empty", func(t *testing.T) {
		rr := testGet(t, handler, "/api/search?q=zzzznotfound")

		var resp []any
		readResponseJSON(t, rr, &resp)
		if len(resp) != 0 {
			t.Errorf("want 0 results; got %d", len(resp))
		}
	})

	t.Run("searches description too", func(t *testing.T) {
		rr := testGet(t, handler, "/api/search?q=programming")

		var resp []map[string]any
		readResponseJSON(t, rr, &resp)
		if len(resp) < 1 {
			t.Error("want at least 1 result matching description")
		}
	})
}

func Test_application_GetBookById(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 books
	for i := range 100 {
		testutil.CreateBook(t, app.db, fmt.Sprintf("Book %d", i), fmt.Sprintf("Description %d", i), "")
	}
	created := testutil.CreateBook(t, app.db, "Target Book", "Target Description", "cover.jpg")

	handler := http.HandlerFunc(app.GetBookById)

	t.Run("empty uuid", func(t *testing.T) {
		rr := testJSONRequest(t, app, handler, "GET", "/api/books/", nil, testAdminUser(), map[string]string{"uuid": ""})

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("not found among 101 books", func(t *testing.T) {
		rr := testJSONRequest(t, app, handler, "GET", "/api/books/nonexistent", nil, testAdminUser(), map[string]string{"uuid": "nonexistent-uuid"})

		if rr.Code != http.StatusNotFound {
			t.Errorf("want status %d; got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("valid book among 101", func(t *testing.T) {
		rr := testJSONRequest(t, app, handler, "GET", "/api/books/"+created.UUID, nil, testAdminUser(), map[string]string{"uuid": created.UUID})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["name"] != "Target Book" {
			t.Errorf("want name 'Target Book'; got %v", resp["name"])
		}
		if resp["image"] != "cover.jpg" {
			t.Errorf("want image 'cover.jpg'; got %v", resp["image"])
		}
	})
}

func Test_application_GetAllBooks(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 books
	for i := range 100 {
		testutil.CreateBook(t, app.db, fmt.Sprintf("Book %03d", i), fmt.Sprintf("Description %d", i), "")
	}

	handler := http.HandlerFunc(app.GetAllBooks)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "POST", "/api/books/all", "{bad json")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("returns page of 10 from 100 books", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/books/all", map[string]any{
			"page_size": 10, "current_page": 1, "sort_col": "name", "sort_dir": "asc",
		})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		books := resp["books"].([]any)
		if len(books) != 10 {
			t.Errorf("want 10 books; got %d", len(books))
		}
		if resp["total_records"].(float64) != 100 {
			t.Errorf("want total_records 100; got %v", resp["total_records"])
		}
		if resp["last_page"].(float64) != 10 {
			t.Errorf("want last_page 10; got %v", resp["last_page"])
		}
	})

	t.Run("page size 25 gives 4 pages", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/books/all", map[string]any{
			"page_size": 25, "current_page": 1, "sort_col": "name", "sort_dir": "asc",
		})

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["last_page"].(float64) != 4 {
			t.Errorf("want last_page 4; got %v", resp["last_page"])
		}
	})

	t.Run("last page has correct remainder", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/books/all", map[string]any{
			"page_size": 7, "current_page": 15, "sort_col": "name", "sort_dir": "asc",
		})

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		lastPage := int(resp["last_page"].(float64))
		if lastPage != 15 {
			t.Errorf("want last_page 15; got %d", lastPage)
		}

		// 100 / 7 = 14 full pages + 2 remaining = 15 pages, last page has 2
		books := resp["books"].([]any)
		if len(books) != 2 {
			t.Errorf("want 2 books on last page; got %d", len(books))
		}
	})
}

func Test_application_EditBook(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 books
	for i := range 100 {
		testutil.CreateBook(t, app.db, fmt.Sprintf("Book %d", i), fmt.Sprintf("Description %d", i), "")
	}
	created := testutil.CreateBook(t, app.db, "Edit Me", "Original Description", "cover.jpg")

	handler := http.HandlerFunc(app.EditBook)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "PATCH", "/api/books/edit", "{bad json")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/books/edit", map[string]string{
			"uuid": created.UUID, "name": "", "description": "desc",
		})

		if rr.Code != http.StatusUnprocessableEntity {
			t.Errorf("want status %d; got %d", http.StatusUnprocessableEntity, rr.Code)
		}
	})

	t.Run("book not found among 101", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/books/edit", map[string]string{
			"uuid": "nonexistent", "name": "Updated", "description": "desc",
		})

		if rr.Code != http.StatusNotFound {
			t.Errorf("want status %d; got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("successful update among 101", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/books/edit", map[string]string{
			"uuid": created.UUID, "name": "Updated Name", "description": "Updated Desc",
		})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["message"] != "Book updated" {
			t.Errorf("want message 'Book updated'; got %v", resp["message"])
		}

		// Verify the update persisted
		updated, err := app.db.Book.GetByUUID(t.Context(), created.UUID)
		if err != nil {
			t.Fatalf("get updated book: %v", err)
		}
		if updated.Name != "Updated Name" {
			t.Errorf("want name 'Updated Name'; got %q", updated.Name)
		}
		if updated.Description != "Updated Desc" {
			t.Errorf("want description 'Updated Desc'; got %q", updated.Description)
		}
		// Image should be preserved (Update doesn't touch it)
		if updated.Image != "cover.jpg" {
			t.Errorf("want image preserved 'cover.jpg'; got %q", updated.Image)
		}
	})
}

func Test_application_DeleteBook(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 books
	for i := range 100 {
		testutil.CreateBook(t, app.db, fmt.Sprintf("Book %d", i), fmt.Sprintf("Description %d", i), "")
	}
	target := testutil.CreateBook(t, app.db, "Delete Me", "To be deleted", "")

	handler := http.HandlerFunc(app.DeleteBook)

	t.Run("empty uuid", func(t *testing.T) {
		rr := testJSONRequest(t, app, handler, "DELETE", "/api/books/", nil, testAdminUser(), map[string]string{"uuid": ""})

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("not found among 101 books", func(t *testing.T) {
		rr := testJSONRequest(t, app, handler, "DELETE", "/api/books/nonexistent", nil, testAdminUser(), map[string]string{"uuid": "nonexistent-uuid"})

		if rr.Code != http.StatusNotFound {
			t.Errorf("want status %d; got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("successful delete among 101", func(t *testing.T) {
		rr := testJSONRequest(t, app, handler, "DELETE", "/api/books/"+target.UUID, nil, testAdminUser(), map[string]string{"uuid": target.UUID})

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d; got %d", http.StatusOK, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["message"] != "Book deleted" {
			t.Errorf("want message 'Book deleted'; got %v", resp["message"])
		}

		// Verify 100 books remain
		_, _, _, total, err := app.db.Book.GetAllPaginated(t.Context(), 10, 1, "name", "asc")
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if total != 100 {
			t.Errorf("want 100 books remaining; got %d", total)
		}
	})
}

func Test_application_AddBook(t *testing.T) {
	t.Parallel()

	app := newTestApplicationWithDB(t)

	// Create 100 books
	for i := range 100 {
		testutil.CreateBook(t, app.db, fmt.Sprintf("Book %d", i), fmt.Sprintf("Description %d", i), "")
	}

	handler := http.HandlerFunc(app.AddBook)

	t.Run("invalid json", func(t *testing.T) {
		rr := testRawRequest(t, handler, "POST", "/api/books/add", "{bad json")

		if rr.Code != http.StatusBadRequest {
			t.Errorf("want status %d; got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/books/add", map[string]string{
			"name": "", "description": "desc",
		})

		if rr.Code != http.StatusUnprocessableEntity {
			t.Errorf("want status %d; got %d", http.StatusUnprocessableEntity, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["message"] != validator.ValidationFailedErr {
			t.Errorf("want message %q; got %v", validator.ValidationFailedErr, resp["message"])
		}
	})

	t.Run("successful creation among 100", func(t *testing.T) {
		rr := testPostJSON(t, handler, "/api/books/add", map[string]string{
			"name": "New Book", "description": "A brand new book",
		})

		if rr.Code != http.StatusCreated {
			t.Errorf("want status %d; got %d", http.StatusCreated, rr.Code)
		}

		var resp map[string]any
		readResponseJSON(t, rr, &resp)
		if resp["message"] != "Book created" {
			t.Errorf("want message 'Book created'; got %v", resp["message"])
		}

		// Verify total is now 101
		_, _, _, total, err := app.db.Book.GetAllPaginated(t.Context(), 10, 1, "name", "asc")
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if total != 101 {
			t.Errorf("want 101 books; got %d", total)
		}
	})
}
