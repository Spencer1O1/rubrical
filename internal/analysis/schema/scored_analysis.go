package schema

// ScoredAnalysis is the full analysis after server-side rubric scoring.
// Keep aligned with ValidateScoredAnalysis().
type ScoredAnalysis struct {
	OverallSummary      string            `json:"overallSummary"`
	PredictedScore      *float64          `json:"predictedScore,omitempty"`
	PredictedScoreMax   *float64          `json:"predictedScoreMax,omitempty"`
	Confidence          string            `json:"confidence"`
	Criteria            []ScoredCriterion `json:"criteria"`
	Strengths           []string          `json:"strengths"`
	Guidance            []string          `json:"guidance"`
}

// ScoredCriterion combines model judgment with server-derived rubric fields.
// CriterionScore is normalized 0–1 for display and persistence.
type ScoredCriterion struct {
	CriterionName           string                   `json:"criterionName"`
	CriterionScore          float64                  `json:"criterionScore"`
	ScoreRationale          string                   `json:"scoreRationale"`
	FulfilledRequirements   []FulfilledRequirement   `json:"fulfilledRequirements"`
	UnfulfilledRequirements []UnfulfilledRequirement `json:"unfulfilledRequirements"`
	Status                  string                   `json:"status"`
	SelectedRating          string                   `json:"selectedRating,omitempty"`
	PredictedPoints         *float64                 `json:"predictedPoints,omitempty"`
	MaxPoints               *float64                 `json:"maxPoints,omitempty"`
}
