package prompt

import "testing"

func TestBuildSubmission_sharedTextBudget(t *testing.T) {
	got := BuildSubmission(Input{
		DraftMode: "file",
		DraftText: stringsRepeat("a", 80),
		Files: FileContext{
			InlineSections: []FileInlineSection{
				{Path: "b.txt", Text: stringsRepeat("b", 80)},
			},
		},
	}, 100)

	if !contains(got, stringsRepeat("a", 80)) {
		t.Fatal("expected draft text first")
	}
	if contains(got, stringsRepeat("b", 80)) {
		t.Fatal("expected inline file truncated by shared budget")
	}
	if !contains(got, "### b.txt") {
		t.Fatal("expected inline heading when partial content fits")
	}
}

func TestBuildSubmission_textModeExcludesFiles(t *testing.T) {
	got := BuildSubmission(Input{
		DraftMode: "text",
		DraftText: "hello",
		Files: FileContext{
			InlineSections: []FileInlineSection{
				{Path: "secret.txt", Text: "should not appear"},
			},
		},
	}, 1000)

	if !contains(got, "hello") {
		t.Fatal("expected draft text")
	}
	if contains(got, "secret.txt") {
		t.Fatal("text mode must not include file context")
	}
}

func stringsRepeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
