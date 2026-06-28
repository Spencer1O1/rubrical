package schema_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"rubrical/internal/analysis/schema"
)

func TestJSONSchemaForOpenAI_acceptedByAPI(t *testing.T) {
	if os.Getenv("OPENAI_KEY") == "" {
		t.Skip("OPENAI_KEY not set")
	}
	s := schema.JSONSchemaForOpenAI([]schema.CriterionSpec{
		{Name: "Event Details", RatingTitles: []string{"Limited", "Proficient", "Exemplary"}},
		{Name: "General Overview", RatingTitles: []string{"Needs Improvement", "Good", "Excellent"}},
		{Name: "Reflection", RatingTitles: []string{""}},
	})
	payload := map[string]any{
		"model":        "gpt-4o-mini",
		"instructions": "Return valid json",
		"input":        "test",
		"store":        false,
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "rubric_analysis",
				"strict": true,
				"schema": s,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/responses", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_KEY"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		t.Fatalf("HTTP %d: %s", resp.StatusCode, string(b))
	}
}
