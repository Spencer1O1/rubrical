package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const LocalDevEmail = "local@rubrical.dev"

func EnsureLocalUser(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	var userID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, display_name)
		VALUES ($1, 'Local Dev User')
		ON CONFLICT (email) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, LocalDevEmail).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("ensure local user: %w", err)
	}

	if _, err := pool.Exec(ctx, `
		UPDATE assignment_snapshots
		SET user_id = $1
		WHERE user_id IS NULL
	`, userID); err != nil {
		return 0, fmt.Errorf("backfill assignment user_id: %w", err)
	}

	if _, err := pool.Exec(ctx, `
		UPDATE submission_drafts
		SET user_id = $1
		WHERE user_id IS NULL
	`, userID); err != nil {
		return 0, fmt.Errorf("backfill draft user_id: %w", err)
	}

	return userID, nil
}
