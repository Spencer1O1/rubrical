package pages

import (
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	ugcHTMLPolicy    = bluemonday.UGCPolicy()
	strictHTMLPolicy = bluemonday.StrictPolicy()
)

var instructionTablePattern = regexp.MustCompile(`(?is)(<table\b[^>]*>.*?</table>)`)

// decodeHTMLEntities runs once. Canvas ENV may ship entity-encoded markup;
// DOM innerHTML does not. Extension normalizes before import.
func decodeHTMLEntities(raw string) string {
	return html.UnescapeString(strings.TrimSpace(raw))
}

// SanitizeUGCHTML strips unsafe tags/attrs with bluemonday's UGC policy.
func SanitizeUGCHTML(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return ugcHTMLPolicy.Sanitize(decodeHTMLEntities(raw))
}

func SanitizedInstructionsHTML(raw string) string {
	sanitized := SanitizeUGCHTML(raw)
	if sanitized == "" {
		return ""
	}
	return wrapInstructionTables(sanitized)
}

// SanitizedDraftHTML prepares draft body HTML for storage and rich-text display.
func SanitizedDraftHTML(raw string) string {
	return SanitizeUGCHTML(raw)
}

// DraftPlainText strips tags for emptiness checks and word counts.
func DraftPlainText(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return strings.TrimSpace(strictHTMLPolicy.Sanitize(decodeHTMLEntities(raw)))
}

// DraftWordCount counts words in the visible text of draft HTML/plain body.
func DraftWordCount(raw string) int {
	return len(strings.Fields(DraftPlainText(raw)))
}

func wrapInstructionTables(html string) string {
	if html == "" || !strings.Contains(strings.ToLower(html), "<table") {
		return html
	}
	if strings.Contains(html, "assignment-instructions-table-scroll") {
		return html
	}
	return instructionTablePattern.ReplaceAllString(html, `<div class="assignment-instructions-table-scroll">$1</div>`)
}

func PrepareInstructionsHTML(instructions string) string {
	source := strings.TrimSpace(instructions)
	if source == "" {
		return ""
	}
	return SanitizedInstructionsHTML(source)
}

func InstructionsHasHTML(raw string) bool {
	return strings.TrimSpace(raw) != ""
}
