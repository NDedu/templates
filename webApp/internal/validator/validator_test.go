package validator

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	v := New()
	if v == nil {
		t.Fatal("want non-nil validator")
	}
	if v.Errors == nil {
		t.Fatal("want non-nil errors map")
	}
	if len(v.Errors) != 0 {
		t.Errorf("want empty errors; got %d", len(v.Errors))
	}
}

func TestValid(t *testing.T) {
	t.Parallel()

	t.Run("valid when no errors", func(t *testing.T) {

		v := New()
		if !v.Valid() {
			t.Error("want Valid()=true for new validator")
		}
	})

	t.Run("invalid when errors exist", func(t *testing.T) {

		v := New()
		v.AddError("field", "error")
		if v.Valid() {
			t.Error("want Valid()=false after AddError")
		}
	})
}

func TestAddError(t *testing.T) {
	t.Parallel()

	t.Run("adds error for new key", func(t *testing.T) {

		v := New()
		v.AddError("email", "required")
		if v.Errors["email"] != "required" {
			t.Errorf("want error 'required'; got %q", v.Errors["email"])
		}
	})

	t.Run("does not overwrite existing key", func(t *testing.T) {

		v := New()
		v.AddError("email", "first error")
		v.AddError("email", "second error")
		if v.Errors["email"] != "first error" {
			t.Errorf("want first error preserved; got %q", v.Errors["email"])
		}
	})

	t.Run("handles multiple keys", func(t *testing.T) {

		v := New()
		v.AddError("email", "required")
		v.AddError("password", "too short")
		if len(v.Errors) != 2 {
			t.Errorf("want 2 errors; got %d", len(v.Errors))
		}
	})
}

func TestCheck(t *testing.T) {
	t.Parallel()

	t.Run("does not add error when condition is true", func(t *testing.T) {

		v := New()
		v.Check(true, "email", "required")
		if !v.Valid() {
			t.Error("want no error when check passes")
		}
	})

	t.Run("adds error when condition is false", func(t *testing.T) {

		v := New()
		v.Check(false, "email", "required")
		if v.Valid() {
			t.Error("want error when check fails")
		}
		if v.Errors["email"] != "required" {
			t.Errorf("want error 'required'; got %q", v.Errors["email"])
		}
	})

	t.Run("multiple checks accumulate errors", func(t *testing.T) {

		v := New()
		v.Check(false, "email", "required")
		v.Check(false, "password", "too short")
		v.Check(true, "name", "required")
		if len(v.Errors) != 2 {
			t.Errorf("want 2 errors; got %d", len(v.Errors))
		}
	})

	t.Run("first error wins for same key", func(t *testing.T) {

		v := New()
		v.Check(false, "password", "required")
		v.Check(false, "password", "too short")
		if v.Errors["password"] != "required" {
			t.Errorf("want first error 'required'; got %q", v.Errors["password"])
		}
	})
}
