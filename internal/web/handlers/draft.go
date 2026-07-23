package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"rubrical/internal/analysispipeline"
	"rubrical/internal/draftmode"
	"rubrical/internal/drafturl"
	"rubrical/internal/submissiontypes"
	"rubrical/internal/urlfetch"
	"rubrical/internal/web/pages"
)

func (h *Handlers) SaveDraft(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		if renderHTMXDraftStatusError(w, r, id, "invalid form") {
			return
		}
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	body := pages.SanitizedDraftHTML(r.FormValue("draft"))
	if pages.DraftPlainText(body) == "" {
		if err := h.clearDraftTextBody(r.Context(), id); err != nil {
			if renderHTMXDraftStatusError(w, r, id, "failed to clear draft") {
				return
			}
			http.Error(w, "failed to clear draft", http.StatusInternalServerError)
			return
		}
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
		if renderHTMXDraftStatusError(w, r, id, "failed to save draft") {
			return
		}
		http.Error(w, "failed to save draft", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		pages.DraftSaved(pages.DraftSavedMessage(pages.DraftWordCount(body), "")).Render(r.Context(), w)
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
		h.renderHTMXDraftPanelError(w, r, id, "invalid upload")
		return
	}

	headers := r.MultipartForm.File["draft_file"]
	if len(headers) == 0 {
		h.renderHTMXDraftPanelError(w, r, id, "draft_file is required")
		return
	}

	if allowed, err := h.loadAllowedDraftModes(r.Context(), id); err != nil {
		h.renderHTMXDraftPanelError(w, r, id, "assignment not found")
		return
	} else if !submissiontypes.ModeAllowed(draftmode.File, allowed) {
		h.renderHTMXDraftPanelError(w, r, id, "file submission not allowed for this assignment")
		return
	}

	var uploads []decodedDraftFile
	var skippedEmpty int
	canvasFileID := strings.TrimSpace(r.FormValue("canvas_file_id"))
	for _, header := range headers {
		file, err := header.Open()
		if err != nil {
			h.renderHTMXDraftPanelError(w, r, id, "failed to read upload")
			return
		}

		data, err := io.ReadAll(io.LimitReader(file, int64(maxBytes+1)))
		file.Close()
		if err != nil {
			h.renderHTMXDraftPanelError(w, r, id, "failed to read upload")
			return
		}
		if len(data) > maxBytes {
			h.renderHTMXDraftPanelError(w, r, id, "file too large")
			return
		}
		if len(data) == 0 {
			skippedEmpty++
			continue
		}

		uploads = append(uploads, decodedDraftFile{
			FileName:     header.Filename,
			MimeType:     header.Header.Get("Content-Type"),
			Data:         data,
			CanvasFileID: canvasFileID,
		})
	}

	if len(uploads) == 0 {
		if skippedEmpty > 0 {
			h.renderHTMXDraftPanelError(w, r, id, "all selected files are empty")
			return
		}
		h.renderHTMXDraftPanelError(w, r, id, "draft_file is required")
		return
	}

	if err := h.appendUploadedDraftFiles(r.Context(), id, uploads); err != nil {
		h.renderHTMXDraftPanelError(w, r, id, err.Error())
		return
	}

	uploadedNames := make([]string, 0, len(uploads))
	for _, upload := range uploads {
		uploadedNames = append(uploadedNames, upload.FileName)
	}
	flash := pages.DraftFilesSavedMessage(uploadedNames)
	if skippedEmpty > 0 {
		flash = pages.DraftFilesSavedWithSkippedEmpty(flash, skippedEmpty)
	}
	h.renderDraftPanel(w, r, id, flash)
}

func (h *Handlers) UploadDiscussionAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}

	pageType, err := h.loadAssignmentPageType(r.Context(), id)
	if err != nil {
		h.renderHTMXDraftPanelError(w, r, id, "assignment not found")
		return
	}
	if pageType != "discussion" {
		h.renderHTMXDraftPanelError(w, r, id, "discussion attachment upload requires a discussion import")
		return
	}

	maxBytes := h.maxDraftUploadBytes()

	if err := r.ParseMultipartForm(int64(maxBytes)); err != nil {
		h.renderHTMXDraftPanelError(w, r, id, "invalid upload")
		return
	}

	headers := r.MultipartForm.File["draft_file"]
	if len(headers) == 0 {
		h.renderHTMXDraftPanelError(w, r, id, "draft_file is required")
		return
	}
	if len(headers) > 1 {
		h.renderHTMXDraftPanelError(w, r, id, "discussion drafts support at most one attachment")
		return
	}

	header := headers[0]
	file, err := header.Open()
	if err != nil {
		h.renderHTMXDraftPanelError(w, r, id, "failed to read upload")
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, int64(maxBytes+1)))
	file.Close()
	if err != nil {
		h.renderHTMXDraftPanelError(w, r, id, "failed to read upload")
		return
	}
	if len(data) > maxBytes {
		h.renderHTMXDraftPanelError(w, r, id, "file too large")
		return
	}
	if len(data) == 0 {
		h.renderHTMXDraftPanelError(w, r, id, "draft_file is required")
		return
	}

	upload := decodedDraftFile{
		FileName:     header.Filename,
		MimeType:     header.Header.Get("Content-Type"),
		Data:         data,
		CanvasFileID: strings.TrimSpace(r.FormValue("canvas_file_id")),
	}

	if err := h.attachDiscussionDraftFile(r.Context(), id, upload); err != nil {
		h.renderHTMXDraftPanelError(w, r, id, err.Error())
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
		if renderHTMXDraftStatusError(w, r, id, "invalid form") {
			return
		}
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	rawURL := strings.TrimSpace(r.FormValue("submission_url"))
	if rawURL == "" {
		if err := h.clearDraftURL(r.Context(), id); err != nil {
			if renderHTMXDraftStatusError(w, r, id, "failed to clear url") {
				return
			}
			http.Error(w, "failed to clear url", http.StatusInternalServerError)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			pages.DraftStatusText(pages.DraftStatusLabel(0, nil, "", draftmode.URL)).Render(r.Context(), w)
			return
		}
		http.Error(w, "submission_url is required", http.StatusBadRequest)
		return
	}

	normalizedURL, err := drafturl.ParseSubmissionURL(rawURL)
	if err != nil {
		if renderHTMXDraftStatusError(w, r, id, err.Error()) {
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if allowed, err := h.loadAllowedDraftModes(r.Context(), id); err != nil {
		if renderHTMXDraftStatusError(w, r, id, "assignment not found") {
			return
		}
		http.Error(w, "assignment not found", http.StatusNotFound)
		return
	} else if !submissiontypes.ModeAllowed(draftmode.URL, allowed) {
		if renderHTMXDraftStatusError(w, r, id, "website url submission not allowed for this assignment") {
			return
		}
		http.Error(w, "website url submission not allowed for this assignment", http.StatusBadRequest)
		return
	}

	if err := h.saveDraftURL(r.Context(), id, normalizedURL, "manual_website_url", false); err != nil {
		if renderHTMXDraftStatusError(w, r, id, "failed to save url") {
			return
		}
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

	panelError := strings.TrimSpace(r.FormValue("draft_panel_error"))

	mode := draftmode.Normalize(r.FormValue("mode"))
	allowed, err := h.loadAllowedDraftModes(r.Context(), id)
	if err != nil {
		h.renderHTMXDraftPanelError(w, r, id, "assignment not found")
		return
	}
	if !submissiontypes.ModeAllowed(mode, allowed) {
		h.renderHTMXDraftPanelError(w, r, id, "submission type not allowed for this assignment")
		return
	}

	if err := h.switchDraftMode(r.Context(), id, mode); err != nil {
		h.renderHTMXDraftPanelError(w, r, id, "failed to switch draft mode")
		return
	}

	if panelError != "" {
		h.renderDraftPanelWithError(w, r, id, panelError)
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
		h.renderHTMXDraftPanelError(w, r, assignmentID, "failed to remove draft file")
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

	userID, err := userIDFrom(r.Context())
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	embed := requestEmbed(r)

	draft, err := h.loadLatestDraftRow(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to load draft", http.StatusInternalServerError)
		return
	}

	mode := draftmode.Text
	if draft != nil {
		mode = draftmode.Normalize(draft.Mode)
	} else {
		allowed, err := h.loadAllowedDraftModes(r.Context(), id)
		if err != nil {
			http.Error(w, "assignment not found", http.StatusNotFound)
			return
		}
		mode = draftmode.Normalize(pages.DefaultDraftMode(allowed))
	}

	if err := h.persistDraftFromAnalyzeForm(r.Context(), id, r, mode); err != nil {
		if draftmode.Normalize(mode) == draftmode.URL {
			renderAnalysisValidationError(w, r.Context(), err.Error(), embed, id)
			return
		}
		http.Error(w, "failed to save draft before analyze", http.StatusInternalServerError)
		return
	}

	if h.analysis == nil {
		h.renderAnalysisError(w, r, analysispipeline.ErrNotConfigured)
		return
	}

	result, err := h.analysis.Run(r.Context(), id, userID)
	if err != nil {
		h.renderAnalysisError(w, r, err)
		return
	}

	h.renderAnalysisResults(w, r, h.analysisResultsView(r.Context(), id, &result))
}

func (h *Handlers) persistDraftFromAnalyzeForm(ctx context.Context, assignmentID int64, r *http.Request, mode string) error {
	switch draftmode.Normalize(mode) {
	case draftmode.Text:
		if _, ok := r.PostForm["draft"]; !ok {
			return nil
		}
		body := pages.SanitizedDraftHTML(r.FormValue("draft"))
		if pages.DraftPlainText(body) == "" {
			return h.clearDraftTextBody(ctx, assignmentID)
		}
		return h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
			Mode:       draftmode.Text,
			Body:       body,
			SourceType: "manual_paste",
		})
	case draftmode.URL:
		if _, ok := r.PostForm["submission_url"]; !ok {
			return nil
		}
		rawURL := strings.TrimSpace(r.FormValue("submission_url"))
		if rawURL == "" {
			return h.clearDraftURL(ctx, assignmentID)
		}
		normalizedURL, err := drafturl.ParseSubmissionURL(rawURL)
		if err != nil {
			return err
		}
		allowed, err := h.loadAllowedDraftModes(ctx, assignmentID)
		if err != nil {
			return err
		}
		if !submissiontypes.ModeAllowed(draftmode.URL, allowed) {
			return fmt.Errorf("website url submission not allowed for this assignment")
		}
		return h.saveDraftURL(ctx, assignmentID, normalizedURL, "manual_website_url", false)
	default:
		return nil
	}
}

func (h *Handlers) renderAnalysisError(w http.ResponseWriter, r *http.Request, err error) {
	embed := requestEmbed(r)
	assignmentID, _ := parseID(chi.URLParam(r, "id"))
	message := "Analysis failed. Try again in a moment."
	settingsURL := ""
	switch {
	case h.analysis == nil || errors.Is(err, analysispipeline.ErrNotConfigured):
		message = "Configure AI before analyzing: choose a provider, model, and API key."
		settingsURL = pages.SettingsURL(embed, assignmentID)
	case errors.Is(err, analysispipeline.ErrNothingToAnalyze):
		message = "Add draft text, upload a file, or enter a submission URL before analyzing."
	case errors.Is(err, analysispipeline.ErrNoAnalyzableContent):
		message = "No analyzable submission content. Upload supported files or add draft text."
	case errors.Is(err, urlfetch.ErrNonHTMLContent):
		message = "The submission URL must point to an HTML page."
	case errors.Is(err, analysispipeline.ErrURLFetchFailed):
		message = "Could not fetch content from the submission URL. Check the link and try again."
	case errors.Is(err, analysispipeline.ErrAnalysisInFlight):
		message = "An analysis is already running for this assignment. Wait for it to finish."
	case errors.Is(err, analysispipeline.ErrFeedbackPersistFailed):
		message = "Analysis finished but feedback could not be saved. Try again in a moment."
	default:
		if errors.Is(err, analysispipeline.ErrRateLimited) {
			message = err.Error()
		} else if trimmed := strings.TrimSpace(err.Error()); trimmed != "" {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "openai api error") || strings.Contains(lower, "anthropic api error") {
				message = "The AI provider returned an error. Check your API key and settings, then try again."
			} else {
				message = trimmed
			}
		}
	}
	pages.AnalysisError(message, settingsURL).Render(r.Context(), w)
}

func (h *Handlers) renderAnalysisResults(w http.ResponseWriter, r *http.Request, view pages.AnalysisResultsView) {
	pages.AnalysisResults(view).Render(r.Context(), w)
}

func (h *Handlers) renderDraftPanel(w http.ResponseWriter, r *http.Request, id int64, statusFlash string) {
	if r.Header.Get("HX-Request") == "true" {
		assignment, err := h.getAssignment(r.Context(), id, requestEmbed(r))
		if err != nil {
			http.Error(w, "assignment not found", http.StatusNotFound)
			return
		}
		assignment.DraftStatusFlash = statusFlash
		pages.DraftPanelBody(assignment).Render(r.Context(), w)
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

	if _, err := userIDFrom(r.Context()); err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	if _, err := h.getAssignment(r.Context(), id, false); err != nil {
		http.Error(w, "assignment not found", http.StatusNotFound)
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

	h.renderAnalysisResults(w, r, h.analysisResultsView(r.Context(), id, result))
}
