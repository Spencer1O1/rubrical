package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"rubrical/internal/draftfiles"
	"rubrical/internal/draftmode"
	"rubrical/internal/importpayload"
	"rubrical/internal/web/pages"
)

type draftStoredFile struct {
	StorageKey   string
	FileName     string
	MimeType     string
	ByteSize     int64
	UploadedAt   time.Time
	CanvasFileID string
}

type draftFileRow struct {
	ID           int64
	FileName     string
	StorageKey   string
	MimeType     string
	ByteSize     int64
	UploadedAt   time.Time
	SortOrder    int
	CanvasFileID string
}

type draftUpsertOptions struct {
	Mode       string
	Body       string
	URL        string
	SourceType string
	FromCanvas bool
	ClearText  bool
	ClearURL   bool
}

type draftRow struct {
	ID         int64
	Mode       string
	Body       string
	URL        *string
	SourceType string
	FromCanvas bool
}

type decodedDraftFile struct {
	FileName     string
	MimeType     string
	Data         []byte
	CanvasFileID string
}

func (h *Handlers) saveDraftFromImport(ctx context.Context, assignmentID int64, payload importpayload.Payload) error {
	if payload.PageType == "discussion" {
		return h.saveDiscussionDraftFromImport(ctx, assignmentID, payload)
	}

	text := strings.TrimSpace(payload.DraftText)
	url := strings.TrimSpace(payload.DraftURL)
	importFiles, err := h.decodeImportDraftFiles(payload.DraftFiles)
	if err != nil {
		return err
	}

	kind := draftmode.Infer(payload.DraftKind, text, url, len(importFiles) > 0)

	switch kind {
	case draftmode.File:
		if len(importFiles) > 0 || len(payload.DraftFileRefs) > 0 {
			merged, err := h.mergeImportDraftFiles(ctx, assignmentID, importFiles, payload.DraftFileRefs)
			if err != nil {
				return err
			}
			return h.replaceDraftFilesMerged(ctx, assignmentID, merged, "canvas_file_upload", true)
		}

		return h.switchDraftModeOnly(ctx, assignmentID, draftmode.File)

	case draftmode.URL:
		if url != "" {
			return h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
				Mode:       draftmode.URL,
				URL:        url,
				SourceType: "canvas_website_url",
				FromCanvas: true,
			})
		}

		return h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
			Mode:       draftmode.URL,
			URL:        "",
			SourceType: canvasDraftSourceType(payload),
			FromCanvas: true,
		})

	default:
		if text != "" {
			return h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
				Mode:       draftmode.Text,
				Body:       pages.SanitizedDraftHTML(text),
				SourceType: canvasDraftSourceType(payload),
				FromCanvas: true,
			})
		}

		return h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
			Mode:       draftmode.Text,
			Body:       "",
			SourceType: canvasDraftSourceType(payload),
			FromCanvas: true,
		})
	}
}

func (h *Handlers) saveDiscussionDraftFromImport(ctx context.Context, assignmentID int64, payload importpayload.Payload) error {
	text := strings.TrimSpace(payload.DraftText)
	importFiles, err := h.decodeImportDraftFiles(payload.DraftFiles)
	if err != nil {
		return err
	}
	if len(importFiles) > 1 {
		return fmt.Errorf("discussion drafts support at most one attachment")
	}

	if err := h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
		Mode:       draftmode.Text,
		Body:       pages.SanitizedDraftHTML(text),
		SourceType: "canvas_discussion_reply",
		FromCanvas: true,
	}); err != nil {
		return err
	}

	draft, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		return err
	}
	if draft == nil {
		return fmt.Errorf("discussion draft row missing after upsert")
	}

	if len(importFiles) == 0 && len(payload.DraftFileRefs) == 0 {
		return h.clearDraftFiles(ctx, draft.ID)
	}

	if len(payload.DraftFileRefs) > 0 {
		merged, err := h.mergeImportDraftFiles(ctx, assignmentID, importFiles, payload.DraftFileRefs)
		if err != nil {
			return err
		}
		stored := make([]draftStoredFile, len(merged))
		for i, item := range merged {
			stored[i] = item.stored
		}
		return h.replaceDraftFilesOnDraft(ctx, draft.ID, stored)
	}

	stored, err := h.storeDraftFileBytes(ctx, assignmentID, importFiles)
	if err != nil {
		return err
	}
	return h.replaceDraftFilesOnDraft(ctx, draft.ID, stored)
}

