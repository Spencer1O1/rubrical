package pages

import "rubrical/internal/analysis/files"

type FilePreviewView struct {
	HasPreview    bool
	ProviderLabel string
	AttachedPaths []string
	InlinePaths   []string
	SkippedNotes  []string
}

func FilePreviewFromResult(providerName string, result files.ProcessResult) FilePreviewView {
	if len(result.Attachments) == 0 && len(result.InlineSections) == 0 && len(result.SkippedNotes) == 0 {
		return FilePreviewView{}
	}
	view := FilePreviewView{
		HasPreview:    true,
		ProviderLabel: providerDisplayName(providerName),
		SkippedNotes:  append([]string(nil), result.SkippedNotes...),
	}
	for _, file := range result.Attachments {
		view.AttachedPaths = append(view.AttachedPaths, file.Path.String())
	}
	for _, section := range result.InlineSections {
		view.InlinePaths = append(view.InlinePaths, section.Path.String())
	}
	return view
}

func providerDisplayName(name string) string {
	switch name {
	case "anthropic":
		return "Anthropic"
	default:
		return "OpenAI"
	}
}
