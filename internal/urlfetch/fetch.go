package urlfetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"rubrical/internal/config"
	"rubrical/internal/drafturl"
)

var ErrNonHTMLContent = errors.New("url fetch returned non-html content")

const maxResponseBytes = 512 << 10

type SafeFetcher struct {
	client     *http.Client
	allowLocal bool
}

func NewSafeFetcher(allowLocal bool) *SafeFetcher {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			if err := drafturl.ValidateFetchHost(ctx, host, allowLocal); err != nil {
				return nil, err
			}
			dialer := &net.Dialer{Timeout: config.DefaultURLFetchTimeout}
			return dialer.DialContext(ctx, network, addr)
		},
	}

	return &SafeFetcher{
		allowLocal: allowLocal,
		client: &http.Client{
			Timeout:   config.DefaultURLFetchTimeout,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				_, err := drafturl.ValidateFetchURL(req.Context(), req.URL.String(), allowLocal)
				return err
			},
		},
	}
}

func (f *SafeFetcher) Fetch(ctx context.Context, rawURL string) (string, error) {
	if f == nil || f.client == nil {
		return "", fmt.Errorf("url fetcher unavailable")
	}

	normalized, err := drafturl.ValidateFetchURL(ctx, rawURL, f.allowLocal)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalized, nil)
	if err != nil {
		return "", err
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("url fetch status %d", resp.StatusCode)
	}

	if !isHTMLContentType(resp.Header.Get("Content-Type")) {
		return "", ErrNonHTMLContent
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", err
	}

	text := htmlToText(string(body))
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("url fetch returned no text content")
	}

	return text, nil
}

func htmlToText(rawHTML string) string {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return collapseWhitespace(rawHTML)
	}

	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteByte(' ')
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	return collapseWhitespace(b.String())
}

func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func isHTMLContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if contentType == "" {
		return true
	}
	return contentType == "text/html" || contentType == "application/xhtml+xml"
}
