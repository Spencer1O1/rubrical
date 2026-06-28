package analysis

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"rubrical/internal/analysis/schema"
	"rubrical/internal/importmeta"
)

var criterionPointsSuffixRE = regexp.MustCompile(`(?i)\s*\(\s*\d+(?:\.\d+)?\s*(?:pts?\.?|points?)\s*\)\s*$`)

type scoredBand struct {
	rating RubricRating
	points float64
}

type RatingBandUI struct {
	Title       string
	Description string
	Points      float64
}

func ApplyRubricScoring(resp *schema.ProviderResponse, rubric RubricContext) (*schema.ScoredAnalysis, error) {
	if resp == nil {
		return nil, fmt.Errorf("provider response is nil")
	}

	scored := &schema.ScoredAnalysis{
		OverallSummary:      resp.OverallSummary,
		Confidence:          resp.Confidence,
		MissingRequirements: resp.MissingRequirements,
		Strengths:           resp.Strengths,
		RevisionSuggestions: resp.RevisionSuggestions,
		Criteria:            make([]schema.ScoredCriterion, len(resp.Criteria)),
	}

	var total, totalMax float64

	for i := range resp.Criteria {
		assessment := resp.Criteria[i]
		row, ok := matchRubricRow(rubric, assessment.CriterionName)
		if !ok {
			return nil, fmt.Errorf("criteria[%d] criterionName %q not found in rubric", i, assessment.CriterionName)
		}
		if assessment.CriterionScore < 0 || assessment.CriterionScore > 1 {
			return nil, fmt.Errorf("criteria[%d] %q: criterionScore must be between 0 and 1", i, assessment.CriterionName)
		}

		maxPts, ok := criterionMaxPoints(row)
		if !ok {
			return nil, fmt.Errorf("criteria[%d] %q: could not determine max points from rubric", i, assessment.CriterionName)
		}

		title, pts := scoreToBand(row, assessment.CriterionScore)
		scored.Criteria[i] = schema.ScoredCriterion{
			CriterionAssessment: assessment,
			Status:              criterionStatusFromScore(assessment.CriterionScore, len(parseRatingBands(row.Ratings))),
			SelectedRating:      title,
			MaxPoints:           floatPtr(maxPts),
			PredictedPoints:     floatPtr(roundScore(pts)),
		}
		total += pts
		totalMax += maxPts
	}

	scored.PredictedScore = floatPtr(roundScore(total))
	scored.PredictedScoreMax = floatPtr(roundScore(totalMax))
	return scored, nil
}

// scoreToBand maps criterionScore to a rubric band using equal [0,1) slices per band (worst→best by points).
// Point-only rows (no bands): linear score × maxPoints.
func scoreToBand(row RubricRow, score float64) (title string, points float64) {
	maxPts, _ := criterionMaxPoints(row)
	bands := parseRatingBands(row.Ratings)
	if len(bands) == 0 {
		return "", roundScore(score * maxPts)
	}
	b := bands[bandIndex(score, len(bands))]
	return strings.TrimSpace(b.rating.Title), b.points
}

// bandIndex: n bands divide [0,1] into n equal slices; band 0 = lowest points, band n-1 = highest.
func bandIndex(score float64, n int) int {
	if n <= 1 {
		return 0
	}
	if score >= 1 {
		return n - 1
	}
	idx := int(score * float64(n))
	if idx >= n {
		return n - 1
	}
	return idx
}

// criterionStatusFromScore maps score to met | partially_met | not_met.
// Banded rows: bottom band → not_met, top band → met, middle → partially_met.
// Point-only rows: [0, 0.25) not_met, [0.25, 0.75) partially_met, [0.75, 1] met.
func criterionStatusFromScore(score float64, bandCount int) string {
	if bandCount >= 2 {
		idx := bandIndex(score, bandCount)
		switch {
		case idx == 0:
			return "not_met"
		case idx == bandCount-1:
			return "met"
		default:
			return "partially_met"
		}
	}
	switch {
	case score >= 0.75:
		return "met"
	case score >= 0.25:
		return "partially_met"
	default:
		return "not_met"
	}
}

func parseRatingBands(ratings []RubricRating) []scoredBand {
	var bands []scoredBand
	for _, rating := range ratings {
		if pts, ok := ratingPoints(rating); ok {
			bands = append(bands, scoredBand{rating: rating, points: pts})
		}
	}
	sort.Slice(bands, func(i, j int) bool { return bands[i].points < bands[j].points })
	return bands
}

func RatingBandsForUI(row RubricRow) []RatingBandUI {
	bands := parseRatingBands(row.Ratings)
	sort.Slice(bands, func(i, j int) bool { return bands[i].points > bands[j].points })
	out := make([]RatingBandUI, len(bands))
	for i, b := range bands {
		out[i] = RatingBandUI{Title: b.rating.Title, Description: b.rating.Description, Points: b.points}
	}
	return out
}

// ArrowPercentForScore: score 1 → 0% (left/green), score 0 → 100% (right/red).
func ArrowPercentForScore(score float64) float64 {
	score = clamp01(score)
	return (1 - score) * 100
}

func matchRubricRow(rubric RubricContext, name string) (RubricRow, bool) {
	want := normalizeCriterionLabel(name)
	for _, row := range rubric.Rows {
		if normalizeCriterionLabel(row.Criterion) == want {
			return row, true
		}
	}
	return RubricRow{}, false
}

func MatchRubricRow(rubric RubricContext, name string) (RubricRow, bool) {
	return matchRubricRow(rubric, name)
}

func criterionMaxPoints(row RubricRow) (float64, bool) {
	bands := parseRatingBands(row.Ratings)
	if len(bands) > 0 {
		return bands[len(bands)-1].points, true
	}
	return importmeta.ParsePointsPossible(row.Points)
}

func ratingPoints(rating RubricRating) (float64, bool) {
	if pts, ok := importmeta.ParsePointsPossible(rating.Points); ok {
		return pts, true
	}
	return importmeta.ParsePointsPossible(rating.Title)
}

func NormalizeCriterionLabel(s string) string {
	s = strings.TrimSpace(criterionPointsSuffixRE.ReplaceAllString(s, ""))
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

func normalizeCriterionLabel(s string) string { return NormalizeCriterionLabel(s) }

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

func roundScore(v float64) float64 { return math.Round(v*10) / 10 }

func floatPtr(v float64) *float64 { return &v }
