# Multi-stage Dockerfile for SLURM Prometheus Exporter
# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Build arguments for version info
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG COMMIT=unknown
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETOS

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

# Build the application with version info and platform-specific optimizations
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" \
    -a -installsuffix cgo \
    -trimpath \
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