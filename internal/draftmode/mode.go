package draftmode

import "strings"

const (
	Text = "text"
	File = "file"
	URL  = "url"
)

// All is the concrete channel list (internal wire values).
func All() []string {
	return []string{Text, File, URL}
}

// PromptLabel is the human/channel name used in LLM prompts.
// Internal File stays "file"; prompts say "files".
func PromptLabel(mode string) string {
	switch Normalize(mode) {
	case File:
		return "files"
	case URL:
		return "URL"
	default:
		return Text
	}
}

// PromptLabels maps wire channel values to prompt labels.
func PromptLabels(modes []string) []string {
	if len(modes) == 0 {
		modes = All()
	}
	out := make([]string, 0, len(modes))
	for _, m := range modes {
		out = append(out, PromptLabel(m))
	}
	return out
}

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
