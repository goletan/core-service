# Build Stage
FROM golang:1.23 AS builder
WORKDIR /app

COPY go.work .
COPY core-service ./core-service
COPY config-library ./config-library
COPY logger-library ./logger-library
COPY observability-library ./observability-library
COPY resilience-library ./resilience-library
COPY security-library ./security-library
COPY services-library ./services-library

RUN go work sync
RUN CGO_ENABLED=0 GOOS=linux go build -o core-service core-service/cmd/core/main.go

# Debug Stage
FROM debian:bullseye AS debug
WORKDIR /app
COPY --from=builder /app/core-service .
RUN chmod +x core-service

# Validate the binary in the debug-friendly container
RUN ./core-service --help || true

# Runtime Stage
FROM gcr.io/distroless/static:nonroot AS runtime
WORKDIR /app/
COPY --from=builder /app/core-service .
USER nonroot:nonroot
ENTRYPOINT ["./core-service"]
