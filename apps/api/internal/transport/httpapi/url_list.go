package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
)

type urlListResponse struct {
	Items      []urlResponse `json:"items"`
	NextCursor string        `json:"nextCursor,omitempty"`
}

func newListURLHandler(lister URLLister, baseURL string) http.HandlerFunc {
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
		page, err := lister.ListPageByOwner(r.Context(), service.ListParams{
			OwnerID: ownerID,
			Limit:   limit,
			Cursor:  r.URL.Query().Get("cursor"),
		})
		if err != nil {
			switch {
			case errors.Is(err, service.ErrURLListLimitInvalid):
				writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 100")
			case isCursorError(err):
				writeError(w, http.StatusBadRequest, "invalid_cursor", "cursor is invalid")
			default:
				writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
			}
			return
		}
		responses := make([]urlResponse, 0, len(page.Items))
		for _, record := range page.Items {
			responses = append(responses, newURLResponse(record, baseURL))
		}
		writeJSON(w, http.StatusOK, urlListResponse{Items: responses, NextCursor: page.NextCursor})
	}
}

func isCursorError(err error) bool {
	return errors.Is(err, urlmodel.ErrCursorInvalid) ||
		errors.Is(err, urlmodel.ErrCursorTimestampAbsent) ||
		errors.Is(err, urlmodel.ErrCursorIDRequired)
}
