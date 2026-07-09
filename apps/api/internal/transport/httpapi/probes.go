package httpapi

import "net/http"

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

type probeResponse struct {
	Status string `json:"status"`
}
