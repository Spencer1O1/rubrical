package schema

// JSONSchema returns the strict JSON Schema for ProviderResponse (AI structured output).
// Keep aligned with ProviderResponse and ValidateProviderResponse().
func JSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"overallSummary":      map[string]any{"type": "string"},
			"confidence":          map[string]any{"type": "string", "enum": []any{"low", "medium", "high"}},
			"criteria":            map[string]any{"type": "array", "items": criterionJSONSchema()},
			"strengths":           stringArraySchema(),
			"guidance":            stringArraySchema(),
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

func criterionJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"criterionName":           map[string]any{"type": "string"},
			"selectedRating":          map[string]any{"type": "string"},
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
