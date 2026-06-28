package schema

import "testing"

func TestValidate(t *testing.T) {
	score := 4.0
	max := 5.0
	out := ModelOutput{
		OverallSummary:    "Solid draft with room to improve connections.",
		EstimatedScore:    &score,
		EstimatedScoreMax: &max,
		Confidence:        "medium",
		Criteria: []CriterionFeedback{
			{
				CriterionName: "Content quality",
				Status:        "partially_met",
				Evidence:      "Uses a specific example.",
				Suggestion:    "Connect the example to the rubric language.",
			},
		},
	}

	if err := Validate(&out); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_rejectsScoreAboveMax(t *testing.T) {
	score := 6.0
	max := 5.0
	out := ModelOutput{
		OverallSummary:    "Summary",
		EstimatedScore:    &score,
		EstimatedScoreMax: &max,
		Confidence:        "medium",
		Criteria: []CriterionFeedback{
			{CriterionName: "A", Status: "met"},
		},
	}
	if err := Validate(&out); err == nil {
		t.Fatal("expected score above max error")
	}
}

func TestValidate_rejectsCriterionPointsAboveMax(t *testing.T) {
	points := 4.0
	max := 3.0
	out := ModelOutput{
		OverallSummary: "Summary",
		Confidence:     "medium",
		Criteria: []CriterionFeedback{
			{
				CriterionName:   "A",
				Status:          "met",
				EstimatedPoints: &points,
				MaxPoints:       &max,
			},
		},
	}
	if err := Validate(&out); err == nil {
		t.Fatal("expected criterion points above max error")
	}
}

func TestValidate_rejectsInvalidConfidence(t *testing.T) {
	out := ModelOutput{
		OverallSummary: "Summary",
		Confidence:     "very-high",
		Criteria: []CriterionFeedback{
			{CriterionName: "A", Status: "met"},
		},
	}
	if err := Validate(&out); err == nil {
		t.Fatal("expected invalid confidence error")
	}
}
