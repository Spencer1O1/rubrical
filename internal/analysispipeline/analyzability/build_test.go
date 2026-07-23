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
	if strings.Contains(req.UserPrompt, "Can inspect:") || strings.Contains(req.UserPrompt, "Cannot inspect:") {
		t.Fatalf("capabilities belong in system prompt, not user:\n%s", req.UserPrompt)
	}
	if !strings.Contains(req.SystemPrompt, "Can inspect:") {
		t.Fatal("expected capabilities in system prompt")
	}
	if !strings.Contains(req.SystemPrompt, "Office (pptx") {
		t.Fatal("openai system prompt should include Office in can-inspect")
	}
	if !strings.Contains(req.SystemPrompt, "text / files / URL") {
		t.Fatalf("expected plural files in channel list:\n%s", req.SystemPrompt)
	}
	if !strings.Contains(req.UserPrompt, "Classmate Reply") || !strings.Contains(req.UserPrompt, "id=classmate-reply") {
		t.Fatalf("expected criterion id+name in prompt:\n%s", req.UserPrompt)
	}
	if !strings.Contains(req.UserPrompt, "Allowed channels: text") {
		t.Fatalf("expected channels in user prompt:\n%s", req.UserPrompt)
	}
	if len(req.Attachments) != 0 {
		t.Fatal("pass 1 must not attach files")
	}
}

func TestBuildRequest_injectsAnthropicCapabilities(t *testing.T) {
	req := analyzability.BuildRequest(analyzability.Input{
		AllowedChannels: []string{"file"},
		Criteria:        criterionname.Index([]string{"Upload"}),
	}, "anthropic")
	canIdx := strings.Index(req.SystemPrompt, "Can inspect:")
	cannotIdx := strings.Index(req.SystemPrompt, "Cannot inspect:")
	if canIdx < 0 || cannotIdx < 0 || cannotIdx <= canIdx {
		t.Fatalf("missing can/cannot blocks:\n%s", req.SystemPrompt)
	}
	canBlock := req.SystemPrompt[canIdx:cannotIdx]
	cannotBlock := req.SystemPrompt[cannotIdx:]
	if strings.Contains(canBlock, "Office (pptx") {
		t.Fatal("anthropic should not list Office under Can inspect")
	}
	if !strings.Contains(cannotBlock, "Office (pptx") {
		t.Fatal("anthropic should list Office under Cannot inspect")
	}
	if !strings.Contains(req.UserPrompt, "Allowed channels: files") {
		t.Fatalf("file channel should prompt as files:\n%s", req.UserPrompt)
	}
	if !strings.Contains(req.SystemPrompt, "text files (txt, csv, tsv") {
		t.Fatalf("expected text/csv capability in system prompt:\n%s", req.SystemPrompt)
	}
}

func TestSystemPrompt_isFieldSpec(t *testing.T) {
	sys := analyzability.SystemPrompt("openai")
	for _, want := range []string{"criterionId", "analyzable", "reason", "howToEarnPoints", "Can inspect:", "text / files / URL", "possible channels"} {
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
