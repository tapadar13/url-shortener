package metrics

import (
	"net/http"
	"strconv"
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
	return result
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
