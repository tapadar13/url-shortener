package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryWritesJSONInternalServerError(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := Recovery(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("database password leaked")
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	assertStatus(t, response, http.StatusInternalServerError)
	assertAPIError(t, response, "internal_error")

	if strings.Contains(recorder.Body.String(), "database password leaked") {
		t.Fatalf("expected panic value to stay out of response, got %q", recorder.Body.String())
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON panic log: %v", err)
	}

	if entry["level"] != "ERROR" || entry["msg"] != "http request panicked" {
		t.Fatalf("expected panic error log, got %#v", entry)
	}

	if entry["path"] != "/healthz" || entry["panic"] != "database password leaked" {
		t.Fatalf("expected panic metadata, got %#v", entry)
	}

	if stack, ok := entry["stack"].(string); !ok || stack == "" {
		t.Fatalf("expected stack trace, got %#v", entry)
	}
}

func TestRecoveryDoesNotOverwriteStartedResponse(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	handler := Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
		panic("after response")
	}))

	request := httptest.NewRequest(http.MethodPost, "/shorten", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected original status %d, got %d", http.StatusCreated, recorder.Code)
	}

	if recorder.Body.String() != "created" {
		t.Fatalf("expected original response body, got %q", recorder.Body.String())
	}
}
