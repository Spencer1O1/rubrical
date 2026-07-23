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
