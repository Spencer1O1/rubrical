package aisettings

import "testing"

func TestMerge_keepsExistingKeysWhenBlank(t *testing.T) {
	current := Settings{
		Provider:        "openai",
		Model:           "gpt-4o-mini",
		OpenAIAPIKey:    "sk-existing",
		AnthropicAPIKey: "sk-ant-existing",
	}
	next := Merge(current, Settings{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-20250514",
	})
	if next.OpenAIAPIKey != "sk-existing" {
		t.Fatalf("openai key = %q", next.OpenAIAPIKey)
	}
	if next.AnthropicAPIKey != "sk-ant-existing" {
		t.Fatalf("anthropic key = %q", next.AnthropicAPIKey)
	}
}

func TestValidateSave_requiresActiveProviderKey(t *testing.T) {
	err := validateSave(Settings{
		Provider: "openai",
		Model:    "gpt-4o-mini",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
