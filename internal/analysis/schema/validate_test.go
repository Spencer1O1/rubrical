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
