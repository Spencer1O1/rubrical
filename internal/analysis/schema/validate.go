package schema

import (
	"fmt"
	"strings"
)

func Validate(out *ModelOutput) error {
	if out == nil {
		return fmt.Errorf("analysis output is nil")
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
		c := &out.Criteria[i]
		c.CriterionName = strings.TrimSpace(c.CriterionName)
		if c.CriterionName == "" {
			return fmt.Errorf("criteria[%d].criterionName is required", i)
		}
		c.Status = normalizeCriterionStatus(c.Status)
		if c.Status == "" {
			return fmt.Errorf("criteria[%d].status is required", i)
		}
		c.Evidence = strings.TrimSpace(c.Evidence)
		c.Suggestion = strings.TrimSpace(c.Suggestion)
	}

	out.MissingRequirements = trimStrings(out.MissingRequirements)
	out.Strengths = trimStrings(out.Strengths)
	out.RevisionSuggestions = trimStrings(out.RevisionSuggestions)

	return nil
}

func SeverityForStatus(status string) string {
	switch normalizeCriterionStatus(status) {
	case "met":
		return "info"
	case "partially_met":
		return "warning"
	case "not_met":
		return "critical"
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
	case "met", "partially_met", "not_met":
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
