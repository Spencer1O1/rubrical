package schema

import (
	"fmt"
	"strings"
)

func ValidateProviderResponse(out *ProviderResponse) error {
	if out == nil {
		return fmt.Errorf("provider response is nil")
	}

	out.OverallSummary = strings.TrimSpace(out.OverallSummary)
	if out.OverallSummary == "" {
		return fmt.Errorf("overallSummary is required")
	}

	out.Confidence = normalizeConfidence(out.Confidence)
	if out.Confidence == "" {
		return fmt.Errorf("confidence is required")
	}

	if len(out.Criteria) == 0 {
		return fmt.Errorf("at least one criterion feedback entry is required")
	}

	for i := range out.Criteria {
		if err := validateCriterionAssessment(i, &out.Criteria[i]); err != nil {
			return err
		}
	}

	out.Strengths = trimStrings(out.Strengths)
	out.Guidance = trimStrings(out.Guidance)

	return nil
}

func ValidateScoredAnalysis(out *ScoredAnalysis) error {
	if out == nil {
		return fmt.Errorf("scored analysis is nil")
	}

	out.OverallSummary = strings.TrimSpace(out.OverallSummary)
	if out.OverallSummary == "" {
		return fmt.Errorf("overallSummary is required")
	}

	out.Confidence = normalizeConfidence(out.Confidence)
	if out.Confidence == "" {
		return fmt.Errorf("confidence is required")
	}

	if len(out.Criteria) == 0 {
		return fmt.Errorf("at least one criterion feedback entry is required")
	}

	if out.PredictedScore != nil && out.PredictedScoreMax != nil && *out.PredictedScore > *out.PredictedScoreMax+0.001 {
		return fmt.Errorf("predictedScore exceeds predictedScoreMax")
	}

	for i := range out.Criteria {
		c := &out.Criteria[i]
		c.CriterionName = strings.TrimSpace(c.CriterionName)
		if c.CriterionName == "" {
			return fmt.Errorf("criteria[%d].criterionName is required", i)
		}
		c.Status = normalizeCriterionStatus(c.Status)
		if c.Status == "" {
			return fmt.Errorf("criteria[%d].status is required", i)
		}
		c.HowToEarnPoints = strings.TrimSpace(c.HowToEarnPoints)
		c.ScoreRationale = strings.TrimSpace(c.ScoreRationale)
		if c.ScoreRationale == "" {
			return fmt.Errorf("criteria[%d].scoreRationale is required", i)
		}

		if c.Status == "not_analyzable" {
			if c.HowToEarnPoints == "" {
				return fmt.Errorf("criteria[%d].howToEarnPoints is required when not analyzable", i)
			}
			c.CriterionScore = 0
			c.PredictedPoints = nil
			c.SelectedRating = ""
			c.FulfilledRequirements = nil
			c.UnfulfilledRequirements = nil
			continue
		}

		if c.CriterionScore < 0 || c.CriterionScore > 1 {
			return fmt.Errorf("criteria[%d].criterionScore must be between 0 and 1", i)
		}
		if c.PredictedPoints != nil && c.MaxPoints != nil && *c.PredictedPoints > *c.MaxPoints+0.001 {
			return fmt.Errorf("criteria[%d].predictedPoints exceeds maxPoints", i)
		}
		if c.PredictedPoints != nil && *c.PredictedPoints < 0 {
			return fmt.Errorf("criteria[%d].predictedPoints must be >= 0", i)
		}
		if err := validateFulfilledRequirements(i, &c.FulfilledRequirements); err != nil {
			return err
		}
		if err := validateUnfulfilledRequirements(i, &c.UnfulfilledRequirements); err != nil {
			return err
		}
		dropFulfilledOverlappingUnfulfilled(&c.FulfilledRequirements, c.UnfulfilledRequirements)
	}

	out.Strengths = trimStrings(out.Strengths)
	out.Guidance = trimStrings(out.Guidance)

	return nil
}

