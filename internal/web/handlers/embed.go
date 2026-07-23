package handlers

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"rubrical/internal/auth"
)

// EmbedURL (GET /auth/embed-url) — step 1: extension SW, first-party session → handoff URL.
func (h *Handlers) EmbedURL(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFrom(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if !h.embedAllowed {
		http.Error(w, "embed handoff requires HTTPS public URL (or http://localhost for local dev)", http.StatusBadRequest)
		return
	}
	next := auth.WithEmbedQuery(auth.SanitizeNextPath(r.URL.Query().Get("next")))
	if !validEmbedNext(next) {
		http.Error(w, "invalid next path", http.StatusBadRequest)
		return
	}
	token, err := auth.MintEmbedHandoffToken(h.embedSecret, userID, time.Now())
	if err != nil {
		http.Error(w, "failed to mint embed handoff", http.StatusInternalServerError)
		return
	}
	handoff, err := url.Parse(strings.TrimRight(h.publicURL, "/") + "/auth/embed")
	if err != nil {
		http.Error(w, "invalid public URL", http.StatusInternalServerError)
		return
	}
	q := handoff.Query()
	q.Set("token", token)
	q.Set("next", next)
	handoff.RawQuery = q.Encode()
	writeJSON(w, map[string]string{"url": handoff.String()})
}

// EmbedHandoff (GET /auth/embed) — step 2: iframe first load → CHIPS cookie → redirect to next.
func (h *Handlers) EmbedHandoff(w http.ResponseWriter, r *http.Request) {
	if !h.embedAllowed {
		http.Error(w, "embed handoff requires HTTPS (or http://localhost for local dev)", http.StatusBadRequest)
		return
	}
	userID, err := auth.ParseEmbedHandoffToken(h.embedSecret, r.URL.Query().Get("token"), time.Now())
	if err != nil {
		http.Redirect(w, r, "/login?error=embed_handoff&embed=1", http.StatusSeeOther)
		return
	}
	next := auth.WithEmbedQuery(auth.SanitizeNextPath(r.URL.Query().Get("next")))
	if !validEmbedNext(next) {
		http.Redirect(w, r, "/login?error=embed_handoff&embed=1", http.StatusSeeOther)
		return
	}
	session, err := h.auth.CreateSession(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	auth.SetEmbedSessionCookie(w, session.Token, session.ExpiresAt)
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func validEmbedNext(next string) bool {
	if next == "" {
		return false
	}
	path := next
	if i := strings.IndexByte(next, '?'); i >= 0 {
		path = next[:i]
	}
	return strings.HasPrefix(path, "/assignments/") || path == "/settings" || strings.HasPrefix(path, "/settings?")
}

func requestEmbed(r *http.Request) bool {
	return auth.WantsEmbed(r)
}
