package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	keySize   = 32
	nonceSize = 12
	prefix    = "enc:v1:"
)

var (
	ErrInvalidKey  = errors.New("secrets encryption key must decode to 32 bytes")
	ErrMissingKey  = errors.New("SECRETS_ENCRYPTION_KEY is required")
	ErrNotEncrypted = errors.New("secret value is not encrypted")
)

type Cipher struct {
	aead cipher.AEAD
}

func NewCipher(key []byte) (*Cipher, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("%w (got %d bytes)", ErrInvalidKey, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

func NewCipherFromEnv(raw string) (*Cipher, error) {
	key, err := parseKey(raw)
	if err != nil {
		return nil, err
	}
	return NewCipher(key)
}

func parseKey(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("%w (run: pnpm setup:secrets-key)", ErrMissingKey)
	}

	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == keySize {
		return decoded, nil
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(decoded) == keySize {
		return decoded, nil
	}
	if decoded, err := hex.DecodeString(raw); err == nil && len(decoded) == keySize {
		return decoded, nil
	}
	return nil, fmt.Errorf("%w: use pnpm setup:secrets-key or openssl rand -base64 32", ErrInvalidKey)
}

func (c *Cipher) IsEncrypted(value string) bool {
	return strings.HasPrefix(value, prefix)
}

func (c *Cipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if c == nil || c.aead == nil {
		return "", errors.New("secrets cipher is not configured")
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := c.aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, sealed...)
	return prefix + base64.RawStdEncoding.EncodeToString(payload), nil
}

func (c *Cipher) Decrypt(stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if !c.IsEncrypted(stored) {
		return "", ErrNotEncrypted
	}
	if c == nil || c.aead == nil {
		return "", errors.New("secrets cipher is not configured")
	}

	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(stored, prefix))
	if err != nil {
		return "", fmt.Errorf("decode encrypted secret: %w", err)
	}
	if len(payload) < nonceSize {
		return "", errors.New("encrypted secret payload is too short")
	}

	nonce := payload[:nonceSize]
	ciphertext := payload[nonceSize:]
	plain, err := c.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: %w", err)
	}
	return string(plain), nil
}
