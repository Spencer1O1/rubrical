package userprompt

import "strings"

// Draft context labels — shared by analyzability (pass 1) and analysis (pass 2).
const (
	DraftContextAssignment          = "Assignment submission"
	DraftContextDiscussionMainTopic = "Discussion main topic reply"
	DraftContextDiscussionClassmate = "Discussion classmate thread reply" // reserved; not emitted yet
)

// DraftContextLabel returns the fixed vocabulary label for pageType.
// Discussion classmate thread reply is reserved until that submission shape exists.
func DraftContextLabel(pageType string) string {
	switch strings.TrimSpace(pageType) {
	case "discussion":
		return DraftContextDiscussionMainTopic
	default:
		return DraftContextAssignment
	}
}
