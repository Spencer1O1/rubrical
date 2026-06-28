package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"rubrical/internal/analysis/provider"
	"rubrical/internal/analysis/schema"
)

func loadRubricContext(ctx context.Context, pool *pgxpool.Pool, assignmentID int64) (RubricContext, error) {
	var rubric RubricContext

	var headerContent *string
	err := pool.QueryRow(ctx, `
		SELECT normalized_content
		FROM extracted_sources
		WHERE assignment_snapshot_id = $1 AND source_kind = 'rubric_header'
		LIMIT 1
	`, assignmentID).Scan(&headerContent)
	if err != nil && !errorsIsNoRows(err) {
		return rubric, err
	}
	if headerContent != nil {
		_ = json.Unmarshal([]byte(*headerContent), &rubric.Header)
	}

	rows, err := pool.Query(ctx, `
		SELECT COALESCE(name, ''), COALESCE(description, ''), COALESCE(points_possible, 0)::float8, ratings_json
		FROM rubric_criteria
		WHERE assignment_snapshot_id = $1
		ORDER BY sort_order
	`, assignmentID)
	if err != nil {
		return rubric, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, description string
		var points float64
		var ratingsJSON []byte
		if err := rows.Scan(&name, &description, &points, &ratingsJSON); err != nil {
			return rubric, err
		}

		row := RubricRow{
			Criterion:                name,
			CriterionLongDescription: description,
		}
		if points > 0 {
			row.Points = fmt.Sprintf("%.2f", points)
		}

		if len(ratingsJSON) > 0 {
			var payload struct {
				Criterion                string         `json:"criterion"`
				CriterionLongDescription string         `json:"criterionLongDescription"`
				Ratings                  []RubricRating `json:"ratings"`
				Points                   string         `json:"points"`
			}
			if err := json.Unmarshal(ratingsJSON, &payload); err == nil {
				if payload.Criterion != "" {
					row.Criterion = payload.Criterion
				}
				if payload.CriterionLongDescription != "" {
					row.CriterionLongDescription = payload.CriterionLongDescription
				}
				row.Ratings = payload.Ratings
				if payload.Points != "" {
					row.Points = payload.Points
				}
			}
		}

		rubric.Rows = append(rubric.Rows, row)
	}
	return rubric, rows.Err()
}

func (s *Service) createRun(ctx context.Context, assignmentID, draftID int64, provider, model string, inputLog []byte) (int64, error) {
	var runID int64
	var draftRef any
	if draftID > 0 {
		draftRef = draftID
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO analysis_runs (
			assignment_snapshot_id,
			submission_draft_id,
			provider,
			model,
			status,
			raw_model_input
		) VALUES ($1, $2, $3, $4, 'pending', $5::jsonb)
		RETURNING id
	`, assignmentID, draftRef, provider, model, string(inputLog)).Scan(&runID)
	return runID, err
}

func (s *Service) markRunStatus(ctx context.Context, runID int64, status string, output []byte, completedAt *time.Time) error {
	if completedAt != nil {
		_, err := s.pool.Exec(ctx, `
			UPDATE analysis_runs
			SET status = $2, raw_model_output = $3::jsonb, completed_at = $4
			WHERE id = $1
		`, runID, status, nullJSON(output), *completedAt)
		return err
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE analysis_runs
		SET status = $2
		WHERE id = $1
	`, runID, status)
	return err
}

func (s *Service) markRunFailed(ctx context.Context, runID int64, runErr error) error {
	payload, _ := json.Marshal(map[string]string{"error": runErr.Error()})
	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		UPDATE analysis_runs
		SET status = 'failed', raw_model_output = $2::jsonb, completed_at = $3
		WHERE id = $1
	`, runID, string(payload), now)
	return err
}

func (s *Service) persistSuccess(ctx context.Context, runID, assignmentID int64, ai provider.Provider, out *schema.ModelOutput) (Result, error) {
	outputJSON, err := json.Marshal(out)
	if err != nil {
		return Result{}, err
	}

	criterionIDs, err := loadCriterionIDs(ctx, s.pool, assignmentID)
	if err != nil {
		return Result{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		UPDATE analysis_runs
		SET status = 'completed',
			overall_summary = $2,
			estimated_score = $3,
			estimated_score_max = $4,
			confidence = $5,
			raw_model_output = $6::jsonb,
			completed_at = $7
		WHERE id = $1
	`, runID,
		out.OverallSummary,
		out.EstimatedScore,
		out.EstimatedScoreMax,
		out.Confidence,
		string(outputJSON),
		now,
	)
	if err != nil {
		return Result{}, err
	}

	sortOrder := 0
	var feedback []FeedbackItem

	for _, criterion := range out.Criteria {
		item := FeedbackItem{
			Category:    "criterion",
			Severity:    schema.SeverityForStatus(criterion.Status),
			Title:       criterion.CriterionName,
			Explanation: criterionStatusLabel(criterion.Status),
			Evidence:    criterion.Evidence,
			Suggestion:  criterion.Suggestion,
			Status:      "open",
			SortOrder:   sortOrder,
		}
		sortOrder++
		criterionID := matchCriterionID(criterionIDs, criterion.CriterionName)
		id, err := insertFeedbackItem(ctx, tx, runID, criterionID, item)
		if err != nil {
			return Result{}, err
		}
		item.ID = id
		feedback = append(feedback, item)
	}

	for _, missing := range out.MissingRequirements {
		item := FeedbackItem{
			Category:    "missing_requirement",
			Severity:    "warning",
			Title:       missing,
			Explanation: "This requirement appears to be missing or incomplete.",
			Status:      "open",
			SortOrder:   sortOrder,
		}
		sortOrder++
		id, err := insertFeedbackItem(ctx, tx, runID, nil, item)
		if err != nil {
			return Result{}, err
		}
		item.ID = id
		feedback = append(feedback, item)
	}

	for _, strength := range out.Strengths {
		item := FeedbackItem{
			Category:    "strength",
			Severity:    "info",
			Title:       strength,
			Status:      "open",
			SortOrder:   sortOrder,
		}
		sortOrder++
		id, err := insertFeedbackItem(ctx, tx, runID, nil, item)
		if err != nil {
			return Result{}, err
		}
		item.ID = id
		feedback = append(feedback, item)
	}

	for _, suggestion := range out.RevisionSuggestions {
		item := FeedbackItem{
			Category:    "suggestion",
			Severity:    "info",
			Title:       suggestion,
			Status:      "open",
			SortOrder:   sortOrder,
		}
		sortOrder++
		id, err := insertFeedbackItem(ctx, tx, runID, nil, item)
		if err != nil {
			return Result{}, err
		}
		item.ID = id
		feedback = append(feedback, item)
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, err
	}

	return Result{
		RunID:             runID,
		Provider:          ai.Name(),
		Model:             ai.Model(),
		OverallSummary:    out.OverallSummary,
		EstimatedScore:    out.EstimatedScore,
		EstimatedScoreMax: out.EstimatedScoreMax,
		Confidence:        out.Confidence,
		Feedback:          feedback,
		CompletedAt:       now,
	}, nil
}

