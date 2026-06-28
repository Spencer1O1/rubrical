package schema

// JSONSchema returns the strict JSON Schema used by provider structured-output APIs.
// Keep aligned with ModelOutput and Validate.
func JSONSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"overallSummary":      map[string]any{"type": "string"},
			"estimatedScore":      nullableNumber(),
			"estimatedScoreMax":   nullableNumber(),
			"confidence":          map[string]any{"type": "string", "enum": []any{"low", "medium", "high"}},
			"criteria":            map[string]any{"type": "array", "items": criterionJSONSchema()},
			"missingRequirements": stringArraySchema(),
			"strengths":           stringArraySchema(),
			"revisionSuggestions": stringArraySchema(),
		},
		"required": []any{
			"overallSummary",
			"estimatedScore",
			"estimatedScoreMax",
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
			"criterionName":   map[string]any{"type": "string"},
			"status":          map[string]any{"type": "string", "enum": []any{"met", "partially_met", "not_met"}},
			"estimatedPoints": nullableNumber(),
			"maxPoints":       nullableNumber(),
			"evidence":        map[string]any{"type": "string"},
			"suggestion":      map[string]any{"type": "string"},
		},
		"required": []any{
			"criterionName",
			"status",
			"estimatedPoints",
			"maxPoints",
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

func nullableNumber() map[string]any {
	return map[string]any{"type": []any{"number", "null"}}
}
