package encryption

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}

	cases := []string{
		"sk-test-api-key-12345",
		"",
		strings.Repeat("x", 1000),
		"含有 UTF-8 字符的密钥",
	}

	for _, plaintext := range cases {
		enc, err := Encrypt(key, plaintext)
		if err != nil {
			t.Fatalf("Encrypt(%q): %v", plaintext, err)
		}
		got, err := Decrypt(key, enc)
		if err != nil {
			t.Fatalf("Decrypt(%q): %v", enc, err)
		}
		if got != plaintext {
			t.Errorf("round-trip: got %q, want %q", got, plaintext)
		}
	}
}

func TestEncrypt_DifferentNonceEachCall(t *testing.T) {
	key := make([]byte, 32)
	enc1, _ := Encrypt(key, "same plaintext")
	enc2, _ := Encrypt(key, "same plaintext")
	if enc1 == enc2 {
		t.Error("expected different ciphertexts for separate calls (nonce should differ)")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key := make([]byte, 32)
	enc, _ := Encrypt(key, "secret")

	wrongKey := make([]byte, 32)
	wrongKey[0] = 0xFF
	if _, err := Decrypt(wrongKey, enc); err == nil {
		t.Error("expected error decrypting with wrong key")
	}
}

func TestDecrypt_Tampered(t *testing.T) {
	key := make([]byte, 32)
	enc, _ := Encrypt(key, "secret")

	// Flip the last byte.
	raw, _ := base64.RawURLEncoding.DecodeString(enc)
	raw[len(raw)-1] ^= 0xFF
	tampered := base64.RawURLEncoding.EncodeToString(raw)

	if _, err := Decrypt(key, tampered); err == nil {
		t.Error("expected error decrypting tampered ciphertext")
	}
}

func TestEncrypt_InvalidKeyLen(t *testing.T) {
	_, err := Encrypt([]byte("too-short"), "x")
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestParseKey(t *testing.T) {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i)
	}

	encoded := base64.RawURLEncoding.EncodeToString(raw)
	key, err := ParseKey(encoded)
	if err != nil {
		t.Fatalf("ParseKey: %v", err)
	}
	if len(key) != 32 {
		t.Errorf("expected 32-byte key, got %d", len(key))
	}

	// Standard base64 should also work.
	encodedStd := base64.StdEncoding.EncodeToString(raw)
	key2, err := ParseKey(encodedStd)
	if err != nil {
		t.Fatalf("ParseKey (std): %v", err)
	}
	if len(key2) != 32 {
		t.Errorf("expected 32-byte key, got %d", len(key2))
	}
}

func TestParseKey_TooShort(t *testing.T) {
	short := base64.RawURLEncoding.EncodeToString([]byte("16bytesonly12345"))
	if _, err := ParseKey(short); err == nil {
		t.Error("expected error for short key")
	}
}
