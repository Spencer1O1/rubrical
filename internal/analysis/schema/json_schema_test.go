package schema

import "testing"

func TestJSONSchema_hasRequiredTopLevelFields(t *testing.T) {
	schema := JSONSchema()
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("expected required array")
	}
	if len(required) != 8 {
		t.Fatalf("required fields = %d", len(required))
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties map")
	}
	if _, ok := props["criteria"]; !ok {
		t.Fatal("expected criteria property")
	}
}
