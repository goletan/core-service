# Use the latest Go base image
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Copy go.work and all referenced directories
COPY go.work .
COPY kernel-service ./kernel-service
COPY config-library ./config-library
COPY logger-library ./logger-library
COPY observability-library ./observability-library
COPY resilience-library ./resilience-library
COPY security-library ./security-library
COPY services-library ./services-library

# Sync Go work dependencies
RUN go work sync

# Build the kernel-service service
RUN go build -o kernel-service kernel-service/cmd/kernel/main.go

# Use a lightweight base image for the runtime
FROM gcr.io/distroless/base-debian11

# Copy the built binary
COPY --from=builder /app/kernel /kernel

# Set the entry point
ENTRYPOINT ["/kernel"]
