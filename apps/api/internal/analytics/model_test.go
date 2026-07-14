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

func TestNewRangeNormalizesUTCDateBoundaries(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("IST", 5*60*60+30*60)
	start := time.Date(2026, 7, 14, 2, 15, 0, 0, location)
	end := time.Date(2026, 7, 16, 2, 15, 0, 0, location)

	rangeValue, err := NewRange(" AbC123 ", start, end)
	if err != nil {
		t.Fatalf("expected analytics range to be valid: %v", err)
	}

	expectedStart := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	if rangeValue.ShortCode != "AbC123" ||
		!rangeValue.Start.Equal(expectedStart) ||
		!rangeValue.EndExclusive.Equal(expectedEnd) ||
		rangeValue.Start.Location() != time.UTC ||
		rangeValue.EndExclusive.Location() != time.UTC {
		t.Fatalf("expected normalized analytics range, got %+v", rangeValue)
	}
}

func TestNewRangeRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	day := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name        string
		shortCode   string
		start       time.Time
		end         time.Time
		expectedErr error
	}{
		{name: "short code", shortCode: "invalid-code", start: day, end: day.Add(24 * time.Hour), expectedErr: shortcode.ErrInvalidChars},
		{name: "start", shortCode: "AbC123", end: day, expectedErr: ErrRangeStartRequired},
		{name: "end", shortCode: "AbC123", start: day, expectedErr: ErrRangeEndRequired},
		{name: "equal boundaries", shortCode: "AbC123", start: day, end: day, expectedErr: ErrRangeInvalid},
		{name: "reversed boundaries", shortCode: "AbC123", start: day, end: day.Add(-24 * time.Hour), expectedErr: ErrRangeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewRange(tt.shortCode, tt.start, tt.end); !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestNewDailyClicksNormalizesDayAndValidatesCount(t *testing.T) {
	t.Parallel()

	day := time.Date(2026, 7, 14, 2, 15, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	daily, err := NewDailyClicks(day, 42)
	if err != nil {
		t.Fatalf("expected daily clicks to be valid: %v", err)
	}

	expectedDay := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	if !daily.DayStart.Equal(expectedDay) || daily.DayStart.Location() != time.UTC || daily.Clicks != 42 {
		t.Fatalf("expected normalized daily clicks, got %+v", daily)
	}

	if _, err := NewDailyClicks(time.Time{}, 0); !errors.Is(err, ErrDayStartRequired) {
		t.Fatalf("expected day start error, got %v", err)
	}

	if _, err := NewDailyClicks(day, -1); !errors.Is(err, ErrNegativeClicks) {
		t.Fatalf("expected negative clicks error, got %v", err)
	}
}
