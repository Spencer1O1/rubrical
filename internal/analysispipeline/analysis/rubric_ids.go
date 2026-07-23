package analysis

import (
	"strings"

	"rubrical/internal/analysispipeline/criterionname"
)

// AssignCriterionIDs sets stable slug ids on each row (rubric order).
// If every row already has an id, those ids are kept — required so filterRubric
// subsets retain full-rubric slugs (duplicate names must not re-index as r0-only).
func (r *RubricContext) AssignCriterionIDs() []criterionname.Ref {
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
	refs := criterionname.Index(names)
	for i := range r.Rows {
		r.Rows[i].ID = refs[i].ID
	}
	return refs
}

// CriterionRefs returns id/name pairs (IDs must already be assigned).
func (r RubricContext) CriterionRefs() []criterionname.Ref {
	out := make([]criterionname.Ref, len(r.Rows))
	for i, row := range r.Rows {
		out[i] = criterionname.Ref{ID: row.ID, Name: row.Criterion}
	}
	return out
}
