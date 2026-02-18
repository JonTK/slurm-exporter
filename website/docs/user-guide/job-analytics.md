# Job Analytics

SLURM Exporter includes job metrics that provide insights into job performance and resource utilization efficiency.

## Overview

The job metrics collected by the exporter enable analysis of:

- **Efficiency Calculations**: CPU and memory utilization metrics
- **Wait Time Analysis**: Queue wait time patterns
- **Resource Usage**: Tracking allocated vs. used resources

## Configuration

Enable job-related collectors in your configuration:

```yaml
collectors:
  jobs:
    enabled: true
    interval: 15s      # More frequent collection for job data
    timeout: 10s
    filters:
      job_states: []   # Collect all states, or filter: ["RUNNING", "PENDING"]
```

## Efficiency Metrics

### CPU Efficiency

CPU efficiency measures how effectively jobs utilize allocated CPU resources:

```promql
# CPU efficiency by user
avg(slurm_job_cpu_efficiency_ratio) by (user)

# Jobs with low CPU efficiency (< 50%)
slurm_job_cpu_efficiency_ratio < 0.5

# Average CPU efficiency by partition
avg(slurm_job_cpu_efficiency_ratio) by (partition)
```

**Interpretation:**
- **> 0.8**: Excellent CPU utilization
- **0.6 - 0.8**: Good utilization
- **0.4 - 0.6**: Moderate utilization (optimization opportunity)
- **< 0.4**: Poor utilization (investigation needed)

### Memory Efficiency

Memory efficiency tracks how well jobs use allocated memory:

```promql
# Memory efficiency distribution
histogram_quantile(0.95, slurm_job_memory_efficiency_ratio)

# Users with consistently low memory efficiency
avg_over_time(slurm_job_memory_efficiency_ratio[7d]) by (user) < 0.5
```

## Queue Wait Time Analysis

### Wait Time Queries

```promql
# Average queue wait time by partition
avg(slurm_job_wait_time_seconds) by (partition) / 60

# Queue wait time percentiles
histogram_quantile(0.95, slurm_job_wait_time_seconds)

# Wait time trend over time
avg_over_time(slurm_job_wait_time_seconds[1h]) by (partition)
```

## Resource Usage Analysis

### CPU Allocation Patterns

```promql
# CPUs allocated per job by user
avg(slurm_job_cpus_allocated) by (user)

# Job runtime distribution
histogram_quantile(0.95, slurm_job_runtime_seconds)

# Jobs exceeding time limits
slurm_job_runtime_seconds / slurm_job_time_limit_seconds > 0.9
```

### Partition Utilization

```promql
# CPU utilization by partition
(slurm_partition_cpus_allocated / slurm_partition_cpus_total) * 100

# Memory utilization by partition
(slurm_partition_memory_allocated_bytes / slurm_partition_memory_total_bytes) * 100

# Node availability by partition
slurm_partition_nodes_idle / slurm_partition_nodes_total
```

## Analytics Dashboard

### Key Performance Indicators

Create dashboards tracking these KPIs:

- Average CPU efficiency across cluster
- Memory utilization trends
- Job completion rates
- Queue wait times
- Throughput metrics (jobs per hour)

### Sample Dashboard Queries

**Efficiency Overview:**
```promql
# Overall cluster CPU utilization
sum(slurm_partition_cpus_allocated) / sum(slurm_partition_cpus_total)

# Efficiency by partition
avg(slurm_job_cpu_efficiency_ratio) by (partition)
```

**Performance Overview:**
```promql
# Average job runtime (in hours)
avg(slurm_job_runtime_seconds) / 3600

# Throughput metrics
rate(slurm_jobs_completed_total[1h])
```

## Alerting on Analytics

Set up alerts for analytics insights:

```yaml
# Efficiency alerts
- alert: LowClusterEfficiency
  expr: avg(slurm_job_cpu_efficiency_ratio) < 0.6
  for: 1h
  annotations:
    summary: "Cluster CPU efficiency below 60%"

- alert: LongQueueWaitTime
  expr: histogram_quantile(0.95, slurm_job_wait_time_seconds) > 3600
  for: 30m
  annotations:
    summary: "95th percentile queue wait time exceeds 1 hour"
```

## Best Practices

**For Users:**
1. **Profile jobs** before production runs
2. **Use efficiency metrics** to optimize resource requests
3. **Monitor wait times** and adjust submission patterns

**For Administrators:**
1. **Monitor cluster-wide efficiency trends** for capacity planning
2. **Set up alerts** for abnormal wait times and low efficiency
3. **Use partition-level metrics** to balance workload distribution

For more advanced monitoring setups, see the [Alerting Setup](alerting.md) guide.
