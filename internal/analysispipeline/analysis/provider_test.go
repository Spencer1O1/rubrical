package analysis

import (
	"testing"

	"rubrical/internal/analysispipeline/analysis/schema"
)

func TestRatingIDsForSchema_withBands(t *testing.T) {
	row := threeBandRubric().Rows[0]
	got := ratingIDsForSchema(row)
	want := []string{"r0", "r1", "r2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRatingIDsForSchema_noBands(t *testing.T) {
	got := ratingIDsForSchema(RubricRow{Criterion: "Points only", Points: "10"})
	if len(got) != 1 || got[0] != "" {
		t.Fatalf("got %v, want [\"\"]", got)
	}
}

func TestApplyRubricScoring_emptyTitlePointsLabel(t *testing.T) {
	rubric := RubricContext{Rows: []RubricRow{{
		Criterion: "Overview",
		Ratings: []RubricRating{
			{Title: "", Points: "0 pts"},
			{Title: "", Points: "5 pts"},
			{Title: "", Points: "8 pts"},
		},
	}}}
	rubric.AssignCriterionIDs()
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionID:      rubric.Rows[0].ID,
		SelectedRatingID: "r2",
		BandPosition:     90,
		ScoreRationale:   "Strong work overall.",
	}}}
	scored, err := ApplyRubricScoring(resp, rubric)
	if err != nil {
		t.Fatal(err)
	}
	if scored.Criteria[0].SelectedRating != "8 pts" {
		t.Fatalf("selected = %q", scored.Criteria[0].SelectedRating)
	}
	if *scored.Criteria[0].PredictedPoints != 8 {
		t.Fatalf("points = %v", scored.Criteria[0].PredictedPoints)
	}
}

func TestRubricCriterionSpecs(t *testing.T) {
	specs := rubricCriterionSpecs(threeBandRubric())
	if len(specs) != 1 {
		t.Fatalf("len = %d", len(specs))
	}
	if specs[0].Name != "General Overview" || specs[0].ID != "general-overview" {
		t.Fatalf("spec = %+v", specs[0])
	}
	if len(specs[0].RatingIDs) != 3 || specs[0].RatingIDs[2] != "r2" {
		t.Fatalf("rating ids = %v", specs[0].RatingIDs)
	}
}
