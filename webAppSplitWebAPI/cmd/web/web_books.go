package main

import (
	"net/http"
	"webProj/internal/validator"
)

func (app *application) AllBooks(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "all-books", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}

func (app *application) ShowBook(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "book", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}
