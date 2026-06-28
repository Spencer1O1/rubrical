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

	"rubrical/internal/analysis/schema"
	"rubrical/internal/config"
)

const defaultModel = config.DefaultAnthropicModel
const defaultBaseURL = config.DefaultAnthropicBaseURL
const anthropicVersion = "2023-06-01"

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type Request struct {
	SystemPrompt string
	UserPrompt   string
	Attachments  []Attachment
}

type Attachment struct {
	Path     string
	MimeType string
	Data     []byte
	Delivery string
}

func New(apiKey, model string) *Client {
	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModel
	}
	return &Client{
		apiKey:  strings.TrimSpace(apiKey),
		model:   model,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: config.DefaultProviderTimeout,
		},
	}
}

func (c *Client) Name() string { return "anthropic" }
func (c *Client) Model() string {
	if c == nil || c.model == "" {
		return defaultModel
	}
	return c.model
}

func (c *Client) CompleteJSON(ctx context.Context, req Request) ([]byte, error) {
	if c == nil || c.apiKey == "" {
		return nil, fmt.Errorf("anthropic api key is not configured")
	}

	content := buildUserContent(req)
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
				Schema: schema.JSONSchema(),
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

func buildUserContent(req Request) any {
	parts := []contentBlock{{Type: "text", Text: req.UserPrompt}}
	for _, file := range req.Attachments {
		switch file.Delivery {
		case "image":
			parts = append(parts, contentBlock{
				Type: "image",
				Source: &sourceBlock{
					Type:      "base64",
					MediaType: file.MimeType,
					Data:      base64.StdEncoding.EncodeToString(file.Data),
				},
			})
		case "pdf":
			parts = append(parts, contentBlock{
				Type: "document",
				Source: &sourceBlock{
					Type:      "base64",
					MediaType: "application/pdf",
					Data:      base64.StdEncoding.EncodeToString(file.Data),
				},
			})
		}
	}
	if len(parts) == 1 {
		return req.UserPrompt
	}
	return parts
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
