package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"rubrical/internal/analysis/provider"
	"rubrical/internal/analysis/schema"
)

const DefaultStaleRunTTL = 10 * time.Minute

func failStaleRuns(ctx context.Context, pool *pgxpool.Pool, assignmentID int64, ttl time.Duration) error {
	if pool == nil || ttl <= 0 {
		return nil
	}
	payload, _ := json.Marshal(map[string]string{
		"error": "analysis timed out or did not finish",
	})
	_, err := pool.Exec(ctx, `
		UPDATE analysis_runs
		SET status = 'failed',
		    raw_model_output = $3::jsonb,
		    completed_at = NOW()
		WHERE assignment_snapshot_id = $1
		  AND status IN ('pending', 'running')
		  AND created_at < NOW() - $2::interval
	`, assignmentID, fmt.Sprintf("%f seconds", ttl.Seconds()), string(payload))
	return err
}

func FailAllStaleRuns(ctx context.Context, pool *pgxpool.Pool, ttl time.Duration) error {
	if pool == nil || ttl <= 0 {
		return nil
	}
	payload, _ := json.Marshal(map[string]string{
		"error": "analysis timed out or did not finish",
	})
	_, err := pool.Exec(ctx, `
		UPDATE analysis_runs
		SET status = 'failed',
		    raw_model_output = $2::jsonb,
		    completed_at = NOW()
		WHERE status IN ('pending', 'running')
		  AND created_at < NOW() - $1::interval
	`, fmt.Sprintf("%f seconds", ttl.Seconds()), string(payload))
	return err
}

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

func (s *Service) beginRun(ctx context.Context, userID, assignmentID, draftID int64, provider, model string, inputLog []byte) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	if err := failStaleRuns(ctx, s.pool, assignmentID, DefaultStaleRunTTL); err != nil {
		return 0, err
	}

	if s.limiter != nil {
		if err := s.limiter.checkInFlightInTx(ctx, tx, assignmentID); err != nil {
			return 0, err
		}
		if s.enforceLimits {
			if err := s.limiter.checkRateLimitsInTx(ctx, tx, userID, assignmentID); err != nil {
				return 0, err
			}
		}
	}

	var runID int64
	var draftRef any
	if draftID > 0 {
		draftRef = draftID
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO analysis_runs (
			assignment_snapshot_id,
			submission_draft_id,
			provider,
			model,
			status,
			raw_model_input
		) VALUES ($1, $2, $3, $4, 'running', $5::jsonb)
		RETURNING id
	`, assignmentID, draftRef, provider, model, string(inputLog)).Scan(&runID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return runID, nil
}

func (s *Service) saveScoredAnalysis(ctx context.Context, runID int64, out *schema.ScoredAnalysis) error {
	outputJSON, err := json.Marshal(out)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE analysis_runs
		SET raw_model_output = $2::jsonb
		WHERE id = $1
	`, runID, string(outputJSON))
	return err
}

