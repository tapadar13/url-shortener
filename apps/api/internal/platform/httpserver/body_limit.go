package httpserver

import "net/http"

// MaxRequestBody limits request bodies before they reach application handlers.
func MaxRequestBody(bytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if next == nil || bytes <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, bytes)
			next.ServeHTTP(w, r)
		})
	}
}
