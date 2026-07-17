package analytics

import (
	"errors"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

var (
	ErrClickedAtRequired  = errors.New("click timestamp is required")
	ErrRangeStartRequired = errors.New("analytics range start is required")
	ErrRangeEndRequired   = errors.New("analytics range end is required")
	ErrRangeInvalid       = errors.New("analytics range end must be after start")
	ErrDayStartRequired   = errors.New("analytics day start is required")
	ErrNegativeClicks     = errors.New("analytics click count cannot be negative")
)

type Click struct {
	ShortCode string
	ClickedAt time.Time
}

type Range struct {
	ShortCode    string
	Start        time.Time
	EndExclusive time.Time
}

type DailyClicks struct {
	DayStart time.Time
	Clicks   int64
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
	return utcDayStart(c.ClickedAt)
}

func NewRange(shortCode string, start time.Time, endExclusive time.Time) (Range, error) {
	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return Range{}, err
	}

	if start.IsZero() {
		return Range{}, ErrRangeStartRequired
	}

	if endExclusive.IsZero() {
		return Range{}, ErrRangeEndRequired
	}

	rangeValue := Range{
		ShortCode:    normalizedShortCode,
		Start:        utcDayStart(start),
		EndExclusive: utcDayStart(endExclusive),
	}
	if !rangeValue.EndExclusive.After(rangeValue.Start) {
		return Range{}, ErrRangeInvalid
	}

	return rangeValue, nil
}

func NewDailyClicks(dayStart time.Time, clicks int64) (DailyClicks, error) {
	if dayStart.IsZero() {
		return DailyClicks{}, ErrDayStartRequired
	}

	if clicks < 0 {
		return DailyClicks{}, ErrNegativeClicks
	}

	return DailyClicks{
		DayStart: utcDayStart(dayStart),
		Clicks:   clicks,
	}, nil
}

func utcDayStart(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
