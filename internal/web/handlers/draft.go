package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"rubrical/internal/analysis"
	"rubrical/internal/draftmode"
	"rubrical/internal/drafturl"
	"rubrical/internal/submissiontypes"
	"rubrical/internal/web/pages"
)

func renderHTMXDraftStatusError(w http.ResponseWriter, r *http.Request, assignmentID int64, message string) bool {
	if r.Header.Get("HX-Request") != "true" {
		return false
	}

	w.Header().Set("HX-Retarget", "#"+pages.DraftStatusID(assignmentID))
	w.Header().Set("HX-Reswap", "innerHTML")
	pages.DraftStatusError(message).Render(r.Context(), w)
	return true
}

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
		if r.Header.Get("HX-Request") == "true" {
			pages.DraftStatusText(pages.DraftStatusLabel(0, nil, "", draftmode.Text)).Render(r.Context(), w)
			return
		}
		http.Error(w, "draft is required", http.StatusBadRequest)
		return
	}

	if err := h.upsertLatestDraft(r.Context(), id, draftUpsertOptions{
		Mode:       draftmode.Text,
		Body:       body,
		SourceType: "manual_paste",
	}); err != nil {
		http.Error(w, "failed to save draft", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "draft-saved")
		wordCount := len(strings.Fields(body))
		pages.DraftSaved(pages.DraftSavedMessage(wordCount, "")).Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
}

func (h *Handlers) UploadDraft(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	maxBytes := h.maxDraftUploadBytes()

	if err := r.ParseMultipartForm(int64(maxBytes)); err != nil {
		http.Error(w, "invalid upload", http.StatusBadRequest)
		return
	}

	headers := r.MultipartForm.File["draft_file"]
	if len(headers) == 0 {
		http.Error(w, "draft_file is required", http.StatusBadRequest)
		return
	}

	if allowed, err := h.loadAllowedDraftModes(r.Context(), id); err != nil {
		http.Error(w, "assignment not found", http.StatusNotFound)
		return
	} else if !submissiontypes.ModeAllowed(draftmode.File, allowed) {
		http.Error(w, "file submission not allowed for this assignment", http.StatusBadRequest)
		return
	}

	var uploads []decodedDraftFile
	for _, header := range headers {
		file, err := header.Open()
		if err != nil {
			http.Error(w, "failed to read upload", http.StatusInternalServerError)
			return
		}

		data, err := io.ReadAll(io.LimitReader(file, int64(maxBytes+1)))
		file.Close()
		if err != nil {
			http.Error(w, "failed to read upload", http.StatusInternalServerError)
			return
		}
		if len(data) > maxBytes {
			http.Error(w, "file too large", http.StatusBadRequest)
			return
		}
		if len(data) == 0 {
			continue
		}

		uploads = append(uploads, decodedDraftFile{
			FileName: header.Filename,
			MimeType: header.Header.Get("Content-Type"),
			Data:     data,
		})
	}

	if len(uploads) == 0 {
		http.Error(w, "draft_file is required", http.StatusBadRequest)
		return
	}

	if err := h.appendUploadedDraftFiles(r.Context(), id, uploads); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uploadedNames := make([]string, 0, len(uploads))
	for _, upload := range uploads {
		uploadedNames = append(uploadedNames, upload.FileName)
	}
	h.renderDraftPanel(w, r, id, pages.DraftFilesSavedMessage(uploadedNames))
}

