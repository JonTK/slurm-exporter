# Configuration Examples

This directory contains example configuration files for different deployment scenarios. Each configuration is optimized for specific use cases and environments.

## Configuration Files

### `config.yaml` - Default Configuration
The standard configuration file with balanced settings suitable for most deployments.

**Use case**: General purpose SLURM monitoring
**Characteristics**:
- 30-second collection intervals for most metrics
- 15-second intervals for jobs
- Moderate concurrency settings
- JSON logging to stdout
- No authentication or TLS (configure as needed)

### `production.yaml` - Production Environment
Enterprise-ready configuration with security, reliability, and performance optimizations.

**Use case**: Production SLURM clusters
**Characteristics**:
- Enhanced security (TLS, authentication)
- Higher concurrency and rate limits
- File-based logging with rotation
- Environment variable placeholders
- Extended timeouts and retry logic
- Comprehensive labeling

**Required environment variables**:
- `SLURM_REST_URL` - SLURM REST API endpoint
- `METRICS_PASSWORD` - Password for metrics endpoint
- `CLUSTER_NAME` - Cluster identifier
- `VERSION` - Application version

### `development.yaml` - Development Environment
Developer-friendly configuration with frequent collection and verbose logging.

**Use case**: Local development and testing
**Characteristics**:
- High-frequency collection (5-10 seconds)
- Debug-level logging in text format
- CORS enabled for development tools
- Insecure TLS settings
- Fast failure for quick feedback
- Lower resource limits

### `kubernetes.yaml` - Kubernetes Deployment
Container-optimized configuration for Kubernetes environments.

**Use case**: Containerized deployments in Kubernetes
**Characteristics**:
- Service discovery integration
- Secrets and ConfigMap integration
- Downward API for pod information
- Structured logging for log aggregation
- Health check endpoints
- Resource-aware settings

**Required environment variables**:
- `SLURM_REST_URL` - SLURM REST API endpoint
- `POD_NAMESPACE` - Kubernetes namespace (from downward API)
- `POD_NAME` - Pod name (from downward API)
- `NODE_NAME` - Node name (from downward API)
- `CLUSTER_NAME` - Cluster identifier

**Required mounted volumes**:
- `/etc/slurm-exporter/token` - JWT token (from secret)
- `/etc/ssl/certs/` - TLS certificates (from secret/configmap)

### `minimal.yaml` - Minimal Configuration
Lightweight configuration with only essential collectors enabled.

**Use case**: Resource-constrained environments or simple monitoring
**Characteristics**:
- Only cluster, nodes, and jobs collectors enabled
- 60-second intervals (30s for jobs)
- Minimal logging
- Low resource usage
- Essential metrics only

### `high-frequency.yaml` - Real-time Monitoring
Ultra-high-frequency configuration for real-time alerting and monitoring.

**Use case**: Critical environments requiring immediate detection of changes
**Characteristics**:
- 2-5 second collection intervals
- High concurrency and rate limits
- Fast failure and minimal retries
- Reduced logging to minimize overhead
- High cardinality limits
- Optimized for performance

**Warning**: This configuration generates significant load on the SLURM REST API and should only be used when necessary.

## Environment Variable Patterns

All configurations support environment variable overrides using the pattern:
```
SLURM_EXPORTER_<SECTION>_<FIELD>
```

Examples:
- `SLURM_EXPORTER_SERVER_ADDRESS=":9090"`
- `SLURM_EXPORTER_SLURM_BASE_URL="https://slurm.example.com:6820"`
- `SLURM_EXPORTER_LOGGING_LEVEL="debug"`
- `SLURM_EXPORTER_COLLECTORS_JOBS_INTERVAL="10s"`

## Security Considerations

### Production Deployments
- Always use TLS for the metrics endpoint in production
- Store sensitive values (tokens, passwords) in files rather than configuration
- Use least-privilege authentication for SLURM API access
- Regularly rotate authentication credentials
- Monitor for authentication failures and rate limiting

### Authentication Methods
- **JWT Token**: Recommended for production deployments
- **Basic Auth**: Simple but requires HTTPS
- **API Key**: Alternative to JWT for some SLURM configurations
- **None**: Only for development or internal networks

## Performance Tuning

### Collection Intervals
- **Jobs**: 15-30 seconds (jobs change frequently)
- **Nodes**: 30-60 seconds (hardware status changes less frequently)
- **Cluster/Partitions**: 60 seconds (configuration changes are rare)
- **Users**: 60-300 seconds (user metrics change slowly)

### Concurrency Settings
- Start with low concurrency (2-5) and increase based on SLURM API performance
- Monitor SLURM API response times and error rates
- Higher concurrency reduces collection time but increases API load

### Rate Limiting
- Configure based on SLURM API capacity
- Typical values: 10-20 requests/second for production
- Use burst settings 2-3x the sustained rate

## Troubleshooting

### Common Issues
1. **Connection timeouts**: Increase `slurm.timeout` and collector timeouts
2. **Rate limiting**: Reduce `rate_limit.requests_per_second` or increase intervals
3. **High memory usage**: Reduce `cardinality.max_series` or disable optional collectors
4. **Authentication failures**: Verify token/credentials and permissions

### Monitoring the Exporter
All configurations include system collector for self-monitoring:
- Collection duration metrics
- Error rate metrics
- Memory and CPU usage
- SLURM API response times

### Log Analysis
- Use structured logging (JSON) for production environments
- Monitor for ERROR and WARN level messages
- Track authentication and connection failures
- Watch for cardinality warnings

## Customization

### Adding Custom Labels
```yaml
collectors:
  jobs:
    labels:
      datacenter: "us-east-1"
      environment: "production"
```

### Filtering Specific Resources
```yaml
collectors:
  nodes:
    filters:
      exclude_nodes: ["maintenance-*", "test-*"]
      node_states: ["IDLE", "ALLOCATED", "MIXED"]
```

### Adjusting Error Handling
```yaml
collectors:
  jobs:
    error_handling:
      max_retries: 5
      retry_delay: "10s"
      backoff_factor: 2.0
      fail_fast: false
```