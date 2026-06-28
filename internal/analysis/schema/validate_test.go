package schema

import "testing"

func TestValidateProviderResponse(t *testing.T) {
	out := ProviderResponse{
		OverallSummary: "Solid draft with room to improve connections.",
		Confidence:     "medium",
		Criteria: []CriterionAssessment{
			{
				CriterionName:  "Content quality",
				CriterionScore: 0.72,
				Evidence:       "Uses a specific example.",
				Suggestion:     "Connect the example to the rubric language.",
			},
		},
	}

	if err := ValidateProviderResponse(&out); err != nil {
		t.Fatal(err)
	}
}

func TestValidateScoredAnalysis_requiresStatus(t *testing.T) {
	out := ScoredAnalysis{
		OverallSummary: "Summary",
		Confidence:     "medium",
		Criteria: []ScoredCriterion{
			{
				CriterionAssessment: CriterionAssessment{
					CriterionName:  "A",
					CriterionScore: 0.5,
					Evidence:       "e",
					Suggestion:     "s",
				},
			},
		},
	}
	if err := ValidateScoredAnalysis(&out); err == nil {
		t.Fatal("expected missing status error")
	}
}

func TestValidateScoredAnalysis(t *testing.T) {
	out := ScoredAnalysis{
		OverallSummary: "Solid draft with room to improve connections.",
		Confidence:     "medium",
		Criteria: []ScoredCriterion{
			{
				CriterionAssessment: CriterionAssessment{
					CriterionName:  "Content quality",
					CriterionScore: 0.72,
					Evidence:       "Uses a specific example.",
					Suggestion:     "Connect the example to the rubric language.",
				},
				Status: "partially_met",
			},
		},
	}

	if err := ValidateScoredAnalysis(&out); err != nil {
		t.Fatal(err)
	}
}

func TestValidateScoredAnalysis_rejectsScoreAboveMax(t *testing.T) {
	score := 6.0
	max := 5.0
	out := ScoredAnalysis{
		OverallSummary:    "Summary",
		PredictedScore:    &score,
		PredictedScoreMax: &max,
		Confidence:        "medium",
		Criteria: []ScoredCriterion{
			{
				CriterionAssessment: CriterionAssessment{CriterionName: "A", CriterionScore: 1},
				Status:              "met",
			},
		},
	}
	if err := ValidateScoredAnalysis(&out); err == nil {
		t.Fatal("expected score above max error")
	}
}

func TestValidateProviderResponse_rejectsInvalidCriterionScore(t *testing.T) {
	out := ProviderResponse{
		OverallSummary: "Summary",
		Confidence:     "medium",
		Criteria: []CriterionAssessment{
			{CriterionName: "A", CriterionScore: 1.2},
		},
	}
	if err := ValidateProviderResponse(&out); err == nil {
		t.Fatal("expected invalid criterionScore error")
	}
}

func TestValidateProviderResponse_rejectsInvalidConfidence(t *testing.T) {
	out := ProviderResponse{
		OverallSummary: "Summary",
		Confidence:     "very-high",
		Criteria: []CriterionAssessment{
			{CriterionName: "A", CriterionScore: 0.5},
		},
	}
	if err := ValidateProviderResponse(&out); err == nil {
		t.Fatal("expected invalid confidence error")
	}
}
