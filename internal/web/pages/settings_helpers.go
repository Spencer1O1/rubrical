package pages

import (
	"strings"

	"rubrical/internal/aisettings"
)

func ProviderSavedModel(settings aisettings.Settings, providerID string) string {
	if settings.Provider == providerID && strings.TrimSpace(settings.Model) != "" {
		return settings.Model
	}
	switch providerID {
	case "anthropic":
		return AnthropicModelOptions[0].ID
	default:
		return OpenAIModelOptions[0].ID
	}
}
