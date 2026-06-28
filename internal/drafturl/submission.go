package drafturl

import (
	"context"
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

	if !validDisplayHost(parsed.Host) {
		return "", errors.New(invalidSubmissionURLMessage)
	}

	normalized := parsed.String()
	return strings.TrimSuffix(normalized, "/"), nil
}

// ValidateFetchURL parses rawURL and ensures resolved addresses are safe for server-side fetch.
func ValidateFetchURL(ctx context.Context, rawURL string, allowLocal bool) (string, error) {
	normalized, err := ParseSubmissionURL(rawURL)
	if err != nil {
		return "", err
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", errors.New(invalidSubmissionURLMessage)
	}

	if err := ValidateFetchHost(ctx, parsed.Hostname(), allowLocal); err != nil {
		return "", err
	}

	return normalized, nil
}

// ValidateFetchHost resolves hostname and rejects private/link-local addresses unless allowLocal is true.
func ValidateFetchHost(ctx context.Context, hostname string, allowLocal bool) error {
	return checkFetchHost(ctx, hostname, allowLocal)
}

func validDisplayHost(host string) bool {
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	hostname = strings.ToLower(strings.TrimSuffix(hostname, "."))
	if net.ParseIP(hostname) != nil {
		return false
	}
	if hostname == "localhost" {
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

func checkFetchHost(ctx context.Context, hostname string, allowLocal bool) error {
	hostname = strings.ToLower(strings.TrimSuffix(hostname, "."))
	if hostname == "" {
		return errors.New(invalidSubmissionURLMessage)
	}

	if ip := net.ParseIP(hostname); ip != nil {
		return checkFetchIP(ip, allowLocal)
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", hostname)
	if err != nil {
		return errors.New("submission url host could not be resolved")
	}
	if len(ips) == 0 {
		return errors.New("submission url host could not be resolved")
	}

	for _, ip := range ips {
		if err := checkFetchIP(ip, allowLocal); err != nil {
			return err
		}
	}

	return nil
}

func checkFetchIP(ip net.IP, allowLocal bool) error {
	if allowLocal {
		return nil
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return errors.New("submission url host is not allowed")
	}
	return nil
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
