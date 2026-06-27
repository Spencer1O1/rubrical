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
	"rubrical/internal/importurl"
	"rubrical/internal/web/pages"
)

type rubricRatingPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Points      string `json:"points"`
}

type rubricTableRowPayload struct {
	Criterion string                `json:"criterion"`
	Ratings   []rubricRatingPayload `json:"ratings"`
	Points    string                `json:"points"`
}

type rubricTablePayload struct {
	Header []string                `json:"header"`
	Rows   []rubricTableRowPayload `json:"rows"`
}

type importPayload struct {
	SourceURL        string              `json:"sourceUrl"`
	PageType         string              `json:"pageType"`
	Title            string              `json:"title"`
	VisibleText      string              `json:"visibleText"`
	InstructionsText string              `json:"instructionsText"`
	DraftText        string              `json:"draftText"`
	Rubric           *rubricTablePayload `json:"rubric"`
	Metadata         importMetadata      `json:"metadata"`
	CaptureMode      string              `json:"captureMode"`
	CapturedAt       time.Time           `json:"capturedAt"`
}

type importMetadata struct {
	DueDateText        string `json:"dueDateText"`
	PointsPossibleText string `json:"pointsPossibleText"`
	SubmissionTypeText string `json:"submissionTypeText"`
	CourseName         string `json:"courseName"`
}

func (h *Handlers) ImportAssignment(w http.ResponseWriter, r *http.Request) {
	var payload importPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid import payload", http.StatusBadRequest)
		return
	}

	sourceURL := importurl.NormalizeSourceURL(payload.SourceURL)
	if sourceURL == "" {
		http.Error(w, "sourceUrl is required", http.StatusBadRequest)
		return
	}
	payload.SourceURL = sourceURL

	id, created, err := h.upsertAssignmentSnapshot(r.Context(), payload)
	if err != nil {
		http.Error(w, "failed to import assignment", http.StatusInternalServerError)
		return
	}

	if strings.TrimSpace(payload.DraftText) != "" {
		if err := h.saveCanvasDraft(r.Context(), id, payload.DraftText); err != nil {
			http.Error(w, "failed to save draft", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":       id,
		"created":  created,
		"redirect": fmt.Sprintf("/assignments/%d", id),
	})
}

func (h *Handlers) upsertAssignmentSnapshot(ctx context.Context, payload importPayload) (int64, bool, error) {
	tx, err := h.db.Pool.Begin(ctx)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback(ctx)

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
			imported_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (user_id, source_url) DO UPDATE SET
			page_type = EXCLUDED.page_type,
			course_name = EXCLUDED.course_name,
			assignment_title = EXCLUDED.assignment_title,
			raw_text = EXCLUDED.raw_text,
			instructions_text = EXCLUDED.instructions_text,
			submission_type = EXCLUDED.submission_type,
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

func (h *Handlers) createDraft(ctx context.Context, assignmentID int64, body, sourceType string, fromCanvas bool) error {
	wordCount := len(strings.Fields(body))
	_, err := h.db.Pool.Exec(ctx, `
		INSERT INTO submission_drafts (
			assignment_snapshot_id,
			user_id,
			body,
			word_count,
			source_type,
			captured_from_canvas
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, assignmentID, h.userID, body, wordCount, sourceType, fromCanvas)
	return err
}

func (h *Handlers) saveCanvasDraft(ctx context.Context, assignmentID int64, body string) error {
	wordCount := len(strings.Fields(body))

	var draftID int64
	err := h.db.Pool.QueryRow(ctx, `
		SELECT id
		FROM submission_drafts
		WHERE assignment_snapshot_id = $1
			AND user_id = $2
			AND captured_from_canvas = true
			AND source_type = 'canvas_text_entry'
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, assignmentID, h.userID).Scan(&draftID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return h.createDraft(ctx, assignmentID, body, "canvas_text_entry", true)
	}

	_, err = h.db.Pool.Exec(ctx, `
		UPDATE submission_drafts
		SET body = $3,
			word_count = $4,
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, draftID, h.userID, body, wordCount)
	return err
}

func (h *Handlers) getAssignment(ctx context.Context, id int64) (pages.AssignmentView, error) {
	var view pages.AssignmentView
	var importedAt time.Time
	var draftWordCount int
	err := h.db.Pool.QueryRow(ctx, `
		SELECT
			id,
			COALESCE(assignment_title, 'Untitled assignment'),
			COALESCE(course_name, ''),
			COALESCE(instructions_text, ''),
			COALESCE(raw_text, ''),
			imported_at
		FROM assignment_snapshots
		WHERE id = $1 AND user_id = $2
	`, id, h.userID).Scan(
		&view.ID,
		&view.Title,
		&view.CourseName,
		&view.Instructions,
		&view.RawText,
		&importedAt,
	)
	if err != nil {
		return pages.AssignmentView{}, err
	}

	view.ImportedAtLabel = pages.ImportedAtLabel(importedAt)
	view.AnalyzeURL = pages.AnalyzeURL(view.ID)
	view.StrictExtraction = h.strictExtraction
	view.InstructionsHTML = pages.PrepareInstructionsHTML(view.Instructions)

	rubric, err := h.loadRubricTable(ctx, id, h.strictExtraction)
	if err != nil {
		return pages.AssignmentView{}, err
	}
	view.Rubric = rubric

	err = h.db.Pool.QueryRow(ctx, `
		SELECT COALESCE(body, ''), COALESCE(word_count, 0)
		FROM submission_drafts
		WHERE assignment_snapshot_id = $1 AND user_id = $2
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, id, h.userID).Scan(&view.DraftBody, &draftWordCount)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return pages.AssignmentView{}, err
	}

	view.DraftStatusLabel = pages.DraftStatusLabel(draftWordCount)

	return view, nil
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
			var row rubricTableRowPayload
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

func rubricRowViewFromPayload(row rubricTableRowPayload) pages.RubricRowView {
	view := pages.RubricRowView{
		Criterion: row.Criterion,
		Points:    row.Points,
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
