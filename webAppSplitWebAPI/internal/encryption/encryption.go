package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encryption holds the initialized AES-GCM cipher
type Encryption struct {
	aead cipher.AEAD
}

// New creates a reusable Encryption instance.
// Key should be 32 bytes for AES-256.
func New(key []byte) (*Encryption, error) {

	if len(key) != 32 {

		return nil, errors.New("key must be exactly 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {

		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {

		return nil, err
	}

	return &Encryption{
		aead: aesGCM, // Store the initialized AEAD
	}, nil
}

func (e *Encryption) Encrypt(text string) (string, error) {

	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {

		return "", err
	}

	// Seal appends the ciphertext to the nonce
	cipherText := e.aead.Seal(nonce, nonce, []byte(text), nil)

	// RawURLEncoding avoids trailing "=" padding
	return base64.RawURLEncoding.EncodeToString(cipherText), nil
}

func (e *Encryption) Decrypt(cryptoText string) (string, error) {

	cipherText, err := base64.RawURLEncoding.DecodeString(cryptoText)
	if err != nil {

		return "", err
	}

	nonceSize := e.aead.NonceSize()
	if len(cipherText) < nonceSize {

		return "", errors.New("ciphertext too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	plainText, err := e.aead.Open(nil, nonce, cipherText, nil)
	if err != nil {

		return "", err // Will return an error if data was tampered with or key is wrong
	}

	return string(plainText), nil
}
