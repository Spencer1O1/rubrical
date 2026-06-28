package handlers

import (
	"context"
	"net/http"

	"rubrical/internal/web/pages"
)

func (h *Handlers) renderHTMXDraftPanelError(w http.ResponseWriter, r *http.Request, assignmentID int64, message string) {
	if r.Header.Get("HX-Request") != "true" {
		http.Error(w, message, http.StatusBadRequest)
		return
	}
	h.renderDraftPanelWithError(w, r, assignmentID, message)
}

func (h *Handlers) renderDraftPanelWithError(w http.ResponseWriter, r *http.Request, assignmentID int64, message string) {
	assignment, err := h.getAssignment(r.Context(), assignmentID, requestEmbed(r))
	if err != nil {
		http.Error(w, "assignment not found", http.StatusNotFound)
		return
	}
	assignment.DraftPanelError = message
	pages.DraftPanelBody(assignment).Render(r.Context(), w)
}

func renderHTMXDraftStatusError(w http.ResponseWriter, r *http.Request, assignmentID int64, message string) bool {
	if r.Header.Get("HX-Request") != "true" {
		return false
	}

	w.Header().Set("HX-Retarget", "#"+pages.DraftStatusID(assignmentID))
	w.Header().Set("HX-Reswap", "innerHTML")
	pages.DraftStatusError(message).Render(r.Context(), w)
	return true
}

func renderAnalysisValidationError(w http.ResponseWriter, ctx context.Context, message string, embed bool, assignmentID int64) {
	settingsURL := ""
	if embed {
		settingsURL = pages.SettingsURL(true, assignmentID)
	}
	pages.AnalysisError(message, settingsURL).Render(ctx, w)
}
