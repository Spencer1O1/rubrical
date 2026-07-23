package aisettings

import (
	"testing"

	"rubrical/internal/secrets"
)

func TestStore_encryptDecryptKey(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	cipher, err := secrets.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore(nil, cipher)

	encrypted, err := store.encryptKey("sk-test")
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "sk-test" {
		t.Fatal("expected encrypted value")
	}

	got, err := store.decryptKey(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if got != "sk-test" {
		t.Fatalf("got %q", got)
	}
}

func TestSettingsPublic_omitsRawKeys(t *testing.T) {
	public := Settings{
		Provider:        "openai",
		Model:           "gpt-5.6-luna",
		OpenAIAPIKey:    "sk-secret",
		AnthropicAPIKey: "",
	}.Public()

	if public.OpenAIAPIKeyConfigured != true {
		t.Fatal("expected openai configured")
	}
	if public.AnthropicAPIKeyConfigured {
		t.Fatal("expected anthropic not configured")
	}
}

func TestMerge_keepsExistingKeysWhenBlank(t *testing.T) {
	current := Settings{
		Provider:        "openai",
		Model:           "gpt-5.6-luna",
		OpenAIAPIKey:    "sk-existing",
		AnthropicAPIKey: "sk-ant-existing",
	}
	next := Merge(current, Settings{
		Provider: "anthropic",
		Model:    "claude-sonnet-5",
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
		Model:    "gpt-5.6-luna",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
