package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
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

	ctx, cancel := contextWithSignal(context.Background(), os.Interrupt)
	defer cancel()

	go simulateRequests(ctx, &config)

	mux := http.NewServeMux()

	mux.HandleFunc("/-/health", healthHandler)
	mux.HandleFunc("/-/config/max-duration", setConfigHandler(config.SetMaxDuration))
	mux.HandleFunc("/-/config/errors-percentage", setConfigHandler(config.SetErrorsPercentage))
	mux.HandleFunc("/-/config/request-rate", setConfigHandler(config.SetRequestRate))
	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("error: shutdown server: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %v", err)
	}

	return nil
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

type config struct {
	maxDuration      int64
	errorsPercentage int64
	requestRate      int64
}

func (c *config) MaxDuration() int {
	return int(atomic.LoadInt64(&c.maxDuration))
}

func (c *config) SetMaxDuration(maxDuration int) error {
	if maxDuration < 0 {
		return fmt.Errorf("value is less than zero")
	}

	atomic.StoreInt64(&c.maxDuration, int64(maxDuration))

	return nil
}

func (c *config) ErrorsPercentage() int {
	return int(atomic.LoadInt64(&c.errorsPercentage))
}

func (c *config) SetErrorsPercentage(errorsPercentage int) error {
	if errorsPercentage < 0 || errorsPercentage > 100 {
		return fmt.Errorf("value is not a valid percentage")
	}

	atomic.StoreInt64(&c.errorsPercentage, int64(errorsPercentage))

	return nil
}

func (c *config) RequestRate() int {
	return int(atomic.LoadInt64(&c.requestRate))
}

func (c *config) SetRequestRate(requestRate int) error {
	if requestRate <= 0 {
		return fmt.Errorf("value is less than or equal to zeros")
	}

	atomic.StoreInt64(&c.requestRate, int64(requestRate))

	return nil
}
