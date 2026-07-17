//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	analyticsrepository "github.com/tapadar13/url-shortener/apps/api/internal/analytics/repository/mongodb"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	urlrepository "github.com/tapadar13/url-shortener/apps/api/internal/url/repository/mongodb"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
)

func TestClickAnalyticsServiceLifecycle(t *testing.T) {
	client := newIntegrationMongoClient(t)
	urlRepository := urlrepository.New(client.URLsCollection())
	analyticsRepository := analyticsrepository.New(client.AnalyticsCollection())
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	recordErrors := make(chan error, 8)
	analyticsRecorder, err := analytics.NewAsyncRecorder(analyticsRepository, analytics.AsyncRecorderOptions{
		Workers:   2,
		QueueSize: 8,
		Timeout:   integrationTimeout,
		OnError: func(err error) {
			recordErrors <- err
		},
	})
	if err != nil {
		t.Fatalf("create analytics recorder: %v", err)
	}
	t.Cleanup(func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), integrationTimeout)
		defer closeCancel()
		if err := analyticsRecorder.Close(closeCtx); err != nil {
			t.Errorf("close analytics recorder: %v", err)
		}
	})

	clickedAt := time.Date(2026, 7, 14, 8, 30, 0, 0, time.UTC)
	shortCode := "AnaFlow1"
	record, err := urlmodel.New(urlmodel.NewParams{
		LongURL:   "https://example.com/analytics",
		ShortCode: shortCode,
		Now:       clickedAt,
	})
	if err != nil {
		t.Fatalf("create URL record: %v", err)
	}
	if _, err := urlRepository.Create(ctx, record); err != nil {
		t.Fatalf("persist URL record: %v", err)
	}

	redirectService, err := service.NewRedirectService(urlRepository, service.RedirectOptions{
		Analytics: analyticsRecorder,
		Now:       func() time.Time { return clickedAt },
	})
	if err != nil {
		t.Fatalf("create redirect service: %v", err)
	}

	const redirectCount = 3
	for range redirectCount {
		resolved, err := redirectService.Resolve(ctx, shortCode)
		if err != nil || resolved.LongURL != record.LongURL {
			t.Fatalf("resolve short URL, record=%+v err=%v", resolved, err)
		}
	}

	closeCtx, closeCancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer closeCancel()
	if err := analyticsRecorder.Close(closeCtx); err != nil {
		t.Fatalf("close analytics recorder: %v", err)
	}

	select {
	case err := <-recordErrors:
		t.Fatalf("unexpected analytics recording error: %v", err)
	default:
	}

	storedURL, err := urlRepository.FindByShortCode(ctx, shortCode)
	if err != nil {
		t.Fatalf("read URL after redirects: %v", err)
	}
	if storedURL.AccessCount != redirectCount {
		t.Fatalf("expected access count %d, got %d", redirectCount, storedURL.AccessCount)
	}

	reporter, err := analytics.NewReporter(analyticsRepository, analytics.ReporterOptions{})
	if err != nil {
		t.Fatalf("create analytics reporter: %v", err)
	}
	dayStart := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	report, err := reporter.Get(ctx, analytics.Range{
		ShortCode:    shortCode,
		Start:        dayStart,
		EndExclusive: dayStart.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("get analytics report: %v", err)
	}

	if report.TotalClicks != redirectCount || len(report.Daily) != 1 || report.Daily[0].Clicks != redirectCount {
		t.Fatalf("expected complete click analytics report, got %+v", report)
	}
}
