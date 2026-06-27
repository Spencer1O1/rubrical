package pages

import (
	"strings"
	"testing"
)

func TestSanitizedInstructionsHTML_rendersParagraphs(t *testing.T) {
	raw := `<p><span>Hello </span><strong>world</strong></p>`
	got := SanitizedInstructionsHTML(raw)
	if !strings.Contains(got, "<p>") || !strings.Contains(got, "<strong>world</strong>") {
		t.Fatalf("expected sanitized HTML paragraphs, got %q", got)
	}
}

func TestSanitizedInstructionsHTML_decodesEntities(t *testing.T) {
	raw := `&lt;p&gt;&lt;span&gt;Hello&lt;/span&gt;&lt;/p&gt;`
	got := SanitizedInstructionsHTML(raw)
	if !strings.Contains(got, "<p>") || !strings.Contains(got, "Hello") {
		t.Fatalf("expected decoded HTML, got %q", got)
	}
}

func TestPrepareInstructionsHTML_fromInstructionsText(t *testing.T) {
	raw := `<p>From instructions_text</p>`
	got := PrepareInstructionsHTML(raw)
	if !strings.Contains(got, "From instructions_text") {
		t.Fatalf("expected instructions text, got %q", got)
	}
}

func TestPrepareInstructionsHTML_wrapsWideTables(t *testing.T) {
	raw := `<p>Intro</p><table><tr><td>Visual Art</td><td>Long column content here</td></tr></table>`
	got := PrepareInstructionsHTML(raw)
	if !strings.Contains(got, `assignment-instructions-table-scroll`) {
		t.Fatalf("expected table scroll wrapper, got %q", got)
	}
	if !strings.Contains(got, "<table>") {
		t.Fatalf("expected table preserved, got %q", got)
	}
}

func TestPrepareInstructionsHTML_emptyWhenMissing(t *testing.T) {
	got := PrepareInstructionsHTML("")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestNormalizeRubricRatings_dropsPointsOnlyDuplicates(t *testing.T) {
	got := NormalizeRubricRatings([]RubricRatingView{
		{Title: "Excellent", Description: "Clearly lists event name.", Points: "3 pts"},
		{Points: "3 pts"},
		{Title: "Good", Description: "Missing one detail.", Points: "1.5 pts"},
		{Points: "1.5 pts"},
		{Title: "A", Description: "one"},
		{Title: "B", Description: "two"},
		{Title: "C", Description: "three"},
		{Title: "D", Description: "four"},
	})
	if len(got) != 6 {
		t.Fatalf("expected 6 ratings, got %d", len(got))
	}
}
