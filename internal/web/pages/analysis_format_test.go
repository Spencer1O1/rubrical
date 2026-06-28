package pages

import (
	"testing"
)

func TestFormatPointsLabel(t *testing.T) {
	est, max := 4.0, 5.0
	if got := formatPointsLabel(&est, &max); got != "4.0 / 5.0" {
		t.Fatalf("got %q", got)
	}
	if got := formatPointsLabel(nil, &max); got != "— / 5.0" {
		t.Fatalf("got %q", got)
	}
}

func TestCriterionStatusLabel(t *testing.T) {
	if got := criterionStatusLabel("partially_met"); got != "Partial" {
		t.Fatalf("got %q", got)
	}
}
