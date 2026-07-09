package httpserver

import (
	"net/http"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
)

func TestNewConfiguresHTTPServer(t *testing.T) {
	t.Parallel()

	handler := http.NewServeMux()

	cfg, err := config.LoadFromMap(map[string]string{
		"HTTP_ADDR":       ":9090",
		"REQUEST_TIMEOUT": "3s",
	})
	if err != nil {
		t.Fatalf("expected valid config: %v", err)
	}

	server := New(cfg, handler)

	if server.Addr != ":9090" {
		t.Fatalf("expected server address :9090, got %q", server.Addr)
	}

	if server.Handler != handler {
		t.Fatal("expected configured handler to be used")
	}

	if server.ReadTimeout != 3*time.Second {
		t.Fatalf("expected read timeout 3s, got %s", server.ReadTimeout)
	}

	if server.WriteTimeout != 3*time.Second {
		t.Fatalf("expected write timeout 3s, got %s", server.WriteTimeout)
	}

	if server.ReadHeaderTimeout != readHeaderTimeout {
		t.Fatalf("expected read header timeout %s, got %s", readHeaderTimeout, server.ReadHeaderTimeout)
	}

	if server.IdleTimeout != idleTimeout {
		t.Fatalf("expected idle timeout %s, got %s", idleTimeout, server.IdleTimeout)
	}
}

func TestNewUsesNotFoundHandlerWhenHandlerIsNil(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFromMap(nil)
	if err != nil {
		t.Fatalf("expected valid config: %v", err)
	}

	server := New(cfg, nil)

	if server.Handler == nil {
		t.Fatal("expected fallback handler")
	}
}
