package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/francescomari/metrics-generator/internal/api"
	"github.com/francescomari/metrics-generator/internal/limits"
	"github.com/francescomari/metrics-generator/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var requestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
	Name: "metrics_generator_request_duration_seconds",
	Help: "Request duration in seconds",
})

var requestErrorsCount = promauto.NewCounter(prometheus.CounterOpts{
	Name: "metrics_generator_request_errors_count",
	Help: "Number of errors observed in requests",
})

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	rand.Seed(time.Now().Unix())

	var (
		addr             string
		minDuration      int
		maxDuration      int
		errorsPercentage int
	)

	flag.StringVar(&addr, "addr", ":8080", "The address to listen to")
	flag.IntVar(&minDuration, "min-duration", 1, "Minimum request duration")
	flag.IntVar(&maxDuration, "max-duration", 10, "Maximum request duration")
	flag.IntVar(&errorsPercentage, "errors-percentage", 10, "Which percentage of the requests will fail")
	flag.Parse()

	var config limits.Config

	if err := config.SetDurationInterval(minDuration, maxDuration); err != nil {
		return fmt.Errorf("set max duration: %v", err)
	}

	if err := config.SetErrorsPercentage(errorsPercentage); err != nil {
		return fmt.Errorf("set errors percentage: %v", err)
	}

	log.Printf("using duration %v,%v", minDuration, maxDuration)
	log.Printf("using errors percentage %v", errorsPercentage)

	ctx, cancel := contextWithSignal(context.Background(), os.Interrupt)
	defer cancel()

	generator := metrics.Generator{
		Config:   &config,
		Duration: requestDuration,
		Errors:   requestErrorsCount,
	}

	go func() {
		if err := generator.Run(ctx); err != nil && err != context.Canceled {
			log.Printf("error: run simulator: %v", err)
		}
	}()

	server := api.Server{
		Addr:    addr,
		Config:  &config,
		Metrics: promhttp.Handler(),
	}

	return server.Run(ctx)
}

func contextWithSignal(parent context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	ctx, cancel := context.WithCancel(parent)

	go func() {
		defer cancel()

		select {
		case <-parent.Done():
			// Return if the parent context is cancelled.
		case <-ch:
			// Return if notified by a signal.
		}
	}()

	return ctx, cancel
}
