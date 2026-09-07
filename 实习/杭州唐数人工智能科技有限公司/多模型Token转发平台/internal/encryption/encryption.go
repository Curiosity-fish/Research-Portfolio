// Package encryption provides AES-256-GCM helpers for storing secrets (such
// as upstream API keys) in the database. Each encrypted value is a base64url
// string of the form: base64url(nonce || ciphertext), where nonce is 12 bytes.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var errInvalidKeyLen = errors.New("encryption key must be exactly 32 bytes (AES-256)")

// Encrypt encrypts plaintext with AES-256-GCM using the provided 32-byte key.
// The returned string is safe to store in a database column.
func Encrypt(key []byte, plaintext string) (string, error) {
	if len(key) != 32 {
		return "", errInvalidKeyLen
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	// Seal appends ciphertext+tag to the nonce.
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt. It returns the original plaintext.
func Decrypt(key []byte, encoded string) (string, error) {
	if len(key) != 32 {
		return "", errInvalidKeyLen
	}

	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	ns := gcm.NonceSize()
	if len(data) < ns {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:ns], data[ns:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// ParseKey decodes a 32-byte key from base64 (standard or URL-safe, with or
// without padding). It returns errInvalidKeyLen if the decoded length is wrong.
func ParseKey(b64 string) ([]byte, error) {
	key, err := base64.RawURLEncoding.DecodeString(b64)
	if err != nil {
		// Fall back to standard encoding.
		key, err = base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("decode encryption key: %w", err)
		}
	}
	if len(key) != 32 {
		return nil, errInvalidKeyLen
	}
	return key, nil
}
