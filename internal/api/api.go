package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/francescomari/metrics-generator/internal/limits"
)

type Server struct {
	Addr    string
	Config  *limits.Config
	Metrics http.Handler
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/-/health", s.handleHealth)
	mux.HandleFunc("/-/config/duration", s.handleDuration)
	mux.HandleFunc("/-/config/errors-percentage", s.handleErrorsPercentage)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/", s.handleNotFound)

	server := http.Server{
		Addr:    s.Addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("shutdown server: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %v", err)
	}

	return ctx.Err()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprintln(w, "OK")
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	httpError(w, http.StatusNotFound)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.Metrics.ServeHTTP(w, r)
}

func (s *Server) handleDuration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, http.StatusInternalServerError)
		return
	}

	parts := strings.Split(string(data), ",")

	if len(parts) != 2 {
		httpError(w, http.StatusBadRequest)
		return
	}

	min, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	max, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := s.Config.SetDurationInterval(min, max); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "OK")
}

func (s *Server) handleErrorsPercentage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, http.StatusInternalServerError)
		return
	}

	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := s.Config.SetErrorsPercentage(value); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "OK")
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