func (h *Handlers) UploadDiscussionAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	pageType, err := h.loadAssignmentPageType(r.Context(), id)
	if err != nil {
		http.Error(w, "assignment not found", http.StatusNotFound)
		return
	}
	if pageType != "discussion" {
		http.Error(w, "discussion attachment upload requires a discussion import", http.StatusBadRequest)
		return
	}

	maxBytes := h.maxDraftUploadBytes()

	if err := r.ParseMultipartForm(int64(maxBytes)); err != nil {
		http.Error(w, "invalid upload", http.StatusBadRequest)
		return
	}

	headers := r.MultipartForm.File["draft_file"]
	if len(headers) == 0 {
		http.Error(w, "draft_file is required", http.StatusBadRequest)
		return
	}
	if len(headers) > 1 {
		http.Error(w, "discussion drafts support at most one attachment", http.StatusBadRequest)
		return
	}

	header := headers[0]
	file, err := header.Open()
	if err != nil {
		http.Error(w, "failed to read upload", http.StatusInternalServerError)
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, int64(maxBytes+1)))
	file.Close()
	if err != nil {
		http.Error(w, "failed to read upload", http.StatusInternalServerError)
		return
	}
	if len(data) > maxBytes {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}
	if len(data) == 0 {
		http.Error(w, "draft_file is required", http.StatusBadRequest)
		return
	}

	upload := decodedDraftFile{
		FileName:     header.Filename,
		MimeType:     header.Header.Get("Content-Type"),
		Data:         data,
		CanvasFileID: strings.TrimSpace(r.FormValue("canvas_file_id")),
	}

	if err := h.attachDiscussionDraftFile(r.Context(), id, upload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		h.renderDraftPanel(w, r, id, pages.DraftFilesSavedMessage([]string{upload.FileName}))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) SaveDraftURL(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	rawURL := strings.TrimSpace(r.FormValue("submission_url"))
	if rawURL == "" {
		if r.Header.Get("HX-Request") == "true" {
			pages.DraftStatusText(pages.DraftStatusLabel(0, nil, "", draftmode.URL)).Render(r.Context(), w)
			return
		}
		http.Error(w, "submission_url is required", http.StatusBadRequest)
		return
	}

	normalizedURL, err := drafturl.ParseSubmissionURL(rawURL)
	if err != nil {
		if r.Header.Get("HX-Request") == "true" {
			pages.DraftStatusError(err.Error()).Render(r.Context(), w)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if allowed, err := h.loadAllowedDraftModes(r.Context(), id); err != nil {
		http.Error(w, "assignment not found", http.StatusNotFound)
		return
	} else if !submissiontypes.ModeAllowed(draftmode.URL, allowed) {
		http.Error(w, "website url submission not allowed for this assignment", http.StatusBadRequest)
		return
	}

	if err := h.saveDraftURL(r.Context(), id, normalizedURL, "manual_website_url", false); err != nil {
		http.Error(w, "failed to save url", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		pages.DraftSaved(pages.DraftLinkSavedMessage()).Render(r.Context(), w)
		return
	}

	h.renderDraftPanel(w, r, id, "")
}

func (h *Handlers) SetDraftMode(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	mode := draftmode.Normalize(r.FormValue("mode"))
	allowed, err := h.loadAllowedDraftModes(r.Context(), id)
	if err != nil {
		http.Error(w, "assignment not found", http.StatusNotFound)
		return
	}
	if !submissiontypes.ModeAllowed(mode, allowed) {
		http.Error(w, "submission type not allowed for this assignment", http.StatusBadRequest)
		return
	}

	if err := h.switchDraftMode(r.Context(), id, mode); err != nil {
		http.Error(w, "failed to switch draft mode", http.StatusInternalServerError)
		return
	}

	h.renderDraftPanel(w, r, id, "")
}

func (h *Handlers) RemoveDraftFile(w http.ResponseWriter, r *http.Request) {
	assignmentID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	fileID, err := parseID(chi.URLParam(r, "fileId"))
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	if err := h.removeDraftFileByID(r.Context(), assignmentID, fileID); err != nil {
		http.Error(w, "failed to remove draft file", http.StatusInternalServerError)
		return
	}

	h.renderDraftPanel(w, r, assignmentID, "")
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
		if err := h.upsertLatestDraft(r.Context(), id, draftUpsertOptions{
			Mode:       draftmode.Text,
			Body:       body,
			SourceType: "manual_paste",
		}); err != nil {
			http.Error(w, "failed to save draft", http.StatusInternalServerError)
			return
		}
	}

	rawURL := strings.TrimSpace(r.FormValue("submission_url"))
	if rawURL != "" {
		normalizedURL, err := drafturl.ParseSubmissionURL(rawURL)
		if err != nil {
			if renderHTMXDraftStatusError(w, r, id, err.Error()) {
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if allowed, err := h.loadAllowedDraftModes(r.Context(), id); err != nil {
			http.Error(w, "assignment not found", http.StatusNotFound)
			return
		} else if !submissiontypes.ModeAllowed(draftmode.URL, allowed) {
			http.Error(w, "website url submission not allowed for this assignment", http.StatusBadRequest)
			return
		}

		if err := h.saveDraftURL(r.Context(), id, normalizedURL, "manual_website_url", false); err != nil {
			http.Error(w, "failed to save url", http.StatusInternalServerError)
			return
		}
	}

	if h.analysis == nil {
		h.renderAnalysisError(w, r, analysis.ErrNotConfigured)
		return
	}

	result, err := h.analysis.Run(r.Context(), id, h.userID)
	if err != nil {
		h.renderAnalysisError(w, r, err)
		return
	}

	h.renderAnalysisResults(w, r, pages.AnalysisResultsFromResult(&result))
}

func (h *Handlers) renderAnalysisError(w http.ResponseWriter, r *http.Request, err error) {
	embed := r.FormValue("embed") == "1"
	message := "Analysis failed. Try again in a moment."
	settingsURL := ""
	switch {
	case h.analysis == nil || errors.Is(err, analysis.ErrNotConfigured):
		message = "Configure AI before analyzing: choose a provider, model, and API key."
		settingsURL = pages.SettingsURL(embed)
	case errors.Is(err, analysis.ErrNothingToAnalyze):
		message = "Add draft text, upload a file, or enter a submission URL before analyzing."
	case errors.Is(err, analysis.ErrNoAnalyzableContent):
		message = "No analyzable submission content. Upload supported files or add draft text."
	default:
		if errors.Is(err, analysis.ErrRateLimited) {
			message = err.Error()
		} else if trimmed := strings.TrimSpace(err.Error()); trimmed != "" {
			message = trimmed
		}
	}
	pages.AnalysisError(message, settingsURL).Render(r.Context(), w)
}

func (h *Handlers) renderAnalysisResults(w http.ResponseWriter, r *http.Request, view pages.AnalysisResultsView) {
	pages.AnalysisResults(view).Render(r.Context(), w)
}

func (h *Handlers) renderDraftPanel(w http.ResponseWriter, r *http.Request, id int64, statusFlash string) {
	if r.Header.Get("HX-Request") == "true" {
		assignment, err := h.getAssignment(r.Context(), id)
		if err != nil {
			http.Error(w, "assignment not found", http.StatusNotFound)
			return
		}
		assignment.DraftStatusFlash = statusFlash
		pages.DraftPanel(assignment).Render(r.Context(), w)
		return
	}

	http.Redirect(w, r, pages.AssignmentURL(id), http.StatusSeeOther)
}

func (h *Handlers) AnalysisResults(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	if h.analysis == nil {
		pages.AnalysisEmpty().Render(r.Context(), w)
		return
	}

	result, err := h.analysis.LoadLatestResult(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to load analysis", http.StatusInternalServerError)
		return
	}
	if result == nil {
		pages.AnalysisEmpty().Render(r.Context(), w)
		return
	}

	h.renderAnalysisResults(w, r, pages.AnalysisResultsFromResult(result))
}

func (h *Handlers) ResolveFeedback(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
