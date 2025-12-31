# SLURM Exporter Container Images

This repository provides multiple container images for the SLURM Exporter, each optimized for different use cases and security requirements.

## Available Images

### Standard Image (`slurm-exporter`)
- **Base**: `scratch`
- **Size**: ~15MB
- **Security**: Minimal attack surface
- **Use case**: Production environments where size and security are priorities

```bash
docker pull ghcr.io/jontk/slurm-exporter:latest
```

### Alpine Image (`slurm-exporter-alpine`)
- **Base**: `alpine:3.19`
- **Size**: ~25MB
- **Security**: Good security with debugging capabilities
- **Use case**: Development and environments requiring shell access

```bash
docker pull ghcr.io/jontk/slurm-exporter-alpine:latest
```

### Distroless Image (`slurm-exporter-distroless`)
- **Base**: `gcr.io/distroless/static-debian12:nonroot`
- **Size**: ~20MB
- **Security**: Maximum security, minimal attack surface
- **Use case**: High-security production environments

```bash
docker pull ghcr.io/jontk/slurm-exporter-distroless:latest
```

## Quick Start

### Basic Usage

```bash
# Run with default configuration
docker run -d \
  --name slurm-exporter \
  -p 8080:8080 \
  ghcr.io/jontk/slurm-exporter:latest
```

### With Custom Configuration

```bash
# Create configuration file
cat > config.yaml <<EOF
server:
  address: ":8080"
  metrics_path: "/metrics"

slurm:
  base_url: "http://your-slurm-server:6820"
  timeout: 30s
  auth:
    type: "jwt"
    jwt_path: "/var/run/slurm-jwt"

collectors:
  jobs:
    enabled: true
  nodes:
    enabled: true
  partitions:
    enabled: true
EOF

# Run with custom configuration
docker run -d \
  --name slurm-exporter \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/etc/slurm-exporter/config.yaml:ro \
  ghcr.io/jontk/slurm-exporter:latest
```

### With Environment Variables

```bash
docker run -d \
  --name slurm-exporter \
  -p 8080:8080 \
  -e SLURM_EXPORTER_SLURM_BASE_URL="http://your-slurm-server:6820" \
  -e SLURM_EXPORTER_SLURM_AUTH_TYPE="jwt" \
  -e SLURM_EXPORTER_SLURM_AUTH_JWT_PATH="/var/run/slurm-jwt" \
  ghcr.io/jontk/slurm-exporter:latest
```

### With Secrets (Kubernetes)

```bash
# Mount JWT token from Kubernetes secret
docker run -d \
  --name slurm-exporter \
  -p 8080:8080 \
  -v /path/to/jwt:/var/run/slurm-jwt:ro \
  -e SLURM_EXPORTER_SLURM_BASE_URL="http://your-slurm-server:6820" \
  -e SLURM_EXPORTER_SLURM_AUTH_TYPE="jwt" \
  -e SLURM_EXPORTER_SLURM_AUTH_JWT_PATH="/var/run/slurm-jwt" \
  ghcr.io/jontk/slurm-exporter:latest
```

## Docker Compose

```yaml
version: '3.8'

services:
  slurm-exporter:
    image: ghcr.io/jontk/slurm-exporter:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/etc/slurm-exporter/config.yaml:ro
      - ./jwt-token:/var/run/slurm-jwt:ro
    environment:
      - SLURM_EXPORTER_LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "/usr/local/bin/slurm-exporter", "--health-check"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 5s

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
    depends_on:
      - slurm-exporter
```

## Configuration

### Environment Variables

All configuration options can be set via environment variables using the prefix `SLURM_EXPORTER_`:

