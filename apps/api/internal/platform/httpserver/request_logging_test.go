package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLoggerWritesRequestMetadata(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	request := httptest.NewRequest(http.MethodPost, "/shorten?url=https://example.com/private", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request.WithContext(context.Background()))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON log entry: %v", err)
	}

	if entry["msg"] != "http request completed" {
		t.Fatalf("expected request log message, got %v", entry["msg"])
	}

	if entry["method"] != http.MethodPost || entry["path"] != "/shorten" {
		t.Fatalf("expected method and path metadata, got %#v", entry)
	}

	if entry["status"] != float64(http.StatusCreated) || entry["bytes"] != float64(2) {
		t.Fatalf("expected status and byte metadata, got %#v", entry)
	}

	if _, ok := entry["duration"]; !ok {
		t.Fatalf("expected duration metadata, got %#v", entry)
	}
}

func TestRequestLoggerDefaultsUnwrittenResponseToOK(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := RequestLogger(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON log entry: %v", err)
	}

	if entry["status"] != float64(http.StatusOK) || entry["bytes"] != float64(0) {
		t.Fatalf("expected default status and zero bytes, got %#v", entry)
	}
}

func TestRequestLoggerIncludesRequestID(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := RequestID(RequestLogger(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON log entry: %v", err)
	}

	if entry["request_id"] != recorder.Header().Get(requestIDHeader) {
		t.Fatalf("expected matching request ID in log entry, got %#v", entry)
	}
}
