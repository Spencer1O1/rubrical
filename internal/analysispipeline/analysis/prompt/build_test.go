package prompt

import "testing"

func TestBuildSystem_scoringGuidance(t *testing.T) {
	got := BuildSystem("assignment")
	if !containsAll(got,
		"scoreRationale",
		"fulfilledRequirements",
		"unfulfilledRequirements",
		"Fully met requirements only",
		"Never also listed under fulfilled",
		"split it into distinct requirements",
		"bandPosition",
		"guidance",
		"Exactly one object per **analyzable** rubric criterion",
		"criterionId",
		"Draft context",
		"Assignment submission",
	) {
		t.Fatalf("system prompt missing scoring guidance: %q", got)
	}
	if contains(got, "predictedScore") || contains(got, "criterionScore") {
		t.Fatalf("system prompt should not mention server-only fields: %q", got)
	}
	if contains(got, "{{") {
		t.Fatal("unreplaced template placeholders")
	}
	if contains(got, "- Assignment submission") {
		t.Fatal("draft context value should not be a bullet")
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
	if contains(user, "# Draft context") {
		t.Fatalf("draft context belongs in system prompt, not user: %q", user)
	}
	if !containsAll(system, "# Draft context", "Assignment submission") {
		t.Fatalf("expected draft context in system: %q", system)
	}
}

func TestBuildContext_discussion(t *testing.T) {
	got := BuildContext(Input{
		PageType: "discussion",
		Title:    "Week 3 reply",
	})
	if !contains(got, "Title: Week 3 reply") {
		t.Fatalf("got %q", got)
	}
	if contains(got, "Draft context") || contains(got, "Page type:") || contains(got, "Discussion title:") {
		t.Fatalf("draft context/page type not in user context: %q", got)
	}
}

func TestBuildContext_assignment(t *testing.T) {
	got := BuildContext(Input{
		PageType: "assignment",
		Title:    "Essay 1",
	})
	if !contains(got, "Title: Essay 1") {
		t.Fatalf("got %q", got)
	}
	if contains(got, "Draft context") || contains(got, "Page type:") || contains(got, "Assignment title:") {
		t.Fatalf("draft context/page type not in user context: %q", got)
	}
}

func TestBuildSystem_discussionDraftContext(t *testing.T) {
	got := BuildSystem("discussion")
	if !containsAll(got, "# Draft context", "Discussion main topic reply") {
		t.Fatalf("got %q", got)
	}
	if contains(got, "- Discussion main topic reply") {
		t.Fatal("draft context value should not be a bullet")
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
