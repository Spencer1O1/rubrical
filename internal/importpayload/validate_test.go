package importpayload

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestValidateAndNormalize(t *testing.T) {
	valid := Payload{
		SourceURL: "https://school.instructure.com/courses/1/assignments/2",
		PageType:  "assignment",
		Title:     "Essay",
		Metadata: Metadata{
			DueDateText:        "Due Jun 26 at 11:59pm",
			PointsPossibleText: "25 pts",
		},
		CapturedAt: time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC),
	}
	if err := ValidateAndNormalize(&valid, DefaultLimits()); err != nil {
		t.Fatalf("valid payload: %v", err)
	}
	if valid.SourceURL != "https://school.instructure.com/courses/1/assignments/2" {
		t.Fatalf("normalized url = %q", valid.SourceURL)
	}
}

func TestValidateAndNormalizeRejectsBadSourceURL(t *testing.T) {
	payload := Payload{
		SourceURL: "https://example.com/courses/1/assignments/2",
		PageType:  "assignment",
	}
	if err := ValidateAndNormalize(&payload, DefaultLimits()); err == nil {
		t.Fatal("expected invalid sourceUrl error")
	}
}

func TestValidateAndNormalizeRejectsDuplicateDraftFileRef(t *testing.T) {
	payload := Payload{
		SourceURL: "https://school.instructure.com/courses/1/assignments/2",
		PageType:  "assignment",
		DraftFileRefs: []DraftFileRef{
			{ServerFileID: 1, FileName: "a.txt"},
			{ServerFileID: 1, FileName: "b.txt"},
		},
	}
	if err := ValidateAndNormalize(&payload, DefaultLimits()); err == nil {
		t.Fatal("expected duplicate serverFileId error")
	}
}

func TestValidateAndNormalizeRejectsOversizedDraftFile(t *testing.T) {
	payload := Payload{
		SourceURL: "https://school.instructure.com/courses/1/assignments/2",
		PageType:  "assignment",
		DraftFiles: []DraftFile{{
			FileName:      "big.txt",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", DefaultLimits().MaxUploadBytes+1))),
		}},
	}
	if err := ValidateAndNormalize(&payload, DefaultLimits()); err == nil {
		t.Fatal("expected draft file size error")
	}
}
