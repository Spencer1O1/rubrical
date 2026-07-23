package prompt

import (
	_ "embed"
	"strings"

	"rubrical/internal/analysispipeline/userprompt"
)

//go:embed system.md
var systemPromptTemplate string

func BuildSystem(pageType string) string {
	s := strings.TrimSpace(systemPromptTemplate)
	return strings.ReplaceAll(s, "{{DRAFT_CONTEXT}}", userprompt.DraftContextLabel(pageType))
}
