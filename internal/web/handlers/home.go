package handlers

import (
	"net/http"

	"rubrical/internal/web/pages"
)

// Home serves the dashboard at "/" for signed-in users, otherwise redirects to onboarding.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.currentUser(r); !ok {
		http.Redirect(w, r, pages.OnboardingPath, http.StatusSeeOther)
		return
	}
	h.Dashboard(w, r)
}
