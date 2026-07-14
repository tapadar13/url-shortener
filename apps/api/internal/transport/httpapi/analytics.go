package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
)

const defaultAnalyticsRangeDays = 30

var (
	errAnalyticsDateInvalid = errors.New("analytics date must use YYYY-MM-DD format")
	errAnalyticsRangeFuture = errors.New("analytics range cannot include future dates")
)

type analyticsDailyResponse struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

type analyticsResponse struct {
	ShortCode   string                   `json:"shortCode"`
	From        string                   `json:"from"`
	To          string                   `json:"to"`
	TotalClicks int64                    `json:"totalClicks"`
	Daily       []analyticsDailyResponse `json:"daily"`
}

func newGetURLAnalyticsHandler(
	finder URLFinder,
	reporter URLAnalyticsReporter,
	now func() time.Time,
) http.HandlerFunc {
	if now == nil {
		now = time.Now
	}

	return func(w http.ResponseWriter, r *http.Request) {
		rangeValue, err := analyticsRangeFromRequest(r, now())
		if err != nil {
			writeAnalyticsRangeError(w, err)
			return
		}

		if _, err := finder.GetByShortCode(r.Context(), rangeValue.ShortCode); err != nil {
			writeShortCodeURLError(w, err)
			return
		}

		report, err := reporter.Get(r.Context(), rangeValue)
		if err != nil {
			writeAnalyticsReportError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, newAnalyticsResponse(report))
	}
}

func analyticsRangeFromRequest(r *http.Request, now time.Time) (analytics.Range, error) {
	today := utcDate(now)
	to := today
	if rawTo := r.URL.Query().Get("to"); rawTo != "" {
		parsed, err := time.Parse(time.DateOnly, rawTo)
		if err != nil {
			return analytics.Range{}, fmt.Errorf("%w: to", errAnalyticsDateInvalid)
		}
		to = parsed
	}

	if to.After(today) {
		return analytics.Range{}, errAnalyticsRangeFuture
	}

	from := to.AddDate(0, 0, -(defaultAnalyticsRangeDays - 1))
	if rawFrom := r.URL.Query().Get("from"); rawFrom != "" {
		parsed, err := time.Parse(time.DateOnly, rawFrom)
		if err != nil {
			return analytics.Range{}, fmt.Errorf("%w: from", errAnalyticsDateInvalid)
		}
		from = parsed
	}

	return analytics.NewRange(
		chi.URLParam(r, "shortCode"),
		from,
		to.AddDate(0, 0, 1),
	)
}

func newAnalyticsResponse(report analytics.Report) analyticsResponse {
	daily := make([]analyticsDailyResponse, 0, len(report.Daily))
	for _, item := range report.Daily {
		daily = append(daily, analyticsDailyResponse{
			Date:   item.DayStart.Format(time.DateOnly),
			Clicks: item.Clicks,
		})
	}

	return analyticsResponse{
		ShortCode:   report.ShortCode,
		From:        report.Start.Format(time.DateOnly),
		To:          report.EndExclusive.AddDate(0, 0, -1).Format(time.DateOnly),
		TotalClicks: report.TotalClicks,
		Daily:       daily,
	}
}

func writeAnalyticsRangeError(w http.ResponseWriter, err error) {
	if isShortCodeError(err) {
		writeShortCodeURLError(w, err)
		return
	}

	writeError(
		w,
		http.StatusBadRequest,
		"invalid_analytics_range",
		"from and to must be YYYY-MM-DD dates, from must not be after to, and to cannot be in the future",
	)
}

func writeAnalyticsReportError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "request_timeout", "request timed out")
	case errors.Is(err, analytics.ErrRangeTooLarge):
		writeError(w, http.StatusBadRequest, "analytics_range_too_large", "analytics range cannot exceed 90 days")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func utcDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
