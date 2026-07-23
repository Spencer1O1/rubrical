package pages

import (
	"fmt"
	"strings"
	"time"

	"rubrical/internal/draftmode"
)

func WordCountLabel(count int) string {
	return fmt.Sprintf("%d words", count)
}

func DraftSavedMessage(wordCount int, fileName string) string {
	if strings.TrimSpace(fileName) != "" {
		return fmt.Sprintf("Draft saved from %s (%d words)", fileName, wordCount)
	}
	return fmt.Sprintf("Draft saved (%d words)", wordCount)
}

func DraftLinkSavedMessage() string {
	return "Link saved"
}

func DraftFilesSavedMessage(fileNames []string) string {
	names := make([]string, 0, len(fileNames))
	for _, name := range fileNames {
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			names = append(names, trimmed)
		}
	}
	if len(names) == 0 {
		return "File saved"
	}
	if len(names) == 1 {
		return fmt.Sprintf("File saved: %s", names[0])
	}
	return fmt.Sprintf("Files saved: %s", strings.Join(names, ", "))
}

func DraftFilesSavedWithSkippedEmpty(baseMessage string, skippedCount int) string {
	if skippedCount <= 0 {
		return baseMessage
	}
	label := "file"
	if skippedCount != 1 {
		label = "files"
	}
	return fmt.Sprintf("%s Skipped %d empty %s.", baseMessage, skippedCount, label)
}

func ImportedAtLabel(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04 PM")
}

func DueAtLabel(dueAt *time.Time) string {
	if dueAt == nil || dueAt.IsZero() {
		return ""
	}
	return fmt.Sprintf("Due %s", dueAt.Format("Jan 2, 2006 3:04 PM"))
}

func PointsPossibleLabel(points *float64) string {
	if points == nil {
		return ""
	}
	if *points == float64(int(*points)) {
		return fmt.Sprintf("%d pts possible", int(*points))
	}
	return fmt.Sprintf("%.1f pts possible", *points)
}

func PointsPossibleChip(points *float64) string {
	if points == nil {
		return ""
	}
	if *points == float64(int(*points)) {
		return fmt.Sprintf("%d pts", int(*points))
	}
	return fmt.Sprintf("%.1f pts", *points)
}

func UploadedAtLabel(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04 PM")
}

func DiscussionDraftStatusLabel(wordCount int, fileNames []string) string {
	parts := []string{}
	if wordCount > 0 {
		parts = append(parts, WordCountLabel(wordCount))
	}
	switch len(fileNames) {
	case 1:
		parts = append(parts, fileNames[0])
	case 0:
	default:
		parts = append(parts, fmt.Sprintf("%d files", len(fileNames)))
	}
	if len(parts) == 0 {
		return "No draft saved yet"
	}
	return strings.Join(parts, " · ")
}

func DraftStatusLabel(wordCount int, fileNames []string, submissionURL, mode string) string {
	switch draftmode.Normalize(mode) {
	case draftmode.File:
		switch len(fileNames) {
		case 0:
			return "No files uploaded yet"
		case 1:
			return fileNames[0]
		default:
			return fmt.Sprintf("%d files", len(fileNames))
		}
	case draftmode.URL:
		if strings.TrimSpace(submissionURL) != "" {
			return submissionURL
		}
		return "No URL saved yet"
	default:
		if wordCount > 0 {
			return WordCountLabel(wordCount)
		}
		return "No draft saved yet"
	}
}

func AssignmentURL(id int64) string {
	return fmt.Sprintf("/assignments/%d", id)
}

func DraftTextSaveURL(id int64) string {
	return fmt.Sprintf("/assignments/%d/draft", id)
}

func DraftStatusID(id int64) string {
	return fmt.Sprintf("draft-status-%d", id)
}

func DraftPanelBodyID(id int64) string {
	return fmt.Sprintf("draft-panel-body-%d", id)
}

const DraftAutoSaveDelayMs = 750

func DraftURLSaveTrigger() string {
	return fmt.Sprintf("blur, input changed delay:%dms, keyup[keyEnter] delay:%dms", DraftAutoSaveDelayMs, DraftAutoSaveDelayMs)
}

func AnalyzeURL(id int64) string {
	return fmt.Sprintf("/assignments/%d/analyze", id)
}

func AssignmentEmbedURL(id int64) string {
	return fmt.Sprintf("/assignments/%d?embed=1", id)
}

func SettingsURL(embed bool, assignmentID int64) string {
	if embed {
		url := "/settings?embed=1"
		if assignmentID > 0 {
			url += fmt.Sprintf("&assignment_id=%d", assignmentID)
		}
		return url
	}
	return "/settings"
}

func SettingsSavedRedirectURL(embed bool, assignmentID int64) string {
	if embed {
		url := "/settings?embed=1&saved=1"
		if assignmentID > 0 {
			url += fmt.Sprintf("&assignment_id=%d", assignmentID)
		}
		return url
	}
	return "/settings?saved=1"
}

func SettingsFormAction(embed bool, assignmentID int64) string {
	if embed {
		url := "/settings/ai?embed=1"
		if assignmentID > 0 {
			url += fmt.Sprintf("&assignment_id=%d", assignmentID)
		}
		return url
	}
	return "/settings/ai"
}

func DraftUploadURL(id int64) string {
	return fmt.Sprintf("/assignments/%d/draft/upload", id)
}

func DraftDiscussionUploadURL(id int64) string {
	return fmt.Sprintf("/assignments/%d/draft/discussion-attachment", id)
}

func DraftRemoveFileURL(assignmentID, fileID int64) string {
	return fmt.Sprintf("/assignments/%d/draft/files/%d/remove", assignmentID, fileID)
}

func DraftFileInputID(id int64) string {
	return fmt.Sprintf("draft-file-%d", id)
}

func DraftDiscussionFileInputID(id int64) string {
	return fmt.Sprintf("draft-discussion-file-%d", id)
}

func DraftTextareaID(id int64) string {
	return fmt.Sprintf("draft-text-%d", id)
}

func DraftEditorID(id int64) string {
	return fmt.Sprintf("draft-editor-%d", id)
}

func DraftURLInputID(id int64) string {
	return fmt.Sprintf("draft-url-%d", id)
}

func rubricRatingDescriptionClass(title string) string {
	if title != "" {
		return "mt-2 break-words leading-relaxed text-stone-600"
	}
	return "break-words leading-relaxed text-stone-600 whitespace-pre-wrap"
}
