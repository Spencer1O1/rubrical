package handlers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"rubrical/internal/auth"
	"rubrical/internal/db"
	"rubrical/internal/draftfiles"
	"rubrical/internal/draftmode"
	"rubrical/internal/importpayload"
)

func TestSaveDiscussionDraftFromImportStoresTextAndOneFile(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/discussion_topics/discussion-test")

	payload := importpayload.Payload{
		PageType:  "discussion",
		DraftText: "My reply text",
		DraftKind: "text",
		DraftFiles: []importpayload.DraftFile{{
			FileName:      "resume-1.pdf",
			MimeType:      "application/pdf",
			ContentBase64: "JVBERi0xLjQK",
		}},
	}
	if err := h.saveDraftFromImport(ctx, assignmentID, payload); err != nil {
		t.Fatal(err)
	}

	var body, mode string
	err := pool.QueryRow(ctx, `
		SELECT body, draft_mode FROM submission_drafts
		WHERE assignment_snapshot_id = $1 AND user_id = $2
		ORDER BY updated_at DESC LIMIT 1
	`, assignmentID, h.userID).Scan(&body, &mode)
	if err != nil {
		t.Fatal(err)
	}
	if body != "My reply text" {
		t.Fatalf("body=%q", body)
	}
	if mode != draftmode.Text {
		t.Fatalf("mode=%q, want text (discussion keeps text mode with attachment)", mode)
	}

	var fileName string
	err = pool.QueryRow(ctx, `
		SELECT f.source_file_name
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
	`, assignmentID, h.userID).Scan(&fileName)
	if err != nil {
		t.Fatal(err)
	}
	if fileName != "resume-1.pdf" {
		t.Fatalf("fileName=%q", fileName)
	}
}

func TestSaveDiscussionDraftFromImportClearsAttachmentWhenEmpty(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/discussion_topics/clear-attachment")

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		PageType:  "discussion",
		DraftText: "with file",
		DraftFiles: []importpayload.DraftFile{{
			FileName:      "notes.txt",
			MimeType:      "text/plain",
			ContentBase64: "aGVsbG8=",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		PageType:  "discussion",
		DraftText: "text only",
	}); err != nil {
		t.Fatal(err)
	}

	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
	`, assignmentID, h.userID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("file count=%d, want 0 after re-import without attachment", count)
	}
}

func TestSaveDiscussionDraftFromImportReusesServerFileRef(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/discussion_topics/ref-reuse")

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		PageType:  "discussion",
		DraftText: "first import",
		DraftFiles: []importpayload.DraftFile{{
			FileName:      "resume.pdf",
			MimeType:      "application/pdf",
			ContentBase64: "JVBERi0xLjQK",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	var serverFileID int64
	err := pool.QueryRow(ctx, `
		SELECT f.id
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
	`, assignmentID, h.userID).Scan(&serverFileID)
	if err != nil {
		t.Fatal(err)
	}

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		PageType:  "discussion",
		DraftText: "second import",
		DraftFileRefs: []importpayload.DraftFileRef{{
			ServerFileID: serverFileID,
			FileName:     "resume.pdf",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
	`, assignmentID, h.userID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("file count=%d, want 1 after ref-only re-import", count)
	}
}

func TestSaveDraftFromImportStoresCanvasFileID(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/assignments/canvas-id-store")

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftKind: "file",
		DraftFiles: []importpayload.DraftFile{{
			FileName:      "actions.txt",
			MimeType:      "text/plain",
			ContentBase64: "aGVsbG8=",
			CanvasFileID:  "99543999",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	var canvasFileID string
	err := pool.QueryRow(ctx, `
		SELECT COALESCE(f.canvas_file_id, '')
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
	`, assignmentID, h.userID).Scan(&canvasFileID)
	if err != nil {
		t.Fatal(err)
	}
	if canvasFileID != "99543999" {
		t.Fatalf("canvas_file_id=%q", canvasFileID)
	}
}

func TestSaveDraftFromImportClearsEmptyCanvasDraft(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/assignments/clear-test")

	if err := h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
		Mode:       draftmode.Text,
		Body:       "from canvas",
		SourceType: "canvas_text_entry",
		FromCanvas: true,
	}); err != nil {
		t.Fatal(err)
	}
	payload := importpayload.Payload{DraftText: ""}
	if err := h.saveDraftFromImport(ctx, assignmentID, payload); err != nil {
		t.Fatal(err)
	}

	var body string
	var wordCount int
	err := pool.QueryRow(ctx, `
		SELECT body, word_count FROM submission_drafts
		WHERE assignment_snapshot_id = $1 AND user_id = $2
		ORDER BY updated_at DESC LIMIT 1
	`, assignmentID, h.userID).Scan(&body, &wordCount)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" || wordCount != 0 {
		t.Fatalf("canvas draft not cleared: body=%q words=%d", body, wordCount)
	}
}

