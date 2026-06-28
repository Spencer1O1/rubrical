package analysis

import "testing"

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
