# Use the latest Go base image
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Copy go.work and all referenced directories
COPY go.work .
COPY core ./core
COPY config ./config
COPY logger ./logger
COPY observability ./observability
COPY resilience ./resilience
COPY security ./security
COPY services ./services

# Sync Go work dependencies
RUN go work sync

# Build the core-service service
RUN go build -o core-service core-service/cmd/core-service/main.go

# Use a lightweight base image for the runtime
FROM gcr.io/distroless/base-debian11

# Copy the built binary
COPY --from=builder /app/core /core

# Set the entry point
ENTRYPOINT ["/core"]
