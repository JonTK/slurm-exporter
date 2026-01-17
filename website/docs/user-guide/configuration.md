# Configuration Reference

This page provides a comprehensive reference for all SLURM Exporter configuration options.

## Configuration Methods

SLURM Exporter supports multiple configuration methods in order of precedence:

1. **Configuration file** (highest precedence)
2. **Command-line arguments**
3. **Environment variables** (lowest precedence)

## Configuration File

The default configuration file location is `/etc/slurm-exporter/config.yaml`. You can specify a different location using the `--config.file` flag.

### SLURM Connection Settings

Configure connection to the SLURM REST API:

```yaml
slurm:
  # SLURM REST API endpoint
  host: "slurm-controller.example.com"
  port: 6820
  
  # Authentication methods
  auth:
    type: "jwt"  # Options: jwt, basic, certificate
    token: "/path/to/token"  # JWT token string or file path
    # For basic auth:
    # user: "slurm_user"
    # password: "password"
    # For certificate auth:
    # cert: "/path/to/client.crt"
    # key: "/path/to/client.key"
  
  # TLS configuration
  tls:
    enabled: true
    verify: true  # Verify server certificate
    ca_cert: "/path/to/ca.crt"
    client_cert: "/path/to/client.crt"
    client_key: "/path/to/client.key"
  
  # Connection tuning
  timeout: 30s
  max_retries: 3
  retry_delay: 5s
```

### Metrics Collection

Control which metrics are collected and how:

```yaml
metrics:
  # Enable specific collectors
  enabled_collectors:
    - jobs           # Job metrics
    - nodes          # Node metrics
    - partitions     # Partition metrics
    - accounts       # Account metrics
    - fairshare      # Fairshare policy metrics
    - qos            # Quality of Service metrics
    - reservations   # Reservation metrics
    - licenses       # License usage metrics
    - tres           # Trackable Resources
    - associations   # User/account associations
  
  # Collection frequency
  collection_interval: 30s
  
  # Job analytics features
  job_analytics:
    enabled: true
    efficiency_calculation: true
    waste_detection: true
    bottleneck_analysis: true
    trend_analysis: true
  
  # Metric filtering to reduce cardinality
  filtering:
    enabled: true
    exclude_metrics:
      - "slurm_node_tres_.*_allocated"
    include_partitions: ["compute", "gpu"]
    exclude_partitions: ["debug", "test"]
  
  # Add custom labels to all metrics
  custom_labels:
    cluster: "production"
    datacenter: "dc1"
    environment: "prod"
```

### Performance Optimization

Configure caching, connection pooling, and other performance features:

```yaml
performance:
  # Intelligent caching
  cache:
    enabled: true
    ttl: 300s              # Cache TTL
    max_size: 10000        # Maximum cache entries
  
  # Connection pooling
  connection_pool:
    enabled: true
    max_connections: 10
    max_idle_connections: 5
    connection_lifetime: 300s
  
  # Batch processing
  batch_processing:
    enabled: true
    batch_size: 1000
    batch_timeout: 10s
  
  # Memory management
  memory:
    max_heap_size: 1000000000  # 1GB
    gc_target_percentage: 10
```

### HTTP Server

Configure the metrics HTTP server:

```yaml
server:
  # Listening configuration
  host: "0.0.0.0"
  port: 9341
  
  # Endpoint paths
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"
  
  # Timeouts
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s
  
  # TLS for metrics endpoint
  tls:
    enabled: false
    cert_file: "/path/to/server.crt"
    key_file: "/path/to/server.key"
  
  # Access logging
  access_log:
    enabled: true
    path: "/var/log/slurm-exporter/access.log"
```

### Logging

Configure application logging:

```yaml
logging:
  level: "info"        # debug, info, warn, error
  format: "json"       # json, text
  output: "stdout"     # stdout, stderr, or file path
  
  # File logging configuration
  file:
    path: "/var/log/slurm-exporter/slurm-exporter.log"
    max_size: 100      # MB
    max_backups: 10
    max_age: 30        # days
    compress: true
```

### Health Checks

Configure health and readiness checks:

```yaml
health:
  enabled: true
  
  # Readiness checks
  readiness:
    - slurm_connectivity
    - metric_collection
  
  # Liveness checks
  liveness:
    - http_server
    - memory_usage
```

### Security

Configure authentication and access control:

```yaml
security:
  # Basic authentication for metrics endpoint
  basic_auth:
    enabled: false
    username: "prometheus"
    password: "secret"
  
  # IP-based access control
  access_control:
    enabled: false
    allowed_ips:
      - "127.0.0.1"
      - "10.0.0.0/8"
      - "172.16.0.0/12"
      - "192.168.0.0/16"
```

### Monitoring and Observability

Configure self-monitoring and debugging:

```yaml
monitoring:
  # Self-monitoring metrics
  self_monitoring:
    enabled: true
    include_build_info: true
    include_go_metrics: true
  
  # Profiling endpoint (debug only)
  profiling:
    enabled: false
    path: "/debug/pprof"
  
  # Distributed tracing
  tracing:
    enabled: false
    jaeger:
      endpoint: "http://jaeger:14268/api/traces"
      service_name: "slurm-exporter"
```

## Command-Line Arguments

All configuration options can be overridden using command-line flags:

```bash
slurm-exporter \
  --config.file=/etc/slurm-exporter/config.yaml \
  --slurm.host=slurm-controller.example.com \
  --slurm.port=6820 \
  --slurm.token=/path/to/token \
  --web.listen-address=:9341 \
  --web.telemetry-path=/metrics \
  --log.level=info \
  --log.format=json \
  --collectors.enabled=jobs,nodes,partitions
```

