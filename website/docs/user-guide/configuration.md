# Configuration Reference

This page provides a comprehensive reference for all SLURM Exporter configuration options.

## Configuration Methods

SLURM Exporter supports multiple configuration methods in order of precedence:

1. **Command-line flags** (highest precedence)
2. **Environment variables**
3. **Configuration file** (lowest precedence)

## Command-Line Flags

The following command-line flags are available:

```bash
slurm-exporter \
  --config=/etc/slurm-exporter/config.yaml \
  --log-level=info \
  --addr=:10341 \
  --metrics-path=/metrics \
  --version \
  --health-check
```

### Flag Reference

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to configuration file | `configs/config.yaml` |
| `--log-level` | Log level (debug, info, warn, error) | `info` |
| `--addr` | HTTP server listen address | `:10341` |
| `--metrics-path` | Metrics endpoint path | `/metrics` |
| `--version` | Show version information and exit | - |
| `--health-check` | Perform health check and exit | - |

## Environment Variables

Configuration can be set using environment variables with the `SLURM_EXPORTER_` prefix. Environment variables use underscores to represent nesting:

```bash
# Server configuration
export SLURM_EXPORTER_SERVER_ADDRESS=":10341"
export SLURM_EXPORTER_SERVER_METRICS_PATH="/metrics"
export SLURM_EXPORTER_SERVER_HEALTH_PATH="/health"
export SLURM_EXPORTER_SERVER_READY_PATH="/ready"

# SLURM configuration
export SLURM_EXPORTER_SLURM_BASE_URL="http://slurm-controller:6820"
export SLURM_EXPORTER_SLURM_API_VERSION="v0.0.44"
export SLURM_EXPORTER_SLURM_AUTH_TYPE="jwt"
export SLURM_EXPORTER_SLURM_AUTH_TOKEN="your-jwt-token"
export SLURM_EXPORTER_SLURM_TIMEOUT="30s"

# Collector configuration
export SLURM_EXPORTER_COLLECTORS_JOBS_ENABLED="true"
export SLURM_EXPORTER_COLLECTORS_NODES_ENABLED="true"
export SLURM_EXPORTER_COLLECTORS_PARTITIONS_ENABLED="true"

# Logging configuration
export SLURM_EXPORTER_LOGGING_LEVEL="info"
export SLURM_EXPORTER_LOGGING_FORMAT="json"
```

## Configuration File

The default configuration file path is `configs/config.yaml`. You can specify a different location using the `--config` flag.

### Server Configuration

Configure the HTTP server that serves metrics, health, and readiness endpoints:

```yaml
server:
  # Listen address (host:port or :port)
  address: ":10341"

  # Endpoint paths
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"

  # Timeouts
  timeout: 30s
  read_timeout: 15s
  write_timeout: 10s
  idle_timeout: 60s

  # Maximum request body size
  max_request_size: 1048576  # 1MB

  # TLS for metrics endpoint (optional)
  tls:
    enabled: false
    cert_file: "/path/to/server.crt"
    key_file: "/path/to/server.key"
    min_version: "1.2"
    cipher_suites:
      - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
      - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"

  # Basic authentication for metrics endpoint (optional)
  basic_auth:
    enabled: false
    username: "prometheus"
    password: "secret"

  # CORS configuration (optional)
  cors:
    enabled: false
    allowed_origins: ["https://grafana.example.com"]
    allowed_methods: ["GET", "OPTIONS"]
    allowed_headers: ["Authorization", "Content-Type"]
```

### SLURM Connection Settings

Configure connection to the SLURM REST API:

```yaml
slurm:
  # SLURM REST API base URL
  base_url: "http://slurm-controller.example.com:6820"

  # API version (supported: v0.0.40, v0.0.41, v0.0.42, v0.0.43, v0.0.44)
  # Leave empty for auto-detection (defaults to v0.0.44)
  api_version: "v0.0.44"

  # Enable adapter pattern for better version compatibility
  use_adapters: true

  # Authentication configuration
  auth:
    type: "jwt"           # Options: none, jwt, basic, apikey
    username: "root"      # For JWT: user to authenticate as
    token: "your-token"   # JWT token string
    token_file: ""        # Or path to file containing JWT token
    # For basic auth:
    # username: "slurm_user"
    # password: "password"
    # password_file: "/path/to/password"
    # For API key auth:
    # api_key: "your-api-key"
    # api_key_file: "/path/to/api-key"

  # Connection tuning
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s

  # TLS configuration for SLURM connections
  tls:
    insecure_skip_verify: false
    ca_cert_file: "/path/to/ca.crt"
    client_cert_file: "/path/to/client.crt"
    client_key_file: "/path/to/client.key"

  # Rate limiting for SLURM API calls
  rate_limit:
    requests_per_second: 10.0
    burst_size: 20
```

