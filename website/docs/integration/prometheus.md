# Prometheus Integration

This guide covers integrating SLURM Exporter with Prometheus for comprehensive SLURM cluster monitoring.

## Overview

SLURM Exporter exposes metrics in Prometheus format, enabling:

- **Time-series storage** of SLURM metrics
- **Flexible querying** with PromQL
- **Alerting** based on metric thresholds
- **Integration** with visualization tools like Grafana

## Basic Setup

### Prometheus Configuration

Add SLURM Exporter as a scrape target in `prometheus.yml`:

```yaml
global:
  scrape_interval: 30s
  evaluation_interval: 30s

scrape_configs:
  - job_name: 'slurm-exporter'
    static_configs:
      - targets: ['localhost:9341']
    scrape_interval: 30s
    scrape_timeout: 25s
    metrics_path: /metrics
    honor_labels: true
    
    # Optional: Add external labels
    external_labels:
      cluster: 'production'
      datacenter: 'dc1'
```

### Verification

```bash
# Check if Prometheus is scraping SLURM Exporter
curl -s 'http://prometheus:9090/api/v1/targets' | \
  jq '.data.activeTargets[] | select(.job=="slurm-exporter") | {health: .health, lastScrape: .lastScrape}'

# Verify metrics are available
curl -s 'http://prometheus:9090/api/v1/query?query=up{job="slurm-exporter"}' | \
  jq '.data.result[0].value[1]'
```

## Service Discovery

### Kubernetes Service Discovery

For Kubernetes deployments, use automatic service discovery:

```yaml
scrape_configs:
  - job_name: 'kubernetes-services'
    kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names: [monitoring]
    
    relabel_configs:
      # Only scrape services with annotation
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      
      # Use custom path if specified
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      
      # Use custom port if specified
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      
      # Set job name from service name
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: job
```

Annotate your SLURM Exporter service:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: slurm-exporter
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9341"
    prometheus.io/path: "/metrics"
spec:
  ports:
    - port: 9341
      name: metrics
  selector:
    app: slurm-exporter
```

### Docker Swarm Service Discovery

```yaml
scrape_configs:
  - job_name: 'dockerswarm'
    dockerswarm_sd_configs:
      - host: unix:///var/run/docker.sock
        role: tasks
    
    relabel_configs:
      - source_labels: [__meta_dockerswarm_service_label_prometheus_job]
        target_label: job
      - source_labels: [__meta_dockerswarm_service_label_prometheus_port]
        target_label: __address__
        regex: (.*)
        replacement: ${1}:9341
```

## Recording Rules

Create recording rules for frequently used queries to improve performance:

```yaml
# /etc/prometheus/rules/slurm-recording-rules.yml
groups:
  - name: slurm.rules
    interval: 30s
    rules:
      # Cluster utilization rules
      - record: slurm:cluster_cpu_utilization_ratio
        expr: |
          sum(slurm_partition_cpus_allocated) by (cluster) / 
          sum(slurm_partition_cpus_total) by (cluster)
      
      - record: slurm:cluster_memory_utilization_ratio
        expr: |
          sum(slurm_partition_memory_allocated_bytes) by (cluster) / 
          sum(slurm_partition_memory_total_bytes) by (cluster)
      
      - record: slurm:cluster_node_availability_ratio
        expr: |
          sum(slurm_partition_nodes_idle) by (cluster) / 
          sum(slurm_partition_nodes_total) by (cluster)

      # Partition-level aggregations
      - record: slurm:partition_queue_depth
        expr: |
          sum(slurm_jobs_pending) by (cluster, partition)
      
      - record: slurm:partition_efficiency_ratio
        expr: |
          avg(slurm_job_cpu_efficiency_ratio) by (cluster, partition)

      # User and account summaries
      - record: slurm:user_cpu_hours_24h
        expr: |
          increase(slurm_account_cpu_hours_used_total[24h]) by (cluster, user, account)
      
      - record: slurm:account_job_completion_rate_1h
        expr: |
          rate(slurm_jobs_completed_total[1h]) by (cluster, account)

      # Performance indicators
      - record: slurm:job_wait_time_p95_5m
        expr: |
          histogram_quantile(0.95, 
            rate(slurm_job_wait_time_seconds_bucket[5m])
          ) by (cluster, partition)
      
      - record: slurm:job_failure_rate_1h
        expr: |
          rate(slurm_jobs_failed_total[1h]) by (cluster, partition, user)

      # Resource waste calculations
      - record: slurm:cluster_cpu_waste_rate_1h
        expr: |
          sum(rate(slurm_job_waste_cpu_hours[1h])) by (cluster)
      
      - record: slurm:partition_memory_waste_rate_1h
        expr: |
          sum(rate(slurm_job_waste_memory_gb_hours[1h])) by (cluster, partition)
