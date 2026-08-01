package main

import (
	"net/http"
	"webProj/internal/validator"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}
