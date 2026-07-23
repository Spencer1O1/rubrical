package anthropic

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"rubrical/internal/config"
	"rubrical/internal/llm/request"
)

const (
	providerName     = "anthropic"
	anthropicVersion = "2023-06-01"
)

// Client talks to the Anthropic Messages API.
type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func New(apiKey, model, baseURL string, httpClient *http.Client) *Client {
	model = strings.TrimSpace(model)
	if model == "" {
		model = config.DefaultAnthropicModel
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = config.DefaultAnthropicBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: config.DefaultProviderTimeout}
	}
	return &Client{
		apiKey:     strings.TrimSpace(apiKey),
		model:      model,
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (c *Client) Name() string { return providerName }

func (c *Client) Model() string {
	if c == nil || c.model == "" {
		return config.DefaultAnthropicModel
	}
	return c.model
}

func (c *Client) CompleteJSON(ctx context.Context, req request.Request) ([]byte, error) {
	if c == nil || c.apiKey == "" {
		return nil, fmt.Errorf("anthropic api key is not configured")
	}
	if len(req.Schema) == 0 {
		return nil, fmt.Errorf("json schema is required")
	}

	content, err := buildUserContent(req)
	if err != nil {
		return nil, err
	}

	payload := messagesRequest{
		Model:     c.Model(),
		MaxTokens: config.DefaultAnthropicMaxTokens,
		System:    req.SystemPrompt,
		Messages: []message{
			{Role: "user", Content: content},
		},
		OutputConfig: outputConfig{
			Format: outputFormat{
				Type:   "json_schema",
				Schema: req.Schema,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("anthropic api error (%d): %s", resp.StatusCode, truncateErrorBody(respBody))
	}

	contentText, err := extractMessagesJSON(respBody)
	if err != nil {
		return nil, err
	}
	if !json.Valid(contentText) {
		return nil, fmt.Errorf("anthropic returned non-json content")
	}
	return contentText, nil
}

func extractMessagesJSON(respBody []byte) ([]byte, error) {
	var completion messagesResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return nil, fmt.Errorf("decode anthropic response: %w", err)
	}

	var textParts []string
	for _, block := range completion.Content {
		switch block.Type {
		case "text":
			if strings.TrimSpace(block.Text) != "" {
				textParts = append(textParts, strings.TrimSpace(block.Text))
			}
		case "refusal":
			return nil, fmt.Errorf("anthropic refused the request")
		}
	}

	contentText := strings.TrimSpace(strings.Join(textParts, "\n"))
	if contentText == "" {
		return nil, fmt.Errorf("anthropic returned empty content")
	}
	return []byte(contentText), nil
}

func buildUserContent(req request.Request) (any, error) {
	parts := []contentBlock{{Type: "text", Text: req.UserPrompt}}
	for _, file := range req.Attachments {
		switch file.Delivery {
		case request.DeliveryImage:
			parts = append(parts, contentBlock{
				Type: "image",
				Source: &sourceBlock{
					Type:      "base64",
					MediaType: file.MimeType,
					Data:      base64.StdEncoding.EncodeToString(file.Data),
				},
			})
		case request.DeliveryPDF:
			parts = append(parts, contentBlock{
				Type: "document",
				Source: &sourceBlock{
					Type:      "base64",
					MediaType: "application/pdf",
					Data:      base64.StdEncoding.EncodeToString(file.Data),
				},
			})
		case request.DeliveryProviderFile:
			return nil, fmt.Errorf("anthropic does not support provider_file delivery for %s", file.Filename)
		default:
			return nil, fmt.Errorf("unsupported attachment delivery %q for %s", file.Delivery, file.Filename)
		}
	}
	if len(parts) == 1 {
		return req.UserPrompt, nil
	}
	return parts, nil
}

func truncateErrorBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > 500 {
		return text[:500] + "…"
	}
	return text
}

type messagesRequest struct {
	Model        string       `json:"model"`
	MaxTokens    int          `json:"max_tokens"`
	System       string       `json:"system,omitempty"`
	Messages     []message    `json:"messages"`
	OutputConfig outputConfig `json:"output_config"`
}

type outputConfig struct {
	Format outputFormat `json:"format"`
}

type outputFormat struct {
	Type   string         `json:"type"`
	Schema map[string]any `json:"schema"`
}

type message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type contentBlock struct {
	Type   string       `json:"type"`
	Text   string       `json:"text,omitempty"`
	Source *sourceBlock `json:"source,omitempty"`
}

type sourceBlock struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type messagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}
