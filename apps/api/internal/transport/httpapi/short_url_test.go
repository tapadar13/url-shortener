package httpapi

import "testing"

func TestBuildShortURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		baseURL   string
		shortCode string
		expected  string
	}{
		{
			name:      "root domain",
			baseURL:   "https://sho.rt",
			shortCode: "AbC123",
			expected:  "https://sho.rt/AbC123",
		},
		{
			name:      "trailing slash",
			baseURL:   "https://sho.rt/",
			shortCode: "AbC123",
			expected:  "https://sho.rt/AbC123",
		},
		{
			name:      "path prefix",
			baseURL:   "https://example.com/links",
			shortCode: "AbC123",
			expected:  "https://example.com/links/AbC123",
		},
		{name: "missing base URL", shortCode: "AbC123"},
		{name: "missing short code", baseURL: "https://sho.rt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if actual := buildShortURL(tt.baseURL, tt.shortCode); actual != tt.expected {
				t.Fatalf("expected short URL %q, got %q", tt.expected, actual)
			}
		})
	}
}
