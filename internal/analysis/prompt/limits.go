package prompt

import (
	"rubrical/internal/config"
	"strings"
	"unicode/utf8"
)

const (
	DefaultMaxSubmissionTextChars = config.DefaultAnalysisMaxSubmissionTextChars
	truncateSuffix                = "\n…[truncated]"
)

func normalizeMaxSubmissionTextChars(max int) int {
	if max <= 0 {
		return config.DefaultAnalysisMaxSubmissionTextChars
	}
	return max
}

func normalizeMaxManifestChars(max int) int {
	if max <= 0 {
		return config.DefaultAnalysisMaxManifestChars
	}
	return max
}

// truncateRunes returns up to maxRunes of content. When truncated, the suffix is
// appended outside the budget (only content runes are charged to the caller).
func truncateRunes(value string, maxRunes int) (string, int) {
	if maxRunes <= 0 {
		return "", 0
	}
	if utf8.RuneCountInString(value) <= maxRunes {
		return value, utf8.RuneCountInString(value)
	}
	runes := []rune(value)
	return string(runes[:maxRunes]) + truncateSuffix, maxRunes
}

// textBudget is a single shared pool for student submission text (draft body + inline file extracts).
type textBudget struct {
	remaining int
}

func newTextBudget(max int) textBudget {
	return textBudget{remaining: normalizeMaxSubmissionTextChars(max)}
}

func (b *textBudget) take(text string) string {
	text = strings.TrimSpace(text)
	if text == "" || b.remaining <= 0 {
		return ""
	}
	out, used := truncateRunes(text, b.remaining)
	b.remaining -= used
	if b.remaining < 0 {
		b.remaining = 0
	}
	return out
}

// manifestBudget caps manifest trees and skipped-file notes (separate from submission text).
type manifestBudget struct {
	remaining int
}

func newManifestBudget(max int) manifestBudget {
	return manifestBudget{remaining: normalizeMaxManifestChars(max)}
}

func (b *manifestBudget) take(text string) string {
	text = strings.TrimSpace(text)
	if text == "" || b.remaining <= 0 {
		return ""
	}
	out, used := truncateRunes(text, b.remaining)
	b.remaining -= used
	if b.remaining < 0 {
		b.remaining = 0
	}
	return out
}
