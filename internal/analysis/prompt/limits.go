package prompt

import (
	"rubrical/internal/config"
	"strings"
)

const DefaultMaxSubmissionTextChars = config.DefaultAnalysisMaxSubmissionTextChars

func normalizeMaxSubmissionTextChars(max int) int {
	if max <= 0 {
		return config.DefaultAnalysisMaxSubmissionTextChars
	}
	return max
}

func truncate(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max] + "\n…[truncated]"
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
	out := truncate(text, b.remaining)
	b.remaining -= len(out)
	if b.remaining < 0 {
		b.remaining = 0
	}
	return out
}
