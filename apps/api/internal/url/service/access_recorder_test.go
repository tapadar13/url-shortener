package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

func TestNewAsyncAccessRecorderValidatesDependenciesAndOptions(t *testing.T) {
	t.Parallel()

	validOptions := AccessRecorderOptions{Workers: 1, QueueSize: 1, Timeout: time.Second}
	_, err := NewAsyncAccessRecorder(nil, validOptions)
	if !errors.Is(err, ErrRepositoryRequired) {
		t.Fatalf("expected repository error, got %v", err)
	}

	repository := &recordingAccessRepository{}
	tests := []struct {
		name     string
		options  AccessRecorderOptions
		expected error
	}{
		{name: "workers", options: AccessRecorderOptions{QueueSize: 1, Timeout: time.Second}, expected: ErrAccessWorkersInvalid},
		{name: "queue size", options: AccessRecorderOptions{Workers: 1, Timeout: time.Second}, expected: ErrAccessQueueSizeInvalid},
		{name: "timeout", options: AccessRecorderOptions{Workers: 1, QueueSize: 1}, expected: ErrAccessTimeoutInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewAsyncAccessRecorder(repository, tt.options)
			if !errors.Is(err, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, err)
			}
		})
	}
}

func TestAsyncAccessRecorderProcessesNormalizedAccess(t *testing.T) {
	t.Parallel()

	repository := &recordingAccessRepository{recorded: make(chan urlmodel.RecordAccessParams, 1)}
	recorder := newTestAccessRecorder(t, repository, AccessRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   time.Second,
	})
	t.Cleanup(func() { closeTestAccessRecorder(t, recorder) })

	accessedAt := time.Date(2026, 7, 14, 10, 0, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	if err := recorder.Enqueue(" AbC123 ", accessedAt); err != nil {
		t.Fatalf("expected access to be queued: %v", err)
	}

	select {
	case params := <-repository.recorded:
		if params.ShortCode != "AbC123" || !params.AccessedAt.Equal(accessedAt.UTC()) || params.AccessedAt.Location() != time.UTC {
			t.Fatalf("expected normalized access params, got %+v", params)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for queued access")
	}
}

func TestAsyncAccessRecorderReportsQueueBackpressure(t *testing.T) {
	t.Parallel()

	release := make(chan struct{})
	repository := &recordingAccessRepository{
		started: make(chan struct{}, 1),
		release: release,
	}
	recorder := newTestAccessRecorder(t, repository, AccessRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   time.Second,
	})

	now := time.Now()
	if err := recorder.Enqueue("AbC123", now); err != nil {
		t.Fatalf("enqueue first access: %v", err)
	}

	select {
	case <-repository.started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for worker")
	}

	if err := recorder.Enqueue("XyZ789", now); err != nil {
		t.Fatalf("enqueue buffered access: %v", err)
	}

	if err := recorder.Enqueue("LmN456", now); !errors.Is(err, ErrAccessQueueFull) {
		t.Fatalf("expected queue full error, got %v", err)
	}

	close(release)
	closeTestAccessRecorder(t, recorder)
}

func TestAsyncAccessRecorderDrainsAcceptedAccessesOnClose(t *testing.T) {
	t.Parallel()

	repository := &recordingAccessRepository{}
	recorder := newTestAccessRecorder(t, repository, AccessRecorderOptions{
		Workers:   1,
		QueueSize: 3,
		Timeout:   time.Second,
	})

	now := time.Now()
	for _, code := range []string{"AbC123", "XyZ789", "LmN456"} {
		if err := recorder.Enqueue(code, now); err != nil {
			t.Fatalf("enqueue %s: %v", code, err)
		}
	}

	closeTestAccessRecorder(t, recorder)
	if repository.callCount() != 3 {
		t.Fatalf("expected three recorded accesses, got %d", repository.callCount())
	}

	if err := recorder.Enqueue("QrS987", now); !errors.Is(err, ErrAccessRecorderClosed) {
		t.Fatalf("expected closed recorder error, got %v", err)
	}
}

