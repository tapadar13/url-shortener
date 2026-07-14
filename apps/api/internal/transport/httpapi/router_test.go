package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/metrics"
	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
)

func TestRouterRecordsRequestMetrics(t *testing.T) {
	requestMetrics := metrics.New()
	router := NewRouter(Dependencies{Metrics: requestMetrics})

	response := executeRequestWithBody(t, router, http.MethodGet, "/healthz", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected health check to succeed, got %d", response.StatusCode)
	}

	snapshot := requestMetrics.Snapshot()
	if len(snapshot) != 1 || snapshot[0].Route != "/healthz" || snapshot[0].Status != "2" {
		t.Fatalf("unexpected request metrics: %+v", snapshot)
	}
}

func TestRouterServesPrometheusMetrics(t *testing.T) {
	requestMetrics := metrics.New()
	requestMetrics.Observe(http.MethodGet, "/healthz", http.StatusOK, time.Second)
	router := NewRouter(Dependencies{Metrics: requestMetrics})

	response := executeRequestWithBody(t, router, http.MethodGet, "/metrics", "")
	assertStatus(t, response, http.StatusOK)
	if got := response.Header.Get("Content-Type"); got != "text/plain; version=0.0.4; charset=utf-8" {
		t.Fatalf("unexpected metrics content type %q", got)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read metrics response: %v", err)
	}
	if !bytes.Contains(body, []byte(`url_shortener_http_requests_total{method="GET",route="/healthz",status_class="2x"} 1`)) {
		t.Fatalf("expected Prometheus request metric, got %s", body)
	}
}

func TestRouterHandlesHealthCheck(t *testing.T) {
	t.Parallel()

	response := executeRequest(t, http.MethodGet, "/healthz")

	assertStatus(t, response, http.StatusOK)
	assertJSONContentType(t, response)

	var body probeResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected health status ok, got %q", body.Status)
	}
}

func TestRouterHandlesReadinessCheck(t *testing.T) {
	t.Parallel()

	response := executeRequest(t, http.MethodGet, "/readyz")

	assertStatus(t, response, http.StatusOK)
	assertJSONContentType(t, response)

	var body probeResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if body.Status != "ready" {
		t.Fatalf("expected readiness status ready, got %q", body.Status)
	}
}

func TestRouterChecksDependencyReadiness(t *testing.T) {
	t.Parallel()

	checker := &fakeReadinessChecker{}
	response := executeRequestWithBody(t, NewRouter(Dependencies{ReadinessChecker: checker}), http.MethodGet, "/readyz", "")

	assertStatus(t, response, http.StatusOK)
	assertJSONContentType(t, response)

	if !checker.called {
		t.Fatal("expected readiness checker to be called")
	}
}

func TestRouterReturnsServiceUnavailableWhenDependencyIsNotReady(t *testing.T) {
	t.Parallel()

	checker := &fakeReadinessChecker{err: errors.New("MongoDB unavailable")}
	response := executeRequestWithBody(t, NewRouter(Dependencies{ReadinessChecker: checker}), http.MethodGet, "/readyz", "")

	assertStatus(t, response, http.StatusServiceUnavailable)
	assertAPIError(t, response, "not_ready")
}

func TestRouterReturnsNotFoundForUnknownRoute(t *testing.T) {
	t.Parallel()

	response := executeRequest(t, http.MethodGet, "/missing")

	assertStatus(t, response, http.StatusNotFound)
	assertAPIError(t, response, "not_found")
}

func TestRouterReturnsMethodNotAllowedForUnsupportedMethod(t *testing.T) {
	t.Parallel()

	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: &fakeURLCreator{}}), http.MethodGet, "/shorten", "")

	assertStatus(t, response, http.StatusMethodNotAllowed)
	assertAPIError(t, response, "method_not_allowed")
}

func executeRequest(t *testing.T, method string, path string) *http.Response {
	t.Helper()

	request := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()

	NewRouter(Dependencies{}).ServeHTTP(recorder, request)

	return recorder.Result()
}

func TestRouterCreatesShortURL(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	creator := &fakeURLCreator{
		created: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/articles/123",
			ShortCode: "AbC1234",
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
	}

	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com/articles/123"}`)

	assertStatus(t, response, http.StatusCreated)
	assertJSONContentType(t, response)

	var body urlResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if body.ID != creator.created.ID || body.URL != creator.created.LongURL || body.ShortCode != creator.created.ShortCode {
		t.Fatalf("expected created URL response, got %#v", body)
	}

	if !body.CreatedAt.Equal(createdAt) || !body.UpdatedAt.Equal(createdAt) {
		t.Fatalf("expected response timestamps %s, got created=%s updated=%s", createdAt, body.CreatedAt, body.UpdatedAt)
	}

	if len(creator.params) != 1 || creator.params[0].LongURL != "https://example.com/articles/123" {
		t.Fatalf("expected creation request to be passed to service, got %#v", creator.params)
	}
}

