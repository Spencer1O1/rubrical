package pages

import "strings"

type RubricRatingView struct {
	Title       string
	Description string
	Points      string
}

type RubricRowView struct {
	Criterion string
	Ratings   []RubricRatingView
	Points    string
}

type RubricTableView struct {
	Header []string
	Rows   []RubricRowView
}

func (t RubricTableView) RatingColumnCount() int {
	max := 0
	for _, row := range t.Rows {
		if len(row.Ratings) > max {
			max = len(row.Ratings)
		}
	}
	return max
}

func DefaultRubricHeader() []string {
	return []string{"Criteria", "Ratings", "Points"}
}

func RubricHeaderLabel(header []string, index int, fallback string, strict bool) string {
	if index < len(header) && strings.TrimSpace(header[index]) != "" {
		return header[index]
	}
	if strict {
		return "—"
	}
	return fallback
}
