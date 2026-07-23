package analysispipeline

import (
	"fmt"
	"strings"

	"rubrical/internal/analysispipeline/analysis"
	analysisschema "rubrical/internal/analysispipeline/analysis/schema"
	"rubrical/internal/analysispipeline/checkability"
)

// MergeAnalysis combines pass-1 classifications with pass-2 scored rows.
// fullRubric must be the complete rubric; scored may be nil when nothing was checkable.
func MergeAnalysis(
	class *checkability.Response,
	scored *analysisschema.ScoredAnalysis,
	fullRubric analysis.RubricContext,
) (*analysisschema.ScoredAnalysis, error) {
	if class == nil {
		return nil, fmt.Errorf("checkability response is nil")
	}
	if len(fullRubric.Rows) == 0 {
		if scored != nil {
			return scored, nil
		}
		return nil, fmt.Errorf("empty rubric")
	}
	refs := fullRubric.AssignCriterionIDs()
	if err := checkability.ValidateResponse(class, refs); err != nil {
		return nil, err
	}
	if scored != nil {
		for _, c := range scored.Criteria {
			if strings.TrimSpace(c.CriterionID) == "" {
				return nil, fmt.Errorf("scored criterion missing criterionId")
			}
		}
	}

	scoredByID := map[string]analysisschema.ScoredCriterion{}
	var overall string
	var confidence string
	var strengths, guidance []string
	if scored != nil {
		overall = scored.OverallSummary
		confidence = scored.Confidence
		strengths = scored.Strengths
		guidance = scored.Guidance
		for _, c := range scored.Criteria {
			scoredByID[c.CriterionID] = c
		}
	}

	out := &analysisschema.ScoredAnalysis{
		OverallSummary: overall,
		Confidence:     confidence,
		Strengths:      strengths,
		Guidance:       guidance,
		Criteria:       make([]analysisschema.ScoredCriterion, len(fullRubric.Rows)),
	}

	var total, totalMax float64
	unchecked := 0

	for i, row := range fullRubric.Rows {
		cls := class.Criteria[i]
		maxPts, _ := analysis.CriterionMaxPoints(row)

		if !cls.Checkable() {
			unchecked++
			out.Criteria[i] = analysisschema.ScoredCriterion{
				CriterionID:     row.ID,
				CriterionName:   row.Criterion,
				CriterionScore:  0,
				ScoreRationale:  cls.Reason,
				HowToEarnPoints: cls.HowToEarnPoints,
				Status:          "not_checkable",
				MaxPoints:       analysis.FloatPtr(maxPts),
			}
			continue
		}

		scoredRow, ok := scoredByID[row.ID]
		if !ok {
			return nil, fmt.Errorf("missing scored criterion %q", row.ID)
		}
		out.Criteria[i] = scoredRow
		if scoredRow.PredictedPoints != nil {
			total += *scoredRow.PredictedPoints
		}
		if scoredRow.MaxPoints != nil {
			totalMax += *scoredRow.MaxPoints
		} else {
			totalMax += maxPts
		}
	}

	if scored == nil {
		out.OverallSummary = notCheckableOnlySummary(unchecked, len(fullRubric.Rows))
		out.Confidence = "medium"
		if out.Guidance == nil {
			out.Guidance = []string{}
		}
		if out.Strengths == nil {
			out.Strengths = []string{}
		}
	}

	if totalMax > 0 || unchecked < len(fullRubric.Rows) {
		out.PredictedScore = analysis.FloatPtr(analysis.RoundScore(total))
		out.PredictedScoreMax = analysis.FloatPtr(analysis.RoundScore(totalMax))
	}
	if unchecked == len(fullRubric.Rows) {
		out.PredictedScore = analysis.FloatPtr(0)
		out.PredictedScoreMax = analysis.FloatPtr(0)
	}

	return out, nil
}

func notCheckableOnlySummary(unchecked, total int) string {
	if total == 1 {
		return "Rubrical couldn’t check this rubric criterion from this draft. See the guidance below so you can still earn those points."
	}
	return fmt.Sprintf(
		"Rubrical couldn’t check any of the %d rubric criteria from this draft. See each criterion for how to earn those points outside Rubrical.",
		total,
	)
}

func filterRubric(rubric analysis.RubricContext, class *checkability.Response) analysis.RubricContext {
	if class == nil {
		return rubric
	}
	var rows []analysis.RubricRow
	for i, row := range rubric.Rows {
		if i < len(class.Criteria) && class.Criteria[i].Checkable() {
			rows = append(rows, row)
		}
	}
	return analysis.RubricContext{Header: rubric.Header, Rows: rows}
}
