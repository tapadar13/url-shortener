package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestServeRequiresServer(t *testing.T) {
	t.Parallel()

	err := Serve(context.Background(), nil, time.Second, slog.Default())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServeReturnsListenError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("bind failed")

	err := serve(
		context.Background(),
		time.Second,
		slog.Default(),
		func() error {
			return expectedErr
		},
		func(context.Context) error {
			t.Fatal("shutdown should not be called")
			return nil
		},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected listen error, got %v", err)
	}
}

func TestServeShutsDownWhenContextIsCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	listening := make(chan struct{})
	releaseServer := make(chan struct{})
	shutdownCalled := make(chan context.Context, 1)

	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(
			ctx,
			time.Second,
			slog.Default(),
			func() error {
				close(listening)
				<-releaseServer
				return http.ErrServerClosed
			},
			func(ctx context.Context) error {
				shutdownCalled <- ctx
				close(releaseServer)
				return nil
			},
		)
	}()

	<-listening
	cancel()

	select {
	case shutdownCtx := <-shutdownCalled:
		if _, ok := shutdownCtx.Deadline(); !ok {
			t.Fatal("expected shutdown context to have a deadline")
		}
	case <-time.After(time.Second):
		t.Fatal("expected shutdown to be called")
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected clean shutdown, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("expected serve to return")
	}
}

func TestServeReturnsShutdownError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	listening := make(chan struct{})
	expectedErr := errors.New("shutdown failed")

	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(
			ctx,
			time.Second,
			slog.Default(),
			func() error {
				close(listening)
				<-ctx.Done()
				return http.ErrServerClosed
			},
			func(context.Context) error {
				return expectedErr
			},
		)
	}()

	<-listening
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected shutdown error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("expected serve to return")
	}
}
