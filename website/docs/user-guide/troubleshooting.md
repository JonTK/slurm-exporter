# Troubleshooting Guide

This comprehensive guide helps diagnose and resolve common issues with SLURM Exporter.

## Quick Diagnostic Steps

When SLURM Exporter isn't working correctly, follow these initial steps:

1. **Check service status**
2. **Verify connectivity to SLURM**
3. **Examine logs for errors**
4. **Test metrics endpoint**

```bash
# Quick health check
curl -s http://localhost:9341/health | jq .

# Check if metrics are being collected
curl -s http://localhost:9341/metrics | grep -c slurm_

# View recent logs
sudo journalctl -u slurm-exporter -n 50 --no-pager
```

## Common Issues and Solutions

### 1. Service Won't Start

#### Symptoms
- Service fails to start
- Immediate exit after startup
- No metrics available on endpoint

#### Causes and Solutions

**Configuration File Issues:**

```bash
# Check if config file exists and is readable
ls -la /etc/slurm-exporter/config.yaml

# Validate configuration syntax
slurm-exporter --config.file=/etc/slurm-exporter/config.yaml --validate-config

# Test with minimal config
cat > /tmp/minimal-config.yaml << 'EOF'
slurm:
  host: "localhost"
  port: 6820
server:
  port: 9341
logging:
  level: "debug"
EOF

slurm-exporter --config.file=/tmp/minimal-config.yaml
```

**Permission Issues:**

```bash
# Fix file permissions
sudo chown -R slurm-exporter:slurm-exporter /etc/slurm-exporter
sudo chmod 640 /etc/slurm-exporter/config.yaml

# Check user exists
id slurm-exporter

# Fix systemd service permissions
sudo systemctl show slurm-exporter | grep User
```

**Port Conflicts:**

```bash
# Check if port is in use
sudo netstat -tlnp | grep 9341
sudo lsof -i :9341

# Use different port temporarily
slurm-exporter --web.listen-address=:9342
```

### 2. Cannot Connect to SLURM

#### Symptoms
- "Connection refused" errors
- Authentication failures
- Timeout errors

#### Diagnosis Steps

```bash
# Test basic connectivity
telnet slurm-controller.example.com 6820

# Test SLURM REST API directly
curl -k https://slurm-controller.example.com:6820/slurm/v0.0.40/ping

# Check with authentication
curl -k -H "X-SLURM-USER-NAME: your-user" \
     -H "X-SLURM-USER-TOKEN: your-token" \
     https://slurm-controller.example.com:6820/slurm/v0.0.40/ping
```

#### Common Solutions

**Network Connectivity:**

```yaml
# Update configuration with correct host/port
slurm:
  host: "slurm-controller.example.com"  # Verify DNS resolution
  port: 6820                            # Confirm port is correct
  timeout: 60s                          # Increase timeout
```

**Authentication Issues:**

```bash
# JWT Token problems
# 1. Verify token is valid
scontrol token

# 2. Check token file permissions
ls -la /etc/slurm-exporter/token
sudo chmod 600 /etc/slurm-exporter/token

# 3. Test token manually
export SLURM_JWT=$(cat /etc/slurm-exporter/token)
curl -k -H "X-SLURM-USER-TOKEN: $SLURM_JWT" \
     https://slurm-controller.example.com:6820/slurm/v0.0.40/jobs
```

**TLS/SSL Issues:**

```yaml
# Disable TLS verification temporarily for testing
slurm:
  tls:
    enabled: true
    verify: false  # Only for testing!

# Or provide proper certificates
slurm:
  tls:
    enabled: true
    verify: true
    ca_cert: "/etc/ssl/certs/ca-bundle.crt"
```

### 3. Missing or Incomplete Metrics

#### Symptoms
- Some metrics missing from `/metrics` endpoint
- Empty response from specific collectors
- Partial data in dashboards

#### Diagnosis

```bash
# Check which collectors are enabled
curl -s http://localhost:9341/metrics | grep slurm_exporter_collector_enabled

# Test specific collectors
curl -s http://localhost:9341/metrics?collector=jobs | grep slurm_job

# Check collection errors
curl -s http://localhost:9341/metrics | grep slurm_exporter_collect_errors_total
```