```

Load rules in `prometheus.yml`:

```yaml
rule_files:
  - "/etc/prometheus/rules/slurm-recording-rules.yml"
  - "/etc/prometheus/rules/slurm-alerting-rules.yml"
```

## Query Examples

### Resource Utilization Queries

```promql
# Current CPU utilization by partition
(slurm_partition_cpus_allocated / slurm_partition_cpus_total) * 100

# Memory utilization trend over 24 hours
avg_over_time(
  (slurm_partition_memory_allocated_bytes / slurm_partition_memory_total_bytes)[24h:1h]
) * 100

# Node availability by state
sum(slurm_nodes_total) by (cluster, state)

# GPU utilization across all nodes
sum(slurm_node_gpus_allocated) / sum(slurm_node_gpus_total) * 100
```

### Job Performance Queries

```promql
# Job completion rate per hour
rate(slurm_jobs_completed_total[1h]) * 3600

# Average queue wait time by partition
avg(slurm_job_wait_time_seconds) by (partition) / 60

# Job efficiency distribution
histogram_quantile(0.95, slurm_job_cpu_efficiency_ratio)

# Top users by resource consumption
topk(10, sum(slurm_job_cpu_hours_total) by (user))
```

### Cluster Health Queries

```promql
# Exporter availability
up{job="slurm-exporter"}

# Controller status
slurm_controller_up

# Database connectivity
slurm_database_up

# Collection success rate
rate(slurm_exporter_collect_duration_seconds_count[5m]) - 
rate(slurm_exporter_collect_errors_total[5m])
```

### Capacity Planning Queries

```promql
# Predict CPU demand growth
predict_linear(
  slurm:cluster_cpu_utilization_ratio[7d], 
  86400 * 30  # 30 days
)

# Resource reservation conflicts
(slurm_reservation_cpus / slurm_cluster_cpus_total) * 100

# Queue depth trends
deriv(slurm:partition_queue_depth[1h])

# Peak usage patterns by hour of day
avg_over_time(
  (slurm_partition_cpus_allocated / slurm_partition_cpus_total)[7d:1h]
) by (hour(timestamp()))
```

## Advanced Configuration

### Multi-Cluster Monitoring

For monitoring multiple SLURM clusters:

```yaml
scrape_configs:
  - job_name: 'slurm-cluster-1'
    static_configs:
      - targets: ['cluster1-exporter:9341']
    params:
      cluster: ['cluster1']
    external_labels:
      cluster: 'cluster1'
      datacenter: 'dc1'
  
  - job_name: 'slurm-cluster-2'
    static_configs:
      - targets: ['cluster2-exporter:9341']
    params:
      cluster: ['cluster2']
    external_labels:
      cluster: 'cluster2'
      datacenter: 'dc2'
```

### Metric Relabeling

Optimize metric collection with relabeling:

```yaml
scrape_configs:
  - job_name: 'slurm-exporter'
    static_configs:
      - targets: ['localhost:9341']
    
    metric_relabel_configs:
      # Drop high-cardinality metrics in test environments
      - source_labels: [__name__]
        regex: 'slurm_job_step_.*'
        action: drop
        target_label: '__tmp_drop'
      
      # Rename metrics for consistency
      - source_labels: [__name__]
        regex: 'slurm_node_cpus_total'
        replacement: 'slurm_node_cpu_cores_total'
        target_label: __name__
      
      # Add environment label
      - target_label: environment
        replacement: 'production'
```

### Federation

Set up Prometheus federation for hierarchical monitoring:

```yaml
# Higher-level Prometheus configuration
scrape_configs:
  - job_name: 'federate'
    scrape_interval: 60s
    honor_labels: true
    metrics_path: '/federate'
    params:
      'match[]':
        - '{job=~"slurm-.*"}'
        - 'slurm:cluster_cpu_utilization_ratio'
        - 'slurm:partition_queue_depth'
    static_configs:
      - targets:
        - 'prometheus-cluster1:9090'
        - 'prometheus-cluster2:9090'
```

## Storage and Retention

### Storage Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 30s
  
# Command line flags for prometheus
# --storage.tsdb.path=/prometheus/data
# --storage.tsdb.retention.time=90d
# --storage.tsdb.retention.size=50GB
# --storage.tsdb.wal-compression
```

### Retention Policies

```bash
# Configure retention based on metric importance
prometheus \
  --storage.tsdb.retention.time=30d \
  --storage.tsdb.retention.size=10GB \
  --web.enable-admin-api

# Create different retention for recording rules
# Short-term: Raw metrics (7 days)
# Medium-term: 5-minute aggregates (30 days)  
# Long-term: 1-hour aggregates (1 year)
```

## Performance Optimization

### Scrape Configuration Tuning

```yaml
scrape_configs:
  - job_name: 'slurm-exporter'
    static_configs:
      - targets: ['localhost:9341']
    
    # Optimize for large clusters
    scrape_interval: 60s      # Reduce frequency for large datasets
    scrape_timeout: 45s       # Allow more time for collection
    sample_limit: 50000       # Prevent memory issues
    metric_relabel_configs:
      # Drop unused metrics to reduce cardinality
      - source_labels: [__name__]
        regex: 'slurm_exporter_.*'
        action: drop
```

