package handlers

import (
	"net/http"

	"rubrical/internal/web/components"
	"rubrical/internal/web/pages"
)

func (h *Handlers) Onboarding(w http.ResponseWriter, r *http.Request) {
	user, signedIn := h.currentUser(r)
	nav := components.LayoutUser{SignedIn: signedIn}
	if signedIn {
		nav.Email = user.Email
	}
	pages.Onboarding(nav).Render(r.Context(), w)
}
