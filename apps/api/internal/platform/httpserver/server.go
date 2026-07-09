package httpserver

import (
	"net/http"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
)

const (
	readHeaderTimeout = 5 * time.Second
	idleTimeout       = 60 * time.Second
)

func New(cfg config.Config, handler http.Handler) *http.Server {
	if handler == nil {
		handler = http.NotFoundHandler()
	}

	return &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           handler,
		ReadTimeout:       cfg.RequestTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      cfg.RequestTimeout,
		IdleTimeout:       idleTimeout,
	}
}
