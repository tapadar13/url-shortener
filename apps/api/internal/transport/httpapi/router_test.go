package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

	NewRouter().ServeHTTP(recorder, request)

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
