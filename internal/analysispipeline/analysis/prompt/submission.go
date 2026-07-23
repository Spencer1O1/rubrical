package prompt

import (
	"fmt"
	"strings"

	"rubrical/internal/drafttext"
)

func BuildSubmission(input Input, maxSubmissionTextChars int) string {
	budget := newTextBudget(maxSubmissionTextChars)
	manifestCap := newManifestBudget(0)
	var b strings.Builder
	b.WriteString("# Submission\n")

	switch strings.TrimSpace(input.DraftMode) {
	case "url":
		writeURLSubmission(&b, input, &budget)
	case "file":
		writeFileSubmission(&b, input, &budget)
		writeFileContext(&b, input.Files, &budget, &manifestCap)
	default:
		writeTextSubmission(&b, input, &budget)
	}

	return b.String()
}

func writeURLSubmission(b *strings.Builder, input Input, budget *textBudget) {
	url := strings.TrimSpace(input.DraftURL)
	if url == "" {
		return
	}
	b.WriteString("website URL\n")
	fmt.Fprintf(b, "URL: %s\n", url)
	if text := budget.take(input.DraftText); text != "" {
		b.WriteString("\nFetched text (may be incomplete):\n")
		b.WriteString(text)
		b.WriteByte('\n')
	}
}

func writeFileSubmission(b *strings.Builder, input Input, budget *textBudget) {
	b.WriteString("files\n")
	if text := budget.take(input.DraftText); text != "" {
		b.WriteString("Draft text:\n")
		b.WriteString(text)
		b.WriteByte('\n')
	}
}

func writeTextSubmission(b *strings.Builder, input Input, budget *textBudget) {
	b.WriteString("text\n")
	words := drafttext.WordCount(input.DraftText)
	if words > 0 {
		fmt.Fprintf(b, "Draft word count (computed by Rubrical): %d\n", words)
	}
	draft := budget.take(input.DraftText)
	if draft == "" && !hasFileContext(input.Files) {
		b.WriteString("(empty)\n")
		return
	}
	if draft != "" {
		b.WriteString(draft)
		b.WriteByte('\n')
	}
}

func writeFileContext(b *strings.Builder, files FileContext, budget *textBudget, manifestCap *manifestBudget) {
	for _, manifest := range files.Manifests {
		if tree := manifestCap.take(manifest.Tree); tree != "" {
			b.WriteByte('\n')
			b.WriteString(tree)
		}
	}

	if len(files.InlineSections) > 0 {
		b.WriteString("\n# Text files\n")
		for _, section := range files.InlineSections {
			text := budget.take(section.Text)
			if text == "" {
				continue
			}
			heading := section.Path
			if section.Extracted {
				heading += " (extracted text)"
			}
			fmt.Fprintf(b, "## %s\n", heading)
			b.WriteString(text)
			b.WriteByte('\n')
		}
	}

	if len(files.AttachedFiles) > 0 {
		b.WriteString("\n# Attached files\n")
		for _, file := range files.AttachedFiles {
			fmt.Fprintf(b, "- %s (%s, %d bytes)\n", file.Path, file.MimeType, file.Bytes)
		}
	}

	if len(files.SkippedNotes) > 0 {
		b.WriteString("\n# Skipped files\n")
		for _, note := range files.SkippedNotes {
			line := manifestCap.take("- " + note)
			if line == "" {
				break
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
}

func hasFileContext(files FileContext) bool {
	return len(files.Manifests) > 0 ||
		len(files.InlineSections) > 0 ||
		len(files.AttachedFiles) > 0 ||
		len(files.SkippedNotes) > 0
}
