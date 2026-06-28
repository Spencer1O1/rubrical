package schema

// ModelOutput is the contract every AI provider must return as JSON.
// Keep aligned with JSONSchema() and Validate.
type ModelOutput struct {
	OverallSummary      string              `json:"overallSummary"`
	EstimatedScore      *float64            `json:"estimatedScore"`
	EstimatedScoreMax   *float64            `json:"estimatedScoreMax"`
	Confidence          string              `json:"confidence"`
	Criteria            []CriterionFeedback `json:"criteria"`
	MissingRequirements []string            `json:"missingRequirements"`
	Strengths           []string            `json:"strengths"`
	RevisionSuggestions []string            `json:"revisionSuggestions"`
}

type CriterionFeedback struct {
	CriterionName   string   `json:"criterionName"`
	Status          string   `json:"status"`
	EstimatedPoints *float64 `json:"estimatedPoints"`
	MaxPoints       *float64 `json:"maxPoints"`
	Evidence        string   `json:"evidence"`
	Suggestion      string   `json:"suggestion"`
}
