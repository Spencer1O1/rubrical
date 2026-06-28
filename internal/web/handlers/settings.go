package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"rubrical/internal/aisettings"
	"rubrical/internal/web/pages"
)

func (h *Handlers) SettingsPage(w http.ResponseWriter, r *http.Request) {
	settings, err := h.aiSettings.Get(r.Context(), h.userID)
	if err != nil {
		http.Error(w, "failed to load ai settings", http.StatusInternalServerError)
		return
	}
	pages.SettingsPage(settings, "").Render(r.Context(), w)
}

func (h *Handlers) GetAISettingsAPI(w http.ResponseWriter, r *http.Request) {
	settings, err := h.aiSettings.Get(r.Context(), h.userID)
	if err != nil {
		http.Error(w, "failed to load ai settings", http.StatusInternalServerError)
		return
	}
	writeJSON(w, settings)
}

func (h *Handlers) SaveAISettings(w http.ResponseWriter, r *http.Request) {
	var incoming aisettings.Settings
	if isJSONRequest(r) {
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		incoming = aisettings.Settings{
			Provider:        r.FormValue("provider"),
			Model:           r.FormValue("model"),
			OpenAIAPIKey:    r.FormValue("openaiApiKey"),
			AnthropicAPIKey: r.FormValue("anthropicApiKey"),
		}
	}

	saved, err := h.aiSettings.Save(r.Context(), h.userID, incoming)
	if err != nil {
		if isJSONRequest(r) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		current, loadErr := h.aiSettings.Get(r.Context(), h.userID)
		if loadErr != nil {
			current = incoming
		}
		pages.SettingsPage(current, err.Error()).Render(r.Context(), w)
		return
	}

	if isJSONRequest(r) {
		writeJSON(w, saved)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		pages.SettingsSaved().Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func isJSONRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
