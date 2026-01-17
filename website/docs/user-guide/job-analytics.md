# Job Analytics

SLURM Exporter includes a powerful job analytics engine that provides insights into job performance, resource utilization efficiency, and cluster optimization opportunities.

## Overview

The job analytics engine analyzes job execution patterns and resource usage to provide:

- **Efficiency Calculations**: CPU, memory, and GPU utilization metrics
- **Waste Detection**: Identification of over-provisioned resources
- **Bottleneck Analysis**: Performance constraints and optimization opportunities
- **Trend Analysis**: Historical patterns and forecasting
- **Anomaly Detection**: Unusual job behavior identification

## Configuration

Enable job analytics in your configuration:

```yaml
metrics:
  job_analytics:
    enabled: true
    efficiency_calculation: true
    waste_detection: true
    bottleneck_analysis: true
    trend_analysis: true
    anomaly_detection: true
    
    # Configuration options
    efficiency_threshold: 0.7    # 70% efficiency threshold
    waste_threshold: 0.3         # 30% waste threshold
    analysis_window: "7d"        # 7-day analysis window
    min_runtime: "5m"            # Minimum job runtime for analysis
    
    # Advanced settings
    statistical_window: 100      # Jobs to consider for statistics
    anomaly_sensitivity: 0.95    # Anomaly detection sensitivity
    trend_points: 30             # Data points for trend analysis
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

# Memory waste by partition
sum(slurm_job_waste_memory_gb_hours) by (partition)
```

### GPU Efficiency

For GPU-enabled jobs, track GPU utilization:

```promql
# GPU efficiency by job
slurm_job_gpu_utilization_ratio

# Average GPU efficiency by user
avg(slurm_job_gpu_utilization_ratio) by (user)

# Underutilized GPU jobs
slurm_job_gpu_utilization_ratio < 0.3
```

## Waste Detection

### Resource Waste Metrics

The analytics engine calculates wasted resources:

```yaml
# Example waste metrics
slurm_job_waste_cpu_hours: 245.7        # CPU hours wasted
slurm_job_waste_memory_gb_hours: 1250.3  # Memory GB-hours wasted
slurm_job_waste_gpu_hours: 12.5          # GPU hours wasted
slurm_job_waste_cost_estimate: 89.45     # Estimated cost impact
```

### Waste Analysis Queries

```promql
# Total waste by resource type
sum(rate(slurm_job_waste_cpu_hours[1h])) by (partition)
sum(rate(slurm_job_waste_memory_gb_hours[1h])) by (partition)

# Users generating the most waste
topk(10, sum(slurm_job_waste_cpu_hours) by (user))

# Waste trends over time
rate(slurm_job_waste_cpu_hours[1h])
```

### Waste Reduction Recommendations

The system generates automatic recommendations:

```json
{
  "job_id": "12345",
  "user": "alice",
  "recommendations": [
    {
      "type": "cpu_reduction",
      "current": "16 cores",
      "recommended": "8 cores",
      "potential_savings": "50% CPU waste reduction"
    },
    {
      "type": "memory_reduction", 
      "current": "64GB",
      "recommended": "32GB",
      "potential_savings": "50% memory waste reduction"
    }
  ]
}
```

## Bottleneck Analysis

### Performance Bottlenecks

Identify what's limiting job performance:

```yaml
# Bottleneck types detected
slurm_job_bottleneck_score{resource_type="cpu"}: 0.8
slurm_job_bottleneck_score{resource_type="memory"}: 0.3
slurm_job_bottleneck_score{resource_type="io"}: 0.9
slurm_job_bottleneck_score{resource_type="network"}: 0.1
```

**Bottleneck Score Interpretation:**
- **0.8-1.0**: Severe bottleneck (immediate attention needed)
- **0.6-0.8**: Moderate bottleneck (optimization opportunity)
- **0.3-0.6**: Minor bottleneck (monitor)
- **0.0-0.3**: No significant bottleneck

### Common Bottleneck Patterns

**I/O Bottlenecks:**
```promql
# Jobs with high I/O wait time
slurm_job_bottleneck_score{resource_type="io"} > 0.7

# I/O bottleneck by storage system
slurm_job_bottleneck_score{resource_type="io"} by (storage_path)
```

**Memory Bottlenecks:**
```promql
# Jobs hitting memory limits
slurm_job_bottleneck_score{resource_type="memory"} > 0.8

# Memory pressure by node
avg(slurm_node_memory_utilization_ratio) by (node) > 0.9
```

**Network Bottlenecks:**
```promql
# Network-bound jobs
slurm_job_bottleneck_score{resource_type="network"} > 0.6

# MPI jobs with communication overhead
slurm_job_bottleneck_score{resource_type="network",job_type="mpi"} > 0.5
```

## Trend Analysis

### Usage Trends

Track resource usage patterns over time:

```promql
# CPU usage trend (hourly)
avg_over_time(slurm_usage_trend_hourly{resource_type="cpu"}[24h])

# Memory usage trend (daily)
avg_over_time(slurm_usage_trend_daily{resource_type="memory"}[7d])

# Weekly utilization patterns
avg_over_time(slurm_usage_trend_weekly{resource_type="cpu"}[4w])
```

### Efficiency Trends

Monitor efficiency improvements over time:

```promql
# CPU efficiency trend by user
avg_over_time(slurm_job_cpu_efficiency_ratio[30d]) by (user)

# Partition efficiency improvement
increase(avg(slurm_job_cpu_efficiency_ratio) by (partition)[30d])

# Cluster-wide efficiency trend
avg_over_time(slurm_job_efficiency_score[7d])
```

