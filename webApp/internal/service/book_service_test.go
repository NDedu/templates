package service_test

import (
	"errors"
	"testing"
	"webProj/internal/models"
	"webProj/internal/service"
	"webProj/internal/testutil"
)

func TestBookService_Create(t *testing.T) {

	t.Parallel()

	t.Run("creates book among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateBooks(t, m, 100)

		id, err := svc.Book.Create(t.Context(), &models.Book{
			Name: "New Book", Description: "New Description",
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if id < 1 {
			t.Errorf("want id >= 1; got %d", id)
		}

		book, err := m.Book.Get(t.Context(), id)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if book.Name != "New Book" {
			t.Errorf("want name 'New Book'; got %q", book.Name)
		}
		if book.Description != "New Description" {
			t.Errorf("want description 'New Description'; got %q", book.Description)
		}
	})

	t.Run("total increases to 101", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateBooks(t, m, 100)

		svc.Book.Create(t.Context(), &models.Book{
			Name: "Extra Book", Description: "Extra",
		})

		_, _, _, total, err := m.Book.GetAllPaginated(t.Context(), 10, 1, "name", "asc")
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if total != 101 {
			t.Errorf("want 101 books; got %d", total)
		}
	})
}

func TestBookService_Update(t *testing.T) {

	t.Parallel()

	t.Run("updates book among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		books := testutil.CreateBooks(t, m, 100)
		target := books[50]

		err := svc.Book.Update(t.Context(), target.UUID, "Updated Name", "Updated Desc")
		if err != nil {
			t.Fatalf("update: %v", err)
		}

		updated, err := m.Book.GetByUUID(t.Context(), target.UUID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if updated.Name != "Updated Name" {
			t.Errorf("want name 'Updated Name'; got %q", updated.Name)
		}
		if updated.Description != "Updated Desc" {
			t.Errorf("want description 'Updated Desc'; got %q", updated.Description)
		}
	})

	t.Run("returns ErrBookNotFound among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateBooks(t, m, 100)

		err := svc.Book.Update(t.Context(), "nonexistent", "Name", "Desc")
		if !errors.Is(err, service.ErrBookNotFound) {
			t.Errorf("want ErrBookNotFound; got %v", err)
		}
	})

	t.Run("preserves image field", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		book := testutil.CreateBook(t, m, "With Image", "Has image", "cover.jpg")

		err := svc.Book.Update(t.Context(), book.UUID, "New Name", "New Desc")
		if err != nil {
			t.Fatalf("update: %v", err)
		}

		updated, _ := m.Book.GetByUUID(t.Context(), book.UUID)
		if updated.Image != "cover.jpg" {
			t.Errorf("want image preserved 'cover.jpg'; got %q", updated.Image)
		}
	})

	t.Run("other books unaffected", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		books := testutil.CreateBooks(t, m, 100)
		target := books[50]
		other := books[51]

		svc.Book.Update(t.Context(), target.UUID, "Changed", "Changed")

		otherBook, _ := m.Book.GetByUUID(t.Context(), other.UUID)
		if otherBook.Name == "Changed" {
			t.Error("want other book unaffected")
		}
	})
}

func TestBookService_Delete(t *testing.T) {

	t.Parallel()

	t.Run("deletes one among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		books := testutil.CreateBooks(t, m, 100)
		target := books[50]

		err := svc.Book.Delete(t.Context(), target.UUID)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		_, err = m.Book.GetByUUID(t.Context(), target.UUID)
		if err == nil {
			t.Error("want book to be deleted")
		}
	})

	t.Run("returns ErrBookNotFound among 100", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateBooks(t, m, 100)

		err := svc.Book.Delete(t.Context(), "nonexistent")
		if !errors.Is(err, service.ErrBookNotFound) {
			t.Errorf("want ErrBookNotFound; got %v", err)
		}
	})

	t.Run("other books unaffected", func(t *testing.T) {

		t.Parallel()

		svc, m := testutil.NewTestServices(t)
		testutil.CreateBooks(t, m, 100)

		books, _, _, _, _ := m.Book.GetAllPaginated(t.Context(), 1, 1, "name", "asc")
		svc.Book.Delete(t.Context(), books[0].UUID)

		_, _, _, total, err := m.Book.GetAllPaginated(t.Context(), 10, 1, "name", "asc")
		if err != nil {
			t.Fatalf("get all: %v", err)
		}
		if total != 99 {
			t.Errorf("want 99 books remaining; got %d", total)
		}
	})
}
