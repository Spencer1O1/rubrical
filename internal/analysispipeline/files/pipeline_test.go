package files

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestClassify_codeAndSkipTypes(t *testing.T) {
	if got := Classify("app.js", "application/javascript", []byte("console.log(1)")); got != KindCode {
		t.Fatalf("js kind = %v", got)
	}
	if got := Classify("app.exe", "", []byte{0x4d, 0x5a}); got != KindExecutable {
		t.Fatalf("exe kind = %v", got)
	}
	if got := Classify("clip.mp4", "video/mp4", []byte{0, 0, 0}); got != KindMedia {
		t.Fatalf("mp4 kind = %v", got)
	}
	if got := Classify("legacy.doc", "application/msword", []byte("data")); got != KindLegacyDoc {
		t.Fatalf("doc kind = %v", got)
	}
}

func TestProcess_openAIMixedDelivery(t *testing.T) {
	result, err := Process("openai", []SubmissionInput{
		{FileName: "notes.txt", MimeType: "text/plain", Data: []byte("hello")},
		{FileName: "spec.pdf", MimeType: "application/pdf", Data: []byte("%PDF-1.4\n%%EOF\n")},
	}, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.InlineSections) != 0 {
		t.Fatalf("expected openai text native, got inline: %+v", result.InlineSections)
	}
	if len(result.Attachments) != 2 {
		t.Fatalf("attachments = %d", len(result.Attachments))
	}
}

func TestProcess_anthropicInlineCode(t *testing.T) {
	result, err := Process("anthropic", []SubmissionInput{
		{FileName: "main.go", MimeType: "text/plain", Data: []byte("package main\n")},
	}, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.InlineSections) != 1 {
		t.Fatalf("inline sections = %d", len(result.InlineSections))
	}
	if result.InlineSections[0].Path.String() != "main.go" {
		t.Fatalf("path = %q", result.InlineSections[0].Path)
	}
}

func TestProcess_anthropicInlineCSV(t *testing.T) {
	result, err := Process("anthropic", []SubmissionInput{
		{FileName: "grades.csv", MimeType: "text/csv", Data: []byte("name,score\nAda,100\n")},
	}, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.InlineSections) != 1 {
		t.Fatalf("inline sections = %d, want 1 for CSV with Anthropic", len(result.InlineSections))
	}
	if result.InlineSections[0].Path.String() != "grades.csv" {
		t.Fatalf("path = %q", result.InlineSections[0].Path)
	}
}

func TestProcess_skipsExeContinuesJS(t *testing.T) {
	result, err := Process("openai", []SubmissionInput{
		{FileName: "virus.exe", Data: []byte{0x4d, 0x5a, 0x90}},
		{FileName: "app.js", Data: []byte("export {}")},
	}, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.SkippedNotes) != 1 {
		t.Fatalf("skipped = %#v", result.SkippedNotes)
	}
	if len(result.Attachments) != 1 || result.Attachments[0].Path.String() != "app.js" {
		t.Fatalf("attachments = %+v", result.Attachments)
	}
}

func TestProcess_zipPreservesPaths(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("src/main.js")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("console.log('hi')")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	result, err := Process("anthropic", []SubmissionInput{
		{FileName: "project.zip", MimeType: "application/zip", Data: buf.Bytes()},
	}, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.InlineSections) != 1 {
		t.Fatalf("inline = %+v", result.InlineSections)
	}
	want := "project.zip/src/main.js"
	if result.InlineSections[0].Path.String() != want {
		t.Fatalf("path = %q want %q", result.InlineSections[0].Path, want)
	}
	if len(result.Manifests) == 0 || !bytes.Contains([]byte(result.Manifests[0].Tree), []byte("src/")) {
		t.Fatalf("manifest = %#v", result.Manifests)
	}
}

func TestProcess_largeMediaSkippedAsTypeNotByteBudget(t *testing.T) {
	big := make([]byte, 70<<20) // 70 MiB — over default 64 MiB analysis budget
	copy(big, []byte{0, 0, 0})

	result, err := Process("openai", []SubmissionInput{
		{FileName: "Recording 2026-06-27 225444.mp4", MimeType: "video/mp4", Data: big},
	}, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Attachments) != 0 || len(result.InlineSections) != 0 {
		t.Fatalf("expected no analyzed content, got attachments=%d inline=%d",
			len(result.Attachments), len(result.InlineSections))
	}
	if len(result.SkippedNotes) != 1 {
		t.Fatalf("skipped = %#v", result.SkippedNotes)
	}
	if bytes.Contains([]byte(result.SkippedNotes[0]), []byte("byte budget")) {
		t.Fatalf("skip note = %q, want media/type skip not byte budget", result.SkippedNotes[0])
	}
	if !bytes.Contains([]byte(result.SkippedNotes[0]), []byte("unsupported media")) {
		t.Fatalf("skip note = %q", result.SkippedNotes[0])
	}
}

func TestProcess_analysisByteBudget(t *testing.T) {
	limits := Limits{MaxTotalBytes: 10}
	result, err := Process("anthropic", []SubmissionInput{
		{FileName: "a.txt", Data: []byte("12345")},
		{FileName: "b.txt", Data: []byte("12345")},
		{FileName: "c.txt", Data: []byte("12345")},
	}, limits)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.InlineSections) != 2 {
		t.Fatalf("inline sections = %d", len(result.InlineSections))
	}
	if len(result.SkippedNotes) != 1 {
		t.Fatalf("skipped = %#v", result.SkippedNotes)
	}
	if !bytes.Contains([]byte(result.SkippedNotes[0]), []byte("analysis byte budget")) {
		t.Fatalf("skip note = %q", result.SkippedNotes[0])
	}
}

func TestBuildManifests_flatTopLevel(t *testing.T) {
	manifests := BuildManifests([]LogicalPath{{RelativePath: "a.txt"}, {RelativePath: "b.pdf"}})
	if len(manifests) != 1 {
		t.Fatalf("manifests = %d", len(manifests))
	}
	if !bytes.Contains([]byte(manifests[0].Tree), []byte("a.txt")) {
		t.Fatalf("tree = %q", manifests[0].Tree)
	}
}
