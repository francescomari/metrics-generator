package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/francescomari/metrics-generator/internal/limits"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	Addr   string
	Config *limits.Config
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/-/health", healthHandler)
	mux.HandleFunc("/-/config/max-duration", setConfigHandler(s.Config.SetMaxDuration))
	mux.HandleFunc("/-/config/errors-percentage", setConfigHandler(s.Config.SetErrorsPercentage))
	mux.HandleFunc("/-/config/request-rate", setConfigHandler(s.Config.SetRequestRate))

	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{
		Addr:    s.Addr,
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
