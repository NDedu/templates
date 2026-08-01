package main

import (
	"net/http"
	"webProj/internal/validator"
)

func (app *application) ShowUser(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "user", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}

func (app *application) AllUsers(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "all-users", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}

func (app *application) AddUserPage(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "add-user", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}
