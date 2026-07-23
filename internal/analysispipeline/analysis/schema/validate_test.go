package schema

import "testing"

func sampleCriterion() CriterionAssessment {
	return CriterionAssessment{
		CriterionID:      "content-quality",
		SelectedRatingID: "r1",
		BandPosition:     72,
		ScoreRationale:   "The draft meets the good band with a specific example, but the rubric connection could be sharper.",
		FulfilledRequirements: []FulfilledRequirement{{
			Requirement: "Uses a specific example",
			Evidence:    "The draft describes a live orchestra performance in detail.",
		}},
		UnfulfilledRequirements: []UnfulfilledRequirement{{
			Requirement: "Connects the example to rubric language",
			Severity:    "medium",
			Explanation: "The example is vivid but not tied back to the criterion wording.",
			Suggestion:  "Add a sentence linking the performance details to the rubric's analysis requirement.",
		}},
	}
}

func TestValidateProviderResponse(t *testing.T) {
	out := ProviderResponse{
		OverallSummary: "Solid draft with room to improve connections.",
		Confidence:     "medium",
		Criteria:       []CriterionAssessment{sampleCriterion()},
	}

	if err := ValidateProviderResponse(&out); err != nil {
		t.Fatal(err)
	}
}

func TestValidateScoredAnalysis_requiresStatus(t *testing.T) {
	criterion := sampleCriterion()
	out := ScoredAnalysis{
		OverallSummary: "Summary",
		Confidence:     "medium",
		Criteria: []ScoredCriterion{{
			CriterionName:           "Content quality",
			CriterionScore:          0.5,
			ScoreRationale:          criterion.ScoreRationale,
			FulfilledRequirements:   criterion.FulfilledRequirements,
			UnfulfilledRequirements: criterion.UnfulfilledRequirements,
		}},
	}
	if err := ValidateScoredAnalysis(&out); err == nil {
		t.Fatal("expected missing status error")
	}
}

func TestValidateScoredAnalysis(t *testing.T) {
	criterion := sampleCriterion()
	out := ScoredAnalysis{
		OverallSummary: "Solid draft with room to improve connections.",
		Confidence:     "medium",
		Criteria: []ScoredCriterion{{
			CriterionName:           "Content quality",
			CriterionScore:          0.72,
			Status:                  "partially_met",
			ScoreRationale:          criterion.ScoreRationale,
			FulfilledRequirements:   criterion.FulfilledRequirements,
			UnfulfilledRequirements: criterion.UnfulfilledRequirements,
		}},
	}

	if err := ValidateScoredAnalysis(&out); err != nil {
		t.Fatal(err)
	}
}

func TestValidateProviderResponse_rejectsInvalidBandPosition(t *testing.T) {
	criterion := sampleCriterion()
	criterion.BandPosition = 101
	out := ProviderResponse{
		OverallSummary: "Summary",
		Confidence:     "medium",
		Criteria:       []CriterionAssessment{criterion},
	}
	if err := ValidateProviderResponse(&out); err == nil {
		t.Fatal("expected invalid bandPosition error")
	}
}

func TestValidateProviderResponse_rejectsVacuousSuggestion(t *testing.T) {
	criterion := sampleCriterion()
	criterion.UnfulfilledRequirements[0].Suggestion = "None needed."
	out := ProviderResponse{
		OverallSummary: "Summary",
		Confidence:     "medium",
		Criteria:       []CriterionAssessment{criterion},
	}
	if err := ValidateProviderResponse(&out); err == nil {
		t.Fatal("expected vacuous suggestion error")
	}
}

func TestValidateProviderResponse_rejectsEmptyScoreRationale(t *testing.T) {
	criterion := sampleCriterion()
	criterion.ScoreRationale = ""
	out := ProviderResponse{
		OverallSummary: "Summary",
		Confidence:     "medium",
		Criteria:       []CriterionAssessment{criterion},
	}
	if err := ValidateProviderResponse(&out); err == nil {
		t.Fatal("expected missing scoreRationale error")
	}
}

func TestValidateProviderResponse_rejectsInvalidConfidence(t *testing.T) {
	out := ProviderResponse{
		OverallSummary: "Summary",
		Confidence:     "very-high",
		Criteria:       []CriterionAssessment{sampleCriterion()},
	}
	if err := ValidateProviderResponse(&out); err == nil {
		t.Fatal("expected invalid confidence error")
	}
}

func TestValidateProviderResponse_dropsFulfilledOverlappingUnfulfilled(t *testing.T) {
	criterion := sampleCriterion()
	dup := "Offers a deep, thoughtful reflection on intellectual and emotional responses."
	criterion.FulfilledRequirements = []FulfilledRequirement{{
		Requirement: dup,
		Evidence:    "Questions whether the story reflects the concluding message.",
	}}
	criterion.UnfulfilledRequirements = []UnfulfilledRequirement{{
		Requirement: dup,
		Severity:    "medium",
		Explanation: "The emotional response could be more fully developed.",
		Suggestion:  "Expand on how the piece made you feel.",
	}}
	out := ProviderResponse{
		OverallSummary: "Partial reflection.",
		Confidence:     "medium",
		Criteria:       []CriterionAssessment{criterion},
	}
	if err := ValidateProviderResponse(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Criteria[0].FulfilledRequirements) != 0 {
		t.Fatalf("expected overlapping fulfilled dropped, got %+v", out.Criteria[0].FulfilledRequirements)
	}
	if len(out.Criteria[0].UnfulfilledRequirements) != 1 {
		t.Fatalf("expected gap kept, got %+v", out.Criteria[0].UnfulfilledRequirements)
	}
}

func TestValidateProviderResponse_acceptsEmptyRequirementArrays(t *testing.T) {
	out := ProviderResponse{
		OverallSummary: "Summary",
		Confidence:     "medium",
		Criteria: []CriterionAssessment{{
			CriterionID:             "a",
			SelectedRatingID:        "r1",
			BandPosition:            68,
			ScoreRationale:          "Strong overall fit for the good band.",
			FulfilledRequirements:   []FulfilledRequirement{},
			UnfulfilledRequirements: []UnfulfilledRequirement{},
		}},
	}
	if err := ValidateProviderResponse(&out); err != nil {
		t.Fatal(err)
	}
}

func TestValidateScoredAnalysis_notAnalyzable(t *testing.T) {
	pts := 0.0
	max := 1.0
	out := ScoredAnalysis{
		OverallSummary: "Could not check participation.",
		Confidence:     "medium",
		Strengths:      []string{},
		Guidance:       []string{},
		Criteria: []ScoredCriterion{{
			CriterionName:   "Participation",
			Status:          "not_analyzable",
			ScoreRationale:  "Requires in-class observation.",
			HowToEarnPoints: "Participate in class as required.",
			MaxPoints:       &max,
		}},
		PredictedScore:    &pts,
		PredictedScoreMax: &pts,
	}
	if err := ValidateScoredAnalysis(&out); err != nil {
		t.Fatal(err)
	}
}
