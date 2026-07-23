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
	if !contains(got, "## b.txt") {
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
	if !contains(got, "Draft word count (computed by Rubrical): 1") {
		t.Fatalf("expected computed word count:\n%s", got)
	}
	if contains(got, "secret.txt") {
		t.Fatal("text mode must not include file context")
	}
}

func TestBuildSubmission_textModeWordCountFromHTML(t *testing.T) {
	got := BuildSubmission(Input{
		DraftMode: "text",
		DraftText: `<p>one two three</p>`,
	}, 1000)
	label := "Draft word count (computed by Rubrical, not student text): 3"
	body := "<p>one two three</p>"
	li := indexSubstring(got, label)
	bi := indexSubstring(got, body)
	if li < 0 {
		t.Fatalf("expected HTML word count:\n%s", got)
	}
	if bi < 0 || li > bi {
		t.Fatalf("computed word count should precede draft body:\n%s", got)
	}
}

func stringsRepeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
