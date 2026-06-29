package handlers

import (
	"net/http"

	"rubrical/internal/web/components"
	"rubrical/internal/web/pages"
)

func (h *Handlers) Landing(w http.ResponseWriter, r *http.Request) {
	user, signedIn := h.currentUser(r)
	nav := components.MarketingNav{SignedIn: signedIn}
	if signedIn {
		nav.Email = user.Email
	}
	pages.Landing(nav).Render(r.Context(), w)
}
