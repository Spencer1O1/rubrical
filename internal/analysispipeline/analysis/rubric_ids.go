package analysis

import (
	"strings"

	"rubrical/internal/analysispipeline/criterion"
)

// AssignCriterionIDs sets stable slug ids on each row (rubric order).
// If every row already has an id, those ids are kept — required so filterRubric
// subsets retain full-rubric slugs (duplicate names must not re-index as r0-only).
func (r *RubricContext) AssignCriterionIDs() []criterion.Ref {
	if r == nil {
		return nil
	}
	if len(r.Rows) > 0 {
		allSet := true
		for _, row := range r.Rows {
			if strings.TrimSpace(row.ID) == "" {
				allSet = false
				break
			}
		}
		if allSet {
			return r.CriterionRefs()
		}
	}
	names := make([]string, len(r.Rows))
	for i, row := range r.Rows {
		names[i] = row.Criterion
	}
	refs := criterion.Index(names)
	for i := range r.Rows {
		r.Rows[i].ID = refs[i].ID
		refs[i].Description = r.Rows[i].CriterionLongDescription
	}
	return refs
}

// CriterionRefs returns id/name/description pairs (IDs must already be assigned).
func (r RubricContext) CriterionRefs() []criterion.Ref {
	out := make([]criterion.Ref, len(r.Rows))
	for i, row := range r.Rows {
		out[i] = criterion.Ref{
			ID:          row.ID,
			Name:        row.Criterion,
			Description: row.CriterionLongDescription,
		}
	}
	return out
}