#### Solutions

**Enable Missing Collectors:**

```yaml
metrics:
  enabled_collectors:
    - jobs           # Basic job metrics
    - nodes          # Node status and resources
    - partitions     # Partition information
    - accounts       # Account usage (requires database access)
    - fairshare      # Fairshare data (requires database access)
    - qos            # QoS information (requires database access)
```

**Database Access Issues:**

```bash
# Test database connectivity
mysql -h slurm-db-host -u slurm_user -p slurm_acct_db

# Check SLURM accounting
sacct -X --format=JobID,JobName,User,State

# Verify exporter has database access
curl -s http://localhost:9341/health | jq '.checks.slurm_database'
```

**Collector-Specific Errors:**

```bash
# Enable debug logging
slurm-exporter --log.level=debug

# Check for specific error patterns
sudo journalctl -u slurm-exporter -f | grep -E "(ERROR|WARN)"
```

### 4. High Memory or CPU Usage

#### Symptoms
- Exporter consuming excessive resources
- Collection timeouts
- System performance degradation

#### Performance Analysis

```bash
# Monitor resource usage
top -p $(pgrep slurm-exporter)

# Check goroutine leaks
curl -s http://localhost:9341/debug/vars | jq .runtime.NumGoroutine

# Profile memory usage
curl -o heap.prof http://localhost:9341/debug/pprof/heap
go tool pprof heap.prof
```

#### Optimization Solutions

**Enable Caching:**

```yaml
performance:
  cache:
    enabled: true
    ttl: 300s              # Cache data for 5 minutes
    max_size: 10000        # Limit cache entries
```

**Adjust Collection Intervals:**

```yaml
metrics:
  collection_interval: 60s  # Reduce collection frequency

# Or use different intervals per collector
collectors:
  jobs:
    interval: 30s           # Frequent for job data
  nodes:
    interval: 300s          # Less frequent for node data
```

**Enable Batch Processing:**

```yaml
performance:
  batch_processing:
    enabled: true
    batch_size: 1000        # Process jobs in batches
    batch_timeout: 30s      # Timeout per batch
```

**Memory Management:**

```yaml
performance:
  memory:
    max_heap_size: 1000000000  # 1GB limit
    gc_target_percentage: 10    # Aggressive GC
```

### 5. Metrics Not Appearing in Prometheus

#### Symptoms
- Prometheus not scraping metrics
- Stale data in Prometheus
- Missing targets in Prometheus UI

#### Diagnosis

```bash
# Check Prometheus configuration
curl -s http://prometheus:9090/api/v1/targets | jq '.data.activeTargets[] | select(.job=="slurm-exporter")'

# Verify scrape endpoint
curl -s http://localhost:9341/metrics | head -20

# Check Prometheus logs
docker logs prometheus 2>&1 | grep slurm-exporter
```

#### Solutions

**Prometheus Configuration:**

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'slurm-exporter'
    static_configs:
      - targets: ['localhost:9341']
    scrape_interval: 30s
    scrape_timeout: 30s
    metrics_path: /metrics
```

**Service Discovery Issues:**

```yaml
# For Kubernetes
scrape_configs:
  - job_name: 'slurm-exporter'
    kubernetes_sd_configs:
      - role: endpoints
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        action: keep
        regex: slurm-exporter
```

**Network Connectivity:**

```bash
# Test from Prometheus container
docker exec prometheus curl -s http://slurm-exporter:9341/metrics

# Check firewall rules
sudo iptables -L | grep 9341
```

### 6. Dashboard Issues in Grafana

#### Symptoms
- Empty panels in Grafana
- "No data" messages
- Query errors

#### Diagnosis

```bash
# Test queries directly in Prometheus
curl -s 'http://prometheus:9090/api/v1/query?query=slurm_jobs_total'

# Check metric names and labels
curl -s 'http://prometheus:9090/api/v1/label/__name__/values' | grep slurm
```

#### Solutions

**Update Metric Names:**

Check if metric names have changed between versions:

```promql
# Old format
slurm_job_state{state="running"}

# New format (if changed)
slurm_jobs_running
```

**Fix Label Selectors:**

```promql
# Ensure labels exist
slurm_jobs_total{cluster="production"}

