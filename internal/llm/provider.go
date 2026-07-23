package llm

import "context"

// Provider is the vendor-agnostic AI surface used by product jobs.
type Provider interface {
	Name() string
	Model() string
	// CompleteJSON returns raw JSON bytes that satisfy req.Schema (when set).
	CompleteJSON(ctx context.Context, req Request) ([]byte, error)
}
