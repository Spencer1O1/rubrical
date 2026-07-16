package handlers

import (
	"fmt"
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
	info, err := os.Stat(extensionZipPath)
	if err != nil {
		pages.Install(nav, false, "").Render(r.Context(), w)
		return
	}
	downloadURL := fmt.Sprintf("/install/rubrical-extension.zip?v=%d", info.ModTime().Unix())
	pages.Install(nav, true, downloadURL).Render(r.Context(), w)
}

// ExtensionZip serves the packaged extension with no-store so Cloudflare/browsers
// do not keep serving a stale zip after deploy.
func (h *Handlers) ExtensionZip(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(extensionZipPath); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="rubrical-extension.zip"`)
	http.ServeFile(w, r, extensionZipPath)
}
