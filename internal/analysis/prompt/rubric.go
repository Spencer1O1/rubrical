package prompt

import (
	"encoding/json"
	"strings"
)

func BuildRubric(rubric Rubric) string {
	var b strings.Builder
	b.WriteString("\n## Rubric\n")
	b.WriteString(formatRubric(rubric))
	return b.String()
}

func formatRubric(rubric Rubric) string {
	if len(rubric.Rows) == 0 {
		return "(no rubric extracted)\n"
	}

	payload := map[string]any{
		"header": rubric.Header,
		"rows":   rubric.Rows,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "(rubric unavailable)\n"
	}
	return string(data) + "\n"
}
