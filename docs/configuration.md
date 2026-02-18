# SLURM Exporter Configuration Reference

This comprehensive guide covers all configuration options for SLURM Exporter, with examples and best practices for different deployment scenarios.

## Quick Configuration

For a full annotated example configuration, see `configs/config.yaml` in the repository root. That file documents every supported field with comments and can be used as a starting point for your deployment.

## Table of Contents

- [Configuration File Structure](#configuration-file-structure)
- [Server Configuration](#server-configuration)
- [SLURM Connection Settings](#slurm-connection-settings)
- [Authentication Options](#authentication-options)
- [Collector Configuration](#collector-configuration)
- [Performance Configuration](#performance-configuration)
- [Observability Configuration](#observability-configuration)
- [Debug Configuration](#debug-configuration)
- [Logging Settings](#logging-settings)
- [Advanced Configuration](#advanced-configuration)
- [Environment Variables](#environment-variables)
- [Configuration Examples](#configuration-examples)
- [Configuration Hot Reload](#configuration-hot-reload)
- [Best Practices](#best-practices)

## Configuration File Structure

The SLURM Exporter uses YAML format for configuration. The configuration can be provided via:

1. Configuration file: `--config /path/to/config.yaml`
2. Environment variables (prefixed with `SLURM_EXPORTER_`)
3. Command-line flags
4. Default values

Priority order: Command-line flags > Environment variables > Configuration file > Defaults

### Basic Structure

```yaml
# config.yaml
server:
  # HTTP server settings

slurm:
  # SLURM API connection settings

collectors:
  # Individual collector configurations

logging:
  # Logging configuration

metrics:
  # Metrics exposition settings
```

## Server Configuration

### HTTP Server Settings

```yaml
server:
  # Network address to bind to
  # Default: ":8080"
  address: ":8080"

  # General request timeout
  # Default: "30s"
  timeout: "30s"

  # Read timeout for HTTP requests
  # Default: "15s"
  read_timeout: "15s"

  # Write timeout for HTTP responses
  # Default: "10s"
  write_timeout: "10s"

  # Idle timeout for keep-alive connections
  # Default: "60s"
  idle_timeout: "60s"

  # Maximum request size in bytes
  # Default: 1048576 (1MB)
  max_request_size: 1048576

  # Metrics path
  # Default: "/metrics"
  metrics_path: "/metrics"

  # Health check path
  # Default: "/health"
  health_path: "/health"

  # Ready check path
  # Default: "/ready"
  ready_path: "/ready"
```

### TLS Configuration

```yaml
server:
  tls:
    # Enable TLS
    # Default: false
    enabled: true

    # Path to certificate file
    cert_file: "/etc/ssl/certs/server.crt"

    # Path to private key file
    key_file: "/etc/ssl/private/server.key"

    # Minimum TLS version
    # Options: "1.0", "1.1", "1.2", "1.3"
    # Default: "1.2"
    min_version: "1.2"

    # Cipher suites (if empty, uses Go defaults)
    cipher_suites:
      - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
      - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
```

### HTTP Authentication

```yaml
server:
  basic_auth:
    # Enable basic authentication for metrics endpoint
    # Default: false
    enabled: true

    # Username for basic auth
    username: "prometheus"

    # Password for basic auth
    password: "secure-password-here"
```

## SLURM Connection Settings

### Basic Connection

```yaml
slurm:
  # SLURM REST API base URL
  # Required
  base_url: "http://slurm-server:6820"

  # API version to use
  # Default: "" (empty enables auto-detection, falls back to "v0.0.44")
  # Supported versions: v0.0.40, v0.0.41, v0.0.42, v0.0.43, v0.0.44
  api_version: "v0.0.44"

  # Enable adapter pattern for better version compatibility
  # Default: true
  use_adapters: true

  # Request timeout
  # Default: "30s"
  timeout: "30s"

  # Number of retry attempts
  # Default: 3
  retry_attempts: 3

  # Delay between retries
  # Default: "5s"
  retry_delay: "5s"
```

### TLS for SLURM Connection

```yaml
slurm:
  tls:
    # Skip certificate verification (insecure)
    # Default: false
    insecure_skip_verify: false

    # CA certificate for server verification
    ca_cert_file: "/etc/ssl/certs/slurm-ca.crt"

    # Client certificate for mutual TLS
    client_cert_file: "/etc/ssl/certs/client.crt"
    client_key_file: "/etc/ssl/private/client.key"
```

### Rate Limiting

```yaml
slurm:
  rate_limit:
    # Requests per second to the SLURM API
    # Default: 10.0
    requests_per_second: 10.0

    # Burst size
    # Default: 20
    burst_size: 20
```

## Authentication Options

### JWT Token Authentication

SLURM JWT authentication sets the `X-SLURM-USER-TOKEN` header. When a `username` is also provided, the exporter additionally sets the `X-SLURM-USER-NAME` header, which is required by SLURM REST API to identify the requesting user.

```yaml
slurm:
  auth:
    # Authentication type
    # Options: "none", "jwt", "basic", "apikey"
    type: "jwt"

    # SLURM username (sets X-SLURM-USER-NAME header)
    # Required for SLURM REST API to identify the user
    username: "your-slurm-username"

    # JWT token (direct value)
    token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."

    # JWT token from file (alternative to token)
    token_file: "/run/secrets/slurm-jwt-token"
```

### Basic Authentication

```yaml
slurm:
  auth:
    type: "basic"

    # Username
    username: "prometheus-user"

    # Password (direct value)
    password: "secure-password"

    # Password from file (alternative to password)
    password_file: "/run/secrets/slurm-password"
```

### API Key Authentication

```yaml
slurm:
  auth:
    type: "apikey"

    # API key value
    api_key: "your-api-key"

    # API key from file (alternative to api_key)
    api_key_file: "/run/secrets/slurm-apikey"
```

## Collector Configuration

### Global Collector Settings

```yaml
collectors:
  # Global settings for all collectors
  global:
    # Default collection interval
    # Default: "30s"
    default_interval: "30s"

    # Default collection timeout
    # Default: "10s"
    default_timeout: "10s"

    # Maximum concurrent collectors
    # Default: 5
    max_concurrency: 5

    # Error threshold before degradation
    # Default: 5
    error_threshold: 5

    # Recovery delay after errors
    # Default: "60s"
    recovery_delay: "60s"

    # Enable graceful degradation
    # Default: true
    graceful_degradation: true
```

### Cluster Collector

```yaml
collectors:
  cluster:
    # Enable cluster collector
    # Default: true
    enabled: true

    # Collection interval
    # Default: "30s"
    interval: "30s"

    # Collection timeout
    # Default: "10s"
    timeout: "10s"

    # Maximum concurrency
    max_concurrency: 1

    # Error handling
    error_handling:
      max_retries: 3
      retry_delay: "5s"
      backoff_factor: 2.0
      max_retry_delay: "60s"
      fail_fast: false
```

### Nodes Collector

```yaml
collectors:
  nodes:
    # Enable nodes collector
    # Default: true
    enabled: true

    # Collection interval
    # Default: "30s"
    interval: "30s"

    # Collection timeout
    # Default: "10s"
    timeout: "10s"

    # Maximum concurrency
    max_concurrency: 2

    # Filters
    filters:
      # Optionally filter specific nodes or states
      # include_nodes: ["node001", "node002"]
      # exclude_nodes: ["maintenance-node"]
      # node_states: ["IDLE", "ALLOCATED", "MIXED"]

    # Error handling
    error_handling:
      max_retries: 3
      retry_delay: "5s"
      backoff_factor: 2.0
      max_retry_delay: "60s"
      fail_fast: false
```

### Jobs Collector

```yaml
collectors:
  jobs:
    # Enable jobs collector
    # Default: true
    enabled: true

    # Collection interval
    # Default: "15s"
    interval: "15s"

    # Collection timeout
    # Default: "10s"
    timeout: "10s"

    # Maximum concurrency
    max_concurrency: 3

    # Filters
    filters:
      # Optionally filter job states
      # job_states: ["RUNNING", "PENDING", "COMPLETED"]

    # Error handling
    error_handling:
      max_retries: 3
      retry_delay: "5s"
      backoff_factor: 2.0
      max_retry_delay: "60s"
      fail_fast: false
```

### Partitions Collector

```yaml
collectors:
  partitions:
    # Enable partitions collector
    # Default: true
    enabled: true

    # Collection interval
    # Default: "60s"
    interval: "60s"

    # Collection timeout
    # Default: "10s"
    timeout: "10s"

    # Filters
    filters:
      # include_partitions: ["compute", "gpu"]
      # exclude_partitions: ["debug"]

    # Error handling
    error_handling:
      max_retries: 3
      retry_delay: "5s"
      backoff_factor: 2.0
      max_retry_delay: "60s"
      fail_fast: false
```

### Reservations Collector

```yaml
collectors:
  reservations:
    # Enable reservations collector
    # Default: true
    enabled: true

    # Collection interval
    # Default: "60s"
    interval: "60s"

    # Collection timeout
    # Default: "10s"
    timeout: "10s"

    # Filters
    filters:
      # include_partitions: ["compute", "gpu"]
      # include_accounts: ["research", "engineering"]

    # Error handling
    error_handling:
      max_retries: 3
      retry_delay: "5s"
      backoff_factor: 2.0
      max_retry_delay: "60s"
      fail_fast: false
```

### QoS Collector

```yaml
collectors:
  qos:
    # Enable QoS collector
    # Default: true
    enabled: true

    # Collection interval
    # Default: "60s"
    interval: "60s"

    # Collection timeout
    # Default: "10s"
    timeout: "10s"

    # Filters
    filters:
      # include_qos: ["normal", "high"]
      # exclude_qos: ["low"]

    # Error handling
    error_handling:
      max_retries: 3
      retry_delay: "5s"
      backoff_factor: 2.0
      max_retry_delay: "60s"
      fail_fast: false
```

### Graceful Degradation

```yaml
collectors:
  degradation:
    # Enable graceful degradation
    # Default: true
    enabled: true

    # Maximum failures before degradation
    # Default: 3
    max_failures: 3

    # Time before resetting failure count
    # Default: "5m"
    reset_timeout: "5m"

    # Use cached metrics during degradation
    # Default: true
    use_cached_metrics: true

    # TTL for cached metrics
    # Default: "10m"
    cache_ttl: "10m"
```

## Performance Configuration

### Intelligent Caching
```yaml
performance:
  caching:
    enabled: true
    default_ttl: 5m
    max_entries: 10000
    cleanup_interval: 60s

    # Adaptive TTL based on data change patterns
    adaptive_ttl:
      enabled: true
      min_ttl: 30s
      max_ttl: 30m
      change_threshold: 0.1             # 10% change triggers shorter TTL
      variance_threshold: 0.05          # Variance threshold for stability
      stability_factor: 0.9
      extension_factor: 1.5
      reduction_factor: 0.5

    # Per-collector cache settings
    collectors:
      jobs:
        ttl: 60s
        max_entries: 5000
      nodes:
        ttl: 300s
        max_entries: 1000
      partitions:
        ttl: 600s
        max_entries: 100
```

### Batch Processing
```yaml
performance:
  batch_processing:
    enabled: true
    max_batch_size: 500
    flush_interval: 5s
    max_wait_time: 10s

    # Deduplication settings
    deduplication:
      enabled: true
      window_size: 1000

    # Per-collector batch settings
    collectors:
      jobs:
        batch_size: 1000
        flush_interval: 3s
      nodes:
        batch_size: 200
        flush_interval: 10s
```

### Memory Management
```yaml
performance:
  memory:
    max_heap_size: "512MB"
    gc_target_percentage: 75

    # Cardinality limits
    max_series: 100000
    max_labels_per_series: 50
    max_label_name_length: 64
    max_label_value_length: 256

    # Connection pooling
    max_idle_connections: 100
    connection_lifetime: 30m
```

## Observability Configuration

### Circuit Breaker
```yaml
observability:
  circuit_breaker:
    enabled: true
    failure_threshold: 10               # Failures to trigger open state
    reset_timeout: 60s                  # Time before trying half-open
    half_open_max_requests: 3           # Max requests in half-open state

    # Per-collector circuit breakers
    collectors:
      jobs:
        failure_threshold: 5
        reset_timeout: 30s
      nodes:
        failure_threshold: 10
        reset_timeout: 120s
```

### Health Monitoring
```yaml
observability:
  health_monitoring:
    enabled: true
    check_interval: 30s
    unhealthy_threshold: 3              # Failures before marking unhealthy

    # Health checks
    checks:
      slurm_api:
        enabled: true
        timeout: 10s
        critical: true                  # Failure marks entire service unhealthy
      collectors:
        enabled: true
        timeout: 5s
        critical: false
      memory:
        enabled: true
        max_memory_mb: 1024
        critical: true

    # Health endpoint configuration
    endpoint:
      path: "/health"
      detailed: true                    # Include check details in response
      authentication:
        enabled: false
```

### OpenTelemetry Tracing

**Note:** Tracing configuration exists in the config schema but is NOT currently wired into the application. The fields below are parsed and validated, but enabling tracing will have no effect until a future release connects the tracing pipeline.

```yaml
observability:
  tracing:
    enabled: false                      # Not yet functional
    sample_rate: 0.01                   # 1% sampling when eventually enabled
    endpoint: "localhost:4317"
    insecure: true
```

## Debug Configuration

### Debug Endpoints
```yaml
debug:
  endpoints:
    enabled: true
    path_prefix: "/debug"

    # Authentication for debug endpoints
    authentication:
      enabled: true
      type: "token"                     # token, basic, none
      token: "${DEBUG_TOKEN}"

    # Available endpoints
    endpoints:
      status: true                      # /debug/status
      collectors: true                  # /debug/collectors
      cache: true                      # /debug/cache
      config: false                    # /debug/config (sensitive)
      traces: true                     # /debug/traces

    # Output formats
    formats:
      json: true
      html: true
      prometheus: true
```

## Logging Settings

### Basic Logging

```yaml
logging:
  # Log level
  # Options: "debug", "info", "warn", "error"
  # Default: "info"
  level: "info"

  # Log format
  # Options: "text", "json"
  # Default: "json"
  format: "json"

  # Output destination
  # Options: "stdout", "stderr", "file"
  # Default: "stdout"
  output: "stdout"
```

### File Logging

```yaml
logging:
  # Output to file
  output: "file"
  file: "/var/log/slurm-exporter/exporter.log"

  # Maximum size in MB before rotation
  # Default: 100
  max_size: 100

  # Maximum age in days
  # Default: 7
  max_age: 7

  # Maximum number of old files to keep
  # Default: 3
  max_backups: 3

  # Compress rotated files
  # Default: true
  compress: true
```

### Structured Logging Fields

```yaml
logging:
  # Additional fields to include in all logs
  fields:
    service: "slurm-exporter"
    environment: "production"
    cluster: "hpc-prod"

  # Suppress HTTP request logs
  # Default: false
  suppress_http: false
```

## Advanced Configuration

### Metrics Configuration

```yaml
metrics:
  # Metric namespace
  # Default: "slurm"
  namespace: "slurm"

  # Metric subsystem
  # Default: "exporter"
  subsystem: "exporter"

  # Additional constant labels applied to all metrics
  const_labels: {}

  # Maximum age for summary metrics
  # Default: "5m"
  max_age: "5m"

  # Number of age buckets for summaries
  # Default: 5
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

## Environment Variables

All configuration options can be set via environment variables using the prefix `SLURM_EXPORTER_`.

### Environment Variable Mapping

```bash
# Server configuration
export SLURM_EXPORTER_SERVER_ADDRESS=":8080"
export SLURM_EXPORTER_SERVER_READ_TIMEOUT="30s"
export SLURM_EXPORTER_SERVER_TLS_ENABLED="true"
export SLURM_EXPORTER_SERVER_TLS_CERT_FILE="/etc/ssl/certs/server.crt"

# SLURM configuration
export SLURM_EXPORTER_SLURM_BASE_URL="http://slurm-server:6820"
export SLURM_EXPORTER_SLURM_AUTH_TYPE="jwt"
export SLURM_EXPORTER_SLURM_AUTH_TOKEN="your-jwt-token"
export SLURM_EXPORTER_SLURM_AUTH_USERNAME="your-slurm-username"

# Collector configuration
export SLURM_EXPORTER_COLLECTORS_NODES_ENABLED="true"
export SLURM_EXPORTER_COLLECTORS_NODES_INTERVAL="60s"

# Logging configuration
export SLURM_EXPORTER_LOGGING_LEVEL="info"
export SLURM_EXPORTER_LOGGING_FORMAT="json"
```

### Special Environment Variables

```bash
# Configuration file path
export SLURM_EXPORTER_CONFIG="/etc/slurm-exporter/config.yaml"

# Enable debug mode
export SLURM_EXPORTER_DEBUG="true"

# Disable specific collectors
export SLURM_EXPORTER_DISABLE_COLLECTORS="qos,reservations"

# Enable specific collectors
export SLURM_EXPORTER_ENABLE_COLLECTORS="nodes,jobs,partitions"
```

## Configuration Examples

### Minimal Configuration

```yaml
# Minimal configuration for development
slurm:
  base_url: "http://localhost:6820"
```

### Development Configuration

```yaml
# Development configuration with debug logging
server:
  address: ":8080"

slurm:
  base_url: "http://slurm-dev:6820"
  auth:
    type: "basic"
    username: "dev-user"
    password: "dev-password"

collectors:
  global:
    default_interval: "10s"

logging:
  level: "debug"
  format: "text"
```

### Production Configuration

```yaml
# Production configuration with high availability
server:
  address: ":8080"
  read_timeout: "30s"
  write_timeout: "30s"
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/server.crt"
    key_file: "/etc/ssl/private/server.key"
    min_version: "1.2"
  basic_auth:
    enabled: true
    username: "prometheus"
    password: "secure-password-here"

slurm:
  base_url: "https://slurm-prod.company.com:6820"
  timeout: "30s"
  retry_attempts: 3
  retry_delay: "5s"
  auth:
    type: "jwt"
    username: "slurm-exporter"
    token_file: "/run/secrets/slurm-jwt-token"
  tls:
    ca_cert_file: "/etc/ssl/certs/company-ca.crt"

collectors:
  nodes:
    interval: "60s"
    timeout: "10s"
  jobs:
    interval: "30s"
    filters:
      job_states: ["RUNNING", "PENDING"]
  partitions:
    interval: "300s"
    filters:
      exclude_partitions: ["test", "debug"]

logging:
  level: "warn"
  format: "json"
  output: "file"
  file: "/var/log/slurm-exporter/exporter.log"
  max_size: 100
  max_backups: 5
  compress: true
  fields:
    environment: "production"
    datacenter: "us-east-1"

metrics:
  registry:
    enable_go_collector: false
    enable_process_collector: true
```

### Large Cluster Configuration

```yaml
# Configuration optimized for large SLURM clusters (10k+ nodes)
server:
  address: ":8080"
  read_timeout: "60s"
  write_timeout: "60s"
  max_request_size: 2097152  # 2MB

slurm:
  base_url: "https://slurm-large.company.com:6820"
  timeout: "120s"
  retry_attempts: 5
  retry_delay: "10s"

collectors:
  global:
    default_timeout: "110s"
  nodes:
    enabled: true
    interval: "300s"  # 5 minutes
    timeout: "240s"   # 4 minutes
  jobs:
    enabled: true
    interval: "60s"
    timeout: "50s"
    filters:
      job_states: ["RUNNING", "PENDING"]
  partitions:
    enabled: true
    interval: "600s"  # 10 minutes
    timeout: "300s"   # 5 minutes

logging:
  level: "info"
  format: "json"
  output: "file"
  file: "/var/log/slurm-exporter/exporter.log"
  max_size: 200
  max_backups: 10

metrics:
  registry:
    enable_go_collector: false
    enable_process_collector: false  # Reduce overhead

slurm:
  rate_limit:
    requests_per_second: 1000.0
    burst_size: 2000
```

### Multi-Cluster Configuration

```yaml
# Configuration for monitoring multiple SLURM clusters
# Note: This requires running multiple exporter instances

# Instance 1: Cluster A
slurm:
  base_url: "https://slurm-cluster-a.company.com:6820"
  auth:
    type: "jwt"
    username: "exporter-user"
    token_file: "/run/secrets/slurm-token-cluster-a"

metrics:
  subsystem: "cluster_a"

logging:
  fields:
    cluster: "cluster-a"
    region: "us-east"

---
# Instance 2: Cluster B (separate config file)
slurm:
  base_url: "https://slurm-cluster-b.company.com:6820"
  auth:
    type: "jwt"
    username: "exporter-user"
    token_file: "/run/secrets/slurm-token-cluster-b"

metrics:
  subsystem: "cluster_b"

logging:
  fields:
    cluster: "cluster-b"
    region: "eu-west"
```

## Best Practices

### Security Best Practices

1. **Always use TLS in production**
   ```yaml
   server:
     tls:
       enabled: true
       min_version: "1.2"
   ```

2. **Store sensitive data in secrets**
   ```yaml
   slurm:
     auth:
       token_file: "/run/secrets/slurm-token"  # Good
       # token: "hardcoded-token"               # Bad
   ```

3. **Enable authentication for metrics endpoint**
   ```yaml
   server:
     basic_auth:
       enabled: true
       username: "prometheus"
       password: "secure-password-here"
   ```

4. **Use least privilege principle**
   - Create dedicated SLURM user for exporter
   - Grant only necessary permissions
   - Rotate tokens regularly

### Performance Best Practices

1. **Tune collection intervals based on cluster size**
   - Small clusters (< 100 nodes): 30s intervals
   - Medium clusters (100-1000 nodes): 60s intervals
   - Large clusters (> 1000 nodes): 300s intervals

2. **Adjust concurrency for your environment**
   ```yaml
   collectors:
     nodes:
       max_concurrency: 2
     jobs:
       max_concurrency: 3
   ```

3. **Disable unnecessary collectors**
   ```yaml
   collectors:
     qos:
       enabled: false       # If not needed
     reservations:
       enabled: false       # If not needed
   ```

4. **Disable unnecessary registry metrics**
   ```yaml
   metrics:
     registry:
       enable_go_collector: false       # If not needed
       enable_process_collector: false   # If not needed
   ```

### Operational Best Practices

1. **Use structured logging in production**
   ```yaml
   logging:
     format: "json"
     fields:
       environment: "production"
       service: "slurm-exporter"
   ```

2. **Enable file logging with rotation**
   ```yaml
   logging:
     output: "file"
     file: "/var/log/slurm-exporter/exporter.log"
     max_size: 100
     max_backups: 5
     compress: true
   ```

3. **Monitor exporter health**
   - Set up alerts for exporter availability
   - Monitor scrape duration and errors
   - Track memory and CPU usage

4. **Plan for high availability**
   - Run multiple instances
   - Use load balancing
   - Implement proper health checks

### Configuration Management

1. **Use version control for configurations**
   ```bash
   git add config/production.yaml
   git commit -m "Update production configuration"
   ```

2. **Use environment-specific configurations**
   ```bash
   # Development
   slurm-exporter --config config/dev.yaml

   # Production
   slurm-exporter --config config/prod.yaml
   ```

3. **Document configuration changes**
   ```yaml
   # config.yaml
   # Last updated: 2024-01-01
   # Changed by: John Doe
   # Reason: Increased batch size for better performance
   ```

## Configuration Hot Reload

The exporter supports hot reloading of configuration without restart:

### Reloading Configuration

```bash
# Send SIGHUP to reload configuration
kill -HUP $(pidof slurm-exporter)

# Or use the API endpoint (if debug endpoints enabled)
curl -X POST http://localhost:8080/debug/config/reload

# Check reload status
curl http://localhost:8080/debug/config/status
```

### Hot-Reloadable Settings

The following settings can be reloaded without restarting the exporter:

**Hot-Reloadable:**
- Collector intervals and filters
- Logging level and format
- Cache TTL settings and size limits
- Circuit breaker thresholds
- Rate limiting rules
- Debug endpoint settings
- Health check configurations
- Metrics collection options

**Non-Reloadable (Requires Restart):**
- SLURM connection details (URL, authentication)
- Server address and port
- TLS configuration
- Performance limits (memory, cardinality)
- OpenTelemetry tracing configuration

### Reload Events

Monitor configuration reload events:

```bash
# Watch reload events in logs
tail -f /var/log/slurm-exporter.log | grep "config.reload"

# Get reload metrics
curl http://localhost:8080/metrics | grep slurm_config_reload
```

Available reload metrics:
- `slurm_config_reload_total` - Total number of reloads
- `slurm_config_reload_errors_total` - Number of failed reloads
- `slurm_config_reload_duration_seconds` - Time taken for reload

## Configuration Validation

### Common Validation Errors

1. **Invalid duration format**
   ```yaml
   # Wrong
   interval: 30  # Missing unit

   # Correct
   interval: "30s"
   ```

2. **Invalid URL format**
   ```yaml
   # Wrong
   base_url: "slurm-server:6820"  # Missing protocol

   # Correct
   base_url: "http://slurm-server:6820"
   ```

3. **JWT auth with username**
   ```yaml
   # This is correct - username IS used with JWT auth
   # It sets the X-SLURM-USER-NAME header required by SLURM REST API
   auth:
     type: "jwt"
     username: "slurm-user"
     token: "..."
   ```

For more information, see:
- [Installation Guide](installation.md)
- [Metrics Documentation](../website/docs/user-guide/metrics.md)
- [Troubleshooting Guide](troubleshooting.md)
