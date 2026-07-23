package files

import (
	"fmt"
)

var ErrNoAnalyzableContent = fmt.Errorf("no analyzable submission content")

func Process(provider string, submissionFiles []SubmissionInput, limits Limits) (ProcessResult, error) {
	limits = limits.withDefaults()
	var result ProcessResult
	var analyzedBytes int

	leaves, err := flattenSubmission(submissionFiles, limits)
	if err != nil {
		return ProcessResult{}, err
	}
	result.SkippedNotes = append(result.SkippedNotes, leaves.notes...)

	for _, raw := range leaves.files {
		fileBytes := len(raw.Data)

		inline, attachment, skip := routeFile(provider, raw)
		if skip != "" {
			result.SkippedNotes = append(result.SkippedNotes, skip)
			continue
		}
		if analyzedBytes+fileBytes > limits.MaxTotalBytes {
			result.SkippedNotes = append(result.SkippedNotes,
				fmt.Sprintf("%s: skipped (analysis byte budget)", raw.Path.String()))
			continue
		}
		if inline != nil {
			result.InlineSections = append(result.InlineSections, *inline)
		}
		if attachment != nil {
			result.Attachments = append(result.Attachments, *attachment)
		}
		analyzedBytes += fileBytes
	}

	paths := collectPaths(result.InlineSections, result.Attachments)
	result.Manifests = BuildManifests(paths)

	return result, nil
}

type SubmissionInput struct {
	FileName string
	MimeType string
	Data     []byte
}

type flattenOutput struct {
	files []RawFile
	notes []string
}

func flattenSubmission(submissionFiles []SubmissionInput, limits Limits) (flattenOutput, error) {
	var out flattenOutput

	for _, file := range submissionFiles {
		if len(file.Data) == 0 {
			out.notes = append(out.notes, fmt.Sprintf("%s: empty file", file.FileName))
			continue
		}
		raw := rawFromSubmission(file.FileName, file.MimeType, file.Data)
		kind := Classify(raw.FileName, raw.MimeType, raw.Data)

		if kind == KindZip {
			expanded, notes, err := expandZip(raw, limits)
			if err != nil {
				return flattenOutput{}, err
			}
			out.files = append(out.files, expanded...)
			out.notes = append(out.notes, notes...)
			continue
		}

		out.files = append(out.files, raw)
	}

	out.files, out.notes = dedupeByLogicalPath(out.files, out.notes)
	return out, nil
}

func dedupeByLogicalPath(files []RawFile, notes []string) ([]RawFile, []string) {
	seen := make(map[string]struct{}, len(files))
	deduped := make([]RawFile, 0, len(files))
	for _, file := range files {
		key := file.Path.String()
		if _, ok := seen[key]; ok {
			notes = append(notes, fmt.Sprintf("%s: skipped (duplicate path)", key))
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, file)
	}
	return deduped, notes
}