### Collectors Configuration

Control which metrics are collected and how. Each collector can be individually configured:

```yaml
collectors:
  # Global collector settings
  global:
    default_interval: 30s
    default_timeout: 10s
    max_concurrency: 5
    error_threshold: 5
    recovery_delay: 60s
    graceful_degradation: true

  # Cluster overview metrics
  cluster:
    enabled: true
    interval: 30s
    timeout: 10s

  # Node metrics
  nodes:
    enabled: true
    interval: 30s
    timeout: 10s
    filters:
      include_nodes: []
      exclude_nodes: []
      node_states: []

  # Job metrics
  jobs:
    enabled: true
    interval: 15s     # More frequent for job status changes
    timeout: 10s
    filters:
      job_states: []

  # User and account metrics
  users:
    enabled: true
    interval: 60s
    timeout: 10s
    filters:
      include_users: []
      exclude_users: []

  # Partition metrics
  partitions:
    enabled: true
    interval: 60s
    timeout: 10s
    filters:
      include_partitions: []
      exclude_partitions: []

  # QoS (Quality of Service) metrics
  qos:
    enabled: true
    interval: 60s
    timeout: 10s

  # Reservation metrics
  reservations:
    enabled: true
    interval: 60s
    timeout: 10s

  # System metrics (exporter self-monitoring)
  system:
    enabled: true
    interval: 60s
    timeout: 10s

  # Graceful degradation configuration
  degradation:
    enabled: true
    max_failures: 3
    reset_timeout: 5m
    use_cached_metrics: true
    cache_ttl: 10m
```

#### Available Collectors

| Collector | Description | Default Enabled |
|-----------|-------------|-----------------|
| `cluster` | Cluster overview metrics | Yes |
| `nodes` | Node state, resources, hardware | Yes |
| `jobs` | Job execution, scheduling, resources | Yes |
| `users` | User activity and resource usage | Yes |
| `partitions` | Partition state, resources, limits | Yes |
| `qos` | Quality of Service policies | Yes |
| `reservations` | Reservation status and resources | Yes |
| `system` | Exporter self-monitoring | Yes |
| `licenses` | License usage metrics | Yes |
| `shares` | Fairshare data | Yes |
| `diagnostics` | Scheduler diagnostics | Yes |
| `tres` | Trackable Resources metrics | Yes |
| `clusters` | Multi-cluster information | Yes |
| `wckeys` | WCKey metrics | No |

#### Collector Filtering

Each collector supports filtering to reduce cardinality:

```yaml
collectors:
  nodes:
    enabled: true
    filters:
      include_nodes: ["node001", "node002"]
      exclude_nodes: ["maintenance-node"]
      node_states: ["IDLE", "ALLOCATED", "MIXED"]
      metrics:
        enable_all: true
        exclude_metrics: ["slurm_node_tres_.*"]

  jobs:
    enabled: true
    filters:
      job_states: ["RUNNING", "PENDING", "COMPLETED"]
      include_users: ["user1", "user2"]
      include_partitions: ["compute", "gpu"]

  partitions:
    enabled: true
    filters:
      include_partitions: ["compute", "gpu"]
      exclude_partitions: ["debug", "test"]
```

#### Error Handling Per Collector

```yaml
collectors:
  jobs:
    enabled: true
    error_handling:
      max_retries: 3
      retry_delay: 5s
      backoff_factor: 2.0
      max_retry_delay: 60s
      fail_fast: false
```

### Logging Configuration

Configure application logging:

```yaml
logging:
  level: "info"        # debug, info, warn, error
  format: "json"       # json, text
  output: "stdout"     # stdout, stderr, file

  # File logging configuration (when output is "file")
  file: "/var/log/slurm-exporter/slurm-exporter.log"
  max_size: 100        # Max size in MB
  max_backups: 10      # Max backup files
  max_age: 30          # Max age in days
  compress: true       # Compress rotated files

  # Suppress HTTP request logging
  suppress_http: false
```

### Metrics Configuration

Configure the Prometheus metrics registry:

```yaml
metrics:
  namespace: "slurm"
  subsystem: "exporter"
  const_labels: {}     # Additional constant labels
  max_age: 5m
  age_buckets: 5

  # Registry configuration
  registry:
    enable_go_collector: true
    enable_process_collector: true
    enable_build_info: true

  # Cardinality management
  cardinality:
    max_series: 10000
    max_labels: 100
    max_label_size: 1024
    warn_limit: 8000
```

