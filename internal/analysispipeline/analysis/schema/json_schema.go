package schema

import "rubrical/internal/analysispipeline/jsonschemautil"

// JSONSchema returns the provider response schema with no rubric-specific bounds.
// Prefer JSONSchemaForOpenAI or JSONSchemaForAnthropic when rubric criteria are available.
func JSONSchema() map[string]any {
	return JSONSchemaForOpenAI(nil)
}

// JSONSchemaForOpenAI builds strict JSON Schema for OpenAI structured outputs.
// When criteria is non-empty, criteria[] must have exactly that many entries;
// each entry's criterionId and selectedRatingId must match that rubric row.
func JSONSchemaForOpenAI(criteria []CriterionSpec) map[string]any {
	return providerResponseSchema(criteria, true)
}

// JSONSchemaForAnthropic builds JSON Schema for Anthropic structured outputs.
// Anthropic only allows minItems of 0 or 1, so exact criteria count is not schema-enforced.
func JSONSchemaForAnthropic(criteria []CriterionSpec) map[string]any {
	return providerResponseSchema(criteria, false)
}

func providerResponseSchema(criteria []CriterionSpec, enforceCriteriaCount bool) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"overallSummary": map[string]any{"type": "string"},
			"confidence":     map[string]any{"type": "string", "enum": []any{"low", "medium", "high"}},
			"criteria":       criteriaArraySchema(criteria, enforceCriteriaCount),
			"strengths":      stringArraySchema(),
			"guidance":       stringArraySchema(),
		},
		"required": []any{
			"overallSummary",
			"confidence",
			"criteria",
			"strengths",
			"guidance",
		},
	}
}

func criteriaArraySchema(criteria []CriterionSpec, enforceCount bool) map[string]any {
	n := len(criteria)
	if n == 0 {
		return map[string]any{
			"type":  "array",
			"items": criterionJSONSchema("", nil),
		}
	}

	anyOf := make([]any, n)
	for i, spec := range criteria {
		anyOf[i] = criterionJSONSchema(spec.ID, spec.RatingIDs)
	}
	return jsonschemautil.FixedAnyOfArray(anyOf, enforceCount)
}

func criterionJSONSchema(fixedID string, ratingIDs []string) map[string]any {
	criterionID := map[string]any{"type": "string"}
	if fixedID != "" {
		criterionID["const"] = fixedID
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"criterionId":             criterionID,
			"selectedRatingId":        selectedRatingSchema(ratingIDs),
			"bandPosition":            map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
			"scoreRationale":          map[string]any{"type": "string"},
			"fulfilledRequirements":   fulfilledRequirementsSchema(),
			"unfulfilledRequirements": unfulfilledRequirementsSchema(),
		},
		"required": []any{
			"criterionId",
			"selectedRatingId",
			"bandPosition",
			"scoreRationale",
			"fulfilledRequirements",
			"unfulfilledRequirements",
		},
	}
}

func selectedRatingSchema(ratingIDs []string) map[string]any {
	if len(ratingIDs) == 0 {
		return map[string]any{"type": "string"}
	}
	enum := make([]any, len(ratingIDs))
	for i, id := range ratingIDs {
		enum[i] = id
	}
	return map[string]any{
		"type": "string",
		"enum": enum,
	}
}

func fulfilledRequirementsSchema() map[string]any {
	return map[string]any{
		"type":  "array",
		"items": fulfilledRequirementSchema(),
	}
}

func fulfilledRequirementSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"requirement": map[string]any{"type": "string"},
			"evidence":    map[string]any{"type": "string"},
		},
		"required": []any{"requirement", "evidence"},
	}
}

func unfulfilledRequirementsSchema() map[string]any {
	return map[string]any{
		"type":  "array",
		"items": unfulfilledRequirementSchema(),
	}
}

func unfulfilledRequirementSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"requirement": map[string]any{"type": "string"},
			"severity":    map[string]any{"type": "string", "enum": []any{"low", "medium", "high"}},
			"explanation": map[string]any{"type": "string"},
			"suggestion":  map[string]any{"type": "string"},
		},
		"required": []any{"requirement", "severity", "explanation", "suggestion"},
	}
}

func stringArraySchema() map[string]any {
	return map[string]any{
		"type":  "array",
		"items": map[string]any{"type": "string"},
	}
}
