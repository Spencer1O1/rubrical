package analysis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRateLimited           = errors.New("analysis rate limit exceeded")
	ErrAnalysisInFlight      = errors.New("analysis already in progress for this assignment")
	ErrFeedbackPersistFailed = errors.New("analysis completed but feedback could not be saved")
)

type RateLimits struct {
	MaxPerHour            int
	MaxPerDay             int
	MinSecondsBetweenRuns int
}

func NewRateLimits(maxPerHour, maxPerDay, minSeconds int) RateLimits {
	return RateLimits{
		MaxPerHour:            maxPerHour,
		MaxPerDay:             maxPerDay,
		MinSecondsBetweenRuns: minSeconds,
	}
}

type Limiter struct {
	pool   *pgxpool.Pool
	limits RateLimits
}

func NewLimiter(pool *pgxpool.Pool, limits RateLimits) *Limiter {
	return &Limiter{pool: pool, limits: limits}
}

func (l *Limiter) checkInFlightInTx(ctx context.Context, tx pgx.Tx, assignmentID int64) error {
	if l == nil {
		return nil
	}

	var inFlight bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM analysis_runs
			WHERE assignment_snapshot_id = $1
			  AND status IN ('pending', 'running')
		)
	`, assignmentID).Scan(&inFlight)
	if err != nil {
		return err
	}
	if inFlight {
		return ErrAnalysisInFlight
	}
	return nil
}

func (l *Limiter) checkRateLimitsInTx(ctx context.Context, tx pgx.Tx, userID, assignmentID int64) error {
	if l == nil {
		return nil
	}

	if l.limits.MaxPerHour > 0 {
		var count int
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM analysis_attempts
			WHERE user_id = $1
			  AND created_at > NOW() - INTERVAL '1 hour'
			  AND status IN ('started', 'completed', 'failed')
		`, userID).Scan(&count)
		if err != nil {
			return err
		}
		if count >= l.limits.MaxPerHour {
			return fmt.Errorf("%w: %d analyses per hour (try again later)", ErrRateLimited, l.limits.MaxPerHour)
		}
	}

	if l.limits.MaxPerDay > 0 {
		var count int
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM analysis_attempts
			WHERE user_id = $1
			  AND created_at > NOW() - INTERVAL '1 day'
			  AND status IN ('started', 'completed', 'failed')
		`, userID).Scan(&count)
		if err != nil {
			return err
		}
		if count >= l.limits.MaxPerDay {
			return fmt.Errorf("%w: %d analyses per day (try again tomorrow)", ErrRateLimited, l.limits.MaxPerDay)
		}
	}

	if l.limits.MinSecondsBetweenRuns > 0 {
		var lastAttempt time.Time
		err := tx.QueryRow(ctx, `
			SELECT COALESCE(MAX(created_at), 'epoch'::timestamptz)
			FROM analysis_attempts
			WHERE assignment_snapshot_id = $1
			  AND status IN ('started', 'completed', 'failed')
		`, assignmentID).Scan(&lastAttempt)
		if err != nil {
			return err
		}
		wait := time.Duration(l.limits.MinSecondsBetweenRuns)*time.Second - time.Since(lastAttempt)
		if wait > 0 {
			secs := int(wait.Seconds())
			if secs < 1 {
				secs = 1
			}
			return fmt.Errorf("%w: wait %d seconds before analyzing again", ErrRateLimited, secs)
		}
	}
	return nil
}

func (l *Limiter) checkInTx(ctx context.Context, tx pgx.Tx, userID, assignmentID int64) error {
	if err := l.checkInFlightInTx(ctx, tx, assignmentID); err != nil {
		return err
	}
	return l.checkRateLimitsInTx(ctx, tx, userID, assignmentID)
}
