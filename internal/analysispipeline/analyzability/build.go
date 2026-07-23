package analyzability

import (
	"fmt"
	"strings"

	"rubrical/internal/analysispipeline/criterion"
	"rubrical/internal/analysispipeline/userprompt"
	"rubrical/internal/llm"
)

// Input is everything pass 1 needs — no student draft evidence.
type Input struct {
	PageType        string
	Instructions    string
	AllowedChannels []string // text, file, url — assignment-allowed; injected into system prompt
	Criteria        []criterion.Ref
}

// Criterion is one row from the analyzability response.
type Criterion struct {
	CriterionID        string `json:"criterionId"`
	CriterionName      string `json:"-"` // resolved from CriterionID
	EvidenceProvidable bool   `json:"evidenceProvidable"`
	EvidenceAnalyzable bool   `json:"evidenceAnalyzable"`
	Reason             string `json:"reason"`
	HowToEarnPoints    string `json:"howToEarnPoints"`
}

// Checkable is true when evidence can be provided here and Rubrical can inspect it.
func (c Criterion) Checkable() bool {
	return c.EvidenceProvidable && c.EvidenceAnalyzable
}

// Response is the pass-1 model output.
type Response struct {
	Criteria []Criterion `json:"criteria"`
}

func BuildRequest(input Input, providerName string) llm.Request {
	return llm.Request{
		SystemPrompt: SystemPrompt(providerName, input.PageType, input.AllowedChannels),
		UserPrompt:   buildUserPrompt(input),
		SchemaName:   "analyzability",
		Schema:       JSONSchema(input.Criteria, providerName),
	}
}

func buildUserPrompt(input Input) string {
	var b strings.Builder
	b.WriteString(userprompt.Instructions(input.Instructions))
	b.WriteByte('\n')
	b.WriteString("# Criteria\n")
	for i, ref := range input.Criteria {
		b.WriteString(fmt.Sprintf("%d. id=%s — %s\n", i+1, ref.ID, ref.Name))
		if desc := strings.TrimSpace(ref.Description); desc != "" {
			b.WriteString(desc)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
