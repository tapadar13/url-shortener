package httpapi

import (
	"net/http"
	"strconv"
)

func newListURLHandler(lister URLLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerID, ok := CurrentUserID(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
			return
		}
		limit := int64(0)
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
				return
			}
			limit = parsed
		}
		urls, err := lister.ListByOwner(r.Context(), ownerID, limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_limit", err.Error())
			return
		}
		responses := make([]urlResponse, 0, len(urls))
		for _, record := range urls {
			responses = append(responses, newURLResponse(record))
		}
		writeJSON(w, http.StatusOK, responses)
	}
}
