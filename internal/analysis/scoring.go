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
	if err := validateCriteriaCoverage(resp, rubric); err != nil {
		return nil, err
	}

	criteria := resp.Criteria
	if len(rubric.Rows) > 0 {
		criteria = orderCriteriaByRubric(resp.Criteria, rubric)
	}

	scored := &schema.ScoredAnalysis{
		OverallSummary:      resp.OverallSummary,
		Confidence:          resp.Confidence,
		Strengths:           resp.Strengths,
		Guidance:            resp.Guidance,
		Criteria:            make([]schema.ScoredCriterion, len(criteria)),
	}

	var total, totalMax float64

	for i := range criteria {
		assessment := criteria[i]
		row, ok := matchRubricRow(rubric, assessment.CriterionName)
		if !ok {
			return nil, fmt.Errorf("criteria[%d] criterionName %q not found in rubric", i, assessment.CriterionName)
		}
		if assessment.BandPosition < 0 || assessment.BandPosition > 100 {
			return nil, fmt.Errorf("criteria[%d] %q: bandPosition must be between 0 and 100", i, assessment.CriterionName)
		}

		maxPts, ok := criterionMaxPoints(row)
		if !ok {
			return nil, fmt.Errorf("criteria[%d] %q: could not determine max points from rubric", i, assessment.CriterionName)
		}

		bands := parseRatingBands(row.Ratings)
		var score float64
		var title string
		var pts float64
		var status string

		if len(bands) == 0 {
			score = float64(assessment.BandPosition) / 100
			pts = roundScore(score * maxPts)
			title = strings.TrimSpace(assessment.SelectedRating)
			status = criterionStatusFromScore(score, 0)
		} else {
			bandIdx, band, err := matchRatingBand(row, assessment.SelectedRating)
			if err != nil {
				return nil, fmt.Errorf("criteria[%d] %q: %w", i, assessment.CriterionName, err)
			}
			score = continuousScore(bandIdx, len(bands), assessment.BandPosition)
			title = strings.TrimSpace(band.rating.Title)
			pts = band.points
			status = criterionStatusFromBandIndex(bandIdx, len(bands))
		}

		scored.Criteria[i] = schema.ScoredCriterion{
			CriterionName:           assessment.CriterionName,
			CriterionScore:          score,
			ScoreRationale:          assessment.ScoreRationale,
			FulfilledRequirements:   append([]schema.FulfilledRequirement(nil), assessment.FulfilledRequirements...),
			UnfulfilledRequirements: append([]schema.UnfulfilledRequirement(nil), assessment.UnfulfilledRequirements...),
			Status:                  status,
			SelectedRating:          title,
			MaxPoints:               floatPtr(maxPts),
			PredictedPoints:         floatPtr(roundScore(pts)),
		}
		total += pts
		totalMax += maxPts
	}

	scored.PredictedScore = floatPtr(roundScore(total))
	scored.PredictedScoreMax = floatPtr(roundScore(totalMax))
	return scored, nil
}

func validateCriteriaCoverage(resp *schema.ProviderResponse, rubric RubricContext) error {
	if len(rubric.Rows) == 0 {
		return nil
	}
	if len(resp.Criteria) != len(rubric.Rows) {
		return fmt.Errorf("expected %d criteria (one per rubric row), got %d", len(rubric.Rows), len(resp.Criteria))
	}

	seen := make(map[string]bool, len(resp.Criteria))
	for i, assessment := range resp.Criteria {
		if _, ok := matchRubricRow(rubric, assessment.CriterionName); !ok {
			return fmt.Errorf("criteria[%d] criterionName %q not found in rubric", i, assessment.CriterionName)
		}
		key := normalizeCriterionLabel(assessment.CriterionName)
		if seen[key] {
			return fmt.Errorf("criteria[%d] duplicate criterion %q", i, assessment.CriterionName)
		}
		seen[key] = true
	}

	for _, row := range rubric.Rows {
		key := normalizeCriterionLabel(row.Criterion)
		if !seen[key] {
			return fmt.Errorf("missing criterion %q", row.Criterion)
		}
	}
	return nil
}

func orderCriteriaByRubric(assessments []schema.CriterionAssessment, rubric RubricContext) []schema.CriterionAssessment {
	byName := make(map[string]schema.CriterionAssessment, len(assessments))
	for _, assessment := range assessments {
		byName[normalizeCriterionLabel(assessment.CriterionName)] = assessment
	}
	ordered := make([]schema.CriterionAssessment, 0, len(rubric.Rows))
	for _, row := range rubric.Rows {
		ordered = append(ordered, byName[normalizeCriterionLabel(row.Criterion)])
	}
	return ordered
}

// continuousScore maps band index + within-band score to [0,1] for the gradient arrow.
// Band i occupies [i/n, (i+1)/n]; bandPosition 0–100 picks a point inside that slice.
func continuousScore(bandIdx, bandCount, bandPosition int) float64 {
	if bandCount <= 0 {
		return float64(bandPosition) / 100
	}
	return (float64(bandIdx) + float64(bandPosition)/100) / float64(bandCount)
}

func matchRatingBand(row RubricRow, title string) (int, scoredBand, error) {
	bands := parseRatingBands(row.Ratings)
	want := normalizeCriterionLabel(title)
	for i, b := range bands {
		if normalizeCriterionLabel(b.rating.Title) == want {
			return i, b, nil
		}
	}
	return 0, scoredBand{}, fmt.Errorf("selectedRating %q not found in rubric", strings.TrimSpace(title))
}

func criterionStatusFromBandIndex(bandIdx, bandCount int) string {
	if bandCount <= 1 {
		return "met"
	}
	switch {
	case bandIdx == 0:
		return "not_met"
	case bandIdx == bandCount-1:
		return "met"
	default:
		return "partially_met"
	}
}

// criterionStatusFromScore for point-only rows without rating bands.
func criterionStatusFromScore(score float64, bandCount int) string {
	if bandCount >= 2 {
		return criterionStatusFromBandIndex(bandIndex(score, bandCount), bandCount)
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
