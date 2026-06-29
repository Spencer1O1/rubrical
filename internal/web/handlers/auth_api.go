package handlers

import (
	"encoding/json"
	"net/http"

	"rubrical/internal/auth"
	"rubrical/internal/email"
)

type authCredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authSignupRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type authForgotRequest struct {
	Email string `json:"email"`
}

type authUserResponse struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type authConfigResponse struct {
	GoogleEnabled    bool `json:"googleEnabled"`
	StrictExtraction bool `json:"strictExtraction"`
}

func (h *Handlers) AuthConfigAPI(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, authConfigResponse{
		GoogleEnabled:    h.google.Enabled(),
		StrictExtraction: h.strictExtraction,
	})
}

func (h *Handlers) LoginAPI(w http.ResponseWriter, r *http.Request) {
	var req authCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	user, err := h.auth.AuthenticatePassword(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	if err := writeAuthSession(w, r, h.auth, h.authSecure, user); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	writeJSON(w, authUserResponse{Email: user.Email, DisplayName: user.DisplayName})
}

func (h *Handlers) SignupAPI(w http.ResponseWriter, r *http.Request) {
	var req authSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	user, err := h.auth.CreateUserWithPassword(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		msg := err.Error()
		if err == auth.ErrEmailTaken {
			msg = "An account with that email already exists."
		}
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if err := writeAuthSession(w, r, h.auth, h.authSecure, user); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}
	writeJSON(w, authUserResponse{Email: user.Email, DisplayName: user.DisplayName})
}

func (h *Handlers) ForgotPasswordAPI(w http.ResponseWriter, r *http.Request) {
	var req authForgotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	token, user, err := h.auth.CreatePasswordResetToken(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "failed to process request", http.StatusInternalServerError)
		return
	}
	if token != "" && user.ID > 0 && h.mailer != nil {
		msg := email.PasswordResetMessage(h.publicURL, token)
		msg.To = user.Email
		_ = h.mailer.Send(r.Context(), msg)
	}
	writeJSON(w, map[string]string{
		"message": "If an account exists for that email, a reset link has been sent.",
	})
}
