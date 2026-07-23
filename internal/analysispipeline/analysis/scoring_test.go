package analysis

import (
	"math"
	"testing"

	"rubrical/internal/analysispipeline/analysis/schema"
)

func threeBandRubric() RubricContext {
	rubric := RubricContext{Rows: []RubricRow{{
		Criterion: "General Overview",
		Ratings: []RubricRating{
			{Title: "Needs Improvement", Points: "0 pts"},
			{Title: "Good", Points: "5 pts"},
			{Title: "Excellent", Points: "8 pts"},
		},
	}}}
	rubric.AssignCriterionIDs()
	return rubric
}

func TestContinuousScore(t *testing.T) {
	cases := []struct {
		bandIdx, bandCount, bandPosition int
		want                             float64
	}{
		{2, 3, 100, 1.0},
		{2, 3, 85, 0.95},
		{2, 3, 0, 2.0 / 3.0},
		{0, 3, 50, 0.5 / 3.0},
	}
	for _, tc := range cases {
		got := continuousScore(tc.bandIdx, tc.bandCount, tc.bandPosition)
		if math.Abs(got-tc.want) > 0.001 {
			t.Fatalf("band %d position %d -> %v, want %v", tc.bandIdx, tc.bandPosition, got, tc.want)
		}
	}
}

func TestApplyRubricScoring_excellentMostlyMet(t *testing.T) {
	rubric := threeBandRubric()
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionID:      rubric.Rows[0].ID,
		SelectedRatingID: "r2",
		BandPosition:     85,
		ScoreRationale:   "Strong work.",
	}}}
	scored, err := ApplyRubricScoring(resp, rubric)
	if err != nil {
		t.Fatal(err)
	}
	c := scored.Criteria[0]
	if c.SelectedRating != "Excellent" {
		t.Fatalf("rating = %q", c.SelectedRating)
	}
	if *c.PredictedPoints != 8 {
		t.Fatalf("points = %v", c.PredictedPoints)
	}
	if c.Status != "met" {
		t.Fatalf("status = %q, want met", c.Status)
	}
	if math.Abs(c.CriterionScore-0.95) > 0.001 {
		t.Fatalf("normalized score = %v, want ~0.95", c.CriterionScore)
	}
	if math.Abs(ArrowPercentForScore(c.CriterionScore)-5) > 0.01 {
		t.Fatalf("arrow = %v, want ~5", ArrowPercentForScore(c.CriterionScore))
	}
}

func TestApplyRubricScoring_goodBand(t *testing.T) {
	rubric := threeBandRubric()
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionID:      rubric.Rows[0].ID,
		SelectedRatingID: "r1",
		BandPosition:     60,
		ScoreRationale:   "Decent.",
	}}}
	scored, err := ApplyRubricScoring(resp, rubric)
	if err != nil {
		t.Fatal(err)
	}
	c := scored.Criteria[0]
	if *c.PredictedPoints != 5 {
		t.Fatalf("points = %v", c.PredictedPoints)
	}
	if c.Status != "partially_met" {
		t.Fatalf("status = %q", c.Status)
	}
	want := (1 + 0.6) / 3.0
	if math.Abs(c.CriterionScore-want) > 0.001 {
		t.Fatalf("score = %v, want %v", c.CriterionScore, want)
	}
}

func TestApplyRubricScoring_pointOnlyRow(t *testing.T) {
	rubric := RubricContext{Rows: []RubricRow{{Criterion: "Participation", Points: "10 pts"}}}
	rubric.AssignCriterionIDs()
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionID:    rubric.Rows[0].ID,
		BandPosition:   65,
		ScoreRationale: "Mostly there.",
	}}}
	scored, err := ApplyRubricScoring(resp, rubric)
	if err != nil {
		t.Fatal(err)
	}
	if *scored.Criteria[0].PredictedPoints != 6.5 {
		t.Fatalf("points = %v", scored.Criteria[0].PredictedPoints)
	}
}

func TestApplyRubricScoring_rejectsPartialCriteria(t *testing.T) {
	rubric := RubricContext{Rows: []RubricRow{
		{Criterion: "General Overview", Ratings: threeBandRubric().Rows[0].Ratings},
		{Criterion: "Event Details", Ratings: threeBandRubric().Rows[0].Ratings},
	}}
	rubric.AssignCriterionIDs()
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionID:      rubric.Rows[0].ID,
		SelectedRatingID: "r2",
		BandPosition:     85,
		ScoreRationale:   "ok",
	}}}
	_, err := ApplyRubricScoring(resp, rubric)
	if err == nil {
		t.Fatal("expected missing criterion error")
	}
}

func TestApplyRubricScoring_rejectsUnknownBand(t *testing.T) {
	rubric := threeBandRubric()
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionID:      rubric.Rows[0].ID,
		SelectedRatingID: "r9",
		BandPosition:     80,
		ScoreRationale:   "ok",
	}}}
	_, err := ApplyRubricScoring(resp, rubric)
	if err == nil {
		t.Fatal("expected unknown band error")
	}
}

func TestArrowPercentForScore_linear(t *testing.T) {
	cases := []struct {
		score float64
		want  float64
	}{
		{1, 0},
		{0.75, 25},
		{0, 100},
	}
	for _, tc := range cases {
		if got := ArrowPercentForScore(tc.score); got != tc.want {
			t.Fatalf("score %.2f -> arrow %v%%, want %v%%", tc.score, got, tc.want)
		}
	}
}
