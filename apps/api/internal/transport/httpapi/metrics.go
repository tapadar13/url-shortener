package httpapi

import (
	"net/http"

	"github.com/tapadar13/url-shortener/apps/api/internal/metrics"
)

func newMetricsHandler(requestMetrics *metrics.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(requestMetrics.Prometheus()))
	}
}
