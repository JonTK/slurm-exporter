# SLURM Exporter Configuration Reference

This comprehensive guide covers all configuration options for SLURM Exporter, with examples and best practices for different deployment scenarios.

## Quick Configuration

**🚀 Use the Interactive Wizard:**
```bash
curl -s https://raw.githubusercontent.com/jontk/slurm-exporter/main/scripts/config-wizard.sh | bash
```

The configuration wizard will help you generate a production-ready configuration in under 2 minutes by:
- Selecting your deployment type (Kubernetes, Docker, Binary)
- Configuring SLURM API connection and authentication
- Choosing appropriate collectors for your use case
- Enabling advanced features based on your environment
- Generating deployment instructions and files

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
  
  # Read timeout for HTTP requests
  # Default: "30s"
  readTimeout: "30s"
  
  # Write timeout for HTTP responses
  # Default: "30s"
  writeTimeout: "30s"
  
  # Idle timeout for keep-alive connections
  # Default: "120s"
  idleTimeout: "120s"
  
  # Maximum header size in bytes
  # Default: 1048576 (1MB)
  maxHeaderBytes: 1048576
  
  # Enable pprof debug endpoints
  # Default: false
  enablePprof: false
  
  # Enable metrics endpoint
  # Default: true
  enableMetrics: true
  
  # Metrics path
  # Default: "/metrics"
  metricsPath: "/metrics"
  
  # Health check path
  # Default: "/health"
  healthPath: "/health"
  
  # Ready check path
  # Default: "/ready"
  readyPath: "/ready"
```

### TLS Configuration

```yaml
server:
  tls:
    # Enable TLS
    # Default: false
    enabled: true
    
    # Path to certificate file
    certFile: "/etc/ssl/certs/server.crt"
    
    # Path to private key file
    keyFile: "/etc/ssl/private/server.key"
    
    # Path to CA certificate file (for client verification)
    caFile: "/etc/ssl/certs/ca.crt"
    
    # Minimum TLS version
    # Options: "1.0", "1.1", "1.2", "1.3"
    # Default: "1.2"
    minVersion: "1.2"
    
    # Cipher suites (if empty, uses Go defaults)
    cipherSuites:
      - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
      - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
      - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
    
    # Enable client certificate verification
    # Default: false
    clientAuth: false
    
    # Client CA certificates for verification
    clientCAFile: "/etc/ssl/certs/client-ca.crt"
```

### HTTP Authentication

```yaml
server:
  auth:
    # Enable basic authentication for metrics endpoint
    # Default: false
    enabled: true
    
    # Authentication realm
    # Default: "SLURM Exporter"
    realm: "SLURM Exporter"
    
    # Users configuration
    users:
      - username: "prometheus"
        # Password can be plain text or bcrypt hash
        password: "$2a$10$Qx5rIu3gYSnAqBNlHAYMZ.nKro1I.KlPbHVlkwPGZeGxHPvNqUqRq"
      - username: "admin"
        password: "admin-password"
    
    # Path to htpasswd file (alternative to users list)
    htpasswdFile: "/etc/slurm-exporter/htpasswd"
```

## SLURM Connection Settings

### Basic Connection

```yaml
slurm:
  # SLURM REST API base URL
  # Required
  baseURL: "http://slurm-server:6820"
  
  # API version to use
  # Default: "v0.0.40"
  apiVersion: "v0.0.40"
  
  # Connection timeout
  # Default: "10s"
  connectionTimeout: "10s"
  
  # Request timeout
  # Default: "30s"
  requestTimeout: "30s"
  
  # Maximum idle connections
  # Default: 10
  maxIdleConns: 10
  
  # Maximum connections per host
  # Default: 10
  maxConnsPerHost: 10
  
  # Idle connection timeout
  # Default: "90s"
  idleConnTimeout: "90s"
  
  # Enable keep-alive
  # Default: true
  keepAlive: true
  
  # Keep-alive interval
  # Default: "30s"
  keepAliveInterval: "30s"
