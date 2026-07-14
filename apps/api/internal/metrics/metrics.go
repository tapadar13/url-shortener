package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// Metrics stores bounded request measurements for the API process.
type Metrics struct {
	mu       sync.RWMutex
	requests map[requestKey]*RequestMetric
}

type requestKey struct {
	method string
	route  string
	status string
}

// RequestMetric is a snapshot of requests grouped by method, route, and status class.
type RequestMetric struct {
	Method        string
	Route         string
	Status        string
	Requests      uint64
	DurationNanos uint64
}

func New() *Metrics {
	return &Metrics{requests: make(map[requestKey]*RequestMetric)}
}

func (m *Metrics) Observe(method, route string, status int, duration time.Duration) {
	if m == nil {
		return
	}
	if route == "" {
		route = "unknown"
	}
	key := requestKey{method: method, route: route, status: strconv.Itoa(status / 100)}

	m.mu.Lock()
	defer m.mu.Unlock()
	metric := m.requests[key]
	if metric == nil {
		metric = &RequestMetric{Method: method, Route: route, Status: key.status}
		m.requests[key] = metric
	}
	metric.Requests++
	metric.DurationNanos += uint64(max(duration, 0))
}

func (m *Metrics) Snapshot() []RequestMetric {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]RequestMetric, 0, len(m.requests))
	for _, metric := range m.requests {
		result = append(result, *metric)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Method != result[j].Method {
			return result[i].Method < result[j].Method
		}
		if result[i].Route != result[j].Route {
			return result[i].Route < result[j].Route
		}
		return result[i].Status < result[j].Status
	})
	return result
}

// Prometheus renders the current measurements in Prometheus text format.
func (m *Metrics) Prometheus() string {
	var output strings.Builder
	output.WriteString("# HELP url_shortener_http_requests_total Total HTTP requests by route and status class.\n")
	output.WriteString("# TYPE url_shortener_http_requests_total counter\n")
	output.WriteString("# HELP url_shortener_http_request_duration_seconds_sum Total HTTP request duration in seconds.\n")
	output.WriteString("# TYPE url_shortener_http_request_duration_seconds_sum counter\n")
	for _, metric := range m.Snapshot() {
		labels := fmt.Sprintf("method=\"%s\",route=\"%s\",status_class=\"%sx\"", escapeLabel(metric.Method), escapeLabel(metric.Route), metric.Status)
		fmt.Fprintf(&output, "url_shortener_http_requests_total{%s} %d\n", labels, metric.Requests)
		fmt.Fprintf(&output, "url_shortener_http_request_duration_seconds_sum{%s} %.9f\n", labels, float64(metric.DurationNanos)/1e9)
	}
	return output.String()
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return strings.ReplaceAll(value, "\n", `\n`)
}

func Middleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			writer := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(writer, r)
			route := ""
			if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
				route = routeContext.RoutePattern()
			}
			m.Observe(r.Method, route, writer.status, time.Since(started))
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func max(value, floor time.Duration) time.Duration {
	if value < floor {
		return floor
	}
	return value
}
