package api

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Config interface {
	SetDurationInterval(min, max int) error
	SetErrorsPercentage(value int) error
}

type Handler struct {
	Config  Config
	Metrics http.Handler

	once    sync.Once
	handler http.Handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.once.Do(func() {
		router := mux.NewRouter()

		router.
			Methods(http.MethodGet).
			Path("/-/health").
			HandlerFunc(h.handleHealth)

		router.
			Methods(http.MethodPut).
			Path("/-/config/duration").
			HandlerFunc(h.handleDuration)

		router.
			Methods(http.MethodPut).
			Path("/-/config/errors-percentage").
			HandlerFunc(h.handleErrorsPercentage)

		router.
			Methods(http.MethodGet).
			Path("/metrics").
			Handler(h.Metrics)

		h.handler = router
	})

	h.handler.ServeHTTP(w, r)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func (h *Handler) handleDuration(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) handleErrorsPercentage(w http.ResponseWriter, r *http.Request) {
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
