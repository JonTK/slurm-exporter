# Performance Tuning Guide for Large Clusters

This guide provides recommendations for optimizing the SLURM exporter performance in large-scale HPC environments with thousands of nodes and high job throughput.

## Table of Contents

- [Performance Considerations](#performance-considerations)
- [Scaling Strategies](#scaling-strategies)
- [Configuration Optimization](#configuration-optimization)
- [Resource Management](#resource-management)
- [Metric Optimization](#metric-optimization)
- [High Availability Setup](#high-availability-setup)
- [Monitoring Performance](#monitoring-performance)
- [Troubleshooting Performance Issues](#troubleshooting-performance-issues)

## Performance Considerations

### Cluster Size Impact

| Cluster Size | Nodes | Jobs/Hour | Recommended Strategy |
|--------------|-------|-----------|---------------------|
| Small | <100 | <1,000 | Single exporter, default config |
| Medium | 100-1,000 | 1,000-10,000 | Single exporter, optimized config |
| Large | 1,000-5,000 | 10,000-50,000 | Multiple exporters, selective collection |
| Enterprise | >5,000 | >50,000 | Federated approach, dedicated instances |

### Performance Bottlenecks

1. **SLURM REST API response time**
   - Large node lists (>1000 nodes)
   - High job throughput (>10,000 jobs/hour)
   - Complex accounting queries

2. **Network latency**
   - Geographic distribution
   - Network congestion
   - Firewall/proxy overhead

3. **Memory consumption**
   - High metric cardinality
   - Large time series retention
   - Inefficient label usage

4. **CPU utilization**
   - JSON parsing overhead
   - Prometheus metric calculation
   - Concurrent collection processing

## Scaling Strategies

### Horizontal Scaling

#### Multiple Exporter Instances

Deploy specialized exporters for different metric categories:

```yaml
# cluster-exporter.yaml
collectors:
  cluster:
    enabled: true
    interval: 30s
  nodes:
    enabled: false
  jobs:
    enabled: false
  users:
    enabled: false

collection:
  intervals:
    cluster: 30s

metrics:
  prefix: "slurm_cluster_"
```

```yaml
# jobs-exporter.yaml
collectors:
  cluster:
    enabled: false
  nodes:
    enabled: false
  jobs:
    enabled: true
    max_jobs: 5000        # Limit job collection
    states: ["running", "pending"]  # Only active jobs
  users:
    enabled: false

collection:
  intervals:
    jobs: 60s             # Less frequent updates

metrics:
  prefix: "slurm_jobs_"
```

```yaml
# nodes-exporter.yaml
collectors:
  cluster:
    enabled: false
  nodes:
    enabled: true
    partitions: ["gpu", "compute"]  # Specific partitions only
  jobs:
    enabled: false
  users:
    enabled: false

collection:
  intervals:
    nodes: 15s

metrics:
  prefix: "slurm_nodes_"
```

#### Federated Prometheus Setup

```yaml
# prometheus-federation.yaml
global:
  scrape_interval: 15s

scrape_configs:
  # Cluster-level metrics
  - job_name: 'slurm-cluster'
    static_configs:
      - targets: ['cluster-exporter:8080']
    scrape_interval: 30s

  # Node metrics  
  - job_name: 'slurm-nodes'
    static_configs:
      - targets: ['nodes-exporter:8080']
    scrape_interval: 15s

  # Job metrics
  - job_name: 'slurm-jobs'
    static_configs:
      - targets: ['jobs-exporter:8080']
    scrape_interval: 60s

  # Federation from regional Prometheus instances
  - job_name: 'federation'
    scrape_interval: 15s
    honor_labels: true
    metrics_path: '/federate'
    params:
      'match[]':
        - '{job=~"slurm.*"}'
    static_configs:
      - targets:
        - 'prometheus-region1:9090'
        - 'prometheus-region2:9090'
```

### Vertical Scaling

#### Resource Allocation

```yaml
# Large cluster deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slurm-exporter
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: slurm-exporter
        image: slurm-exporter:latest
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 4Gi
        env:
        - name: GOGC
          value: "100"
        - name: GOMEMLIMIT
          value: "3GiB"
```

## Configuration Optimization

### Collection Intervals

```yaml
# Optimized for large clusters
collection:
  intervals:
    cluster: 60s          # Cluster info changes infrequently
    nodes: 30s            # Balance between freshness and load
    jobs: 120s            # Jobs can be collected less frequently
    partitions: 300s      # Partitions rarely change
    users: 300s           # User stats for longer-term analysis

  # Stagger collection start times to spread load
  jitter: 10s
  
  # Timeout settings for large clusters
  timeout: 60s
  
  # Retry configuration
  retry_attempts: 3
  retry_delay: 10s
  retry_backoff: 2.0
```

### Selective Collection

```yaml
# Reduce collection scope
collectors:
  nodes:
    enabled: true
    # Only collect specific partitions
    partitions: ["gpu", "bigmem", "compute"]
    # Skip detailed node features
    include_features: false
    # Limit node states collected
    states: ["idle", "allocated", "down", "drain"]
    
  jobs:
    enabled: true
    # Limit job collection
    max_jobs: 10000
    # Only active jobs
    states: ["running", "pending", "completing"]
    # Skip completed jobs older than 1 hour
    max_age: "1h"
    # Limit job details
    include_steps: false
    
  users:
    enabled: true
    # Top users only
    max_users: 500
    # Exclude system users
    exclude_users: ["root", "slurm", "nobody"]
```

### Connection Optimization

```yaml
slurm:
  # Connection pooling
  max_connections: 20
  max_idle_connections: 10
  connection_lifetime: 300s
  keep_alive: true
  
  # Timeout settings
  connect_timeout: 10s
  request_timeout: 60s
  
  # HTTP/2 settings
  http2: true
  max_frame_size: 1048576
  
  # Compression
  compression: true
  
  # Circuit breaker
  circuit_breaker:
    enabled: true
    failure_threshold: 5
    reset_timeout: 30s
    success_threshold: 3
```

## Resource Management

### Memory Optimization

#### Metric Cardinality Management

```yaml
metrics:
  # Limit total series
  max_series: 100000
  
  # Label cardinality limits
  label_limits:
    user: 1000            # Top 1000 users
    partition: 50         # Maximum partitions
    node: 5000           # All nodes for large cluster
    job_name: 2000       # Top job names
    account: 500         # Top accounts
    
  # Metric retention
  retention:
    high_frequency: 24h   # Detailed metrics
    medium_frequency: 7d  # Aggregated metrics  
    low_frequency: 30d    # Summary metrics
```

#### Go Runtime Tuning

```bash
# Environment variables for large clusters
export GOGC=50                    # More aggressive GC
export GOMEMLIMIT=8GiB           # Set memory limit
export GOMAXPROCS=4              # Limit CPU cores
export GODEBUG=gctrace=1         # GC debugging
```

### CPU Optimization

#### Concurrent Collection

```yaml
collection:
  # Parallel collection workers
  max_workers: 4
  
  # Queue settings
  queue_size: 1000
  
  # Collection batching
  batch_size: 100
  batch_timeout: 5s
  
  # CPU throttling prevention
  cpu_throttle_threshold: 0.8
  backoff_on_throttle: true
```

#### Efficient Data Processing

```yaml
processing:
  # Streaming JSON parsing
  streaming_parser: true
  
  # Memory pooling
  buffer_pool_size: 100
  max_buffer_size: 1048576
  
  # Compression for large responses
  response_compression: true
  
  # Caching frequently accessed data
  cache:
    enabled: true
    ttl: 300s
    max_size: 1000
```

## Metric Optimization

### Efficient Metric Design

#### Reduce Label Dimensions

```yaml
# Instead of high-cardinality labels
# BAD: slurm_job_info{user="user1", account="proj1", partition="gpu", qos="high", job_name="simulation_run_12345"}

# Use aggregated metrics with selective labels
# GOOD: slurm_user_jobs_total{user="user1", state="running"}
# GOOD: slurm_partition_utilization{partition="gpu"}
# GOOD: slurm_account_usage{account="proj1"}

metrics:
  # Enable metric aggregation
  aggregation:
    enabled: true
    levels: ["cluster", "partition", "user", "account"]
    
  # Sampling for high-frequency metrics
  sampling:
    jobs:
      enabled: true
      rate: 0.1           # Sample 10% of jobs
      stratified: true    # Stratify by partition/state
      
  # Histogram optimization
  histograms:
    job_duration_buckets: [60, 300, 1800, 3600, 14400, 86400]
    memory_usage_buckets: [1e6, 1e7, 1e8, 1e9, 1e10, 1e11]
```

#### Metric Hierarchies

```yaml
metrics:
  levels:
    # Level 1: Critical (5s intervals)
    critical:
      - slurm_cluster_health
      - slurm_controller_status
      - slurm_node_count_by_state
      
    # Level 2: Important (30s intervals)  
    important:
      - slurm_cluster_utilization
      - slurm_job_queue_length
      - slurm_partition_utilization
      
    # Level 3: Detailed (5m intervals)
    detailed:
      - slurm_node_details
      - slurm_job_details
      - slurm_user_statistics
```

### Time Series Optimization

```yaml
# Prometheus configuration for large clusters
global:
  scrape_interval: 30s
  evaluation_interval: 30s
  
  # Reduce memory usage
  external_labels:
    cluster: 'main'
    
# Storage optimization
storage:
  tsdb:
    retention.time: 15d
    retention.size: 500GB
    min-block-duration: 2h
    max-block-duration: 36h
    
    # Compaction settings
    wal-compression: true
    head-chunks-write-queue-size: 1000
```

## High Availability Setup

### Active-Passive Configuration

```yaml
# Primary exporter
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slurm-exporter-primary
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: slurm-exporter
        image: slurm-exporter:latest
        env:
        - name: HA_MODE
          value: "primary"
        - name: LEADER_ELECTION
          value: "true"
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

```yaml
# Secondary exporter (standby)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slurm-exporter-secondary
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: slurm-exporter
        image: slurm-exporter:latest
        env:
        - name: HA_MODE
          value: "secondary"
        - name: LEADER_ELECTION
          value: "true"
        - name: STANDBY_INTERVAL
          value: "60s"
```

### Load Balancing

```yaml
# Service with load balancing
apiVersion: v1
kind: Service
metadata:
  name: slurm-exporter
spec:
  selector:
    app: slurm-exporter
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
  
  # Session affinity for consistent scraping
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800
```

## Monitoring Performance

### Exporter Self-Monitoring

```yaml
# Performance metrics configuration
metrics:
  exporter:
    # Collection performance
    collection_duration_seconds:
      enabled: true
      buckets: [0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0]
      
    # API performance
    api_request_duration_seconds:
      enabled: true
      buckets: [0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0]
      
    # Memory usage
    memory_usage_bytes:
      enabled: true
      
    # Cache performance
    cache_hit_ratio:
      enabled: true
      
    # Error rates
    collection_errors_total:
      enabled: true
      by_collector: true
      by_error_type: true
```

### Alerting for Performance

```yaml
# Performance monitoring alerts
groups:
  - name: slurm_exporter_performance
    rules:
    - alert: SlowCollectionPerformance
      expr: histogram_quantile(0.95, rate(slurm_exporter_collection_duration_seconds_bucket[5m])) > 30
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "Slow metric collection performance"
        description: "95th percentile collection time is {{ $value }}s"
        
    - alert: HighMemoryUsage
      expr: slurm_exporter_memory_usage_bytes > 2e9  # 2GB
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High memory usage"
        description: "Memory usage is {{ $value | humanizeBytes }}"
        
    - alert: CollectionErrors
      expr: rate(slurm_exporter_collection_errors_total[5m]) > 0.1
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High collection error rate"
        description: "Error rate is {{ $value | humanizePercentage }}"
```

### Performance Dashboard

```json
{
  "dashboard": {
    "title": "SLURM Exporter Performance",
    "panels": [
      {
        "title": "Collection Duration",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(slurm_exporter_collection_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(slurm_exporter_collection_duration_seconds_bucket[5m]))",
            "legendFormat": "Median"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "targets": [
          {
            "expr": "slurm_exporter_memory_usage_bytes",
            "legendFormat": "Memory Usage"
          }
        ]
      },
      {
        "title": "API Request Rate",
        "targets": [
          {
            "expr": "rate(slurm_exporter_api_requests_total[5m])",
            "legendFormat": "Requests/sec"
          }
        ]
      }
    ]
  }
}
```

## Troubleshooting Performance Issues

### Diagnosis Commands

```bash
# Check collection performance
curl http://slurm-exporter:8080/metrics | grep collection_duration

# Monitor memory usage
kubectl top pod slurm-exporter

# Check for CPU throttling
kubectl exec slurm-exporter -- cat /sys/fs/cgroup/cpu/cpu.stat

# Analyze garbage collection
kubectl logs slurm-exporter | grep "gc\|GC"

# Profile CPU usage
kubectl exec slurm-exporter -- go tool pprof http://localhost:8080/debug/pprof/profile

# Profile memory usage  
kubectl exec slurm-exporter -- go tool pprof http://localhost:8080/debug/pprof/heap
```

### Performance Analysis

#### Collection Timing Analysis

```bash
# Analyze collection patterns
prometheus_query='
  histogram_quantile(0.95, 
    rate(slurm_exporter_collection_duration_seconds_bucket[5m])
  ) by (collector)
'

curl "http://prometheus:9090/api/v1/query?query=${prometheus_query}"
```

#### Memory Growth Analysis

```bash
# Track memory growth over time
prometheus_query='
  increase(slurm_exporter_memory_usage_bytes[1h]) > 0
'

curl "http://prometheus:9090/api/v1/query_range?query=${prometheus_query}&start=$(date -d '1 hour ago' +%s)&end=$(date +%s)&step=300"
```

### Common Performance Issues

#### Slow SLURM API Responses

**Problem:** High API response times
**Solution:**
```yaml
slurm:
  # Add circuit breaker
  circuit_breaker:
    enabled: true
    failure_threshold: 3
    timeout: 30s
    
  # Implement request timeout
  timeout: 30s
  
  # Add retries with backoff
  retry_attempts: 3
  retry_delay: 5s
```

#### Memory Leaks

**Problem:** Continuous memory growth
**Solution:**
```yaml
# Enable garbage collection tuning
runtime:
  gc_target_percentage: 50
  max_heap_size: "2GiB"
  
# Implement metric cardinality limits
metrics:
  max_series: 50000
  label_limits:
    user: 1000
    job_name: 2000
```

#### High CPU Usage

**Problem:** Excessive CPU consumption
**Solution:**
```yaml
# Reduce collection frequency
collection:
  intervals:
    cluster: 60s
    nodes: 60s
    jobs: 300s
    
# Limit concurrent workers
max_workers: 2

# Enable request batching
batch_requests: true
batch_size: 50
```

## Best Practices Summary

### Configuration
- Start with conservative collection intervals
- Implement selective collection for large clusters
- Use connection pooling and keep-alive
- Set appropriate timeouts and retries

### Scaling
- Use horizontal scaling for specialized collection
- Implement federation for multi-region setups
- Consider active-passive HA configuration
- Monitor resource usage continuously

### Optimization
- Limit metric cardinality proactively
- Use aggregation and sampling for high-volume metrics
- Implement efficient caching strategies
- Regular performance monitoring and tuning

### Monitoring
- Alert on collection performance degradation
- Monitor memory usage and growth patterns
- Track API response times and error rates
- Regular performance reviews and optimization