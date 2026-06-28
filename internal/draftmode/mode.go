package draftmode

import "strings"

const (
	Text = "text"
	File = "file"
	URL  = "url"
)

func Valid(mode string) bool {
	switch mode {
	case Text, File, URL:
		return true
	default:
		return false
	}
}

func Normalize(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if Valid(mode) {
		return mode
	}
	return Text
}

// Infer picks the active draft mode from captured Canvas content.
func Infer(kind, draftText, draftURL string, hasFile bool) string {
	if k := Normalize(kind); kind != "" && k != Text {
		return k
	}
	if strings.TrimSpace(draftText) != "" {
		return Text
	}
	if strings.TrimSpace(draftURL) != "" {
		return URL
	}
	if hasFile {
		return File
	}
	if kind != "" {
		return Normalize(kind)
	}
	return Text
}
