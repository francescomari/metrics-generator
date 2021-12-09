# Metrics Generator

Metrics Generator pretends to continuously receive requests with a fixed rate of
1 request/sec and exposes two metrics related to these requests:

- `metrics_generator_request_duration_seconds` - histogram - The duration of the
  requests, in seconds.
- `metrics_generator_request_errors_count` - counter - The number of requests
  resulting in an error.

## CLI

Metrics Generator accepts flags to initialize the rate of the simulated request,
the maximum request duration and the percentage of requests that will result in
an error. Use the `-help` flag to see the command's help.

## API

Metrics Generator exposes a minimal API for reporting its health and for
changing at runtime the behaviour around the simulated requests.

```
GET /-/health
```

Always return a 200 response.

```
PUT /-/config/max-duration
```

Set the maximum duration of the simulated requests to the value passed in the
body of the request. The value must be a  positive integer.

```
PUT /-/config/errors-percentage
```

Set the percentage of the simulated requests that will result in an error to the
value passed in the body of the request. It must be an integer between 0 and
100.
