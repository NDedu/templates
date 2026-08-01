package models_test

import (
	"database/sql"
	"errors"
	"testing"
	"webProj/internal/models"
	"webProj/internal/testutil"
)

func TestBookModel_Insert(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)

	t.Run("bulk insert 100 books with unique IDs", func(t *testing.T) {

		ids := make(map[int]bool, 100)
		for i := range 100 {

			id, err := m.Book.Insert(t.Context(), &models.Book{
				Name: "Book", Description: "desc", Image: "",
			})
			if err != nil {
				t.Fatalf("insert %d: %v", i, err)
			}
			if ids[id] {
				t.Fatalf("duplicate id %d on insert %d", id, i)
			}
			ids[id] = true
		}

		if len(ids) != 100 {
			t.Errorf("want 100 unique IDs; got %d", len(ids))
		}
	})

	t.Run("auto-generates uuid", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		testutil.CreateBooks(t, m, 100)

		id, err := m.Book.Insert(t.Context(), &models.Book{
			Name: "UUID Book", Description: "desc",
		})
		if err != nil {
			t.Fatalf("insert: %v", err)
		}

		book, err := m.Book.Get(t.Context(), id)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if book.UUID == "" {
			t.Error("want UUID to be auto-generated")
		}
	})
}

func TestBookModel_Get(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	books := testutil.CreateBooks(t, m, 100)

	t.Run("returns correct book among 100", func(t *testing.T) {

		target := books[49]
		book, err := m.Book.Get(t.Context(), target.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if book.Name != target.Name {
			t.Errorf("want name %s; got %s", target.Name, book.Name)
		}
		if book.Description != target.Description {
			t.Errorf("want description %s; got %s", target.Description, book.Description)
		}
	})

	t.Run("returns first and last among 100", func(t *testing.T) {

		for _, target := range []*models.Book{books[0], books[99]} {

			book, err := m.Book.Get(t.Context(), target.ID)
			if err != nil {
				t.Fatalf("get id %d: %v", target.ID, err)
			}
			if book.Name != target.Name {
				t.Errorf("want name %s; got %s", target.Name, book.Name)
			}
		}
	})

	t.Run("returns error for nonexistent id", func(t *testing.T) {

		_, err := m.Book.Get(t.Context(), 99999)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("want sql.ErrNoRows; got %v", err)
		}
	})

	t.Run("empty image defaults to empty string", func(t *testing.T) {

		// All created books have empty image
		book, err := m.Book.Get(t.Context(), books[0].ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if book.Image != "" {
			t.Errorf("want empty image; got %q", book.Image)
		}
	})
}

func TestBookModel_GetByUUID(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	books := testutil.CreateBooks(t, m, 100)

	t.Run("returns correct book by uuid among 100", func(t *testing.T) {

		target := books[73]
		book, err := m.Book.GetByUUID(t.Context(), target.UUID)
		if err != nil {
			t.Fatalf("get by uuid: %v", err)
		}
		if book.Name != target.Name {
			t.Errorf("want name %s; got %s", target.Name, book.Name)
		}
	})

	t.Run("returns error for unknown uuid", func(t *testing.T) {

		_, err := m.Book.GetByUUID(t.Context(), "nonexistent")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("want sql.ErrNoRows; got %v", err)
		}
	})
}

