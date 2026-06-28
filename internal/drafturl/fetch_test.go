package drafturl

import (
	"context"
	"testing"
)

func TestParseSubmissionURLRejectsIPLiteral(t *testing.T) {
	cases := []string{
		"http://127.0.0.1",
		"http://192.168.1.1/page",
		"http://8.8.8.8",
	}
	for _, raw := range cases {
		if _, err := ParseSubmissionURL(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}

func TestValidateFetchURLBlocksLocalhostWithoutAllowLocal(t *testing.T) {
	ctx := context.Background()
	if _, err := ValidateFetchURL(ctx, "http://localhost/page", false); err == nil {
		t.Fatal("expected localhost fetch to be blocked")
	}
}

func TestValidateFetchURLAllowsLocalhostWithAllowLocal(t *testing.T) {
	ctx := context.Background()
	got, err := ValidateFetchURL(ctx, "http://localhost/page", true)
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://localhost/page" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateFetchHostRejectsPrivateIPLiteral(t *testing.T) {
	ctx := context.Background()
	if err := ValidateFetchHost(ctx, "127.0.0.1", false); err == nil {
		t.Fatal("expected private IP to be blocked")
	}
}

func TestValidateFetchHostAllowsPrivateIPLiteralWithAllowLocal(t *testing.T) {
	ctx := context.Background()
	if err := ValidateFetchHost(ctx, "127.0.0.1", true); err != nil {
		t.Fatal(err)
	}
}
