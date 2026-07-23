package analysispipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func insertFeedbackItem(ctx context.Context, tx pgx.Tx, runID int64, criterionID *int64, item FeedbackItem) (int64, error) {
	fulfilledJSON, err := json.Marshal(item.FulfilledRequirements)
	if err != nil {
		return 0, fmt.Errorf("encode fulfilled requirements: %w", err)
	}
	unfulfilledJSON, err := json.Marshal(item.UnfulfilledRequirements)
	if err != nil {
		return 0, fmt.Errorf("encode unfulfilled requirements: %w", err)
	}
	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO feedback_items (
			analysis_run_id,
			rubric_criterion_id,
			category,
			severity,
			title,
			explanation,
			score_rationale,
			fulfilled_requirements,
			unfulfilled_requirements,
			criterion_status,
			criterion_score,
			predicted_points,
			max_points,
			selected_rating,
			status,
			sort_order
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id
	`, runID, criterionID, item.Category, item.Severity, item.Title, item.Explanation, nullIfEmpty(item.ScoreRationale),
		fulfilledJSON, unfulfilledJSON,
		nullIfEmpty(item.CriterionStatus), item.CriterionScore, item.PredictedPoints, item.MaxPoints, nullIfEmpty(item.SelectedRating),
		item.Status, item.SortOrder).Scan(&id)
	return id, err
}

// loadCriterionIDs returns rubric_criteria.id in sort_order (same order as analysis criteria).
func loadCriterionIDs(ctx context.Context, pool *pgxpool.Pool, assignmentID int64) ([]int64, error) {
	rows, err := pool.Query(ctx, `
		SELECT id
		FROM rubric_criteria
		WHERE assignment_snapshot_id = $1
		ORDER BY sort_order
	`, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func criterionIDAt(ids []int64, index int) *int64 {
	if index < 0 || index >= len(ids) {
		return nil
	}
	id := ids[index]
	return &id
}
