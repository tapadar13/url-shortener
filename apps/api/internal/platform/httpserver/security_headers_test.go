package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersAddsBrowserProtections(t *testing.T) {
	t.Parallel()

	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	expected := map[string]string{
		"Content-Security-Policy": contentSecurityPolicy,
		"Permissions-Policy":      permissionsPolicy,
		"Referrer-Policy":         "no-referrer",
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
	}

	for header, value := range expected {
		if actual := recorder.Header().Get(header); actual != value {
			t.Fatalf("expected %s=%q, got %q", header, value, actual)
		}
	}
}