### Memory Optimization

```yaml
# Recording rules to pre-aggregate data
groups:
  - name: slurm.aggregates
    interval: 60s
    rules:
      # Pre-calculate expensive aggregations
      - record: slurm:node_utilization_5m
        expr: |
          avg_over_time(
            (slurm_node_cpus_allocated / slurm_node_cpus_total)[5m:]
          ) by (cluster, node)
      
      # Reduce cardinality for job metrics
      - record: slurm:job_states_by_partition
        expr: |
          sum(slurm_jobs_total) by (cluster, partition, state)
```

### Query Performance

```promql
# Use recording rules for complex queries
slurm:cluster_cpu_utilization_ratio

# Optimize with label selection
slurm_jobs_total{cluster="production", partition=~"compute|gpu"}

# Use rate() for counters
rate(slurm_jobs_completed_total[5m])

# Limit time ranges for expensive queries
slurm_job_wait_time_seconds[1h]
```

## Integration with External Systems

### Prometheus Operator

For Kubernetes with Prometheus Operator:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: slurm-exporter
  labels:
    app: slurm-exporter
spec:
  selector:
    matchLabels:
      app: slurm-exporter
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
    honorLabels: true
    scrapeTimeout: 25s
```

### Remote Write

Send metrics to remote storage:

```yaml
remote_write:
  - url: "https://prometheus-remote-write.example.com/api/v1/write"
    write_relabel_configs:
      # Only send SLURM metrics
      - source_labels: [__name__]
        regex: 'slurm_.*'
        action: keep
    
    queue_config:
      batch_send_deadline: 5s
      max_samples_per_send: 1000
      capacity: 10000
```

### Thanos Integration

For long-term storage with Thanos:

```yaml
# prometheus.yml with Thanos sidecar
external_labels:
  cluster: 'slurm-production'
  replica: '0'

# Thanos sidecar configuration
--tsdb.path=/prometheus/data
--prometheus.url=http://localhost:9090
--objstore.config-file=/etc/thanos/bucket.yml
--reloader.config-file=/etc/prometheus/prometheus.yml
```

## Monitoring Prometheus Itself

### Prometheus Self-Monitoring

```yaml
scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: /metrics
    scrape_interval: 15s
```

### Key Metrics to Monitor

```promql
# Prometheus ingestion rate
rate(prometheus_tsdb_samples_appended_total[5m])

# Query performance
histogram_quantile(0.95, rate(prometheus_http_request_duration_seconds_bucket[5m]))

# Storage usage
prometheus_tsdb_size_bytes / 1024 / 1024 / 1024  # GB

# WAL corruption or issues
increase(prometheus_tsdb_wal_corruptions_total[1h])
```

## Backup and Recovery

### Configuration Backup

```bash
#!/bin/bash
# backup-prometheus-config.sh

BACKUP_DIR="/backups/prometheus/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Backup configuration
cp /etc/prometheus/prometheus.yml "$BACKUP_DIR/"
cp -r /etc/prometheus/rules/ "$BACKUP_DIR/"

# Backup runtime configuration
curl -s http://localhost:9090/api/v1/status/config > "$BACKUP_DIR/runtime-config.json"

# Create tarball
tar -czf "$BACKUP_DIR.tar.gz" -C /backups/prometheus "$(basename $BACKUP_DIR)"
```

### Data Backup

```bash
#!/bin/bash
# backup-prometheus-data.sh

# Create snapshot
curl -XPOST http://localhost:9090/api/v1/admin/tsdb/snapshot

# Get snapshot name
SNAPSHOT=$(curl -s http://localhost:9090/api/v1/admin/tsdb/snapshot | jq -r '.data.name')

# Copy snapshot to backup location
cp -r "/prometheus/data/snapshots/$SNAPSHOT" "/backups/prometheus-data/$(date +%Y%m%d)"
```

## Troubleshooting

### Common Issues

**High Cardinality:**
```bash
# Check series count
curl -s http://localhost:9090/api/v1/label/__name__/values | jq '.data | length'

# Identify high-cardinality metrics
curl -s 'http://localhost:9090/api/v1/query?query=topk(10,count by (__name__)({__name__=~".%2B"}))'
```

**Slow Queries:**
```bash
# Enable query logging
--web.enable-admin-api
--log.level=debug

# Check slow queries
curl -s http://localhost:9090/api/v1/status/tsdb | jq '.data.seriesCountByMetricName'
```

**Storage Issues:**
```bash
# Check disk usage
du -sh /prometheus/data

# Monitor WAL size
ls -lah /prometheus/data/wal/
```

For more troubleshooting information, see the [Troubleshooting Guide](../user-guide/troubleshooting.md).