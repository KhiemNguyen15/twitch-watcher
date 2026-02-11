FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY services/stream-poller/go.mod services/stream-poller/go.sum* ./services/stream-poller/

RUN GOEXPERIMENT=jsonv2 go work sync

COPY pkg/ ./pkg/
COPY services/stream-poller/ ./services/stream-poller/

RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /bin/stream-poller \
    ./services/stream-poller/cmd

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/stream-poller /stream-poller
ENTRYPOINT ["/stream-poller"]
