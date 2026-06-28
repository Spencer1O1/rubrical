package importpayload

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"rubrical/internal/drafturl"
	"rubrical/internal/importurl"
)

var allowedPageTypes = map[string]struct{}{
	"assignment":  {},
	"discussion":  {},
	"unknown":     {},
}

var allowedDraftEditorRoles = map[string]struct{}{
	"":            {},
	"topic_reply": {},
	"thread_reply": {},
}

// ValidateAndNormalize checks production import payloads and trims string fields.
func ValidateAndNormalize(payload *Payload, limits Limits) error {
	limits = limits.WithDefaults()
	if payload == nil {
		return fmt.Errorf("import payload is required")
	}

	payload.SourceURL = strings.TrimSpace(payload.SourceURL)
	sourceURL, err := importurl.ValidateSourceURL(payload.SourceURL)
	if err != nil {
		return err
	}
	payload.SourceURL = sourceURL

	payload.PageType = strings.ToLower(strings.TrimSpace(payload.PageType))
	if payload.PageType == "" {
		payload.PageType = "unknown"
	}
	if _, ok := allowedPageTypes[payload.PageType]; !ok {
		return fmt.Errorf("pageType must be assignment, discussion, or unknown")
	}

	if err := limitRunes("title", payload.Title, limits.MaxTitleRunes); err != nil {
		return err
	}
	payload.Title = strings.TrimSpace(payload.Title)

	if err := limitBytes("visibleText", payload.VisibleText, limits.MaxVisibleTextBytes); err != nil {
		return err
	}
	if err := limitBytes("instructionsText", payload.InstructionsText, limits.MaxInstructionsBytes); err != nil {
		return err
	}
	if err := limitBytes("draftText", payload.DraftText, limits.MaxDraftTextBytes); err != nil {
		return err
	}
	payload.DraftURL = strings.TrimSpace(payload.DraftURL)
	if payload.DraftURL != "" {
		normalized, err := drafturl.ParseSubmissionURL(payload.DraftURL)
		if err != nil {
			return fmt.Errorf("draftUrl: %w", err)
		}
		payload.DraftURL = normalized
	}
	if err := limitRunes("draftUrl", payload.DraftURL, limits.MaxMetadataFieldRunes); err != nil {
		return err
	}

	payload.DraftKind = strings.ToLower(strings.TrimSpace(payload.DraftKind))
	if payload.DraftKind != "" && payload.DraftKind != "text" && payload.DraftKind != "file" && payload.DraftKind != "url" {
		return fmt.Errorf("draftKind must be text, file, or url")
	}

	payload.DraftEditorRole = strings.ToLower(strings.TrimSpace(payload.DraftEditorRole))
	if _, ok := allowedDraftEditorRoles[payload.DraftEditorRole]; !ok {
		return fmt.Errorf("draftEditorRole must be topic_reply or thread_reply")
	}

	if err := validateMetadata(&payload.Metadata, limits); err != nil {
		return err
	}

	if err := validateDraftFiles(payload.DraftFiles, limits); err != nil {
		return err
	}

	if err := validateDraftFileRefs(payload.DraftFileRefs, len(payload.DraftFiles), limits); err != nil {
		return err
	}

	if err := validateRubric(payload.Rubric, limits); err != nil {
		return err
	}

	if payload.CapturedAt.IsZero() {
		payload.CapturedAt = time.Now().UTC()
	}

	return nil
}

func validateMetadata(metadata *Metadata, limits Limits) error {
	if err := limitRunes("metadata.dueDateText", metadata.DueDateText, limits.MaxMetadataFieldRunes); err != nil {
		return err
	}
	if err := limitRunes("metadata.dueAt", metadata.DueAt, limits.MaxMetadataFieldRunes); err != nil {
		return err
	}
	if err := limitRunes("metadata.pointsPossibleText", metadata.PointsPossibleText, limits.MaxMetadataFieldRunes); err != nil {
		return err
	}
	if err := limitRunes("metadata.submissionTypeText", metadata.SubmissionTypeText, limits.MaxMetadataFieldRunes); err != nil {
		return err
	}
	if err := limitRunes("metadata.courseName", metadata.CourseName, limits.MaxCourseNameRunes); err != nil {
		return err
	}

	metadata.DueDateText = strings.TrimSpace(metadata.DueDateText)
	metadata.DueAt = strings.TrimSpace(metadata.DueAt)
	metadata.PointsPossibleText = strings.TrimSpace(metadata.PointsPossibleText)
	metadata.SubmissionTypeText = strings.TrimSpace(metadata.SubmissionTypeText)
	metadata.CourseName = strings.TrimSpace(metadata.CourseName)

	seen := map[string]struct{}{}
	var allowed []string
	for _, raw := range metadata.AllowedSubmissionTypes {
		key := strings.ToLower(strings.TrimSpace(raw))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		allowed = append(allowed, key)
	}
	metadata.AllowedSubmissionTypes = allowed

	return nil
}

