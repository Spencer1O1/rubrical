package analyzability

import (
	_ "embed"
	"strings"

	"rubrical/internal/analysispipeline/files"
	"rubrical/internal/draftmode"
)

//go:embed system.md
var systemPromptTemplate string

func SystemPrompt(provider string) string {
	channels := strings.Join(draftmode.PromptLabels(nil), " / ")
	caps := files.FormatPromptCapabilities(provider)
	s := strings.TrimSpace(systemPromptTemplate)
	s = strings.ReplaceAll(s, "{{CHANNELS}}", channels)
	s = strings.ReplaceAll(s, "{{CAPABILITIES}}", caps)
	return s
}
