package schema

// CriterionSpec describes one rubric row for provider JSON Schema generation.
type CriterionSpec struct {
	Name         string
	RatingTitles []string // selectedRating enum; [""] when the row has no rating bands
}
