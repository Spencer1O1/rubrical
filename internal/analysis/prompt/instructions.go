package prompt

import "strings"

func BuildInstructions(instructions string, maxChars int) string {
	var b strings.Builder
	b.WriteString("\n## Instructions\n")
	b.WriteString(truncate(strings.TrimSpace(instructions), maxChars))
	b.WriteByte('\n')
	return b.String()
}
