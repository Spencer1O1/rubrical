package pages

import (
	"fmt"
	"strings"
	"time"

	"rubrical/internal/analysis"
)

type AnalysisFeedbackView struct {
	ID          int64
	Category    string
	Severity    string
	Title       string
	Explanation string
	Evidence    string
	Suggestion  string
}

type AnalysisResultsView struct {
	HasResults        bool
	OverallSummary    string
	ScoreLabel        string
	ConfidenceLabel   string
	CompletedAtLabel  string
	Criteria          []AnalysisFeedbackView
	MissingRequirements []AnalysisFeedbackView
	Strengths         []AnalysisFeedbackView
	Suggestions       []AnalysisFeedbackView
}

func AnalysisResultsFromResult(result *analysis.Result) AnalysisResultsView {
	if result == nil {
		return AnalysisResultsView{}
	}

	view := AnalysisResultsView{
		HasResults:       true,
		OverallSummary:   result.OverallSummary,
		ScoreLabel:       formatScoreLabel(result.EstimatedScore, result.EstimatedScoreMax),
		ConfidenceLabel:  formatConfidenceLabel(result.Confidence),
		CompletedAtLabel: formatAnalysisTime(result.CompletedAt),
	}

	for _, item := range result.Feedback {
		feedback := AnalysisFeedbackView{
			ID:          item.ID,
			Category:    item.Category,
			Severity:    item.Severity,
			Title:       item.Title,
			Explanation: item.Explanation,
			Evidence:    item.Evidence,
			Suggestion:  item.Suggestion,
		}
		switch item.Category {
		case "criterion":
			view.Criteria = append(view.Criteria, feedback)
		case "missing_requirement":
			view.MissingRequirements = append(view.MissingRequirements, feedback)
		case "strength":
			view.Strengths = append(view.Strengths, feedback)
		case "suggestion":
			view.Suggestions = append(view.Suggestions, feedback)
		}
	}

	return view
}

func formatScoreLabel(score, max *float64) string {
	if score == nil && max == nil {
		return ""
	}
	if score != nil && max != nil {
		return fmt.Sprintf("Estimated score: %.1f / %.1f", *score, *max)
	}
	if score != nil {
		return fmt.Sprintf("Estimated score: %.1f", *score)
	}
	return fmt.Sprintf("Score out of %.1f", *max)
}

func formatConfidenceLabel(confidence string) string {
	switch strings.ToLower(strings.TrimSpace(confidence)) {
	case "low":
		return "Confidence: low"
	case "medium":
		return "Confidence: medium"
	case "high":
		return "Confidence: high"
	default:
		return ""
	}
}

func formatAnalysisTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2, 2006 3:04 PM")
}

func feedbackSeverityClass(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return "border-red-200 bg-red-50"
	case "warning":
		return "border-amber-200 bg-amber-50"
	default:
		return "border-stone-200 bg-white"
	}
}

func feedbackCategoryLabel(category string) string {
	switch category {
	case "criterion":
		return "Rubric criterion"
	case "missing_requirement":
		return "Missing requirement"
	case "strength":
		return "Strength"
	case "suggestion":
		return "Suggestion"
	default:
		return category
	}
}
