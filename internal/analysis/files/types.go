package files

// LogicalPath identifies a file relative to the submission root.
type LogicalPath struct {
	ArchiveRoot  string // empty for top-level uploads
	RelativePath string
}

func (p LogicalPath) String() string {
	if p.ArchiveRoot == "" {
		return p.RelativePath
	}
	return p.ArchiveRoot + "/" + p.RelativePath
}

type RawFile struct {
	Path     LogicalPath
	FileName string
	MimeType string
	Data     []byte
}

type StructureManifest struct {
	ArchiveRoot string
	Tree        string
}

type InlineSection struct {
	Path      LogicalPath
	Text      string
	Extracted bool
}

type DeliveryKind string

const (
	DeliveryPDF          DeliveryKind = "pdf"
	DeliveryImage        DeliveryKind = "image"
	DeliveryProviderFile DeliveryKind = "provider_file"
)

type Attachment struct {
	Path     LogicalPath
	MimeType string
	Data     []byte
	Delivery DeliveryKind
}

type ProcessResult struct {
	Manifests      []StructureManifest
	InlineSections []InlineSection
	Attachments    []Attachment
	SkippedNotes   []string
}

func (r ProcessResult) HasContent() bool {
	return len(r.InlineSections) > 0 || len(r.Attachments) > 0
}
