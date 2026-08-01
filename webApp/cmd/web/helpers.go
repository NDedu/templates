package main

import (
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"webProj/internal/tokens"
	"webProj/internal/validator"
)

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data any) error {

	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576) // 1MB

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(data); err != nil {

		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {

		return errors.New(validator.InvalidJSONErr)
	}

	return nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) {

	out, err := json.Marshal(data)
	if err != nil {

		app.errorLog.Println(err)
		http.Error(w, validator.InternalErr, http.StatusInternalServerError)
		return
	}

	for _, h := range headers {

		maps.Copy(w.Header(), h)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(out); err != nil {

		app.errorLog.Println("writeJSON write error:", err)
	}
}

func (app *application) errorJSON(w http.ResponseWriter, status int, msg string, err ...error) {

	if len(err) > 0 && err[0] != nil {

		app.errorLog.Println(err[0])
	}

	app.writeJSON(w, status, map[string]any{
		"error":   true,
		"message": msg,
	})
}

func (app *application) failedValidation(w http.ResponseWriter, errs map[string]string) {

	app.writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
		"error":   true,
		"message": validator.ValidationFailedErr,
		"errors":  errs,
	})
}

func (app *application) getOrCreateDeviceID(w http.ResponseWriter, r *http.Request) string {

	if c, err := r.Cookie("device_id"); err == nil && len(c.Value) == tokens.DeviceTokenLen {

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
