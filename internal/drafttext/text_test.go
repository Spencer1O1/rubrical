package drafttext_test

import (
	"testing"

	"rubrical/internal/drafttext"
)

func TestPlainText_stripsTags(t *testing.T) {
	got := drafttext.PlainText(`<p>Hello <em>world</em></p>`)
	if got != "Hello world" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainText_emptyEditorShell(t *testing.T) {
	if drafttext.PlainText(`<p><br></p>`) != "" {
		t.Fatal("expected empty")
	}
	if drafttext.WordCount(`<div><br></div>`) != 0 {
		t.Fatal("expected 0 words")
	}
}

func TestWordCount(t *testing.T) {
	if got := drafttext.WordCount(`<p>one two three</p>`); got != 3 {
		t.Fatalf("got %d", got)
	}
}
