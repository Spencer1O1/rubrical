package provider

import (
	"context"

	"rubrical/internal/analysis/request"
	"rubrical/internal/analysis/schema"
)

type Provider interface {
	Name() string
	Model() string
	Analyze(ctx context.Context, req request.Request) (*schema.ModelOutput, error)
}
