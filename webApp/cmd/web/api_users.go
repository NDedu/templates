package main

import (
	"database/sql"
	"errors"
	"net/http"
	"webProj/internal/models"
	"webProj/internal/service"
	"webProj/internal/validator"

	"github.com/go-chi/chi/v5"
)

func (app *application) GetUser(w http.ResponseWriter, r *http.Request) {

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidIDErr)
		return
	}

	user, err := app.db.User.GetByUUID(r.Context(), uuid)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {

			app.errorJSON(w, http.StatusNotFound, validator.UserNotFoundErr)
			return
		}
		app.errorJSON(w, http.StatusInternalServerError, validator.InternalErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error": false,
		"user":  user,
	})
}

func (app *application) EditUser(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		UUID      string `json:"uuid"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		IsActive  bool   `json:"is_active"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	v := validator.New()
	v.Check(payload.FirstName != "", "first_name", validator.FirstNameRequiredErr)
	v.Check(payload.LastName != "", "last_name", validator.LastNameRequiredErr)
	v.Check(payload.Email != "", "email", validator.EmailRequiredErr)
	v.Check(payload.Role == models.RoleUser || payload.Role == models.RoleEditor || payload.Role == models.RoleAdmin, "role", validator.InvalidRoleErr)
	if !v.Valid() {

		app.failedValidation(w, v.Errors)
		return
	}

	err := app.services.User.Update(r.Context(), payload.UUID, payload.FirstName, payload.LastName, payload.Email, payload.Role, payload.IsActive)
	if err != nil {

		if errors.Is(err, service.ErrUserNotFound) {

			app.errorJSON(w, http.StatusNotFound, validator.UserNotFoundErr)
			return
		}

		if errors.Is(err, service.ErrEmailAlreadyTaken) {

			app.errorJSON(w, http.StatusBadRequest, validator.EmailAlreadyRegisteredErr)
			return
		}

		app.errorJSON(w, http.StatusInternalServerError, validator.UpdateUserErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":   false,
		"message": "User updated",
	})
}

func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidIDErr)
		return
	}

	currentUser := app.contextGetUser(r)
	currentUUID := ""
	if currentUser != nil {

		currentUUID = currentUser.UUID
	}

	err := app.services.User.Delete(r.Context(), currentUUID, uuid)
	if err != nil {

		if errors.Is(err, service.ErrCannotDeleteSelf) {

			app.errorJSON(w, http.StatusForbidden, validator.CannotDeleteSelfErr)
			return
		}

		app.errorJSON(w, http.StatusInternalServerError, validator.DeleteUserErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":   false,
		"message": "User deleted",
	})
}

func (app *application) AddUser(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		Role      string `json:"role"`
		IsActive  bool   `json:"is_active"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	v := validator.New()
	v.Check(payload.FirstName != "", "first_name", validator.FirstNameRequiredErr)
	v.Check(payload.LastName != "", "last_name", validator.LastNameRequiredErr)
	v.Check(payload.Email != "", "email", validator.EmailRequiredErr)
	v.Check(payload.Password != "", "password", validator.PasswordRequiredErr)
	v.Check(len(payload.Password) >= service.MinPasswordBytes, "password", validator.PasswordTooShortErr)
	v.Check(len(payload.Password) <= service.MaxPasswordBytes, "password", validator.PasswordTooLongErr)
	v.Check(payload.Role == models.RoleUser || payload.Role == models.RoleEditor || payload.Role == models.RoleAdmin, "role", validator.InvalidRoleErr)
	if !v.Valid() {

		app.failedValidation(w, v.Errors)
		return
	}

	user := &models.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Role:      payload.Role,
		IsActive:  payload.IsActive,
	}

	id, err := app.services.User.Create(r.Context(), user, payload.Password)
	if err != nil {

		if errors.Is(err, service.ErrEmailAlreadyTaken) {

			app.errorJSON(w, http.StatusBadRequest, validator.EmailAlreadyRegisteredErr)
			return
		}

		app.errorJSON(w, http.StatusInternalServerError, validator.CreateUserErr, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]any{
		"error":   false,
		"message": "User created",
		"id":      id,
	})
}

func (app *application) ResetUserPassword(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		UUID            string `json:"uuid"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	if err := app.readJSON(w, r, &payload); err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidJSONErr)
		return
	}

	v := validator.New()
	v.Check(payload.Password != "", "password", validator.PasswordRequiredErr)
	v.Check(len(payload.Password) >= service.MinPasswordBytes, "password", validator.PasswordTooShortErr)
	v.Check(len(payload.Password) <= service.MaxPasswordBytes, "password", validator.PasswordTooLongErr)
	v.Check(payload.Password == payload.ConfirmPassword, "confirm_password", validator.PasswordsDoNotMatchErr)
	if !v.Valid() {

		app.failedValidation(w, v.Errors)
		return
	}

	user, err := app.db.User.GetByUUID(r.Context(), payload.UUID)
	if err != nil {

		app.errorJSON(w, http.StatusNotFound, validator.UserNotFoundErr)
		return
	}

	err = app.services.User.UpdatePassword(r.Context(), user, payload.Password)
	if err != nil {

		app.errorJSON(w, http.StatusInternalServerError, validator.InternalErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":   false,
		"message": "Password updated",
	})
}

func (app *application) UploadProfileImage(w http.ResponseWriter, r *http.Request) {

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {

		app.errorJSON(w, http.StatusBadRequest, validator.InvalidIDErr)
		return
	}

	// 5MB max
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(5 << 20); err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.ImageTooLargeErr)
		return
	}

	file, _, err := r.FormFile("profile_image")
	if err != nil {

		app.errorJSON(w, http.StatusBadRequest, validator.UploadImageErr)
		return
	}
	defer file.Close()

	imagePath, err := app.services.User.UploadProfileImage(r.Context(), uuid, file, app.staticPath)
	if err != nil {

		if errors.Is(err, service.ErrInvalidImageType) {

			app.errorJSON(w, http.StatusBadRequest, validator.InvalidImageTypeErr)
			return
		}

		if errors.Is(err, service.ErrUserNotFound) {

			app.errorJSON(w, http.StatusNotFound, validator.UserNotFoundErr)
			return
		}

		app.errorJSON(w, http.StatusInternalServerError, validator.UploadImageErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":         false,
		"message":       "Profile image updated",
		"profile_image": imagePath,
	})
}

func (app *application) GetAllUsers(w http.ResponseWriter, r *http.Request) {

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

	users, currentPage, lastPage, totalRecords, err := app.db.User.GetAllPaginated(r.Context(), payload.PageSize, payload.CurrentPage, payload.SortCol, payload.SortDir)
	if err != nil {

		app.errorJSON(w, http.StatusInternalServerError, validator.RetrieveUsersErr, err)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]any{
		"error":         false,
		"users":         users,
		"current_page":  currentPage,
		"page_size":     payload.PageSize,
		"last_page":     lastPage,
		"total_records": totalRecords,
	})
}
