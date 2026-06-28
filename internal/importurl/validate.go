package importurl

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateSourceURL ensures the import origin is a normalized Canvas URL.
func ValidateSourceURL(raw string) (string, error) {
	normalized := NormalizeSourceURL(raw)
	if normalized == "" {
		return "", fmt.Errorf("sourceUrl is required")
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", fmt.Errorf("sourceUrl is invalid")
	}

	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("sourceUrl must use http or https")
	}

	host := strings.ToLower(parsed.Hostname())
	if !strings.HasSuffix(host, "instructure.com") {
		return "", fmt.Errorf("sourceUrl must be a Canvas instructure.com URL")
	}

	path := strings.ToLower(parsed.Path)
	if !strings.Contains(path, "/assignments/") && !strings.Contains(path, "/discussion_topics/") {
		return "", fmt.Errorf("sourceUrl must be a Canvas assignment or discussion URL")
	}

	return normalized, nil
}
