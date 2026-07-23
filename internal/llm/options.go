package llm

import "net/http"

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures adapter HTTP clients (base URL, transport) for tests.
type Option func(*clientConfig)

func WithBaseURL(baseURL string) Option {
	return func(c *clientConfig) {
		c.baseURL = baseURL
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = client
	}
}

func applyOptions(defaults clientConfig, opts []Option) clientConfig {
	cfg := defaults
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: DefaultProviderTimeout}
	}
	return cfg
}
