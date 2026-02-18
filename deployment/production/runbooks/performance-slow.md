# Performance Issues Runbook

**Alert**: `SlurmExporterHighScrapeDuration`  
**Severity**: P1 Warning  
**Response Time**: < 30 minutes  

## Symptoms
- SLURM Exporter response times > 5 seconds (SLO breach)
- Slow metric collection affecting monitoring freshness
- Increased CPU/memory usage on exporter pods
- Timeouts in Prometheus scraping

## Immediate Actions (< 10 minutes)

### 1. Assess Impact
```bash
# Check current performance metrics
kubectl port-forward svc/slurm-exporter 10341:10341 -n slurm-exporter &
curl "http://localhost:10341/metrics" | grep slurm_exporter_scrape_duration

# Check Prometheus scrape health
kubectl exec prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'up{job="slurm-exporter"}'

# Review Grafana dashboard
# https://grafana.example.com/d/slurm-exporter-ops
```

### 2. Quick Resource Check
```bash
# Pod resource usage
kubectl top pods -n slurm-exporter

# Node resource usage
kubectl top nodes

# Check for resource constraints
kubectl describe pods -n slurm-exporter | grep -A 5 -B 5 "cpu\|memory"
```

## Investigation Procedures

### Performance Profiling
```bash
# Enable pprof endpoint if not already enabled
kubectl port-forward deployment/slurm-exporter 6060:6060 -n slurm-exporter &

# CPU profiling (30 seconds)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine analysis
curl http://localhost:6060/debug/pprof/goroutine?debug=1 > goroutines.txt
```

### Application-Level Analysis
```bash
# Check collection timing by collector
curl -s http://localhost:10341/metrics | grep -E "(duration_seconds|collection_errors)"

# Check cardinality metrics
curl -s http://localhost:10341/metrics | grep -E "(cardinality|series_count)"

# Application logs for performance patterns
kubectl logs deployment/slurm-exporter -n slurm-exporter --tail=100 | grep -E "(slow|timeout|error)"
```

### SLURM API Performance
```bash
# Test SLURM API response time
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  time curl -s "$SLURM_REST_URL/slurm/v0.0.44/ping"

# Check SLURM API load
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  curl -s "$SLURM_REST_URL/slurm/v0.0.44/jobs" | jq '.jobs | length'

# Test specific slow endpoints
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  time curl -s "$SLURM_REST_URL/slurm/v0.0.44/nodes" > /dev/null
```

### Infrastructure Analysis
```bash
# Network latency to SLURM cluster
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  ping -c 5 slurm-head.cluster.local

# DNS resolution time
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  time nslookup slurm-head.cluster.local

# Check for network throttling
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  iperf3 -c slurm-head.cluster.local -t 10 -P 1
```

## Common Causes and Solutions

### 1. High Cardinality
**Symptoms**: High memory usage, slow metric processing

**Investigation**:
```bash
# Check metric cardinality
curl -s http://localhost:10341/metrics | grep "^slurm_" | wc -l

# Find high-cardinality metrics
curl -s http://localhost:10341/metrics | grep "^slurm_" | cut -d'{' -f1 | sort | uniq -c | sort -nr | head -20

# Check cardinality limits
kubectl get configmap slurm-exporter-config -n slurm-exporter -o yaml | grep -A 10 cardinality
```

**Solutions**:
```bash
# Enable metric filtering
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "updated-config-with-filtering"
  }
}'

# Reduce cardinality limits
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "config-with-lower-cardinality-limits"
  }
}'

# Restart to apply changes
kubectl rollout restart deployment/slurm-exporter -n slurm-exporter
```

### 2. SLURM API Slowness
**Symptoms**: High API response times, SLURM timeouts

**Investigation**:
```bash
# Check SLURM cluster load
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  curl -s "$SLURM_REST_URL/slurm/v0.0.44/diag" | jq '.statistics'

# Monitor SLURM response patterns
for i in {1..10}; do
  kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
    time curl -s "$SLURM_REST_URL/slurm/v0.0.44/ping" 2>&1 | grep real
  sleep 1
done
```

**Solutions**:
```bash
# Increase collection intervals
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "config-with-longer-intervals"
  }
}'

# Enable connection pooling
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "config-with-connection-pooling"
  }
}'

# Add request rate limiting
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "config-with-rate-limiting"
  }
}'
```

### 3. Resource Constraints
**Symptoms**: CPU/memory limits hit, throttling

**Investigation**:
```bash
# Check resource usage vs limits
kubectl describe pods -n slurm-exporter | grep -A 10 "Limits\|Requests"

# Check throttling metrics
kubectl exec prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'container_cpu_cfs_throttled_seconds_total{pod=~"slurm-exporter-.*"}'

# Memory pressure indicators
kubectl exec prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'container_memory_working_set_bytes{pod=~"slurm-exporter-.*"}'
```

