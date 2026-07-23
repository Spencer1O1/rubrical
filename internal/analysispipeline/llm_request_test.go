package analysispipeline_test

import (
	"encoding/json"
	"testing"

	"rubrical/internal/analysispipeline"
	"rubrical/internal/analysispipeline/analysis"
	"rubrical/internal/analysispipeline/analysis/prompt"
	"rubrical/internal/analysispipeline/files"
	"rubrical/internal/llm"
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

	req := analysis.BuildAnalysisRequest(analysis.DraftInput{}, fileResult, prompt.DefaultMaxSubmissionTextChars, "openai", analysis.RubricContext{})
	data, err := analysispipeline.EncodePromptLog(req)
	if err != nil {
		t.Fatal(err)
	}
	var decoded analysispipeline.PromptLog
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

func TestEncodePipelinePromptLog_includesBothPasses(t *testing.T) {
	pass1 := llm.Request{SystemPrompt: "sys1", UserPrompt: "user1", Schema: map[string]any{"type": "object"}}
	pass2 := llm.Request{SystemPrompt: "sys2", UserPrompt: "user2", Schema: map[string]any{"type": "object"}}
	data, err := analysispipeline.EncodePipelinePromptLog(pass1, &pass2)
	if err != nil {
		t.Fatal(err)
	}
	var decoded analysispipeline.PipelinePromptLog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Checkability.System != "sys1" || decoded.Analysis == nil || decoded.Analysis.System != "sys2" {
		t.Fatalf("unexpected pipeline log: %+v", decoded)
	}
}
