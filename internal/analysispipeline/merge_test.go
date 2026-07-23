package analysispipeline

import (
	"testing"

	"rubrical/internal/analysispipeline/analysis"
	analysisschema "rubrical/internal/analysispipeline/analysis/schema"
	"rubrical/internal/analysispipeline/analyzability"
)

func TestMergeAnalysis_classmateNotAnalyzable(t *testing.T) {
	rubric := analysis.RubricContext{Rows: []analysis.RubricRow{
		{Criterion: "Topic Response", Ratings: []analysis.RubricRating{{Title: "Full", Points: "1"}, {Title: "None", Points: "0"}}},
		{Criterion: "Classmate Reply", Ratings: []analysis.RubricRating{{Title: "Full", Points: "1"}, {Title: "None", Points: "0"}}},
	}}
	refs := rubric.AssignCriterionIDs()
	class := &analyzability.Response{Criteria: []analyzability.Criterion{
		{CriterionID: refs[0].ID, Analyzable: true, Reason: "text draft"},
		{CriterionID: refs[1].ID, Analyzable: false, Reason: "peer reply not in this draft", HowToEarnPoints: "Write a thoughtful classmate reply in Canvas."},
	}}
	scored := &analysisschema.ScoredAnalysis{
		OverallSummary: "Solid topic post.",
		Confidence:     "high",
		Strengths:      []string{},
		Guidance:       []string{},
		Criteria: []analysisschema.ScoredCriterion{{
			CriterionID:     refs[0].ID,
			CriterionName:   "Topic Response",
			CriterionScore:  1,
			ScoreRationale:  "Clear response.",
			Status:          "met",
			SelectedRating:  "Full",
			PredictedPoints: analysis.FloatPtr(1),
			MaxPoints:       analysis.FloatPtr(1),
		}},
		PredictedScore:    analysis.FloatPtr(1),
		PredictedScoreMax: analysis.FloatPtr(1),
	}

	merged, err := MergeAnalysis(class, scored, rubric)
	if err != nil {
		t.Fatal(err)
	}
	if err := analysisschema.ValidateScoredAnalysis(merged); err != nil {
		t.Fatal(err)
	}
	if len(merged.Criteria) != 2 {
		t.Fatalf("criteria = %d", len(merged.Criteria))
	}
	if merged.Criteria[1].Status != "not_analyzable" {
		t.Fatalf("status = %q", merged.Criteria[1].Status)
	}
	if merged.Criteria[1].HowToEarnPoints == "" {
		t.Fatal("expected howToEarnPoints")
	}
	if merged.PredictedScoreMax == nil || *merged.PredictedScoreMax != 1 {
		t.Fatalf("checkable max = %v", merged.PredictedScoreMax)
	}
}

func TestMergeAnalysis_missingPhotoStillAnalyzableViaScoring(t *testing.T) {
	rubric := analysis.RubricContext{Rows: []analysis.RubricRow{
		{Criterion: "Photo Evidence", Ratings: []analysis.RubricRating{{Title: "Full", Points: "2"}, {Title: "None", Points: "0"}}},
	}}
	refs := rubric.AssignCriterionIDs()
	class := &analyzability.Response{Criteria: []analyzability.Criterion{
		{CriterionID: refs[0].ID, Analyzable: true, Reason: "image upload expected"},
	}}
	scored := &analysisschema.ScoredAnalysis{
		OverallSummary: "Missing photo.",
		Confidence:     "high",
		Strengths:      []string{},
		Guidance:       []string{"Upload the required photo."},
		Criteria: []analysisschema.ScoredCriterion{{
			CriterionID:     refs[0].ID,
			CriterionName:   "Photo Evidence",
			CriterionScore:  0,
			ScoreRationale:  "No photo attached.",
			Status:          "not_met",
			SelectedRating:  "None",
			PredictedPoints: analysis.FloatPtr(0),
			MaxPoints:       analysis.FloatPtr(2),
		}},
		PredictedScore:    analysis.FloatPtr(0),
		PredictedScoreMax: analysis.FloatPtr(2),
	}
	merged, err := MergeAnalysis(class, scored, rubric)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Criteria[0].Status != "not_met" {
		t.Fatalf("status = %q, want not_met", merged.Criteria[0].Status)
	}
}

func TestFilterRubric_dropsNotAnalyzable(t *testing.T) {
	rubric := analysis.RubricContext{Rows: []analysis.RubricRow{
		{Criterion: "A"},
		{Criterion: "B"},
	}}
	class := &analyzability.Response{Criteria: []analyzability.Criterion{
		{CriterionID: "a", Analyzable: true, Reason: "ok"},
		{CriterionID: "b", Analyzable: false, Reason: "live", HowToEarnPoints: "Attend."},
	}}
	filtered := filterRubric(rubric, class)
	if len(filtered.Rows) != 1 || filtered.Rows[0].Criterion != "A" {
		t.Fatalf("filtered = %+v", filtered.Rows)
	}
}

func TestMergeAnalysis_preservesIdsThroughFilteredPass2(t *testing.T) {
	rubric := analysis.RubricContext{Rows: []analysis.RubricRow{
		{Criterion: "Content", Ratings: []analysis.RubricRating{{Title: "Full", Points: "1"}, {Title: "None", Points: "0"}}},
		{Criterion: "Content", Ratings: []analysis.RubricRating{{Title: "Full", Points: "1"}, {Title: "None", Points: "0"}}},
	}}
	refs := rubric.AssignCriterionIDs()
	class := &analyzability.Response{Criteria: []analyzability.Criterion{
		{CriterionID: refs[0].ID, Analyzable: false, Reason: "skip", HowToEarnPoints: "N/A"},
		{CriterionID: refs[1].ID, Analyzable: true, Reason: "ok"},
	}}
	filtered := filterRubric(rubric, class)
	filtered.AssignCriterionIDs() // must not rewrite content-2 → content
	scored := &analysisschema.ScoredAnalysis{
		OverallSummary: "ok",
		Confidence:     "high",
		Strengths:      []string{},
		Guidance:       []string{},
		Criteria: []analysisschema.ScoredCriterion{{
			CriterionID:     filtered.Rows[0].ID,
			CriterionName:   "Content",
			CriterionScore:  1,
			ScoreRationale:  "Good.",
			Status:          "met",
			SelectedRating:  "Full",
			PredictedPoints: analysis.FloatPtr(1),
			MaxPoints:       analysis.FloatPtr(1),
		}},
	}
	merged, err := MergeAnalysis(class, scored, rubric)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Criteria[1].Status != "met" || merged.Criteria[1].CriterionID != "content-2" {
		t.Fatalf("merged = %+v", merged.Criteria[1])
	}
}

func TestMergeAnalysis_allNotAnalyzableSkipsPass2(t *testing.T) {
	rubric := analysis.RubricContext{Rows: []analysis.RubricRow{
		{Criterion: "Participation", Ratings: []analysis.RubricRating{{Title: "Full", Points: "1"}, {Title: "None", Points: "0"}}},
	}}
	refs := rubric.AssignCriterionIDs()
	class := &analyzability.Response{Criteria: []analyzability.Criterion{
		{CriterionID: refs[0].ID, Analyzable: false, Reason: "in class", HowToEarnPoints: "Participate in class."},
	}}
	merged, err := MergeAnalysis(class, nil, rubric)
	if err != nil {
		t.Fatal(err)
	}
	if err := analysisschema.ValidateScoredAnalysis(merged); err != nil {
		t.Fatal(err)
	}
	if merged.Criteria[0].Status != "not_analyzable" {
		t.Fatal(merged.Criteria[0].Status)
	}
}
