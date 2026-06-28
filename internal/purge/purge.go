package purge

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rubrical/internal/draftfiles"
)

// Policy is when draft file bytes are eligible for purge.
type Policy struct {
	PostDueDateRetention time.Duration
	PostUploadRetention  time.Duration
}

func (p Policy) enabled() bool {
	return p.PostDueDateRetention > 0 || p.PostUploadRetention > 0
}

// PurgeDraftFiles deletes uploaded draft file bytes and rows past retention.
func PurgeDraftFiles(
	ctx context.Context,
	pool *pgxpool.Pool,
	files *draftfiles.Store,
	policy Policy,
) (int, error) {
	if !policy.enabled() {
		return 0, nil
	}

	rows, err := pool.Query(ctx, `
		DELETE FROM submission_draft_files f
		USING submission_drafts d, assignment_snapshots a
		WHERE f.submission_draft_id = d.id
		  AND d.assignment_snapshot_id = a.id
		  AND (
		    ($1::bigint > 0
		      AND a.due_at IS NOT NULL
		      AND a.due_at + ($1::bigint * INTERVAL '1 second') < NOW())
		    OR
		    ($2::bigint > 0
		      AND a.due_at IS NULL
		      AND f.uploaded_at + ($2::bigint * INTERVAL '1 second') < NOW())
		  )
		RETURNING f.file_storage_key
	`, int64(policy.PostDueDateRetention.Seconds()), int64(policy.PostUploadRetention.Seconds()))
	if err != nil {
		return 0, fmt.Errorf("delete draft files: %w", err)
	}
	defer rows.Close()

	purged := 0
	for rows.Next() {
		var storageKey string
		if err := rows.Scan(&storageKey); err != nil {
			return purged, fmt.Errorf("scan storage key: %w", err)
		}
		if err := files.Delete(storageKey); err != nil {
			log.Printf("purge: delete blob %q: %v", storageKey, err)
			continue
		}
		purged++
	}
	if err := rows.Err(); err != nil {
		return purged, err
	}

	return purged, nil
}

// RunBackground purges on interval until ctx is cancelled. Runs once immediately.
func RunBackground(
	ctx context.Context,
	pool *pgxpool.Pool,
	files *draftfiles.Store,
	policy Policy,
	interval time.Duration,
) {
	if !policy.enabled() {
		return
	}
	if interval <= 0 {
		interval = time.Hour
	}

	go func() {
		run := func() {
			n, err := PurgeDraftFiles(ctx, pool, files, policy)
			if err != nil {
				log.Printf("purge: %v", err)
				return
			}
			if n > 0 {
				log.Printf("purge: removed %d draft file(s)", n)
			}
		}

		run()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}
