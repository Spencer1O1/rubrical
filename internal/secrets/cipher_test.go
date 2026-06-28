package secrets

import (
	"encoding/base64"
	"errors"
	"testing"
)

func testCipher(t *testing.T) *Cipher {
	t.Helper()
	key := make([]byte, keySize)
	c, err := NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCipher_encryptDecrypt_roundTrip(t *testing.T) {
	c := testCipher(t)

	plain := "sk-test-key-12345"
	encrypted, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == plain {
		t.Fatal("expected ciphertext")
	}
	if !c.IsEncrypted(encrypted) {
		t.Fatalf("expected encrypted prefix, got %q", encrypted)
	}

	got, err := c.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("got %q want %q", got, plain)
	}
}

func TestCipher_decryptRejectsPlaintext(t *testing.T) {
	c := testCipher(t)

	_, err := c.Decrypt("sk-plaintext")
	if !errors.Is(err, ErrNotEncrypted) {
		t.Fatalf("expected ErrNotEncrypted, got %v", err)
	}
}

func TestNewCipherFromEnv_parsesBase64(t *testing.T) {
	key := make([]byte, keySize)
	encoded := base64.StdEncoding.EncodeToString(key)
	c, err := NewCipherFromEnv(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("expected cipher")
	}
}

func TestNewCipherFromEnv_requiresKey(t *testing.T) {
	_, err := NewCipherFromEnv("")
	if !errors.Is(err, ErrMissingKey) {
		t.Fatalf("expected ErrMissingKey, got %v", err)
	}
}
