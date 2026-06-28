package prompt

import "testing"

func TestBuildSystem_scoringGuidance(t *testing.T) {
	got := BuildSystem()
	if !containsAll(got,
		"scoreRationale",
		"fulfilledRequirements",
		"unfulfilledRequirements",
		"bandPosition",
		"guidance",
		"every rubric criterion",
		"passion for helping students reach their full potential",
	) {
		t.Fatalf("system prompt missing scoring guidance: %q", got)
	}
	if contains(got, "predictedScore") || contains(got, "Do NOT return") {
		t.Fatalf("system prompt should not mention server-only fields: %q", got)
	}
}

func TestBuild_includesRubricAndDraft(t *testing.T) {
	system, user := Build(Input{
		Title:        "Essay 1",
		Instructions: "Write about live performance.",
		DraftMode:    "text",
		DraftText:    "My draft text.",
		Rubric: Rubric{
			Rows: []RubricRow{{Criterion: "Analysis", Points: "5"}},
		},
	}, DefaultMaxSubmissionTextChars)
	if system == "" || user == "" {
		t.Fatal("expected prompts")
	}
	if !containsAll(user, "Essay 1", "Write about live performance", "My draft text", "Analysis") {
		t.Fatalf("user prompt missing context: %q", user)
	}
}

func TestBuildContext_discussion(t *testing.T) {
	got := BuildContext(Input{
		PageType: "discussion",
		Title:    "Week 3 reply",
	})
	if !containsAll(got, "Discussion title: Week 3 reply", "Page type: discussion") {
		t.Fatalf("got %q", got)
	}
}

func TestBuildRubric_empty(t *testing.T) {
	got := formatRubric(Rubric{})
	if got != "(no rubric extracted)\n" {
		t.Fatalf("got %q", got)
	}
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !contains(text, part) {
			return false
		}
	}
	return true
}

func contains(text, part string) bool {
	return len(part) == 0 || len(text) >= len(part) && indexSubstring(text, part) >= 0
}

func indexSubstring(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