func TestRouterCreatesExpiringShortURL(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(24 * time.Hour)
	creator := &fakeURLCreator{
		created: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/articles/123",
			ShortCode: "AbC1234",
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
			ExpiresAt: &expiresAt,
		},
	}

	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com/articles/123","expiresAt":"2026-07-11T09:00:00Z"}`)

	assertStatus(t, response, http.StatusCreated)

	var body urlResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if len(creator.params) != 1 || creator.params[0].ExpiresAt == nil || !creator.params[0].ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected expiration to be passed to service, got %#v", creator.params)
	}

	if body.ExpiresAt == nil || !body.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected response expiration %s, got %v", expiresAt, body.ExpiresAt)
	}
}

func TestRouterCreatesShortURLWithCustomCode(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	creator := &fakeURLCreator{
		created: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/articles/123",
			ShortCode: "Custom123",
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
	}

	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com/articles/123","shortCode":"Custom123"}`)

	assertStatus(t, response, http.StatusCreated)

	if len(creator.params) != 1 || creator.params[0].ShortCode == nil || *creator.params[0].ShortCode != "Custom123" {
		t.Fatalf("expected custom short code to be passed to service, got %#v", creator.params)
	}
}

func TestRouterRejectsInvalidCreateURLRequest(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com","unexpected":true}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_request")

	if len(creator.params) != 0 {
		t.Fatalf("expected service not to be called, got %#v", creator.params)
	}
}

func TestRouterMapsInvalidURLToBadRequest(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: urlmodel.ErrLongURLSchemeUnsupported}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"ftp://example.com"}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_url")
}

func TestRouterMapsInvalidExpirationToBadRequest(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: urlmodel.ErrExpirationNotFuture}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com","expiresAt":"2026-07-01T00:00:00Z"}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_expiration")
}

func TestRouterMapsInvalidCustomShortCodeToBadRequest(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: shortcode.ErrInvalidChars}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com","shortCode":"invalid-code"}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_short_code")
}

func TestRouterMapsCustomShortCodeCollisionToConflict(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: urlmodel.ErrDuplicateShortCode}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com","shortCode":"Custom123"}`)

	assertStatus(t, response, http.StatusConflict)
	assertAPIError(t, response, "short_code_taken")
}

func TestRouterMapsShortCodeRetryExhaustionToServiceUnavailable(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: service.ErrShortCodeRetriesExhausted}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com"}`)

	assertStatus(t, response, http.StatusServiceUnavailable)
	assertAPIError(t, response, "short_code_unavailable")
}

func TestRouterMapsCreateTimeoutToGatewayTimeout(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: context.DeadlineExceeded}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com"}`)

	assertStatus(t, response, http.StatusGatewayTimeout)
	assertAPIError(t, response, "request_timeout")
}

func TestRouterMapsUnexpectedCreateErrorToInternalServerError(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: errors.New("database unavailable")}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com"}`)

	assertStatus(t, response, http.StatusInternalServerError)
	assertAPIError(t, response, "internal_error")
}

func TestRouterGetsShortURL(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	finder := &fakeURLFinder{
		found: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/articles/123",
			ShortCode: "AbC123",
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
	}

	response := executeRequestWithBody(t, NewRouter(Dependencies{URLFinder: finder}), http.MethodGet, "/shorten/AbC123", "")

	assertStatus(t, response, http.StatusOK)
	assertJSONContentType(t, response)

	var body urlResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if body.ID != finder.found.ID || body.URL != finder.found.LongURL || body.ShortCode != finder.found.ShortCode {
		t.Fatalf("expected found URL response, got %#v", body)
	}

	if finder.shortCode != "AbC123" {
		t.Fatalf("expected short code AbC123, got %q", finder.shortCode)
	}
}

func TestRouterMapsInvalidShortCodeToBadRequest(t *testing.T) {
	t.Parallel()

	finder := &fakeURLFinder{err: shortcode.ErrInvalidChars}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLFinder: finder}), http.MethodGet, "/shorten/invalid-code", "")

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_short_code")
}

