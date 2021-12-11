package api

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Config interface {
	DurationInterval() (int, int)
	SetDurationInterval(min, max int) error
	ErrorsPercentage() int
	SetErrorsPercentage(value int) error
}

type Handler struct {
	Config  Config
	Metrics http.Handler

	once    sync.Once
	handler http.Handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.once.Do(h.setupHandlers)
	h.handler.ServeHTTP(w, r)
}

func (h *Handler) setupHandlers() {
	router := mux.NewRouter()

	h.setupHealthHandler(router)
	h.setupDurationIntervalHandlers(router)
	h.setupErrorsPercentageHandlers(router)
	h.setupMetricsHandler(router)

	h.handler = router
}

func (h *Handler) setupHealthHandler(router *mux.Router) {
	router.
		Methods(http.MethodGet).
		Path("/-/health").
		HandlerFunc(h.handleHealth)
}

func (h *Handler) setupDurationIntervalHandlers(router *mux.Router) {
	sub := router.
		PathPrefix("/-/config/duration").
		Subrouter()

	sub.
		Methods(http.MethodGet).
		HandlerFunc(h.handleGetDurationInterval)

	sub.
		Methods(http.MethodPut).
		HandlerFunc(h.handleSetDurationInterval)
}

func (h *Handler) setupErrorsPercentageHandlers(router *mux.Router) {
	sub := router.
		PathPrefix("/-/config/errors-percentage").
		Subrouter()

	sub.
		Methods(http.MethodGet).
		HandlerFunc(h.handleGetErrorsPercentage)

	sub.
		Methods(http.MethodPut).
		HandlerFunc(h.handleSetErrorsPercentage)
}

func (h *Handler) setupMetricsHandler(router *mux.Router) {
	router.
		Methods(http.MethodGet).
		Path("/metrics").
		Handler(h.Metrics)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func (h *Handler) handleGetDurationInterval(w http.ResponseWriter, r *http.Request) {
	min, max := h.Config.DurationInterval()
	fmt.Fprintf(w, "%d,%d\n", min, max)
}

func (h *Handler) handleSetDurationInterval(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, http.StatusInternalServerError)
		return
	}

	min, max, err := parseDuration(string(data))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := h.Config.SetDurationInterval(min, max); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "OK")
}

func (h *Handler) handleGetErrorsPercentage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d\n", h.Config.ErrorsPercentage())
}

func (h *Handler) handleSetErrorsPercentage(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, http.StatusInternalServerError)
		return
	}

	value, err := parseInt(string(data))
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := h.Config.SetErrorsPercentage(value); err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	fmt.Fprintln(w, "OK")
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
