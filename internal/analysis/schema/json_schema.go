package schema

// JSONSchema returns the provider response schema with no rubric-specific bounds.
// Prefer JSONSchemaForOpenAI or JSONSchemaForAnthropic when rubric criteria are available.
func JSONSchema() map[string]any {
	return JSONSchemaForOpenAI(nil)
}

// JSONSchemaForOpenAI builds strict JSON Schema for OpenAI structured outputs.
// When criteria is non-empty, criteria[] must have exactly that many entries;
// each entry's criterionName and selectedRating must match that rubric row.
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
		anyOf[i] = criterionJSONSchema(spec.Name, spec.RatingTitles)
	}

	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"anyOf": anyOf,
		},
	}
	if enforceCount && n > 0 {
		schema["minItems"] = n
		schema["maxItems"] = n
	}
	return schema
}

func criterionJSONSchema(fixedName string, ratingTitles []string) map[string]any {
	criterionName := map[string]any{"type": "string"}
	if fixedName != "" {
		criterionName["const"] = fixedName
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"criterionName":           criterionName,
			"selectedRating":          selectedRatingSchema(ratingTitles),
			"bandPosition":            map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
			"scoreRationale":          map[string]any{"type": "string"},
			"fulfilledRequirements":   fulfilledRequirementsSchema(),
			"unfulfilledRequirements": unfulfilledRequirementsSchema(),
		},
		"required": []any{
			"criterionName",
			"selectedRating",
			"bandPosition",
			"scoreRationale",
			"fulfilledRequirements",
			"unfulfilledRequirements",
		},
	}
}

func selectedRatingSchema(ratingTitles []string) map[string]any {
	if len(ratingTitles) == 0 {
		return map[string]any{"type": "string"}
	}
	enum := make([]any, len(ratingTitles))
	for i, title := range ratingTitles {
		enum[i] = title
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
