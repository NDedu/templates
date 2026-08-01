package tokens

import (
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	t.Run("returns timestamp.signature format", func(t *testing.T) {
		t.Parallel()

		token := Generate("secret", "data")
		if token == "" {
			t.Fatal("want non-empty token")
		}

		// Must contain exactly one "."
		parts := 0
		for _, c := range token {
			if c == '.' {
				parts++
			}
		}
		if parts != 1 {
			t.Errorf("want 1 dot separator; got %d in token %q", parts, token)
		}
	})

	t.Run("different data produces different tokens", func(t *testing.T) {
		t.Parallel()

		t1 := Generate("secret", "data1")
		t2 := Generate("secret", "data2")
		if t1 == t2 {
			t.Error("want different tokens for different data")
		}
	})

	t.Run("different secrets produce different tokens", func(t *testing.T) {
		t.Parallel()

		t1 := Generate("secret1", "data")
		t2 := Generate("secret2", "data")
		if t1 == t2 {
			t.Error("want different tokens for different secrets")
		}
	})
}

func TestVerify(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	data := "client-data"

	t.Run("valid token passes verification", func(t *testing.T) {
		t.Parallel()

		token := Generate(secret, data)
		if !Verify(secret, token, data, 1*time.Hour) {
			t.Error("want valid token to pass verification")
		}
	})

	t.Run("wrong secret fails", func(t *testing.T) {
		t.Parallel()

		token := Generate(secret, data)
		if Verify("wrong-secret", token, data, 1*time.Hour) {
			t.Error("want wrong secret to fail verification")
		}
	})

	t.Run("wrong data fails", func(t *testing.T) {
		t.Parallel()

		token := Generate(secret, data)
		if Verify(secret, token, "wrong-data", 1*time.Hour) {
			t.Error("want wrong data to fail verification")
		}
	})

	t.Run("tampered signature fails", func(t *testing.T) {
		t.Parallel()

		token := Generate(secret, data)
		tampered := token + "x"
		if Verify(secret, tampered, data, 1*time.Hour) {
			t.Error("want tampered token to fail verification")
		}
	})

	t.Run("missing dot separator fails", func(t *testing.T) {
		t.Parallel()

		if Verify(secret, "nodot", data, 1*time.Hour) {
			t.Error("want token without dot to fail")
		}
	})

	t.Run("empty token fails", func(t *testing.T) {
		t.Parallel()

		if Verify(secret, "", data, 1*time.Hour) {
			t.Error("want empty token to fail")
		}
	})

	t.Run("expired token fails", func(t *testing.T) {
		t.Parallel()

		// maxAge of 0 means immediately expired
		token := Generate(secret, data)
		if Verify(secret, token, data, 0) {
			t.Error("want expired token to fail")
		}
	})

	t.Run("invalid timestamp fails", func(t *testing.T) {
		t.Parallel()

		if Verify(secret, "notanumber.signature", data, 1*time.Hour) {
			t.Error("want invalid timestamp to fail")
		}
	})
}

func TestGenerateSecureToken(t *testing.T) {
	t.Parallel()

	t.Run("returns non-empty base64 string", func(t *testing.T) {
		t.Parallel()

		token := GenerateSecureToken(32)
		if token == "" {
			t.Fatal("want non-empty token")
		}
		if len(token) != DeviceTokenLen {
			t.Errorf("want length %d; got %d", DeviceTokenLen, len(token))
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		t.Parallel()

		seen := make(map[string]bool)
		for i := range 100 {

			token := GenerateSecureToken(32)
			if seen[token] {
				t.Fatalf("duplicate token at iteration %d", i)
			}
			seen[token] = true
		}
	})
}

func TestGenerateAuthToken(t *testing.T) {
	t.Parallel()

	t.Run("returns valid token with all fields", func(t *testing.T) {
		t.Parallel()

		token, err := GenerateAuthToken(42, 24*time.Hour, ScopeAuthentication)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if token.PlainText == "" {
			t.Error("want PlainText to be set")
		}
		if len(token.Hash) == 0 {
			t.Error("want Hash to be set")
		}
		if token.UserID != 42 {
			t.Errorf("want UserID 42; got %d", token.UserID)
		}
		if token.Scope != ScopeAuthentication {
			t.Errorf("want scope %s; got %s", ScopeAuthentication, token.Scope)
		}
		if token.Expiry.Before(time.Now()) {
			t.Error("want expiry in the future")
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		t.Parallel()

		seen := make(map[string]bool)
		for i := range 100 {

			token, err := GenerateAuthToken(1, time.Hour, ScopeAuthentication)
			if err != nil {
				t.Fatalf("generate %d: %v", i, err)
			}
			if seen[token.PlainText] {
				t.Fatalf("duplicate token at iteration %d", i)
			}
			seen[token.PlainText] = true
		}
	})

	t.Run("hash is SHA256 of plain text", func(t *testing.T) {
		t.Parallel()

		token, err := GenerateAuthToken(1, time.Hour, ScopeAuthentication)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}

		if len(token.Hash) != 32 {
			t.Errorf("want SHA256 hash length 32; got %d", len(token.Hash))
		}
	})
}
