package httprun_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/francescomari/metrics-generator/internal/httprun"
)

var (
	errListenAndServe = errors.New("listen and serve")
	errShutdown       = errors.New("shutdown")
)

type mockServer struct {
	doListenAndServe func() error
	doShutdown       func(ctx context.Context) error
}

func (s mockServer) ListenAndServe() error {
	return s.doListenAndServe()
}

func (s mockServer) Shutdown(ctx context.Context) error {
	return s.doShutdown(ctx)
}

func TestListenAndServe(t *testing.T) {
	var (
		listenAndServeCalled = make(chan struct{})
		shutdownCalled       = make(chan struct{})
	)

	server := mockServer{
		doListenAndServe: func() error {
			close(listenAndServeCalled)
			<-shutdownCalled
			return nil
		},
		doShutdown: func(context.Context) error {
			close(shutdownCalled)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-listenAndServeCalled
		cancel()
	}()

	errs := httprun.ListenAndServe(ctx, server, time.Second)

	checkErrorsLength(t, errs, 0)
}

func TestListenAndServeErrors(t *testing.T) {
	var (
		listenAndServeCalled = make(chan struct{})
		shutdownCalled       = make(chan struct{})
	)

	server := mockServer{
		doListenAndServe: func() error {
			close(listenAndServeCalled)
			<-shutdownCalled
			return errListenAndServe
		},
		doShutdown: func(context.Context) error {
			close(shutdownCalled)
			return errShutdown
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-listenAndServeCalled
		cancel()
	}()

	errs := httprun.ListenAndServe(ctx, server, time.Second)

	checkErrorsLength(t, errs, 2)
	checkErrorsContain(t, errs, errListenAndServe)
	checkErrorsContain(t, errs, errShutdown)
}

func TestListenAndServeSetupError(t *testing.T) {
	server := mockServer{
		doListenAndServe: func() error {
			return errListenAndServe
		},
		doShutdown: func(context.Context) error {
			panic("shutdown should not be called")
		},
	}

	errs := httprun.ListenAndServe(context.Background(), server, time.Second)

	checkErrorsLength(t, errs, 1)
	checkErrorsContain(t, errs, errListenAndServe)
}

func checkErrorsLength(t *testing.T, errors []error, length int) {
	t.Helper()

	if len(errors) != length {
		t.Fatalf("expected %d errors, got %d", len(errors), length)
	}
}

func checkErrorsContain(t *testing.T, errors []error, target error) {
	t.Helper()

	for _, err := range errors {
		if err == target {
			return
		}
	}

	t.Fatalf("error '%v' not found", target)
}
