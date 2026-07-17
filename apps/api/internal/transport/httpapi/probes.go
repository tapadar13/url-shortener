package httpapi

import (
	"context"
	"net/http"
)

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, probeResponse{
		Status: "ok",
	})
}

func handleReadiness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, probeResponse{
		Status: "ready",
	})
}

func newReadinessHandler(checker ReadinessChecker) http.HandlerFunc {
	if checker == nil {
		return handleReadiness
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := checker.Ping(r.Context()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "not_ready", "service dependencies are not ready")
			return
		}

		handleReadiness(w, r)
	}
}

type probeResponse struct {
	Status string `json:"status"`
}
