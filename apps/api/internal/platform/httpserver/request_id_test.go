package httpserver

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestRequestIDAddsHeaderAndContextValue(t *testing.T) {
	t.Parallel()

	var contextRequestID string
	handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		contextRequestID = RequestIDFromContext(r.Context())
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(requestIDHeader, "client-provided-id")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	responseRequestID := recorder.Header().Get(requestIDHeader)
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(responseRequestID) {
		t.Fatalf("expected generated request ID, got %q", responseRequestID)
	}

	if contextRequestID != responseRequestID {
		t.Fatalf("expected matching context and response IDs, got context=%q response=%q", contextRequestID, responseRequestID)
	}

	if responseRequestID == "client-provided-id" {
		t.Fatal("expected server-generated request ID")
	}
}

func TestRequestIDGeneratesUniqueValues(t *testing.T) {
	t.Parallel()

	handler := RequestID(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if first.Header().Get(requestIDHeader) == second.Header().Get(requestIDHeader) {
		t.Fatal("expected distinct request IDs")
	}
}
