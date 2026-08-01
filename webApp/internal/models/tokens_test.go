package models_test

import (
	"testing"
	"time"
	"webProj/internal/models"
	"webProj/internal/testutil"
	"webProj/internal/tokens"
)

func TestTokenModel_Insert(t *testing.T) {
	t.Parallel()

	t.Run("inserts token for user", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		user, token := testutil.CreateUserAndToken(t, m)

		// verify we can retrieve the user by token
		got, err := m.Token.GetUserByToken(t.Context(), token.PlainText)
		if err != nil {
			t.Fatalf("get user by token: %v", err)
		}
		if got.ID != user.ID {
			t.Errorf("want user ID %d; got %d", user.ID, got.ID)
		}
	})

	t.Run("allows multiple active tokens for same user", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		user := testutil.CreateUser(t, m, "multi@test.com", "password123", models.RoleUser, true)

		var plainTexts []string
		for range 5 {

			token, err := tokens.GenerateAuthToken(user.ID, 24*time.Hour, tokens.ScopeAuthentication)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			err = m.Token.Insert(t.Context(), token, user)
			if err != nil {
				t.Fatalf("insert: %v", err)
			}
			plainTexts = append(plainTexts, token.PlainText)
		}

		// all tokens should still resolve to the user
		for i, pt := range plainTexts {

			got, err := m.Token.GetUserByToken(t.Context(), pt)
			if err != nil {
				t.Fatalf("token %d: get user: %v", i, err)
			}
			if got.ID != user.ID {
				t.Errorf("token %d: want user ID %d; got %d", i, user.ID, got.ID)
			}
		}
	})

	t.Run("cleans up expired tokens on insert", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		user := testutil.CreateUser(t, m, "expired@test.com", "password123", models.RoleUser, true)

		// insert a token that's already expired
		expiredToken, err := tokens.GenerateAuthToken(user.ID, -1*time.Hour, tokens.ScopeAuthentication)
		if err != nil {
			t.Fatalf("generate expired: %v", err)
		}
		err = m.Token.Insert(t.Context(), expiredToken, user)
		if err != nil {
			t.Fatalf("insert expired: %v", err)
		}

		// insert a new valid token — should clean expired ones
		newToken, err := tokens.GenerateAuthToken(user.ID, 24*time.Hour, tokens.ScopeAuthentication)
		if err != nil {
			t.Fatalf("generate new: %v", err)
		}
		err = m.Token.Insert(t.Context(), newToken, user)
		if err != nil {
			t.Fatalf("insert new: %v", err)
		}

		// expired token should no longer resolve
		_, err = m.Token.GetUserByToken(t.Context(), expiredToken.PlainText)
		if err == nil {
			t.Error("want error for expired token; got nil")
		}

		// new token should work
		got, err := m.Token.GetUserByToken(t.Context(), newToken.PlainText)
		if err != nil {
			t.Fatalf("get new token: %v", err)
		}
		if got.ID != user.ID {
			t.Errorf("want user ID %d; got %d", user.ID, got.ID)
		}
	})
}

