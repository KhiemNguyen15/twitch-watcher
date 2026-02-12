FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY . .

RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /bin/subscription-service \
    github.com/khiemnguyen15/twitch-watcher/services/subscription-service/cmd

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/subscription-service /subscription-service
EXPOSE 8080
ENTRYPOINT ["/subscription-service"]
