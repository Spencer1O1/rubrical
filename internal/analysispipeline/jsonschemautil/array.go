package jsonschemautil

import "strings"

// EnforceExactItemCount is true for providers that accept minItems/maxItems
// equal to the criteria length (OpenAI). Anthropic only allows 0 or 1.
func EnforceExactItemCount(provider string) bool {
	return !strings.EqualFold(strings.TrimSpace(provider), "anthropic")
}

// FixedAnyOfArray builds a JSON Schema array whose items are anyOf variants.
// When enforceCount is true, minItems/maxItems match len(variants).
// Callers must pass a non-empty variants slice (empty arrays keep local item schemas).
func FixedAnyOfArray(variants []any, enforceCount bool) map[string]any {
	n := len(variants)
	schema := map[string]any{
		"type": "array",
		"items": map[string]any{
			"anyOf": variants,
		},
	}
	if enforceCount && n > 0 {
		schema["minItems"] = n
		schema["maxItems"] = n
	}
	return schema
}
