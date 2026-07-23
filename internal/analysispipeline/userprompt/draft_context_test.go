package userprompt_test

import (
	"testing"

	"rubrical/internal/analysispipeline/userprompt"
)

func TestDraftContextLabel(t *testing.T) {
	cases := []struct {
		pageType string
		want     string
	}{
		{"assignment", userprompt.DraftContextAssignment},
		{"", userprompt.DraftContextAssignment},
		{"  ", userprompt.DraftContextAssignment},
		{"discussion", userprompt.DraftContextDiscussionMainTopic},
		{" discussion ", userprompt.DraftContextDiscussionMainTopic},
	}
	for _, tc := range cases {
		got := userprompt.DraftContextLabel(tc.pageType)
		if got != tc.want {
			t.Fatalf("pageType %q: got %q, want %q", tc.pageType, got, tc.want)
		}
	}
}

func TestInstructions(t *testing.T) {
	empty := userprompt.Instructions("  ")
	if empty != "# Instructions\n(none)\n" {
		t.Fatalf("empty: %q", empty)
	}
	got := userprompt.Instructions("Write a reply.")
	if got != "# Instructions\nWrite a reply.\n" {
		t.Fatalf("got %q", got)
	}
}