func TestAsyncAccessRecorderReportsWorkerErrors(t *testing.T) {
	t.Parallel()

	reported := make(chan error, 1)
	repository := &recordingAccessRepository{waitForContext: true}
	recorder := newTestAccessRecorder(t, repository, AccessRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   20 * time.Millisecond,
		OnError: func(err error) {
			reported <- err
		},
	})
	t.Cleanup(func() { closeTestAccessRecorder(t, recorder) })

	if err := recorder.Enqueue("AbC123", time.Now()); err != nil {
		t.Fatalf("enqueue access: %v", err)
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

func TestAsyncAccessRecorderCloseHonorsContext(t *testing.T) {
	t.Parallel()

	finished := make(chan error, 1)
	repository := &recordingAccessRepository{
		started:        make(chan struct{}, 1),
		finished:       finished,
		waitForContext: true,
	}
	recorder := newTestAccessRecorder(t, repository, AccessRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   time.Second,
	})

	if err := recorder.Enqueue("AbC123", time.Now()); err != nil {
		t.Fatalf("enqueue access: %v", err)
	}
	<-repository.started

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

	closeTestAccessRecorder(t, recorder)
}

func TestAsyncAccessRecorderValidatesEventsBeforeQueueing(t *testing.T) {
	t.Parallel()

	repository := &recordingAccessRepository{}
	recorder := newTestAccessRecorder(t, repository, AccessRecorderOptions{
		Workers:   1,
		QueueSize: 1,
		Timeout:   time.Second,
	})
	t.Cleanup(func() { closeTestAccessRecorder(t, recorder) })

	if err := recorder.Enqueue("invalid-code", time.Now()); !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code error, got %v", err)
	}

	if err := recorder.Enqueue("AbC123", time.Time{}); !errors.Is(err, urlmodel.ErrTimestampRequired) {
		t.Fatalf("expected timestamp error, got %v", err)
	}

	if repository.callCount() != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repository.callCount())
	}
}

func TestNilAsyncAccessRecorderIsSafe(t *testing.T) {
	t.Parallel()

	var recorder *AsyncAccessRecorder
	if err := recorder.Enqueue("AbC123", time.Now()); !errors.Is(err, ErrAccessRecorderClosed) {
		t.Fatalf("expected closed recorder error, got %v", err)
	}

	if err := recorder.Close(nil); err != nil {
		t.Fatalf("expected nil recorder close to be a no-op: %v", err)
	}
}

func newTestAccessRecorder(t *testing.T, repository AccessRepository, options AccessRecorderOptions) *AsyncAccessRecorder {
	t.Helper()

	recorder, err := NewAsyncAccessRecorder(repository, options)
	if err != nil {
		t.Fatalf("create access recorder: %v", err)
	}

	return recorder
}

func closeTestAccessRecorder(t *testing.T, recorder *AsyncAccessRecorder) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := recorder.Close(ctx); err != nil {
		t.Fatalf("close access recorder: %v", err)
	}
}

type recordingAccessRepository struct {
	mu             sync.Mutex
	calls          int
	recorded       chan urlmodel.RecordAccessParams
	started        chan struct{}
	finished       chan error
	release        chan struct{}
	waitForContext bool
}

func (r *recordingAccessRepository) RecordAccess(ctx context.Context, params urlmodel.RecordAccessParams) (urlmodel.URL, error) {
	if r.started != nil {
		r.started <- struct{}{}
	}

	if r.waitForContext {
		<-ctx.Done()
		if r.finished != nil {
			r.finished <- ctx.Err()
		}
		return urlmodel.URL{}, ctx.Err()
	}

	if r.release != nil {
		select {
		case <-r.release:
		case <-ctx.Done():
			return urlmodel.URL{}, ctx.Err()
		}
	}

	r.mu.Lock()
	r.calls++
	r.mu.Unlock()

	if r.recorded != nil {
		r.recorded <- params
	}

	return urlmodel.URL{}, nil
}

func (r *recordingAccessRepository) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.calls
}
