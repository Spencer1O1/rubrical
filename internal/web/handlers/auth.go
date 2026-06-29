package handlers

import (
	"net/http"

	"rubrical/internal/auth"
	"rubrical/internal/email"
	"rubrical/internal/web/components"
	"rubrical/internal/web/pages"
)

func (h *Handlers) AuthPage(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.UserFromContext(r.Context()); err == nil {
		http.Redirect(w, r, pages.DashboardPath, http.StatusSeeOther)
		return
	}
	h.renderAuthPage(w, r, pages.AuthFormView{
		Mode:          pages.ParseAuthMode(r.URL.Query().Get("mode")),
		Next:          auth.SanitizeNextPath(r.URL.Query().Get("next")),
		GoogleEnabled: h.google.Enabled(),
		ErrorMessage:  authPageErrorMessage(r),
		SuccessMessage: authPageSuccessMessage(r),
	})
}

func (h *Handlers) ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	h.renderAuthPage(w, r, pages.AuthFormView{
		Mode:       pages.AuthModeReset,
		ResetToken: r.URL.Query().Get("token"),
	})
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	user, err := h.auth.AuthenticatePassword(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		h.renderAuthPage(w, r, pages.AuthFormView{
			Mode:          pages.AuthModeLogin,
			Next:          auth.SanitizeNextPath(r.FormValue("next")),
			GoogleEnabled: h.google.Enabled(),
			ErrorMessage:  "Invalid email or password.",
		})
		return
	}
	if err := writeAuthSession(w, r, h.auth, h.authSecure, user); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	redirectAfterLogin(w, r)
}

func (h *Handlers) Signup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	user, err := h.auth.CreateUserWithPassword(r.Context(), r.FormValue("email"), r.FormValue("password"), r.FormValue("displayName"))
	if err != nil {
		msg := err.Error()
		if err == auth.ErrEmailTaken {
			msg = "An account with that email already exists."
		}
		h.renderAuthPage(w, r, pages.AuthFormView{
			Mode:          pages.AuthModeSignup,
			Next:          auth.SanitizeNextPath(r.FormValue("next")),
			GoogleEnabled: h.google.Enabled(),
			ErrorMessage:  msg,
		})
		return
	}
	if err := writeAuthSession(w, r, h.auth, h.authSecure, user); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	redirectAfterLogin(w, r)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	token := auth.SessionTokenFromRequest(r)
	_ = h.auth.RevokeSession(r.Context(), token)
	auth.ClearSessionCookie(w, h.authSecure)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	token, user, err := h.auth.CreatePasswordResetToken(r.Context(), r.FormValue("email"))
	if err != nil {
		http.Error(w, "failed to process request", http.StatusInternalServerError)
		return
	}
	if token != "" && user.ID > 0 && h.mailer != nil {
		msg := email.PasswordResetMessage(h.publicURL, token)
		msg.To = user.Email
		_ = h.mailer.Send(r.Context(), msg)
	}
	h.renderAuthPage(w, r, pages.AuthFormView{
		Mode:           pages.AuthModeForgot,
		SuccessMessage: "If an account exists for that email, a reset link has been sent.",
	})
}

func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	token := r.FormValue("token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if err := h.auth.ResetPassword(r.Context(), token, r.FormValue("password")); err != nil {
		h.renderAuthPage(w, r, pages.AuthFormView{
			Mode:         pages.AuthModeReset,
			ResetToken:   token,
			ErrorMessage: "Invalid or expired reset link.",
		})
		return
	}
	http.Redirect(w, r, "/login?saved=1", http.StatusSeeOther)
}

func (h *Handlers) GoogleAuthStart(w http.ResponseWriter, r *http.Request) {
	if !h.google.Enabled() {
		http.Error(w, "google sign-in is not configured", http.StatusNotFound)
		return
	}
	state, err := auth.NewOAuthState()
	if err != nil {
		http.Error(w, "failed to start google sign-in", http.StatusInternalServerError)
		return
	}
	auth.SetOAuthStateCookie(w, state, h.authSecure)
	next := auth.SanitizeNextPath(r.URL.Query().Get("next"))
	if next != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "rubrical_oauth_next",
			Value:    next,
			Path:     "/",
			HttpOnly: true,
			Secure:   h.authSecure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   600,
		})
	}
	http.Redirect(w, r, h.google.AuthCodeURL(state), http.StatusSeeOther)
}

func (h *Handlers) GoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if !h.google.Enabled() {
		http.Error(w, "google sign-in is not configured", http.StatusNotFound)
		return
	}
	state := auth.OAuthStateFromRequest(r)
	auth.ClearOAuthStateCookie(w, h.authSecure)
	if state == "" || state != r.URL.Query().Get("state") {
		http.Redirect(w, r, "/login?error=oauth_state", http.StatusSeeOther)
		return
	}
	profile, err := h.google.ExchangeProfile(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		http.Redirect(w, r, "/login?error=oauth_exchange", http.StatusSeeOther)
		return
	}
	user, err := h.auth.FindOrCreateGoogleUser(r.Context(), profile.Sub, profile.Email, profile.Name, profile.EmailVerified)
	if err != nil {
		http.Redirect(w, r, "/login?error=oauth_user", http.StatusSeeOther)
		return
	}
	if err := writeAuthSession(w, r, h.auth, h.authSecure, user); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	next := ""
	if cookie, err := r.Cookie("rubrical_oauth_next"); err == nil {
		next = auth.SanitizeNextPath(cookie.Value)
		http.SetCookie(w, &http.Cookie{Name: "rubrical_oauth_next", Value: "", Path: "/", MaxAge: -1})
	}
	if next == "" {
		http.Redirect(w, r, pages.DashboardPath, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (h *Handlers) SessionAPI(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromContext(r.Context())
	if err != nil {
		token := auth.SessionTokenFromRequest(r)
		if token != "" {
			user, err = h.auth.UserFromSessionToken(r.Context(), token)
		}
	}
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	writeJSON(w, authUserResponse{Email: user.Email, DisplayName: user.DisplayName})
}

func (h *Handlers) renderAuthPage(w http.ResponseWriter, r *http.Request, view pages.AuthFormView) {
	if view.Mode == "" {
		view.Mode = pages.AuthModeLogin
	}
	pages.AuthPage(view).Render(r.Context(), w)
}

func authPageErrorMessage(r *http.Request) string {
	switch r.URL.Query().Get("error") {
	case "oauth_state":
		return "Google sign-in expired. Try again."
	case "oauth_exchange":
		return "Google sign-in failed. Try again."
	case "oauth_user":
		return "Could not sign in with Google."
	default:
		return r.URL.Query().Get("error")
	}
}

func authPageSuccessMessage(r *http.Request) string {
	if r.URL.Query().Get("saved") == "1" {
		return "Password updated. Sign in with your new password."
	}
	return ""
}

func (h *Handlers) currentUser(r *http.Request) (auth.User, bool) {
	user, err := auth.UserFromContext(r.Context())
	return user, err == nil
}

func (h *Handlers) layoutUser(r *http.Request) components.LayoutUser {
	user, ok := h.currentUser(r)
	if !ok {
		return components.LayoutUser{}
	}
	return components.LayoutUser{Email: user.Email, SignedIn: true}
}
