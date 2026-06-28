package schema

// ScoredAnalysis is the full analysis after server-side rubric scoring.
// Keep aligned with ValidateScoredAnalysis().
type ScoredAnalysis struct {
	OverallSummary      string            `json:"overallSummary"`
	PredictedScore      *float64          `json:"predictedScore,omitempty"`
	PredictedScoreMax   *float64          `json:"predictedScoreMax,omitempty"`
	Confidence          string            `json:"confidence"`
	Criteria            []ScoredCriterion `json:"criteria"`
	MissingRequirements []string          `json:"missingRequirements"`
	Strengths           []string          `json:"strengths"`
	RevisionSuggestions []string          `json:"revisionSuggestions"`
}

// ScoredCriterion combines model judgment with server-derived rubric fields.
type ScoredCriterion struct {
	CriterionAssessment
	Status          string   `json:"status"`
	SelectedRating  string   `json:"selectedRating,omitempty"`
	PredictedPoints *float64 `json:"predictedPoints,omitempty"`
	MaxPoints       *float64 `json:"maxPoints,omitempty"`
}
