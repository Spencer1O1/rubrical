package analysis

import "sort"

// RatingBandUI is one rubric rating band for the analysis UI scale.
type RatingBandUI struct {
	Title       string
	Description string
	Points      float64
}

func RatingBandsForUI(row RubricRow) []RatingBandUI {
	bands := parseRatingBands(row.Ratings)
	sort.Slice(bands, func(i, j int) bool { return bands[i].points > bands[j].points })
	out := make([]RatingBandUI, len(bands))
	for i, b := range bands {
		out[i] = RatingBandUI{Title: b.rating.Title, Description: b.rating.Description, Points: b.points}
	}
	return out
}

// ArrowPercentForScore: score 1 → 0% (left/green), score 0 → 100% (right/red).
func ArrowPercentForScore(score float64) float64 {
	return (1 - clamp01(score)) * 100
}
