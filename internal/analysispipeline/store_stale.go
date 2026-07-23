package analysispipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultStaleRunTTL = 10 * time.Minute

func failStaleRuns(ctx context.Context, pool *pgxpool.Pool, assignmentID int64, ttl time.Duration) error {
	if pool == nil || ttl <= 0 {
		return nil
	}
	payload, _ := json.Marshal(map[string]string{
		"error": "analysis timed out or did not finish",
	})
	rows, err := pool.Query(ctx, `
		UPDATE analysis_runs
		SET status = 'failed',
		    raw_model_output = $3::jsonb,
		    completed_at = NOW()
		WHERE assignment_snapshot_id = $1
		  AND status IN ('pending', 'running')
		  AND created_at < NOW() - $2::interval
		RETURNING id
	`, assignmentID, fmt.Sprintf("%f seconds", ttl.Seconds()), string(payload))
	if err != nil {
		return err
	}
	defer rows.Close()

	var staleRunIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		staleRunIDs = append(staleRunIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return markAttemptsFailed(ctx, pool, staleRunIDs)
}

func FailAllStaleRuns(ctx context.Context, pool *pgxpool.Pool, ttl time.Duration) error {
	if pool == nil || ttl <= 0 {
		return nil
	}
	payload, _ := json.Marshal(map[string]string{
		"error": "analysis timed out or did not finish",
	})
	rows, err := pool.Query(ctx, `
		UPDATE analysis_runs
		SET status = 'failed',
		    raw_model_output = $2::jsonb,
		    completed_at = NOW()
		WHERE status IN ('pending', 'running')
		  AND created_at < NOW() - $1::interval
		RETURNING id
	`, fmt.Sprintf("%f seconds", ttl.Seconds()), string(payload))
	if err != nil {
		return err
	}
	defer rows.Close()

	var staleRunIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		staleRunIDs = append(staleRunIDs, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return markAttemptsFailed(ctx, pool, staleRunIDs)
}

func markAttemptsFailed(ctx context.Context, pool *pgxpool.Pool, runIDs []int64) error {
	if len(runIDs) == 0 {
		return nil
	}
	_, err := pool.Exec(ctx, `
		UPDATE analysis_attempts
		SET status = 'failed',
		    completed_at = NOW()
		WHERE analysis_run_id = ANY($1)
		  AND status = 'started'
	`, runIDs)
	return err
}

