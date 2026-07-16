package handlers

import (
	"net/http"
	"os"

	"rubrical/internal/web/components"
	"rubrical/internal/web/pages"
)

const extensionZipPath = "static/downloads/rubrical-extension.zip"

func (h *Handlers) Install(w http.ResponseWriter, r *http.Request) {
	user, signedIn := h.currentUser(r)
	nav := components.MarketingNav{SignedIn: signedIn}
	if signedIn {
		nav.Email = user.Email
	}
	_, err := os.Stat(extensionZipPath)
	pages.Install(nav, err == nil).Render(r.Context(), w)
}
