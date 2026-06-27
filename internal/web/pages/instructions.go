package pages

import (
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var instructionsHTMLPolicy = bluemonday.UGCPolicy()

var instructionTablePattern = regexp.MustCompile(`(?is)(<table\b[^>]*>.*?</table>)`)

// decodeInstructionHTML runs once. Canvas ENV may ship entity-encoded markup;
// DOM innerHTML does not. Extension normalizes before import.
func decodeInstructionHTML(raw string) string {
	return html.UnescapeString(strings.TrimSpace(raw))
}

func SanitizedInstructionsHTML(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return wrapInstructionTables(instructionsHTMLPolicy.Sanitize(decodeInstructionHTML(raw)))
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
