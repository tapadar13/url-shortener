package analytics

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewReporterValidatesDependenciesAndOptions(t *testing.T) {
	t.Parallel()

	if _, err := NewReporter(nil, ReporterOptions{}); !errors.Is(err, ErrReaderRequired) {
		t.Fatalf("expected reader error, got %v", err)
	}

	if _, err := NewReporter(&fakeReader{}, ReporterOptions{MaxRangeDays: -1}); !errors.Is(err, ErrMaxRangeDaysInvalid) {
		t.Fatalf("expected maximum range error, got %v", err)
	}

	reporter, err := NewReporter(&fakeReader{}, ReporterOptions{})
	if err != nil {
		t.Fatalf("expected default reporter options: %v", err)
	}
	if reporter.maxRangeDays != DefaultMaxRangeDays {
		t.Fatalf("expected default maximum %d days, got %d", DefaultMaxRangeDays, reporter.maxRangeDays)
	}
}

func TestReporterBuildsContiguousDailyReport(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 7, 11, 0, 0, 0, 0, time.UTC)
	reader := &fakeReader{daily: []DailyClicks{
		{DayStart: start.Add(48 * time.Hour), Clicks: 4},
		{DayStart: start, Clicks: 2},
	}}
	reporter := newTestReporter(t, reader, ReporterOptions{MaxRangeDays: 3})

	report, err := reporter.Get(context.Background(), Range{
		ShortCode:    " AbC123 ",
		Start:        start.Add(12 * time.Hour),
		EndExclusive: start.Add(72 * time.Hour),
	})
	if err != nil {
		t.Fatalf("expected analytics report: %v", err)
	}

	if report.ShortCode != "AbC123" || !report.Start.Equal(start) || !report.EndExclusive.Equal(start.Add(72*time.Hour)) || report.TotalClicks != 6 {
		t.Fatalf("expected normalized report metadata, got %+v", report)
	}

	if len(report.Daily) != 3 ||
		report.Daily[0].Clicks != 2 ||
		report.Daily[1].Clicks != 0 ||
		report.Daily[2].Clicks != 4 {
		t.Fatalf("expected contiguous daily counts, got %+v", report.Daily)
	}

	if reader.rangeValue != (Range{ShortCode: "AbC123", Start: start, EndExclusive: start.Add(72 * time.Hour)}) {
		t.Fatalf("expected normalized reader range, got %+v", reader.rangeValue)
	}
}

func TestReporterRejectsOversizedRangeBeforeQuerying(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	reader := &fakeReader{}
	reporter := newTestReporter(t, reader, ReporterOptions{MaxRangeDays: 2})

	_, err := reporter.Get(context.Background(), Range{
		ShortCode:    "AbC123",
		Start:        start,
		EndExclusive: start.Add(72 * time.Hour),
	})
	if !errors.Is(err, ErrRangeTooLarge) {
		t.Fatalf("expected range too large error, got %v", err)
	}

	if reader.called {
		t.Fatal("expected reader not to be called")
	}
}

func TestReporterReturnsReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	reporter := newTestReporter(t, &fakeReader{err: expectedErr}, ReporterOptions{})

	_, err := reporter.Get(context.Background(), validReporterRange())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected reader error, got %v", err)
	}
}

func TestReporterRejectsInvalidStoredDailyClicks(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		daily    []DailyClicks
		expected error
	}{
		{name: "negative", daily: []DailyClicks{{DayStart: start, Clicks: -1}}, expected: ErrNegativeClicks},
		{name: "before range", daily: []DailyClicks{{DayStart: start.Add(-24 * time.Hour), Clicks: 1}}, expected: ErrDailyClicksOutOfRange},
		{name: "after range", daily: []DailyClicks{{DayStart: start.Add(48 * time.Hour), Clicks: 1}}, expected: ErrDailyClicksOutOfRange},
		{name: "duplicate", daily: []DailyClicks{{DayStart: start, Clicks: 1}, {DayStart: start, Clicks: 2}}, expected: ErrDuplicateDailyClicks},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reporter := newTestReporter(t, &fakeReader{daily: tt.daily}, ReporterOptions{})
			_, err := reporter.Get(context.Background(), Range{
				ShortCode:    "AbC123",
				Start:        start,
				EndExclusive: start.Add(48 * time.Hour),
			})
			if !errors.Is(err, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, err)
			}
		})
	}
}

func TestNilReporterIsSafe(t *testing.T) {
	t.Parallel()

	var reporter *Reporter
	if _, err := reporter.Get(context.Background(), validReporterRange()); !errors.Is(err, ErrReaderRequired) {
		t.Fatalf("expected reader error, got %v", err)
	}
}

func newTestReporter(t *testing.T, reader Reader, options ReporterOptions) *Reporter {
	t.Helper()

	reporter, err := NewReporter(reader, options)
	if err != nil {
		t.Fatalf("create analytics reporter: %v", err)
	}

	return reporter
}

func validReporterRange() Range {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	return Range{
		ShortCode:    "AbC123",
		Start:        start,
		EndExclusive: start.Add(24 * time.Hour),
	}
}

type fakeReader struct {
	called     bool
	rangeValue Range
	daily      []DailyClicks
	err        error
}

func (r *fakeReader) FindDailyClicks(_ context.Context, rangeValue Range) ([]DailyClicks, error) {
	r.called = true
	r.rangeValue = rangeValue
	return r.daily, r.err
}