```

### Advanced Connection Settings

```yaml
slurm:
  # Retry configuration
  retry:
    # Enable automatic retries
    # Default: true
    enabled: true
    
    # Maximum number of retries
    # Default: 3
    maxRetries: 3
    
    # Initial retry delay
    # Default: "1s"
    initialDelay: "1s"
    
    # Maximum retry delay
    # Default: "30s"
    maxDelay: "30s"
    
    # Backoff multiplier
    # Default: 2.0
    multiplier: 2.0
    
    # Retry on these HTTP status codes
    # Default: [429, 502, 503, 504]
    retryStatusCodes:
      - 429  # Too Many Requests
      - 502  # Bad Gateway
      - 503  # Service Unavailable
      - 504  # Gateway Timeout
  
  # Circuit breaker configuration
  circuitBreaker:
    # Enable circuit breaker
    # Default: false
    enabled: true
    
    # Failure threshold before opening
    # Default: 5
    threshold: 5
    
    # Time window for failures
    # Default: "60s"
    window: "60s"
    
    # Time to wait before half-open
    # Default: "30s"
    timeout: "30s"
```

### TLS for SLURM Connection

```yaml
slurm:
  tls:
    # Enable TLS for SLURM connection
    # Default: false
    enabled: true
    
    # Skip certificate verification (insecure)
    # Default: false
    insecureSkipVerify: false
    
    # Client certificate for mutual TLS
    certFile: "/etc/ssl/certs/client.crt"
    keyFile: "/etc/ssl/private/client.key"
    
    # CA certificate for server verification
    caFile: "/etc/ssl/certs/slurm-ca.crt"
    
    # Server name for verification
    serverName: "slurm-server.example.com"
```

## Authentication Options

### JWT Token Authentication

```yaml
slurm:
  auth:
    # Authentication type
    # Options: "none", "jwt", "basic", "apikey"
    type: "jwt"
    
    # JWT token (direct value)
    token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
    
    # JWT token from file
    tokenFile: "/run/secrets/slurm-jwt-token"
    
    # JWT token from environment variable
    tokenEnv: "SLURM_JWT_TOKEN"
    
    # Token refresh configuration
    refresh:
      # Enable automatic token refresh
      # Default: false
      enabled: true
      
      # Refresh before expiry
      # Default: "5m"
      beforeExpiry: "5m"
      
      # Command to execute for token refresh
      command: "/usr/local/bin/refresh-slurm-token.sh"
      
      # Refresh interval (if not using expiry)
      # Default: "1h"
      interval: "1h"
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
    
    # Password from file
    passwordFile: "/run/secrets/slurm-password"
    
    # Password from environment variable
    passwordEnv: "SLURM_PASSWORD"
```

### API Key Authentication

```yaml
slurm:
  auth:
    type: "apikey"
    
    # API key header name
    # Default: "X-API-Key"
    headerName: "X-API-Key"
    
    # API key value
    apiKey: "your-api-key"
    
    # API key from file
    apiKeyFile: "/run/secrets/slurm-apikey"
    
    # API key from environment variable
    apiKeyEnv: "SLURM_API_KEY"
```

### Custom Authentication

```yaml
slurm:
  auth:
    type: "custom"
    
    # Custom headers to add to requests
    headers:
      X-Custom-Auth: "custom-value"
      X-Request-ID: "{{.RequestID}}"
    
    # Headers from environment variables
    headersFromEnv:
      X-Secret-Token: "SECRET_TOKEN_ENV"
```

## Collector Configuration

### Global Collector Settings

```yaml
collectors:
  # Global settings for all collectors
  global:
    # Default collection interval
    # Default: "30s"
    interval: "30s"
    
    # Default collection timeout
    # Default: "25s"
    timeout: "25s"
    
    # Namespace for metrics
    # Default: "slurm"
    namespace: "slurm"
    
    # Subsystem for metrics
    # Default: ""
    subsystem: ""
    
    # Enable all collectors by default
    # Default: true
    enabledByDefault: true
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
    # Default: "25s"
    timeout: "25s"
    
    # Metrics to collect
    metrics:
      # Cluster configuration metrics
      config: true
      
      # Controller statistics
      controller: true
      
      # License information
      licenses: true
      
      # Cluster totals
      totals: true
