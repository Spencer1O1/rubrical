package prompt

import (
	_ "embed"
	"strings"
)

//go:embed system.md
var systemPrompt string

func BuildSystem() string {
	return strings.TrimSpace(systemPrompt)
}
