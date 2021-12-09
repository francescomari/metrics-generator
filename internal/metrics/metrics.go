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
		g.Duration.Observe(float64(rand.Intn(g.Config.MaxDuration())))

		if rand.Intn(100) < g.Config.ErrorsPercentage() {
			g.Errors.Inc()
		}

		select {
		case <-time.After(1 * time.Second):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
