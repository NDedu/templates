package main

import (
	"fmt"
	"net/http"
	"webProj/internal/tokens"
)

func FormatCurrency(n int) string {

	nr := float32(n) / float32(100)
	return fmt.Sprintf("$%.2f", nr)
}

func (app *application) getOrCreateDeviceID(w http.ResponseWriter, r *http.Request) string {

	c, err := r.Cookie("device_id")
	if err == nil && len(c.Value) == tokens.DeviceTokenLen {

		return c.Value
	}

	id := tokens.GenerateSecureToken(tokens.DeviceTokenBytes)
	http.SetCookie(w, &http.Cookie{
		Name:     "device_id",
		Value:    id,
		Path:     "/",
		MaxAge:   365 * 24 * 3600,
		HttpOnly: true,
		Secure:   app.config.env == "production",
		SameSite: http.SameSiteLaxMode,
	})

	return id
}
