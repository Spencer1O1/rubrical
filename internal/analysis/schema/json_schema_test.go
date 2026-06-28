package schema

import "testing"

func TestJSONSchema_hasRequiredTopLevelFields(t *testing.T) {
	schema := JSONSchema()
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("expected required array")
	}
	if len(required) != 6 {
		t.Fatalf("required fields = %d", len(required))
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties map")
	}
	if _, ok := props["criteria"]; !ok {
		t.Fatal("expected criteria property")
	}
	if _, ok := props["predictedScore"]; ok {
		t.Fatal("predictedScore should not be in model schema")
	}
	criteria, ok := props["criteria"].(map[string]any)
	if !ok {
		t.Fatal("expected criteria schema")
	}
	items, ok := criteria["items"].(map[string]any)
	if !ok {
		t.Fatal("expected criteria items")
	}
	itemProps, ok := items["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected criterion properties")
	}
	if _, ok := itemProps["criterionScore"]; !ok {
		t.Fatal("expected criterionScore on criterion")
	}
	if _, ok := itemProps["status"]; ok {
		t.Fatal("status should not be in model schema")
	}
}
