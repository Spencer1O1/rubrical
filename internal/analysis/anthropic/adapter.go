package anthropic

import (
	"context"
	"encoding/json"

	"rubrical/internal/analysis/request"
	"rubrical/internal/analysis/schema"
)

type Provider struct {
	client *Client
}

func NewProvider(apiKey, model string) *Provider {
	return &Provider{client: New(apiKey, model)}
}

func (p *Provider) Name() string {
	if p == nil || p.client == nil {
		return "anthropic"
	}
	return p.client.Name()
}

func (p *Provider) Model() string {
	if p == nil || p.client == nil {
		return defaultModel
	}
	return p.client.Model()
}

func (p *Provider) Analyze(ctx context.Context, req request.Request) (*schema.ModelOutput, error) {
	attachments := make([]Attachment, len(req.Attachments))
	for i, file := range req.Attachments {
		attachments[i] = Attachment{
			Path:     file.Path,
			MimeType: file.MimeType,
			Data:     file.Data,
			Delivery: string(file.Delivery),
		}
	}

	raw, err := p.client.CompleteJSON(ctx, Request{
		SystemPrompt: req.SystemPrompt,
		UserPrompt:   req.UserPrompt,
		Attachments:  attachments,
	})
	if err != nil {
		return nil, err
	}

	var out schema.ModelOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if err := schema.Validate(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
