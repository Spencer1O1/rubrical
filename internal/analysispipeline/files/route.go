package files

import (
	"fmt"
	"path/filepath"
	"strings"

	"rubrical/internal/analysispipeline/files/extract"
)

func routeFile(provider string, raw RawFile) (inline *InlineSection, attachment *Attachment, skipNote string) {
	if len(raw.Data) == 0 {
		return nil, nil, fmt.Sprintf("%s: empty file", raw.Path.String())
	}

	if note := invalidPDFNote(raw); note != "" {
		return nil, nil, note
	}

	kind := Classify(raw.FileName, raw.MimeType, raw.Data)
	mime := normalizeMime(raw.MimeType, raw.FileName)

	switch kind {
	case KindExecutable, KindMedia, KindArchiveUnsupported, KindLegacyDoc:
		return nil, nil, fmt.Sprintf("%s: %s", raw.Path.String(), SkipReason(kind))
	}

	provider = NormalizeProvider(provider)

	switch provider {
	case "openai":
		if delivery, ok := openAINativeDelivery(kind); ok {
			att := &Attachment{
				Path:     raw.Path,
				MimeType: mimeForDelivery(kind, mime),
				Data:     raw.Data,
				Delivery: delivery,
			}
			return nil, att, ""
		}
	case "anthropic":
		if delivery, ok := anthropicNativeDelivery(kind); ok {
			att := &Attachment{
				Path:     raw.Path,
				MimeType: mimeForDelivery(kind, mime),
				Data:     raw.Data,
				Delivery: delivery,
			}
			return nil, att, ""
		}
		if reason, skip := anthropicSkipReason(kind); skip {
			return nil, nil, fmt.Sprintf("%s: %s", raw.Path.String(), reason)
		}
		if !anthropicNeedsInline(kind) {
			return nil, nil, fmt.Sprintf("%s: %s", raw.Path.String(), SkipReason(kind))
		}
	}

	text, extracted, err := extractInline(kind, raw.Data)
	if err != nil {
		if kind == KindUnknown {
			return nil, nil, fmt.Sprintf("%s: %s", raw.Path.String(), SkipReason(kind))
		}
		return nil, nil, fmt.Sprintf("%s: %s", raw.Path.String(), err.Error())
	}
	if strings.TrimSpace(text) == "" {
		return nil, nil, fmt.Sprintf("%s: empty extract", raw.Path.String())
	}

	return &InlineSection{
		Path:      raw.Path,
		Text:      text,
		Extracted: extracted,
	}, nil, ""
}

func extractInline(kind FileKind, data []byte) (text string, extracted bool, err error) {
	switch kind {
	case KindDocx:
		text, err = extract.Docx(data)
		return text, true, err
	default:
		text, err = extract.Text(data)
		return text, false, err
	}
}

func mimeForDelivery(kind FileKind, mime string) string {
	if kind == KindPDF {
		return "application/pdf"
	}
	if kind == KindImage && strings.HasPrefix(mime, "image/") {
		return mime
	}
	if mime != "" && mime != "application/octet-stream" {
		return mime
	}
	return "text/plain"
}

func invalidPDFNote(raw RawFile) string {
	ext := strings.ToLower(filepath.Ext(raw.FileName))
	mime := normalizeMime(raw.MimeType, raw.FileName)
	if ext != ".pdf" && mime != "application/pdf" {
		return ""
	}
	if IsValidPDF(raw.Data) {
		return ""
	}
	return fmt.Sprintf("%s: file is not a valid PDF — delete it and re-upload", raw.Path.String())
}
