package llm

import (
	"context"
	"fmt"
	"strings"

	"rubrical/internal/llm/anthropic"
	"rubrical/internal/llm/openai"
	"rubrical/internal/llm/request"
)

// New builds a Provider for the given vendor name, model, and API key.
func New(provider, model, apiKey string, opts ...Option) (Provider, error) {
	name := strings.ToLower(strings.TrimSpace(provider))
	apiKey = strings.TrimSpace(apiKey)
	model = strings.TrimSpace(model)
	if apiKey == "" {
		return nil, fmt.Errorf("%s api key is required", name)
	}

	cfg := applyOptions(clientConfig{}, opts)

	var inner interface {
		Name() string
		Model() string
		CompleteJSON(context.Context, request.Request) ([]byte, error)
	}

	switch name {
	case ProviderOpenAI:
		baseURL := cfg.baseURL
		if baseURL == "" {
			baseURL = DefaultOpenAIBaseURL
		}
		inner = openai.New(apiKey, model, baseURL, cfg.httpClient)
	case ProviderAnthropic:
		baseURL := cfg.baseURL
		if baseURL == "" {
			baseURL = DefaultAnthropicBaseURL
		}
		inner = anthropic.New(apiKey, model, baseURL, cfg.httpClient)
	default:
		return nil, fmt.Errorf("unsupported ai provider %q", provider)
	}

	return validatingProvider{inner: inner}, nil
}

type validatingProvider struct {
	inner interface {
		Name() string
		Model() string
		CompleteJSON(context.Context, request.Request) ([]byte, error)
	}
}

func (p validatingProvider) Name() string  { return p.inner.Name() }
func (p validatingProvider) Model() string { return p.inner.Model() }

func (p validatingProvider) CompleteJSON(ctx context.Context, req Request) ([]byte, error) {
	raw, err := p.inner.CompleteJSON(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := ValidateJSON(raw, req.Schema); err != nil {
		return nil, err
	}
	return raw, nil
}
