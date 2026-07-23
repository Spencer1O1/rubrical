package checkability

import (
	"encoding/json"
	"fmt"
	"strings"

	"rubrical/internal/analysispipeline/criterion"
	"rubrical/internal/analysispipeline/jsonschemautil"
)

// JSONSchema builds structured-output schema for pass 1.
func JSONSchema(refs []criterion.Ref, providerName string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"criteria": criteriaArraySchema(refs, jsonschemautil.EnforceExactItemCount(providerName)),
		},
		"required": []any{"criteria"},
	}
}

func criteriaArraySchema(refs []criterion.Ref, enforceCount bool) map[string]any {
	n := len(refs)
	if n == 0 {
		return map[string]any{
			"type":  "array",
			"items": criterionJSONSchema(""),
		}
	}
	anyOf := make([]any, n)
	for i, ref := range refs {
		anyOf[i] = criterionJSONSchema(ref.ID)
	}
	return jsonschemautil.FixedAnyOfArray(anyOf, enforceCount)
}

func criterionJSONSchema(fixedID string) map[string]any {
	criterionID := map[string]any{"type": "string"}
	if fixedID != "" {
		criterionID["const"] = fixedID
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"criterionId":        criterionID,
			"evidenceProvidable": map[string]any{"type": "boolean"},
			"evidenceAnalyzable": map[string]any{"type": "boolean"},
			"reason":             map[string]any{"type": "string"},
			"howToEarnPoints":    map[string]any{"type": "string"},
		},
		"required": []any{"criterionId", "evidenceProvidable", "evidenceAnalyzable", "reason", "howToEarnPoints"},
	}
}

// ValidateResponse checks coverage and howToEarnPoints; resolves CriterionName from ids.
func ValidateResponse(resp *Response, refs []criterion.Ref) error {
	if resp == nil {
		return fmt.Errorf("checkability response is nil")
	}
	if len(refs) == 0 {
		return nil
	}
	if len(resp.Criteria) != len(refs) {
		return fmt.Errorf("expected %d criteria, got %d", len(refs), len(resp.Criteria))
	}

	seen := make(map[string]bool, len(resp.Criteria))
	for i := range resp.Criteria {
		c := &resp.Criteria[i]
		c.CriterionID = strings.TrimSpace(c.CriterionID)
		c.Reason = strings.TrimSpace(c.Reason)
		c.HowToEarnPoints = strings.TrimSpace(c.HowToEarnPoints)
		if c.CriterionID == "" {
			return fmt.Errorf("criteria[%d].criterionId is required", i)
		}
		ref, ok := criterion.Lookup(refs, c.CriterionID)
		if !ok {
			return fmt.Errorf("criteria[%d] criterionId %q not in rubric", i, c.CriterionID)
		}
		c.CriterionName = ref.Name
		if seen[c.CriterionID] {
			return fmt.Errorf("criteria[%d] duplicate criterionId %q", i, c.CriterionID)
		}
		seen[c.CriterionID] = true
		if c.Reason == "" {
			return fmt.Errorf("criteria[%d].reason is required", i)
		}
		if !c.Checkable() && c.HowToEarnPoints == "" {
			return fmt.Errorf("criteria[%d].howToEarnPoints is required when evidence is not checkable", i)
		}
		if c.Checkable() {
			c.HowToEarnPoints = ""
		}
	}

	for _, ref := range refs {
		if !seen[ref.ID] {
			return fmt.Errorf("missing criterionId %q", ref.ID)
		}
	}

	byID := make(map[string]Criterion, len(resp.Criteria))
	for _, c := range resp.Criteria {
		byID[c.CriterionID] = c
	}
	ordered := make([]Criterion, len(refs))
	for i, ref := range refs {
		ordered[i] = byID[ref.ID]
	}
	resp.Criteria = ordered
	return nil
}

// ParseResponse unmarshals and validates the pass-1 JSON.
func ParseResponse(raw []byte, refs []criterion.Ref) (*Response, error) {
	var resp Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode checkability response: %w", err)
	}
	if err := ValidateResponse(&resp, refs); err != nil {
		return nil, err
	}
	return &resp, nil
}
