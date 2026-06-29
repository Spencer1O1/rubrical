package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"rubrical/internal/aisettings"
	"rubrical/internal/web/pages"
)

func (h *Handlers) SettingsPage(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFrom(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	settings, err := h.aiSettings.Get(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to load ai settings", http.StatusInternalServerError)
		return
	}
	saved := r.URL.Query().Get("saved") == "1"
	if r.URL.Query().Get("embed") == "1" {
		id := assignmentIDFromRequest(r)
		pages.SettingsEmbed(settings, "", assignmentEmbedBackURL(r), id, saved).Render(r.Context(), w)
		return
	}
	pages.SettingsPage(settings, h.layoutUser(r), "", saved).Render(r.Context(), w)
}

func (h *Handlers) GetAISettingsAPI(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFrom(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	settings, err := h.aiSettings.Get(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to load ai settings", http.StatusInternalServerError)
		return
	}
	writeJSON(w, settings.Public())
}

func (h *Handlers) SaveAISettings(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFrom(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	embed := r.FormValue("embed") == "1" || r.URL.Query().Get("embed") == "1"
	assignmentID := assignmentIDFromRequest(r)

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
		embed = embed || r.FormValue("embed") == "1"
		assignmentID = assignmentIDFromRequest(r)
		incoming = aisettings.Settings{
			Provider:        r.FormValue("provider"),
			Model:           r.FormValue("model"),
			OpenAIAPIKey:    r.FormValue("openaiApiKey"),
			AnthropicAPIKey: r.FormValue("anthropicApiKey"),
		}
	}

	saved, err := h.aiSettings.Save(r.Context(), userID, incoming)
	if err != nil {
		if isJSONRequest(r) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			pages.SettingsStatusError(err.Error()).Render(r.Context(), w)
			return
		}
		current, loadErr := h.aiSettings.Get(r.Context(), userID)
		if loadErr != nil {
			current = incoming
		}
		if embed {
			pages.SettingsEmbed(current, err.Error(), assignmentEmbedBackURL(r), assignmentID, false).Render(r.Context(), w)
		} else {
			pages.SettingsPage(current, h.layoutUser(r), err.Error(), false).Render(r.Context(), w)
		}
		return
	}

	if isJSONRequest(r) {
		writeJSON(w, saved.Public())
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		pages.SettingsSaved().Render(r.Context(), w)
		return
	}

	if embed {
		http.Redirect(w, r, pages.SettingsSavedRedirectURL(true, assignmentID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, pages.SettingsSavedRedirectURL(false, 0), http.StatusSeeOther)
}

func assignmentIDFromRequest(r *http.Request) int64 {
	if r == nil {
		return 0
	}
	if raw := strings.TrimSpace(r.FormValue("assignment_id")); raw != "" {
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
			return id
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("assignment_id")); raw != "" {
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
			return id
		}
	}
	return assignmentIDFromReferer(r.Header.Get("Referer"))
}

func assignmentIDFromReferer(referer string) int64 {
	referer = strings.TrimSpace(referer)
	if referer == "" {
		return 0
	}
	parsed, err := url.Parse(referer)
	if err != nil {
		return 0
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "assignments" {
		return 0
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func assignmentEmbedBackURL(r *http.Request) string {
	if id := assignmentIDFromRequest(r); id > 0 {
		return pages.AssignmentEmbedURL(id)
	}
	return "/"
}

func isJSONRequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
