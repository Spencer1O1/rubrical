package analysis

import (
	"encoding/json"
	"testing"

	"rubrical/internal/analysis/files"
	"rubrical/internal/analysis/prompt"
)

func TestEncodePromptLog(t *testing.T) {
	fileResult, err := files.Process("openai", []files.SubmissionInput{{
		FileName: "essay.pdf",
		MimeType: "application/pdf",
		Data:     []byte("%PDF-1.4\nabc\n%%EOF\n"),
	}}, files.Limits{})
	if err != nil {
		t.Fatal(err)
	}

	data, err := EncodePromptLog(BuildProviderRequest(Input{}, fileResult, prompt.DefaultMaxSubmissionTextChars))
	if err != nil {
		t.Fatal(err)
	}
	var decoded PromptLog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Files) != 1 || decoded.Files[0].Bytes != len("%PDF-1.4\nabc\n%%EOF\n") {
		t.Fatalf("unexpected log: %+v", decoded)
	}
	if decoded.Files[0].Path != "essay.pdf" {
		t.Fatalf("path = %q", decoded.Files[0].Path)
	}
}
