package schema

// ProviderResponse is JSON returned by AI providers (OpenAI, Anthropic).
// Keep aligned with JSONSchema() and ValidateProviderResponse().
type ProviderResponse struct {
	OverallSummary      string                `json:"overallSummary"`
	Confidence          string                `json:"confidence"`
	Criteria            []CriterionAssessment `json:"criteria"`
	Strengths           []string              `json:"strengths"`
	Guidance            []string              `json:"guidance"`
}

// CriterionAssessment is one rubric row as judged by the model.
type CriterionAssessment struct {
	CriterionName            string                   `json:"criterionName"`
	SelectedRating           string                   `json:"selectedRating"`
	BandPosition             int                      `json:"bandPosition"`
	ScoreRationale           string                   `json:"scoreRationale"`
	FulfilledRequirements    []FulfilledRequirement   `json:"fulfilledRequirements"`
	UnfulfilledRequirements  []UnfulfilledRequirement `json:"unfulfilledRequirements"`
}

type FulfilledRequirement struct {
	Requirement string `json:"requirement"`
	Evidence    string `json:"evidence"`
}

type UnfulfilledRequirement struct {
	Requirement string `json:"requirement"`
	Severity    string `json:"severity"`
	Explanation string `json:"explanation"`
	Suggestion  string `json:"suggestion"`
}
