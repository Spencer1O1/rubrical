package auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestWithEmbedQuery(t *testing.T) {
	if got := WithEmbedQuery("/assignments/1"); got != "/assignments/1?embed=1" {
		t.Fatalf("got %q", got)
	}
	if got := WithEmbedQuery("/assignments/1?embed=1"); got != "/assignments/1?embed=1" {
		t.Fatalf("idempotent got %q", got)
	}
	if got := WithEmbedQuery("/settings?x=1"); got != "/settings?x=1&embed=1" {
		t.Fatalf("got %q", got)
	}
}

func TestRequestIsEmbed(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/login?embed=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !RequestIsEmbed(req) {
		t.Fatal("expected embed query")
	}
	req, err = http.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{"next": {"/a?embed=1"}}.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	if !RequestIsEmbed(req) {
		t.Fatal("expected embed next")
	}
}

func TestEmbedHandoffTokenRoundTrip(t *testing.T) {
	secret := "test-secret-key-for-embed-handoff"
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	token, err := MintEmbedHandoffToken(secret, 42, now)
	if err != nil {
		t.Fatal(err)
	}
	userID, err := ParseEmbedHandoffToken(secret, token, now.Add(30*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if userID != 42 {
		t.Fatalf("userID=%d", userID)
	}
	if _, err := ParseEmbedHandoffToken(secret, token, now.Add(EmbedHandoffTTL+time.Second)); err == nil {
		t.Fatal("expected expiry")
	}
	if _, err := ParseEmbedHandoffToken("wrong", token, now); err == nil {
		t.Fatal("expected bad secret to fail")
	}
}
