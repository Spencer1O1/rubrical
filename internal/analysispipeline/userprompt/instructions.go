package userprompt

import "strings"

// Instructions emits the shared "## Instructions" section for both analysis passes.
func Instructions(text string) string {
	var b strings.Builder
	b.WriteString("## Instructions\n")
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		b.WriteString("(none)\n")
		return b.String()
	}
	b.WriteString(trimmed)
	b.WriteByte('\n')
	return b.String()
}
