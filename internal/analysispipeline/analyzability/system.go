package analyzability

import (
	_ "embed"
	"strings"

	"rubrical/internal/analysispipeline/files"
	"rubrical/internal/analysispipeline/userprompt"
	"rubrical/internal/draftmode"
)

//go:embed system.md
var systemPromptTemplate string

func SystemPrompt(provider string, pageType string, allowedChannels []string) string {
	channels := strings.Join(draftmode.PromptLabels(allowedChannels), ", ")
	caps := files.FormatPromptCapabilities(provider)
	s := strings.TrimSpace(systemPromptTemplate)
	s = strings.ReplaceAll(s, "{{DRAFT_CONTEXT}}", userprompt.DraftContextLabel(pageType))
	s = strings.ReplaceAll(s, "{{CHANNELS}}", channels)
	s = strings.ReplaceAll(s, "{{CAPABILITIES}}", caps)
	return s
}
