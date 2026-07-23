package files

import (
	"strings"

	"rubrical/internal/llm/request"
)

func openAINativeDelivery(kind FileKind) (request.DeliveryKind, bool) {
	switch kind {
	case KindPDF:
		return request.DeliveryPDF, true
	case KindImage:
		return request.DeliveryImage, true
	case KindDocx, KindOffice, KindSpreadsheet,
		KindText, KindCode, KindMarkdown,
		KindJSON, KindHTML, KindXML:
		return request.DeliveryProviderFile, true
	default:
		return "", false
	}
}

func anthropicNativeDelivery(kind FileKind) (request.DeliveryKind, bool) {
	switch kind {
	case KindPDF:
		return request.DeliveryPDF, true
	case KindImage:
		return request.DeliveryImage, true
	default:
		return "", false
	}
}

func anthropicNeedsInline(kind FileKind) bool {
	switch kind {
	case KindPDF, KindImage:
		return false
	case KindExecutable, KindMedia, KindArchiveUnsupported,
		KindLegacyDoc, KindUnknown, KindSpreadsheet, KindOffice, KindZip:
		return false
	default:
		return true
	}
}

func anthropicSkipReason(kind FileKind) (string, bool) {
	switch kind {
	case KindOffice:
		return "office file not supported with Anthropic (use docx/pdf or switch to OpenAI)", true
	case KindSpreadsheet:
		return "Excel not supported with Anthropic (use csv or switch to OpenAI)", true
	default:
		return "", false
	}
}

// CanInspectKind reports whether this provider can feed kind into the model
// (native attachment or inline extract). Same predicates as routeFile.
func CanInspectKind(provider string, kind FileKind) bool {
	switch kind {
	case KindExecutable, KindMedia, KindArchiveUnsupported, KindLegacyDoc, KindZip:
		return false
	case KindUnknown:
		return false
	}

	switch NormalizeProvider(provider) {
	case "anthropic":
		if _, ok := anthropicNativeDelivery(kind); ok {
			return true
		}
		if _, skip := anthropicSkipReason(kind); skip {
			return false
		}
		return anthropicNeedsInline(kind)
	default:
		if _, ok := openAINativeDelivery(kind); ok {
			return true
		}
		// OpenAI falls through to extract for odd kinds; only UTF-8/docx extractables.
		return extractableKind(kind)
	}
}

func extractableKind(kind FileKind) bool {
	switch kind {
	case KindDocx, KindText, KindCode, KindMarkdown, KindJSON, KindHTML, KindXML:
		return true
	default:
		return false
	}
}

// NormalizeProvider maps settings strings to routing keys.
func NormalizeProvider(name string) string {
	if strings.EqualFold(strings.TrimSpace(name), "anthropic") {
		return "anthropic"
	}
	return "openai"
}
