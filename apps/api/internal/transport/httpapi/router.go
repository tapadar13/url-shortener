package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	router := chi.NewRouter()

	router.Get("/healthz", handleHealth)
	router.Get("/readyz", handleReadiness)

	return router
}
