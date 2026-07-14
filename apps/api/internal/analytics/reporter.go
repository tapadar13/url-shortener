package analytics

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const DefaultMaxRangeDays = 90

var (
	ErrReaderRequired        = errors.New("analytics reader is required")
	ErrMaxRangeDaysInvalid   = errors.New("analytics maximum range days must be greater than zero")
	ErrRangeTooLarge         = errors.New("analytics range exceeds the maximum number of days")
	ErrDailyClicksOutOfRange = errors.New("daily clicks fall outside the analytics range")
	ErrDuplicateDailyClicks  = errors.New("duplicate daily clicks were returned")
)

type ReporterOptions struct {
	MaxRangeDays int
}

type Report struct {
	ShortCode    string
	Start        time.Time
	EndExclusive time.Time
	TotalClicks  int64
	Daily        []DailyClicks
}

type Reporter struct {
	reader       Reader
	maxRangeDays int
}

func NewReporter(reader Reader, options ReporterOptions) (*Reporter, error) {
	if reader == nil {
		return nil, ErrReaderRequired
	}

	if options.MaxRangeDays == 0 {
		options.MaxRangeDays = DefaultMaxRangeDays
	}
	if options.MaxRangeDays < 0 {
		return nil, ErrMaxRangeDaysInvalid
	}

	return &Reporter{
		reader:       reader,
		maxRangeDays: options.MaxRangeDays,
	}, nil
}

func (r *Reporter) Get(ctx context.Context, rangeValue Range) (Report, error) {
	if r == nil || r.reader == nil {
		return Report{}, ErrReaderRequired
	}

	normalized, err := NewRange(rangeValue.ShortCode, rangeValue.Start, rangeValue.EndExclusive)
	if err != nil {
		return Report{}, err
	}

	rangeDays := int(normalized.EndExclusive.Sub(normalized.Start) / (24 * time.Hour))
	if rangeDays > r.maxRangeDays {
		return Report{}, fmt.Errorf("%w: maximum is %d days", ErrRangeTooLarge, r.maxRangeDays)
	}

	stored, err := r.reader.FindDailyClicks(ctx, normalized)
	if err != nil {
		return Report{}, fmt.Errorf("find daily click analytics: %w", err)
	}

	counts := make(map[time.Time]int64, len(stored))
	for _, item := range stored {
		daily, err := NewDailyClicks(item.DayStart, item.Clicks)
		if err != nil {
			return Report{}, err
		}

		if daily.DayStart.Before(normalized.Start) || !daily.DayStart.Before(normalized.EndExclusive) {
			return Report{}, fmt.Errorf("%w: %s", ErrDailyClicksOutOfRange, daily.DayStart.Format(time.DateOnly))
		}

		if _, exists := counts[daily.DayStart]; exists {
			return Report{}, fmt.Errorf("%w: %s", ErrDuplicateDailyClicks, daily.DayStart.Format(time.DateOnly))
		}

		counts[daily.DayStart] = daily.Clicks
	}

	daily := make([]DailyClicks, 0, rangeDays)
	var total int64
	for day := normalized.Start; day.Before(normalized.EndExclusive); day = day.AddDate(0, 0, 1) {
		clicks := counts[day]
		daily = append(daily, DailyClicks{DayStart: day, Clicks: clicks})
		total += clicks
	}

	return Report{
		ShortCode:    normalized.ShortCode,
		Start:        normalized.Start,
		EndExclusive: normalized.EndExclusive,
		TotalClicks:  total,
		Daily:        daily,
	}, nil
}