**Solutions**:
```bash
# Increase CPU limits
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "resources": {
            "limits": {
              "cpu": "1000m",
              "memory": "1Gi"
            },
            "requests": {
              "cpu": "200m",
              "memory": "256Mi"
            }
          }
        }]
      }
    }
  }
}'

# Scale horizontally
kubectl scale deployment slurm-exporter --replicas=5 -n slurm-exporter
```

### 4. Memory Leaks
**Symptoms**: Increasing memory usage over time, GC pressure

**Investigation**:
```bash
# Memory trend analysis
kubectl exec prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'increase(process_resident_memory_bytes{job="slurm-exporter"}[1h])'

# GC metrics
curl -s http://localhost:6060/debug/pprof/heap > heap.pprof
go tool pprof heap.pprof

# Memory allocation patterns
curl http://localhost:6060/debug/pprof/allocs?debug=1 > allocs.txt
```

**Solutions**:
```bash
# Enable memory optimization
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "config-with-memory-optimization"
  }
}'

# Restart pods to clear memory
kubectl rollout restart deployment/slurm-exporter -n slurm-exporter

# Implement memory limits with restart on limit
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "resources": {
            "limits": {
              "memory": "800Mi"
            }
          }
        }]
      }
    }
  }
}'
```

### 5. Database/Cache Performance
**Symptoms**: Slow cache lookups, high cache miss rates

**Investigation**:
```bash
# Cache performance metrics
curl -s http://localhost:10341/metrics | grep -E "(cache_hits|cache_misses|cache_size)"

# Cache configuration
kubectl get configmap slurm-exporter-config -n slurm-exporter -o yaml | grep -A 10 caching
```

**Solutions**:
```bash
# Increase cache size
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "config-with-larger-cache"
  }
}'

# Optimize cache TTL
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "config-with-optimized-ttl"
  }
}'
```

## Performance Optimization Actions

### Immediate Optimizations
```bash
# Enable performance mode configuration
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "performance-optimized-config"
  }
}'

# Scale up to handle load
kubectl scale deployment slurm-exporter --replicas=3 -n slurm-exporter

# Increase resource limits
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "resources": {
            "limits": {
              "cpu": "1",
              "memory": "1Gi"
            }
          }
        }]
      }
    }
  }
}'
```

### Configuration Tuning
```yaml
# Example performance configuration
collectors:
  jobs:
    collection_interval: 60s  # Increase from 30s
    timeout: 30s             # Increase timeout
    cardinality:
      max_labels: 3000       # Reduce cardinality
      sample_rate: 0.8       # Sample some metrics
  
  nodes:
    collection_interval: 120s # Less frequent collection
    timeout: 45s
    
performance:
  memory_optimizer:
    enabled: true
    gc_percent: 50           # More aggressive GC
    
  connection_pool:
    enabled: true
    max_connections: 5       # Limit concurrent connections
    
  caching:
    enabled: true
    default_ttl: 60s        # Longer cache TTL
    max_size: 2000          # Larger cache
```

## Monitoring and Prevention

### Enhanced Monitoring
```bash
# Add performance alerting
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: slurm-exporter-performance
  namespace: slurm-exporter
spec:
  groups:
  - name: performance
    rules:
    - alert: SlurmExporterSlowTrend
      expr: increase(slurm_exporter_scrape_duration_seconds[1h]) > 0.1
      for: 30m
      labels:
        severity: warning
      annotations:
        summary: "SLURM Exporter performance trending slower"
EOF
```

### Capacity Planning
```bash
# Performance trend analysis
kubectl exec prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'predict_linear(slurm_exporter_scrape_duration_seconds[1h], 3600)'

# Resource growth prediction
kubectl exec prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'predict_linear(process_resident_memory_bytes{job="slurm-exporter"}[1h], 86400)'
```

## Escalation Criteria

### Escalate to L2 if:
- Performance not improved within 1 hour
- Multiple performance issues occurring
- Infrastructure tuning required
- Application code optimization needed

### Escalate to L3 if:
- Architectural changes required
- SLURM cluster performance issues
- Hardware scaling needed
- SLA impact continues > 4 hours

## Communication Template

### Performance Alert
```
⚠️ PERFORMANCE: SLURM Exporter response time degraded
CURRENT: 95th percentile at [X]s (SLO: <5s)
IMPACT: Slower monitoring updates
INVESTIGATING: [Specific area]
ETA: [Estimated resolution time]
```

### Performance Update
```
🔧 UPDATE: SLURM Exporter performance optimization in progress
ACTIONS: [Optimizations applied]
IMPROVEMENT: [Current performance level]
NEXT: [Next optimization steps]
ETA: [Updated resolution time]
```

## Related Runbooks
- [Memory Issues](./memory-issues.md)
- [Cardinality Explosion](./cardinality-explosion.md)
- [SLURM Connectivity](./slurm-connectivity.md)
- [Capacity Planning](./capacity-planning.md)

## Prevention Actions
1. Regular performance testing
2. Capacity planning reviews
3. Performance benchmark automation
4. Configuration optimization guidelines
5. Proactive monitoring improvements