package analysis

import (
	"rubrical/internal/analysispipeline/criterionname"
)

// AssignCriterionIDs sets stable slug ids on each row (rubric order). Idempotent if already set
// with the same names; always recomputes from current Criterion names.
func (r *RubricContext) AssignCriterionIDs() []criterionname.Ref {
	if r == nil {
		return nil
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
