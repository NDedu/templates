package main

import (
	"net/http"
	"webProj/internal/validator"
)

func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}

func (app *application) RegisterPage(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "register", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}
