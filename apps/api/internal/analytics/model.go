package analytics

import (
	"errors"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

var ErrClickedAtRequired = errors.New("click timestamp is required")

type Click struct {
	ShortCode string
	ClickedAt time.Time
}

func NewClick(shortCode string, clickedAt time.Time) (Click, error) {
	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return Click{}, err
	}

	click := Click{
		ShortCode: normalizedShortCode,
		ClickedAt: clickedAt.UTC(),
	}
	if err := click.Validate(); err != nil {
		return Click{}, err
	}

	return click, nil
}

func (c Click) Validate() error {
	var errs []error

	if err := shortcode.Validate(c.ShortCode); err != nil {
		errs = append(errs, err)
	}

	if c.ClickedAt.IsZero() {
		errs = append(errs, ErrClickedAtRequired)
	}

	return errors.Join(errs...)
}

func (c Click) DayStart() time.Time {
	clickedAt := c.ClickedAt.UTC()
	return time.Date(clickedAt.Year(), clickedAt.Month(), clickedAt.Day(), 0, 0, 0, 0, time.UTC)
}
