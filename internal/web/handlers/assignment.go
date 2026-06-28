package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"rubrical/internal/web/pages"
)

func (h *Handlers) AssignmentDetail(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	assignment, err := h.getAssignment(r.Context(), id)
	if err != nil {
		http.Error(w, "assignment not found", http.StatusNotFound)
		return
	}
	assignment.Embed = r.URL.Query().Get("embed") == "1"

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	if r.URL.Query().Get("embed") == "1" {
		pages.AssignmentEmbed(assignment).Render(r.Context(), w)
		return
	}

	pages.Assignment(assignment).Render(r.Context(), w)
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
