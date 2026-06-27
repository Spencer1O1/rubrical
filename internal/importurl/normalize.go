package importurl

import (
	"net/url"
	"strings"
)

// NormalizeSourceURL strips query/fragment so Canvas revisits dedupe to one snapshot.
func NormalizeSourceURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}

	parsed.RawQuery = ""
	parsed.Fragment = ""
	normalized := parsed.String()
	return strings.TrimSuffix(normalized, "/")
}
