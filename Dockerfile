FROM golang:1.16.2 AS builder
WORKDIR /workdir
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 go install ./bin/metrics-generator

FROM alpine:3.13.2
COPY  --from=builder /go/bin/metrics-generator /bin/
ENTRYPOINT ["metrics-generator"]