func TestRouterMapsMissingShortURLToNotFound(t *testing.T) {
	t.Parallel()

	finder := &fakeURLFinder{err: urlmodel.ErrNotFound}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLFinder: finder}), http.MethodGet, "/shorten/AbC123", "")

	assertStatus(t, response, http.StatusNotFound)
	assertAPIError(t, response, "not_found")
}

func TestRouterMapsLookupTimeoutToGatewayTimeout(t *testing.T) {
	t.Parallel()

	finder := &fakeURLFinder{err: context.DeadlineExceeded}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLFinder: finder}), http.MethodGet, "/shorten/AbC123", "")

	assertStatus(t, response, http.StatusGatewayTimeout)
	assertAPIError(t, response, "request_timeout")
}

func TestRouterMapsUnexpectedLookupErrorToInternalServerError(t *testing.T) {
	t.Parallel()

	finder := &fakeURLFinder{err: errors.New("database unavailable")}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLFinder: finder}), http.MethodGet, "/shorten/AbC123", "")

	assertStatus(t, response, http.StatusInternalServerError)
	assertAPIError(t, response, "internal_error")
}

func TestRouterGetsShortURLStats(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 9, 9, 0, 0, 0, time.UTC)
	lastAccessedAt := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	finder := &fakeURLFinder{
		found: urlmodel.URL{
			ID:             "507f1f77bcf86cd799439011",
			LongURL:        "https://example.com/articles/123",
			ShortCode:      "AbC123",
			AccessCount:    42,
			CreatedAt:      createdAt,
			UpdatedAt:      lastAccessedAt,
			LastAccessedAt: &lastAccessedAt,
		},
	}

	response := executeRequestWithBody(t, NewRouter(Dependencies{URLFinder: finder}), http.MethodGet, "/shorten/AbC123/stats", "")

	assertStatus(t, response, http.StatusOK)
	assertJSONContentType(t, response)

	var body urlStatsResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if body.ID != finder.found.ID || body.URL != finder.found.LongURL || body.ShortCode != finder.found.ShortCode || body.AccessCount != 42 {
		t.Fatalf("expected URL stats response, got %#v", body)
	}

	if body.LastAccessedAt == nil || !body.LastAccessedAt.Equal(lastAccessedAt) {
		t.Fatalf("expected last accessed timestamp %s, got %v", lastAccessedAt, body.LastAccessedAt)
	}

	if finder.shortCode != "AbC123" {
		t.Fatalf("expected short code AbC123, got %q", finder.shortCode)
	}
}

func TestRouterMapsMissingStatsURLToNotFound(t *testing.T) {
	t.Parallel()

	finder := &fakeURLFinder{err: urlmodel.ErrNotFound}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLFinder: finder}), http.MethodGet, "/shorten/AbC123/stats", "")

	assertStatus(t, response, http.StatusNotFound)
	assertAPIError(t, response, "not_found")
}

func TestRouterUpdatesShortURL(t *testing.T) {
	t.Parallel()

	updatedAt := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	updater := &fakeURLUpdater{
		updated: urlmodel.URL{
			ID:        "507f1f77bcf86cd799439011",
			LongURL:   "https://example.com/updated",
			ShortCode: "AbC123",
			CreatedAt: updatedAt.Add(-time.Hour),
			UpdatedAt: updatedAt,
		},
	}

	response := executeRequestWithBody(t, NewRouter(Dependencies{URLUpdater: updater}), http.MethodPut, "/shorten/AbC123", `{"url":"https://example.com/updated"}`)

	assertStatus(t, response, http.StatusOK)
	assertJSONContentType(t, response)

	var body urlResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON response: %v", err)
	}

	if body.ID != updater.updated.ID || body.URL != updater.updated.LongURL || body.ShortCode != updater.updated.ShortCode {
		t.Fatalf("expected updated URL response, got %#v", body)
	}

	if updater.params.ShortCode != "AbC123" || updater.params.LongURL != "https://example.com/updated" {
		t.Fatalf("expected update params, got %#v", updater.params)
	}
}

func TestRouterRejectsInvalidUpdateURLRequest(t *testing.T) {
	t.Parallel()

	updater := &fakeURLUpdater{}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLUpdater: updater}), http.MethodPut, "/shorten/AbC123", `{"url":"https://example.com","unexpected":true}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_request")

	if updater.called {
		t.Fatal("expected update service not to be called")
	}
}

func TestRouterRejectsExpirationOnUpdate(t *testing.T) {
	t.Parallel()

	updater := &fakeURLUpdater{}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLUpdater: updater}), http.MethodPut, "/shorten/AbC123", `{"url":"https://example.com","expiresAt":"2026-07-11T09:00:00Z"}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_request")

	if updater.called {
		t.Fatal("expected update service not to be called")
	}
}

