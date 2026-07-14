package analytics

import "context"

type Recorder interface {
	RecordClick(ctx context.Context, click Click) error
}
