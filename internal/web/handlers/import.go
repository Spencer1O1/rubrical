package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"rubrical/internal/draftmode"
	"rubrical/internal/importmeta"
	"rubrical/internal/importpayload"
	"rubrical/internal/submissiontypes"
	"rubrical/internal/web/pages"
)


func (h *Handlers) ImportAssignment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(h.importLimits.MaxBodyBytes))

	var payload importpayload.Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid import payload", http.StatusBadRequest)
		return
	}

	if err := importpayload.ValidateAndNormalize(&payload, h.importLimits); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, created, err := h.upsertAssignmentSnapshot(r.Context(), payload)
	if err != nil {
		http.Error(w, "failed to import assignment", http.StatusInternalServerError)
		return
	}

	if err := h.saveDraftFromImport(r.Context(), id, payload); err != nil {
		// Assignment context still imports when draft/file capture fails.
		_ = err
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":       id,
		"created":  created,
		"redirect": fmt.Sprintf("/assignments/%d", id),
	})
}

func (h *Handlers) upsertAssignmentSnapshot(ctx context.Context, payload importpayload.Payload) (int64, bool, error) {
	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback(ctx)

	dueAt, dueOK := importmeta.ParseDueAtISO(payload.Metadata.DueAt)
	if !dueOK {
		dueAt, dueOK = importmeta.ParseDueAt(payload.Metadata.DueDateText, payload.CapturedAt)
	}
	pointsPossible, pointsOK := importmeta.ParsePointsPossible(payload.Metadata.PointsPossibleText)

	var dueAtValue any
	if dueOK {
		dueAtValue = dueAt
	}
	var pointsValue any
	if pointsOK {
		pointsValue = pointsPossible
	}

	allowedTypesJSON, err := json.Marshal(payload.Metadata.AllowedSubmissionTypes)
	if err != nil {
		return 0, false, fmt.Errorf("marshal allowed submission types: %w", err)
	}

	var id int64
	var created bool
	err = tx.QueryRow(ctx, `
		INSERT INTO assignment_snapshots (
			user_id,
			source_url,
			source_platform,
			page_type,
			course_name,
			assignment_title,
			raw_text,
			instructions_text,
			submission_type,
			allowed_submission_types,
			due_at,
			points_possible,
			imported_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (user_id, source_url) DO UPDATE SET
			page_type = EXCLUDED.page_type,
			course_name = EXCLUDED.course_name,
			assignment_title = EXCLUDED.assignment_title,
			raw_text = EXCLUDED.raw_text,
			instructions_text = EXCLUDED.instructions_text,
			submission_type = EXCLUDED.submission_type,
			allowed_submission_types = EXCLUDED.allowed_submission_types,
			due_at = EXCLUDED.due_at,
			points_possible = EXCLUDED.points_possible,
			imported_at = NOW(),
			updated_at = NOW()
		RETURNING id, (xmax = 0) AS created
	`,
		h.userID,
		payload.SourceURL,
		"canvas",
		nullIfEmpty(payload.PageType),
		nullIfEmpty(payload.Metadata.CourseName),
		nullIfEmpty(payload.Title),
		nullIfEmpty(payload.VisibleText),
		nullIfEmpty(pages.SanitizedInstructionsHTML(payload.InstructionsText)),
		nullIfEmpty(payload.Metadata.SubmissionTypeText),
		allowedTypesJSON,
		dueAtValue,
		pointsValue,
	).Scan(&id, &created)
	if err != nil {
		return 0, false, err
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM rubric_criteria WHERE assignment_snapshot_id = $1
	`, id); err != nil {
		return 0, false, err
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM extracted_sources
		WHERE assignment_snapshot_id = $1 AND source_kind = 'rubric_header'
	`, id); err != nil {
		return 0, false, err
	}

	if payload.Rubric != nil {
		if len(payload.Rubric.Header) > 0 {
			headerJSON, err := json.Marshal(payload.Rubric.Header)
			if err != nil {
				return 0, false, err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO extracted_sources (
					assignment_snapshot_id,
					source_kind,
					normalized_content
				) VALUES ($1, 'rubric_header', $2)
			`, id, string(headerJSON)); err != nil {
				return 0, false, err
			}
		}

		for i, row := range payload.Rubric.Rows {
			rowJSON, err := json.Marshal(row)
			if err != nil {
				return 0, false, err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO rubric_criteria (
					assignment_snapshot_id,
					name,
					ratings_json,
					sort_order
				) VALUES ($1, $2, $3, $4)
			`, id, nullIfEmpty(row.Criterion), rowJSON, i); err != nil {
				return 0, false, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, false, err
	}

	return id, created, nil
}

func (h *Handlers) getAssignment(ctx context.Context, id int64) (pages.AssignmentView, error) {
	var view pages.AssignmentView
	var importedAt time.Time
	var draftWordCount int
	var draftMode *string
	var draftSubmissionURL *string
	var draftID int64
	var dueAt *time.Time
	var pointsPossible *float64
	var pageType *string
	var allowedSubmissionTypesJSON []byte
	err := h.db.Pool.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(page_type, ''),
			COALESCE(assignment_title, 'Untitled assignment'),
			COALESCE(course_name, ''),
			COALESCE(instructions_text, ''),
			COALESCE(raw_text, ''),
			imported_at,
			COALESCE(submission_type, ''),
			allowed_submission_types,
			due_at,
			points_possible
		FROM assignment_snapshots
		WHERE id = $1 AND user_id = $2
	`, id, h.userID).Scan(
		&view.ID,
		&pageType,
		&view.Title,
		&view.CourseName,
		&view.Instructions,
		&view.RawText,
		&importedAt,
		&view.SubmissionType,
		&allowedSubmissionTypesJSON,
		&dueAt,
		&pointsPossible,
	)
	if err != nil {
		return pages.AssignmentView{}, err
	}

	canvasAllowedTypes := decodeAllowedSubmissionTypes(allowedSubmissionTypesJSON)
	if pageType != nil {
		view.PageType = strings.TrimSpace(*pageType)
	}
	if view.PageType == "discussion" {
		view.AllowedDraftModes = []string{draftmode.Text}
	} else {
		view.AllowedDraftModes = pages.AllowedDraftModes(canvasAllowedTypes)
	}

	view.ImportedAtLabel = pages.ImportedAtLabel(importedAt)
	view.DueAtLabel = pages.DueAtLabel(dueAt)
	view.PointsPossibleLabel = pages.PointsPossibleLabel(pointsPossible)
	view.PointsPossibleChip = pages.PointsPossibleChip(pointsPossible)
	view.AnalyzeURL = pages.AnalyzeURL(view.ID)
	view.DraftUploadURL = pages.DraftUploadURL(view.ID)
	view.DraftDiscussionUploadURL = pages.DraftDiscussionUploadURL(view.ID)
	view.DiscussionAttachmentsAllowed = view.PageType == "discussion" &&
		submissiontypes.AttachmentsAllowed(canvasAllowedTypes)
	view.DraftModeURL = pages.DraftModeURL(view.ID)
	view.DraftTextSaveURL = pages.DraftTextSaveURL(view.ID)
	view.DraftSaveURL = pages.DraftSaveURL(view.ID)
	view.StrictExtraction = h.strictExtraction
	view.InstructionsHTML = pages.PrepareInstructionsHTML(view.Instructions)

	rubric, err := h.loadRubricTable(ctx, id, h.strictExtraction)
	if err != nil {
		return pages.AssignmentView{}, err
	}
	view.Rubric = rubric

	err = h.db.Pool.QueryRow(ctx, `
		SELECT id, COALESCE(body, ''), COALESCE(word_count, 0),
			COALESCE(draft_mode, ''), submission_url
		FROM submission_drafts
		WHERE assignment_snapshot_id = $1 AND user_id = $2
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, id, h.userID).Scan(&draftID, &view.DraftBody, &draftWordCount, &draftMode, &draftSubmissionURL)
	if errors.Is(err, pgx.ErrNoRows) {
		view.DraftBody = ""
		draftWordCount = 0
		view.DraftMode = pages.DefaultDraftMode(view.AllowedDraftModes)
	} else if err != nil {
		return pages.AssignmentView{}, err
	}

	if draftMode != nil && strings.TrimSpace(*draftMode) != "" {
		view.DraftMode = *draftMode
	} else if view.DraftMode == "" {
		view.DraftMode = pages.DefaultDraftMode(view.AllowedDraftModes)
	}
	view.DraftMode = pages.ClampDraftMode(view.DraftMode, view.AllowedDraftModes)
	if draftSubmissionURL != nil {
		view.DraftSubmissionURL = *draftSubmissionURL
	}

	if draftID > 0 {
		files, err := h.loadDraftFiles(ctx, draftID)
		if err != nil {
			return pages.AssignmentView{}, err
		}
		for _, file := range files {
			view.DraftFiles = append(view.DraftFiles, pages.DraftFileView{
				ID:              file.ID,
				FileName:        file.FileName,
				UploadedAtLabel: pages.UploadedAtLabel(file.UploadedAt),
				RemoveURL:       pages.DraftRemoveFileURL(view.ID, file.ID),
			})
		}
	}

	fileNames := make([]string, len(view.DraftFiles))
	for i, file := range view.DraftFiles {
		fileNames[i] = file.FileName
	}
	view.HasDraftFiles = len(view.DraftFiles) > 0
	if view.PageType == "discussion" {
		view.DraftStatusLabel = pages.DiscussionDraftStatusLabel(draftWordCount, fileNames)
	} else {
		view.DraftStatusLabel = pages.DraftStatusLabel(draftWordCount, fileNames, view.DraftSubmissionURL, view.DraftMode)
	}

	if h.analysis != nil {
		if result, err := h.analysis.LoadLatestResult(ctx, id); err == nil && result != nil {
			view.Analysis = pages.AnalysisResultsFromResult(result)
		}
	}

	return view, nil
}

func decodeAllowedSubmissionTypes(raw []byte) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var types []string
	if err := json.Unmarshal(raw, &types); err != nil {
		return nil
	}
	return types
}

