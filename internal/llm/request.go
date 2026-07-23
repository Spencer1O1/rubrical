package llm

import "rubrical/internal/llm/request"

// Re-export request types so product code can depend on package llm only.
type (
	DeliveryKind = request.DeliveryKind
	Attachment   = request.Attachment
	Request      = request.Request
)

const (
	DeliveryPDF          = request.DeliveryPDF
	DeliveryImage        = request.DeliveryImage
	DeliveryProviderFile = request.DeliveryProviderFile
)