func validateDraftFiles(files []DraftFile, limits Limits) error {
	if len(files) == 0 {
		return nil
	}
	if len(files) > limits.MaxUploadSlots {
		return fmt.Errorf("draftFiles exceeds maximum upload slots")
	}

	for i := range files {
		if err := validateDraftFileEntry(i, &files[i], limits); err != nil {
			return err
		}
	}
	return nil
}

func validateDraftFileEntry(index int, file *DraftFile, limits Limits) error {
	field := fmt.Sprintf("draftFiles[%d].fileName", index)
	if err := limitRunes(field, file.FileName, limits.MaxDraftFileNameRunes); err != nil {
		return err
	}
	file.FileName = strings.TrimSpace(file.FileName)
	if file.FileName == "" {
		return fmt.Errorf("%s is required", field)
	}

	file.MimeType = strings.TrimSpace(file.MimeType)
	file.ContentBase64 = strings.TrimSpace(file.ContentBase64)
	if file.ContentBase64 == "" {
		return fmt.Errorf("draftFiles[%d].contentBase64 is required", index)
	}

	decoded, err := base64.StdEncoding.DecodeString(file.ContentBase64)
	if err != nil {
		return fmt.Errorf("draftFiles[%d].contentBase64 is invalid", index)
	}
	if len(decoded) > limits.MaxUploadBytes {
		return fmt.Errorf("draftFiles[%d] exceeds maximum size", index)
	}

	file.CanvasFileID = strings.TrimSpace(file.CanvasFileID)

	return nil
}

func validateDraftFileRefs(refs []DraftFileRef, newFileCount int, limits Limits) error {
	if len(refs) == 0 {
		return nil
	}
	if newFileCount+len(refs) > limits.MaxUploadSlots {
		return fmt.Errorf("draft upload slot count exceeds maximum")
	}

	seen := map[int64]struct{}{}
	for i := range refs {
		ref := &refs[i]
		if ref.ServerFileID <= 0 {
			return fmt.Errorf("draftFileRefs[%d].serverFileId is required", i)
		}
		if _, ok := seen[ref.ServerFileID]; ok {
			return fmt.Errorf("draftFileRefs[%d].serverFileId is duplicated", i)
		}
		seen[ref.ServerFileID] = struct{}{}

		field := fmt.Sprintf("draftFileRefs[%d].fileName", i)
		if err := limitRunes(field, ref.FileName, limits.MaxDraftFileNameRunes); err != nil {
			return err
		}
		ref.FileName = strings.TrimSpace(ref.FileName)

		ref.CanvasFileID = strings.TrimSpace(ref.CanvasFileID)
	}
	return nil
}

func validateRubric(rubric *RubricTable, limits Limits) error {
	if rubric == nil {
		return nil
	}

	if len(rubric.Header) > limits.MaxRubricHeaderColumns {
		return fmt.Errorf("rubric header exceeds maximum columns")
	}
	for i, header := range rubric.Header {
		if err := limitRunes(fmt.Sprintf("rubric.header[%d]", i), header, limits.MaxRubricFieldRunes); err != nil {
			return err
		}
	}

	if len(rubric.Rows) > limits.MaxRubricRows {
		return fmt.Errorf("rubric exceeds maximum rows")
	}

	for i, row := range rubric.Rows {
		if err := limitRunes(fmt.Sprintf("rubric.rows[%d].criterion", i), row.Criterion, limits.MaxRubricFieldRunes); err != nil {
			return err
		}
		if err := limitRunes(fmt.Sprintf("rubric.rows[%d].criterionLongDescription", i), row.CriterionLongDescription, limits.MaxRubricFieldRunes); err != nil {
			return err
		}
		if err := limitRunes(fmt.Sprintf("rubric.rows[%d].points", i), row.Points, limits.MaxRubricFieldRunes); err != nil {
			return err
		}
		if len(row.Ratings) > limits.MaxRatingsPerRow {
			return fmt.Errorf("rubric row %d exceeds maximum ratings", i)
		}
		for j, rating := range row.Ratings {
			if err := limitRunes(fmt.Sprintf("rubric.rows[%d].ratings[%d].title", i, j), rating.Title, limits.MaxRubricFieldRunes); err != nil {
				return err
			}
			if err := limitRunes(fmt.Sprintf("rubric.rows[%d].ratings[%d].description", i, j), rating.Description, limits.MaxRubricFieldRunes); err != nil {
				return err
			}
			if err := limitRunes(fmt.Sprintf("rubric.rows[%d].ratings[%d].points", i, j), rating.Points, limits.MaxRubricFieldRunes); err != nil {
				return err
			}
		}
	}

	return nil
}

func limitRunes(field, value string, max int) error {
	if utf8.RuneCountInString(value) > max {
		return fmt.Errorf("%s exceeds maximum length", field)
	}
	return nil
}

func limitBytes(field, value string, max int) error {
	if len(value) > max {
		return fmt.Errorf("%s exceeds maximum size", field)
	}
	return nil
}
