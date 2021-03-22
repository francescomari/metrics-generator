package metrics

import (
	"context"
	"math/rand"
	"time"

	"github.com/francescomari/metrics-generator/internal/limits"
)

type Histogram interface {
	Observe(float64)
}

type Counter interface {
	Inc()
}

type Generator struct {
	Config   *limits.Config
	Duration Histogram
	Errors   Counter
}

func (g *Generator) Run(ctx context.Context) error {
	for {

		// Observe a request that took a random amount of time between (0,
		// N) seconds. The default for N is 10s, which fits the highest
		// bucket defined by default by a Prometheus histogram.

		g.Duration.Observe(float64(rand.Intn(g.Config.MaxDuration())))

		// Simulate the failure of a certain percentage of the requests.

		if rand.Intn(100) < g.Config.ErrorsPercentage() {
			g.Errors.Inc()
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
