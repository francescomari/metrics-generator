package api

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

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

func TestHandlerHealth(t *testing.T) {
	handler := Handler{}

	response := doRequest(&handler, http.MethodGet, "/-/health")

	checkStatusCode(t, response, http.StatusOK)
	checkBody(t, response, "OK\n")
}

func TestHandlerGetDurationInterval(t *testing.T) {
	config := mockConfig{
		doDurationInterval: func() (int, int) {
			return 12, 34
		},
	}

	handler := Handler{
		Config: config,
	}

	response := doRequest(&handler, http.MethodGet, "/-/config/duration-interval")

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

	handler := Handler{
		Config: config,
	}

	response := doRequestWithBody(&handler, http.MethodPut, "/-/config/duration", strings.NewReader("12,34"))

	checkStatusCode(t, response, http.StatusOK)
	checkBody(t, response, "OK\n")

	if minDuration != 12 {
		t.Fatalf("invalid minimum duration interval: %v", minDuration)
	}

	if maxDuration != 34 {
		t.Fatalf("invalid maximum duration interval: %v", maxDuration)
	}
}

func TestHandlerSetDurationIntervalInvalid(t *testing.T) {
	handler := Handler{}

	response := doRequestWithBody(&handler, http.MethodPut, "/-/config/duration", strings.NewReader("boom"))

	checkStatusCode(t, response, http.StatusBadRequest)
}

func TestHandlerSetDurationIntervalReadError(t *testing.T) {
	handler := Handler{}

	response := doRequestWithBody(&handler, http.MethodPut, "/-/config/duration", iotest.ErrReader(errors.New("error")))

	checkStatusCode(t, response, http.StatusInternalServerError)
}

func TestHandlerSetDurationIntervalConfigError(t *testing.T) {
	config := mockConfig{
		doSetDurationInterval: func(min, max int) error {
			return errors.New("error")
		},
	}

	handler := Handler{
		Config: config,
	}

	response := doRequestWithBody(&handler, http.MethodPut, "/-/config/duration", strings.NewReader("12,34"))

	checkStatusCode(t, response, http.StatusBadRequest)
}

func TestHandlerGetErrorsPercentage(t *testing.T) {
	config := mockConfig{
		doErrorsPercentage: func() int {
			return 12
		},
	}

	handler := Handler{
		Config: config,
	}

	response := doRequest(&handler, http.MethodGet, "/-/config/errors-percentage")

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

	handler := Handler{
		Config: config,
	}

	response := doRequestWithBody(&handler, http.MethodPut, "/-/config/errors-percentage", strings.NewReader("12"))

	checkStatusCode(t, response, http.StatusOK)
	checkBody(t, response, "OK\n")

	if errorsPercentage != 12 {
		t.Fatalf("invalid errors percentage: %v", errorsPercentage)
	}
}

func TestHandlerSetErrorsPercentageInvalid(t *testing.T) {
	handler := Handler{}

	response := doRequestWithBody(&handler, http.MethodPut, "/-/config/errors-percentage", strings.NewReader("boom"))

	checkStatusCode(t, response, http.StatusBadRequest)
}

func TestHandlerSetErrorsPercentageReadError(t *testing.T) {
	handler := Handler{}

	response := doRequestWithBody(&handler, http.MethodPut, "/-/config/errors-percentage", iotest.ErrReader(errors.New("error")))

	checkStatusCode(t, response, http.StatusInternalServerError)
}

func TestHandlerSetErrorsPercentageConfigError(t *testing.T) {
	config := mockConfig{
		doSetErrorsPercentage: func(value int) error {
			return errors.New("error")
		},
	}

	handler := Handler{
		Config: config,
	}

	response := doRequestWithBody(&handler, http.MethodPut, "/-/config/errors-percentage", strings.NewReader("12"))

	checkStatusCode(t, response, http.StatusBadRequest)
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