```

### Nodes Collector

```yaml
collectors:
  nodes:
    # Enable nodes collector
    # Default: true
    enabled: true
    
    # Collection interval
    # Default: "60s"
    interval: "60s"
    
    # Collection timeout
    # Default: "50s"
    timeout: "50s"
    
    # Batch size for node queries
    # Default: 100
    batchSize: 100
    
    # Concurrent requests
    # Default: 5
    concurrency: 5
    
    # Node states to include
    # Default: all states
    includeStates:
      - "idle"
      - "allocated"
      - "mixed"
      - "down"
      - "drain"
      - "draining"
    
    # Node metrics configuration
    metrics:
      # Basic node info
      info: true
      
      # CPU metrics
      cpu: true
      
      # Memory metrics
      memory: true
      
      # GPU metrics (if available)
      gpu: true
      
      # Network metrics
      network: false
      
      # Disk metrics
      disk: false
    
    # Label configuration
    labels:
      # Include partition as label
      partition: true
      
      # Include features as labels
      features: true
      
      # Include reason as label
      reason: true
      
      # Custom labels from node properties
      custom:
        - property: "Arch"
          label: "architecture"
        - property: "OS"
          label: "operating_system"
```

### Jobs Collector

```yaml
collectors:
  jobs:
    # Enable jobs collector
    # Default: true
    enabled: true
    
    # Collection interval
    # Default: "30s"
    interval: "30s"
    
    # Collection timeout
    # Default: "25s"
    timeout: "25s"
    
    # Batch size for job queries
    # Default: 500
    batchSize: 500
    
    # Job states to include
    # Default: ["RUNNING", "PENDING"]
    includeStates:
      - "RUNNING"
      - "PENDING"
      - "SUSPENDED"
      - "COMPLETING"
    
    # Job time window
    timeWindow:
      # Include jobs from this duration ago
      # Default: "0" (no limit)
      lookback: "24h"
      
      # Include future jobs (for pending)
      # Default: true
      includeFuture: true
    
    # Job metrics configuration
    metrics:
      # Job counts by state
      counts: true
      
      # Job wait time
      waitTime: true
      
      # Job runtime
      runtime: true
      
      # Resource usage
      resources: true
      
      # Array job metrics
      arrays: true
    
    # Label configuration
    labels:
      # Include user as label
      user: true
      
      # Include account as label
      account: true
      
      # Include partition as label
      partition: true
      
      # Include QoS as label
      qos: true
      
      # Maximum number of users to track
      # Default: 100
      maxUsers: 100
      
      # Maximum number of accounts to track
      # Default: 50
      maxAccounts: 50
```

### Partitions Collector

```yaml
collectors:
  partitions:
    # Enable partitions collector
    # Default: true
    enabled: true
    
    # Collection interval
    # Default: "120s"
    interval: "120s"
    
    # Collection timeout
    # Default: "60s"
    timeout: "60s"
    
    # Partitions to include (empty = all)
    includePartitions: []
    
    # Partitions to exclude
    excludePartitions:
      - "test"
      - "debug"
    
    # Partition metrics configuration
    metrics:
      # Node counts by state
      nodes: true
      
      # CPU metrics
      cpus: true
      
      # Memory metrics
      memory: true
      
      # Job metrics
      jobs: true
      
      # Limits and constraints
      limits: true
```

### Reservations Collector

```yaml
collectors:
  reservations:
    # Enable reservations collector
    # Default: false
    enabled: true
    
    # Collection interval
    # Default: "300s"
    interval: "300s"
    
    # Collection timeout
    # Default: "60s"
    timeout: "60s"
    
    # Reservation metrics
    metrics:
      # Active reservations
      active: true
      
      # Resource usage
      resources: true
      
      # Time metrics
      time: true
