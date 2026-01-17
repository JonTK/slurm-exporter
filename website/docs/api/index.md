# API Reference

SLURM Exporter provides several HTTP endpoints for metrics collection, health checks, and operational management.

## Base URL

The default base URL for all endpoints is:
```
http://localhost:9341
```

This can be configured using the `server.host` and `server.port` configuration options.

## Endpoints Overview

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/metrics` | GET | Prometheus metrics | Optional |
| `/health` | GET | Health check | None |
| `/ready` | GET | Readiness check | None |
| `/config` | GET | Configuration dump | Admin |
| `/debug/pprof/*` | GET | Go profiling | Admin |
| `/debug/vars` | GET | Runtime variables | Admin |

## Authentication

### Basic Authentication

When basic authentication is enabled in the configuration:

```yaml
security:
  basic_auth:
    enabled: true
    username: "prometheus"
    password: "secret"
```

Include credentials in requests:

```bash
curl -u prometheus:secret http://localhost:9341/metrics
```

### IP Access Control

Restrict access by IP address:

```yaml
security:
  access_control:
    enabled: true
    allowed_ips:
      - "10.0.0.0/8"
      - "192.168.1.100"
```

## Metrics Endpoint

### GET /metrics

Returns Prometheus-formatted metrics.

**Request:**
```http
GET /metrics HTTP/1.1
Host: localhost:9341
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

**Query Parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `collector` | Filter metrics by collector | `?collector=jobs,nodes` |
| `format` | Output format (prometheus, json) | `?format=json` |

**Examples:**

```bash
# Get all metrics
curl http://localhost:9341/metrics

# Get only job metrics
curl http://localhost:9341/metrics?collector=jobs

# Get metrics in JSON format
curl http://localhost:9341/metrics?format=json

# With authentication
curl -u prometheus:secret http://localhost:9341/metrics
```

**Error Responses:**

```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Basic realm="SLURM Exporter"

HTTP/1.1 403 Forbidden
Content-Type: text/plain
Access denied from IP: 192.168.1.200
```

## Health Endpoints

### GET /health

Returns overall system health status.

**Request:**
```http
GET /health HTTP/1.1
Host: localhost:9341
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
    },
    "memory_usage": {
      "status": "healthy",
      "last_check": "2024-01-15T10:30:00Z",
      "details": "Memory usage: 245MB / 1000MB (24.5%)"
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
Host: localhost:9341
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

## Configuration Endpoint

### GET /config

Returns current configuration (admin only).

**Request:**
```http
GET /config HTTP/1.1
Host: localhost:9341
Authorization: Basic cHJvbWV0aGV1czpzZWNyZXQ=
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "slurm": {
    "host": "slurm-controller.example.com",
    "port": 6820,
    "timeout": "30s"
  },
  "metrics": {
    "enabled_collectors": ["jobs", "nodes", "partitions"],
    "collection_interval": "30s"
  },
  "server": {
    "host": "0.0.0.0",
    "port": 9341
  }
}
```

**Query Parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `section` | Return specific config section | `?section=slurm` |
| `masked` | Mask sensitive values | `?masked=true` |

## Debug Endpoints

Debug endpoints are only available when enabled in development mode.

### GET /debug/vars

Returns runtime variables and statistics.

**Request:**
```http
GET /debug/vars HTTP/1.1
Host: localhost:9341
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "memstats": {
    "Alloc": 2548736,
    "TotalAlloc": 157834496,
    "Sys": 71893248,
    "NumGC": 1247
  },
  "slurm_exporter": {
    "uptime_seconds": 3600,
    "last_collection": "2024-01-15T10:29:50Z",
    "total_collections": 120,
    "collection_errors": 2
  }
}
```

### GET /debug/pprof/

Go profiling endpoints for performance analysis.

**Available Profiles:**

| Endpoint | Description |
|----------|-------------|
| `/debug/pprof/` | Profile index |
| `/debug/pprof/profile` | CPU profile |
| `/debug/pprof/heap` | Memory heap profile |
| `/debug/pprof/goroutine` | Goroutine profile |
| `/debug/pprof/block` | Block profile |
| `/debug/pprof/mutex` | Mutex profile |

**Examples:**

```bash
# Get CPU profile for 30 seconds
curl http://localhost:9341/debug/pprof/profile?seconds=30 > cpu.prof

# Get heap profile
curl http://localhost:9341/debug/pprof/heap > heap.prof

# Analyze with go tool
go tool pprof cpu.prof
```

## Response Formats

### JSON Format

All endpoints support JSON responses. Use `Accept: application/json` header or `?format=json` parameter.

**Metrics in JSON:**
```json
{
  "metrics": [
    {
      "name": "slurm_jobs_total",
      "type": "counter",
      "help": "Total number of jobs by state",
      "samples": [
        {
          "labels": {
            "cluster": "production",
            "state": "completed",
            "user": "alice"
          },
          "value": 142,
          "timestamp": "2024-01-15T10:30:00Z"
        }
      ]
    }
  ]
}
```

### Error Responses

All endpoints return consistent error responses:

```json
{
  "error": {
    "code": "SLURM_CONNECTION_FAILED",
    "message": "Failed to connect to SLURM controller",
    "details": "dial tcp 192.168.1.10:6820: connection refused",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
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

## Rate Limiting

SLURM Exporter implements basic rate limiting to prevent abuse:

- **Metrics endpoint**: 100 requests per minute per IP
- **Health endpoints**: 300 requests per minute per IP
- **Debug endpoints**: 10 requests per minute per IP

Rate limit headers are included in responses:

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1642244400
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

## TLS Configuration

Enable HTTPS for secure communication:

```yaml
server:
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/slurm-exporter.crt"
    key_file: "/etc/ssl/private/slurm-exporter.key"
```

Access endpoints via HTTPS:
```bash
curl https://localhost:9341/metrics
```

## Client Libraries

### Go Client

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
)

func main() {
    client := &http.Client{Timeout: 30 * time.Second}
    
    resp, err := client.Get("http://localhost:9341/health")
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
import json

class SlurmExporterClient:
    def __init__(self, base_url="http://localhost:9341"):
        self.base_url = base_url
        self.session = requests.Session()
        
    def get_metrics(self, collector=None):
        params = {}
        if collector:
            params['collector'] = collector
            
        response = self.session.get(
            f"{self.base_url}/metrics",
            params=params,
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

BASE_URL="http://localhost:9341"
AUTH="prometheus:secret"

# Check health
health_status=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/health")
if [ "$health_status" -eq 200 ]; then
    echo "Service is healthy"
else
    echo "Service is unhealthy (HTTP $health_status)"
fi

# Get metrics with authentication
curl -u "$AUTH" -s "$BASE_URL/metrics" | grep slurm_jobs_total

# Check if ready
if curl -s "$BASE_URL/ready" | grep -q "Ready"; then
    echo "Service is ready"
fi
```

## Troubleshooting API Issues

### Common Issues

1. **Connection Refused**
   - Check if service is running: `systemctl status slurm-exporter`
   - Verify listen address: `netstat -tlnp | grep 9341`

2. **Authentication Failures**
   - Verify credentials in configuration
   - Check basic auth encoding: `echo -n 'user:pass' | base64`

3. **Timeout Errors**
   - Check SLURM controller connectivity
   - Increase timeout values in configuration

4. **High Response Times**
   - Enable caching to improve performance
   - Reduce collection intervals
   - Filter collectors to reduce load

### Debug Commands

```bash
# Test basic connectivity
curl -v http://localhost:9341/health

# Check metrics availability
curl -s http://localhost:9341/metrics | wc -l

# Verify authentication
curl -u user:pass -v http://localhost:9341/metrics

# Get runtime statistics
curl -s http://localhost:9341/debug/vars | jq .

# Profile CPU usage
curl http://localhost:9341/debug/pprof/profile?seconds=10 > profile.out
go tool pprof profile.out
```