package llm

import (
	"encoding/json"
	"fmt"
)

// ValidateJSON checks that raw is valid JSON and, when schema is provided,
// that object responses include every required property.
func ValidateJSON(raw []byte, schema map[string]any) error {
	if !json.Valid(raw) {
		return fmt.Errorf("llm returned non-json content")
	}
	if len(schema) == 0 {
		return nil
	}

	required, _ := schema["required"].([]any)
	if len(required) == 0 {
		return nil
	}

	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return fmt.Errorf("llm json is not an object: %w", err)
	}
	for _, item := range required {
		key, ok := item.(string)
		if !ok || key == "" {
			continue
		}
		if _, exists := obj[key]; !exists {
			return fmt.Errorf("llm json missing required field %q", key)
		}
	}
	return nil
}
