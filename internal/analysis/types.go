package analysis

import "time"

type Input struct {
	AssignmentID   int64
	UserID         int64
	PageType       string
	Title          string
	CourseName     string
	Instructions   string
	PointsPossible *float64
	DraftMode      string
	DraftText      string
	DraftURL       string
	Files          []SubmissionFile
	Rubric         RubricContext
}

type SubmissionFile struct {
	FileName string
	MimeType string
	Data     []byte
}

type RubricContext struct {
	Header []string
	Rows   []RubricRow
}

type RubricRow struct {
	Criterion                string
	CriterionLongDescription string
	Points                   string
	Ratings                  []RubricRating
}

type RubricRating struct {
	Title       string
	Description string
	Points      string
}

type Result struct {
	RunID            int64
	Provider         string
	Model            string
	OverallSummary   string
	PredictedScore    *float64
	PredictedScoreMax *float64
	Confidence       string
	Feedback         []FeedbackItem
	CompletedAt      time.Time
}

type FeedbackItem struct {
	ID              int64
	Category        string
	Severity        string
	Title           string
	Explanation     string
	Evidence        string
	Suggestion      string
	CriterionStatus string
	CriterionScore  *float64
	SelectedRating  string
	PredictedPoints *float64
	MaxPoints       *float64
	Status          string
	SortOrder       int
}
