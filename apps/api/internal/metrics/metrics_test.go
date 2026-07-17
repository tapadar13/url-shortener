package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestMetricsGroupsRequestsByRouteAndStatusClass(t *testing.T) {
	measurements := New()
	measurements.Observe(http.MethodGet, "/shorten/{shortCode}", http.StatusOK, time.Millisecond)
	measurements.Observe(http.MethodGet, "/shorten/{shortCode}", http.StatusNotFound, time.Millisecond)

	snapshot := measurements.Snapshot()
	if len(snapshot) != 2 {
		t.Fatalf("expected two metric groups, got %d", len(snapshot))
	}
}

func TestMiddlewareUsesChiRoutePatternAndCapturesStatus(t *testing.T) {
	measurements := New()
	router := chi.NewRouter()
	router.Use(Middleware(measurements))
	router.Get("/shorten/{shortCode}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/shorten/AbC123", nil))

	snapshot := measurements.Snapshot()
	if len(snapshot) != 1 || snapshot[0].Route != "/shorten/{shortCode}" || snapshot[0].Status != "4" {
		t.Fatalf("unexpected metrics snapshot: %+v", snapshot)
	}
	if snapshot[0].Requests != 1 {
		t.Fatalf("expected one request, got %d", snapshot[0].Requests)
	}
}
