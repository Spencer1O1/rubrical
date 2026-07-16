package auth

// Canvas embed session handoff (CHIPS) — correct model for third-party iframes:
//
//  1. Extension service worker (first-party cookie, SameSite=Lax) GET /auth/embed-url?next=…
//  2. Server returns a short-lived signed URL: /auth/embed?token=…&next=…
//  3. Modal iframe navigates to that URL as its first document load.
//  4. Server validates the token, creates a session, Set-Cookie with
//     SameSite=None; Secure; Partitioned (CHIPS, keyed to the top-level site),
//     then 303 to next.
//  5. Later iframe navigations / HTMX send that partitioned cookie.
//
// embed=1 means “inside the extension modal”: EmbedLayout / bare auth (modal owns chrome).

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const EmbedHandoffTTL = 2 * time.Minute

// MintEmbedHandoffToken returns a short-lived signed token for step (2) above.
func MintEmbedHandoffToken(secret string, userID int64, now time.Time) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", fmt.Errorf("embed handoff secret is required")
	}
	if userID <= 0 {
		return "", fmt.Errorf("invalid user id")
	}
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	exp := now.UTC().Add(EmbedHandoffTTL).Unix()
	payload := fmt.Sprintf("%d.%d.%s", userID, exp, hex.EncodeToString(nonce))
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// ParseEmbedHandoffToken validates a handoff token and returns the user id.
func ParseEmbedHandoffToken(secret, token string, now time.Time) (int64, error) {
	secret = strings.TrimSpace(secret)
	token = strings.TrimSpace(token)
	if secret == "" || token == "" {
		return 0, fmt.Errorf("invalid embed handoff token")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid embed handoff token")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid embed handoff token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid embed handoff token")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payloadBytes)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return 0, fmt.Errorf("invalid embed handoff token")
	}
	fields := strings.Split(string(payloadBytes), ".")
	if len(fields) != 3 {
		return 0, fmt.Errorf("invalid embed handoff token")
	}
	userID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil || userID <= 0 {
		return 0, fmt.Errorf("invalid embed handoff token")
	}
	exp, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid embed handoff token")
	}
	if now.UTC().Unix() > exp {
		return 0, fmt.Errorf("embed handoff token expired")
	}
	return userID, nil
}

// WantsEmbed is true when the request carries embed=1 (query or form).
func WantsEmbed(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.URL.Query().Get("embed") == "1" {
		return true
	}
	return r.FormValue("embed") == "1"
}

// IsEmbedNext reports whether a sanitized next path targets embed UI.
func IsEmbedNext(next string) bool {
	return strings.Contains(next, "embed=1")
}

// RequestIsEmbed is true for Canvas-modal UI (explicit embed=1 or next path with embed=1).
func RequestIsEmbed(r *http.Request) bool {
	if WantsEmbed(r) {
		return true
	}
	if r == nil {
		return false
	}
	if IsEmbedNext(SanitizeNextPath(r.FormValue("next"))) {
		return true
	}
	return IsEmbedNext(SanitizeNextPath(r.URL.Query().Get("next")))
}

// WithEmbedQuery ensures next carries embed=1 for modal chrome.
func WithEmbedQuery(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || IsEmbedNext(next) {
		return next
	}
	if strings.Contains(next, "?") {
		return next + "&embed=1"
	}
	return next + "?embed=1"
}
