package submissiontypes

import (
	"strings"

	"rubrical/internal/draftmode"
)

var canvasToDraftMode = map[string]string{
	"online_text_entry": draftmode.Text,
	"online_upload":     draftmode.File,
	"online_url":        draftmode.URL,
}

var draftCapableCanvasTypes = map[string]struct{}{
	"online_text_entry": {},
	"online_upload":     {},
	"online_url":        {},
}

func IsDraftCapableCanvasType(raw string) bool {
	_, ok := draftCapableCanvasTypes[strings.ToLower(strings.TrimSpace(raw))]
	return ok
}

func DraftModeForCanvasType(raw string) string {
	return canvasToDraftMode[strings.ToLower(strings.TrimSpace(raw))]
}

// AllowedDraftModes returns Rubrical draft tabs from Canvas allowed submission types
// stored on import. When none were captured, default to all draft-capable modes.
func AllowedDraftModes(canvasTypes []string) []string {
	modes := uniqueDraftModes(canvasTypes)
	if len(modes) > 0 {
		return modes
	}
	return []string{draftmode.Text, draftmode.File, draftmode.URL}
}

func uniqueDraftModes(canvasTypes []string) []string {
	seen := map[string]struct{}{}
	var modes []string
	for _, raw := range canvasTypes {
		mode := DraftModeForCanvasType(raw)
		if mode == "" {
			continue
		}
		if _, ok := seen[mode]; ok {
			continue
		}
		seen[mode] = struct{}{}
		modes = append(modes, mode)
	}
	return modes
}

func NormalizeDraftMode(mode string, allowed []string) string {
	mode = draftmode.Normalize(mode)
	for _, item := range allowed {
		if item == mode {
			return mode
		}
	}
	if len(allowed) > 0 {
		return allowed[0]
	}
	return draftmode.Text
}

func ModeAllowed(mode string, allowed []string) bool {
	mode = draftmode.Normalize(mode)
	for _, item := range allowed {
		if item == mode {
			return true
		}
	}
	return false
}

func AttachmentsAllowed(canvasTypes []string) bool {
	for _, raw := range canvasTypes {
		if strings.ToLower(strings.TrimSpace(raw)) == "online_upload" {
			return true
		}
	}
	return false
}
