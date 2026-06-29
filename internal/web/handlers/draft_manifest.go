package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"rubrical/internal/importurl"
)

type draftManifestFile struct {
	ServerFileID int64     `json:"serverFileId"`
	FileName     string    `json:"fileName"`
	CanvasFileID string    `json:"canvasFileId,omitempty"`
	ByteSize     int64     `json:"byteSize"`
	UploadedAt   time.Time `json:"uploadedAt"`
}

type draftManifestResponse struct {
	AssignmentID int64               `json:"assignmentId,omitempty"`
	Files        []draftManifestFile `json:"files"`
}

func (h *Handlers) DraftManifest(w http.ResponseWriter, r *http.Request) {
	sourceURL, err := importurl.ValidateSourceURL(r.URL.Query().Get("sourceUrl"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	userID, err := userIDFrom(ctx)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	var assignmentID int64
	err = h.db.Pool.QueryRow(ctx, `
		SELECT id FROM assignment_snapshots
		WHERE user_id = $1 AND source_url = $2
	`, userID, sourceURL).Scan(&assignmentID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeDraftManifest(w, draftManifestResponse{Files: []draftManifestFile{}})
		return
	}
	if err != nil {
		http.Error(w, "failed to load assignment", http.StatusInternalServerError)
		return
	}

	draft, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		http.Error(w, "failed to load draft", http.StatusInternalServerError)
		return
	}
	if draft == nil {
		writeDraftManifest(w, draftManifestResponse{AssignmentID: assignmentID, Files: []draftManifestFile{}})
		return
	}

	rows, err := h.loadDraftFiles(ctx, draft.ID)
	if err != nil {
		http.Error(w, "failed to load draft files", http.StatusInternalServerError)
		return
	}

	files := make([]draftManifestFile, 0, len(rows))
	for _, row := range rows {
		files = append(files, draftManifestFile{
			ServerFileID: row.ID,
			FileName:     row.FileName,
			CanvasFileID: row.CanvasFileID,
			ByteSize:     row.ByteSize,
			UploadedAt:   row.UploadedAt,
		})
	}

	writeDraftManifest(w, draftManifestResponse{
		AssignmentID: assignmentID,
		Files:        files,
	})
}

func writeDraftManifest(w http.ResponseWriter, payload draftManifestResponse) {
	if payload.Files == nil {
		payload.Files = []draftManifestFile{}
	}
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
