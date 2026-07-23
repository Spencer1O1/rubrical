package analysispipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"rubrical/internal/analysispipeline/analysis"
)

func loadRubricContext(ctx context.Context, pool *pgxpool.Pool, assignmentID int64) (analysis.RubricContext, error) {
	var rubric analysis.RubricContext

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

		row := analysis.RubricRow{
			Criterion:                name,
			CriterionLongDescription: description,
		}
		if points > 0 {
			row.Points = fmt.Sprintf("%.2f", points)
		}

		if len(ratingsJSON) > 0 {
			var payload struct {
				Criterion                string                  `json:"criterion"`
				CriterionLongDescription string                  `json:"criterionLongDescription"`
				Ratings                  []analysis.RubricRating `json:"ratings"`
				Points                   string                  `json:"points"`
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

