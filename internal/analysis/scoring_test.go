package analysis

import (
	"testing"

	"rubrical/internal/analysis/schema"
)

func TestCriterionStatusFromScore_banded(t *testing.T) {
	cases := []struct {
		score float64
		want  string
	}{
		{0.24, "not_met"},
		{0.25, "partially_met"},
		{0.74, "partially_met"},
		{0.75, "met"},
	}
	for _, tc := range cases {
		if got := criterionStatusFromScore(tc.score, 4); got != tc.want {
			t.Fatalf("score %.2f -> %q, want %q", tc.score, got, tc.want)
		}
	}
}

func TestCriterionStatusFromScore_pointOnly(t *testing.T) {
	if got := criterionStatusFromScore(0.65, 0); got != "partially_met" {
		t.Fatalf("got %q", got)
	}
}

func TestBandIndex_equalSlices(t *testing.T) {
	n := 4
	cases := []struct {
		score float64
		want  int
	}{
		{0, 0},
		{0.24, 0},
		{0.25, 1},
		{0.5, 2},
		{0.74, 2},
		{0.75, 3},
		{1, 3},
	}
	for _, tc := range cases {
		if got := bandIndex(tc.score, n); got != tc.want {
			t.Fatalf("score %.2f -> band %d, want %d", tc.score, got, tc.want)
		}
	}
}

func TestApplyRubricScoring_topBandAt075(t *testing.T) {
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionName: "General Overview", CriterionScore: 0.75,
	}}}
	rubric := RubricContext{Rows: []RubricRow{{
		Criterion: "General Overview",
		Ratings: []RubricRating{
			{Title: "Poor", Points: "2 pts"},
			{Title: "Fair", Points: "4 pts"},
			{Title: "Good", Points: "6 pts"},
			{Title: "Excellent", Points: "8 pts"},
		},
	}}}
	scored, err := ApplyRubricScoring(resp, rubric)
	if err != nil {
		t.Fatal(err)
	}
	if scored.Criteria[0].SelectedRating != "Excellent" {
		t.Fatalf("rating = %q, want Excellent", scored.Criteria[0].SelectedRating)
	}
	if scored.Criteria[0].Status != "met" {
		t.Fatalf("status = %q, want met", scored.Criteria[0].Status)
	}
	if *scored.Criteria[0].PredictedPoints != 8 {
		t.Fatalf("points = %v", scored.Criteria[0].PredictedPoints)
	}
}

func TestApplyRubricScoring_pointOnlyRow(t *testing.T) {
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionName: "Participation", CriterionScore: 0.65,
	}}}
	rubric := RubricContext{Rows: []RubricRow{{Criterion: "Participation", Points: "10 pts"}}}
	scored, err := ApplyRubricScoring(resp, rubric)
	if err != nil {
		t.Fatal(err)
	}
	if *scored.Criteria[0].PredictedPoints != 6.5 {
		t.Fatalf("points = %v", scored.Criteria[0].PredictedPoints)
	}
	if scored.Criteria[0].Status != "partially_met" {
		t.Fatalf("status = %q, want partially_met", scored.Criteria[0].Status)
	}
}

func TestArrowPercentForScore_topBandLeft(t *testing.T) {
	row := RubricRow{Ratings: []RubricRating{
		{Title: "Poor", Points: "2 pts"},
		{Title: "Fair", Points: "4 pts"},
		{Title: "Good", Points: "6 pts"},
		{Title: "Excellent", Points: "8 pts"},
	}}
	if got := ArrowPercentForScore(row, 0.75); got != 12.5 {
		t.Fatalf("arrow = %v, want 12.5", got)
	}
}
