package url

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrLongURLRequired   = errors.New("long URL is required")
	ErrShortCodeRequired = errors.New("short code is required")
	ErrTimestampRequired = errors.New("timestamp is required")
	ErrNegativeAccesses  = errors.New("access count cannot be negative")
)

type URL struct {
	ID             string
	LongURL        string
	ShortCode      string
	AccessCount    int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LastAccessedAt *time.Time
}

type NewParams struct {
	LongURL   string
	ShortCode string
	Now       time.Time
}

func New(params NewParams) (URL, error) {
	now := params.Now.UTC()

	record := URL{
		LongURL:     strings.TrimSpace(params.LongURL),
		ShortCode:   strings.TrimSpace(params.ShortCode),
		AccessCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := record.Validate(); err != nil {
		return URL{}, err
	}

	return record, nil
}

func (u URL) Validate() error {
	var errs []error

	if strings.TrimSpace(u.LongURL) == "" {
		errs = append(errs, ErrLongURLRequired)
	}

	if strings.TrimSpace(u.ShortCode) == "" {
		errs = append(errs, ErrShortCodeRequired)
	}

	if u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		errs = append(errs, ErrTimestampRequired)
	}

	if u.AccessCount < 0 {
		errs = append(errs, ErrNegativeAccesses)
	}

	return errors.Join(errs...)
}

func (u URL) WithLongURL(longURL string, now time.Time) (URL, error) {
	updated := u
	updated.LongURL = strings.TrimSpace(longURL)
	updated.UpdatedAt = now.UTC()

	if err := updated.Validate(); err != nil {
		return URL{}, err
	}

	return updated, nil
}

func (u URL) WithAccessRecorded(now time.Time) (URL, error) {
	accessedAt := now.UTC()
	updated := u
	updated.AccessCount++
	updated.LastAccessedAt = &accessedAt

	if err := updated.Validate(); err != nil {
		return URL{}, err
	}

	return updated, nil
}