func (h *Handlers) attachDiscussionDraftFile(ctx context.Context, assignmentID int64, upload decodedDraftFile) error {
	draft, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		return err
	}
	if draft == nil {
		return fmt.Errorf("discussion draft row missing")
	}
	if draft.Mode != draftmode.Text {
		return fmt.Errorf("discussion attachment requires text draft mode")
	}

	stored, err := h.storeDraftFileBytes(ctx, assignmentID, []decodedDraftFile{upload})
	if err != nil {
		return err
	}
	return h.replaceDraftFilesOnDraft(ctx, draft.ID, stored)
}

func (h *Handlers) replaceDraftFilesOnDraft(ctx context.Context, draftID int64, files []draftStoredFile) error {
	if err := h.clearDraftFiles(ctx, draftID); err != nil {
		for _, file := range files {
			_ = h.files.Delete(file.StorageKey)
		}
		return err
	}

	for i, file := range files {
		if err := h.insertDraftFile(ctx, draftID, file, i); err != nil {
			return err
		}
	}

	return h.touchDraftUpdatedAt(ctx, draftID)
}

func (h *Handlers) decodeImportDraftFiles(files []importpayload.DraftFile) ([]decodedDraftFile, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > h.importLimits.MaxUploadSlots {
		return nil, fmt.Errorf("too many draft files")
	}

	var decoded []decodedDraftFile
	for i, file := range files {
		content := strings.TrimSpace(file.ContentBase64)
		if content == "" {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("decode draftFiles[%d]: %w", i, err)
		}
		fileName := strings.TrimSpace(file.FileName)
		if fileName == "" {
			return nil, fmt.Errorf("draftFiles[%d].fileName is required", i)
		}
		if len(data) > h.importLimits.MaxUploadBytes {
			return nil, fmt.Errorf("draftFiles[%d] exceeds maximum size", i)
		}
		decoded = append(decoded, decodedDraftFile{
			FileName:     fileName,
			MimeType:     strings.TrimSpace(file.MimeType),
			Data:         data,
			CanvasFileID: strings.TrimSpace(file.CanvasFileID),
		})
	}
	return decoded, nil
}

func canvasDraftSourceType(payload importpayload.Payload) string {
	if payload.PageType == "discussion" {
		return "canvas_discussion_reply"
	}
	return "canvas_text_entry"
}

func (h *Handlers) appendUploadedDraftFiles(ctx context.Context, assignmentID int64, uploads []decodedDraftFile) error {
	if len(uploads) == 0 {
		return fmt.Errorf("no files to upload")
	}

	stored, err := h.storeDraftFileBytes(ctx, assignmentID, uploads)
	if err != nil {
		return err
	}

	draft, err := h.ensureLatestDraftRow(ctx, assignmentID, draftUpsertOptions{
		Mode:       draftmode.File,
		SourceType: "file_upload",
	})
	if err != nil {
		return err
	}

	existing, err := h.loadDraftFiles(ctx, draft.ID)
	if err != nil {
		return err
	}
	if len(existing)+len(stored) > h.importLimits.MaxUploadSlots {
		for _, file := range stored {
			_ = h.files.Delete(file.StorageKey)
		}
		return fmt.Errorf("draft upload slot limit exceeded")
	}

	for i, file := range stored {
		if err := h.insertDraftFile(ctx, draft.ID, file, len(existing)+i); err != nil {
			return err
		}
	}

	return h.touchDraftUpdatedAt(ctx, draft.ID)
}

func (h *Handlers) storeDraftFileBytes(ctx context.Context, assignmentID int64, uploads []decodedDraftFile) ([]draftStoredFile, error) {
	userID, err := userIDFrom(ctx)
	if err != nil {
		return nil, err
	}

	var stored []draftStoredFile
	for _, upload := range uploads {
		storageKey, err := h.files.Save(userID, assignmentID, upload.FileName, upload.Data)
		if err != nil {
			for _, saved := range stored {
				_ = h.files.Delete(saved.StorageKey)
			}
			return nil, err
		}
		stored = append(stored, draftStoredFile{
			StorageKey:   storageKey,
			FileName:     upload.FileName,
			MimeType:     upload.MimeType,
			ByteSize:     int64(len(upload.Data)),
			UploadedAt:   time.Now().UTC(),
			CanvasFileID: upload.CanvasFileID,
		})
	}
	return stored, nil
}