func TestTokenModel_GetUserByToken(t *testing.T) {
	t.Parallel()

	t.Run("returns correct user fields", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		user, token := testutil.CreateUserAndToken(t, m)

		got, err := m.Token.GetUserByToken(t.Context(), token.PlainText)
		if err != nil {
			t.Fatalf("get user: %v", err)
		}

		if got.ID != user.ID {
			t.Errorf("want ID %d; got %d", user.ID, got.ID)
		}
		if got.UUID != user.UUID {
			t.Errorf("want UUID %s; got %s", user.UUID, got.UUID)
		}
		if got.Email != user.Email {
			t.Errorf("want email %s; got %s", user.Email, got.Email)
		}
		if got.Role != user.Role {
			t.Errorf("want role %s; got %s", user.Role, got.Role)
		}
		if got.IsActive != user.IsActive {
			t.Errorf("want is_active %v; got %v", user.IsActive, got.IsActive)
		}
	})

	t.Run("returns error for invalid token", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		testutil.CreateUserAndToken(t, m)

		_, err := m.Token.GetUserByToken(t.Context(), "invalid-token-string")
		if err == nil {
			t.Error("want error for invalid token; got nil")
		}
	})

	t.Run("returns error for expired token", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		user := testutil.CreateUser(t, m, "exp@test.com", "password123", models.RoleUser, true)

		token, err := tokens.GenerateAuthToken(user.ID, -1*time.Hour, tokens.ScopeAuthentication)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		err = m.Token.Insert(t.Context(), token, user)
		if err != nil {
			t.Fatalf("insert: %v", err)
		}

		_, err = m.Token.GetUserByToken(t.Context(), token.PlainText)
		if err == nil {
			t.Error("want error for expired token; got nil")
		}
	})

	t.Run("distinguishes tokens among many users", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		users := testutil.CreateUsers(t, m, 10)

		tokenMap := make(map[string]int) // plainText -> userID
		for _, u := range users {

			token, err := tokens.GenerateAuthToken(u.ID, 24*time.Hour, tokens.ScopeAuthentication)
			if err != nil {
				t.Fatalf("generate for user %d: %v", u.ID, err)
			}
			err = m.Token.Insert(t.Context(), token, u)
			if err != nil {
				t.Fatalf("insert for user %d: %v", u.ID, err)
			}
			tokenMap[token.PlainText] = u.ID
		}

		for pt, expectedID := range tokenMap {

			got, err := m.Token.GetUserByToken(t.Context(), pt)
			if err != nil {
				t.Fatalf("get user for token: %v", err)
			}
			if got.ID != expectedID {
				t.Errorf("want user ID %d; got %d", expectedID, got.ID)
			}
		}
	})
}

func TestTokenModel_DeleteByToken(t *testing.T) {
	t.Parallel()

	t.Run("deletes token so it no longer resolves", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		_, token := testutil.CreateUserAndToken(t, m)

		err := m.Token.DeleteByToken(t.Context(), token.PlainText)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		_, err = m.Token.GetUserByToken(t.Context(), token.PlainText)
		if err == nil {
			t.Error("want error after delete; got nil")
		}
	})

	t.Run("does not error on nonexistent token", func(t *testing.T) {

		m := testutil.NewTestDB(t)

		err := m.Token.DeleteByToken(t.Context(), "nonexistent-token")
		if err != nil {
			t.Errorf("want nil error; got %v", err)
		}
	})

	t.Run("only deletes the specific token", func(t *testing.T) {

		m := testutil.NewTestDB(t)
		user := testutil.CreateUser(t, m, "multi@test.com", "password123", models.RoleUser, true)

		token1, err := tokens.GenerateAuthToken(user.ID, 24*time.Hour, tokens.ScopeAuthentication)
		if err != nil {
			t.Fatalf("generate token1: %v", err)
		}
		m.Token.Insert(t.Context(), token1, user)

		token2, err := tokens.GenerateAuthToken(user.ID, 24*time.Hour, tokens.ScopeAuthentication)
		if err != nil {
			t.Fatalf("generate token2: %v", err)
		}
		m.Token.Insert(t.Context(), token2, user)

		// delete only token1
		err = m.Token.DeleteByToken(t.Context(), token1.PlainText)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}

		// token1 gone
		_, err = m.Token.GetUserByToken(t.Context(), token1.PlainText)
		if err == nil {
			t.Error("want error for deleted token1; got nil")
		}

		// token2 still works
		got, err := m.Token.GetUserByToken(t.Context(), token2.PlainText)
		if err != nil {
			t.Fatalf("get token2: %v", err)
		}
		if got.ID != user.ID {
			t.Errorf("want user ID %d; got %d", user.ID, got.ID)
		}
	})
}