# Check available labels
{__name__=~"slurm_.*"}
```

**Datasource Configuration:**

```bash
# Verify Grafana can reach Prometheus
curl -s http://prometheus:9090/api/v1/status/config
```

### 7. Authentication and Authorization Issues

#### Symptoms
- 401 Unauthorized errors
- 403 Forbidden responses
- JWT token rejected

#### Diagnosis and Solutions

**JWT Token Issues:**

```bash
# Generate new token
scontrol token username=your-user

# Check token expiration
echo "your-jwt-token" | base64 -d | jq .exp

# Verify token format
echo "your-jwt-token" | base64 -d | jq .
```

**Certificate Authentication:**

```bash
# Check certificate validity
openssl x509 -in /etc/slurm-exporter/client.crt -text -noout

# Verify certificate matches key
openssl rsa -noout -modulus -in /etc/slurm-exporter/client.key | openssl md5
openssl x509 -noout -modulus -in /etc/slurm-exporter/client.crt | openssl md5
```

**Basic Authentication:**

```yaml
# Test basic auth configuration
slurm:
  auth:
    type: "basic"
    user: "slurm_user"
    password: "password"
```

### 8. Performance Bottlenecks

#### Symptoms
- Slow metric collection
- Timeouts in Prometheus scrapes
- High latency in SLURM API calls

#### Optimization Strategies

**Connection Pooling:**

```yaml
performance:
  connection_pool:
    enabled: true
    max_connections: 10
    max_idle_connections: 5
    connection_lifetime: 300s
```

**Parallel Collection:**

```yaml
metrics:
  collection:
    parallel: true
    max_concurrent: 5     # Limit concurrent collectors
```

**Selective Collection:**

```yaml
metrics:
  filtering:
    enabled: true
    exclude_metrics:
      - "slurm_node_tres_.*"      # High cardinality metrics
    include_partitions: ["compute", "gpu"]  # Only important partitions
```

## Advanced Troubleshooting

### Using Debug Endpoints

Enable debug endpoints for detailed diagnostics:

```yaml
development:
  enabled: true
  debug_endpoints:
    enabled: true
```

**Available Debug Endpoints:**

```bash
# Runtime variables
curl -s http://localhost:9341/debug/vars | jq .

# Current configuration
curl -s http://localhost:9341/debug/config | jq .

# Profiling data
curl -o cpu.prof http://localhost:9341/debug/pprof/profile?seconds=30
go tool pprof cpu.prof
```

### Log Analysis

**Enable Structured Logging:**

```yaml
logging:
  level: "debug"
  format: "json"
  output: "/var/log/slurm-exporter/debug.log"
