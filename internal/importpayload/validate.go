package importpayload

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"rubrical/internal/importurl"
)

const (
	MaxImportBodyBytes      = 8 << 20
	MaxTitleRunes           = 500
	MaxCourseNameRunes      = 300
	MaxMetadataFieldRunes   = 500
	MaxVisibleTextBytes     = 512 << 10
	MaxInstructionsBytes    = 512 << 10
	MaxDraftTextBytes       = 512 << 10
	MaxRubricRows           = 100
	MaxRubricHeaderColumns  = 20
	MaxRatingsPerRow        = 50
	MaxRubricFieldRunes     = 4000
	MaxDraftFileBytes       = 32 << 20
	MaxDraftFileNameRunes   = 255
	MaxDraftFiles           = 20
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
func ValidateAndNormalize(payload *Payload) error {
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

	if err := limitRunes("title", payload.Title, MaxTitleRunes); err != nil {
		return err
	}
	payload.Title = strings.TrimSpace(payload.Title)

	if err := limitBytes("visibleText", payload.VisibleText, MaxVisibleTextBytes); err != nil {
		return err
	}
	if err := limitBytes("instructionsText", payload.InstructionsText, MaxInstructionsBytes); err != nil {
		return err
	}
	if err := limitBytes("draftText", payload.DraftText, MaxDraftTextBytes); err != nil {
		return err
	}
	payload.DraftURL = strings.TrimSpace(payload.DraftURL)
	if err := limitRunes("draftUrl", payload.DraftURL, MaxMetadataFieldRunes); err != nil {
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

	if err := validateMetadata(&payload.Metadata); err != nil {
		return err
	}

	if err := validateDraftFiles(payload.DraftFiles); err != nil {
		return err
	}

	if err := validateDraftFileRefs(payload.DraftFileRefs, len(payload.DraftFiles)); err != nil {
		return err
	}

	if err := validateRubric(payload.Rubric); err != nil {
		return err
	}

	if payload.CapturedAt.IsZero() {
		payload.CapturedAt = time.Now().UTC()
	}

	return nil
}

func validateMetadata(metadata *Metadata) error {
	if err := limitRunes("metadata.dueDateText", metadata.DueDateText, MaxMetadataFieldRunes); err != nil {
		return err
	}
	if err := limitRunes("metadata.dueAt", metadata.DueAt, MaxMetadataFieldRunes); err != nil {
		return err
	}
	if err := limitRunes("metadata.pointsPossibleText", metadata.PointsPossibleText, MaxMetadataFieldRunes); err != nil {
		return err
	}
	if err := limitRunes("metadata.submissionTypeText", metadata.SubmissionTypeText, MaxMetadataFieldRunes); err != nil {
		return err
	}
	if err := limitRunes("metadata.courseName", metadata.CourseName, MaxCourseNameRunes); err != nil {
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

func validateDraftFiles(files []DraftFile) error {
	if len(files) == 0 {
		return nil
	}
	if len(files) > MaxDraftFiles {
		return fmt.Errorf("draftFiles exceeds maximum count")
	}

	for i := range files {
		if err := validateDraftFileEntry(i, &files[i]); err != nil {
			return err
		}
	}
	return nil
}

func validateDraftFileEntry(index int, file *DraftFile) error {
	field := fmt.Sprintf("draftFiles[%d].fileName", index)
	if err := limitRunes(field, file.FileName, MaxDraftFileNameRunes); err != nil {
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
	if len(decoded) > MaxDraftFileBytes {
		return fmt.Errorf("draftFiles[%d] exceeds maximum size", index)
	}

	file.CanvasFileID = strings.TrimSpace(file.CanvasFileID)

	return nil
}

func validateDraftFileRefs(refs []DraftFileRef, newFileCount int) error {
	if len(refs) == 0 {
		return nil
	}
	if newFileCount+len(refs) > MaxDraftFiles {
		return fmt.Errorf("draft file count exceeds maximum")
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
		if err := limitRunes(field, ref.FileName, MaxDraftFileNameRunes); err != nil {
			return err
		}
		ref.FileName = strings.TrimSpace(ref.FileName)

		ref.CanvasFileID = strings.TrimSpace(ref.CanvasFileID)
	}
	return nil
}

func validateRubric(rubric *RubricTable) error {
	if rubric == nil {
		return nil
	}

	if len(rubric.Header) > MaxRubricHeaderColumns {
		return fmt.Errorf("rubric header exceeds maximum columns")
	}
	for i, header := range rubric.Header {
		if err := limitRunes(fmt.Sprintf("rubric.header[%d]", i), header, MaxRubricFieldRunes); err != nil {
			return err
		}
	}

	if len(rubric.Rows) > MaxRubricRows {
		return fmt.Errorf("rubric exceeds maximum rows")
	}

	for i, row := range rubric.Rows {
		if err := limitRunes(fmt.Sprintf("rubric.rows[%d].criterion", i), row.Criterion, MaxRubricFieldRunes); err != nil {
			return err
		}
		if err := limitRunes(fmt.Sprintf("rubric.rows[%d].criterionLongDescription", i), row.CriterionLongDescription, MaxRubricFieldRunes); err != nil {
			return err
		}
		if err := limitRunes(fmt.Sprintf("rubric.rows[%d].points", i), row.Points, MaxRubricFieldRunes); err != nil {
			return err
		}
		if len(row.Ratings) > MaxRatingsPerRow {
			return fmt.Errorf("rubric row %d exceeds maximum ratings", i)
		}
		for j, rating := range row.Ratings {
			if err := limitRunes(fmt.Sprintf("rubric.rows[%d].ratings[%d].title", i, j), rating.Title, MaxRubricFieldRunes); err != nil {
				return err
			}
			if err := limitRunes(fmt.Sprintf("rubric.rows[%d].ratings[%d].description", i, j), rating.Description, MaxRubricFieldRunes); err != nil {
				return err
			}
			if err := limitRunes(fmt.Sprintf("rubric.rows[%d].ratings[%d].points", i, j), rating.Points, MaxRubricFieldRunes); err != nil {
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