```

### QoS Collector

```yaml
collectors:
  qos:
    # Enable QoS collector
    # Default: false
    enabled: true
    
    # Collection interval
    # Default: "300s"
    interval: "300s"
    
    # QoS metrics
    metrics:
      # Limits and thresholds
      limits: true
      
      # Usage statistics
      usage: true
      
      # Priority information
      priority: true
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
```yaml
observability:
  tracing:
    enabled: true
    service_name: "slurm-exporter"
    service_version: "1.0.0"
    
    # OTLP exporter
    otlp:
      endpoint: "http://jaeger:14268/api/traces"
      headers:
        authorization: "Bearer ${TRACING_TOKEN}"
      timeout: 30s
      
    # Sampling configuration
    sampling:
      type: "ratio"                     # ratio, always, never
      ratio: 0.1                        # 10% sampling
      
    # Resource attributes
    resource:
      attributes:
        environment: "production"
        cluster.name: "hpc-cluster-01"
        datacenter: "us-west-2"
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
  # Options: "debug", "info", "warn", "error", "fatal"
  # Default: "info"
  level: "info"
  
  # Log format
  # Options: "text", "json"
  # Default: "json"
  format: "json"
  
  # Enable colored output (text format only)
  # Default: false
  color: false
  
  # Include caller information
  # Default: false
  caller: false
  
  # Timestamp format
  # Default: "2006-01-02T15:04:05.000Z07:00"
  timestampFormat: "2006-01-02T15:04:05.000Z07:00"
```

### File Logging

```yaml
logging:
  # Output to file
  file:
    # Enable file logging
    # Default: false
    enabled: true
    
    # Log file path
    path: "/var/log/slurm-exporter/exporter.log"
    
    # Maximum size in MB before rotation
    # Default: 100
    maxSize: 100
    
    # Maximum number of old files to keep
    # Default: 5
    maxBackups: 5
    
    # Maximum age in days
    # Default: 30
    maxAge: 30
    
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
    
  # Fields from environment variables
  fieldsFromEnv:
    pod_name: "HOSTNAME"
    namespace: "KUBERNETES_NAMESPACE"
```

## Advanced Configuration

### Metrics Configuration

```yaml
metrics:
  # Include Go runtime metrics
  # Default: true
  includeGoMetrics: true
  
  # Include process metrics
  # Default: true
  includeProcessMetrics: true
  
  # Metric name prefix
  # Default: "slurm"
  prefix: "slurm"
  
  # Histogram buckets configuration
  histograms:
    # Request duration buckets (seconds)
    requestDuration:
      - 0.001
      - 0.005
      - 0.01
      - 0.05
      - 0.1
      - 0.5
      - 1
      - 5
      - 10
    
    # Response size buckets (bytes)
    responseSize:
      - 100
      - 1000
      - 10000
      - 100000
      - 1000000
```

### Caching Configuration

```yaml
cache:
  # Enable caching
  # Default: true
  enabled: true
  
  # Cache backend
  # Options: "memory", "redis"
  # Default: "memory"
  backend: "memory"
  
  # Default TTL for cache entries
  # Default: "60s"
  defaultTTL: "60s"
  
  # Maximum cache size (memory backend)
  # Default: 1000
  maxSize: 1000
  
  # Cache key prefix
  # Default: "slurm-exporter"
  keyPrefix: "slurm-exporter"
  
  # Redis configuration (if backend is redis)
  redis:
    address: "localhost:6379"
    password: ""
    db: 0
    maxRetries: 3
    poolSize: 10
```

### Rate Limiting

```yaml
rateLimit:
  # Enable rate limiting
  # Default: false
  enabled: true
  
  # Requests per second
  # Default: 10
  requestsPerSecond: 10
  
  # Burst size
  # Default: 20
  burst: 20
  
  # Rate limit by
  # Options: "global", "ip", "user"
  # Default: "global"
  by: "ip"
  
  # Exclude certain IPs from rate limiting
  excludeIPs:
    - "127.0.0.1"
    - "10.0.0.0/8"
```

### Feature Flags

```yaml
features:
  # Enable experimental features
  experimental:
    # Enable experimental collectors
    collectors: false
    
    # Enable experimental metrics
    metrics: false
    
    # Enable experimental API endpoints
    api: false
  
  # Deprecated features (for backward compatibility)
  deprecated:
    # Enable v1 metrics format
    v1Metrics: false
    
    # Enable legacy authentication
    legacyAuth: false
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

# Collector configuration
export SLURM_EXPORTER_COLLECTORS_NODES_ENABLED="true"
export SLURM_EXPORTER_COLLECTORS_NODES_INTERVAL="60s"
export SLURM_EXPORTER_COLLECTORS_NODES_BATCH_SIZE="100"

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
  baseURL: "http://localhost:6820"
```

