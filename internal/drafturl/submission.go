package drafturl

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

var ErrEmpty = errors.New("submission url is required")

const invalidSubmissionURLMessage = "Enter a valid website URL (e.g. example.com or https://example.com)"

// ParseSubmissionURL trims input, adds https:// when no scheme is present, and validates http(s).
func ParseSubmissionURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrEmpty
	}

	candidate := trimmed
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}

	parsed, err := url.ParseRequestURI(candidate)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New(invalidSubmissionURLMessage)
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return "", errors.New(invalidSubmissionURLMessage)
	}

	if !validSubmissionHost(parsed.Host) {
		return "", errors.New(invalidSubmissionURLMessage)
	}

	normalized := parsed.String()
	return strings.TrimSuffix(normalized, "/"), nil
}

func validSubmissionHost(host string) bool {
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	hostname = strings.ToLower(strings.TrimSuffix(hostname, "."))
	if hostname == "localhost" {
		return true
	}
	if ip := net.ParseIP(hostname); ip != nil {
		return true
	}
	if !strings.Contains(hostname, ".") {
		return false
	}

	for _, label := range strings.Split(hostname, ".") {
		if !validHostnameLabel(label) {
			return false
		}
	}

	return true
}

func validHostnameLabel(label string) bool {
	if len(label) == 0 || len(label) > 63 {
		return false
	}

	for i, r := range label {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		case r == '-' && i > 0 && i < len(label)-1:
		default:
			return false
		}
	}

	return true
}

func InvalidSubmissionURLMessage() string {
	return invalidSubmissionURLMessage
}
