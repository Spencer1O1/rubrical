package request

type DeliveryKind string

const (
	DeliveryPDF          DeliveryKind = "pdf"
	DeliveryImage        DeliveryKind = "image"
	DeliveryProviderFile DeliveryKind = "provider_file"
)

type Attachment struct {
	Path     string // full logical path
	MimeType string
	Data     []byte
	Delivery DeliveryKind
}

type Request struct {
	SystemPrompt string
	UserPrompt   string
	Attachments  []Attachment
}
