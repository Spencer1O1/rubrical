package anthropic_test

import (
	"strings"
	"testing"

	"rubrical/internal/llm/anthropic"
	"rubrical/internal/llm/request"
)

func TestCompleteJSON_rejectsProviderFileDelivery(t *testing.T) {
	client := anthropic.New("test-key", "claude-test", "http://127.0.0.1:1", nil)
	_, err := client.CompleteJSON(t.Context(), request.Request{
		UserPrompt: "hi",
		Schema:     map[string]any{"type": "object"},
		Attachments: []request.Attachment{{
			Filename: "deck.pptx",
			MimeType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
			Data:     []byte("pk"),
			Delivery: request.DeliveryProviderFile,
		}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "provider_file") {
		t.Fatalf("error = %v", err)
	}
}
