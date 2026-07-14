package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

func TestRouterReturnsDailyURLAnalytics(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	reporter := &fakeAnalyticsReporter{report: analytics.Report{
		ShortCode:    "AbC123",
		Start:        start,
		EndExclusive: start.Add(48 * time.Hour),
		TotalClicks:  5,
		Daily: []analytics.DailyClicks{
			{DayStart: start, Clicks: 2},
			{DayStart: start.Add(24 * time.Hour), Clicks: 3},
		},
	}}
	finder := &fakeURLFinder{found: urlmodel.URL{ShortCode: "AbC123"}}
	router := NewRouter(Dependencies{
		URLFinder:         finder,
		AnalyticsReporter: reporter,
		AnalyticsNow:      func() time.Time { return time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC) },
	})

	response := executeRequestWithBody(t, router, http.MethodGet, "/shorten/AbC123/analytics?from=2026-07-12&to=2026-07-13", "")
	assertStatus(t, response, http.StatusOK)
	assertJSONContentType(t, response)

	var body analyticsResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode analytics response: %v", err)
	}

	if body.ShortCode != "AbC123" || body.From != "2026-07-12" || body.To != "2026-07-13" || body.TotalClicks != 5 {
		t.Fatalf("expected analytics response metadata, got %+v", body)
	}
	if len(body.Daily) != 2 || body.Daily[0].Date != "2026-07-12" || body.Daily[0].Clicks != 2 || body.Daily[1].Clicks != 3 {
		t.Fatalf("expected daily analytics response, got %+v", body.Daily)
	}

	expectedRange := analytics.Range{
		ShortCode:    "AbC123",
		Start:        start,
		EndExclusive: start.Add(48 * time.Hour),
	}
	if reporter.rangeValue != expectedRange || finder.shortCode != "AbC123" {
		t.Fatalf("expected link lookup and report range, finder=%q range=%+v", finder.shortCode, reporter.rangeValue)
	}
}

func TestRouterDefaultsAnalyticsToLatestThirtyDays(t *testing.T) {
	t.Parallel()

	today := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	reporter := &fakeAnalyticsReporter{}
	router := NewRouter(Dependencies{
		URLFinder:         &fakeURLFinder{},
		AnalyticsReporter: reporter,
		AnalyticsNow:      func() time.Time { return today.Add(12 * time.Hour) },
	})

	response := executeRequestWithBody(t, router, http.MethodGet, "/shorten/AbC123/analytics", "")
	assertStatus(t, response, http.StatusOK)

	if !reporter.rangeValue.Start.Equal(today.AddDate(0, 0, -29)) || !reporter.rangeValue.EndExclusive.Equal(today.AddDate(0, 0, 1)) {
		t.Fatalf("expected latest 30-day range, got %+v", reporter.rangeValue)
	}
}

func TestRouterRejectsInvalidAnalyticsRanges(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
	}{
		{name: "invalid from", query: "from=07-01-2026"},
		{name: "invalid to", query: "to=tomorrow"},
		{name: "reversed", query: "from=2026-07-14&to=2026-07-13"},
		{name: "future", query: "to=2026-07-15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			finder := &fakeURLFinder{}
			reporter := &fakeAnalyticsReporter{}
			router := NewRouter(Dependencies{
				URLFinder:         finder,
				AnalyticsReporter: reporter,
				AnalyticsNow:      func() time.Time { return time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC) },
			})

			response := executeRequestWithBody(t, router, http.MethodGet, "/shorten/AbC123/analytics?"+tt.query, "")
			assertStatus(t, response, http.StatusBadRequest)
			assertAPIError(t, response, "invalid_analytics_range")

			if finder.shortCode != "" || reporter.called {
				t.Fatal("expected invalid range to be rejected before dependencies are called")
			}
		})
	}
}

func TestRouterMapsMissingAnalyticsURLToNotFound(t *testing.T) {
	t.Parallel()

	reporter := &fakeAnalyticsReporter{}
	router := NewRouter(Dependencies{
		URLFinder:         &fakeURLFinder{err: urlmodel.ErrNotFound},
		AnalyticsReporter: reporter,
		AnalyticsNow:      func() time.Time { return time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC) },
	})

	response := executeRequestWithBody(t, router, http.MethodGet, "/shorten/AbC123/analytics", "")
	assertStatus(t, response, http.StatusNotFound)
	assertAPIError(t, response, "not_found")

	if reporter.called {
		t.Fatal("expected reporter not to be called for missing URL")
	}
}

func TestRouterMapsAnalyticsReporterErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		statusCode int
		code       string
	}{
		{name: "range too large", err: analytics.ErrRangeTooLarge, statusCode: http.StatusBadRequest, code: "analytics_range_too_large"},
		{name: "timeout", err: context.DeadlineExceeded, statusCode: http.StatusGatewayTimeout, code: "request_timeout"},
		{name: "unexpected", err: errors.New("database unavailable"), statusCode: http.StatusInternalServerError, code: "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewRouter(Dependencies{
				URLFinder:         &fakeURLFinder{},
				AnalyticsReporter: &fakeAnalyticsReporter{err: tt.err},
				AnalyticsNow:      func() time.Time { return time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC) },
			})

			response := executeRequestWithBody(t, router, http.MethodGet, "/shorten/AbC123/analytics", "")
			assertStatus(t, response, tt.statusCode)
			assertAPIError(t, response, tt.code)
		})
	}
}

type fakeAnalyticsReporter struct {
	called     bool
	rangeValue analytics.Range
	report     analytics.Report
	err        error
}

func (r *fakeAnalyticsReporter) Get(_ context.Context, rangeValue analytics.Range) (analytics.Report, error) {
	r.called = true
	r.rangeValue = rangeValue
	if r.err != nil {
		return analytics.Report{}, r.err
	}

	if r.report.ShortCode == "" {
		return analytics.Report{
			ShortCode:    rangeValue.ShortCode,
			Start:        rangeValue.Start,
			EndExclusive: rangeValue.EndExclusive,
			Daily:        []analytics.DailyClicks{},
		}, nil
	}

	return r.report, nil
}
