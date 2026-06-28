package pages

import (
	"fmt"

	"rubrical/internal/analysis"
)

type RatingCellView struct {
	Title    string
	Points   string
	Selected bool
}

type CriterionScaleView struct {
	HasBands     bool
	ArrowPercent float64
	Ratings      []RatingCellView
}

type AnalysisFeedbackView struct {
	ID              int64
	Category        string
	Severity        string
	Title           string
	Explanation     string
	Evidence        string
	Suggestion      string
	CriterionStatus string
	SelectedRating  string
	PointsLabel     string
	Scale           CriterionScaleView
}

type AnalysisResultsView struct {
	HasResults          bool
	OverallSummary      string
	ScoreLabel          string
	ConfidenceLabel     string
	CompletedAtLabel    string
	Criteria            []AnalysisFeedbackView
	MissingRequirements []AnalysisFeedbackView
	Strengths           []AnalysisFeedbackView
	Suggestions         []AnalysisFeedbackView
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
			ID:              item.ID,
			Category:        item.Category,
			Severity:        item.Severity,
			Title:           item.Title,
			Explanation:     item.Explanation,
			Evidence:        item.Evidence,
			Suggestion:      item.Suggestion,
			CriterionStatus: item.CriterionStatus,
			SelectedRating:  item.SelectedRating,
			PointsLabel:     formatPointsLabel(item.PredictedPoints, item.MaxPoints),
		}
		switch item.Category {
		case "criterion":
			feedback.Scale = buildCriterionScale(rubric, item)
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

func buildCriterionScale(rubric analysis.RubricContext, item analysis.FeedbackItem) CriterionScaleView {
	row, ok := analysis.MatchRubricRow(rubric, item.Title)
	if !ok {
		return fallbackScale(item)
	}

	score := 0.0
	if item.CriterionScore != nil {
		score = *item.CriterionScore
	}
	arrow := analysis.ArrowPercentForScore(row, score)

	bands := analysis.RatingBandsForUI(row)
	if len(bands) == 0 {
		return CriterionScaleView{
			HasBands:     false,
			ArrowPercent: arrow,
			Ratings: []RatingCellView{{
				Title:    "Predicted",
				Points:   formatPointsOnly(item.PredictedPoints, item.MaxPoints),
				Selected: true,
			}},
		}
	}

	want := normalizeLabel(item.SelectedRating)
	cells := make([]RatingCellView, len(bands))
	for i, band := range bands {
		title := band.Title
		cells[i] = RatingCellView{
			Title:    title,
			Points:   formatBandPoints(band.Points),
			Selected: normalizeLabel(title) == want,
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
		ArrowPercent: (1 - score) * 100,
		Ratings: []RatingCellView{{
			Title:    item.SelectedRating,
			Points:   formatPointsOnly(item.PredictedPoints, item.MaxPoints),
			Selected: true,
		}},
	}
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