func (h *Handlers) loadAllowedDraftModes(ctx context.Context, assignmentID int64) ([]string, error) {
	var allowedJSON []byte
	err := h.db.Pool.QueryRow(ctx, `
		SELECT allowed_submission_types
		FROM assignment_snapshots
		WHERE id = $1 AND user_id = $2
	`, assignmentID, h.userID).Scan(&allowedJSON)
	if err != nil {
		return nil, err
	}
	return submissiontypes.AllowedDraftModes(decodeAllowedSubmissionTypes(allowedJSON)), nil
}

func (h *Handlers) loadAssignmentPageType(ctx context.Context, assignmentID int64) (string, error) {
	var pageType *string
	err := h.db.Pool.QueryRow(ctx, `
		SELECT page_type
		FROM assignment_snapshots
		WHERE id = $1 AND user_id = $2
	`, assignmentID, h.userID).Scan(&pageType)
	if err != nil {
		return "", err
	}
	if pageType == nil {
		return "", nil
	}
	return strings.TrimSpace(*pageType), nil
}

func nullIfEmpty(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (h *Handlers) loadRubricTable(ctx context.Context, assignmentID int64, strict bool) (pages.RubricTableView, error) {
	table := pages.RubricTableView{}
	if !strict {
		table.Header = pages.DefaultRubricHeader()
	}

	var headerContent *string
	err := h.db.Pool.QueryRow(ctx, `
		SELECT normalized_content
		FROM extracted_sources
		WHERE assignment_snapshot_id = $1 AND source_kind = 'rubric_header'
		LIMIT 1
	`, assignmentID).Scan(&headerContent)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return table, err
	}
	if headerContent != nil {
		_ = json.Unmarshal([]byte(*headerContent), &table.Header)
	}

	rows, err := h.db.Pool.Query(ctx, `
		SELECT COALESCE(name, ''), COALESCE(raw_text, ''), ratings_json
		FROM rubric_criteria
		WHERE assignment_snapshot_id = $1
		ORDER BY sort_order
	`, assignmentID)
	if err != nil {
		return table, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var rawText string
		var ratingsJSON []byte
		if err := rows.Scan(&name, &rawText, &ratingsJSON); err != nil {
			return table, err
		}

		if len(ratingsJSON) > 0 {
			var row importpayload.RubricRow
			if err := json.Unmarshal(ratingsJSON, &row); err != nil {
				return table, err
			}
			table.Rows = append(table.Rows, rubricRowViewFromPayload(row))
			continue
		}

		if strict {
			continue
		}

		table.Rows = append(table.Rows, rubricRowViewFromLegacy(name, rawText))
	}
	if err := rows.Err(); err != nil {
		return table, err
	}

	return table, nil
}

func rubricRowViewFromPayload(row importpayload.RubricRow) pages.RubricRowView {
	view := pages.RubricRowView{
		Criterion:                row.Criterion,
		CriterionLongDescription: row.CriterionLongDescription,
		Points:                   row.Points,
	}
	for _, rating := range row.Ratings {
		view.Ratings = append(view.Ratings, pages.RubricRatingView{
			Title:       rating.Title,
			Description: rating.Description,
			Points:      rating.Points,
		})
	}
	view.Ratings = pages.NormalizeRubricRatings(view.Ratings)
	return view
}

func rubricRowViewFromLegacy(name, rawText string) pages.RubricRowView {
	parts := strings.Split(rawText, " | ")
	criterion := name
	points := ""
	description := rawText

	if len(parts) >= 3 {
		criterion = parts[0]
		description = parts[1]
		points = parts[2]
	} else if len(parts) == 2 {
		description = parts[1]
	}

	return pages.RubricRowView{
		Criterion: criterion,
		Ratings: []pages.RubricRatingView{
			{Description: description},
		},
		Points: points,
	}
}
