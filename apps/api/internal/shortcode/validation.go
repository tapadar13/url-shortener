package shortcode

import (
	"errors"
	"strings"
)

const (
	MinLength = 4
	MaxLength = 32
)

var (
	ErrRequired     = errors.New("short code is required")
	ErrTooShort     = errors.New("short code is too short")
	ErrTooLong      = errors.New("short code is too long")
	ErrInvalidChars = errors.New("short code must contain only letters and numbers")
	ErrReserved     = errors.New("short code is reserved")
)

var reservedCodes = map[string]struct{}{
	"api":     {},
	"healthz": {},
	"readyz":  {},
	"shorten": {},
}

func Normalize(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", ErrRequired
	}

	if len(normalized) < MinLength {
		return "", ErrTooShort
	}

	if len(normalized) > MaxLength {
		return "", ErrTooLong
	}

	if !isBase62(normalized) {
		return "", ErrInvalidChars
	}

	if IsReserved(normalized) {
		return "", ErrReserved
	}

	return normalized, nil
}

func Validate(value string) error {
	_, err := Normalize(value)
	return err
}

func IsReserved(value string) bool {
	_, ok := reservedCodes[strings.ToLower(strings.TrimSpace(value))]
	return ok
}

func isBase62(value string) bool {
	for _, char := range value {
		if char >= '0' && char <= '9' {
			continue
		}

		if char >= 'a' && char <= 'z' {
			continue
		}

		if char >= 'A' && char <= 'Z' {
			continue
		}

		return false
	}

	return true
}
