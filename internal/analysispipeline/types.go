package analysispipeline

import (
	"rubrical/internal/analysispipeline/analysis"
	analysisschema "rubrical/internal/analysispipeline/analysis/schema"
	"time"
)

type Input struct {
	AssignmentID    int64
	UserID          int64
	PageType        string
	Title           string
	CourseName      string
	Instructions    string
	PointsPossible  *float64
	AllowedChannels []string
	DraftMode       string
	DraftText       string
	DraftURL        string
	Files           []SubmissionFile
	Rubric          analysis.RubricContext
}

type SubmissionFile struct {
	FileName string
	MimeType string
	Data     []byte
}

type Result struct {
	RunID             int64
	Provider          string
	Model             string
	OverallSummary    string
	PredictedScore    *float64
	PredictedScoreMax *float64
	Confidence        string
	Feedback          []FeedbackItem
	CompletedAt       time.Time
}

type FeedbackItem struct {
	ID                      int64
	Category                string
	Severity                string
	Title                   string
	Explanation             string
	ScoreRationale          string
	HowToEarnPoints         string
	FulfilledRequirements   []analysisschema.FulfilledRequirement
	UnfulfilledRequirements []analysisschema.UnfulfilledRequirement
	CriterionStatus         string
	CriterionScore          *float64
	SelectedRating          string
	PredictedPoints         *float64
	MaxPoints               *float64
	Status                  string
	SortOrder               int
}
