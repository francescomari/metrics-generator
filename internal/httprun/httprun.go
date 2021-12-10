package httprun

import (
	"context"
	"time"
)

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

func ListenAndServe(ctx context.Context, server Server, shutdownTimeout time.Duration) []error {
	var (
		errors       = make(chan error, 2)
		shutdownDone = make(chan struct{})
		serverDone   = make(chan struct{})
	)

	go func() {
		defer close(serverDone)

		errors <- server.ListenAndServe()
	}()

	go func() {
		defer close(shutdownDone)

		select {
		case <-serverDone:
			return
		case <-ctx.Done():
			errors <- shutdownGracefully(server, shutdownTimeout)
		}
	}()

	go func() {
		defer close(errors)

		<-shutdownDone
		<-serverDone
	}()

	var result []error

	for err := range errors {
		if err != nil {
			result = append(result, err)
		}
	}

	return result
}

func shutdownGracefully(server Server, shutdownTimeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return server.Shutdown(ctx)
}
