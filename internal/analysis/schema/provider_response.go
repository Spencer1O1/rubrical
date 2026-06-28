package schema

// ProviderResponse is JSON returned by AI providers (OpenAI, Anthropic).
// Keep aligned with JSONSchema() and ValidateProviderResponse().
type ProviderResponse struct {
	OverallSummary      string                `json:"overallSummary"`
	Confidence          string                `json:"confidence"`
	Criteria            []CriterionAssessment `json:"criteria"`
	MissingRequirements []string              `json:"missingRequirements"`
	Strengths           []string              `json:"strengths"`
	RevisionSuggestions []string              `json:"revisionSuggestions"`
}

// CriterionAssessment is one rubric row as judged by the model.
type CriterionAssessment struct {
	CriterionName  string  `json:"criterionName"`
	CriterionScore float64 `json:"criterionScore"`
	Evidence       string  `json:"evidence"`
	Suggestion     string  `json:"suggestion"`
}
