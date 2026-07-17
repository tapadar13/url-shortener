package analytics

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrRecorderRequired = errors.New("analytics recorder is required")
	ErrRecorderClosed   = errors.New("analytics recorder is closed")
	ErrQueueFull        = errors.New("analytics queue is full")
	ErrWorkersInvalid   = errors.New("analytics workers must be greater than zero")
	ErrQueueSizeInvalid = errors.New("analytics queue size must be greater than zero")
	ErrTimeoutInvalid   = errors.New("analytics write timeout must be greater than zero")
)

type Enqueuer interface {
	Enqueue(shortCode string, clickedAt time.Time) error
}

type AsyncRecorderOptions struct {
	Workers   int
	QueueSize int
	Timeout   time.Duration
	OnError   func(error)
}

type AsyncRecorder struct {
	recorder      Recorder
	queue         chan Click
	timeout       time.Duration
	onError       func(error)
	workerCtx     context.Context
	cancelWorkers context.CancelFunc

	mu        sync.RWMutex
	closed    bool
	closeOnce sync.Once
	workers   sync.WaitGroup
	done      chan struct{}
}

func NewAsyncRecorder(recorder Recorder, options AsyncRecorderOptions) (*AsyncRecorder, error) {
	if recorder == nil {
		return nil, ErrRecorderRequired
	}

	if options.Workers <= 0 {
		return nil, ErrWorkersInvalid
	}

	if options.QueueSize <= 0 {
		return nil, ErrQueueSizeInvalid
	}

	if options.Timeout <= 0 {
		return nil, ErrTimeoutInvalid
	}

	workerCtx, cancel := context.WithCancel(context.Background())
	asyncRecorder := &AsyncRecorder{
		recorder:      recorder,
		queue:         make(chan Click, options.QueueSize),
		timeout:       options.Timeout,
		onError:       options.OnError,
		workerCtx:     workerCtx,
		cancelWorkers: cancel,
		done:          make(chan struct{}),
	}

	asyncRecorder.workers.Add(options.Workers)
	for range options.Workers {
		go asyncRecorder.runWorker()
	}

	return asyncRecorder, nil
}

func (r *AsyncRecorder) Enqueue(shortCode string, clickedAt time.Time) error {
	if r == nil || r.recorder == nil || r.queue == nil {
		return ErrRecorderClosed
	}

	click, err := NewClick(shortCode, clickedAt)
	if err != nil {
		return err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return ErrRecorderClosed
	}

	select {
	case r.queue <- click:
		return nil
	default:
		return ErrQueueFull
	}
}

func (r *AsyncRecorder) Close(ctx context.Context) error {
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
		r.cancelWorkers()
		return nil
	case <-ctx.Done():
		r.cancelWorkers()
		<-r.done
		return fmt.Errorf("close analytics recorder: %w", ctx.Err())
	}
}

func (r *AsyncRecorder) runWorker() {
	defer r.workers.Done()

	for click := range r.queue {
		if r.workerCtx.Err() != nil {
			return
		}

		ctx, cancel := context.WithTimeout(r.workerCtx, r.timeout)
		err := r.recorder.RecordClick(ctx, click)
		cancel()

		if err != nil && r.onError != nil {
			r.onError(fmt.Errorf("record queued click analytics: %w", err))
		}
	}
}

var _ Enqueuer = (*AsyncRecorder)(nil)