func loadLatestResult(ctx context.Context, pool *pgxpool.Pool, assignmentID int64) (*Result, error) {
	var result Result
	var completedAt *time.Time
	err := pool.QueryRow(ctx, `
		SELECT id, COALESCE(provider, ''), COALESCE(model, ''), COALESCE(overall_summary, ''),
			estimated_score, estimated_score_max, COALESCE(confidence, ''), completed_at
		FROM analysis_runs
		WHERE assignment_snapshot_id = $1 AND status = 'completed'
		ORDER BY completed_at DESC NULLS LAST, id DESC
		LIMIT 1
	`, assignmentID).Scan(
		&result.RunID,
		&result.Provider,
		&result.Model,
		&result.OverallSummary,
		&result.EstimatedScore,
		&result.EstimatedScoreMax,
		&result.Confidence,
		&completedAt,
	)
	if errorsIsNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if completedAt != nil {
		result.CompletedAt = *completedAt
	}

	rows, err := pool.Query(ctx, `
		SELECT id, category, severity, title, COALESCE(explanation, ''), COALESCE(evidence, ''),
			COALESCE(suggestion, ''), status, sort_order
		FROM feedback_items
		WHERE analysis_run_id = $1
		ORDER BY sort_order ASC, id ASC
	`, result.RunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item FeedbackItem
		if err := rows.Scan(
			&item.ID,
			&item.Category,
			&item.Severity,
			&item.Title,
			&item.Explanation,
			&item.Evidence,
			&item.Suggestion,
			&item.Status,
			&item.SortOrder,
		); err != nil {
			return nil, err
		}
		result.Feedback = append(result.Feedback, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &result, nil
}

func insertFeedbackItem(ctx context.Context, tx pgx.Tx, runID int64, criterionID *int64, item FeedbackItem) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO feedback_items (
			analysis_run_id,
			rubric_criterion_id,
			category,
			severity,
			title,
			explanation,
			evidence,
			suggestion,
			status,
			sort_order
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`, runID, criterionID, item.Category, item.Severity, item.Title, item.Explanation, item.Evidence, item.Suggestion, item.Status, item.SortOrder).Scan(&id)
	return id, err
}

type criterionRef struct {
	ID   int64
	Name string
}

func loadCriterionIDs(ctx context.Context, pool *pgxpool.Pool, assignmentID int64) ([]criterionRef, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, COALESCE(name, '')
		FROM rubric_criteria
		WHERE assignment_snapshot_id = $1
		ORDER BY sort_order
	`, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []criterionRef
	for rows.Next() {
		var ref criterionRef
		if err := rows.Scan(&ref.ID, &ref.Name); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

func matchCriterionID(refs []criterionRef, name string) *int64 {
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, ref := range refs {
		if strings.ToLower(strings.TrimSpace(ref.Name)) == normalized {
			id := ref.ID
			return &id
		}
	}
	for _, ref := range refs {
		refName := strings.ToLower(strings.TrimSpace(ref.Name))
		if strings.Contains(refName, normalized) || strings.Contains(normalized, refName) {
			id := ref.ID
			return &id
		}
	}
	return nil
}

func criterionStatusLabel(status string) string {
	switch status {
	case "met":
		return "This criterion appears to be met."
	case "partially_met":
		return "This criterion is partially met."
	case "not_met":
		return "This criterion does not appear to be met."
	default:
		return ""
	}
}

func nullJSON(data []byte) *string {
	if len(data) == 0 {
		return nil
	}
	s := string(data)
	return &s
}

func errorsIsNoRows(err error) bool {
	return err == pgx.ErrNoRows
}
