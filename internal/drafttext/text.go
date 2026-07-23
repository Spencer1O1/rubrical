package drafttext

import (
	"html"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var strictHTML = bluemonday.StrictPolicy()

// PlainText strips tags for emptiness checks and word counts.
func PlainText(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return strings.TrimSpace(strictHTML.Sanitize(html.UnescapeString(raw)))
}

// WordCount counts words in the visible text of draft HTML/plain body.
func WordCount(raw string) int {
	return len(strings.Fields(PlainText(raw)))
}
