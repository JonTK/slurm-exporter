# Grafana Integration

This guide covers setting up Grafana dashboards and visualizations for SLURM cluster monitoring using metrics from SLURM Exporter.

## Overview

Grafana provides powerful visualization capabilities for SLURM metrics, enabling:

- **Real-time dashboards** for cluster monitoring
- **Historical analysis** of resource utilization
- **User-specific views** for different stakeholder needs
- **Alerting integration** with visual feedback
- **Custom panels** for specific use cases

## Quick Setup

### Datasource Configuration

Add Prometheus as a datasource in Grafana:

```json
{
  "name": "Prometheus-SLURM",
  "type": "prometheus",
  "url": "http://prometheus:9090",
  "access": "proxy",
  "basicAuth": false,
  "isDefault": true,
  "jsonData": {
    "httpMethod": "POST",
    "timeInterval": "30s",
    "queryTimeout": "60s"
  }
}
```

Via Grafana UI:
1. Go to **Configuration** -> **Data Sources**
2. Click **Add data source**
3. Select **Prometheus**
4. Configure URL and settings
5. Click **Save & Test**

### Dashboard Import

Import pre-built dashboards:

1. **Via Grafana UI:**
   - Go to **Dashboards** -> **Import**
   - Upload JSON file or paste JSON content
   - Select Prometheus datasource
   - Click **Import**

2. **Via API:**
   ```bash
   curl -X POST \
     -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
     -H "Content-Type: application/json" \
     -d @slurm-overview.json \
     http://grafana:3000/api/dashboards/db
   ```

3. **Via Provisioning:**
   ```yaml
   # /etc/grafana/provisioning/dashboards/slurm.yaml
   apiVersion: 1
   providers:
     - name: 'slurm'
       orgId: 1
       folder: 'SLURM'
       type: file
       disableDeletion: false
       updateIntervalSeconds: 10
       options:
         path: /var/lib/grafana/dashboards/slurm
   ```

## Available Dashboards

### 1. SLURM Cluster Overview

**Purpose:** High-level cluster monitoring and health status

**Key Panels:**
- Cluster health indicators
- Resource utilization gauges
- Job state distribution
- Queue metrics
- Node availability

**Import:** Use `dashboards/slurm-overview.json`

### 2. SLURM Job Analytics

**Purpose:** Detailed job performance and user activity analysis

**Key Panels:**
- Job completion rates
- Queue wait time analysis
- Resource efficiency metrics
- User activity patterns
- Account resource usage

**Import:** Use `dashboards/slurm-job-analysis.json`

### 3. SLURM Exporter Performance

**Purpose:** Monitor the exporter itself for reliability

**Key Panels:**
- Collection success rates
- API response times
- Memory and CPU usage
- Error tracking
- SLA compliance

**Import:** Use `dashboards/slurm-exporter-performance.json`

## Custom Dashboard Creation

### Basic Panel Setup

Create a new dashboard with essential SLURM metrics:

```json
{
  "dashboard": {
    "title": "SLURM Custom Dashboard",
    "tags": ["slurm", "hpc"],
    "panels": [
      {
        "title": "CPU Utilization",
        "type": "stat",
        "targets": [
          {
            "expr": "(sum(slurm_partition_cpus_allocated) / sum(slurm_partition_cpus_total)) * 100",
            "legendFormat": "CPU Utilization %"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "thresholds": {
              "steps": [
                {"color": "green", "value": 0},
                {"color": "yellow", "value": 70},
                {"color": "red", "value": 90}
              ]
            }
          }
        }
      }
    ]
  }
}
```

### Panel Types and Use Cases

#### Stat Panels
Best for current values and KPIs:

```promql
# Current CPU utilization
(sum(slurm_partition_cpus_allocated) / sum(slurm_partition_cpus_total)) * 100

# Active jobs count
sum(slurm_jobs_running)

# Nodes down
sum(slurm_nodes_total{state="down"})
```

#### Time Series Panels
Perfect for trends and historical data:

```promql
# Job completion rate over time
rate(slurm_jobs_completed_total[5m]) * 60

# CPU utilization trend
avg_over_time((sum(slurm_partition_cpus_allocated) / sum(slurm_partition_cpus_total))[1h:])

# Queue depth over time
sum(slurm_jobs_pending) by (partition)
```

#### Bar Gauge Panels
Great for comparing values across categories:

```promql
# CPU utilization by partition
(slurm_partition_cpus_allocated / slurm_partition_cpus_total) * 100

# Job count by user
sum(slurm_jobs_running) by (user)

# Memory usage by partition
(slurm_partition_memory_allocated_bytes / slurm_partition_memory_total_bytes) * 100
```

#### Pie Chart Panels
Ideal for showing proportions:

```promql
# Job states distribution
sum(slurm_jobs_total) by (state)

# Node states distribution
sum(slurm_nodes_total) by (state)
```

## Dashboard Variables

### Dynamic Filtering

Create template variables for flexible filtering:

```json
{
  "templating": {
    "list": [
      {
        "name": "datasource",
        "type": "datasource",
        "query": "prometheus"
      },
      {
        "name": "cluster",
        "type": "query",
        "datasource": "$datasource",
        "query": "label_values(slurm_nodes_total, cluster)",
        "refresh": "on_time_range_change",
        "includeAll": true,
        "multi": true
      },
      {
        "name": "partition",
        "type": "query",
        "datasource": "$datasource",
        "query": "label_values(slurm_partition_nodes_total{cluster=~\"$cluster\"}, partition)",
        "refresh": "on_time_range_change",
        "includeAll": true,
        "multi": true
      },
      {
        "name": "user",
        "type": "query",
        "datasource": "$datasource",
        "query": "label_values(slurm_jobs_total{cluster=~\"$cluster\"}, user)",
        "refresh": "on_time_range_change",
        "includeAll": true,
        "multi": true
      }
    ]
  }
}
```

## Performance Optimization

### Query Optimization

Optimize dashboard queries for better performance:

```promql
# Use recording rules for complex calculations
slurm:cluster_cpu_utilization_ratio

# Add specific label filters
slurm_jobs_total{cluster="production", partition=~"compute|gpu"}

# Limit time ranges appropriately
rate(slurm_jobs_completed_total[5m])

# Use appropriate aggregation intervals
avg_over_time(slurm_partition_cpus_allocated[1h])
```

## User Access and Permissions

### Role-Based Dashboards

Create different dashboards for different user roles:

**Administrators:**
- Full cluster overview
- Performance metrics
- Capacity planning

**Users:**
- Personal job status
- Queue information
- Resource availability

**Managers:**
- High-level summaries
- Utilization trends
- SLA compliance

## Troubleshooting

### Common Issues

**No Data in Panels:**
```bash
# Check Prometheus connectivity
curl -s http://prometheus:9090/api/v1/targets | jq '.data.activeTargets[] | select(.job=="slurm-exporter")'

# Verify metric existence
curl -s 'http://prometheus:9090/api/v1/query?query=up{job="slurm-exporter"}'
```

**Slow Dashboard Loading:**
```promql
# Optimize queries with specific label filters
slurm_jobs_total{cluster="production", partition="compute"}

# Use recording rules for complex calculations
slurm:cluster_cpu_utilization_ratio

# Limit time ranges
rate(slurm_jobs_completed_total[5m])
```

**Variable Issues:**
```bash
# Check variable query syntax in Grafana
label_values(slurm_nodes_total, cluster)

# Verify label existence in Prometheus
curl -s 'http://prometheus:9090/api/v1/label/cluster/values'
```

For more detailed troubleshooting information, see the [Troubleshooting Guide](../user-guide/troubleshooting.md).
