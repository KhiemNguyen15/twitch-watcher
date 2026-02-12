FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY . .

RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /bin/stream-filter \
    github.com/khiemnguyen15/twitch-watcher/services/stream-filter/cmd

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/stream-filter /stream-filter
ENTRYPOINT ["/stream-filter"]
