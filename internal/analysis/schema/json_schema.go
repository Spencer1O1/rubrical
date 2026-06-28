package schema

// JSONSchema returns the strict JSON Schema for ProviderResponse (AI structured output).
// Keep aligned with ProviderResponse and ValidateProviderResponse.
func JSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"overallSummary":      map[string]any{"type": "string"},
			"confidence":          map[string]any{"type": "string", "enum": []any{"low", "medium", "high"}},
			"criteria":            map[string]any{"type": "array", "items": criterionJSONSchema()},
			"missingRequirements": stringArraySchema(),
			"strengths":           stringArraySchema(),
			"revisionSuggestions": stringArraySchema(),
		},
		"required": []any{
			"overallSummary",
			"confidence",
			"criteria",
			"missingRequirements",
			"strengths",
			"revisionSuggestions",
		},
	}
}

func criterionJSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"criterionName":  map[string]any{"type": "string"},
			"criterionScore": map[string]any{"type": "number", "minimum": 0, "maximum": 1},
			"evidence":       map[string]any{"type": "string"},
			"suggestion":     map[string]any{"type": "string"},
		},
		"required": []any{
			"criterionName",
			"criterionScore",
			"evidence",
			"suggestion",
		},
	}
}

func stringArraySchema() map[string]any {
	return map[string]any{
		"type":  "array",
		"items": map[string]any{"type": "string"},
	}
}
