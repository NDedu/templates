package main

import (
	"errors"
	"net/http"
	"strings"
	"webProj/internal/models"
	"webProj/internal/service"
	"webProj/internal/validator"
)

func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {

	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &userInput)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	v := validator.New()
	v.Check(userInput.Email != "", "email", validator.EmailRequiredErr)
	v.Check(userInput.Password != "", "password", validator.PasswordRequiredErr)
	if !v.Valid() {

		app.failedValidation(w, v.Errors)
		return
	}

	user, token, err := app.services.User.Login(r.Context(), userInput.Email, userInput.Password)
	if err != nil {

		if errors.Is(err, service.ErrAccountNotActive) {

			_ = user
			app.errorJSON(w, http.StatusForbidden, validator.AccountNotActiveResendErr)
			return
		}

		if errors.Is(err, service.ErrInvalidCredentials) {

			app.errorJSON(w, http.StatusUnauthorized, validator.InvalidCredentialsErr)
			return
		}

		app.errorJSON(w, http.StatusInternalServerError, validator.InternalErr, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token.PlainText,
		Path:     "/",
		Expires:  token.Expiry,
		HttpOnly: true,
		Secure:   app.config.env == "production",
		SameSite: http.SameSiteLaxMode,
	})

	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = false
	payload.Message = "Authenticated successfully"

	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) GetCurrentUser(w http.ResponseWriter, r *http.Request) {

	user := app.contextGetUser(r)

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error": false,
		"user":  user,
	})
}

func (app *application) APILogout(w http.ResponseWriter, r *http.Request) {

	c, err := r.Cookie("auth_token")
	if err == nil {
		// delete token from DB
		app.db.Token.DeleteByToken(r.Context(), c.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   app.config.env == "production",
		SameSite: http.SameSiteLaxMode,
	})

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":   false,
		"message": "logged out",
	})
}

func (app *application) RegisterUser(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		Email           string `json:"email"`
		ConfirmEmail    string `json:"confirm_email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.ConfirmEmail = strings.ToLower(strings.TrimSpace(payload.ConfirmEmail))

	v := validator.New()
	v.Check(payload.FirstName != "", "first_name", validator.FirstNameRequiredErr)
	v.Check(payload.LastName != "", "last_name", validator.LastNameRequiredErr)
	v.Check(payload.Email != "", "email", validator.EmailRequiredErr)
	v.Check(payload.Email == payload.ConfirmEmail, "confirm_email", validator.EmailsDoNotMatchErr)
	v.Check(payload.Password != "", "password", validator.PasswordRequiredErr)
	v.Check(len(payload.Password) >= service.MinPasswordBytes, "password", validator.PasswordTooShortErr)
	v.Check(len(payload.Password) <= service.MaxPasswordBytes, "password", validator.PasswordTooLongErr)
	v.Check(payload.Password == payload.ConfirmPassword, "confirm_password", validator.PasswordsDoNotMatchErr)
	if !v.Valid() {

		app.failedValidation(w, v.Errors)
		return
	}

	user := &models.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Role:      models.RoleUser,
		IsActive:  false,
	}

	_, err = app.services.User.Create(r.Context(), user, payload.Password)
	if err != nil {

		if errors.Is(err, service.ErrEmailAlreadyTaken) {

			app.errorJSON(w, http.StatusBadRequest, validator.EmailAlreadyRegisteredErr)
			return
		}

		app.errorJSON(w, http.StatusInternalServerError, validator.InternalErr, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]any{

		"error":   false,
		"message": "Registration successful, your account will be activated by an administrator",
	})
}