### Predictive Analytics

The system provides usage forecasting:

```yaml
# Forecasted metrics
slurm_forecast_cpu_demand{horizon="1d"}: 850     # CPUs needed in 1 day
slurm_forecast_memory_demand{horizon="1w"}: 2.5  # TB memory needed in 1 week
slurm_forecast_queue_length{horizon="1h"}: 45    # Queue length in 1 hour
```

## Anomaly Detection

### Anomaly Metrics

Detect unusual patterns in job behavior:

```yaml
# Anomaly scores (0-1, higher = more anomalous)
slurm_anomaly_score{entity_type="job",entity_id="12345"}: 0.95
slurm_anomaly_score{entity_type="user",entity_id="alice"}: 0.72
slurm_anomaly_score{entity_type="partition",entity_id="gpu"}: 0.45
```

### Types of Anomalies Detected

**Resource Usage Anomalies:**
- Jobs using significantly more/less resources than typical
- Unusual resource request patterns
- Abnormal job duration vs. resource allocation

**Performance Anomalies:**
- Jobs running much slower/faster than expected
- Unusual efficiency patterns
- Abnormal failure rates

**Behavioral Anomalies:**
- Users submitting atypical job patterns
- Unusual submission times or frequencies
- Abnormal resource request distributions

### Anomaly Queries

```promql
# High-anomaly jobs (score > 0.8)
slurm_anomaly_score{entity_type="job"} > 0.8

# Users with anomalous behavior
slurm_anomaly_score{entity_type="user"} > 0.7

# Anomaly trends over time
rate(slurm_anomaly_score[1h])
```

## Analytics Dashboard

### Key Performance Indicators

Create dashboards tracking these KPIs:

```yaml
Efficiency KPIs:
  - Average CPU efficiency across cluster
  - Memory utilization trends
  - GPU efficiency by partition
  - Resource waste metrics

Performance KPIs:
  - Job completion rates
  - Queue wait times
  - Throughput metrics
  - Bottleneck frequencies

Cost KPIs:
  - Estimated waste costs
  - Cost per CPU-hour
  - ROI on resource allocations
  - Budget utilization tracking
```

### Sample Dashboard Queries

**Efficiency Overview:**
```promql
# Overall cluster efficiency
avg(slurm_job_efficiency_score)

# Efficiency by partition
avg(slurm_job_cpu_efficiency_ratio) by (partition)

# Top wasteful users
topk(10, sum(slurm_job_waste_cpu_hours) by (user))
```

**Performance Overview:**
```promql
# Average job runtime
avg(slurm_job_runtime_seconds) / 3600

# Queue wait time percentiles
histogram_quantile(0.95, slurm_job_wait_time_seconds)

# Throughput metrics
rate(slurm_jobs_completed_total[1h])
```

## Optimization Recommendations

### Automated Recommendations

The analytics engine provides actionable insights:

```json
{
  "cluster_recommendations": [
    {
      "type": "partition_rebalancing",
      "description": "Move 20% of CPU nodes from 'debug' to 'compute' partition",
      "impact": "Reduce average queue wait time by 15%"
    },
    {
      "type": "memory_optimization",
      "description": "Reduce default memory allocation in 'gpu' partition",
      "impact": "Increase memory efficiency by 25%"
    }
  ],
  "user_recommendations": [
    {
      "user": "alice",
      "type": "resource_rightsizing",
      "jobs_affected": 145,
      "potential_savings": "30% CPU waste reduction"
    }
  ]
}
```

### Best Practices

**For Users:**
1. **Profile jobs** before production runs
2. **Use efficiency metrics** to optimize resource requests
3. **Monitor waste alerts** and adjust accordingly
4. **Follow bottleneck recommendations** for performance tuning

**For Administrators:**
1. **Set efficiency thresholds** and enforce policies
2. **Monitor cluster-wide trends** for capacity planning
3. **Use anomaly detection** for security and resource abuse
4. **Implement automated recommendations** where appropriate

## Integration with External Tools

### Slurm Integration

Export analytics back to Slurm:

```bash
# Generate efficiency reports
slurm-exporter --generate-efficiency-report --user=alice --period=30d

# Export recommendations to Slurm accounting
slurm-exporter --export-recommendations --format=sacct
```

### ML/AI Integration

Feed analytics data to machine learning systems:

```yaml
# Export configuration
analytics:
  ml_export:
    enabled: true
    endpoint: "http://ml-service:8080/api/v1/training-data"
    features: ["efficiency", "bottlenecks", "trends"]
    export_interval: "1h"
```

### Custom Analytics Plugins

Extend analytics with custom plugins:

```go
// Example custom analytics plugin
type CustomAnalyzer struct {
    config AnalyticsConfig
}

func (a *CustomAnalyzer) AnalyzeJob(job *JobMetrics) (*AnalysisResult, error) {
    // Custom analysis logic
    return &AnalysisResult{
        EfficiencyScore: calculateCustomEfficiency(job),
        Recommendations: generateRecommendations(job),
        Anomalies: detectAnomalies(job),
    }, nil
}
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

- alert: HighResourceWaste
  expr: sum(rate(slurm_job_waste_cpu_hours[1h])) > 100
  for: 30m
  annotations:
    summary: "High CPU waste detected (>100 CPU-hours/hour)"

# Anomaly alerts
- alert: JobAnomalyDetected
  expr: slurm_anomaly_score{entity_type="job"} > 0.9
  for: 5m
  annotations:
    summary: "Highly anomalous job detected"
```

For more advanced monitoring setups, see the [Alerting Setup](alerting.md) guide.