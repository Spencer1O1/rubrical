package request

// DeliveryKind controls how an attachment is sent to the vendor API.
type DeliveryKind string

const (
	DeliveryPDF          DeliveryKind = "pdf"
	DeliveryImage        DeliveryKind = "image"
	DeliveryProviderFile DeliveryKind = "provider_file"
)

// Attachment is binary evidence attached to a completion request.
type Attachment struct {
	Filename string
	MimeType string
	Data     []byte
	Delivery DeliveryKind
}

// Request is the vendor-agnostic completion input.
// Schema is a JSON Schema object used for structured output.
type Request struct {
	SystemPrompt string
	UserPrompt   string
	Attachments  []Attachment
	SchemaName   string
	Schema       map[string]any
}
