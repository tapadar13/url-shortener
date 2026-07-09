package shortcode

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeAcceptsBase62Codes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "letters and numbers",
			input:    " AbC123 ",
			expected: "AbC123",
		},
		{
			name:     "minimum length",
			input:    "aB12",
			expected: "aB12",
		},
		{
			name:     "maximum length",
			input:    strings.Repeat("A", MaxLength),
			expected: strings.Repeat("A", MaxLength),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := Normalize(tt.input)
			if err != nil {
				t.Fatalf("expected short code to be valid: %v", err)
			}

			if actual != tt.expected {
				t.Fatalf("expected normalized code %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestNormalizeRejectsInvalidCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{
			name:     "empty",
			input:    " ",
			expected: ErrRequired,
		},
		{
			name:     "too short",
			input:    "abc",
			expected: ErrTooShort,
		},
		{
			name:     "too long",
			input:    strings.Repeat("A", MaxLength+1),
			expected: ErrTooLong,
		},
		{
			name:     "underscore",
			input:    "abc_123",
			expected: ErrInvalidChars,
		},
		{
			name:     "dash",
			input:    "abc-123",
			expected: ErrInvalidChars,
		},
		{
			name:     "unicode",
			input:    "abcé123",
			expected: ErrInvalidChars,
		},
		{
			name:     "reserved route",
			input:    "shorten",
			expected: ErrReserved,
		},
		{
			name:     "reserved route with different case",
			input:    "READYZ",
			expected: ErrReserved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Normalize(tt.input)
			if !errors.Is(err, tt.expected) {
				t.Fatalf("expected error %q, got %v", tt.expected, err)
			}
		})
	}
}

func TestIsReserved(t *testing.T) {
	t.Parallel()

	if !IsReserved(" Healthz ") {
		t.Fatal("expected healthz to be reserved")
	}

	if IsReserved("AbC123") {
		t.Fatal("did not expect generated-looking code to be reserved")
	}
}