```

**Common Log Patterns:**

```bash
# Connection errors
grep "connection refused\|timeout\|network" /var/log/slurm-exporter/*.log

# Authentication errors
grep "401\|403\|unauthorized\|forbidden" /var/log/slurm-exporter/*.log

# Collection errors
grep "collect_error\|collection_failed" /var/log/slurm-exporter/*.log

# Performance issues
grep "slow\|timeout\|duration" /var/log/slurm-exporter/*.log
```

### Performance Profiling

**CPU Profiling:**

```bash
# Collect 30-second CPU profile
curl -o cpu.prof http://localhost:9341/debug/pprof/profile?seconds=30

# Analyze with pprof
go tool pprof cpu.prof
(pprof) top
(pprof) web
```

**Memory Profiling:**

```bash
# Collect heap profile
curl -o heap.prof http://localhost:9341/debug/pprof/heap

# Analyze memory usage
go tool pprof heap.prof
(pprof) top
(pprof) list main.main
```

**Goroutine Analysis:**

```bash
# Check for goroutine leaks
curl -o goroutine.prof http://localhost:9341/debug/pprof/goroutine

go tool pprof goroutine.prof
(pprof) top
(pprof) traces
```

## Environment-Specific Issues

### Docker/Container Issues

**Container Networking:**

```bash
# Check container connectivity
docker exec slurm-exporter curl -s http://localhost:9341/health

# Network troubleshooting
docker network ls
docker exec slurm-exporter nslookup slurm-controller
```

**Volume Mounts:**

```bash
# Verify config file is mounted correctly
docker exec slurm-exporter cat /etc/slurm-exporter/config.yaml

# Check file permissions in container  
docker exec slurm-exporter ls -la /etc/slurm-exporter/
```

### Kubernetes Issues

**Pod Status:**

```bash
# Check pod status
kubectl get pods -l app=slurm-exporter

# View pod logs
kubectl logs -l app=slurm-exporter --tail=100

# Describe pod for events
kubectl describe pod slurm-exporter-xxx
```

**Service and Ingress:**

```bash
# Check service endpoints
kubectl get endpoints slurm-exporter

# Test service internally
kubectl run debug --image=busybox -it --rm -- wget -qO- http://slurm-exporter:9341/health
```

**ConfigMap and Secrets:**

```bash
# Verify ConfigMap
kubectl get configmap slurm-exporter-config -o yaml

# Check mounted secrets
kubectl exec slurm-exporter-xxx -- ls -la /etc/slurm-exporter/
```

## Recovery Procedures

### Service Recovery

```bash
# Restart service
sudo systemctl restart slurm-exporter

# Clear any stuck processes
sudo pkill -f slurm-exporter
sudo systemctl start slurm-exporter

# Reset to known good configuration
sudo cp /etc/slurm-exporter/config.yaml.backup /etc/slurm-exporter/config.yaml
sudo systemctl restart slurm-exporter
```

### Cache Reset

```bash
# Clear cache directory
sudo rm -rf /var/lib/slurm-exporter/cache/*
sudo systemctl restart slurm-exporter

# Disable cache temporarily
# Edit config: performance.cache.enabled = false
```

### Configuration Reset

```bash
# Backup current config
sudo cp /etc/slurm-exporter/config.yaml /etc/slurm-exporter/config.yaml.$(date +%s)

# Restore from package defaults
sudo cp /usr/share/doc/slurm-exporter/examples/config.yaml /etc/slurm-exporter/

# Validate and restart
slurm-exporter --config.file=/etc/slurm-exporter/config.yaml --validate-config
sudo systemctl restart slurm-exporter
```

## Getting Help

### Information to Collect

When reporting issues, provide:

1. **Version information:**
   ```bash
   slurm-exporter --version
   ```

2. **Configuration (sanitized):**
   ```bash
   # Remove sensitive data before sharing
   cat /etc/slurm-exporter/config.yaml
   ```

3. **Logs with debug level:**
   ```bash
   sudo journalctl -u slurm-exporter -n 100 --no-pager
   ```

4. **System information:**
   ```bash
   uname -a
   systemctl --version
   ```

5. **SLURM version:**
   ```bash
   scontrol version
   ```

### Health Check Script

Create a comprehensive health check:

```bash
#!/bin/bash
# slurm-exporter-healthcheck.sh

echo "=== SLURM Exporter Health Check ==="
echo "Timestamp: $(date)"
echo

echo "=== Service Status ==="
systemctl status slurm-exporter --no-pager
echo

echo "=== Process Information ==="
ps aux | grep slurm-exporter | grep -v grep
echo

echo "=== Network Ports ==="
netstat -tlnp | grep 9341
echo

echo "=== Metrics Endpoint ==="
curl -s -m 5 http://localhost:9341/health | jq . 2>/dev/null || echo "Health endpoint not responding"
echo

echo "=== Metrics Count ==="
curl -s -m 5 http://localhost:9341/metrics | grep -c "^slurm_" || echo "Metrics not available"
echo

echo "=== Recent Errors ==="
journalctl -u slurm-exporter -n 10 --no-pager | grep -i error
echo

echo "=== SLURM Connectivity ==="
telnet slurm-controller 6820 < /dev/null && echo "SLURM reachable" || echo "SLURM unreachable"
```

### Support Channels

- **GitHub Issues**: [Repository Issues](https://github.com/jontk/slurm-exporter/issues)
- **Documentation**: [Project Documentation](https://jontk.github.io/slurm-exporter)
- **Community**: [Discussions](https://github.com/jontk/slurm-exporter/discussions)

### Escalation Process

1. **Check documentation** and common issues
2. **Run health check script** and collect diagnostics
3. **Search existing issues** for similar problems
4. **Create detailed issue** with all collected information
5. **Follow up** with additional information as requested