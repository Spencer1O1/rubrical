package aisettings

// PublicSettings is returned to browsers and the extension — never includes raw API keys.
type PublicSettings struct {
	Provider                    string `json:"provider"`
	Model                       string `json:"model"`
	OpenAIAPIKeyConfigured      bool   `json:"openaiApiKeyConfigured"`
	AnthropicAPIKeyConfigured   bool   `json:"anthropicApiKeyConfigured"`
}

func (s Settings) Public() PublicSettings {
	s = Normalize(s)
	return PublicSettings{
		Provider:                  s.Provider,
		Model:                     s.Model,
		OpenAIAPIKeyConfigured:    s.OpenAIAPIKey != "",
		AnthropicAPIKeyConfigured: s.AnthropicAPIKey != "",
	}
}
