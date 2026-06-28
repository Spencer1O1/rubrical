package prompt

import (
	"fmt"
	"strings"
)

func BuildSubmission(input Input, maxDraftChars int) string {
	var b strings.Builder
	b.WriteString("\n\n## Student submission\n")

	switch strings.TrimSpace(input.DraftMode) {
	case "url":
		writeURLSubmission(&b, input, maxDraftChars)
	case "file":
		writeFileSubmission(&b, input, maxDraftChars)
	default:
		writeTextSubmission(&b, input, maxDraftChars)
	}

	return b.String()
}

func writeURLSubmission(b *strings.Builder, input Input, maxDraftChars int) {
	url := strings.TrimSpace(input.DraftURL)
	if url == "" {
		return
	}
	fmt.Fprintf(b, "Submission type: website URL\nURL: %s\n", url)
	if text := strings.TrimSpace(input.DraftText); text != "" {
		b.WriteString("\nFetched page text (may be incomplete):\n")
		b.WriteString(truncate(text, maxDraftChars/4))
		b.WriteByte('\n')
	}
}

func writeFileSubmission(b *strings.Builder, input Input, maxDraftChars int) {
	b.WriteString("Submission type: file upload\n")
	if text := strings.TrimSpace(input.DraftText); text != "" {
		b.WriteString("Extracted file text:\n")
		b.WriteString(truncate(text, maxDraftChars))
		b.WriteByte('\n')
	}
	for _, file := range input.Files {
		fmt.Fprintf(b, "- Attached file: %s (%s, %d bytes)\n", file.FileName, file.MimeType, len(file.Data))
	}
}

func writeTextSubmission(b *strings.Builder, input Input, maxDraftChars int) {
	b.WriteString("Submission type: text\n")
	draft := strings.TrimSpace(input.DraftText)
	if draft == "" {
		b.WriteString("(empty)\n")
	} else {
		b.WriteString(truncate(draft, maxDraftChars))
		b.WriteByte('\n')
	}
	for _, file := range input.Files {
		fmt.Fprintf(b, "- Additional attachment: %s (%s, %d bytes)\n", file.FileName, file.MimeType, len(file.Data))
	}
}
