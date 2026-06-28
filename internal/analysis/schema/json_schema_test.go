package schema

import "testing"

func TestJSONSchema_hasRequiredTopLevelFields(t *testing.T) {
	schema := JSONSchema()
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("expected required array")
	}
	if len(required) != 5 {
		t.Fatalf("required fields = %d", len(required))
	}
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties map")
	}
	if _, ok := props["criteria"]; !ok {
		t.Fatal("expected criteria property")
	}
	if _, ok := props["guidance"]; !ok {
		t.Fatal("expected guidance property")
	}
	if _, ok := props["missingRequirements"]; ok {
		t.Fatal("missingRequirements should not be in model schema")
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
	for _, key := range []string{"selectedRating", "bandPosition", "scoreRationale", "fulfilledRequirements", "unfulfilledRequirements"} {
		if _, ok := itemProps[key]; !ok {
			t.Fatalf("expected %s on criterion", key)
		}
	}
	if _, ok := itemProps["status"]; ok {
		t.Fatal("status should not be in model schema")
	}
}
