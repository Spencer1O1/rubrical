package analysispipeline

import (
	"context"
	"encoding/json"
	"time"

	analysisschema "rubrical/internal/analysispipeline/analysis/schema"
)

func (s *Service) beginRun(ctx context.Context, userID, assignmentID, draftID int64, provider, model string, inputLog []byte) (RunHandle, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RunHandle{}, err
	}
	defer tx.Rollback(ctx)

	if err := failStaleRuns(ctx, s.pool, assignmentID, DefaultStaleRunTTL); err != nil {
		return RunHandle{}, err
	}

	if s.limiter != nil {
		if err := s.limiter.checkInFlightInTx(ctx, tx, assignmentID); err != nil {
			return RunHandle{}, err
		}
		if s.enforceLimits {
			if err := s.limiter.checkRateLimitsInTx(ctx, tx, userID, assignmentID); err != nil {
				return RunHandle{}, err
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
		return RunHandle{}, err
	}

	var attemptID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO analysis_attempts (
			user_id,
			assignment_snapshot_id,
			analysis_run_id,
			status
		) VALUES ($1, $2, $3, 'started')
		RETURNING id
	`, userID, assignmentID, runID).Scan(&attemptID)
	if err != nil {
		return RunHandle{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return RunHandle{}, err
	}
	return RunHandle{RunID: runID, AttemptID: attemptID}, nil
}

func (s *Service) updateRunPromptLog(ctx context.Context, runID int64, inputLog []byte) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE analysis_runs
		SET raw_model_input = $2::jsonb
		WHERE id = $1
	`, runID, string(inputLog))
	return err
}

func (s *Service) saveScoredAnalysis(ctx context.Context, runID int64, out *analysisschema.ScoredAnalysis) error {
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

func (s *Service) markRunFailed(ctx context.Context, handle RunHandle, runErr error) error {
	payload, _ := json.Marshal(map[string]string{"error": runErr.Error()})
	now := time.Now().UTC()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE analysis_runs
		SET status = 'failed', raw_model_output = $2::jsonb, completed_at = $3
		WHERE id = $1
	`, handle.RunID, string(payload), now); err != nil {
		return err
	}
	if handle.AttemptID > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE analysis_attempts
			SET status = 'failed', completed_at = $2
			WHERE id = $1 AND status = 'started'
		`, handle.AttemptID, now); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

