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
curl -s http://localhost:10341/health | jq .

# Check readiness
curl -s http://localhost:10341/ready

# Check if metrics are being collected
curl -s http://localhost:10341/metrics | grep -c slurm_

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

# Test with minimal config
cat > /tmp/minimal-config.yaml << 'EOF'
server:
  address: ":10341"
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"

slurm:
  base_url: "http://localhost:6820"
  api_version: "v0.0.44"
  auth:
    type: "none"

logging:
  level: "debug"
  format: "text"
EOF

slurm-exporter --config=/tmp/minimal-config.yaml
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
sudo netstat -tlnp | grep 10341
sudo lsof -i :10341

# Use different port via CLI flag
slurm-exporter --config=/etc/slurm-exporter/config.yaml --addr=:9090
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
curl -k https://slurm-controller.example.com:6820/slurm/v0.0.44/ping

# Check with authentication
curl -k -H "X-SLURM-USER-NAME: your-user" \
     -H "X-SLURM-USER-TOKEN: your-token" \
     https://slurm-controller.example.com:6820/slurm/v0.0.44/ping
```

#### Common Solutions

**Network Connectivity:**

```yaml
# Update configuration with correct base_url
slurm:
  base_url: "http://slurm-controller.example.com:6820"
  timeout: 60s                          # Increase timeout
  retry_attempts: 5
  retry_delay: 10s
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
     https://slurm-controller.example.com:6820/slurm/v0.0.44/jobs
```

**TLS/SSL Issues:**

```yaml
# Allow insecure connections temporarily for testing
validation:
  allow_insecure_connections: true

slurm:
  tls:
    insecure_skip_verify: true  # Only for testing!

# Or provide proper certificates
slurm:
  tls:
    insecure_skip_verify: false
    ca_cert_file: "/etc/ssl/certs/ca-bundle.crt"
```

### 3. Missing or Incomplete Metrics

#### Symptoms
- Some metrics missing from `/metrics` endpoint
- Empty response from specific collectors
- Partial data in dashboards

#### Diagnosis

```bash
# Check which metrics are available
curl -s http://localhost:10341/metrics | grep slurm_exporter

# Check collection errors
curl -s http://localhost:10341/metrics | grep slurm_exporter_collect_errors_total
```

#### Solutions

**Enable Missing Collectors:**

```yaml
collectors:
  jobs:
    enabled: true
  nodes:
    enabled: true
  partitions:
    enabled: true
  users:
    enabled: true
  qos:
    enabled: true
  reservations:
    enabled: true
  cluster:
    enabled: true
  system:
    enabled: true
```

**Collector-Specific Errors:**

```bash
# Enable debug logging
slurm-exporter --config=/etc/slurm-exporter/config.yaml --log-level=debug

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
```

#### Optimization Solutions

**Adjust Collection Intervals:**

```yaml
collectors:
  jobs:
    interval: 30s           # Reduce from 15s default
  nodes:
    interval: 60s           # Less frequent for node data
  partitions:
    interval: 120s
```

**Enable Filtering:**

```yaml
collectors:
  nodes:
    filters:
      include_partitions: ["compute", "gpu"]  # Only important partitions
      metrics:
        exclude_metrics:
          - "slurm_node_tres_.*"      # High cardinality metrics
```

**Adjust Cardinality Limits:**

```yaml
metrics:
  cardinality:
    max_series: 5000        # Reduce max series
    warn_limit: 4000
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
curl -s http://localhost:10341/metrics | head -20

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
      - targets: ['localhost:10341']
    scrape_interval: 30s
    scrape_timeout: 30s
    metrics_path: /metrics
```

**Network Connectivity:**

```bash
# Test from Prometheus container
docker exec prometheus curl -s http://slurm-exporter:10341/metrics

# Check firewall rules
sudo iptables -L | grep 10341
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

# Verify token format
echo "your-jwt-token" | base64 -d | jq .
```

### 8. API Version Mismatch

#### Symptoms
- 404 errors from SLURM REST API
- "unsupported API version" in logs

#### Solutions

Verify your SLURM REST API version and set the correct version in config:

```yaml
slurm:
  # Supported versions: v0.0.40, v0.0.41, v0.0.42, v0.0.43, v0.0.44
  api_version: "v0.0.44"
```

```bash
# Check which API versions your SLURM supports
curl http://slurm-controller:6820/openapi/v3
```

## Environment-Specific Issues

### Docker/Container Issues

**Container Networking:**

```bash
# Check container connectivity
docker exec slurm-exporter curl -s http://localhost:10341/health

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
kubectl run debug --image=busybox -it --rm -- wget -qO- http://slurm-exporter:10341/health
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
netstat -tlnp | grep 10341
echo

echo "=== Health Endpoint ==="
curl -s -m 5 http://localhost:10341/health | jq . 2>/dev/null || echo "Health endpoint not responding"
echo

echo "=== Ready Endpoint ==="
curl -s -m 5 http://localhost:10341/ready || echo "Ready endpoint not responding"
echo

echo "=== Metrics Count ==="
curl -s -m 5 http://localhost:10341/metrics | grep -c "^slurm_" || echo "Metrics not available"
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