func (s *Service) scoredAnalysisSaved(ctx context.Context, runID int64) (bool, error) {
	var saved bool
	err := s.pool.QueryRow(ctx, `
		SELECT raw_model_output IS NOT NULL
		FROM analysis_runs
		WHERE id = $1
	`, runID).Scan(&saved)
	return saved, err
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

func (s *Service) persistSuccess(ctx context.Context, runID, assignmentID int64, ai provider.Provider, out *schema.ScoredAnalysis) (Result, error) {
	if existing, err := loadRunResult(ctx, s.pool, runID); err != nil {
		return Result{}, err
	} else if existing != nil {
		return *existing, nil
	}

	outputJSON, err := json.Marshal(out)
	if err != nil {
		return Result{}, err
	}

	outputSaved, err := s.scoredAnalysisSaved(ctx, runID)
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
	if outputSaved {
		_, err = tx.Exec(ctx, `
			UPDATE analysis_runs
			SET status = 'completed',
				overall_summary = $2,
				predicted_score = $3,
				predicted_score_max = $4,
				confidence = $5,
				completed_at = $6
			WHERE id = $1
		`, runID,
			out.OverallSummary,
			out.PredictedScore,
			out.PredictedScoreMax,
			out.Confidence,
			now,
		)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE analysis_runs
			SET status = 'completed',
				overall_summary = $2,
				predicted_score = $3,
				predicted_score_max = $4,
				confidence = $5,
				raw_model_output = $6::jsonb,
				completed_at = $7
			WHERE id = $1
		`, runID,
			out.OverallSummary,
			out.PredictedScore,
			out.PredictedScoreMax,
			out.Confidence,
			string(outputJSON),
			now,
		)
	}
	if err != nil {
		return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
	}

	sortOrder := 0
	var feedback []FeedbackItem

	for _, criterion := range out.Criteria {
		item := FeedbackItem{
			Category:        "criterion",
			Severity:        schema.SeverityForStatus(criterion.Status),
			Title:           criterion.CriterionName,
			Explanation:     criterionStatusLabel(criterion.Status),
			Evidence:        criterion.Evidence,
			Suggestion:      criterion.Suggestion,
			CriterionStatus: criterion.Status,
			CriterionScore:  floatPtr(criterion.CriterionScore),
			SelectedRating:  criterion.SelectedRating,
			PredictedPoints: criterion.PredictedPoints,
			MaxPoints:       criterion.MaxPoints,
			Status:          "open",
			SortOrder:       sortOrder,
		}
		sortOrder++
		criterionID := matchCriterionID(criterionIDs, criterion.CriterionName)
		id, err := insertFeedbackItem(ctx, tx, runID, criterionID, item)
		if err != nil {
			return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
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
			return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
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
			return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
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
			return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
		}
		item.ID = id
		feedback = append(feedback, item)
	}

	if err := tx.Commit(ctx); err != nil {
		return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
	}

	return Result{
		RunID:             runID,
		Provider:          ai.Name(),
		Model:             ai.Model(),
		OverallSummary:    out.OverallSummary,
		PredictedScore:    out.PredictedScore,
		PredictedScoreMax: out.PredictedScoreMax,
		Confidence:        out.Confidence,
		Feedback:          feedback,
		CompletedAt:       now,
	}, nil
}

func (s *Service) persistFeedbackFailure(outputSaved bool, runID int64, ai provider.Provider, out *schema.ScoredAnalysis, completedAt time.Time, err error) (Result, error) {
	if !outputSaved {
		return Result{}, err
	}
	return Result{
		RunID:             runID,
		Provider:          ai.Name(),
		Model:             ai.Model(),
		OverallSummary:    out.OverallSummary,
		PredictedScore:    out.PredictedScore,
		PredictedScoreMax: out.PredictedScoreMax,
		Confidence:        out.Confidence,
		CompletedAt:       completedAt,
	}, fmt.Errorf("%w: %w", ErrFeedbackPersistFailed, err)
}

func loadRunResult(ctx context.Context, pool *pgxpool.Pool, runID int64) (*Result, error) {
	var result Result
	var completedAt *time.Time
	err := pool.QueryRow(ctx, `
		SELECT id, COALESCE(provider, ''), COALESCE(model, ''), COALESCE(overall_summary, ''),
			predicted_score, predicted_score_max, COALESCE(confidence, ''), completed_at
		FROM analysis_runs
		WHERE id = $1 AND status = 'completed'
	`, runID).Scan(
		&result.RunID,
		&result.Provider,
		&result.Model,
		&result.OverallSummary,
		&result.PredictedScore,
		&result.PredictedScoreMax,
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

	feedback, err := loadFeedbackItems(ctx, pool, runID)
	if err != nil {
		return nil, err
	}
	result.Feedback = feedback
	return &result, nil
}

func loadLatestResult(ctx context.Context, pool *pgxpool.Pool, assignmentID int64) (*Result, error) {
	var runID int64
	err := pool.QueryRow(ctx, `
		SELECT id
		FROM analysis_runs
		WHERE assignment_snapshot_id = $1 AND status = 'completed'
		ORDER BY completed_at DESC NULLS LAST, id DESC
		LIMIT 1
	`, assignmentID).Scan(&runID)
	if errorsIsNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return loadRunResult(ctx, pool, runID)
}

func loadFeedbackItems(ctx context.Context, pool *pgxpool.Pool, runID int64) ([]FeedbackItem, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, category, severity, title, COALESCE(explanation, ''), COALESCE(evidence, ''),
			COALESCE(suggestion, ''), COALESCE(criterion_status, ''), criterion_score, predicted_points, max_points,
			COALESCE(selected_rating, ''), status, sort_order
		FROM feedback_items
		WHERE analysis_run_id = $1
		ORDER BY sort_order ASC, id ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []FeedbackItem
	for rows.Next() {
		var item FeedbackItem
		var estPoints, maxPoints, criterionScore pgtype.Numeric
		if err := rows.Scan(
			&item.ID,
			&item.Category,
			&item.Severity,
			&item.Title,
			&item.Explanation,
			&item.Evidence,
			&item.Suggestion,
			&item.CriterionStatus,
			&criterionScore,
			&estPoints,
			&maxPoints,
			&item.SelectedRating,
			&item.Status,
			&item.SortOrder,
		); err != nil {
			return nil, err
		}
		item.CriterionScore = numericToFloatPtr(criterionScore)
		item.PredictedPoints = numericToFloatPtr(estPoints)
		item.MaxPoints = numericToFloatPtr(maxPoints)
		items = append(items, item)
	}
	return items, rows.Err()
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
			criterion_status,
			criterion_score,
			predicted_points,
			max_points,
			selected_rating,
			status,
			sort_order
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`, runID, criterionID, item.Category, item.Severity, item.Title, item.Explanation, item.Evidence, item.Suggestion,
		nullIfEmpty(item.CriterionStatus), item.CriterionScore, item.PredictedPoints, item.MaxPoints, nullIfEmpty(item.SelectedRating),
		item.Status, item.SortOrder).Scan(&id)
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

func errorsIsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func nullIfEmpty(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func numericToFloatPtr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}
