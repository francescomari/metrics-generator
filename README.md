# Metrics Generator

Metrics Generator pretends to continuously receive requests with a fixed rate of
1 request/sec and exposes two metrics related to these requests:

- `metrics_generator_request_duration_seconds` - histogram - The duration of the
  requests, in seconds.
- `metrics_generator_request_errors_count` - counter - The number of requests
  resulting in an error.

## CLI

Metrics Generator accepts flags to initialize the minimum and maximum request
duration and the percentage of requests that will result in an error. Use the
`-help` flag to see the command's help.

## API

Metrics Generator exposes a minimal API for reporting its health and for
changing at runtime the behaviour of the simulated requests.

```
GET /-/health
```

Always return a 200 response.

```
GET /-/config/duration-interval
```

Returns the current minimum and maximum values for the duration interval in the
form `min,max`.

```
PUT /-/config/duration-interval
```

Set the minimum and maximum value for the simulated duration to the values
passed in the body of the request. The body must be in the form `min,max`. Both
the minimum and the maximum must be numbers greater than zero. The minimum must
be less than the maximum.

```
GET /-/config/errors-percentage
```

Returns the current errors percentage.

```
PUT /-/config/errors-percentage
```

Set the percentage of the simulated requests that will result in an error to the
value passed in the body of the request. It must be an integer between 0 and
100.

### Examples

Read the current duration interval:

```
curl http://localhost:8080/-/config/duration-interval
```

Simulate the duration to be a random number between 15s and 45s:

```
curl -X PUT http://localhost:8080/-/config/duration-interval -d 15,45
```

Simulate the duration to be exactly 10s:

```
curl -X PUT http://localhost:8080/-/config/duration-interval -d 10,10
```

Read the current errors percentage:

```
curl http://localhost:8080/-/config/errors-percentage
```

Simulate the error rate to be 25%:

```
curl -X PUT http://localhost:8080/-/config/errors-percentage -d 25
```
