package purge

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rubrical/internal/draftfiles"
)

func TestPurgeDraftFiles_removesFilesPastDuePlusRetention(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	userID := testUserID(t, pool)

	files, err := draftfiles.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	var assignmentID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO assignment_snapshots (user_id, source_url, assignment_title, due_at)
		VALUES ($1, $2, 'Due date purge', NOW() - INTERVAL '10 days')
		RETURNING id
	`, userID, "https://school.instructure.com/courses/1/assignments/due-purge-test").Scan(&assignmentID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM assignment_snapshots WHERE id = $1`, assignmentID)
	})

	storageKey := insertDraftFile(t, ctx, pool, files, userID, assignmentID, "essay.pdf", []byte("%PDF-1.4\ndraft bytes"))

	n, err := PurgeDraftFiles(ctx, pool, files, Policy{PostDueDateRetention: 7 * 24 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("purged = %d", n)
	}

	assertBlobGone(t, files, storageKey)
	assertDraftFileCount(t, ctx, pool, assignmentID, 0)
}

func TestPurgeDraftFiles_removesFilesWithoutDueAtAfterUploadRetention(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	userID := testUserID(t, pool)

	files, err := draftfiles.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	var assignmentID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO assignment_snapshots (user_id, source_url, assignment_title, due_at)
		VALUES ($1, $2, 'No due date', NULL)
		RETURNING id
	`, userID, "https://school.instructure.com/courses/1/assignments/upload-purge-test").Scan(&assignmentID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM assignment_snapshots WHERE id = $1`, assignmentID)
	})

	storageKey := insertDraftFile(t, ctx, pool, files, userID, assignmentID, "notes.txt", []byte("keep"), time.Now().Add(-31*24*time.Hour))

	n, err := PurgeDraftFiles(ctx, pool, files, Policy{PostUploadRetention: 30 * 24 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("purged = %d", n)
	}

	assertBlobGone(t, files, storageKey)
	assertDraftFileCount(t, ctx, pool, assignmentID, 0)
}

func TestPurgeDraftFiles_keepsRecentUploadWhenNoDueAt(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	userID := testUserID(t, pool)

	files, err := draftfiles.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	var assignmentID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO assignment_snapshots (user_id, source_url, assignment_title, due_at)
		VALUES ($1, $2, 'No due date recent', NULL)
		RETURNING id
	`, userID, "https://school.instructure.com/courses/1/assignments/recent-upload-test").Scan(&assignmentID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM assignment_snapshots WHERE id = $1`, assignmentID)
	})

	insertDraftFile(t, ctx, pool, files, userID, assignmentID, "notes.txt", []byte("keep"), time.Now())

	n, err := PurgeDraftFiles(ctx, pool, files, Policy{
		PostDueDateRetention: 7 * 24 * time.Hour,
		PostUploadRetention:  30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("purged = %d", n)
	}

	assertDraftFileCount(t, ctx, pool, assignmentID, 1)
}

func insertDraftFile(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	files *draftfiles.Store,
	userID, assignmentID int64,
	name string,
	data []byte,
	uploadedAt ...time.Time,
) string {
	t.Helper()

	var draftID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO submission_drafts (assignment_snapshot_id, user_id, body, draft_mode)
		VALUES ($1, $2, '', 'file')
		RETURNING id
	`, assignmentID, userID).Scan(&draftID)
	if err != nil {
		t.Fatal(err)
	}

	storageKey, err := files.Save(userID, assignmentID, name, data)
	if err != nil {
		t.Fatal(err)
	}

	if len(uploadedAt) > 0 {
		_, err = pool.Exec(ctx, `
			INSERT INTO submission_draft_files (submission_draft_id, source_file_name, file_storage_key, uploaded_at)
			VALUES ($1, $2, $3, $4)
		`, draftID, name, storageKey, uploadedAt[0])
	} else {
		_, err = pool.Exec(ctx, `
			INSERT INTO submission_draft_files (submission_draft_id, source_file_name, file_storage_key)
			VALUES ($1, $2, $3)
		`, draftID, name, storageKey)
	}
	if err != nil {
		t.Fatal(err)
	}

	return storageKey
}

func assertBlobGone(t *testing.T, files *draftfiles.Store, storageKey string) {
	t.Helper()
	if _, err := os.Stat(files.Path(storageKey)); !os.IsNotExist(err) {
		t.Fatalf("blob still on disk: %v", err)
	}
}

func assertDraftFileCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, assignmentID int64, want int) {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1
	`, assignmentID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("file count = %d, want %d", count, want)
	}
}

