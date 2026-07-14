package analytics

import (
	"errors"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

func TestNewClickNormalizesValues(t *testing.T) {
	t.Parallel()

	clickedAt := time.Date(2026, 7, 14, 23, 45, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	click, err := NewClick(" AbC123 ", clickedAt)
	if err != nil {
		t.Fatalf("expected click to be valid: %v", err)
	}

	if click.ShortCode != "AbC123" {
		t.Fatalf("expected normalized short code, got %q", click.ShortCode)
	}

	if !click.ClickedAt.Equal(clickedAt.UTC()) || click.ClickedAt.Location() != time.UTC {
		t.Fatalf("expected UTC click timestamp %s, got %s", clickedAt.UTC(), click.ClickedAt)
	}
}

func TestNewClickRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	if _, err := NewClick("invalid-code", time.Now()); !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code error, got %v", err)
	}

	if _, err := NewClick("AbC123", time.Time{}); !errors.Is(err, ErrClickedAtRequired) {
		t.Fatalf("expected click timestamp error, got %v", err)
	}
}

func TestClickValidateRejectsInvalidPersistedValues(t *testing.T) {
	t.Parallel()

	err := (Click{ShortCode: "invalid-code"}).Validate()
	if !errors.Is(err, shortcode.ErrInvalidChars) || !errors.Is(err, ErrClickedAtRequired) {
		t.Fatalf("expected short code and timestamp errors, got %v", err)
	}
}

func TestClickDayStartUsesUTCBoundary(t *testing.T) {
	t.Parallel()

	clickedAt := time.Date(2026, 7, 14, 2, 15, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	click, err := NewClick("AbC123", clickedAt)
	if err != nil {
		t.Fatalf("expected click to be valid: %v", err)
	}

	expected := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	if dayStart := click.DayStart(); !dayStart.Equal(expected) || dayStart.Location() != time.UTC {
		t.Fatalf("expected UTC day start %s, got %s", expected, dayStart)
	}
}