func (h *Handlers) replaceDraftFiles(ctx context.Context, assignmentID int64, files []draftStoredFile, sourceType string, fromCanvas bool) error {
	draft, err := h.ensureLatestDraftRow(ctx, assignmentID, draftUpsertOptions{
		Mode:       draftmode.File,
		SourceType: sourceType,
		FromCanvas: fromCanvas,
	})
	if err != nil {
		for _, file := range files {
			_ = h.files.Delete(file.StorageKey)
		}
		return err
	}

	if err := h.clearDraftFiles(ctx, draft.ID); err != nil {
		for _, file := range files {
			_ = h.files.Delete(file.StorageKey)
		}
		return err
	}

	for i, file := range files {
		if err := h.insertDraftFile(ctx, draft.ID, file, i); err != nil {
			return err
		}
	}

	return h.touchDraftUpdatedAt(ctx, draft.ID)
}

func (h *Handlers) removeDraftFileByID(ctx context.Context, assignmentID, fileID int64) error {
	draft, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		return err
	}
	if draft == nil {
		return nil
	}

	var storageKey string
	err = h.db.Pool.QueryRow(ctx, `
		DELETE FROM submission_draft_files
		WHERE id = $1 AND submission_draft_id = $2
		RETURNING file_storage_key
	`, fileID, draft.ID).Scan(&storageKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	if err := h.files.Delete(storageKey); err != nil {
		return err
	}

	return h.touchDraftUpdatedAt(ctx, draft.ID)
}

func (h *Handlers) clearDraftTextBody(ctx context.Context, assignmentID int64) error {
	userID, err := userIDFrom(ctx)
	if err != nil {
		return err
	}

	draft, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		return err
	}
	if draft == nil {
		return nil
	}
	_, err = h.db.Pool.Exec(ctx, `
		UPDATE submission_drafts
		SET body = '', word_count = 0, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, draft.ID, userID)
	return err
}

func (h *Handlers) clearDraftURL(ctx context.Context, assignmentID int64) error {
	userID, err := userIDFrom(ctx)
	if err != nil {
		return err
	}

	draft, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		return err
	}
	if draft == nil {
		return nil
	}
	_, err = h.db.Pool.Exec(ctx, `
		UPDATE submission_drafts
		SET submission_url = NULL, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, draft.ID, userID)
	return err
}

