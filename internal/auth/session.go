package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const SessionCookieName = "rubrical_session"

type Session struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
}

func NewSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func SetSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

// SetEmbedSessionCookie sets a CHIPS cookie for the Canvas iframe partition.
// Requires HTTPS. Used after embed handoff and after password login inside the modal.
func SetEmbedSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:        SessionCookieName,
		Value:       token,
		Path:        "/",
		HttpOnly:    true,
		Secure:      true,
		SameSite:    http.SameSiteNoneMode,
		Partitioned: true,
		Expires:     expiresAt,
	})
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// ClearEmbedSessionCookie clears the CHIPS session in the Canvas iframe partition.
func ClearEmbedSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:        SessionCookieName,
		Value:       "",
		Path:        "/",
		HttpOnly:    true,
		Secure:      true,
		SameSite:    http.SameSiteNoneMode,
		Partitioned: true,
		MaxAge:      -1,
	})
}

func SessionTokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func NewOAuthState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func OAuthStateCookieName() string {
	return "rubrical_oauth_state"
}

func SetOAuthStateCookie(w http.ResponseWriter, state string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthStateCookieName(),
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
}

func ClearOAuthStateCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthStateCookieName(),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func OAuthStateFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	cookie, err := r.Cookie(OAuthStateCookieName())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidateEmail(email string) error {
	email = NormalizeEmail(email)
	if email == "" || !strings.Contains(email, "@") || strings.HasPrefix(email, "@") {
		return fmt.Errorf("invalid email address")
	}
	return nil
}
