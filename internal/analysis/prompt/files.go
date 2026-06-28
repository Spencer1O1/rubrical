package prompt

type FileContext struct {
	Manifests      []FileManifest
	InlineSections []FileInlineSection
	AttachedFiles  []AttachedFileIndex
	SkippedNotes   []string
}

type FileManifest struct {
	Tree string
}

type FileInlineSection struct {
	Path      string
	Text      string
	Extracted bool
}

type AttachedFileIndex struct {
	Path     string
	MimeType string
	Bytes    int
}