```bash
# Server configuration
SLURM_EXPORTER_SERVER_ADDRESS=":8080"
SLURM_EXPORTER_SERVER_METRICS_PATH="/metrics"

# SLURM connection
SLURM_EXPORTER_SLURM_BASE_URL="http://slurm-server:6820"
SLURM_EXPORTER_SLURM_TIMEOUT="30s"

# Authentication
SLURM_EXPORTER_SLURM_AUTH_TYPE="jwt"
SLURM_EXPORTER_SLURM_AUTH_JWT_PATH="/var/run/slurm-jwt"

# Logging
SLURM_EXPORTER_LOG_LEVEL="info"
SLURM_EXPORTER_LOG_FORMAT="json"

# Collectors
SLURM_EXPORTER_COLLECTORS_JOBS_ENABLED="true"
SLURM_EXPORTER_COLLECTORS_NODES_ENABLED="true"
SLURM_EXPORTER_COLLECTORS_PARTITIONS_ENABLED="true"
```

### Volume Mounts

- `/etc/slurm-exporter/config.yaml` - Configuration file
- `/var/run/slurm-jwt` - SLURM JWT token file
- `/var/lib/slurm-exporter` - Data directory (if needed)
- `/var/log/slurm-exporter` - Log directory (Alpine image only)

### Exposed Ports

- `8080` - Default HTTP port for metrics and API endpoints

## Health Checks

The container includes health check endpoints:

```bash
# Health check (all images)
curl http://localhost:8080/health

# Ready check
curl http://localhost:8080/ready

# Metrics endpoint
curl http://localhost:8080/metrics
```

## Security

### Image Signatures

All images are signed with Cosign for supply chain security:

```bash
# Verify image signature
cosign verify ghcr.io/jontk/slurm-exporter:latest \
  --certificate-identity-regexp="https://github.com/jontk/slurm-exporter" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com"
```

### Security Scanning

Images are scanned for vulnerabilities using Trivy. View scan results in the GitHub Security tab.

### Running as Non-Root

All images run as non-root users:
- Standard/Alpine: `slurm-exporter` user (UID: 65534)
- Distroless: `nonroot` user (UID: 65532)

### Recommended Security Practices

```bash
# Run with read-only root filesystem
docker run -d \
  --name slurm-exporter \
  --read-only \
  --tmpfs /tmp \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/etc/slurm-exporter/config.yaml:ro \
  ghcr.io/jontk/slurm-exporter:latest

# Run with limited capabilities
docker run -d \
  --name slurm-exporter \
  --cap-drop=ALL \
  --cap-add=NET_BIND_SERVICE \
  -p 8080:8080 \
  ghcr.io/jontk/slurm-exporter:latest

# Run with security options
docker run -d \
  --name slurm-exporter \
  --security-opt=no-new-privileges:true \
  --security-opt=seccomp=unconfined \
  -p 8080:8080 \
  ghcr.io/jontk/slurm-exporter:latest
```

## Troubleshooting

### Debug Mode (Alpine Image Only)

```bash
# Run with shell access for debugging
docker run -it --rm \
  ghcr.io/jontk/slurm-exporter-alpine:latest \
  /bin/sh

# Check configuration
docker exec -it slurm-exporter cat /etc/slurm-exporter/config.yaml

# View logs
docker logs slurm-exporter

# Test connectivity
docker exec -it slurm-exporter curl -f http://localhost:8080/health
```

### Common Issues

1. **Connection refused**: Check SLURM server URL and authentication
2. **Permission denied**: Verify JWT token file permissions and mount
3. **Metrics not appearing**: Check collector configuration and SLURM API access

### Performance Tuning

```bash
# Run with increased memory limits
docker run -d \
  --name slurm-exporter \
  --memory=512m \
  --memory-swap=1g \
  -p 8080:8080 \
  ghcr.io/jontk/slurm-exporter:latest

# Run with CPU limits
docker run -d \
  --name slurm-exporter \
  --cpus="1.0" \
  -p 8080:8080 \
  ghcr.io/jontk/slurm-exporter:latest
```

## Tags and Versioning

- `latest` - Latest stable release
- `v1.2.3` - Specific version
- `main` - Latest development build
- `YYYY-MM-DD-<sha>` - Daily builds with commit SHA

## Support

- GitHub Issues: [Report bugs or request features](https://github.com/jontk/slurm-exporter/issues)
- Documentation: [Full documentation](https://github.com/jontk/slurm-exporter/docs)
- Discussions: [Community discussions](https://github.com/jontk/slurm-exporter/discussions)