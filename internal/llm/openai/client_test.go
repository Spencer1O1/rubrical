package openai

import (
	"testing"

	"rubrical/internal/llm/request"
)

func TestBuildResponsesInput_providerFile(t *testing.T) {
	input := buildResponsesInput(request.Request{
		UserPrompt: "grade this",
		Attachments: []request.Attachment{{
			Filename: "project.zip/src/app.js",
			MimeType: "application/javascript",
			Data:     []byte("export {}"),
			Delivery: request.DeliveryProviderFile,
		}},
	})
	if len(input) != 1 {
		t.Fatalf("messages = %d", len(input))
	}
	parts, ok := input[0].Content.([]inputContentPart)
	if !ok {
		t.Fatalf("content type %T", input[0].Content)
	}
	if len(parts) != 2 {
		t.Fatalf("parts = %d", len(parts))
	}
	if parts[1].Type != "input_file" || parts[1].Filename != "project.zip/src/app.js" {
		t.Fatalf("part = %+v", parts[1])
	}
}

func TestModelSupportsTemperature(t *testing.T) {
	if !modelSupportsTemperature("gpt-4o-mini") {
		t.Fatal("gpt-4o-mini should support temperature")
	}
	if !modelSupportsTemperature("gpt-4.1") {
		t.Fatal("gpt-4.1 should support temperature")
	}
	for _, model := range []string{"gpt-5.6-luna", "gpt-5.6-terra", "gpt-5.6-sol", "o3-mini"} {
		if modelSupportsTemperature(model) {
			t.Fatalf("%s should not support temperature", model)
		}
	}
}
