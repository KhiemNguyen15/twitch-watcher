FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY services/stream-filter/go.mod services/stream-filter/go.sum* ./services/stream-filter/

RUN GOEXPERIMENT=jsonv2 go work sync

COPY pkg/ ./pkg/
COPY services/stream-filter/ ./services/stream-filter/

RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /bin/stream-filter \
    ./services/stream-filter/cmd

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/stream-filter /stream-filter
ENTRYPOINT ["/stream-filter"]
