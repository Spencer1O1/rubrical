package pages

import (
	"fmt"

	"rubrical/internal/draftmode"
	"rubrical/internal/submissiontypes"
)

func AllowedDraftModes(canvasTypes []string) []string {
	return submissiontypes.AllowedDraftModes(canvasTypes)
}

func DefaultDraftMode(allowed []string) string {
	if len(allowed) > 0 {
		return allowed[0]
	}
	return draftmode.Text
}

func ClampDraftMode(mode string, allowed []string) string {
	return submissiontypes.NormalizeDraftMode(mode, allowed)
}

func DraftModeAllowed(mode string, allowed []string) bool {
	return submissiontypes.ModeAllowed(mode, allowed)
}

func DraftModeLabel(mode string) string {
	switch draftmode.Normalize(mode) {
	case draftmode.File:
		return "File"
	case draftmode.URL:
		return "Web URL"
	default:
		return "Text"
	}
}

func DraftModeIsActive(viewMode, tabMode string) bool {
	return draftmode.Normalize(viewMode) == draftmode.Normalize(tabMode)
}

func DraftModeIcon(mode string) string {
	switch draftmode.Normalize(mode) {
	case draftmode.File:
		return "upload_file"
	case draftmode.URL:
		return "link"
	default:
		return "title"
	}
}

func DraftModeAriaPressed(viewMode, tabMode string) string {
	if DraftModeIsActive(viewMode, tabMode) {
		return "true"
	}
	return "false"
}

func DraftModeTileClass(viewMode, tabMode string) string {
	base := "inline-flex items-center justify-center rounded-xl border p-3 shrink-0 transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
	if DraftModeIsActive(viewMode, tabMode) {
		return base + " border-indigo-600 bg-indigo-600 text-white"
	}
	return base + " border-stone-200 bg-white text-stone-700 hover:border-stone-300 hover:bg-stone-50"
}

func DraftModeURL(id int64) string {
	return fmt.Sprintf("/assignments/%d/draft/mode", id)
}

func DraftSaveURL(id int64) string {
	return fmt.Sprintf("/assignments/%d/draft/url", id)
}
