package drafturl

import (
	"errors"
	"testing"
)

func TestParseSubmissionURLAddsHTTPS(t *testing.T) {
	got, err := ParseSubmissionURL("spencerls.dev")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://spencerls.dev" {
		t.Fatalf("got %q", got)
	}
}

func TestParseSubmissionURLPreservesExplicitScheme(t *testing.T) {
	got, err := ParseSubmissionURL("http://example.com/path")
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://example.com/path" {
		t.Fatalf("got %q", got)
	}
}

func TestParseSubmissionURLEmpty(t *testing.T) {
	_, err := ParseSubmissionURL("   ")
	if !errors.Is(err, ErrEmpty) {
		t.Fatalf("expected empty error, got %v", err)
	}
}

func TestParseSubmissionURLInvalid(t *testing.T) {
	cases := []string{
		"not a url",
		"LIJSEF**#&&#",
		"bad*host.com",
	}
	for _, raw := range cases {
		if _, err := ParseSubmissionURL(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestParseSubmissionURLWithPath(t *testing.T) {
	got, err := ParseSubmissionURL("example.com/paper")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/paper" {
		t.Fatalf("got %q", got)
	}
}
