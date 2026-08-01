package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"webProj/internal/models"
	"webProj/internal/service"
	"webProj/internal/urlsigner"
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

			app.sendActivationEmail(user.Email)
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

	// payload.Token = token
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

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {

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

func (app *application) SendPasswordResetEmail(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	_, err = app.db.User.GetUserByEmail(r.Context(), payload.Email)
	if err != nil {

		app.writeJSON(w, http.StatusCreated, map[string]any{

			"error":   false,
			"message": "If an account with that email exists, a reset link has been sent",
		})
		return
	}

	encryptedEmail, err := app.enc.Encrypt(payload.Email)
	if err != nil {

		app.errorJSON(w, http.StatusInternalServerError, validator.InternalErr, err)
		return
	}

	link := fmt.Sprintf("%s/reset-password?email=%s", app.config.frontend, encryptedEmail)

	sign := urlsigner.New([]byte(app.config.secretkey))
	signedLink := sign.GenerateTokenFromString(link)

	var data struct {
		Link string
	}

	data.Link = signedLink

	err = app.SendEmail(app.config.smtp.sender, payload.Email, "Password reset", "password-reset", data)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.EmailQueueErr, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]any{

		"error":   false,
		"message": "If an account with that email exists, a reset link has been sent",
	})
}

func (app *application) ResetPassword(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		SignedURL string `json:"signed_url"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	v := validator.New()
	v.Check(payload.Password != "", "password", validator.PasswordRequiredErr)
	v.Check(len(payload.Password) >= service.MinPasswordBytes, "password", validator.PasswordTooShortErr)
	v.Check(len(payload.Password) <= service.MaxPasswordBytes, "password", validator.PasswordTooLongErr)
	if !v.Valid() {

		app.failedValidation(w, v.Errors)
		return
	}

	signer := urlsigner.New([]byte(app.config.secretkey))
	err = signer.IsValid(payload.SignedURL, 30)
	if err != nil {

		app.errorJSON(w, http.StatusRequestTimeout, validator.ResetLinkExpiredErr)
		return
	}

	decryptedEmail, err := app.enc.Decrypt(payload.Email)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.CouldNotResetPasswordErr)
		return
	}

	user, err := app.db.User.GetUserByEmail(r.Context(), decryptedEmail)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.CouldNotResetPasswordErr)
		return
	}

	err = app.services.User.UpdatePassword(r.Context(), user, payload.Password)
	if err != nil {

		app.errorJSON(w, http.StatusInternalServerError, validator.InternalErr, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]any{

		"error":   false,
		"message": "Password changed",
	})
}

func (app *application) RegisterUser(w http.ResponseWriter, r *http.Request) {

	var payload struct {
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
		Email:    payload.Email,
		Role:     models.RoleUser,
		IsActive: false,
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

	err = app.sendActivationEmail(payload.Email)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.EmailQueueErr, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]any{

		"error":   false,
		"message": "Registration successful, please check your email to activate your account",
	})
}

func (app *application) sendActivationEmail(email string) error {

	encryptedEmail, err := app.enc.Encrypt(email)
	if err != nil {

		return err
	}

	link := fmt.Sprintf("%s/activate-account?email=%s", app.config.frontend, encryptedEmail)

	sign := urlsigner.New([]byte(app.config.secretkey))
	signedLink := sign.GenerateTokenFromString(link)

	var data struct {
		Link string
	}
	data.Link = signedLink

	return app.SendEmail(app.config.smtp.sender, email, "Activate your account", "account-activation", data)
}

func (app *application) ActivateAccount(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		Email     string `json:"email"`
		SignedURL string `json:"signed_url"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	signer := urlsigner.New([]byte(app.config.secretkey))
	err = signer.IsValid(payload.SignedURL, 30)
	if err != nil {

		app.errorJSON(w, http.StatusRequestTimeout, validator.ActivationLinkExpiredErr)
		return
	}

	decryptedEmail, err := app.enc.Decrypt(payload.Email)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.AccountActivationFailedErr)
		return
	}

	user, err := app.db.User.GetUserByEmail(r.Context(), decryptedEmail)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.AccountActivationFailedErr)
		return
	}

	if user.IsActive {

		app.writeJSON(w, http.StatusOK, map[string]any{
			"error":   false,
			"message": "Account is already active",
		})
		return
	}

	err = app.db.User.Activate(r.Context(), user.ID)
	if err != nil {

		app.errorJSON(w, http.StatusInternalServerError, validator.InternalErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":   false,
		"message": "Account activated successfully",
	})
}

func (app *application) ResendActivation(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	user, err := app.db.User.GetUserByEmail(r.Context(), payload.Email)
	if err != nil || user.IsActive {

		app.writeJSON(w, http.StatusOK, map[string]any{
			"error":   false,
			"message": "If an account with that email exists and is inactive, an activation link has been sent",
		})
		return
	}

	app.sendActivationEmail(user.Email)

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":   false,
		"message": "If an account with that email exists and is inactive, an activation link has been sent",
	})
}
