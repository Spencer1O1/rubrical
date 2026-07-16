package handlers

import (
	"context"
	"net/http"

	"rubrical/internal/auth"
	"rubrical/internal/web/pages"
)

func userIDFrom(ctx context.Context) (int64, error) {
	return auth.UserID(ctx)
}

func writeAuthSession(w http.ResponseWriter, r *http.Request, authSvc *auth.Service, secure bool, user auth.User) error {
	session, err := authSvc.CreateSession(r.Context(), user.ID)
	if err != nil {
		return err
	}
	// Iframe on Canvas: first-party Lax cookies are not stored/sent — use CHIPS.
	if auth.RequestIsEmbed(r) {
		auth.SetEmbedSessionCookie(w, session.Token, session.ExpiresAt)
		return nil
	}
	auth.SetSessionCookie(w, session.Token, session.ExpiresAt, secure)
	return nil
}

func redirectAfterLogin(w http.ResponseWriter, r *http.Request) {
	next := auth.SanitizeNextPath(r.FormValue("next"))
	if next == "" {
		next = auth.SanitizeNextPath(r.URL.Query().Get("next"))
	}
	if next == "" {
		http.Redirect(w, r, pages.DashboardPath, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}
