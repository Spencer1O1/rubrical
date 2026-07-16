package analysis

import (
	"testing"

	"rubrical/internal/analysis/schema"
)

func TestRatingTitlesForSchema_withBands(t *testing.T) {
	row := threeBandRubric().Rows[0]
	got := ratingTitlesForSchema(row)
	want := []string{"Needs Improvement", "Good", "Excellent"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRatingTitlesForSchema_noBands(t *testing.T) {
	got := ratingTitlesForSchema(RubricRow{Criterion: "Points only", Points: "10"})
	if len(got) != 1 || got[0] != "" {
		t.Fatalf("got %v, want [\"\"]", got)
	}
}

func TestRatingTitlesForSchema_emptyTitlesUsePoints(t *testing.T) {
	row := RubricRow{
		Criterion: "Overview",
		Ratings: []RubricRating{
			{Title: "", Points: "0 pts", Description: "Weak"},
			{Title: "", Points: "5 pts", Description: "OK"},
			{Title: "", Points: "8 pts", Description: "Strong"},
		},
	}
	got := ratingTitlesForSchema(row)
	want := []string{"0 pts", "5 pts", "8 pts"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNormalizeRowRatingTitles_matchesSchemaEnum(t *testing.T) {
	row := RubricRow{
		Criterion: "Overview",
		Ratings: []RubricRating{
			{Title: "", Points: "8 pts"},
			{Title: "", Points: "0 pts"},
			{Title: "", Points: "5 pts"},
		},
	}
	normalized := NormalizeRowRatingTitles(row)
	enum := ratingTitlesForSchema(row)
	byTitle := map[string]bool{}
	for _, title := range enum {
		byTitle[title] = true
	}
	for _, rating := range normalized.Ratings {
		if !byTitle[rating.Title] {
			t.Fatalf("prompt title %q not in schema enum %v", rating.Title, enum)
		}
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
	resp := &schema.ProviderResponse{Criteria: []schema.CriterionAssessment{{
		CriterionName:  "Overview",
		SelectedRating: "8 pts",
		BandPosition:   90,
		ScoreRationale: "Strong work overall.",
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
	if specs[0].Name != "General Overview" {
		t.Fatalf("name = %q", specs[0].Name)
	}
	if len(specs[0].RatingTitles) != 3 {
		t.Fatalf("rating titles = %v", specs[0].RatingTitles)
	}
}