func validateCriterionAssessment(index int, c *CriterionAssessment) error {
	c.CriterionID = strings.TrimSpace(c.CriterionID)
	if c.CriterionID == "" {
		return fmt.Errorf("criteria[%d].criterionId is required", index)
	}
	c.SelectedRatingID = strings.TrimSpace(c.SelectedRatingID)
	if c.BandPosition < 0 || c.BandPosition > 100 {
		return fmt.Errorf("criteria[%d].bandPosition must be between 0 and 100", index)
	}
	c.ScoreRationale = strings.TrimSpace(c.ScoreRationale)
	if c.ScoreRationale == "" {
		return fmt.Errorf("criteria[%d].scoreRationale is required", index)
	}
	if c.FulfilledRequirements == nil {
		return fmt.Errorf("criteria[%d].fulfilledRequirements is required", index)
	}
	if c.UnfulfilledRequirements == nil {
		return fmt.Errorf("criteria[%d].unfulfilledRequirements is required", index)
	}
	if err := validateFulfilledRequirements(index, &c.FulfilledRequirements); err != nil {
		return err
	}
	if err := validateUnfulfilledRequirements(index, &c.UnfulfilledRequirements); err != nil {
		return err
	}
	dropFulfilledOverlappingUnfulfilled(&c.FulfilledRequirements, c.UnfulfilledRequirements)
	return nil
}

func validateFulfilledRequirements(index int, items *[]FulfilledRequirement) error {
	for j := range *items {
		item := &(*items)[j]
		item.Requirement = strings.TrimSpace(item.Requirement)
		item.Evidence = strings.TrimSpace(item.Evidence)
		if item.Requirement == "" {
			return fmt.Errorf("criteria[%d].fulfilledRequirements[%d].requirement is required", index, j)
		}
		if item.Evidence == "" {
			return fmt.Errorf("criteria[%d].fulfilledRequirements[%d].evidence is required", index, j)
		}
	}
	return nil
}

func validateUnfulfilledRequirements(index int, items *[]UnfulfilledRequirement) error {
	for j := range *items {
		item := &(*items)[j]
		item.Requirement = strings.TrimSpace(item.Requirement)
		item.Explanation = strings.TrimSpace(item.Explanation)
		item.Suggestion = strings.TrimSpace(item.Suggestion)
		if item.Requirement == "" {
			return fmt.Errorf("criteria[%d].unfulfilledRequirements[%d].requirement is required", index, j)
		}
		item.Severity = normalizeGapSeverity(item.Severity)
		if item.Severity == "" {
			return fmt.Errorf("criteria[%d].unfulfilledRequirements[%d].severity must be low, medium, or high", index, j)
		}
		if item.Explanation == "" {
			return fmt.Errorf("criteria[%d].unfulfilledRequirements[%d].explanation is required", index, j)
		}
		if item.Suggestion == "" {
			return fmt.Errorf("criteria[%d].unfulfilledRequirements[%d].suggestion is required", index, j)
		}
		if isVacuousSuggestion(item.Suggestion) {
			return fmt.Errorf("criteria[%d].unfulfilledRequirements[%d].suggestion must be a concrete improvement, not a dismissal", index, j)
		}
	}
	return nil
}

// dropFulfilledOverlappingUnfulfilled removes fulfilled rows whose requirement text
// also appears under unfulfilled — a requirement cannot be both met and a gap.
func dropFulfilledOverlappingUnfulfilled(fulfilled *[]FulfilledRequirement, unfulfilled []UnfulfilledRequirement) {
	if fulfilled == nil || len(*fulfilled) == 0 || len(unfulfilled) == 0 {
		return
	}
	gaps := make(map[string]struct{}, len(unfulfilled))
	for _, item := range unfulfilled {
		gaps[normalizeRequirementKey(item.Requirement)] = struct{}{}
	}
	kept := (*fulfilled)[:0]
	for _, item := range *fulfilled {
		if _, hit := gaps[normalizeRequirementKey(item.Requirement)]; hit {
			continue
		}
		kept = append(kept, item)
	}
	*fulfilled = kept
}

func normalizeRequirementKey(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

func SeverityForStatus(status string) string {
	switch normalizeCriterionStatus(status) {
	case "met":
		return "info"
	case "partially_met":
		return "warning"
	case "not_met":
		return "critical"
	case "not_analyzable":
		return "info"
	default:
		return "info"
	}
}

func normalizeConfidence(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low", "medium", "high":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func normalizeCriterionStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "met", "partially_met", "not_met", "not_analyzable":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func normalizeGapSeverity(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "low", "medium", "high":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func trimStrings(values []string) []string {
	var out []string
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func isVacuousSuggestion(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "none", "none.", "none needed", "none needed.",
		"no change needed", "no change needed.",
		"no major change needed", "no major change needed.",
		"n/a", "n/a.":
		return true
	default:
		return false
	}
}
