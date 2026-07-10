package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
)

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

func TestRouterReturnsNotFoundForUnknownRoute(t *testing.T) {
	t.Parallel()

	response := executeRequest(t, http.MethodGet, "/missing")

	assertStatus(t, response, http.StatusNotFound)
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

	var body createURLResponse
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

func TestRouterMapsShortCodeRetryExhaustionToServiceUnavailable(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: service.ErrShortCodeRetriesExhausted}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com"}`)

	assertStatus(t, response, http.StatusServiceUnavailable)
	assertAPIError(t, response, "short_code_unavailable")
}

func TestRouterMapsUnexpectedCreateErrorToInternalServerError(t *testing.T) {
	t.Parallel()

	creator := &fakeURLCreator{err: errors.New("database unavailable")}
	response := executeRequestWithBody(t, NewRouter(Dependencies{URLCreator: creator}), http.MethodPost, "/shorten", `{"url":"https://example.com"}`)

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

func (c *fakeURLCreator) Create(_ context.Context, params service.CreateParams) (urlmodel.URL, error) {
	c.params = append(c.params, params)

	if c.err != nil {
		return urlmodel.URL{}, c.err
	}

	return c.created, nil
}
