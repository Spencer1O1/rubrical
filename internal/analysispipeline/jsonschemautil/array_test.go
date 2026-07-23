package jsonschemautil_test

import (
	"testing"

	"rubrical/internal/analysispipeline/jsonschemautil"
)

func TestEnforceExactItemCount(t *testing.T) {
	if !jsonschemautil.EnforceExactItemCount("openai") {
		t.Fatal("openai should enforce")
	}
	if jsonschemautil.EnforceExactItemCount("anthropic") {
		t.Fatal("anthropic should not enforce")
	}
}

func TestFixedAnyOfArray(t *testing.T) {
	got := jsonschemautil.FixedAnyOfArray([]any{"a", "b"}, true)
	if got["minItems"] != 2 || got["maxItems"] != 2 {
		t.Fatalf("count missing: %#v", got)
	}
	noCount := jsonschemautil.FixedAnyOfArray([]any{"a"}, false)
	if _, ok := noCount["minItems"]; ok {
		t.Fatalf("unexpected minItems: %#v", noCount)
	}
}