func TestSaveDraftFromImportClearsStoredFilesWhenCanvasCaptureEmpty(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/assignments/preserve-file-test")

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftFiles: []importpayload.DraftFile{{
			FileName:      "notes.txt",
			MimeType:      "text/plain",
			ContentBase64: "aGVsbG8gd29ybGQ=",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftKind: "file",
	}); err != nil {
		t.Fatal(err)
	}

	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
	`, assignmentID, h.userID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("file count=%d, want 0 after empty canvas file import", count)
	}
}

func TestSaveDraftFromImportStoresMultipleFiles(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/assignments/file-test")

	payload := importpayload.Payload{
		DraftFiles: []importpayload.DraftFile{
			{
				FileName:      "notes.txt",
				MimeType:      "text/plain",
				ContentBase64: "aGVsbG8gd29ybGQ=",
			},
			{
				FileName:      "archive.zip",
				MimeType:      "application/zip",
				ContentBase64: "UEsFBg==",
			},
		},
	}
	if err := h.saveDraftFromImport(ctx, assignmentID, payload); err != nil {
		t.Fatal(err)
	}

	rows, err := pool.Query(ctx, `
		SELECT f.source_file_name, f.file_storage_key, d.draft_mode
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
		ORDER BY f.sort_order ASC
	`, assignmentID, h.userID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var names []string
	var mode string
	for rows.Next() {
		var name, storageKey string
		if err := rows.Scan(&name, &storageKey, &mode); err != nil {
			t.Fatal(err)
		}
		if storageKey == "" {
			t.Fatal("expected storage key")
		}
		names = append(names, name)
	}
	if len(names) != 2 {
		t.Fatalf("names=%v", names)
	}
	if names[0] != "notes.txt" || names[1] != "archive.zip" {
		t.Fatalf("names=%v", names)
	}
	if mode != draftmode.File {
		t.Fatalf("mode=%q", mode)
	}
}

func TestSwitchDraftModeClearsOtherFields(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/assignments/mode-test")

	if err := h.upsertLatestDraft(ctx, assignmentID, draftUpsertOptions{
		Mode:       draftmode.Text,
		Body:       "hello",
		SourceType: "manual_paste",
	}); err != nil {
		t.Fatal(err)
	}
	if err := h.switchDraftMode(ctx, assignmentID, draftmode.File); err != nil {
		t.Fatal(err)
	}

	var body string
	var mode string
	err := pool.QueryRow(ctx, `
		SELECT body, draft_mode FROM submission_drafts
		WHERE assignment_snapshot_id = $1 AND user_id = $2
	`, assignmentID, h.userID).Scan(&body, &mode)
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		t.Fatalf("expected empty body after switch to file, got %q", body)
	}
	if mode != draftmode.File {
		t.Fatalf("mode=%q", mode)
	}
}

func TestSaveDraftFromImportRefsOnlyPreservesBytes(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/assignments/refs-only")

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftFiles: []importpayload.DraftFile{{
			FileName:      "notes.txt",
			MimeType:      "text/plain",
			ContentBase64: "aGVsbG8gd29ybGQ=",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	var serverFileID int64
	err := pool.QueryRow(ctx, `
		SELECT f.id
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
	`, assignmentID, h.userID).Scan(&serverFileID)
	if err != nil {
		t.Fatal(err)
	}

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftKind: "file",
		DraftFileRefs: []importpayload.DraftFileRef{{
			ServerFileID: serverFileID,
			FileName:     "notes.txt",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND d.user_id = $2
	`, assignmentID, h.userID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("file count=%d", count)
	}
}

func TestSaveDraftFromImportPrunesOrphanedServerFile(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	h := testHandler(t, pool)

	assignmentID := insertAssignment(t, pool, h.userID, "https://school.instructure.com/courses/1/assignments/prune-orphan")

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftFiles: []importpayload.DraftFile{
			{FileName: "keep.txt", MimeType: "text/plain", ContentBase64: "aGVsbG8="},
			{FileName: "drop.txt", MimeType: "text/plain", ContentBase64: "Ynll"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	var keepID int64
	err := pool.QueryRow(ctx, `
		SELECT f.id FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1 AND f.source_file_name = 'keep.txt'
	`, assignmentID).Scan(&keepID)
	if err != nil {
		t.Fatal(err)
	}

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftKind: "file",
		DraftFileRefs: []importpayload.DraftFileRef{{
			ServerFileID: keepID,
			FileName:     "keep.txt",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM submission_draft_files f
		JOIN submission_drafts d ON d.id = f.submission_draft_id
		WHERE d.assignment_snapshot_id = $1
	`, assignmentID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("file count=%d", count)
	}
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	database, err := db.Connect(context.Background(), "postgres://rubrical:rubrical@localhost:5432/rubrical?sslmode=disable")
	if err != nil {
		t.Skipf("database unavailable: %v", err)
	}
	t.Cleanup(database.Close)
	return database.Pool
}

func testHandler(t *testing.T, pool *pgxpool.Pool) *Handlers {
	t.Helper()
	userID, err := auth.EnsureLocalUser(context.Background(), pool)
	if err != nil {
		t.Fatalf("local user: %v", err)
	}
	files, err := draftfiles.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return &Handlers{
		db:          &db.DB{Pool: pool},
		files:       files,
		userID:      userID,
		importLimits: importpayload.DefaultLimits(),
	}
}

func insertAssignment(t *testing.T, pool *pgxpool.Pool, userID int64, sourceURL string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO assignment_snapshots (user_id, source_url, assignment_title)
		VALUES ($1, $2, 'Clear test')
		ON CONFLICT (user_id, source_url) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, userID, sourceURL).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM assignment_snapshots WHERE id = $1`, id)
	})
	return id
}
