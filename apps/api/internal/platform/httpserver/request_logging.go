package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			response := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(response, r)

			statusCode := response.Status()
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			logger.InfoContext(r.Context(), "http request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", statusCode),
				slog.Int("bytes", response.BytesWritten()),
				slog.Duration("duration", time.Since(startedAt)),
			)
		})
	}
}
