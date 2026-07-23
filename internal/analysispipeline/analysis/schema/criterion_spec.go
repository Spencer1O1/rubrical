package schema

// CriterionSpec describes one rubric row for provider JSON Schema generation.
type CriterionSpec struct {
	ID        string   // criterionId slug
	Name      string   // display name (prompts / scoring resolve)
	RatingIDs []string // selectedRatingId enum; [""] when the row has no rating bands
}