### Development Configuration

```yaml
# Development configuration with debug logging
server:
  address: ":8080"
  enablePprof: true

slurm:
  baseURL: "http://slurm-dev:6820"
  auth:
    type: "basic"
    username: "dev-user"
    password: "dev-password"

collectors:
  global:
    interval: "10s"
  nodes:
    batchSize: 10
  jobs:
    batchSize: 50

logging:
  level: "debug"
  format: "text"
  color: true
  caller: true
```

### Production Configuration

```yaml
# Production configuration with high availability
server:
  address: ":8080"
  readTimeout: "30s"
  writeTimeout: "30s"
  tls:
    enabled: true
    certFile: "/etc/ssl/certs/server.crt"
    keyFile: "/etc/ssl/private/server.key"
    minVersion: "1.2"
  auth:
    enabled: true
    htpasswdFile: "/etc/slurm-exporter/htpasswd"

slurm:
  baseURL: "https://slurm-prod.company.com:6820"
  connectionTimeout: "10s"
  requestTimeout: "30s"
  maxConnsPerHost: 20
  auth:
    type: "jwt"
    tokenFile: "/run/secrets/slurm-jwt-token"
    refresh:
      enabled: true
      beforeExpiry: "5m"
  retry:
    enabled: true
    maxRetries: 3
  tls:
    enabled: true
    caFile: "/etc/ssl/certs/company-ca.crt"

collectors:
  nodes:
    interval: "60s"
    timeout: "50s"
    batchSize: 200
    concurrency: 10
  jobs:
    interval: "30s"
    batchSize: 1000
    includeStates: ["RUNNING", "PENDING"]
  partitions:
    interval: "300s"
    excludePartitions: ["test", "debug"]

cache:
  enabled: true
  backend: "redis"
  redis:
    address: "redis-cluster:6379"
    poolSize: 20

logging:
  level: "warn"
  format: "json"
  file:
    enabled: true
    path: "/var/log/slurm-exporter/exporter.log"
    maxSize: 100
    maxBackups: 5
    compress: true
  fields:
    environment: "production"
    datacenter: "us-east-1"

metrics:
  includeGoMetrics: false
  includeProcessMetrics: true

rateLimit:
  enabled: true
  requestsPerSecond: 100
  burst: 200
  by: "ip"
```

### Large Cluster Configuration

```yaml
# Configuration optimized for large SLURM clusters (10k+ nodes)
server:
  address: ":8080"
  readTimeout: "60s"
  writeTimeout: "60s"
  maxHeaderBytes: 2097152  # 2MB

slurm:
  baseURL: "https://slurm-large.company.com:6820"
  connectionTimeout: "30s"
  requestTimeout: "120s"
  maxIdleConns: 100
  maxConnsPerHost: 100
  retry:
    enabled: true
    maxRetries: 5
    maxDelay: "60s"

collectors:
  global:
    timeout: "110s"
  nodes:
    enabled: true
    interval: "300s"  # 5 minutes
    timeout: "240s"   # 4 minutes
    batchSize: 500
    concurrency: 20
    metrics:
      info: true
      cpu: true
      memory: true
      gpu: false      # Disable if not needed
      network: false
      disk: false
  jobs:
    enabled: true
    interval: "60s"
    timeout: "50s"
    batchSize: 2000
    includeStates: ["RUNNING", "PENDING"]
    timeWindow:
      lookback: "1h"  # Only recent jobs
    labels:
      maxUsers: 500
      maxAccounts: 200
  partitions:
    enabled: true
    interval: "600s"  # 10 minutes
    timeout: "300s"   # 5 minutes

cache:
  enabled: true
  backend: "memory"
  maxSize: 10000
  defaultTTL: "300s"  # 5 minutes

logging:
  level: "info"
  format: "json"
  file:
    enabled: true
    path: "/var/log/slurm-exporter/exporter.log"
    maxSize: 200
    maxBackups: 10

metrics:
  includeGoMetrics: false
  includeProcessMetrics: false  # Reduce overhead

rateLimit:
  enabled: true
  requestsPerSecond: 1000
  burst: 2000
```

