package request

type Attachment struct {
	FileName string
	MimeType string
	Data     []byte
	Kind     string
}

type Request struct {
	SystemPrompt string
	UserPrompt   string
	Attachments  []Attachment
}
