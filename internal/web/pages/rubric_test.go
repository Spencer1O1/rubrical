package pages

import "testing"

func TestRubricTableRatingColumnCount(t *testing.T) {
	t.Parallel()

	table := RubricTableView{
		Rows: []RubricRowView{
			{Ratings: make([]RubricRatingView, 2)},
			{Ratings: make([]RubricRatingView, 3)},
		},
	}

	if got := table.RatingColumnCount(); got != 3 {
		t.Fatalf("RatingColumnCount() = %d, want 3", got)
	}
}
