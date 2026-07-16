package auth

import (
	"net/http"
	"net/url"
	"strings"

	"rubrical/internal/config"
)

type Middleware struct {
	auth   *Service
	secure bool
}

func NewMiddleware(auth *Service, cfg config.Config) *Middleware {
	secure := cfg.CookieSecure()
	return &Middleware{auth: auth, secure: secure}
}

func (m *Middleware) OptionalUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := SessionTokenFromRequest(r)
		if token != "" {
			if user, err := m.auth.UserFromSessionToken(r.Context(), token); err == nil {
				r = r.WithContext(WithUser(r.Context(), user))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := SessionTokenFromRequest(r)
		user, err := m.auth.UserFromSessionToken(r.Context(), token)
		if err != nil {
			m.unauthorized(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	})
}

func (m *Middleware) unauthorized(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	next := SanitizeNextPath(r.URL.Query().Get("next"))
	if next == "" && r.URL.Path != "/" {
		next = r.URL.Path
		if r.URL.Query().Get("embed") == "1" {
			next += "?embed=1"
		}
	}
	http.Redirect(w, r, loginRedirectURL(next), http.StatusSeeOther)
}

func wantsJSON(r *http.Request) bool {
	if r == nil {
		return false
	}
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		return true
	}
	if r.URL.Query().Get("format") == "json" {
		return true
	}
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return true
	}
	if r.Header.Get("HX-Request") == "true" {
		return false
	}
	return false
}

func loginRedirectURL(next string) string {
	if next == "" {
		return "/login"
	}
	u := "/login?next=" + url.QueryEscape(next)
	if IsEmbedNext(next) {
		u += "&embed=1"
	}
	return u
}

func SanitizeNextPath(next string) string {
	next = strings.TrimSpace(next)
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return ""
	}
	// Allow relative paths with query (e.g. /assignments/1?embed=1); reject absolute URLs.
	if strings.Contains(next, "://") {
		return ""
	}
	return next
}

func (m *Middleware) SecureCookies() bool {
	return m.secure
}
