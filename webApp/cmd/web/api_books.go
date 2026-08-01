package main

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"
	"webProj/internal/models"
	"webProj/internal/service"
	"webProj/internal/validator"

	"github.com/go-chi/chi/v5"
)

const maxSearchResults = 10

func (app *application) Search(w http.ResponseWriter, r *http.Request) {

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(q) < 3 {

		app.writeJSON(w, http.StatusOK, []*models.SearchResult{})
		return
	}

	var all []*models.SearchResult
	for _, s := range app.db.Searchables() {
		results, err := s.Search(r.Context(), q)
		if err != nil {

			app.errorJSON(w, http.StatusInternalServerError, validator.SearchFailedErr, err)
			return
		}
		all = append(all, results...)
	}

	if all == nil {
		all = []*models.SearchResult{}
	}

	slices.SortFunc(all, func(a, b *models.SearchResult) int {

		if a.Rank != b.Rank {

			return a.Rank - b.Rank
		}

		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	if len(all) > maxSearchResults {

		all = all[:maxSearchResults]
	}

	app.writeJSON(w, http.StatusOK, all)
}

func (app *application) GetBookById(w http.ResponseWriter, r *http.Request) {

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidIDErr)
		return
	}

	book, err := app.db.Book.GetByUUID(r.Context(), uuid)
	if err != nil {

		app.errorJSON(w, http.StatusNotFound, validator.BookNotFoundErr)
		return
	}

	app.writeJSON(w, http.StatusOK, book)
}

func (app *application) GetAllBooks(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		PageSize    int    `json:"page_size"`
		CurrentPage int    `json:"current_page"`
		SortCol     string `json:"sort_col"`
		SortDir     string `json:"sort_dir"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	books, currentPage, lastPage, totalRecords, err := app.db.Book.GetAllPaginated(r.Context(), payload.PageSize, payload.CurrentPage, payload.SortCol, payload.SortDir)
	if err != nil {

		app.errorJSON(w, http.StatusInternalServerError, validator.RetrieveBooksErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":         false,
		"books":         books,
		"current_page":  currentPage,
		"page_size":     payload.PageSize,
		"last_page":     lastPage,
		"total_records": totalRecords,
	})
}

func (app *application) EditBook(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		UUID        string `json:"uuid"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	v := validator.New()
	v.Check(payload.Name != "", "name", validator.BookNameRequiredErr)
	if !v.Valid() {

		app.failedValidation(w, v.Errors)
		return
	}

	err := app.services.Book.Update(r.Context(), payload.UUID, payload.Name, payload.Description)
	if err != nil {

		if errors.Is(err, service.ErrBookNotFound) {

			app.errorJSON(w, http.StatusNotFound, validator.BookNotFoundErr)
			return
		}

		app.errorJSON(w, http.StatusInternalServerError, validator.UpdateBookErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":   false,
		"message": "Book updated",
	})
}

func (app *application) DeleteBook(w http.ResponseWriter, r *http.Request) {

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidIDErr)
		return
	}

	err := app.services.Book.Delete(r.Context(), uuid)
	if err != nil {

		if errors.Is(err, service.ErrBookNotFound) {

			app.errorJSON(w, http.StatusNotFound, validator.BookNotFoundErr)
			return
		}

		app.errorJSON(w, http.StatusInternalServerError, validator.DeleteBookErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":   false,
		"message": "Book deleted",
	})
}

func (app *application) AddBook(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	v := validator.New()
	v.Check(payload.Name != "", "name", validator.BookNameRequiredErr)
	if !v.Valid() {

		app.failedValidation(w, v.Errors)
		return
	}

	book := &models.Book{
		Name:        payload.Name,
		Description: payload.Description,
	}

	id, err := app.services.Book.Create(r.Context(), book)
	if err != nil {

		app.errorJSON(w, http.StatusInternalServerError, validator.CreateBookErr, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]any{
		"error":   false,
		"message": "Book created",
		"id":      id,
	})
}
