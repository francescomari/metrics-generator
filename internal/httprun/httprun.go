package httprun

import (
	"context"
	"time"
)

// Server is an HTTP server that can be started and shut down with the functions
// in this package. Server mimicks the interface of http.Server.
type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// ListenAndServe starts the provided server using the ListenAndServe method.
// The server will be shut down when the provided context is cancelled. When
// shutting the server, ListenAndServe calls the Shutdown method of the server
// with a context that will be cancelled after shutdownTimeout. Any error
// returned by the server will be returned to the caller. ListenAndServe returns
// only when the server is fully shut down.
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
