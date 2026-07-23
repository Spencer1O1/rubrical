package analysis

import "testing"

func TestAssignCriterionIDs_preservesFilteredDuplicateSlugs(t *testing.T) {
	full := RubricContext{Rows: []RubricRow{
		{Criterion: "Content"},
		{Criterion: "Content"},
	}}
	full.AssignCriterionIDs()
	if full.Rows[0].ID != "content" || full.Rows[1].ID != "content-2" {
		t.Fatalf("full ids = %q %q", full.Rows[0].ID, full.Rows[1].ID)
	}

	// Pass 2 often sees only a subset (e.g. first row not analyzable).
	filtered := RubricContext{Rows: []RubricRow{full.Rows[1]}}
	filtered.AssignCriterionIDs()
	if filtered.Rows[0].ID != "content-2" {
		t.Fatalf("filtered re-indexed to %q, want content-2", filtered.Rows[0].ID)
	}
}

func TestAssignCriterionIDs_assignsWhenMissing(t *testing.T) {
	rubric := RubricContext{Rows: []RubricRow{{Criterion: "Word Count"}}}
	refs := rubric.AssignCriterionIDs()
	if refs[0].ID != "word-count" || rubric.Rows[0].ID != "word-count" {
		t.Fatalf("got %+v", refs)
	}
}
