package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

	var config config

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

	go simulateRequests(context.Background(), &config)

	http.HandleFunc("/-/health", healthHandler)
	http.HandleFunc("/-/config/max-duration", setConfigHandler(config.SetMaxDuration))
	http.HandleFunc("/-/config/errors-percentage", setConfigHandler(config.SetErrorsPercentage))
	http.HandleFunc("/-/config/request-rate", setConfigHandler(config.SetRequestRate))
	http.Handle("/metrics", promhttp.Handler())

	return http.ListenAndServe(addr, nil)
}

func simulateRequests(ctx context.Context, config *config) error {
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

		requestDuration.Observe(float64(rand.Intn(config.MaxDuration())))

		// Simulate the failure of a certain percentage of the requests.

		if rand.Intn(100) < config.ErrorsPercentage() {
			requestErrorsCount.Inc()
		}

		// Simulate the configured request rate.

		select {
		case <-time.After(time.Duration(float64(time.Second) / float64(config.RequestRate()))):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "OK")
}

func setConfigHandler(set func(int) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		value, err := strconv.Atoi(string(data))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := set(value); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "OK")
	}
}

type config struct {
	mu               sync.RWMutex
	maxDuration      int
	errorsPercentage int
	requestRate      int
}

func (c *config) MaxDuration() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxDuration
}

func (c *config) SetMaxDuration(maxDuration int) error {
	if maxDuration < 0 {
		return fmt.Errorf("value is less than zero")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxDuration = maxDuration
	return nil
}

func (c *config) ErrorsPercentage() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.errorsPercentage
}

func (c *config) SetErrorsPercentage(errorsPercentage int) error {
	if errorsPercentage < 0 || errorsPercentage > 100 {
		return fmt.Errorf("value is not a valid percentage")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorsPercentage = errorsPercentage
	return nil
}

func (c *config) RequestRate() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requestRate
}

func (c *config) SetRequestRate(requestRate int) error {
	if requestRate <= 0 {
		return fmt.Errorf("value is less than or equal to zeros")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestRate = requestRate
	return nil
}
