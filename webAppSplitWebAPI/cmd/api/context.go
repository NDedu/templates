package main

import (
	"context"
	"net/http"
	"webProj/internal/models"
)

type contextKey string

const contextKeyUser = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *models.User) *http.Request {

	ctx := context.WithValue(r.Context(), contextKeyUser, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *models.User {

	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {

		return nil
	}
	return user
}
