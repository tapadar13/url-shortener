package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"github.com/tapadar13/url-shortener/apps/api/internal/url/service"
)

const defaultRedirectStatusCode = http.StatusFound

type createURLRequest struct {
	URL       string     `json:"url"`
	ExpiresAt *time.Time `json:"expiresAt"`
	ShortCode *string    `json:"shortCode"`
}

type updateURLRequest struct {
	URL string `json:"url"`
}

type urlResponse struct {
	ID        string     `json:"id"`
	URL       string     `json:"url"`
	ShortCode string     `json:"shortCode"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type urlStatsResponse struct {
	ID             string     `json:"id"`
	URL            string     `json:"url"`
	ShortCode      string     `json:"shortCode"`
	AccessCount    int64      `json:"accessCount"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	LastAccessedAt *time.Time `json:"lastAccessedAt,omitempty"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

func newCreateURLHandler(creator URLCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request createURLRequest
		if err := decodeURLRequest(w, r, &request); err != nil {
			if isRequestBodyTooLarge(err) {
				writeError(w, http.StatusRequestEntityTooLarge, "request_entity_too_large", "request body exceeds the configured size limit")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid_request", "request body must be a valid JSON object")
			return
		}

		created, err := creator.Create(r.Context(), service.CreateParams{
			LongURL:   request.URL,
			OwnerID:   currentOwnerID(r.Context()),
			ExpiresAt: request.ExpiresAt,
			ShortCode: request.ShortCode,
		})
		if err != nil {
			writeCreateURLError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, newURLResponse(created))
	}
}

func currentOwnerID(ctx context.Context) string {
	ownerID, _ := CurrentUserID(ctx)
	return ownerID
}

func newUpdateURLHandler(updater URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request updateURLRequest
		if err := decodeURLRequest(w, r, &request); err != nil {
			if isRequestBodyTooLarge(err) {
				writeError(w, http.StatusRequestEntityTooLarge, "request_entity_too_large", "request body exceeds the configured size limit")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid_request", "request body must be a valid JSON object")
			return
		}

		params := service.UpdateParams{
			ShortCode: chi.URLParam(r, "shortCode"),
			LongURL:   request.URL,
		}
		var updated urlmodel.URL
		var err error
		if ownerID, ok := CurrentUserID(r.Context()); ok {
			if ownerUpdater, supported := updater.(OwnerURLUpdater); supported {
				updated, err = ownerUpdater.UpdateLongURLForOwner(r.Context(), ownerID, params)
			} else {
				updated, err = updater.UpdateLongURL(r.Context(), params)
			}
		} else {
			updated, err = updater.UpdateLongURL(r.Context(), params)
		}
		if err != nil {
			writeUpdateURLError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, newURLResponse(updated))
	}
}

func newGetURLHandler(finder URLFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		found, err := findURLForRequest(r, finder)
		if err != nil {
			writeShortCodeURLError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, newURLResponse(found))
	}
}

func newGetURLStatsHandler(finder URLFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		found, err := findURLForRequest(r, finder)
		if err != nil {
			writeShortCodeURLError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, newURLStatsResponse(found))
	}
}

func newDeleteURLHandler(deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := deleteURLForRequest(r, deleter)
		if err != nil {
			writeShortCodeURLError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func findURLForRequest(r *http.Request, finder URLFinder) (urlmodel.URL, error) {
	if ownerID, ok := CurrentUserID(r.Context()); ok {
		if ownerFinder, supported := finder.(OwnerURLFinder); supported {
			return ownerFinder.GetByShortCodeForOwner(r.Context(), ownerID, chi.URLParam(r, "shortCode"))
		}
	}
	return finder.GetByShortCode(r.Context(), chi.URLParam(r, "shortCode"))
}

func deleteURLForRequest(r *http.Request, deleter URLDeleter) error {
	if ownerID, ok := CurrentUserID(r.Context()); ok {
		if ownerDeleter, supported := deleter.(OwnerURLDeleter); supported {
			return ownerDeleter.DeleteByShortCodeForOwner(r.Context(), ownerID, chi.URLParam(r, "shortCode"))
		}
	}
	return deleter.DeleteByShortCode(r.Context(), chi.URLParam(r, "shortCode"))
}

func newRedirectHandler(redirector URLRedirector, statusCode int) http.HandlerFunc {
	if statusCode == 0 {
		statusCode = defaultRedirectStatusCode
	}

	return func(w http.ResponseWriter, r *http.Request) {
		resolved, err := redirector.Resolve(r.Context(), chi.URLParam(r, "shortCode"))
		if err != nil {
			writeShortCodeURLError(w, err)
			return
		}

		w.Header().Set("Location", resolved.LongURL)
		w.WriteHeader(statusCode)
	}
}

func newURLResponse(record urlmodel.URL) urlResponse {
	return urlResponse{
		ID:        record.ID,
		URL:       record.LongURL,
		ShortCode: record.ShortCode,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		ExpiresAt: record.ExpiresAt,
	}
}

func newURLStatsResponse(record urlmodel.URL) urlStatsResponse {
	return urlStatsResponse{
		ID:             record.ID,
		URL:            record.LongURL,
		ShortCode:      record.ShortCode,
		AccessCount:    record.AccessCount,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
		LastAccessedAt: record.LastAccessedAt,
		ExpiresAt:      record.ExpiresAt,
	}
}

func decodeURLRequest(_ http.ResponseWriter, r *http.Request, request any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(request); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func isRequestBodyTooLarge(err error) bool {
	var maxBytesError *http.MaxBytesError
	return errors.As(err, &maxBytesError)
}

func writeCreateURLError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "request_timeout", "request timed out")
	case isExpirationError(err):
		writeError(w, http.StatusBadRequest, "invalid_expiration", "expiresAt must be a future RFC3339 timestamp")
	case isShortCodeError(err):
		writeError(w, http.StatusBadRequest, "invalid_short_code", "short code is invalid")
	case errors.Is(err, urlmodel.ErrDuplicateShortCode):
		writeError(w, http.StatusConflict, "short_code_taken", "short code is already in use")
	case isLongURLError(err):
		writeError(w, http.StatusBadRequest, "invalid_url", "url must be a valid http or https URL")
	case errors.Is(err, service.ErrShortCodeRetriesExhausted):
		writeError(w, http.StatusServiceUnavailable, "short_code_unavailable", "unable to generate a unique short code")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func writeShortCodeURLError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "request_timeout", "request timed out")
	case isShortCodeError(err):
		writeError(w, http.StatusBadRequest, "invalid_short_code", "short code is invalid")
	case errors.Is(err, urlmodel.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "short URL was not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func writeUpdateURLError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "request_timeout", "request timed out")
	case isShortCodeError(err):
		writeError(w, http.StatusBadRequest, "invalid_short_code", "short code is invalid")
	case isLongURLError(err):
		writeError(w, http.StatusBadRequest, "invalid_url", "url must be a valid http or https URL")
	case errors.Is(err, urlmodel.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "short URL was not found")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func isLongURLError(err error) bool {
	return errors.Is(err, urlmodel.ErrLongURLRequired) ||
		errors.Is(err, urlmodel.ErrLongURLTooLong) ||
		errors.Is(err, urlmodel.ErrLongURLInvalid) ||
		errors.Is(err, urlmodel.ErrLongURLSchemeUnsupported) ||
		errors.Is(err, urlmodel.ErrLongURLHostRequired)
}

func isExpirationError(err error) bool {
	return errors.Is(err, urlmodel.ErrExpirationInvalid) ||
		errors.Is(err, urlmodel.ErrExpirationNotFuture)
}

func isShortCodeError(err error) bool {
	return errors.Is(err, shortcode.ErrRequired) ||
		errors.Is(err, shortcode.ErrTooShort) ||
		errors.Is(err, shortcode.ErrTooLong) ||
		errors.Is(err, shortcode.ErrInvalidChars) ||
		errors.Is(err, shortcode.ErrReserved)
}
