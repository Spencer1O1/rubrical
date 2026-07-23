package analysis

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"rubrical/internal/analysispipeline/analysis/schema"
	"rubrical/internal/analysispipeline/criterion"
	"rubrical/internal/importmeta"
)

type scoredBand struct {
	rating RubricRating
	points float64
}

func ApplyRubricScoring(resp *schema.ProviderResponse, rubric RubricContext) (*schema.ScoredAnalysis, error) {
	if resp == nil {
		return nil, fmt.Errorf("provider response is nil")
	}
	if err := ensureRubricIDs(&rubric); err != nil {
		return nil, err
	}
	if err := validateCriteriaCoverage(resp, rubric); err != nil {
		return nil, err
	}

	criteria := resp.Criteria
	if len(rubric.Rows) > 0 {
		criteria = orderCriteriaByRubric(resp.Criteria, rubric)
	}

	scored := &schema.ScoredAnalysis{
		OverallSummary: resp.OverallSummary,
		Confidence:     resp.Confidence,
		Strengths:      resp.Strengths,
		Guidance:       resp.Guidance,
		Criteria:       make([]schema.ScoredCriterion, len(criteria)),
	}

	var total, totalMax float64
	byID := map[string]RubricRow{}
	for _, row := range rubric.Rows {
		byID[row.ID] = row
	}

	for i := range criteria {
		assessment := criteria[i]
		row := byID[assessment.CriterionID]
		if assessment.BandPosition < 0 || assessment.BandPosition > 100 {
			return nil, fmt.Errorf("criteria[%d] %q: bandPosition must be between 0 and 100", i, assessment.CriterionID)
		}

		maxPts, ok := CriterionMaxPoints(row)
		if !ok {
			return nil, fmt.Errorf("criteria[%d] %q: could not determine max points from rubric", i, assessment.CriterionID)
		}

		bands := parseRatingBands(row.Ratings)
		var score float64
		var title string
		var pts float64
		var status string

		if len(bands) == 0 {
			score = float64(assessment.BandPosition) / 100
			pts = roundScore(score * maxPts)
			title = ""
			status = criterionStatusFromScore(score)
		} else {
			bandIdx, band, err := matchRatingBandByID(row, assessment.SelectedRatingID)
			if err != nil {
				return nil, fmt.Errorf("criteria[%d] %q: %w", i, assessment.CriterionID, err)
			}
			score = continuousScore(bandIdx, len(bands), assessment.BandPosition)
			title = strings.TrimSpace(band.rating.Title)
			pts = band.points
			status = criterionStatusFromBandIndex(bandIdx, len(bands))
		}

		scored.Criteria[i] = schema.ScoredCriterion{
			CriterionID:             assessment.CriterionID,
			CriterionName:           row.Criterion,
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

func ensureRubricIDs(rubric *RubricContext) error {
	if rubric == nil || len(rubric.Rows) == 0 {
		return nil
	}
	for _, row := range rubric.Rows {
		if strings.TrimSpace(row.ID) == "" {
			rubric.AssignCriterionIDs()
			return nil
		}
	}
	return nil
}

func validateCriteriaCoverage(resp *schema.ProviderResponse, rubric RubricContext) error {
	if len(rubric.Rows) == 0 {
		return nil
	}
	if len(resp.Criteria) != len(rubric.Rows) {
		return fmt.Errorf("expected %d criteria (one per rubric row), got %d", len(rubric.Rows), len(resp.Criteria))
	}

	want := make(map[string]bool, len(rubric.Rows))
	for _, row := range rubric.Rows {
		want[row.ID] = true
	}
	seen := make(map[string]bool, len(resp.Criteria))
	for i, assessment := range resp.Criteria {
		id := strings.TrimSpace(assessment.CriterionID)
		if !want[id] {
			return fmt.Errorf("criteria[%d] criterionId %q not found in rubric", i, assessment.CriterionID)
		}
		if seen[id] {
			return fmt.Errorf("criteria[%d] duplicate criterionId %q", i, assessment.CriterionID)
		}
		seen[id] = true
	}
	for _, row := range rubric.Rows {
		if !seen[row.ID] {
			return fmt.Errorf("missing criterionId %q", row.ID)
		}
	}
	return nil
}

func orderCriteriaByRubric(assessments []schema.CriterionAssessment, rubric RubricContext) []schema.CriterionAssessment {
	byID := make(map[string]schema.CriterionAssessment, len(assessments))
	for _, assessment := range assessments {
		byID[assessment.CriterionID] = assessment
	}
	ordered := make([]schema.CriterionAssessment, 0, len(rubric.Rows))
	for _, row := range rubric.Rows {
		ordered = append(ordered, byID[row.ID])
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

func matchRatingBandByID(row RubricRow, id string) (int, scoredBand, error) {
	bands := parseRatingBands(row.Ratings)
	id = strings.TrimSpace(id)
	for i, b := range bands {
		if criterion.RatingID(i) == id {
			return i, b, nil
		}
	}
	return 0, scoredBand{}, fmt.Errorf("selectedRatingId %q not found in rubric", id)
}

// ratingSelectionLabel is the display title for a band in prompts/UI.
// Prefer Canvas title; if empty, use points; else a short description.
func ratingSelectionLabel(rating RubricRating) string {
	if title := strings.TrimSpace(rating.Title); title != "" {
		return title
	}
	if pts, ok := ratingPoints(rating); ok {
		return formatRatingPointsLabel(pts)
	}
	if desc := strings.TrimSpace(rating.Description); desc != "" {
		return truncateSelectionLabel(desc, 48)
	}
	return ""
}

func formatRatingPointsLabel(pts float64) string {
	if pts == float64(int64(pts)) {
		return fmt.Sprintf("%d pts", int64(pts))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", pts), "0"), ".") + " pts"
}

func truncateSelectionLabel(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if max <= 0 || len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max]) + "…"
}

// uniquifySelectionLabels ensures display titles are unique within a criterion.
func uniquifySelectionLabels(bands []scoredBand) {
	used := make(map[string]bool, len(bands))
	for i := range bands {
		label := strings.TrimSpace(bands[i].rating.Title)
		if label == "" {
			label = formatRatingPointsLabel(bands[i].points)
		}
		candidate := label
		if used[strings.ToLower(candidate)] {
			candidate = label + " · " + formatRatingPointsLabel(bands[i].points)
		}
		n := 2
		for used[strings.ToLower(candidate)] {
			candidate = fmt.Sprintf("%s (%d)", label, n)
			n++
		}
		bands[i].rating.Title = candidate
		used[strings.ToLower(candidate)] = true
	}
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
func criterionStatusFromScore(score float64) string {
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
			normalized := rating
			normalized.Title = ratingSelectionLabel(rating)
			bands = append(bands, scoredBand{rating: normalized, points: pts})
		}
	}
	sort.Slice(bands, func(i, j int) bool { return bands[i].points < bands[j].points })
	uniquifySelectionLabels(bands)
	return bands
}

func matchRubricRow(rubric RubricContext, name string) (RubricRow, bool) {
	want := strings.TrimSpace(name)
	for _, row := range rubric.Rows {
		if strings.TrimSpace(row.Criterion) == want {
			return row, true
		}
	}
	return RubricRow{}, false
}

func MatchRubricRow(rubric RubricContext, name string) (RubricRow, bool) {
	return matchRubricRow(rubric, name)
}

func CriterionMaxPoints(row RubricRow) (float64, bool) {
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

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

func roundScore(v float64) float64 { return math.Round(v*10) / 10 }

func FloatPtr(v float64) *float64 { return &v }

func floatPtr(v float64) *float64 { return FloatPtr(v) }

func RoundScore(v float64) float64 { return roundScore(v) }
