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
		maxDuration      int
		errorsPercentage int
		requestRate      int
	)

	flag.StringVar(&addr, "addr", ":8080", "The address to listen to")
	flag.IntVar(&maxDuration, "max-duration", 10, "Max duration of the simulated requests")
	flag.IntVar(&errorsPercentage, "errors-percentage", 10, "Which percentage of the requests will fail")
	flag.IntVar(&requestRate, "request-rate", 1, "How many requests per seconds to simulate")
	flag.Parse()

	var config limits.Config

	if err := config.SetMaxDuration(maxDuration); err != nil {
		return fmt.Errorf("set max duration: %v", err)
	}

	if err := config.SetErrorsPercentage(errorsPercentage); err != nil {
		return fmt.Errorf("set errors percentage: %v", err)
	}

	if err := config.SetRequestRate(requestRate); err != nil {
		return fmt.Errorf("set request rate: %v", err)
	}

	log.Printf("using max duration %v", maxDuration)
	log.Printf("using errors percentage %v", errorsPercentage)
	log.Printf("using request rate %v", requestRate)

	ctx, cancel := contextWithSignal(context.Background(), os.Interrupt)
	defer cancel()

	simulator := metrics.Generator{
		Config:   &config,
		Duration: requestDuration,
		Errors:   requestErrorsCount,
	}

	go simulator.Run(ctx)

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
