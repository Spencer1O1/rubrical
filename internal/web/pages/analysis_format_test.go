package pages

import (
	"strings"
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

func TestFormatArrowStyle_clampsToBarEdges(t *testing.T) {
	if got := formatArrowStyle(0); !strings.Contains(got, "clamp(6px") {
		t.Fatalf("got %q", got)
	}
	if got := formatArrowStyle(100); !strings.Contains(got, "calc(100% - 6px)") {
		t.Fatalf("got %q", got)
	}
}

func TestFormatRatingLabel(t *testing.T) {
	if got := formatRatingLabel("Good", "5"); got != "Good (5)" {
		t.Fatalf("got %q", got)
	}
}
