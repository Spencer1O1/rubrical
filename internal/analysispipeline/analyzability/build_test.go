package analyzability_test

import (
	"strings"
	"testing"

	"rubrical/internal/analysispipeline/analyzability"
	"rubrical/internal/analysispipeline/criterionname"
)

func TestBuildRequest_excludesDraftEvidence(t *testing.T) {
	refs := criterionname.Index([]string{"Topic Response", "Classmate Reply"})
	req := analyzability.BuildRequest(analyzability.Input{
		PageType:        "discussion",
		Instructions:    "Reply to the prompt and respond to a classmate.",
		AllowedChannels: []string{"text"},
		Criteria:        refs,
	}, "openai")

	if strings.Contains(req.UserPrompt, "Draft text") || strings.Contains(req.UserPrompt, "Submission") {
		t.Fatalf("pass 1 prompt should not include draft evidence:\n%s", req.UserPrompt)
	}
	if strings.Contains(req.UserPrompt, "Can:") || strings.Contains(req.UserPrompt, "Cannot:") {
		t.Fatalf("capabilities belong in system prompt, not user:\n%s", req.UserPrompt)
	}
	if strings.Contains(req.UserPrompt, "Allowed channels:") || strings.Contains(req.UserPrompt, "Draft context") {
		t.Fatalf("channels/draft context belong in system prompt, not user:\n%s", req.UserPrompt)
	}
	if !strings.Contains(req.SystemPrompt, "Can:") {
		t.Fatal("expected capabilities in system prompt")
	}
	if !strings.Contains(req.SystemPrompt, "Office (pptx") {
		t.Fatal("openai system prompt should include Office in can-inspect")
	}
	if !strings.Contains(req.SystemPrompt, "via allowed channels (text)") {
		t.Fatalf("expected assignment channels in system prompt:\n%s", req.SystemPrompt)
	}
	if !strings.Contains(req.SystemPrompt, "# Draft context") || !strings.Contains(req.SystemPrompt, "Discussion main topic reply") {
		t.Fatalf("expected discussion draft context in system:\n%s", req.SystemPrompt)
	}
	if strings.Contains(req.SystemPrompt, "- Discussion main topic reply") {
		t.Fatal("draft context value should not be a bullet")
	}
	if !strings.Contains(req.UserPrompt, "Classmate Reply") || !strings.Contains(req.UserPrompt, "id=classmate-reply") {
		t.Fatalf("expected criterion id+name in prompt:\n%s", req.UserPrompt)
	}
	if len(req.Attachments) != 0 {
		t.Fatal("pass 1 must not attach files")
	}
}

func TestBuildRequest_assignmentDraftContext(t *testing.T) {
	req := analyzability.BuildRequest(analyzability.Input{
		PageType:        "assignment",
		AllowedChannels: []string{"text"},
		Criteria:        criterionname.Index([]string{"Essay"}),
	}, "openai")
	if !strings.Contains(req.SystemPrompt, "# Draft context") || !strings.Contains(req.SystemPrompt, "Assignment submission") {
		t.Fatalf("expected assignment draft context in system:\n%s", req.SystemPrompt)
	}
	if strings.Contains(req.UserPrompt, "Draft context") || strings.Contains(req.UserPrompt, "Page type:") {
		t.Fatalf("draft context/page type not in user prompt:\n%s", req.UserPrompt)
	}
}

