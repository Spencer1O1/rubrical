package prompt

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateRunes_suffixOutsideBudget(t *testing.T) {
	content := strings.Repeat("é", 10) // 10 runes, 20 bytes
	out, used := truncateRunes(content, 5)
	if used != 5 {
		t.Fatalf("used=%d want 5", used)
	}
	if !strings.HasSuffix(out, truncateSuffix) {
		t.Fatalf("expected suffix, got %q", out)
	}
	contentRunes := utf8.RuneCountInString(strings.TrimSuffix(out, truncateSuffix))
	if contentRunes != 5 {
		t.Fatalf("content runes=%d want 5", contentRunes)
	}
}

func TestTextBudget_takeCountsRunes(t *testing.T) {
	budget := newTextBudget(5)
	got := budget.take(strings.Repeat("é", 10))
	if utf8.RuneCountInString(strings.TrimSuffix(got, truncateSuffix)) != 5 {
		t.Fatalf("got %q", got)
	}
	if budget.remaining != 0 {
		t.Fatalf("remaining=%d want 0", budget.remaining)
	}
}

func TestManifestBudget_separateFromText(t *testing.T) {
	textBudget := newTextBudget(10)
	manifestBudget := newManifestBudget(5)

	textBudget.take(strings.Repeat("a", 10))
	tree := manifestBudget.take(strings.Repeat("b", 10))

	if textBudget.remaining != 0 {
		t.Fatalf("text remaining=%d", textBudget.remaining)
	}
	if !strings.HasSuffix(tree, truncateSuffix) {
		t.Fatalf("manifest tree not truncated: %q", tree)
	}
}
