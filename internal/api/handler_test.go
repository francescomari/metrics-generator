package api_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/francescomari/metrics-generator/internal/api"
	"github.com/google/go-cmp/cmp"
)

type mockConfig struct {
	doDurationInterval    func() (int, int)
	doSetDurationInterval func(min, max int) error
	doErrorsPercentage    func() int
	doSetErrorsPercentage func(value int) error
}

func (c mockConfig) DurationInterval() (int, int) {
	return c.doDurationInterval()
}

func (c mockConfig) SetDurationInterval(min, max int) error {
	return c.doSetDurationInterval(min, max)
}

func (c mockConfig) ErrorsPercentage() int {
	return c.doErrorsPercentage()
}

func (c mockConfig) SetErrorsPercentage(value int) error {
	return c.doSetErrorsPercentage(value)
}
func TestHandlerRoot(t *testing.T) {
	config := mockConfig{
		doDurationInterval: func() (int, int) {
			return 2, 4
		},
		doErrorsPercentage: func() int {
			return 10
		},
	}

	response := doIndexRequest(handlerForConfig(config))
	checkStatusCode(t, response, http.StatusOK)

	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	want := "Request time duration: 2s - 4s"
	if !strings.Contains(string(data), want) {
		t.Errorf("index page does not contain expected string:%s", want)
	}
}

func TestHandlerHealth(t *testing.T) {
	handler := api.Handler{}

	response := doHealthRequest(&handler)

	checkStatusCode(t, response, http.StatusOK)
	checkBody(t, response, "OK\n")
}

func TestHandlerGetDurationInterval(t *testing.T) {
	config := mockConfig{
		doDurationInterval: func() (int, int) {
			return 12, 34
		},
	}

	response := doGetDurationIntervalRequest(handlerForConfig(config))

	checkStatusCode(t, response, http.StatusOK)
	checkBody(t, response, "12,34\n")
}

func TestHandlerSetDurationInterval(t *testing.T) {
	var minDuration, maxDuration int

	config := mockConfig{
		doSetDurationInterval: func(min, max int) error {
			minDuration = min
			maxDuration = max
			return nil
		},
	}

	response := doSetDurationIntervalRequest(handlerForConfig(config), strings.NewReader("12,34"))

	checkStatusCode(t, response, http.StatusOK)
	checkBody(t, response, "OK\n")
	checkIntEqual(t, "minimum duration", minDuration, 12)
	checkIntEqual(t, "maximum duration", maxDuration, 34)
}

func TestHandlerSetDurationIntervalInvalid(t *testing.T) {
	handler := api.Handler{}

	response := doSetDurationIntervalRequest(&handler, strings.NewReader("boom"))

	checkStatusCode(t, response, http.StatusBadRequest)
}

func TestHandlerSetDurationIntervalReadError(t *testing.T) {
	handler := api.Handler{}

	response := doSetDurationIntervalRequest(&handler, iotest.ErrReader(errors.New("error")))

	checkStatusCode(t, response, http.StatusInternalServerError)
}

func TestHandlerSetDurationIntervalConfigError(t *testing.T) {
	config := mockConfig{
		doSetDurationInterval: func(min, max int) error {
			return errors.New("error")
		},
	}

	response := doSetDurationIntervalRequest(handlerForConfig(config), strings.NewReader("12,34"))

	checkStatusCode(t, response, http.StatusBadRequest)
}

func TestHandlerGetErrorsPercentage(t *testing.T) {
	config := mockConfig{
		doErrorsPercentage: func() int {
			return 12
		},
	}

	response := doGetErrorsPercentageRequest(handlerForConfig(config))

	checkStatusCode(t, response, http.StatusOK)
	checkBody(t, response, "12\n")
}

func TestHandlerSetErrorsPercentage(t *testing.T) {
	var errorsPercentage int

	config := mockConfig{
		doSetErrorsPercentage: func(value int) error {
			errorsPercentage = value
			return nil
		},
	}

	response := doSetErrorsPercentageRequest(handlerForConfig(config), strings.NewReader("12"))

	checkStatusCode(t, response, http.StatusOK)
	checkBody(t, response, "OK\n")
	checkIntEqual(t, "errors percentage", errorsPercentage, 12)
}

func TestHandlerSetErrorsPercentageInvalid(t *testing.T) {
	handler := api.Handler{}

	response := doSetErrorsPercentageRequest(&handler, strings.NewReader("boom"))

	checkStatusCode(t, response, http.StatusBadRequest)
}

func TestHandlerSetErrorsPercentageReadError(t *testing.T) {
	handler := api.Handler{}

	response := doSetErrorsPercentageRequest(&handler, iotest.ErrReader(errors.New("error")))

	checkStatusCode(t, response, http.StatusInternalServerError)
}

func TestHandlerSetErrorsPercentageConfigError(t *testing.T) {
	config := mockConfig{
		doSetErrorsPercentage: func(value int) error {
			return errors.New("error")
		},
	}

	response := doSetErrorsPercentageRequest(handlerForConfig(config), strings.NewReader("12"))

	checkStatusCode(t, response, http.StatusBadRequest)
}

func handlerForConfig(config api.Config) http.Handler {
	return &api.Handler{
		Config: config,
	}
}

func doGetDurationIntervalRequest(handler http.Handler) *http.Response {
	return doRequest(handler, http.MethodGet, "/-/config/duration-interval")
}

func doSetDurationIntervalRequest(handler http.Handler, body io.Reader) *http.Response {
	return doRequestWithBody(handler, http.MethodPut, "/-/config/duration-interval", body)
}

func doGetErrorsPercentageRequest(handler http.Handler) *http.Response {
	return doRequest(handler, http.MethodGet, "/-/config/errors-percentage")
}

func doSetErrorsPercentageRequest(handler http.Handler, body io.Reader) *http.Response {
	return doRequestWithBody(handler, http.MethodPut, "/-/config/errors-percentage", body)
}

func doIndexRequest(handler http.Handler) *http.Response {
	return doRequest(handler, http.MethodGet, "/")
}

func doHealthRequest(handler http.Handler) *http.Response {
	return doRequest(handler, http.MethodGet, "/-/health")
}

func doRequest(handler http.Handler, method string, path string) *http.Response {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))
	return recorder.Result()
}

func doRequestWithBody(handler http.Handler, method string, path string, body io.Reader) *http.Response {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(method, path, body))
	return recorder.Result()
}

func checkStatusCode(t *testing.T, response *http.Response, wanted int) {
	t.Helper()

	if got := response.StatusCode; got != wanted {
		t.Fatalf("invalid status code: wanted %d, got %d", wanted, got)
	}
}

func checkBody(t *testing.T, response *http.Response, wanted string) {
	t.Helper()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if diff := cmp.Diff(string(data), wanted); diff != "" {
		t.Fatalf("invalid body:\n%s", diff)
	}
}

func checkIntEqual(t *testing.T, name string, got, wanted int) {
	t.Helper()

	if got != wanted {
		t.Fatalf("invalid %s: wanted %d, got %d", name, wanted, got)
	}
}
