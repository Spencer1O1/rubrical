package analysispipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
		SELECT id, category, severity, title, COALESCE(explanation, ''), COALESCE(score_rationale, ''),
			COALESCE(fulfilled_requirements, '[]'::jsonb), COALESCE(unfulfilled_requirements, '[]'::jsonb),
			COALESCE(criterion_status, ''), criterion_score, predicted_points, max_points,
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
		var fulfilledJSON, unfulfilledJSON []byte
		if err := rows.Scan(
			&item.ID,
			&item.Category,
			&item.Severity,
			&item.Title,
			&item.Explanation,
			&item.ScoreRationale,
			&fulfilledJSON,
			&unfulfilledJSON,
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
		if len(fulfilledJSON) > 0 {
			if err := json.Unmarshal(fulfilledJSON, &item.FulfilledRequirements); err != nil {
				return nil, fmt.Errorf("decode fulfilled requirements: %w", err)
			}
		}
		if item.CriterionStatus == "not_analyzable" {
			item.HowToEarnPoints = item.Explanation
		}
		if len(unfulfilledJSON) > 0 {
			if err := json.Unmarshal(unfulfilledJSON, &item.UnfulfilledRequirements); err != nil {
				return nil, fmt.Errorf("decode unfulfilled requirements: %w", err)
			}
		}
		item.CriterionScore = numericToFloatPtr(criterionScore)
		item.PredictedPoints = numericToFloatPtr(estPoints)
		item.MaxPoints = numericToFloatPtr(maxPoints)
		items = append(items, item)
	}
	return items, rows.Err()
}

