package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"rubrical/internal/web/pages"
)

func (h *Handlers) SaveDraft(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	body := strings.TrimSpace(r.FormValue("draft"))
	if body == "" {
		http.Error(w, "draft is required", http.StatusBadRequest)
		return
	}

	if err := h.createDraft(r.Context(), id, body, "manual_paste", false); err != nil {
		http.Error(w, "failed to save draft", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "draft-saved")
		wordCount := len(strings.Fields(body))
		pages.DraftSaved(pages.DraftSavedMessage(wordCount)).Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

func (h *Handlers) AnalyzeDraft(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	body := strings.TrimSpace(r.FormValue("draft"))
	if body != "" {
		if err := h.createDraft(r.Context(), id, body, "manual_paste", false); err != nil {
			http.Error(w, "failed to save draft", http.StatusInternalServerError)
			return
		}
	}

	pages.AnalysisPending().Render(r.Context(), w)
}

func (h *Handlers) AnalysisResults(w http.ResponseWriter, r *http.Request) {
	pages.AnalysisPending().Render(r.Context(), w)
}

func (h *Handlers) ResolveFeedback(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