### Validation Configuration

```yaml
validation:
  allow_insecure_connections: true
```

### Observability Configuration

Configure performance monitoring, caching, and other observability features:

```yaml
observability:
  # Performance monitoring
  performance_monitoring:
    enabled: true
    interval: 30s
    memory_threshold: 104857600  # 100MB
    cpu_threshold: 80.0

  # Tracing (OpenTelemetry)
  tracing:
    enabled: false
    sample_rate: 0.01
    endpoint: "localhost:4317"
    insecure: true

  # Adaptive collection intervals
  adaptive_collection:
    enabled: true
    min_interval: 30s
    max_interval: 5m
    base_interval: 1m
    score_window: 10m

  # Circuit breaker for SLURM API
  circuit_breaker:
    enabled: true
    failure_threshold: 5
    reset_timeout: 30s
    half_open_requests: 3

  # Health checks
  health:
    enabled: true
    interval: 30s
    timeout: 10s
    checks: ["slurm_api", "collectors", "memory"]

  # Intelligent caching
  caching:
    intelligent: true
    base_ttl: 1m
    max_entries: 50000
    cleanup_interval: 5m
    change_tracking: true
    adaptive_ttl:
      enabled: true
      min_ttl: 30s
      max_ttl: 30m
```

## Hot-Reload Configuration

SLURM Exporter supports hot-reloading configuration without restart by watching the config file for changes. When the configuration file is modified, the exporter will automatically reload it.

Not all configuration changes can be hot-reloaded. Changes requiring restart:

- HTTP server address/port
- TLS certificate changes

## Configuration Examples

### Minimal Configuration

```yaml
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
```

### Production Configuration

```yaml
server:
  address: ":10341"
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"
  timeout: 30s
  read_timeout: 15s
  write_timeout: 10s
  idle_timeout: 60s
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/slurm-exporter.crt"
    key_file: "/etc/ssl/private/slurm-exporter.key"
    min_version: "1.2"
  basic_auth:
    enabled: true
    username: "prometheus"
    password: "secure-password"

slurm:
  base_url: "https://slurm-controller.prod.example.com:6820"
  api_version: "v0.0.44"
  auth:
    type: "jwt"
    username: "root"
    token_file: "/etc/slurm-exporter/token"
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s
  tls:
    insecure_skip_verify: false
    ca_cert_file: "/etc/ssl/certs/ca-bundle.crt"
  rate_limit:
    requests_per_second: 10.0
    burst_size: 20

collectors:
  jobs:
    enabled: true
    interval: 15s
    timeout: 10s
  nodes:
    enabled: true
    interval: 30s
    timeout: 10s
  partitions:
    enabled: true
    interval: 60s
    timeout: 10s
    filters:
      exclude_partitions: ["debug", "test"]
  users:
    enabled: true
    interval: 60s
    timeout: 10s
  qos:
    enabled: true
    interval: 60s
    timeout: 10s
  reservations:
    enabled: true
    interval: 60s
    timeout: 10s
  cluster:
    enabled: true
    interval: 30s
    timeout: 10s
  system:
    enabled: true
    interval: 60s
    timeout: 10s
  degradation:
    enabled: true
    max_failures: 3
    reset_timeout: 5m
    use_cached_metrics: true
    cache_ttl: 10m

logging:
  level: "info"
  format: "json"
  output: "file"
  file: "/var/log/slurm-exporter/slurm-exporter.log"
  max_size: 100
  max_backups: 5
  max_age: 30
  compress: true

metrics:
  namespace: "slurm"
  subsystem: "exporter"
  cardinality:
    max_series: 10000
    max_labels: 100
    max_label_size: 1024
    warn_limit: 8000
```

## Troubleshooting Configuration

### Common Issues

1. **Connection Timeouts**
   ```yaml
   slurm:
     timeout: 60s          # Increase timeout
     retry_attempts: 5     # More retries
     retry_delay: 10s      # Longer delay between retries
   ```

2. **Too Many Metrics (High Cardinality)**
   ```yaml
   collectors:
     nodes:
       filters:
         exclude_partitions: ["debug", "test"]
         metrics:
           exclude_metrics:
             - "slurm_node_tres_.*"
   metrics:
     cardinality:
       max_series: 5000    # Reduce max series
   ```

3. **Debug Configuration**
   ```yaml
   logging:
     level: "debug"
     format: "text"    # Easier to read than json
   ```
