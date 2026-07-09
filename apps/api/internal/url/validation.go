package url

import (
	"errors"
	neturl "net/url"
	"strings"
	"unicode"
)

const MaxLongURLLength = 2048

var (
	ErrLongURLTooLong           = errors.New("long URL is too long")
	ErrLongURLInvalid           = errors.New("long URL is invalid")
	ErrLongURLSchemeUnsupported = errors.New("long URL scheme must be http or https")
	ErrLongURLHostRequired      = errors.New("long URL host is required")
)

func NormalizeLongURL(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", ErrLongURLRequired
	}

	if len(normalized) > MaxLongURLLength {
		return "", ErrLongURLTooLong
	}

	if strings.ContainsFunc(normalized, unicode.IsSpace) {
		return "", ErrLongURLInvalid
	}

	parsed, err := neturl.Parse(normalized)
	if err != nil {
		return "", ErrLongURLInvalid
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return "", ErrLongURLSchemeUnsupported
	}

	if parsed.Hostname() == "" {
		return "", ErrLongURLHostRequired
	}

	return normalized, nil
}

func ValidateLongURL(value string) error {
	_, err := NormalizeLongURL(value)
	return err
}
