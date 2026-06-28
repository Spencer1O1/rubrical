package importpayload

import "time"

type Payload struct {
	SourceURL        string         `json:"sourceUrl"`
	PageType         string         `json:"pageType"`
	Title            string         `json:"title"`
	VisibleText      string         `json:"visibleText"`
	InstructionsText string         `json:"instructionsText"`
	DraftText        string         `json:"draftText"`
	DraftURL         string         `json:"draftUrl"`
	DraftKind        string         `json:"draftKind"`
	DraftFiles       []DraftFile    `json:"draftFiles"`
	DraftFileRefs    []DraftFileRef `json:"draftFileRefs"`
	Rubric           *RubricTable   `json:"rubric"`
	Metadata         Metadata       `json:"metadata"`
	CapturedAt       time.Time      `json:"capturedAt"`
	DraftEditorRole  string         `json:"draftEditorRole"`
}

type DraftFile struct {
	FileName      string `json:"fileName"`
	MimeType      string `json:"mimeType"`
	ContentBase64 string `json:"contentBase64"`
	CanvasFileID  string `json:"canvasFileId,omitempty"`
	SortOrder     int    `json:"sortOrder,omitempty"`
}

type DraftFileRef struct {
	ServerFileID int64  `json:"serverFileId"`
	FileName     string `json:"fileName"`
	CanvasFileID string `json:"canvasFileId,omitempty"`
	SortOrder    int    `json:"sortOrder,omitempty"`
}

type RubricTable struct {
	Header []string     `json:"header"`
	Rows   []RubricRow  `json:"rows"`
}

type RubricRow struct {
	Criterion                string         `json:"criterion"`
	CriterionLongDescription string         `json:"criterionLongDescription,omitempty"`
	Ratings                  []RubricRating `json:"ratings"`
	Points                   string         `json:"points"`
}

type RubricRating struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Points      string `json:"points"`
}

type Metadata struct {
	DueDateText        string `json:"dueDateText"`
	DueAt              string `json:"dueAt"`
	PointsPossibleText string `json:"pointsPossibleText"`
	SubmissionTypeText       string   `json:"submissionTypeText"`
	AllowedSubmissionTypes   []string `json:"allowedSubmissionTypes"`
	CourseName               string   `json:"courseName"`
}
