package provider

import (
	"fmt"
	"strings"

	"rubrical/internal/analysis/anthropic"
	"rubrical/internal/analysis/openai"
)

type UserCredentials struct {
	Provider string
	Model    string
	APIKey   string
}

func (c UserCredentials) Valid() bool {
	switch strings.ToLower(strings.TrimSpace(c.Provider)) {
	case "openai", "anthropic":
		return strings.TrimSpace(c.APIKey) != "" && strings.TrimSpace(c.Model) != ""
	default:
		return false
	}
}

func NewFromUser(creds UserCredentials) (Provider, error) {
	providerName := strings.ToLower(strings.TrimSpace(creds.Provider))
	apiKey := strings.TrimSpace(creds.APIKey)
	model := strings.TrimSpace(creds.Model)

	switch providerName {
	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("openai api key is required")
		}
		return openai.NewProvider(apiKey, model), nil
	case "anthropic":
		if apiKey == "" {
			return nil, fmt.Errorf("anthropic api key is required")
		}
		return anthropic.NewProvider(apiKey, model), nil
	default:
		return nil, fmt.Errorf("unsupported ai provider %q", creds.Provider)
	}
}
