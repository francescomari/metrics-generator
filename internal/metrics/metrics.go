package metrics

import (
	"context"
	"math/rand"
	"time"

	"github.com/francescomari/metrics-generator/internal/limits"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Generator struct {
	Config *limits.Config
}

func (g *Generator) Run(ctx context.Context) error {
	requestDuration := promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "metrics_generator_request_duration_seconds",
		Help: "Request duration in seconds",
	})

	requestErrorsCount := promauto.NewCounter(prometheus.CounterOpts{
		Name: "metrics_generator_request_errors_count",
		Help: "Number of errors observed in requests",
	})

	for {

		// Observe a request that took a random amount of time between (0,
		// N) seconds. The default for N is 10s, which fits the highest
		// bucket defined by default by a Prometheus histogram.

		requestDuration.Observe(float64(rand.Intn(g.Config.MaxDuration())))

		// Simulate the failure of a certain percentage of the requests.

		if rand.Intn(100) < g.Config.ErrorsPercentage() {
			requestErrorsCount.Inc()
		}

		// Simulate the configured request rate.

		select {
		case <-time.After(time.Duration(float64(time.Second) / float64(g.Config.RequestRate()))):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
