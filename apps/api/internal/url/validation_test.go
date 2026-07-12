package url

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNormalizeLongURLAcceptsValidHTTPURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "http URL",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "https URL with path and query",
			input:    " https://example.com/articles/123?utm_source=test#section ",
			expected: "https://example.com/articles/123?utm_source=test#section",
		},
		{
			name:     "uppercase scheme",
			input:    "HTTPS://example.com",
			expected: "HTTPS://example.com",
		},
		{
			name:     "localhost with port",
			input:    "http://localhost:3000/path",
			expected: "http://localhost:3000/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := NormalizeLongURL(tt.input)
			if err != nil {
				t.Fatalf("expected URL to be valid: %v", err)
			}

			if actual != tt.expected {
				t.Fatalf("expected normalized URL %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestNormalizeExpiresAt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC)
	expiresAt := now.Add(time.Hour).In(time.FixedZone("IST", 5*60*60+30*60))

	normalized, err := NormalizeExpiresAt(&expiresAt, now)
	if err != nil {
		t.Fatalf("expected expiration to be valid: %v", err)
	}

	if normalized == nil || !normalized.Equal(now.Add(time.Hour)) || normalized.Location() != time.UTC {
		t.Fatalf("expected UTC expiration %s, got %v", now.Add(time.Hour), normalized)
	}
}

func TestNormalizeExpiresAtAllowsNoExpiration(t *testing.T) {
	t.Parallel()

	normalized, err := NormalizeExpiresAt(nil, time.Now())
	if err != nil {
		t.Fatalf("expected no expiration to be valid: %v", err)
	}

	if normalized != nil {
		t.Fatalf("expected no expiration, got %v", normalized)
	}
}

func TestNormalizeLongURLRejectsInvalidURLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "empty",
			input:    " ",
			expected: ErrLongURLRequired,
		},
		{
			name:     "too long",
			input:    "https://example.com/" + strings.Repeat("a", MaxLongURLLength),
			expected: ErrLongURLTooLong,
		},
		{
			name:     "relative URL",
			input:    "/articles/123",
			expected: ErrLongURLSchemeUnsupported,
		},
		{
			name:     "missing scheme",
			input:    "example.com/articles/123",
			expected: ErrLongURLSchemeUnsupported,
		},
		{
			name:     "unsupported scheme",
			input:    "ftp://example.com/file",
			expected: ErrLongURLSchemeUnsupported,
		},
		{
			name:     "dangerous scheme",
			input:    "javascript:alert(1)",
			expected: ErrLongURLSchemeUnsupported,
		},
		{
			name:     "missing host",
			input:    "https:///articles/123",
			expected: ErrLongURLHostRequired,
		},
		{
			name:     "contains spaces",
			input:    "https://example.com/a path",
			expected: ErrLongURLInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NormalizeLongURL(tt.input)
			if !errors.Is(err, tt.expected) {
				t.Fatalf("expected error %q, got %v", tt.expected, err)
			}
		})
	}
}
