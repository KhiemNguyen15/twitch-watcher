FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy workspace and module files
COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY services/subscription-service/go.mod services/subscription-service/go.sum* ./services/subscription-service/

# Download dependencies
RUN GOEXPERIMENT=jsonv2 go work sync

# Copy source
COPY pkg/ ./pkg/
COPY services/subscription-service/ ./services/subscription-service/

# Build
RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /bin/subscription-service \
    ./services/subscription-service/cmd

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/subscription-service /subscription-service
EXPOSE 8080
ENTRYPOINT ["/subscription-service"]
