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
	"golang.org/x/sync/errgroup"
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

	var g metricsGenerator

	flag.StringVar(&g.address, "addr", ":8080", "The address to listen to")
	flag.IntVar(&g.minDuration, "min-duration", 1, "Minimum request duration")
	flag.IntVar(&g.maxDuration, "max-duration", 10, "Maximum request duration")
	flag.IntVar(&g.errorsPercentage, "errors-percentage", 10, "Which percentage of the requests will fail")
	flag.Parse()

	return g.run()
}

type metricsGenerator struct {
	address          string
	minDuration      int
	maxDuration      int
	errorsPercentage int
}

func (g *metricsGenerator) run() error {
	config, err := g.buildLimitsConfig()
	if err != nil {
		return fmt.Errorf("build limits configuration: %v", err)
	}

	generator := metrics.Generator{
		Config:   config,
		Duration: requestDuration,
		Errors:   requestErrorsCount,
	}

	server := api.Server{
		Addr:    g.address,
		Config:  config,
		Metrics: promhttp.Handler(),
	}

	ctx, cancel := contextWithSignal(context.Background(), os.Interrupt)
	defer cancel()

	var group errgroup.Group

	group.Go(func() error {
		return generator.Run(ctx)
	})

	group.Go(func() error {
		return server.Run(ctx)
	})

	return group.Wait()
}

func (g *metricsGenerator) buildLimitsConfig() (*limits.Config, error) {
	var config limits.Config

	if err := config.SetDurationInterval(g.minDuration, g.maxDuration); err != nil {
		return nil, fmt.Errorf("set max duration: %v", err)
	}

	if err := config.SetErrorsPercentage(g.errorsPercentage); err != nil {
		return nil, fmt.Errorf("set errors percentage: %v", err)
	}

	return &config, nil
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
