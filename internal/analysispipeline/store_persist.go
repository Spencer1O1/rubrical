package analysispipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rubrical/internal/analysispipeline/analysis"
	analysisschema "rubrical/internal/analysispipeline/analysis/schema"
	"rubrical/internal/llm"
)

func (s *Service) persistSuccess(ctx context.Context, handle RunHandle, assignmentID int64, ai llm.Provider, out *analysisschema.ScoredAnalysis) (Result, error) {
	runID := handle.RunID
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
	if len(out.Criteria) != len(criterionIDs) {
		return Result{}, fmt.Errorf(
			"rubric criterion count mismatch: analysis has %d, database has %d",
			len(out.Criteria), len(criterionIDs),
		)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		DELETE FROM analysis_runs
		WHERE assignment_snapshot_id = $1
		  AND id <> $2
	`, assignmentID, runID); err != nil {
		return Result{}, err
	}

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

	for i, criterion := range out.Criteria {
		explanation := criterionStatusLabel(criterion.Status)
		if criterion.Status == "not_analyzable" && criterion.HowToEarnPoints != "" {
			explanation = criterion.HowToEarnPoints
		}
		item := FeedbackItem{
			Category:                "criterion",
			Severity:                analysisschema.SeverityForStatus(criterion.Status),
			Title:                   criterion.CriterionName,
			Explanation:             explanation,
			ScoreRationale:          criterion.ScoreRationale,
			HowToEarnPoints:         criterion.HowToEarnPoints,
			FulfilledRequirements:   append([]analysisschema.FulfilledRequirement(nil), criterion.FulfilledRequirements...),
			UnfulfilledRequirements: append([]analysisschema.UnfulfilledRequirement(nil), criterion.UnfulfilledRequirements...),
			CriterionStatus:         criterion.Status,
			CriterionScore:          analysis.FloatPtr(criterion.CriterionScore),
			SelectedRating:          criterion.SelectedRating,
			PredictedPoints:         criterion.PredictedPoints,
			MaxPoints:               criterion.MaxPoints,
			Status:                  "open",
			SortOrder:               sortOrder,
		}
		sortOrder++
		criterionID := criterionIDAt(criterionIDs, i)
		id, err := insertFeedbackItem(ctx, tx, runID, criterionID, item)
		if err != nil {
			return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
		}
		item.ID = id
		feedback = append(feedback, item)
	}

	for _, strength := range out.Strengths {
		item := FeedbackItem{
			Category:  "strength",
			Severity:  "info",
			Title:     strength,
			Status:    "open",
			SortOrder: sortOrder,
		}
		sortOrder++
		id, err := insertFeedbackItem(ctx, tx, runID, nil, item)
		if err != nil {
			return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
		}
		item.ID = id
		feedback = append(feedback, item)
	}

	for _, item := range out.Guidance {
		feedbackItem := FeedbackItem{
			Category:  "guidance",
			Severity:  "info",
			Title:     item,
			Status:    "open",
			SortOrder: sortOrder,
		}
		sortOrder++
		id, err := insertFeedbackItem(ctx, tx, runID, nil, feedbackItem)
		if err != nil {
			return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
		}
		feedbackItem.ID = id
		feedback = append(feedback, feedbackItem)
	}

	if handle.AttemptID > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE analysis_attempts
			SET status = 'completed', completed_at = $2
			WHERE id = $1 AND status = 'started'
		`, handle.AttemptID, now); err != nil {
			return s.persistFeedbackFailure(outputSaved, runID, ai, out, now, err)
		}
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

func (s *Service) persistFeedbackFailure(outputSaved bool, runID int64, ai llm.Provider, out *analysisschema.ScoredAnalysis, completedAt time.Time, err error) (Result, error) {
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

