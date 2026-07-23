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
	for _, key := range []string{"criterionId", "selectedRatingId", "bandPosition", "scoreRationale", "fulfilledRequirements", "unfulfilledRequirements"} {
		if _, ok := itemProps[key]; !ok {
			t.Fatalf("expected %s on criterion", key)
		}
	}
	if _, ok := itemProps["status"]; ok {
		t.Fatal("status should not be in model schema")
	}
}

func TestJSONSchemaForOpenAI_fixesCriteriaCountIdsAndRatings(t *testing.T) {
	criteria := []CriterionSpec{
		{ID: "event-details", Name: "Event Details", RatingIDs: []string{"r0", "r1"}},
		{ID: "general-overview", Name: "General Overview", RatingIDs: []string{"r0", "r1", "r2"}},
		{ID: "reflection", Name: "Reflection", RatingIDs: []string{""}},
	}
	schema := JSONSchemaForOpenAI(criteria)
	props := schema["properties"].(map[string]any)
	criteriaSchema := props["criteria"].(map[string]any)

	if criteriaSchema["minItems"] != 3 || criteriaSchema["maxItems"] != 3 {
		t.Fatalf("criteria bounds = %v / %v", criteriaSchema["minItems"], criteriaSchema["maxItems"])
	}
	items := criteriaSchema["items"].(map[string]any)
	anyOf, ok := items["anyOf"].([]any)
	if !ok || len(anyOf) != 3 {
		t.Fatalf("items.anyOf = %T len=%d", items["anyOf"], len(anyOf))
	}
	for i, spec := range criteria {
		item := anyOf[i].(map[string]any)
		itemProps := item["properties"].(map[string]any)
		criterionID := itemProps["criterionId"].(map[string]any)
		if criterionID["const"] != spec.ID {
			t.Fatalf("anyOf[%d].criterionId.const = %v, want %q", i, criterionID["const"], spec.ID)
		}
		selectedRating := itemProps["selectedRatingId"].(map[string]any)
		enum, ok := selectedRating["enum"].([]any)
		if !ok {
			t.Fatalf("anyOf[%d].selectedRatingId.enum missing", i)
		}
		if len(enum) != len(spec.RatingIDs) {
			t.Fatalf("anyOf[%d].selectedRatingId.enum len = %d, want %d", i, len(enum), len(spec.RatingIDs))
		}
		for j, id := range spec.RatingIDs {
			if enum[j] != id {
				t.Fatalf("anyOf[%d].selectedRatingId.enum[%d] = %v, want %q", i, j, enum[j], id)
			}
		}
	}
}

func TestJSONSchemaForAnthropic_omitsCriteriaCountBounds(t *testing.T) {
	criteria := []CriterionSpec{
		{ID: "event-details", RatingIDs: []string{"r0", "r1"}},
		{ID: "general-overview", RatingIDs: []string{"r0", "r1"}},
		{ID: "reflection", RatingIDs: []string{""}},
	}
	schema := JSONSchemaForAnthropic(criteria)
	props := schema["properties"].(map[string]any)
	criteriaSchema := props["criteria"].(map[string]any)

	if _, ok := criteriaSchema["minItems"]; ok {
		t.Fatal("Anthropic schema must not set minItems")
	}
	if _, ok := criteriaSchema["maxItems"]; ok {
		t.Fatal("Anthropic schema must not set maxItems")
	}
}

func TestSelectedRatingSchema_emptyBands(t *testing.T) {
	got := selectedRatingSchema([]string{""})
	enum, ok := got["enum"].([]any)
	if !ok || len(enum) != 1 || enum[0] != "" {
		t.Fatalf("selectedRatingSchema([\"\"]) = %v", got)
	}
}
