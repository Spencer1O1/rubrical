package files

import "strings"

// textDrafts is always inspectable (draft channel, not a file kind).
const textDraftsCapability = "text drafts"

// capabilityGroups is the prompt-facing map: label → FileKinds.
// Providers for each label are derived via CanInspectKind (same rules as routeFile).
// Enable a new provider for a kind by updating delivery.go — this table stays labels only.
var capabilityGroups = []struct {
	Name  string
	Kinds []FileKind
}{
	{textDraftsCapability, nil}, // special: always both providers
	{"images", []FileKind{KindImage}},
	{"PDFs", []FileKind{KindPDF}},
	{"docx", []FileKind{KindDocx}},
	{"text files (txt, csv, tsv, …)", []FileKind{KindText}},
	{"Excel (xlsx/xls)", []FileKind{KindSpreadsheet}},
	{"code and structured text (code, md, json, html, xml)", []FileKind{KindCode, KindMarkdown, KindJSON, KindHTML, KindXML}},
	{"Office (pptx, odp, odt, rtf, …)", []FileKind{KindOffice}},
	{"video and audio", []FileKind{KindMedia}},
	{"executables and unknown binaries", []FileKind{KindExecutable, KindUnknown}},
	{"legacy .doc", []FileKind{KindLegacyDoc}},
	{"non-zip archives (rar, tar, …)", []FileKind{KindArchiveUnsupported}},
}

// PromptCapabilities splits capability groups into can/cannot for one provider.
func PromptCapabilities(provider string) (can, cannot []string) {
	provider = NormalizeProvider(provider)
	for _, g := range capabilityGroups {
		if capabilitySupported(provider, g.Name, g.Kinds) {
			can = append(can, g.Name)
		} else {
			cannot = append(cannot, g.Name)
		}
	}
	return can, cannot
}

func capabilitySupported(provider, name string, kinds []FileKind) bool {
	if name == textDraftsCapability {
		return true
	}
	for _, kind := range kinds {
		if CanInspectKind(provider, kind) {
			return true
		}
	}
	return false
}

// FormatPromptCapabilities renders Can/Cannot lines for system prompts
// (under a "# Capabilities" heading — heading already implies inspect).
func FormatPromptCapabilities(provider string) string {
	can, cannot := PromptCapabilities(provider)
	return "Can: " + strings.Join(can, "; ") + ".\n" +
		"Cannot: " + strings.Join(cannot, "; ") + "."
}
