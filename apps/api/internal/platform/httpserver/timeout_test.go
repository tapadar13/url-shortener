package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutAddsRequestDeadline(t *testing.T) {
	t.Parallel()

	const duration = 2 * time.Second
	var deadline time.Time
	var hasDeadline bool

	handler := Timeout(duration)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		deadline, hasDeadline = r.Context().Deadline()
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !hasDeadline {
		t.Fatal("expected request context to have a deadline")
	}

	remaining := time.Until(deadline)
	if remaining <= time.Second || remaining > duration {
		t.Fatalf("expected deadline within %s, got %s remaining", duration, remaining)
	}
}

func TestTimeoutCancelsContextAfterHandlerReturns(t *testing.T) {
	t.Parallel()

	var requestContext context.Context
	handler := Timeout(time.Second)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		requestContext = r.Context()
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	select {
	case <-requestContext.Done():
		if requestContext.Err() != context.Canceled {
			t.Fatalf("expected canceled context, got %v", requestContext.Err())
		}
	default:
		t.Fatal("expected request context to be canceled")
	}
}

func TestTimeoutLeavesHandlerUnchangedForNonPositiveDuration(t *testing.T) {
	t.Parallel()

	called := false
	handler := Timeout(0)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := r.Context().Deadline(); ok {
			t.Fatal("expected request context to have no deadline")
		}
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !called {
		t.Fatal("expected handler to be called")
	}
}
