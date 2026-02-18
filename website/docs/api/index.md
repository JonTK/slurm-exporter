# API Reference

SLURM Exporter provides several HTTP endpoints for metrics collection, health checks, and operational management.

## Base URL

The default base URL for all endpoints is:
```
http://localhost:10341
```

This can be configured using the `server.address` configuration option or the `--addr` CLI flag.

## Endpoints Overview

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/metrics` | GET | Prometheus metrics | Optional (basic auth) |
| `/health` | GET | Health check | None |
| `/ready` | GET | Readiness check | None |

## Authentication

### Basic Authentication

When basic authentication is enabled in the configuration:

```yaml
server:
  basic_auth:
    enabled: true
    username: "prometheus"
    password: "secret"
```

Include credentials in requests:

```bash
curl -u prometheus:secret http://localhost:10341/metrics
```

## Metrics Endpoint

### GET /metrics

Returns Prometheus-formatted metrics.

**Request:**
```http
GET /metrics HTTP/1.1
Host: localhost:10341
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: text/plain; version=0.0.4; charset=utf-8

# HELP slurm_jobs_total Total number of jobs by state
# TYPE slurm_jobs_total counter
slurm_jobs_total{cluster="production",partition="compute",state="completed",user="alice"} 142
slurm_jobs_total{cluster="production",partition="compute",state="running",user="alice"} 5
slurm_jobs_total{cluster="production",partition="gpu",state="pending",user="bob"} 12

# HELP slurm_nodes_total Total number of nodes by state
# TYPE slurm_nodes_total gauge
slurm_nodes_total{cluster="production",partition="compute",state="idle"} 45
slurm_nodes_total{cluster="production",partition="compute",state="allocated"} 23
slurm_nodes_total{cluster="production",partition="compute",state="down"} 2
```

**Examples:**

```bash
# Get all metrics
curl http://localhost:10341/metrics

# With authentication
curl -u prometheus:secret http://localhost:10341/metrics
```

**Error Responses:**

```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Basic realm="SLURM Exporter"

HTTP/1.1 403 Forbidden
Content-Type: text/plain
Access denied
```

## Health Endpoints

### GET /health

Returns overall system health status.

**Request:**
```http
GET /health HTTP/1.1
Host: localhost:10341
```

**Success Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "slurm_connectivity": {
      "status": "healthy",
      "last_check": "2024-01-15T10:29:45Z",
      "details": "Connected to slurm-controller.example.com:6820"
    },
    "metric_collection": {
      "status": "healthy",
      "last_check": "2024-01-15T10:29:50Z",
      "details": "Last collection completed in 2.3s"
    }
  },
  "uptime_seconds": 3600,
  "version": "1.0.0"
}
```

**Failure Response:**
```http
HTTP/1.1 503 Service Unavailable
Content-Type: application/json

{
  "status": "unhealthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "slurm_connectivity": {
      "status": "unhealthy",
      "last_check": "2024-01-15T10:29:45Z",
      "details": "Connection timeout to slurm-controller.example.com:6820",
      "error": "dial tcp 192.168.1.10:6820: i/o timeout"
    }
  }
}
```

### GET /ready

Returns readiness status for load balancer checks.

**Request:**
```http
GET /ready HTTP/1.1
Host: localhost:10341
```

**Success Response:**
```http
HTTP/1.1 200 OK
Content-Type: text/plain

Ready
```

**Failure Response:**
```http
HTTP/1.1 503 Service Unavailable
Content-Type: text/plain

Not Ready: SLURM connectivity check failed
```

## HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Access denied |
| 404 | Not Found | Endpoint not found |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service unhealthy |

## TLS Configuration

Enable HTTPS for secure communication:

```yaml
server:
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/slurm-exporter.crt"
    key_file: "/etc/ssl/private/slurm-exporter.key"
    min_version: "1.2"
```

Access endpoints via HTTPS:
```bash
curl https://localhost:10341/metrics
```

## CORS Support

Cross-Origin Resource Sharing (CORS) can be enabled:

```yaml
server:
  cors:
    enabled: true
    allowed_origins: ["https://grafana.example.com"]
    allowed_methods: ["GET", "OPTIONS"]
    allowed_headers: ["Authorization", "Content-Type"]
```

## Client Libraries

### Go Client

```go
package main

import (
    "fmt"
    "net/http"
    "time"
)

func main() {
    client := &http.Client{Timeout: 30 * time.Second}

    resp, err := client.Get("http://localhost:10341/health")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Printf("Health check status: %d\n", resp.StatusCode)
}
```

### Python Client

```python
import requests

class SlurmExporterClient:
    def __init__(self, base_url="http://localhost:10341"):
        self.base_url = base_url
        self.session = requests.Session()

    def get_metrics(self):
        response = self.session.get(
            f"{self.base_url}/metrics",
            timeout=30
        )
        response.raise_for_status()
        return response.text

    def get_health(self):
        response = self.session.get(f"{self.base_url}/health")
        response.raise_for_status()
        return response.json()

# Usage
client = SlurmExporterClient()
health = client.get_health()
print(f"Status: {health['status']}")
```

### Bash/cURL Examples

```bash
#!/bin/bash

BASE_URL="http://localhost:10341"

# Check health
health_status=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/health")
if [ "$health_status" -eq 200 ]; then
    echo "Service is healthy"
else
    echo "Service is unhealthy (HTTP $health_status)"
fi

# Get metrics with authentication (if enabled)
curl -u "prometheus:secret" -s "$BASE_URL/metrics" | grep slurm_jobs_total

# Check if ready
if curl -s "$BASE_URL/ready" | grep -q "Ready"; then
    echo "Service is ready"
fi
```

## Troubleshooting API Issues

### Common Issues

1. **Connection Refused**
   - Check if service is running: `systemctl status slurm-exporter`
   - Verify listen address: `netstat -tlnp | grep 10341`

2. **Authentication Failures**
   - Verify credentials in configuration
   - Check basic auth encoding: `echo -n 'user:pass' | base64`

3. **Timeout Errors**
   - Check SLURM controller connectivity
   - Increase timeout values in configuration

### Debug Commands

```bash
# Test basic connectivity
curl -v http://localhost:10341/health

# Check metrics availability
curl -s http://localhost:10341/metrics | wc -l

# Verify authentication
curl -u user:pass -v http://localhost:10341/metrics
```
