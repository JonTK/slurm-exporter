# Multi-stage Dockerfile for SLURM Prometheus Exporter
# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy Go module files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking and optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o slurm-exporter \
    ./cmd/slurm-exporter

# Stage 2: Runtime stage
FROM scratch AS runtime

# Copy CA certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data from builder stage
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary from builder stage
COPY --from=builder /build/slurm-exporter /usr/local/bin/slurm-exporter

# Copy example configuration
COPY --from=builder /build/configs/config.yaml /etc/slurm-exporter/config.yaml

# Create a non-root user (note: we can't use adduser in scratch)
# The application should handle running as a specific UID/GID
USER 65534:65534

# Expose default port
EXPOSE 8080

# Set environment variables
ENV SLURM_EXPORTER_CONFIG_FILE=/etc/slurm-exporter/config.yaml

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/slurm-exporter", "--health-check"]

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/slurm-exporter"]

# Default command arguments
CMD ["--config", "/etc/slurm-exporter/config.yaml"]