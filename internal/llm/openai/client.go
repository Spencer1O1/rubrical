package openai

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

const providerName = "openai"

// Client talks to the OpenAI Responses API.
type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func New(apiKey, model, baseURL string, httpClient *http.Client) *Client {
	model = strings.TrimSpace(model)
	if model == "" {
		model = config.DefaultOpenAIModel
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = config.DefaultOpenAIBaseURL
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
		return config.DefaultOpenAIModel
	}
	return c.model
}

func (c *Client) CompleteJSON(ctx context.Context, req request.Request) ([]byte, error) {
	if c == nil || c.apiKey == "" {
		return nil, fmt.Errorf("openai api key is not configured")
	}
	if len(req.Schema) == 0 {
		return nil, fmt.Errorf("json schema is required")
	}
	schemaName := strings.TrimSpace(req.SchemaName)
	if schemaName == "" {
		schemaName = "response"
	}

	payload := responsesRequest{
		Model:        c.Model(),
		Instructions: req.SystemPrompt,
		Input:        buildResponsesInput(req),
		Store:        false,
		Temperature:  config.DefaultOpenAITemperature,
		Text: responsesTextConfig{
			Format: responsesJSONSchemaFormat{
				Type:   "json_schema",
				Name:   schemaName,
				Strict: true,
				Schema: req.Schema,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
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
		return nil, fmt.Errorf("openai api error (%d): %s", resp.StatusCode, truncateErrorBody(respBody))
	}

	content, err := extractResponsesJSON(respBody)
	if err != nil {
		return nil, err
	}
	if !json.Valid(content) {
		return nil, fmt.Errorf("openai returned non-json content")
	}
	return content, nil
}

func buildResponsesInput(req request.Request) []inputMessage {
	parts := []inputContentPart{{Type: "input_text", Text: req.UserPrompt}}
	for _, file := range req.Attachments {
		switch file.Delivery {
		case request.DeliveryPDF, request.DeliveryProviderFile:
			parts = append(parts, inputContentPart{
				Type:     "input_file",
				Filename: file.Filename,
				FileData: dataURL(file.MimeType, file.Data),
			})
		case request.DeliveryImage:
			parts = append(parts, inputContentPart{
				Type:     "input_image",
				ImageURL: dataURL(file.MimeType, file.Data),
			})
		}
	}

	content := any(req.UserPrompt)
	if len(parts) > 1 {
		content = parts
	}

	return []inputMessage{{
		Role:    "user",
		Content: content,
	}}
}

func extractResponsesJSON(respBody []byte) ([]byte, error) {
	var response responsesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("decode openai response: %w", err)
	}

	for _, item := range response.Output {
		if item.Type == "refusal" {
			return nil, fmt.Errorf("openai refused the request")
		}
	}

	text := strings.TrimSpace(response.OutputText)
	if text == "" {
		text = strings.TrimSpace(extractOutputText(response.Output))
	}
	if text == "" {
		return nil, fmt.Errorf("openai returned empty content")
	}
	return []byte(text), nil
}

func extractOutputText(output []outputItem) string {
	var parts []string
	for _, item := range output {
		if item.Type != "message" {
			continue
		}
		for _, block := range item.Content {
			if block.Type == "output_text" && strings.TrimSpace(block.Text) != "" {
				parts = append(parts, strings.TrimSpace(block.Text))
			}
		}
	}
	return strings.Join(parts, "\n")
}

func dataURL(mimeType string, data []byte) string {
	mime := strings.TrimSpace(mimeType)
	if mime == "" {
		mime = "application/octet-stream"
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data))
}

func truncateErrorBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > 500 {
		return text[:500] + "…"
	}
	return text
}

type responsesRequest struct {
	Model        string              `json:"model"`
	Instructions string              `json:"instructions,omitempty"`
	Input        any                 `json:"input"`
	Store        bool                `json:"store"`
	Temperature  float64             `json:"temperature,omitempty"`
	Text         responsesTextConfig `json:"text"`
}

type responsesTextConfig struct {
	Format responsesJSONSchemaFormat `json:"format"`
}

type responsesJSONSchemaFormat struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type inputMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type inputContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Filename string `json:"filename,omitempty"`
	FileData string `json:"file_data,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type responsesResponse struct {
	OutputText string       `json:"output_text"`
	Output     []outputItem `json:"output"`
}

type outputItem struct {
	Type    string        `json:"type"`
	Content []outputBlock `json:"content,omitempty"`
}

type outputBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
