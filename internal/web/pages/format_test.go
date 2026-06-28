package pages

import "testing"

func TestDraftFilesSavedMessage(t *testing.T) {
	if got := DraftFilesSavedMessage([]string{"report.pdf"}); got != "File saved: report.pdf" {
		t.Fatalf("single file: got %q", got)
	}

	got := DraftFilesSavedMessage([]string{"report.pdf", "study.png"})
	want := "Files saved: report.pdf, study.png"
	if got != want {
		t.Fatalf("multiple files: got %q want %q", got, want)
	}
}

func TestDraftLinkSavedMessage(t *testing.T) {
	if got := DraftLinkSavedMessage(); got != "Link saved" {
		t.Fatalf("got %q", got)
	}
}