func TestBuildRequest_injectsAnthropicCapabilities(t *testing.T) {
	req := analyzability.BuildRequest(analyzability.Input{
		AllowedChannels: []string{"file"},
		Criteria:        criterionname.Index([]string{"Upload"}),
	}, "anthropic")
	canIdx := strings.Index(req.SystemPrompt, "Can:")
	cannotIdx := strings.Index(req.SystemPrompt, "Cannot:")
	if canIdx < 0 || cannotIdx < 0 || cannotIdx <= canIdx {
		t.Fatalf("missing can/cannot blocks:\n%s", req.SystemPrompt)
	}
	canBlock := req.SystemPrompt[canIdx:cannotIdx]
	cannotBlock := req.SystemPrompt[cannotIdx:]
	if strings.Contains(canBlock, "Office (pptx") {
		t.Fatal("anthropic should not list Office under Can")
	}
	if !strings.Contains(cannotBlock, "Office (pptx") {
		t.Fatal("anthropic should list Office under Cannot")
	}
	if !strings.Contains(req.SystemPrompt, "via allowed channels (files)") {
		t.Fatalf("file channel should be in system prompt as files:\n%s", req.SystemPrompt)
	}
	if strings.Contains(req.UserPrompt, "Allowed channels:") {
		t.Fatalf("channels belong in system prompt, not user:\n%s", req.UserPrompt)
	}
	if !strings.Contains(req.SystemPrompt, "text files (txt, csv, tsv") {
		t.Fatalf("expected text/csv capability in system prompt:\n%s", req.SystemPrompt)
	}
}

func TestSystemPrompt_isFieldSpec(t *testing.T) {
	sys := analyzability.SystemPrompt("openai", "assignment", []string{"text", "file"})
	for _, want := range []string{
		"criterionId", "analyzable", "reason", "howToEarnPoints", "Can:",
		"via allowed channels (text, files)",
		"# Draft context",
		"Assignment submission",
	} {
		if !strings.Contains(sys, want) {
			t.Fatalf("system prompt missing %q", want)
		}
	}
	if strings.Contains(sys, "{{") {
		t.Fatal("unreplaced template placeholders")
	}
	if strings.Contains(sys, "Outside locus (analyzable: false):") {
		t.Fatal("system prompt still has verbose locus essay")
	}
	if strings.Contains(sys, "Missing expected") || strings.Contains(sys, "score it poorly") {
		t.Fatal("system prompt must not describe pass-2 scoring of missing evidence")
	}
	if strings.Contains(sys, "False when the locus") {
		t.Fatal("system prompt must not negative-rephrase the channel rule")
	}
	if strings.Contains(sys, "possible channels") || strings.Contains(sys, "this request lists") {
		t.Fatal("assignment channels are injected; no deferred user-prompt channel list")
	}
}

func TestValidateResponse_requiresHowToEarnPoints(t *testing.T) {
	refs := criterionname.Index([]string{"Classmate Reply", "Word Count"})
	resp := &analyzability.Response{Criteria: []analyzability.Criterion{
		{CriterionID: refs[0].ID, Analyzable: false, Reason: "peer reply", HowToEarnPoints: ""},
		{CriterionID: refs[1].ID, Analyzable: true, Reason: "text", HowToEarnPoints: ""},
	}}
	if err := analyzability.ValidateResponse(resp, refs); err == nil {
		t.Fatal("expected error for missing howToEarnPoints")
	}
	resp.Criteria[0].HowToEarnPoints = "Complete the classmate reply in Canvas before submitting."
	if err := analyzability.ValidateResponse(resp, refs); err != nil {
		t.Fatal(err)
	}
	if resp.Criteria[0].CriterionName != "Classmate Reply" {
		t.Fatalf("name resolve = %q", resp.Criteria[0].CriterionName)
	}
}

func TestParseResponse_ordersByRubric(t *testing.T) {
	refs := criterionname.Index([]string{"A", "B"})
	raw := []byte(`{
		"criteria": [
			{"criterionId":"b","analyzable":true,"reason":"text","howToEarnPoints":""},
			{"criterionId":"a","analyzable":false,"reason":"live","howToEarnPoints":"Attend class."}
		]
	}`)
	resp, err := analyzability.ParseResponse(raw, refs)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Criteria[0].CriterionID != "a" || resp.Criteria[1].CriterionID != "b" {
		t.Fatalf("order = %+v", resp.Criteria)
	}
	if resp.Criteria[0].CriterionName != "A" {
		t.Fatalf("resolved name = %q", resp.Criteria[0].CriterionName)
	}
}
