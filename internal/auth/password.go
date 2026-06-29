package auth

import (
	"crypto/subtle"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 8
	bcryptCost        = 12
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrWeakPassword = errors.New("password must be at least 8 characters")

func HashPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < MinPasswordLength {
		return "", ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) bool {
	if hash == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
