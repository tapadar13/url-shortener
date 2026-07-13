package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
)

var (
	ErrAccessRecorderClosed   = errors.New("access recorder is closed")
	ErrAccessQueueFull        = errors.New("access recorder queue is full")
	ErrAccessWorkersInvalid   = errors.New("access recorder workers must be greater than zero")
	ErrAccessQueueSizeInvalid = errors.New("access recorder queue size must be greater than zero")
	ErrAccessTimeoutInvalid   = errors.New("access recorder timeout must be greater than zero")
)

type AccessRecorderOptions struct {
	Workers   int
	QueueSize int
	Timeout   time.Duration
	OnError   func(error)
}

type AsyncAccessRecorder struct {
	repository AccessRepository
	queue      chan urlmodel.RecordAccessParams
	timeout    time.Duration
	onError    func(error)

	mu        sync.RWMutex
	closed    bool
	closeOnce sync.Once
	workers   sync.WaitGroup
	done      chan struct{}
}

func NewAsyncAccessRecorder(repository AccessRepository, options AccessRecorderOptions) (*AsyncAccessRecorder, error) {
	if repository == nil {
		return nil, ErrRepositoryRequired
	}

	if options.Workers <= 0 {
		return nil, ErrAccessWorkersInvalid
	}

	if options.QueueSize <= 0 {
		return nil, ErrAccessQueueSizeInvalid
	}

	if options.Timeout <= 0 {
		return nil, ErrAccessTimeoutInvalid
	}

	recorder := &AsyncAccessRecorder{
		repository: repository,
		queue:      make(chan urlmodel.RecordAccessParams, options.QueueSize),
		timeout:    options.Timeout,
		onError:    options.OnError,
		done:       make(chan struct{}),
	}

	recorder.workers.Add(options.Workers)
	for range options.Workers {
		go recorder.runWorker()
	}

	return recorder, nil
}

func (r *AsyncAccessRecorder) Enqueue(shortCode string, accessedAt time.Time) error {
	if r == nil || r.repository == nil || r.queue == nil {
		return ErrAccessRecorderClosed
	}

	normalizedShortCode, err := shortcode.Normalize(shortCode)
	if err != nil {
		return err
	}

	if accessedAt.IsZero() {
		return urlmodel.ErrTimestampRequired
	}

	params := urlmodel.RecordAccessParams{
		ShortCode:  normalizedShortCode,
		AccessedAt: accessedAt.UTC(),
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return ErrAccessRecorderClosed
	}

	select {
	case r.queue <- params:
		return nil
	default:
		return ErrAccessQueueFull
	}
}

func (r *AsyncAccessRecorder) Close(ctx context.Context) error {
	if r == nil || r.queue == nil {
		return nil
	}

	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		close(r.queue)
		r.mu.Unlock()

		go func() {
			r.workers.Wait()
			close(r.done)
		}()
	})

	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-r.done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("close access recorder: %w", ctx.Err())
	}
}

func (r *AsyncAccessRecorder) runWorker() {
	defer r.workers.Done()

	for params := range r.queue {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		_, err := r.repository.RecordAccess(ctx, params)
		cancel()

		if err != nil && r.onError != nil {
			r.onError(fmt.Errorf("record queued URL access: %w", err))
		}
	}
}
