package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/ratelimit"
)

func TestRateLimitAllowsRequestWithinQuota(t *testing.T) {
	t.Parallel()

	resetAt := time.Date(2026, 7, 13, 10, 1, 0, 0, time.UTC)
	limiter := &fakeRequestRateLimiter{result: ratelimit.Result{
		Allowed:   true,
		Limit:     60,
		Remaining: 42,
		ResetAt:   resetAt,
	}}
	called := false
	handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "/shorten", nil)
	request.RemoteAddr = "203.0.113.10:4321"
	request.Header.Set("X-Forwarded-For", "198.51.100.20")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if !called || recorder.Code != http.StatusNoContent {
		t.Fatalf("expected request to reach handler, called=%t status=%d", called, recorder.Code)
	}

	if limiter.clientKey != "203.0.113.10" {
		t.Fatalf("expected socket client IP, got %q", limiter.clientKey)
	}

	if recorder.Header().Get(rateLimitLimitHeader) != "60" || recorder.Header().Get(rateLimitRemainingHeader) != "42" || recorder.Header().Get(rateLimitResetHeader) != "1783936860" {
		t.Fatalf("expected rate limit headers, got %#v", recorder.Header())
	}
}

func TestRateLimitRejectsRequestOverQuota(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 13, 10, 0, 30, 0, time.UTC)
	limiter := &fakeRequestRateLimiter{result: ratelimit.Result{
		Allowed:   false,
		Limit:     60,
		Remaining: 0,
		ResetAt:   now.Add(30 * time.Second),
	}}
	called := false
	handler := rateLimit(limiter, func() time.Time { return now })(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	response := executeRequestWithBody(t, handler, http.MethodGet, "/AbC123", "")

	if called {
		t.Fatal("expected limited request not to reach handler")
	}

	assertStatus(t, response, http.StatusTooManyRequests)
	assertAPIError(t, response, "rate_limit_exceeded")

	if response.Header.Get("Retry-After") != "30" {
		t.Fatalf("expected retry after 30 seconds, got %q", response.Header.Get("Retry-After"))
	}
}

func TestRateLimitReturnsServiceUnavailableOnLimiterError(t *testing.T) {
	t.Parallel()

	limiter := &fakeRequestRateLimiter{err: errors.New("database unavailable")}
	response := executeRequestWithBody(t, RateLimit(limiter)(http.NotFoundHandler()), http.MethodPost, "/shorten", "")

	assertStatus(t, response, http.StatusServiceUnavailable)
	assertAPIError(t, response, "rate_limit_unavailable")
}

func TestRateLimitBypassesOperationalRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "health", method: http.MethodGet, path: "/healthz"},
		{name: "readiness", method: http.MethodGet, path: "/readyz"},
		{name: "preflight", method: http.MethodOptions, path: "/shorten"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limiter := &fakeRequestRateLimiter{}
			handler := RateLimit(limiter)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			response := executeRequestWithBody(t, handler, tt.method, tt.path, "")

			assertStatus(t, response, http.StatusNoContent)
			if limiter.called {
				t.Fatal("expected rate limiter not to be called")
			}
		})
	}
}

func TestRateLimitAllowsRequestWhenLimiterIsNotConfigured(t *testing.T) {
	t.Parallel()

	handler := RateLimit(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	response := executeRequestWithBody(t, handler, http.MethodGet, "/AbC123", "")

	assertStatus(t, response, http.StatusNoContent)
}

type fakeRequestRateLimiter struct {
	result    ratelimit.Result
	err       error
	called    bool
	clientKey string
}

func (l *fakeRequestRateLimiter) Allow(_ context.Context, clientKey string) (ratelimit.Result, error) {
	l.called = true
	l.clientKey = clientKey

	return l.result, l.err
}
