package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMaxRequestBodyLimitsReads(t *testing.T) {
	handler := MaxRequestBody(4)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Fatal("expected request body limit error")
		}
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader("12345"))
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 response, got %d", response.Code)
	}
}
