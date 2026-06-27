package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":            "ok",
		"strictExtraction": h.strictExtraction,
	})
}