### Key Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config.file` | Configuration file path | `/etc/slurm-exporter/config.yaml` |
| `--slurm.host` | SLURM controller hostname | `localhost` |
| `--slurm.port` | SLURM REST API port | `6820` |
| `--slurm.token` | JWT authentication token | - |
| `--web.listen-address` | HTTP server listen address | `:9341` |
| `--web.telemetry-path` | Metrics endpoint path | `/metrics` |
| `--log.level` | Log level | `info` |
| `--collectors.enabled` | Enabled collectors (comma-separated) | `jobs,nodes,partitions` |

## Environment Variables

Configuration can be set using environment variables with the `SLURM_EXPORTER_` prefix:

```bash
export SLURM_EXPORTER_SLURM_HOST=slurm-controller.example.com
export SLURM_EXPORTER_SLURM_TOKEN=/path/to/token
export SLURM_EXPORTER_LOG_LEVEL=debug
export SLURM_EXPORTER_WEB_LISTEN_ADDRESS=:9341
```

### Environment Variable Mapping

Environment variables use underscores instead of dots and hyphens:

- `slurm.host` → `SLURM_EXPORTER_SLURM_HOST`
- `web.listen-address` → `SLURM_EXPORTER_WEB_LISTEN_ADDRESS`
- `log.level` → `SLURM_EXPORTER_LOG_LEVEL`

## Configuration Validation

SLURM Exporter validates configuration on startup and provides detailed error messages for invalid settings:

```bash
# Validate configuration without starting
slurm-exporter --config.file=/path/to/config.yaml --validate-config

# Test SLURM connectivity
slurm-exporter --config.file=/path/to/config.yaml --test-connection
```

## Hot-Reload Configuration

SLURM Exporter supports hot-reloading configuration without restart:

```bash
# Send SIGHUP to reload configuration
kill -HUP $(pidof slurm-exporter)

# Or use systemd
systemctl reload slurm-exporter
```

Not all configuration changes can be hot-reloaded. Changes requiring restart:

- HTTP server address/port
- TLS certificate changes
- Authentication method changes

## Configuration Examples

### Minimal Configuration

```yaml
slurm:
  host: "slurm-controller"
  auth:
    type: "jwt"
    token: "your-jwt-token"

metrics:
  enabled_collectors: ["jobs", "nodes"]
```

### Production Configuration

```yaml
slurm:
  host: "slurm-controller.prod.example.com"
  port: 6820
  auth:
    type: "jwt"
    token: "/etc/slurm-exporter/token"
  tls:
    enabled: true
    verify: true
    ca_cert: "/etc/ssl/certs/ca-bundle.crt"
  timeout: 30s
  max_retries: 3

metrics:
  enabled_collectors:
    - jobs
    - nodes
    - partitions
    - accounts
    - fairshare
    - qos
  collection_interval: 30s
  job_analytics:
    enabled: true
    efficiency_calculation: true
    waste_detection: true
  filtering:
    enabled: true
    exclude_partitions: ["debug", "test"]
  custom_labels:
    cluster: "production"
    datacenter: "dc1"

performance:
  cache:
    enabled: true
    ttl: 300s
  connection_pool:
    enabled: true
    max_connections: 10

server:
  host: "0.0.0.0"
  port: 9341
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/slurm-exporter.crt"
    key_file: "/etc/ssl/private/slurm-exporter.key"

logging:
  level: "info"
  format: "json"
  file:
    path: "/var/log/slurm-exporter/slurm-exporter.log"
    max_size: 100
    max_backups: 5

security:
  basic_auth:
    enabled: true
    username: "prometheus"
    password: "secure-password"
  access_control:
    enabled: true
    allowed_ips:
      - "10.0.0.0/8"
      - "172.16.0.0/12"
```

### High-Availability Configuration

```yaml
slurm:
  # Multiple SLURM controllers for HA
  hosts:
    - "slurm-controller-1.example.com"
    - "slurm-controller-2.example.com"
  port: 6820
  auth:
    type: "certificate"
    cert: "/etc/ssl/certs/slurm-client.crt"
    key: "/etc/ssl/private/slurm-client.key"
  tls:
    enabled: true
    verify: true
    ca_cert: "/etc/ssl/certs/slurm-ca.crt"

performance:
  cache:
    enabled: true
    ttl: 60s  # Shorter TTL for HA
  connection_pool:
    enabled: true
    max_connections: 20

monitoring:
  self_monitoring:
    enabled: true
    include_build_info: true
```

## Troubleshooting Configuration

### Common Issues

1. **Connection Timeouts**
   ```yaml
   slurm:
     timeout: 60s  # Increase timeout
     max_retries: 5
   ```

2. **High Memory Usage**
   ```yaml
   performance:
     cache:
       max_size: 5000  # Reduce cache size
     memory:
       max_heap_size: 500000000  # 500MB
   ```

3. **Too Many Metrics (High Cardinality)**
   ```yaml
   metrics:
     filtering:
       enabled: true
       exclude_metrics:
         - "slurm_node_tres_.*"
       exclude_partitions: ["debug", "test"]
   ```

### Debug Configuration

Enable debug logging and additional endpoints:

```yaml
logging:
  level: "debug"

development:
  enabled: true
  debug_endpoints:
    enabled: true

monitoring:
  profiling:
    enabled: true
```