func (h *Handlers) switchDraftModeOnly(ctx context.Context, assignmentID int64, mode string) error {
	userID, err := userIDFrom(ctx)
	if err != nil {
		return err
	}

	mode = draftmode.Normalize(mode)
	draft, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		return err
	}
	if draft == nil {
		return h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
			Mode:       mode,
			SourceType: "manual_paste",
		})
	}
	_, err = h.db.Pool.Exec(ctx, `
		UPDATE submission_drafts
		SET draft_mode = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, draft.ID, userID, mode)
	return err
}

func (h *Handlers) saveDraftURL(ctx context.Context, assignmentID int64, url string, sourceType string, fromCanvas bool) error {
	return h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
		Mode:       draftmode.URL,
		URL:        strings.TrimSpace(url),
		SourceType: sourceType,
		FromCanvas: fromCanvas,
	})
}

func (h *Handlers) switchDraftMode(ctx context.Context, assignmentID int64, mode string) error {
	return h.switchDraftModeOnly(ctx, assignmentID, mode)
}

func (h *Handlers) loadLatestDraftRow(ctx context.Context, assignmentID int64) (*draftRow, error) {
	userID, err := userIDFrom(ctx)
	if err != nil {
		return nil, err
	}

	var row draftRow
	err = h.db.Pool.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(draft_mode, 'text'),
			COALESCE(body, ''),
			submission_url,
			source_type,
			captured_from_canvas
		FROM submission_drafts
		WHERE assignment_snapshot_id = $1 AND user_id = $2
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, assignmentID, userID).Scan(
		&row.ID,
		&row.Mode,
		&row.Body,
		&row.URL,
		&row.SourceType,
		&row.FromCanvas,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (h *Handlers) loadDraftFiles(ctx context.Context, draftID int64) ([]draftFileRow, error) {
	rows, err := h.db.Pool.Query(ctx, `
		SELECT
			id,
			source_file_name,
			file_storage_key,
			COALESCE(file_mime_type, ''),
			COALESCE(file_byte_size, 0),
			uploaded_at,
			sort_order,
			COALESCE(canvas_file_id, '')
		FROM submission_draft_files
		WHERE submission_draft_id = $1
		ORDER BY sort_order ASC, id ASC
	`, draftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []draftFileRow
	for rows.Next() {
		var file draftFileRow
		if err := rows.Scan(
			&file.ID,
			&file.FileName,
			&file.StorageKey,
			&file.MimeType,
			&file.ByteSize,
			&file.UploadedAt,
			&file.SortOrder,
			&file.CanvasFileID,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func (h *Handlers) readDraftFileData(fileName, storageKey string) ([]byte, error) {
	data, err := h.files.Read(storageKey)
	if errors.Is(err, draftfiles.ErrNotFound) {
		return nil, fmt.Errorf("draft file %q is missing from storage — re-upload it", fileName)
	}
	if err != nil {
		return nil, fmt.Errorf("read draft file %q: %w", fileName, err)
	}
	return data, nil
}

func (h *Handlers) ensureLatestDraftRow(ctx context.Context, assignmentID int64, opts draftUpsertOptions) (*draftRow, error) {
	if err := h.upsertLatestDraft(ctx, assignmentID, opts); err != nil {
		return nil, err
	}
	return h.loadLatestDraftRow(ctx, assignmentID)
}

func applyDraftUpsert(existing *draftRow, opts draftUpsertOptions) (draftRow, bool) {
	next := draftRow{
		Mode:       draftmode.Text,
		SourceType: "manual_paste",
	}
	if existing != nil {
		next = *existing
	}

	if opts.Mode == "" {
		return next, false
	}

	next.Mode = draftmode.Normalize(opts.Mode)

	if opts.ClearText {
		next.Body = ""
	} else if next.Mode == draftmode.Text && (opts.SourceType != "" || opts.Body != "") {
		next.Body = opts.Body
	}

	if opts.ClearURL {
		next.URL = nil
	} else if next.Mode == draftmode.URL {
		if opts.URL != "" {
			url := strings.TrimSpace(opts.URL)
			next.URL = &url
		} else if opts.SourceType != "" {
			next.URL = nil
		}
	}

	if opts.SourceType != "" {
		next.SourceType = opts.SourceType
		next.FromCanvas = opts.FromCanvas
	}

	return next, true
}

func (h *Handlers) upsertLatestDraft(ctx context.Context, assignmentID int64, opts draftUpsertOptions) error {
	userID, err := userIDFrom(ctx)
	if err != nil {
		return err
	}

	existing, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		return err
	}

	next, changed := applyDraftUpsert(existing, opts)
	if !changed {
		return nil
	}

	wordCount := pages.DraftWordCount(next.Body)

	if existing == nil {
		err = h.db.Pool.QueryRow(ctx, `
			INSERT INTO submission_drafts (
				assignment_snapshot_id,
				user_id,
				draft_mode,
				body,
				word_count,
				submission_url,
				source_type,
				captured_from_canvas
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id
		`, assignmentID, userID, next.Mode, next.Body, wordCount, next.URL, next.SourceType, next.FromCanvas).Scan(&next.ID)
		return err
	}

	_, err = h.db.Pool.Exec(ctx, `
		UPDATE submission_drafts
		SET draft_mode = $3,
			body = $4,
			word_count = $5,
			submission_url = $6,
			source_type = $7,
			captured_from_canvas = $8,
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, existing.ID, userID, next.Mode, next.Body, wordCount, next.URL, next.SourceType, next.FromCanvas)
	return err
}

