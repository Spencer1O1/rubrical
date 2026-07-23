package llm

import "testing"

func TestValidateJSON_requiredFields(t *testing.T) {
	schema := map[string]any{
		"required": []any{"overallSummary", "criteria"},
	}
	if err := ValidateJSON([]byte(`{"overallSummary":"ok","criteria":[]}`), schema); err != nil {
		t.Fatal(err)
	}
	if err := ValidateJSON([]byte(`{"overallSummary":"ok"}`), schema); err == nil {
		t.Fatal("expected missing criteria error")
	}
}
