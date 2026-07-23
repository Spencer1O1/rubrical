package criterionname_test

import (
	"testing"

	"rubrical/internal/analysispipeline/criterionname"
)

func TestIndex_uniqueSlugs(t *testing.T) {
	refs := criterionname.Index([]string{"Word Count", "Word Count", "Classmate Reply!"})
	if refs[0].ID != "word-count" || refs[1].ID != "word-count-2" {
		t.Fatalf("ids = %q %q", refs[0].ID, refs[1].ID)
	}
	if refs[2].ID != "classmate-reply" {
		t.Fatalf("id = %q", refs[2].ID)
	}
	if refs[0].Name != "Word Count" {
		t.Fatalf("name = %q", refs[0].Name)
	}
}

func TestLookup(t *testing.T) {
	refs := criterionname.Index([]string{"A", "B"})
	got, ok := criterionname.Lookup(refs, refs[1].ID)
	if !ok || got.Name != "B" {
		t.Fatalf("lookup = %+v ok=%v", got, ok)
	}
}

func TestRatingID(t *testing.T) {
	if criterionname.RatingID(0) != "r0" || criterionname.RatingID(2) != "r2" {
		t.Fatal("rating ids")
	}
}
