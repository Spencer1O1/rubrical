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
