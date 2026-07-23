package files

import (
	"rubrical/internal/llm/request"
)

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

// DeliveryKind is owned by llm/request — files route into the same values.
type DeliveryKind = request.DeliveryKind

const (
	DeliveryPDF          = request.DeliveryPDF
	DeliveryImage        = request.DeliveryImage
	DeliveryProviderFile = request.DeliveryProviderFile
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
