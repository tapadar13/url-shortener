package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunAcceptsReadyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/readyz" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := run([]string{server.URL + "/readyz"}, server.Client()); err != nil {
		t.Fatalf("expected health check to succeed: %v", err)
	}
}

func TestRunRejectsUnreadyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	err := run([]string{server.URL}, server.Client())
	if err == nil || !strings.Contains(err.Error(), "503 Service Unavailable") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestRunValidatesArguments(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		client *http.Client
	}{
		{name: "too many endpoints", args: []string{"one", "two"}, client: http.DefaultClient},
		{name: "missing client", client: nil},
		{name: "invalid endpoint", args: []string{":"}, client: http.DefaultClient},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := run(test.args, test.client); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
