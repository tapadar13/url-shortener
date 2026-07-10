package httpapi

import (
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

const maxURLRequestBodyBytes = 1 << 20

type urlRequest struct {
	URL string `json:"url"`
}

type urlResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ShortCode string    `json:"shortCode"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func newCreateURLHandler(creator URLCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request urlRequest
		if err := decodeURLRequest(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "request body must contain only a URL field")
			return
		}

		created, err := creator.Create(r.Context(), service.CreateParams{
			LongURL: request.URL,
		})
		if err != nil {
			writeCreateURLError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, newURLResponse(created))
	}
}

func newUpdateURLHandler(updater URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request urlRequest
		if err := decodeURLRequest(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "request body must contain only a URL field")
			return
		}

		updated, err := updater.UpdateLongURL(r.Context(), service.UpdateParams{
			ShortCode: chi.URLParam(r, "shortCode"),
			LongURL:   request.URL,
		})
		if err != nil {
			writeUpdateURLError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, newURLResponse(updated))
	}
}

func newGetURLHandler(finder URLFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		found, err := finder.GetByShortCode(r.Context(), chi.URLParam(r, "shortCode"))
		if err != nil {
			writeShortCodeURLError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, newURLResponse(found))
	}
}

func newDeleteURLHandler(deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := deleter.DeleteByShortCode(r.Context(), chi.URLParam(r, "shortCode")); err != nil {
			writeShortCodeURLError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func newURLResponse(record urlmodel.URL) urlResponse {
	return urlResponse{
		ID:        record.ID,
		URL:       record.LongURL,
		ShortCode: record.ShortCode,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

func decodeURLRequest(w http.ResponseWriter, r *http.Request, request *urlRequest) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxURLRequestBodyBytes)
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

func writeCreateURLError(w http.ResponseWriter, err error) {
	switch {
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

func isShortCodeError(err error) bool {
	return errors.Is(err, shortcode.ErrRequired) ||
		errors.Is(err, shortcode.ErrTooShort) ||
		errors.Is(err, shortcode.ErrTooLong) ||
		errors.Is(err, shortcode.ErrInvalidChars) ||
		errors.Is(err, shortcode.ErrReserved)
}
