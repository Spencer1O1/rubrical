package pages

import (
	"fmt"
	"strings"

	"rubrical/internal/analysis"
	"rubrical/internal/analysis/schema"
)

type RatingCellView struct {
	Label       string
	Description string
	Selected    bool
}

type CriterionScaleView struct {
	HasBands     bool
	ArrowPercent float64
	Ratings      []RatingCellView
}

type FulfilledRequirementView struct {
	Requirement string
	Evidence    string
}

type UnfulfilledRequirementView struct {
	Requirement string
	Severity    string
	Explanation string
	Suggestion  string
}

type AnalysisFeedbackView struct {
	ID                      int64
	Category                string
	Severity                string
	Title                   string
	Explanation             string
	ScoreRationale          string
	FulfilledRequirements   []FulfilledRequirementView
	UnfulfilledRequirements []UnfulfilledRequirementView
	CriterionStatus         string
	SelectedRating          string
	PointsLabel             string
	Scale                   CriterionScaleView
}

type AnalysisResultsView struct {
	HasResults          bool
	OverallSummary      string
	ScoreLabel          string
	ConfidenceLabel     string
	CompletedAtLabel    string
	Criteria            []AnalysisFeedbackView
	Strengths           []AnalysisFeedbackView
	Guidance            []AnalysisFeedbackView
}

func AnalysisResultsFromResult(result *analysis.Result, rubric analysis.RubricContext) AnalysisResultsView {
	if result == nil {
		return AnalysisResultsView{}
	}

	view := AnalysisResultsView{
		HasResults:       true,
		OverallSummary:   result.OverallSummary,
		ScoreLabel:       formatScoreLabel(result.PredictedScore, result.PredictedScoreMax),
		ConfidenceLabel:  formatConfidenceLabel(result.Confidence),
		CompletedAtLabel: formatAnalysisTime(result.CompletedAt),
	}

	for _, item := range result.Feedback {
		feedback := AnalysisFeedbackView{
			ID:                      item.ID,
			Category:                item.Category,
			Severity:                item.Severity,
			Title:                   item.Title,
			Explanation:             item.Explanation,
			ScoreRationale:          item.ScoreRationale,
			FulfilledRequirements:   mapFulfilledRequirements(item.FulfilledRequirements),
			UnfulfilledRequirements: mapUnfulfilledRequirements(item.UnfulfilledRequirements),
			CriterionStatus:         item.CriterionStatus,
			SelectedRating:          item.SelectedRating,
			PointsLabel:             formatPointsLabel(item.PredictedPoints, item.MaxPoints),
		}
		switch item.Category {
		case "criterion":
			feedback.Scale = buildCriterionScale(rubric, item)
			view.Criteria = append(view.Criteria, feedback)
		case "strength":
			view.Strengths = append(view.Strengths, feedback)
		case "guidance":
			view.Guidance = append(view.Guidance, feedback)
		}
	}

	return view
}

func mapFulfilledRequirements(items []schema.FulfilledRequirement) []FulfilledRequirementView {
	out := make([]FulfilledRequirementView, len(items))
	for i, item := range items {
		out[i] = FulfilledRequirementView{
			Requirement: item.Requirement,
			Evidence:    item.Evidence,
		}
	}
	return out
}

func mapUnfulfilledRequirements(items []schema.UnfulfilledRequirement) []UnfulfilledRequirementView {
	out := make([]UnfulfilledRequirementView, len(items))
	for i, item := range items {
		out[i] = UnfulfilledRequirementView{
			Requirement: item.Requirement,
			Severity:    item.Severity,
			Explanation: item.Explanation,
			Suggestion:  item.Suggestion,
		}
	}
	return out
}

func buildCriterionScale(rubric analysis.RubricContext, item analysis.FeedbackItem) CriterionScaleView {
	row, ok := analysis.MatchRubricRow(rubric, item.Title)
	if !ok {
		return fallbackScale(item)
	}

	score := 0.0
	if item.CriterionScore != nil {
		score = *item.CriterionScore
	}
	arrow := analysis.ArrowPercentForScore(score)

	bands := analysis.RatingBandsForUI(row)
	if len(bands) == 0 {
		return CriterionScaleView{
			HasBands:     false,
			ArrowPercent: arrow,
			Ratings: []RatingCellView{{
				Label:    formatRatingLabel("Predicted", formatPointsOnly(item.PredictedPoints, item.MaxPoints)),
				Selected: true,
			}},
		}
	}

	want := normalizeLabel(item.SelectedRating)
	cells := make([]RatingCellView, len(bands))
	for i, band := range bands {
		cells[i] = RatingCellView{
			Label:       formatRatingLabel(band.Title, formatBandPoints(band.Points)),
			Description: strings.TrimSpace(band.Description),
			Selected:    normalizeLabel(band.Title) == want,
		}
	}

	return CriterionScaleView{
		HasBands:     true,
		ArrowPercent: arrow,
		Ratings:      cells,
	}
}

func fallbackScale(item analysis.FeedbackItem) CriterionScaleView {
	score := 0.0
	if item.CriterionScore != nil {
		score = *item.CriterionScore
	}
	return CriterionScaleView{
		HasBands:     false,
		ArrowPercent: analysis.ArrowPercentForScore(score),
		Ratings: []RatingCellView{{
			Label:    formatRatingLabel(item.SelectedRating, formatPointsOnly(item.PredictedPoints, item.MaxPoints)),
			Selected: true,
		}},
	}
}

func formatRatingLabel(title, points string) string {
	title = strings.TrimSpace(title)
	points = strings.TrimSpace(points)
	if title == "" {
		return points
	}
	if points == "" {
		return title
	}
	// Title may already be a normalized points label ("8 pts").
	if strings.Contains(strings.ToLower(title), "pt") {
		return title
	}
	return fmt.Sprintf("%s (%s)", title, points)
}

func formatBandPoints(pts float64) string {
	if pts == float64(int(pts)) {
		return fmt.Sprintf("%d", int(pts))
	}
	return fmt.Sprintf("%.1f", pts)
}

func formatPointsOnly(predicted, max *float64) string {
	if predicted != nil && max != nil {
		return fmt.Sprintf("%.1f / %.1f", *predicted, *max)
	}
	if predicted != nil {
		return fmt.Sprintf("%.1f", *predicted)
	}
	return ""
}

func normalizeLabel(s string) string {
	return analysis.NormalizeCriterionLabel(s)
}
