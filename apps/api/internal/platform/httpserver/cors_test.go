package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	t.Parallel()

	called := false
	handler := CORS([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if !called {
		t.Fatal("expected request handler to be called")
	}

	if recorder.Header().Get(accessControlAllowOrigin) != "http://localhost:3000" {
		t.Fatalf("expected allowed origin header, got %q", recorder.Header().Get(accessControlAllowOrigin))
	}

	if recorder.Header().Get(accessControlExposeHeaders) != exposeHeaders {
		t.Fatalf("expected exposed request ID header, got %q", recorder.Header().Get(accessControlExposeHeaders))
	}
}

func TestCORSHandlesAllowedPreflightRequest(t *testing.T) {
	t.Parallel()

	called := false
	handler := CORS([]string{"http://localhost:3000"})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	request := httptest.NewRequest(http.MethodOptions, "/shorten", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set(accessControlRequestMethod, http.MethodPost)
	request.Header.Set(accessControlRequestHeaders, "Content-Type")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if called {
		t.Fatal("expected preflight request not to reach handler")
	}

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	if recorder.Header().Get(accessControlAllowMethods) != allowMethods || recorder.Header().Get(accessControlAllowHeaders) != allowHeaders {
		t.Fatalf("expected preflight headers, got %#v", recorder.Header())
	}
}

func TestCORSDoesNotAllowUnconfiguredOrigin(t *testing.T) {
	t.Parallel()

	called := false
	handler := CORS([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "https://untrusted.example.com")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if !called {
		t.Fatal("expected request handler to be called")
	}

	if recorder.Header().Get(accessControlAllowOrigin) != "" {
		t.Fatalf("expected no allow-origin header, got %q", recorder.Header().Get(accessControlAllowOrigin))
	}
}

func TestCORSNormalizesConfiguredOrigin(t *testing.T) {
	t.Parallel()

	handler := CORS([]string{"HTTP://LOCALHOST:3000/"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Header().Get(accessControlAllowOrigin) != "http://localhost:3000" {
		t.Fatalf("expected normalized allowed origin, got %q", recorder.Header().Get(accessControlAllowOrigin))
	}
}
