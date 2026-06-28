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

func TestJSONSchemaForOpenAI_fixesCriteriaCountNamesAndRatings(t *testing.T) {
	criteria := []CriterionSpec{
		{Name: "Event Details", RatingTitles: []string{"Limited", "Proficient"}},
		{Name: "General Overview", RatingTitles: []string{"Needs Improvement", "Good", "Excellent"}},
		{Name: "Reflection", RatingTitles: []string{""}},
	}
	schema := JSONSchemaForOpenAI(criteria)
	props := schema["properties"].(map[string]any)
	criteriaSchema := props["criteria"].(map[string]any)

	if criteriaSchema["minItems"] != 3 || criteriaSchema["maxItems"] != 3 {
		t.Fatalf("criteria bounds = %v / %v", criteriaSchema["minItems"], criteriaSchema["maxItems"])
	}
	if _, ok := criteriaSchema["prefixItems"]; ok {
		t.Fatal("OpenAI schema must not use prefixItems")
	}
	items := criteriaSchema["items"].(map[string]any)
	anyOf, ok := items["anyOf"].([]any)
	if !ok || len(anyOf) != 3 {
		t.Fatalf("items.anyOf = %T len=%d", items["anyOf"], len(anyOf))
	}
	for i, spec := range criteria {
		item := anyOf[i].(map[string]any)
		itemProps := item["properties"].(map[string]any)
		criterionName := itemProps["criterionName"].(map[string]any)
		if criterionName["const"] != spec.Name {
			t.Fatalf("anyOf[%d].criterionName.const = %v, want %q", i, criterionName["const"], spec.Name)
		}
		selectedRating := itemProps["selectedRating"].(map[string]any)
		enum, ok := selectedRating["enum"].([]any)
		if !ok {
			t.Fatalf("anyOf[%d].selectedRating.enum missing", i)
		}
		if len(enum) != len(spec.RatingTitles) {
			t.Fatalf("anyOf[%d].selectedRating.enum len = %d, want %d", i, len(enum), len(spec.RatingTitles))
		}
		for j, title := range spec.RatingTitles {
			if enum[j] != title {
				t.Fatalf("anyOf[%d].selectedRating.enum[%d] = %v, want %q", i, j, enum[j], title)
			}
		}
	}
}

func TestJSONSchemaForAnthropic_omitsCriteriaCountBounds(t *testing.T) {
	criteria := []CriterionSpec{
		{Name: "Event Details", RatingTitles: []string{"Limited", "Proficient"}},
		{Name: "General Overview", RatingTitles: []string{"Good", "Excellent"}},
		{Name: "Reflection", RatingTitles: []string{""}},
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
	items := criteriaSchema["items"].(map[string]any)
	anyOf, ok := items["anyOf"].([]any)
	if !ok || len(anyOf) != 3 {
		t.Fatalf("items.anyOf = %T len=%d", items["anyOf"], len(anyOf))
	}
}

func TestSelectedRatingSchema_emptyBands(t *testing.T) {
	got := selectedRatingSchema([]string{""})
	enum, ok := got["enum"].([]any)
	if !ok || len(enum) != 1 || enum[0] != "" {
		t.Fatalf("selectedRatingSchema([\"\"]) = %v", got)
	}
}