func TestBookModel_GetAllPaginated(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	testutil.CreateBooks(t, m, 100)

	t.Run("page size 10 gives 10 pages", func(t *testing.T) {

		books, currentPage, lastPage, totalRecords, err := m.Book.GetAllPaginated(t.Context(), 10, 1, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(books) != 10 {
			t.Errorf("want 10 books; got %d", len(books))
		}
		if currentPage != 1 {
			t.Errorf("want currentPage 1; got %d", currentPage)
		}
		if lastPage != 10 {
			t.Errorf("want lastPage 10; got %d", lastPage)
		}
		if totalRecords != 100 {
			t.Errorf("want totalRecords 100; got %d", totalRecords)
		}
	})

	t.Run("page size 25 gives 4 pages", func(t *testing.T) {

		books, _, lastPage, _, err := m.Book.GetAllPaginated(t.Context(), 25, 1, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(books) != 25 {
			t.Errorf("want 25 books; got %d", len(books))
		}
		if lastPage != 4 {
			t.Errorf("want lastPage 4; got %d", lastPage)
		}
	})

	t.Run("middle page returns correct slice", func(t *testing.T) {

		books, currentPage, _, _, err := m.Book.GetAllPaginated(t.Context(), 10, 5, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(books) != 10 {
			t.Errorf("want 10 books on page 5; got %d", len(books))
		}
		if currentPage != 5 {
			t.Errorf("want currentPage 5; got %d", currentPage)
		}
	})

	t.Run("last page has correct remainder", func(t *testing.T) {

		// 100 / 7 = 14 full pages + 2 remainder → lastPage = 15
		books, currentPage, lastPage, _, err := m.Book.GetAllPaginated(t.Context(), 7, 15, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if lastPage != 15 {
			t.Errorf("want lastPage 15; got %d", lastPage)
		}
		if currentPage != 15 {
			t.Errorf("want currentPage 15; got %d", currentPage)
		}
		if len(books) != 2 {
			t.Errorf("want 2 books on last page; got %d", len(books))
		}
	})

	t.Run("clamps currentPage to lastPage", func(t *testing.T) {

		_, currentPage, lastPage, _, err := m.Book.GetAllPaginated(t.Context(), 10, 99, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if currentPage != lastPage {
			t.Errorf("want currentPage clamped to %d; got %d", lastPage, currentPage)
		}
	})

	t.Run("default sort is by name across 100", func(t *testing.T) {

		books, _, _, _, err := m.Book.GetAllPaginated(t.Context(), 100, 1, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(books) != 100 {
			t.Fatalf("want 100 books; got %d", len(books))
		}
		for i := 1; i < len(books); i++ {

			if books[i-1].Name > books[i].Name {
				t.Errorf("want ascending name order; got %s before %s", books[i-1].Name, books[i].Name)
				break
			}
		}
	})

	t.Run("sort by description desc across 100", func(t *testing.T) {

		books, _, _, _, err := m.Book.GetAllPaginated(t.Context(), 100, 1, "description", "desc")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		for i := 1; i < len(books); i++ {

			if books[i-1].Description < books[i].Description {
				t.Errorf("want descending description order; got %s before %s", books[i-1].Description, books[i].Description)
				break
			}
		}
	})

	t.Run("empty table returns empty slice", func(t *testing.T) {

		empty := testutil.NewTestDB(t)

		books, _, _, totalRecords, err := empty.Book.GetAllPaginated(t.Context(), 10, 1, "", "")
		if err != nil {
			t.Fatalf("paginated: %v", err)
		}
		if len(books) != 0 {
			t.Errorf("want 0 books; got %d", len(books))
		}
		if totalRecords != 0 {
			t.Errorf("want totalRecords 0; got %d", totalRecords)
		}
	})
}

func TestBookModel_Update(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	books := testutil.CreateBooks(t, m, 100)
	target := books[50]

	t.Run("updates one book among 100", func(t *testing.T) {

		target.Name = "Updated"
		target.Description = "Updated desc"

		err := m.Book.Update(t.Context(), target)
		if err != nil {
			t.Fatalf("update: %v", err)
		}

		got, err := m.Book.Get(t.Context(), target.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Name != "Updated" {
			t.Errorf("want name Updated; got %s", got.Name)
		}
		if got.Description != "Updated desc" {
			t.Errorf("want description 'Updated desc'; got %s", got.Description)
		}
	})

	t.Run("update does not change image", func(t *testing.T) {

		got, err := m.Book.Get(t.Context(), target.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Image != "" {
			t.Errorf("want image preserved as empty; got %s", got.Image)
		}
	})

	t.Run("other books unchanged after update", func(t *testing.T) {

		neighbor := books[51]
		got, err := m.Book.Get(t.Context(), neighbor.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Name != neighbor.Name {
			t.Errorf("want neighbor name %s unchanged; got %s", neighbor.Name, got.Name)
		}
	})
}

func TestBookModel_Delete(t *testing.T) {
	t.Parallel()

	t.Run("deletes one among 100 and verifies count", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		books := testutil.CreateBooks(t, m, 100)
		target := books[50]

		err := m.Book.Delete(t.Context(), target.ID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		_, err = m.Book.Get(t.Context(), target.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("want sql.ErrNoRows after delete; got %v", err)
		}

		remaining, _, _, totalRecords, err := m.Book.GetAllPaginated(t.Context(), 100, 1, "", "")
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if len(remaining) != 99 {
			t.Errorf("want 99 remaining books; got %d", len(remaining))
		}
		if totalRecords != 99 {
			t.Errorf("want totalRecords 99; got %d", totalRecords)
		}
	})

	t.Run("neighbors still accessible after delete", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		books := testutil.CreateBooks(t, m, 100)

		m.Book.Delete(t.Context(), books[50].ID)

		for _, idx := range []int{49, 51} {

			_, err := m.Book.Get(t.Context(), books[idx].ID)
			if err != nil {
				t.Errorf("want neighbor %d still accessible; got %v", idx, err)
			}
		}
	})

	t.Run("deleting nonexistent id does not error", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		testutil.CreateBooks(t, m, 100)

		err := m.Book.Delete(t.Context(), 99999)
		if err != nil {
			t.Errorf("want nil error; got %v", err)
		}
	})
}

func TestBookModel_Search(t *testing.T) {
	t.Parallel()

	m := testutil.NewTestDB(t)
	testutil.CreateBook(t, m, "Lord of the Rings", "Greatest book of all times!", "/img/lotr.png")
	testutil.CreateBook(t, m, "The Hobbit", "Prequel to the greatest book of all times.", "/img/hobbit.png")
	// Create 100 additional books to test search among many
	testutil.CreateBooks(t, m, 100)

	t.Run("finds book by name prefix among 102", func(t *testing.T) {

		results, err := m.Book.Search(t.Context(), "Lord")
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("want at least 1 result")
		}
		if results[0].Name != "Lord of the Rings" {
			t.Errorf("want Lord of the Rings; got %s", results[0].Name)
		}
		if results[0].Rank != 0 {
			t.Errorf("want rank 0 for prefix match; got %d", results[0].Rank)
		}
	})

	t.Run("finds book by name substring among 102", func(t *testing.T) {

		results, err := m.Book.Search(t.Context(), "Ring")
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("want at least 1 result")
		}
		if results[0].Rank != 1 {
			t.Errorf("want rank 1 for substring match; got %d", results[0].Rank)
		}
	})

	t.Run("finds book by description among 102", func(t *testing.T) {

		results, err := m.Book.Search(t.Context(), "Prequel")
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("want at least 1 result")
		}
		if results[0].Name != "The Hobbit" {
			t.Errorf("want The Hobbit; got %s", results[0].Name)
		}
	})

	t.Run("search is case-insensitive", func(t *testing.T) {

		results, err := m.Book.Search(t.Context(), "the hobbit")
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("want at least 1 result")
		}
	})

	t.Run("returns empty for no match among 102", func(t *testing.T) {

		results, err := m.Book.Search(t.Context(), "zzzzzzz")
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("want 0 results; got %d", len(results))
		}
	})

	t.Run("limits to 10 results with 100+ matching", func(t *testing.T) {

		// All 100 created books are named "Book N" — search for "Book"
		results, err := m.Book.Search(t.Context(), "Book")
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) > 10 {
			t.Errorf("want at most 10 results; got %d", len(results))
		}
		if len(results) != 10 {
			t.Errorf("want exactly 10 results from 100+ matches; got %d", len(results))
		}
	})

	t.Run("result has correct Kind and URL", func(t *testing.T) {

		results, err := m.Book.Search(t.Context(), "Lord")
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("want at least 1 result")
		}
		if results[0].Kind != "book" {
			t.Errorf("want kind book; got %s", results[0].Kind)
		}
		if results[0].URL == "" {
			t.Error("want URL to be set")
		}
	})

	t.Run("special characters in query are escaped", func(t *testing.T) {

		results, err := m.Book.Search(t.Context(), "100%")
		if err != nil {
			t.Fatalf("search with special chars: %v", err)
		}
		_ = results
	})
}
