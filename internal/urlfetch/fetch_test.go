package urlfetch

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func localhostFetchURL(serverURL string) string {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("http://localhost:%s", parsed.Port())
}

func TestSafeFetcherFetchHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><head><style>.x{}</style><script>alert(1)</script></head><body><p>Hello <b>world</b></p></body></html>`))
	}))
	defer server.Close()

	fetcher := NewSafeFetcher(true)
	got, err := fetcher.Fetch(context.Background(), localhostFetchURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	if got != "Hello world" {
		t.Fatalf("got %q", got)
	}
}

func TestSafeFetcherRejectsRedirectToPrivateHost(t *testing.T) {
	step := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if step == 0 {
			step++
			http.Redirect(w, r, "http://127.0.0.1/private", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte("<p>secret</p>"))
	}))
	defer server.Close()

	fetcher := NewSafeFetcher(true)
	if _, err := fetcher.Fetch(context.Background(), localhostFetchURL(server.URL)); err == nil {
		t.Fatal("expected redirect to private host to fail")
	}
}

func TestSafeFetcherRejectsEmptyText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<style>.x{}</style><script></script>"))
	}))
	defer server.Close()

	fetcher := NewSafeFetcher(true)
	if _, err := fetcher.Fetch(context.Background(), localhostFetchURL(server.URL)); err == nil {
		t.Fatal("expected empty extracted text to fail")
	} else if !strings.Contains(err.Error(), "no text content") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTMLToTextSkipsScriptAndStyle(t *testing.T) {
	got := htmlToText(`<div>Keep<script>drop</script><style>.a{}</style>me</div>`)
	if got != "Keep me" {
		t.Fatalf("got %q", got)
	}
}
