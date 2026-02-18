# Distributed Tracing

SLURM Exporter includes OpenTelemetry tracing configuration in its config schema for future observability into metric collection operations and performance analysis.

!!! warning "Not Currently Wired Into the Application"

    **OpenTelemetry tracing configuration exists in the config schema but is NOT currently wired into the application.** The tracing fields in the configuration file are parsed and validated, but no traces are actually generated or exported. This page documents the configuration schema for when tracing support is fully implemented in a future release.

## Configuration Schema

The tracing configuration lives under the `observability.tracing` section:

```yaml
observability:
  tracing:
    enabled: false          # Disabled by default
    sample_rate: 0.01       # 1% sampling when enabled
    endpoint: "localhost:4317"  # OTLP gRPC endpoint
    insecure: true          # Use insecure connection (no TLS)
```

### Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable tracing (not currently functional) |
| `sample_rate` | float | `0.01` | Sampling rate between 0 and 1 |
| `endpoint` | string | `localhost:4317` | OTLP collector endpoint |
| `insecure` | bool | `true` | Whether to use insecure connection |

### Validation Rules

When `enabled` is set to `true`, the following validation rules apply:

- `sample_rate` must be between 0 and 1
- `endpoint` must not be empty

### Environment Variable Overrides

The tracing configuration can also be set via environment variables:

```bash
export SLURM_EXPORTER_OBSERVABILITY_TRACING_ENABLED=true
export SLURM_EXPORTER_OBSERVABILITY_TRACING_SAMPLE_RATE=0.05
export SLURM_EXPORTER_OBSERVABILITY_TRACING_ENDPOINT=localhost:4317
export SLURM_EXPORTER_OBSERVABILITY_TRACING_INSECURE=true
```

## Example Configuration

```yaml title="config.yaml"
server:
  address: ":10341"
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"

slurm:
  base_url: "http://slurm-controller:6820"
  api_version: "v0.0.44"
  auth:
    type: "jwt"
    username: "root"
    token: "your-jwt-token"

collectors:
  jobs:
    enabled: true
  nodes:
    enabled: true

logging:
  level: "info"
  format: "json"

# Tracing configuration (schema only - not currently functional)
observability:
  tracing:
    enabled: false
    sample_rate: 0.01
    endpoint: "localhost:4317"
    insecure: true
```

## Future Plans

When tracing support is fully implemented, it will provide:

- **End-to-end visibility** into metric collection workflows
- **Performance bottleneck identification** in SLURM API calls
- **Detailed timing analysis** of collection operations
- **Error tracking and debugging** across the entire stack
- **Integration** with popular tracing backends (Jaeger, Zipkin, etc.) via OTLP

## Current Observability Alternatives

While tracing is not yet functional, SLURM Exporter provides other observability features:

### Self-Monitoring Metrics

The exporter exposes metrics about its own performance:

```promql
# Collection duration per collector
slurm_exporter_collect_duration_seconds

# Collection errors
slurm_exporter_collect_errors_total

# API request duration
slurm_exporter_api_request_duration_seconds
```

### Debug Logging

Enable debug logging for detailed operation visibility:

```yaml
logging:
  level: "debug"
  format: "json"
```

Or via CLI flag:

```bash
slurm-exporter --config=config.yaml --log-level=debug
```

### Health Endpoints

Use the health and readiness endpoints for operational monitoring:

```bash
# Health check
curl http://localhost:10341/health

# Readiness check
curl http://localhost:10341/ready
```

### Performance Monitoring

The observability configuration also includes performance monitoring that IS functional:

```yaml
observability:
  performance_monitoring:
    enabled: true
    interval: 30s
    memory_threshold: 104857600  # 100MB
    cpu_threshold: 80.0

  circuit_breaker:
    enabled: true
    failure_threshold: 5
    reset_timeout: 30s
    half_open_requests: 3
```

For more information on monitoring the exporter itself, see the [Metrics Catalog](../user-guide/metrics.md) and [Prometheus Integration](../integration/prometheus.md) guides.
