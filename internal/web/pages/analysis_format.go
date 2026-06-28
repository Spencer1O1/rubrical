package pages

import (
	"fmt"
	"strings"
	"time"
)

func formatScoreLabel(score, max *float64) string {
	if score == nil && max == nil {
		return ""
	}
	if score != nil && max != nil {
		return fmt.Sprintf("Predicted score: %.1f / %.1f", *score, *max)
	}
	if score != nil {
		return fmt.Sprintf("Predicted score: %.1f", *score)
	}
	return fmt.Sprintf("Out of %.1f pts", *max)
}

func formatPointsLabel(predicted, max *float64) string {
	if predicted != nil && max != nil {
		return fmt.Sprintf("%.1f / %.1f", *predicted, *max)
	}
	if max != nil {
		return fmt.Sprintf("— / %.1f", *max)
	}
	if predicted != nil {
		return fmt.Sprintf("%.1f", *predicted)
	}
	return ""
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

func criterionStatusLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "met":
		return "Met"
	case "partially_met":
		return "Partial"
	case "not_met":
		return "Not met"
	default:
		return status
	}
}

func criterionStatusClass(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "met":
		return "bg-emerald-100 text-emerald-800"
	case "partially_met":
		return "bg-amber-100 text-amber-900"
	case "not_met":
		return "bg-red-100 text-red-800"
	default:
		return "bg-stone-100 text-stone-700"
	}
}

func IsNoChangeSuggestion(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	return lower == "no major change needed." || lower == "no change needed." || lower == "none."
}

func formatArrowStyle(percent float64) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	// Keep the 12px-wide arrow tip inside the bar at 0% and 100%.
	return fmt.Sprintf("left: clamp(6px, %.2f%%, calc(100%% - 6px))", percent)
}
