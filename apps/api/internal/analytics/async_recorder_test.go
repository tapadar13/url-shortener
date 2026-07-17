package analytics

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
)

func TestNewAsyncRecorderValidatesDependenciesAndOptions(t *testing.T) {
	t.Parallel()

	validOptions := AsyncRecorderOptions{Workers: 1, QueueSize: 1, Timeout: time.Second}
	if _, err := NewAsyncRecorder(nil, validOptions); !errors.Is(err, ErrRecorderRequired) {
		t.Fatalf("expected recorder error, got %v", err)
	}

	recorder := &recordingClickRecorder{}
	tests := []struct {
		name     string
		options  AsyncRecorderOptions
		expected error
	}{
		{name: "workers", options: AsyncRecorderOptions{QueueSize: 1, Timeout: time.Second}, expected: ErrWorkersInvalid},
		{name: "queue size", options: AsyncRecorderOptions{Workers: 1, Timeout: time.Second}, expected: ErrQueueSizeInvalid},
		{name: "timeout", options: AsyncRecorderOptions{Workers: 1, QueueSize: 1}, expected: ErrTimeoutInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewAsyncRecorder(recorder, tt.options); !errors.Is(err, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, err)
			}
		})
	}
}

func TestAsyncRecorderProcessesNormalizedClick(t *testing.T) {
	t.Parallel()

	sink := &recordingClickRecorder{recorded: make(chan Click, 1)}
	recorder := newTestAsyncRecorder(t, sink, AsyncRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   time.Second,
	})
	t.Cleanup(func() { closeTestAsyncRecorder(t, recorder) })

	clickedAt := time.Date(2026, 7, 14, 10, 0, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	if err := recorder.Enqueue(" AbC123 ", clickedAt); err != nil {
		t.Fatalf("expected click to be queued: %v", err)
	}

	select {
	case click := <-sink.recorded:
		if click.ShortCode != "AbC123" || !click.ClickedAt.Equal(clickedAt.UTC()) || click.ClickedAt.Location() != time.UTC {
			t.Fatalf("expected normalized click, got %+v", click)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for queued click")
	}
}

func TestAsyncRecorderReportsQueueBackpressure(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	sink := &recordingClickRecorder{
		started: make(chan struct{}, 1),
		release: release,
	}
	recorder := newTestAsyncRecorder(t, sink, AsyncRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   time.Second,
	})

	now := time.Now()
	if err := recorder.Enqueue("AbC123", now); err != nil {
		t.Fatalf("enqueue first click: %v", err)
	}

	select {
	case <-sink.started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker")
	}

	if err := recorder.Enqueue("XyZ789", now); err != nil {
		t.Fatalf("enqueue buffered click: %v", err)
	}

	if err := recorder.Enqueue("LmN456", now); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected queue full error, got %v", err)
	}

	close(release)
	closeTestAsyncRecorder(t, recorder)
}

func TestAsyncRecorderDrainsAcceptedClicksOnClose(t *testing.T) {
	t.Parallel()

	sink := &recordingClickRecorder{}
	recorder := newTestAsyncRecorder(t, sink, AsyncRecorderOptions{
		Workers:   1,
		QueueSize: 3,
		Timeout:   time.Second,
	})

	for _, code := range []string{"AbC123", "XyZ789", "LmN456"} {
		if err := recorder.Enqueue(code, time.Now()); err != nil {
			t.Fatalf("enqueue %s: %v", code, err)
		}
	}

	closeTestAsyncRecorder(t, recorder)
	if sink.callCount() != 3 {
		t.Fatalf("expected three recorded clicks, got %d", sink.callCount())
	}

	if err := recorder.Enqueue("QrS987", time.Now()); !errors.Is(err, ErrRecorderClosed) {
		t.Fatalf("expected closed recorder error, got %v", err)
	}
}

