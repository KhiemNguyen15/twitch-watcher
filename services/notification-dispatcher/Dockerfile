FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY services/notification-dispatcher/go.mod services/notification-dispatcher/go.sum* ./services/notification-dispatcher/

RUN GOEXPERIMENT=jsonv2 go work sync

COPY pkg/ ./pkg/
COPY services/notification-dispatcher/ ./services/notification-dispatcher/

RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /bin/notification-dispatcher \
    ./services/notification-dispatcher/cmd

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/notification-dispatcher /notification-dispatcher
ENTRYPOINT ["/notification-dispatcher"]