func (h *Handlers) insertDraftFile(ctx context.Context, draftID int64, file draftStoredFile, sortOrder int) error {
	_, err := h.db.Pool.Exec(ctx, `
		INSERT INTO submission_draft_files (
			submission_draft_id,
			source_file_name,
			file_storage_key,
			file_mime_type,
			file_byte_size,
			canvas_file_id,
			uploaded_at,
			sort_order
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, draftID, file.FileName, file.StorageKey, nullIfEmptyString(file.MimeType), file.ByteSize, nullIfEmptyString(file.CanvasFileID), file.UploadedAt, sortOrder)
	return err
}

func (h *Handlers) clearDraftFiles(ctx context.Context, draftID int64) error {
	files, err := h.loadDraftFiles(ctx, draftID)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	if _, err := h.db.Pool.Exec(ctx, `
		DELETE FROM submission_draft_files WHERE submission_draft_id = $1
	`, draftID); err != nil {
		return err
	}

	for _, file := range files {
		if err := h.files.Delete(file.StorageKey); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handlers) touchDraftUpdatedAt(ctx context.Context, draftID int64) error {
	userID, err := userIDFrom(ctx)
	if err != nil {
		return err
	}

	_, err = h.db.Pool.Exec(ctx, `
		UPDATE submission_drafts SET updated_at = NOW() WHERE id = $1 AND user_id = $2
	`, draftID, userID)
	return err
}

func nullIfEmptyString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

type mergedDraftFile struct {
	stored    draftStoredFile
	sortOrder int
}

func (h *Handlers) mergeImportDraftFiles(
	ctx context.Context,
	assignmentID int64,
	importFiles []decodedDraftFile,
	refs []importpayload.DraftFileRef,
) ([]mergedDraftFile, error) {
	draft, err := h.loadLatestDraftRow(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	existingByID := map[int64]draftFileRow{}
	if draft != nil {
		existing, err := h.loadDraftFiles(ctx, draft.ID)
		if err != nil {
			return nil, err
		}
		for _, file := range existing {
			existingByID[file.ID] = file
		}
	}

	sortedRefs := append([]importpayload.DraftFileRef(nil), refs...)
	sort.SliceStable(sortedRefs, func(i, j int) bool {
		return sortedRefs[i].SortOrder < sortedRefs[j].SortOrder
	})

	var merged []mergedDraftFile
	for _, ref := range sortedRefs {
		row, ok := existingByID[ref.ServerFileID]
		if !ok {
			return nil, fmt.Errorf("draftFileRefs: unknown serverFileId %d", ref.ServerFileID)
		}
		fileName := strings.TrimSpace(ref.FileName)
		if fileName == "" {
			fileName = row.FileName
		}
		canvasFileID := strings.TrimSpace(ref.CanvasFileID)
		if canvasFileID == "" {
			canvasFileID = row.CanvasFileID
		}
		merged = append(merged, mergedDraftFile{
			stored: draftStoredFile{
				StorageKey:   row.StorageKey,
				FileName:     fileName,
				MimeType:     row.MimeType,
				ByteSize:     row.ByteSize,
				UploadedAt:   row.UploadedAt,
				CanvasFileID: canvasFileID,
			},
			sortOrder: ref.SortOrder,
		})
	}

	stored, err := h.storeDraftFileBytes(ctx, assignmentID, importFiles)
	if err != nil {
		return nil, err
	}
	refCount := len(merged)
	for i, file := range stored {
		sortOrder := refCount + i
		merged = append(merged, mergedDraftFile{stored: file, sortOrder: sortOrder})
	}

	if len(merged) > h.importLimits.MaxUploadSlots {
		for _, file := range stored {
			_ = h.files.Delete(file.StorageKey)
		}
		return nil, fmt.Errorf("draft upload slot limit exceeded")
	}

	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].sortOrder < merged[j].sortOrder
	})

	return merged, nil
}

func (h *Handlers) replaceDraftFilesMerged(
	ctx context.Context,
	assignmentID int64,
	files []mergedDraftFile,
	sourceType string,
	fromCanvas bool,
) error {
	draft, err := h.ensureLatestDraftRow(ctx, assignmentID, draftUpsertOptions{
		Mode:       draftmode.File,
		SourceType: sourceType,
		FromCanvas: fromCanvas,
	})
	if err != nil {
		for _, file := range files {
			_ = h.files.Delete(file.stored.StorageKey)
		}
		return err
	}

	existing, err := h.loadDraftFiles(ctx, draft.ID)
	if err != nil {
		for _, file := range files {
			_ = h.files.Delete(file.stored.StorageKey)
		}
		return err
	}

	keptKeys := map[string]struct{}{}
	for _, file := range files {
		keptKeys[file.stored.StorageKey] = struct{}{}
	}
	for _, old := range existing {
		if _, keep := keptKeys[old.StorageKey]; !keep {
			if err := h.files.Delete(old.StorageKey); err != nil {
				return err
			}
		}
	}

	if _, err := h.db.Pool.Exec(ctx, `
		DELETE FROM submission_draft_files WHERE submission_draft_id = $1
	`, draft.ID); err != nil {
		return err
	}

	for i, file := range files {
		if err := h.insertDraftFile(ctx, draft.ID, file.stored, i); err != nil {
			return err
		}
	}

	return h.touchDraftUpdatedAt(ctx, draft.ID)
}
