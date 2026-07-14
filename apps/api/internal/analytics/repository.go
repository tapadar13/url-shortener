package analytics

import "context"

type Recorder interface {
	RecordClick(ctx context.Context, click Click) error
}

type Reader interface {
	FindDailyClicks(ctx context.Context, rangeValue Range) ([]DailyClicks, error)
}
