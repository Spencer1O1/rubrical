package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"rubrical/internal/importpayload"
)

func TestDraftManifestEmptyWhenNeverImported(t *testing.T) {
	pool := testPool(t)
	h, userID := testHandler(t, pool)

	req := httptest.NewRequest(http.MethodGet, "/assignments/draft-manifest?sourceUrl="+urlQueryEscape("https://school.instructure.com/courses/1/assignments/never"), nil)
	req = req.WithContext(testCtx(userID))
	rec := httptest.NewRecorder()

	h.DraftManifest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload draftManifestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Files) != 0 {
		t.Fatalf("files=%v", payload.Files)
	}
}

func TestDraftManifestReturnsStoredFiles(t *testing.T) {
	pool := testPool(t)
	h, userID := testHandler(t, pool)
	ctx := testCtx(userID)

	sourceURL := "https://school.instructure.com/courses/1/assignments/manifest-test"
	assignmentID := insertAssignment(t, pool, userID, sourceURL)

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftFiles: []importpayload.DraftFile{{
			FileName:      "report.pdf",
			MimeType:      "application/pdf",
			ContentBase64: "JVBERi0xLjQK",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/assignments/draft-manifest?sourceUrl="+urlQueryEscape(sourceURL), nil)
	req = req.WithContext(testCtx(userID))
	rec := httptest.NewRecorder()

	h.DraftManifest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload draftManifestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.AssignmentID != assignmentID {
		t.Fatalf("assignmentId=%d", payload.AssignmentID)
	}
	if len(payload.Files) != 1 {
		t.Fatalf("files=%v", payload.Files)
	}
	if payload.Files[0].FileName != "report.pdf" {
		t.Fatalf("fileName=%q", payload.Files[0].FileName)
	}
	if payload.Files[0].ServerFileID == 0 {
		t.Fatal("expected serverFileId")
	}
}

func TestDraftManifestReturnsCanvasFileID(t *testing.T) {
	pool := testPool(t)
	h, userID := testHandler(t, pool)
	ctx := testCtx(userID)

	sourceURL := "https://school.instructure.com/courses/1/assignments/manifest-canvas-id"
	assignmentID := insertAssignment(t, pool, userID, sourceURL)

	if err := h.saveDraftFromImport(ctx, assignmentID, importpayload.Payload{
		DraftFiles: []importpayload.DraftFile{{
			FileName:      "resume.pdf",
			MimeType:      "application/pdf",
			ContentBase64: "JVBERi0xLjQK",
			CanvasFileID:  "99543121",
		}},
	}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/assignments/draft-manifest?sourceUrl="+urlQueryEscape(sourceURL), nil)
	req = req.WithContext(testCtx(userID))
	rec := httptest.NewRecorder()

	h.DraftManifest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var payload draftManifestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Files) != 1 {
		t.Fatalf("files=%v", payload.Files)
	}
	if payload.Files[0].CanvasFileID != "99543121" {
		t.Fatalf("canvasFileId=%q", payload.Files[0].CanvasFileID)
	}
}

func urlQueryEscape(value string) string {
	return url.QueryEscape(value)
}
