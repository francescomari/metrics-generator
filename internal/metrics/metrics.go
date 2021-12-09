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
		g.Duration.Observe(g.randomDuration())

		if g.shouldFailRequest() {
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

func (g *Generator) shouldFailRequest() bool {
	return rand.Intn(100) < g.Config.ErrorsPercentage()
}

func (g *Generator) randomDuration() float64 {
	return float64(randomNumberBetween(g.Config.DurationInterval()))
}

func randomNumberBetween(min, max int) int {
	return min + rand.Intn(max-min)
}