func TestRouterMapsInvalidUpdateURLToBadRequest(t *testing.T) {
	t.Parallel()

	updater := &fakeURLUpdater{err: urlmodel.ErrLongURLSchemeUnsupported}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLUpdater: updater}), http.MethodPut, "/shorten/AbC123", `{"url":"ftp://example.com"}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_url")
}

func TestRouterMapsInvalidUpdateShortCodeToBadRequest(t *testing.T) {
	t.Parallel()

	updater := &fakeURLUpdater{err: shortcode.ErrInvalidChars}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLUpdater: updater}), http.MethodPut, "/shorten/invalid-code", `{"url":"https://example.com"}`)

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_short_code")
}

func TestRouterMapsMissingUpdateURLToNotFound(t *testing.T) {
	t.Parallel()

	updater := &fakeURLUpdater{err: urlmodel.ErrNotFound}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLUpdater: updater}), http.MethodPut, "/shorten/AbC123", `{"url":"https://example.com"}`)

	assertStatus(t, response, http.StatusNotFound)
	assertAPIError(t, response, "not_found")
}

func TestRouterMapsUpdateTimeoutToGatewayTimeout(t *testing.T) {
	t.Parallel()

	updater := &fakeURLUpdater{err: context.DeadlineExceeded}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLUpdater: updater}), http.MethodPut, "/shorten/AbC123", `{"url":"https://example.com"}`)

	assertStatus(t, response, http.StatusGatewayTimeout)
	assertAPIError(t, response, "request_timeout")
}

func TestRouterMapsUnexpectedUpdateErrorToInternalServerError(t *testing.T) {
	t.Parallel()

	updater := &fakeURLUpdater{err: errors.New("database unavailable")}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLUpdater: updater}), http.MethodPut, "/shorten/AbC123", `{"url":"https://example.com"}`)

	assertStatus(t, response, http.StatusInternalServerError)
	assertAPIError(t, response, "internal_error")
}

func TestRouterDeletesShortURL(t *testing.T) {
	t.Parallel()

	deleter := &fakeURLDeleter{}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLDeleter: deleter}), http.MethodDelete, "/shorten/AbC123", "")

	assertStatus(t, response, http.StatusNoContent)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("expected delete response body to be readable: %v", err)
	}

	if len(body) != 0 {
		t.Fatalf("expected empty delete response, got %q", body)
	}

	if deleter.shortCode != "AbC123" {
		t.Fatalf("expected short code AbC123, got %q", deleter.shortCode)
	}
}

func TestRouterMapsInvalidDeleteShortCodeToBadRequest(t *testing.T) {
	t.Parallel()

	deleter := &fakeURLDeleter{err: shortcode.ErrInvalidChars}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLDeleter: deleter}), http.MethodDelete, "/shorten/invalid-code", "")

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_short_code")
}

func TestRouterMapsMissingDeleteURLToNotFound(t *testing.T) {
	t.Parallel()

	deleter := &fakeURLDeleter{err: urlmodel.ErrNotFound}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLDeleter: deleter}), http.MethodDelete, "/shorten/AbC123", "")

	assertStatus(t, response, http.StatusNotFound)
	assertAPIError(t, response, "not_found")
}

func TestRouterMapsUnexpectedDeleteErrorToInternalServerError(t *testing.T) {
	t.Parallel()

	deleter := &fakeURLDeleter{err: errors.New("database unavailable")}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLDeleter: deleter}), http.MethodDelete, "/shorten/AbC123", "")

	assertStatus(t, response, http.StatusInternalServerError)
	assertAPIError(t, response, "internal_error")
}

func TestRouterRedirectsShortURL(t *testing.T) {
	t.Parallel()

	redirector := &fakeURLRedirector{
		resolved: urlmodel.URL{
			LongURL:   "https://example.com/articles/123",
			ShortCode: "AbC123",
		},
	}
	response := executeRequestWithBody(t, NewRouter(Dependencies{
		URLRedirector:      redirector,
		RedirectStatusCode: http.StatusPermanentRedirect,
	}), http.MethodGet, "/AbC123", "")

	assertStatus(t, response, http.StatusPermanentRedirect)

	if response.Header.Get("Location") != redirector.resolved.LongURL {
		t.Fatalf("expected redirect location %q, got %q", redirector.resolved.LongURL, response.Header.Get("Location"))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("expected redirect response body to be readable: %v", err)
	}

	if len(body) != 0 {
		t.Fatalf("expected empty redirect response, got %q", body)
	}

	if redirector.shortCode != "AbC123" {
		t.Fatalf("expected short code AbC123, got %q", redirector.shortCode)
	}
}

func TestRouterRedirectRouteDoesNotShadowHealthCheck(t *testing.T) {
	t.Parallel()

	redirector := &fakeURLRedirector{}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLRedirector: redirector}), http.MethodGet, "/healthz", "")

	assertStatus(t, response, http.StatusOK)

	if redirector.called {
		t.Fatal("expected health check not to invoke redirect service")
	}
}

func TestRouterMapsInvalidRedirectShortCodeToBadRequest(t *testing.T) {
	t.Parallel()

	redirector := &fakeURLRedirector{err: shortcode.ErrInvalidChars}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLRedirector: redirector}), http.MethodGet, "/invalid-code", "")

	assertStatus(t, response, http.StatusBadRequest)
	assertAPIError(t, response, "invalid_short_code")
}

func TestRouterMapsMissingRedirectURLToNotFound(t *testing.T) {
	t.Parallel()

	redirector := &fakeURLRedirector{err: urlmodel.ErrNotFound}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLRedirector: redirector}), http.MethodGet, "/AbC123", "")

	assertStatus(t, response, http.StatusNotFound)
	assertAPIError(t, response, "not_found")
}

func TestRouterMapsUnexpectedRedirectErrorToInternalServerError(t *testing.T) {
	t.Parallel()

	redirector := &fakeURLRedirector{err: errors.New("database unavailable")}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLRedirector: redirector}), http.MethodGet, "/AbC123", "")

	assertStatus(t, response, http.StatusInternalServerError)
	assertAPIError(t, response, "internal_error")
}

func executeRequestWithBody(t *testing.T, router http.Handler, method string, path string, body string) *http.Response {
	t.Helper()

	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	return recorder.Result()
}

func assertStatus(t *testing.T, response *http.Response, expected int) {
	t.Helper()

	if response.StatusCode != expected {
		t.Fatalf("expected status %d, got %d", expected, response.StatusCode)
	}
}

func assertJSONContentType(t *testing.T, response *http.Response) {
	t.Helper()

	if response.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", response.Header.Get("Content-Type"))
	}
}

func assertAPIError(t *testing.T, response *http.Response, expectedCode string) {
	t.Helper()
	assertJSONContentType(t, response)

	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON error response: %v", err)
	}

	if body.Error.Code != expectedCode {
		t.Fatalf("expected error code %q, got %q", expectedCode, body.Error.Code)
	}
}

type fakeURLCreator struct {
	created urlmodel.URL
	err     error
	params  []service.CreateParams
}

type fakeReadinessChecker struct {
	err    error
	called bool
}

func (c *fakeReadinessChecker) Ping(context.Context) error {
	c.called = true
	return c.err
}

func (c *fakeURLCreator) Create(_ context.Context, params service.CreateParams) (urlmodel.URL, error) {
	c.params = append(c.params, params)

	if c.err != nil {
		return urlmodel.URL{}, c.err
	}

	return c.created, nil
}

type fakeURLFinder struct {
	found     urlmodel.URL
	err       error
	shortCode string
}

func (f *fakeURLFinder) GetByShortCode(_ context.Context, shortCode string) (urlmodel.URL, error) {
	f.shortCode = shortCode

	if f.err != nil {
		return urlmodel.URL{}, f.err
	}

	return f.found, nil
}

type fakeURLUpdater struct {
	updated urlmodel.URL
	err     error
	params  service.UpdateParams
	called  bool
}

func (u *fakeURLUpdater) UpdateLongURL(_ context.Context, params service.UpdateParams) (urlmodel.URL, error) {
	u.called = true
	u.params = params

	if u.err != nil {
		return urlmodel.URL{}, u.err
	}

	return u.updated, nil
}

type fakeURLDeleter struct {
	err       error
	shortCode string
}

func (d *fakeURLDeleter) DeleteByShortCode(_ context.Context, shortCode string) error {
	d.shortCode = shortCode

	return d.err
}

type fakeURLRedirector struct {
	resolved  urlmodel.URL
	err       error
	shortCode string
	called    bool
}

func (r *fakeURLRedirector) Resolve(_ context.Context, shortCode string) (urlmodel.URL, error) {
	r.called = true
	r.shortCode = shortCode

	if r.err != nil {
		return urlmodel.URL{}, r.err
	}

	return r.resolved, nil
}