### Multi-Cluster Configuration

```yaml
# Configuration for monitoring multiple SLURM clusters
# Note: This requires running multiple exporter instances

# Instance 1: Cluster A
slurm:
  baseURL: "https://slurm-cluster-a.company.com:6820"
  auth:
    type: "jwt"
    tokenEnv: "SLURM_TOKEN_CLUSTER_A"

collectors:
  global:
    namespace: "slurm"
    subsystem: "cluster_a"

logging:
  fields:
    cluster: "cluster-a"
    region: "us-east"

---
# Instance 2: Cluster B
slurm:
  baseURL: "https://slurm-cluster-b.company.com:6820"
  auth:
    type: "jwt"
    tokenEnv: "SLURM_TOKEN_CLUSTER_B"

collectors:
  global:
    namespace: "slurm"
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
       minVersion: "1.2"
   ```

2. **Store sensitive data in secrets**
   ```yaml
   slurm:
     auth:
       tokenFile: "/run/secrets/slurm-token"  # Good
       # token: "hardcoded-token"             # Bad
   ```

3. **Enable authentication for metrics endpoint**
   ```yaml
   server:
     auth:
       enabled: true
       htpasswdFile: "/etc/slurm-exporter/htpasswd"
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

2. **Adjust batch sizes and concurrency**
   ```yaml
   collectors:
     nodes:
       batchSize: 200      # Increase for better performance
       concurrency: 10     # Increase for parallel processing
   ```

3. **Enable caching for stable data**
   ```yaml
   cache:
     enabled: true
     defaultTTL: "300s"
   ```

4. **Disable unnecessary metrics**
   ```yaml
   metrics:
     includeGoMetrics: false      # If not needed
     includeProcessMetrics: false # If not needed
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
     file:
       enabled: true
       maxSize: 100
       maxBackups: 5
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

2. **Validate configurations before deployment**
   ```bash
   slurm-exporter --config config.yaml --validate
   ```

3. **Use environment-specific configurations**
   ```bash
   # Development
   slurm-exporter --config config/dev.yaml
   
   # Production
   slurm-exporter --config config/prod.yaml
   ```

4. **Document configuration changes**
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

**✅ Hot-Reloadable:**
- Collector intervals and filters
- Logging level and format
- Cache TTL settings and size limits
- Circuit breaker thresholds
- Rate limiting rules
- Debug endpoint settings
- Health check configurations
- Metrics collection options

**❌ Non-Reloadable (Requires Restart):**
- SLURM connection details (URL, authentication)
- Server address and port
- TLS configuration
- Performance limits (memory, cardinality)
- OpenTelemetry tracing configuration

### Reload Configuration

```yaml
# Hot reload configuration
hotReload:
  # Enable configuration hot reload
  enabled: true
  
  # Watch configuration file for changes
  watchFile: true
  
  # Reload interval for file watching
  watchInterval: 30s
  
  # Validation on reload
  validateOnReload: true
  
  # Reload hooks
  hooks:
    preReload: "/usr/local/bin/pre-reload-hook.sh"
    postReload: "/usr/local/bin/post-reload-hook.sh"
```

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

### Command-Line Validation

```bash
# Validate configuration file
slurm-exporter --config config.yaml --validate

# Test configuration with dry-run
slurm-exporter --config config.yaml --dry-run

# Check specific collector configuration
slurm-exporter --config config.yaml --validate-collector nodes
```

### Configuration Schema

The exporter provides a JSON schema for validation:

```bash
# Get configuration schema
slurm-exporter --print-schema > config-schema.json

# Validate using schema
ajv validate -s config-schema.json -d config.yaml
```

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
   baseURL: "slurm-server:6820"  # Missing protocol
   
   # Correct
   baseURL: "http://slurm-server:6820"
   ```

3. **Conflicting settings**
   ```yaml
   # Wrong
   auth:
     type: "jwt"
     username: "user"  # Not used with JWT
   
   # Correct
   auth:
     type: "jwt"
     token: "..."
   ```

For more information, see:
- [Installation Guide](installation.md)
- [Metrics Documentation](metrics.md)
- [Troubleshooting Guide](troubleshooting.md)