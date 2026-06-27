package pages

import (
	"fmt"
	"time"
)

func WordCountLabel(count int) string {
	return fmt.Sprintf("%d words", count)
}

func DraftSavedMessage(wordCount int) string {
	return fmt.Sprintf("Draft saved (%d words)", wordCount)
}

func ImportedAtLabel(t time.Time) string {
	return t.Format("Jan 2, 2006 3:04 PM")
}

func DraftStatusLabel(wordCount int) string {
	if wordCount > 0 {
		return WordCountLabel(wordCount)
	}
	return "No draft saved yet"
}

func AssignmentURL(id int64) string {
	return fmt.Sprintf("/assignments/%d", id)
}

func AnalyzeURL(id int64) string {
	return fmt.Sprintf("/assignments/%d/analyze", id)
}

func rubricRatingDescriptionClass(title string) string {
	if title != "" {
		return "mt-2 break-words leading-relaxed text-stone-600"
	}
	return "break-words leading-relaxed text-stone-600 whitespace-pre-wrap"
}
