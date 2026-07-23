package llm

import "rubrical/internal/config"

const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"

	DefaultOpenAIModel    = config.DefaultOpenAIModel
	DefaultAnthropicModel = config.DefaultAnthropicModel

	DefaultOpenAIBaseURL    = config.DefaultOpenAIBaseURL
	DefaultAnthropicBaseURL = config.DefaultAnthropicBaseURL

	DefaultProviderTimeout    = config.DefaultProviderTimeout
	DefaultOpenAITemperature  = config.DefaultOpenAITemperature
	DefaultAnthropicMaxTokens = config.DefaultAnthropicMaxTokens
)
