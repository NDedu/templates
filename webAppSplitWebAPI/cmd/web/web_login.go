package main

import (
	"fmt"
	"net/http"
	"webProj/internal/urlsigner"
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

func (app *application) ResendActivationPage(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "resend-activation", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}

func (app *application) ForgotPassword(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "forgot-password", &templateData{}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}

func (app *application) ActivateAccount(w http.ResponseWriter, r *http.Request) {

	url := r.RequestURI
	testUrl := fmt.Sprintf("%s%s", app.config.frontend, url)

	signer := urlsigner.New([]byte(app.config.secretkey))

	err := signer.IsValid(testUrl, 30)
	if err != nil {

		app.errorLog.Println(err)
		app.renderTemplate(w, r, "activate-account", &templateData{
			Error: "Activation link is invalid or has expired",
		})
		return
	}

	encryptedEmail := r.URL.Query().Get("email")

	data := make(map[string]any)
	data["email"] = encryptedEmail
	data["signed_url"] = testUrl

	if err := app.renderTemplate(w, r, "activate-account", &templateData{Data: data}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}

func (app *application) ShowResetPassword(w http.ResponseWriter, r *http.Request) {

	url := r.RequestURI
	testUrl := fmt.Sprintf("%s%s", app.config.frontend, url)

	signer := urlsigner.New([]byte(app.config.secretkey))

	err := signer.IsValid(testUrl, 30)
	if err != nil {

		app.errorLog.Println(err)
		app.renderTemplate(w, r, "reset-password", &templateData{
			Error: "Reset link is invalid or has expired",
		})
		return
	}

	encryptedEmail := r.URL.Query().Get("email")

	data := make(map[string]any)
	data["email"] = encryptedEmail
	data["signed_url"] = testUrl

	if err := app.renderTemplate(w, r, "reset-password", &templateData{Data: data}); err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
	}
}
