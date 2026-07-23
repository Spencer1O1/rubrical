package pages

import (
	"strings"
	"testing"
)

func TestSanitizedDraftHTML_keepsFormatting(t *testing.T) {
	raw := `<p>Hello <strong>world</strong></p><script>alert(1)</script>`
	got := SanitizedDraftHTML(raw)
	if !strings.Contains(got, "<strong>world</strong>") {
		t.Fatalf("expected formatting kept, got %q", got)
	}
	if strings.Contains(got, "<script") {
		t.Fatalf("expected script stripped, got %q", got)
	}
}

func TestDraftPlainText_stripsTags(t *testing.T) {
	got := DraftPlainText(`<p>Hello <em>world</em></p>`)
	if got != "Hello world" {
		t.Fatalf("got %q", got)
	}
}

func TestDraftPlainText_emptyEditorShell(t *testing.T) {
	if DraftPlainText(`<p><br></p>`) != "" {
		t.Fatal("expected empty plain text")
	}
	if DraftWordCount(`<div><br></div>`) != 0 {
		t.Fatal("expected zero words")
	}
}

func TestDraftWordCount(t *testing.T) {
	if got := DraftWordCount(`<p>one two three</p>`); got != 3 {
		t.Fatalf("got %d", got)
	}
}