func TestAsyncRecorderReportsWorkerErrors(t *testing.T) {
	t.Parallel()

	reported := make(chan error, 1)
	sink := &recordingClickRecorder{waitForContext: true}
	recorder := newTestAsyncRecorder(t, sink, AsyncRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   20 * time.Millisecond,
		OnError: func(err error) {
			reported <- err
		},
	})
	t.Cleanup(func() { closeTestAsyncRecorder(t, recorder) })

	if err := recorder.Enqueue("AbC123", time.Now()); err != nil {
		t.Fatalf("enqueue click: %v", err)
	}

	select {
	case err := <-reported:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected worker deadline error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker error")
	}
}

func TestAsyncRecorderCloseCancelsWorkersAfterDeadline(t *testing.T) {
	t.Parallel()

	finished := make(chan error, 1)
	sink := &recordingClickRecorder{
		started:        make(chan struct{}, 1),
		finished:       finished,
		waitForContext: true,
	}
	recorder := newTestAsyncRecorder(t, sink, AsyncRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   time.Second,
	})

	if err := recorder.Enqueue("AbC123", time.Now()); err != nil {
		t.Fatalf("enqueue click: %v", err)
	}
	<-sink.started

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := recorder.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected close deadline error, got %v", err)
	}

	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected worker cancellation, got %v", err)
		}
	default:
		t.Fatal("expected close to wait for worker cancellation")
	}

	closeTestAsyncRecorder(t, recorder)
}

func TestAsyncRecorderValidatesClickBeforeQueueing(t *testing.T) {
	t.Parallel()

	sink := &recordingClickRecorder{}
	recorder := newTestAsyncRecorder(t, sink, AsyncRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   time.Second,
	})
	t.Cleanup(func() { closeTestAsyncRecorder(t, recorder) })

	if err := recorder.Enqueue("invalid-code", time.Now()); !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code error, got %v", err)
	}

	if err := recorder.Enqueue("AbC123", time.Time{}); !errors.Is(err, ErrClickedAtRequired) {
		t.Fatalf("expected click timestamp error, got %v", err)
	}

	if sink.callCount() != 0 {
		t.Fatalf("expected recorder not to be called, got %d calls", sink.callCount())
	}
}

func TestNilAsyncRecorderIsSafe(t *testing.T) {
	t.Parallel()

	var recorder *AsyncRecorder
	if err := recorder.Enqueue("AbC123", time.Now()); !errors.Is(err, ErrRecorderClosed) {
		t.Fatalf("expected closed recorder error, got %v", err)
	}

	if err := recorder.Close(nil); err != nil {
		t.Fatalf("expected nil recorder close to be a no-op: %v", err)
	}
}

func newTestAsyncRecorder(t *testing.T, sink Recorder, options AsyncRecorderOptions) *AsyncRecorder {
	t.Helper()

	recorder, err := NewAsyncRecorder(sink, options)
	if err != nil {
		t.Fatalf("create async recorder: %v", err)
	}

	return recorder
}

func closeTestAsyncRecorder(t *testing.T, recorder *AsyncRecorder) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := recorder.Close(ctx); err != nil {
		t.Fatalf("close async recorder: %v", err)
	}
}

type recordingClickRecorder struct {
	mu             sync.Mutex
	calls          int
	recorded       chan Click
	started        chan struct{}
	finished       chan error
	release        chan struct{}
	waitForContext bool
}

func (r *recordingClickRecorder) RecordClick(ctx context.Context, click Click) error {
	if r.started != nil {
		r.started <- struct{}{}
	}

	if r.waitForContext {
		<-ctx.Done()
		if r.finished != nil {
			r.finished <- ctx.Err()
		}
		return ctx.Err()
	}

	if r.release != nil {
		select {
		case <-r.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	r.mu.Lock()
	r.calls++
	r.mu.Unlock()

	if r.recorded != nil {
		r.recorded <- click
	}

	return nil
}

func (r *recordingClickRecorder) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.calls
}
