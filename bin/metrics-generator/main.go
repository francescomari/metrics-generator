package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os/signal"
	"syscall"
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
		return err
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		if err := generator.Run(ctx); err != nil && err != context.Canceled {
			return fmt.Errorf("run generator: %v", err)
		}
		return nil
	})

	group.Go(func() error {
		if err := server.Run(ctx); err != nil && err != context.Canceled {
			return fmt.Errorf("run server: %v", err)
		}
		return nil
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
