package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const (
	ScopeAuthentication = "authentication"

	// DeviceTokenBytes is the number of random bytes used for device tokens.
	DeviceTokenBytes = 32

	// DeviceTokenLen is the expected length of a base64url-encoded device token.
	// base64url encodes 32 bytes as ceil(32*4/3) = 44 characters (with padding).
	DeviceTokenLen = 44
)

type Token struct {
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// Generate creates a signed token: timestamp.signature
// The data (e.g. client IP or session ID) is bound to the signature but not embedded
// in the token, so it isn't leaked to the client and avoids delimiter
// collisions with IPv4/IPv6 addresses.
func Generate(secret, data string) string {

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	payload := timestamp + "|" + data
	sig := sign(secret, payload)
	return timestamp + "." + sig
}

// Verify checks the signature and that the token isn't expired.
func Verify(secret, token, data string, maxAge time.Duration) bool {

	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {

		return false
	}

	timestamp, signature := parts[0], parts[1]

	// verify signature
	payload := timestamp + "|" + data
	expected := sign(secret, payload)
	if !hmac.Equal([]byte(signature), []byte(expected)) {

		return false
	}

	// verify not expired
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {

		return false
	}

	if time.Since(time.Unix(ts, 0)) > maxAge {

		return false
	}

	return true
}

func sign(secret, data string) string {

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSecureToken generates a random base64 encoded token of length n.
func GenerateSecureToken(n int) string {

	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// FIX: Deletes all tokens on login — means logging in on one device logs out everywhere. Consider keeping multiple tokens and deleting by device/scope instead
func GenerateAuthToken(userID int, ttl time.Duration, scope string) (*Token, error) {

	token := &Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {

		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))

	token.Hash = hash[:]

	return token, nil
}
