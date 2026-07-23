package pages

import (
	"strings"
	"testing"

	analysispipeline "rubrical/internal/analysispipeline"
	"rubrical/internal/analysispipeline/analysis"
)

func TestFormatScoreLabel_checkableOnly(t *testing.T) {
	score, max := 7.0, 9.0
	got := formatScoreLabel(&score, &max, 2)
	if !strings.Contains(got, "7.0 / 9.0") || !strings.Contains(got, "checkable") || !strings.Contains(got, "2 criteria") {
		t.Fatalf("got %q", got)
	}
}

func TestAnalysisResultsFromResult_notAnalyzable(t *testing.T) {
	result := &analysispipeline.Result{
		OverallSummary:    "Mixed.",
		Confidence:        "high",
		PredictedScore:    floatPtr(1),
		PredictedScoreMax: floatPtr(1),
		Feedback: []analysispipeline.FeedbackItem{
			{
				Category:        "criterion",
				Title:           "Topic",
				CriterionStatus: "met",
				ScoreRationale:  "Good.",
				PredictedPoints: floatPtr(1),
				MaxPoints:       floatPtr(1),
				CriterionScore:  floatPtr(1),
			},
			{
				Category:        "criterion",
				Title:           "Classmate Reply",
				CriterionStatus: "not_analyzable",
				ScoreRationale:  "Peer reply not in this draft.",
				HowToEarnPoints: "Reply to a classmate in Canvas.",
				Explanation:     "Reply to a classmate in Canvas.",
				MaxPoints:       floatPtr(1),
			},
		},
	}
	view := AnalysisResultsFromResult(result, analysis.RubricContext{})
	if !strings.Contains(view.ScoreLabel, "checkable") {
		t.Fatalf("score label = %q", view.ScoreLabel)
	}
	if len(view.Criteria) != 2 {
		t.Fatalf("criteria = %d", len(view.Criteria))
	}
	na := view.Criteria[1]
	if na.ShowScale || na.PointsLabel != "" {
		t.Fatalf("not analyzable card should hide scale/points: %+v", na)
	}
	if na.HowToEarnPoints == "" {
		t.Fatal("expected howToEarnPoints")
	}
	if criterionStatusLabel("not_analyzable") != "Not analyzable" {
		t.Fatal(criterionStatusLabel("not_analyzable"))
	}
}

func floatPtr(v float64) *float64 { return &v }
