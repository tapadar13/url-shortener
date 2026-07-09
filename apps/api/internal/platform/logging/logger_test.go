package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
)

func TestNewWritesJSONLogs(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := New(config.LogConfig{
		Level:  config.LogLevelInfo,
		Format: config.LogFormatJSON,
	}, &output)

	logger.InfoContext(context.Background(), "service started", slog.String("environment", "test"))

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("expected JSON log entry: %v", err)
	}

	if entry["level"] != "INFO" {
		t.Fatalf("expected INFO level, got %v", entry["level"])
	}

	if entry["msg"] != "service started" {
		t.Fatalf("expected message to be logged, got %v", entry["msg"])
	}

	if entry["environment"] != "test" {
		t.Fatalf("expected attribute to be logged, got %v", entry["environment"])
	}
}

func TestNewWritesTextLogsByDefault(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := New(config.LogConfig{
		Level:  config.LogLevelInfo,
		Format: config.LogFormatText,
	}, &output)

	logger.InfoContext(context.Background(), "service started")

	logLine := output.String()
	if !strings.Contains(logLine, "level=INFO") {
		t.Fatalf("expected text log to include INFO level, got %q", logLine)
	}

	if !strings.Contains(logLine, `msg="service started"`) {
		t.Fatalf("expected text log to include message, got %q", logLine)
	}
}

func TestNewFiltersLogsBelowConfiguredLevel(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := New(config.LogConfig{
		Level:  config.LogLevelWarn,
		Format: config.LogFormatJSON,
	}, &output)

	logger.InfoContext(context.Background(), "ignored")
	logger.WarnContext(context.Background(), "included")

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one log line, got %d: %q", len(lines), output.String())
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("expected JSON log entry: %v", err)
	}

	if entry["level"] != "WARN" {
		t.Fatalf("expected WARN level, got %v", entry["level"])
	}

	if entry["msg"] != "included" {
		t.Fatalf("expected warning message, got %v", entry["msg"])
	}
}
