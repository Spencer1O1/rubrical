package prompt

import "strings"

func BuildInstructions(instructions string) string {
	var b strings.Builder
	b.WriteString("\n## Instructions\n")
	text := strings.TrimSpace(instructions)
	if text == "" {
		b.WriteString("(none)\n")
		return b.String()
	}
	b.WriteString(text)
	b.WriteByte('\n')
	return b.String()
}
