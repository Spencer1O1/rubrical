package pages

import "strings"

// NormalizeRubricRatings drops Canvas companion buttons that only carry points
// (same rating index merged on import). Works for any rating column count.
func NormalizeRubricRatings(ratings []RubricRatingView) []RubricRatingView {
	normalized := make([]RubricRatingView, 0, len(ratings))
	for _, rating := range ratings {
		if strings.TrimSpace(rating.Title) == "" && strings.TrimSpace(rating.Description) == "" {
			continue
		}
		normalized = append(normalized, rating)
	}
	return normalized
}
