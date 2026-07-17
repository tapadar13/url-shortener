package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
)

const requestIDHeader = "X-Request-ID"

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				if recovered := recover(); recovered != nil {
					logger.ErrorContext(r.Context(), "http request panicked",
						slog.String("request_id", response.Header().Get(requestIDHeader)),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("panic", fmt.Sprint(recovered)),
						slog.String("stack", string(debug.Stack())),
					)

					if response.Status() == 0 {
						writeError(response, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
					}
				}
			}()

			next.ServeHTTP(response, r)
		})
	}
}
