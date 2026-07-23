package analyzability

import (
	"fmt"
	"strings"

	"rubrical/internal/analysispipeline/criterionname"
	"rubrical/internal/draftmode"
	"rubrical/internal/llm"
)

// Input is everything pass 1 needs — no student draft evidence.
type Input struct {
	PageType        string
	Instructions    string
	AllowedChannels []string // text, file, url
	Criteria        []criterionname.Ref
}

// Criterion is one row from the analyzability response.
type Criterion struct {
	CriterionID     string `json:"criterionId"`
	CriterionName   string `json:"-"` // resolved from CriterionID
	Analyzable      bool   `json:"analyzable"`
	Reason          string `json:"reason"`
	HowToEarnPoints string `json:"howToEarnPoints"`
}

// Response is the pass-1 model output.
type Response struct {
	Criteria []Criterion `json:"criteria"`
}

func BuildRequest(input Input, providerName string) llm.Request {
	return llm.Request{
		SystemPrompt: SystemPrompt(providerName),
		UserPrompt:   buildUserPrompt(input),
		SchemaName:   "analyzability",
		Schema:       JSONSchema(input.Criteria, providerName),
	}
}

func buildUserPrompt(input Input) string {
	var b strings.Builder

	pageType := strings.TrimSpace(input.PageType)
	if pageType == "" {
		pageType = "assignment"
	}
	b.WriteString("Page type: ")
	b.WriteString(pageType)
	b.WriteString("\n")
	if pageType == "discussion" {
		b.WriteString("Draft context: main topic reply (not a classmate thread reply).\n")
	}
	b.WriteString("Allowed channels: ")
	b.WriteString(formatChannels(input.AllowedChannels))
	b.WriteString("\n\n")

	b.WriteString("## Instructions\n")
	instructions := strings.TrimSpace(input.Instructions)
	if instructions == "" {
		instructions = "(none)"
	}
	b.WriteString(instructions)
	b.WriteString("\n\n")

	b.WriteString("## Criteria\n")
	for i, ref := range input.Criteria {
		b.WriteString(fmt.Sprintf("%d. id=%s — %s\n", i+1, ref.ID, ref.Name))
	}
	return b.String()
}

func formatChannels(channels []string) string {
	return strings.Join(draftmode.PromptLabels(channels), ", ")
}
