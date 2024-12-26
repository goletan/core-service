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
RUN CGO_ENABLED=0 GOOS=linux go build -o core core-service/cmd/core/main.go

# Runtime Stage
FROM gcr.io/distroless/static:nonroot AS runtime
WORKDIR /app

COPY --from=builder /app/core /app/core
COPY core-service/config /app/config

USER nonroot:nonroot

EXPOSE 8081 2112

ENTRYPOINT ["/app/core"]
