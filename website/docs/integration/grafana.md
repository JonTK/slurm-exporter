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
1. Go to **Configuration** → **Data Sources**
2. Click **Add data source**
3. Select **Prometheus**
4. Configure URL and settings
5. Click **Save & Test**

### Dashboard Import

Import pre-built dashboards:

1. **Via Grafana UI:**
   - Go to **Dashboards** → **Import**
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

### 4. SLURM Executive Overview

**Purpose:** High-level summary for management and planning

**Key Panels:**
- Cluster utilization trends
- Cost analysis
- Capacity planning indicators
- SLA compliance
- Growth projections

**Import:** Use `dashboards/slurm-executive-overview.json`

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

# Resource allocation by account
sum(slurm_account_cpu_hours_used_total) by (account)
```

#### Heatmap Panels
Excellent for pattern analysis:

```promql
# Queue wait times by hour of day
avg_over_time(slurm_job_wait_time_seconds[1h]) by (hour(timestamp()))

# Job submission patterns
rate(slurm_jobs_total[1h]) by (hour(timestamp()), day_of_week(timestamp()))
```

## Advanced Visualizations

### Multi-Cluster Dashboards

Template for monitoring multiple SLURM clusters:

```json
{
  "templating": {
    "list": [
      {
        "name": "cluster",
        "type": "query",
        "query": "label_values(slurm_nodes_total, cluster)",
        "refresh": "on_time_range_change",
        "includeAll": true,
        "multi": true
      }
    ]
  },
  "panels": [
    {
      "title": "CPU Utilization by Cluster",
      "targets": [
        {
          "expr": "(sum(slurm_partition_cpus_allocated{cluster=~\"$cluster\"}) by (cluster) / sum(slurm_partition_cpus_total{cluster=~\"$cluster\"}) by (cluster)) * 100"
        }
      ]
    }
  ]
}
```

### Drilldown Dashboards

Create linked dashboards for detailed analysis:

```json
{
  "panels": [
    {
      "title": "Partitions Overview",
      "targets": [
        {
          "expr": "sum(slurm_partition_nodes_total) by (partition)"
        }
      ],
      "options": {
        "dataLinks": [
          {
            "title": "Partition Details",
            "url": "/d/partition-details?var-partition=${__field.labels.partition}"
          }
        ]
      }
    }
  ]
}
```

### Custom Time Ranges

Define dashboard-specific time ranges:

```json
{
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": ["5s", "10s", "30s", "1m", "5m"],
    "time_options": ["5m", "15m", "1h", "6h", "12h", "24h", "2d", "7d"]
  }
}
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

### Calculated Variables

Create custom intervals and thresholds:

```json
{
  "templating": {
    "list": [
      {
        "name": "interval",
        "type": "interval",
        "query": "1m,5m,10m,30m,1h,6h,12h,1d,7d",
        "refresh": "on_time_range_change",
        "auto": true,
        "auto_count": 30,
        "auto_min": "1m"
      },
      {
        "name": "threshold_cpu",
        "type": "constant",
        "query": "80"
      }
    ]
  }
}
```

## Alert Integration

### Visual Alert Status

Display alert status in dashboards:

```json
{
  "panels": [
    {
      "title": "Active Alerts",
      "type": "alertlist",
      "options": {
        "showOptions": "current",
        "maxItems": 20,
        "sortOrder": "importance",
        "dashboardAlerts": false,
        "alertInstanceLabelFilter": "{job=\"slurm-exporter\"}"
      }
    }
  ]
}
```

### Panel Alerts

Create panel-based alerts:

```json
{
  "panels": [
    {
      "title": "CPU Utilization",
      "targets": [
        {
          "expr": "(sum(slurm_partition_cpus_allocated) / sum(slurm_partition_cpus_total)) * 100"
        }
      ],
      "alert": {
        "conditions": [
          {
            "evaluator": {
              "params": [85],
              "type": "gt"
            },
            "operator": {
              "type": "and"
            },
            "query": {
              "params": ["A", "5m", "now"]
            },
            "reducer": {
              "type": "last"
            },
            "type": "query"
          }
        ],
        "executionErrorState": "alerting",
        "for": "5m",
        "frequency": "10s",
        "handler": 1,
        "name": "High CPU Utilization",
        "noDataState": "no_data",
        "notifications": []
      }
    }
  ]
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

### Caching Configuration

Configure dashboard caching:

```json
{
  "refresh": "30s",
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": ["30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"]
  }
}
```

### Lazy Loading

Implement lazy loading for heavy panels:

```json
{
  "panels": [
    {
      "title": "Detailed Job Analysis",
      "datasource": "$datasource",
      "options": {
        "maxDataPoints": 1000
      },
      "transformations": [
        {
          "id": "limit",
          "options": {
            "limitField": 100
          }
        }
      ]
    }
  ]
}
```

## User Access and Permissions

### Role-Based Dashboards

Create different dashboards for different user roles:

**Administrators:**
- Full cluster overview
- Performance metrics
- Cost analysis
- Capacity planning

**Users:**
- Personal job status
- Queue information
- Resource availability
- Fair share status

**Managers:**
- High-level summaries
- Utilization trends
- SLA compliance
- Budget tracking

### Dashboard Permissions

Configure appropriate access controls:

```json
{
  "dashboard": {
    "title": "SLURM User Dashboard",
    "tags": ["slurm", "user"],
    "meta": {
      "canEdit": false,
      "canSave": false,
      "canStar": true
    }
  }
}
```

## Mobile and Responsive Design

### Mobile-Optimized Panels

Create mobile-friendly dashboards:

```json
{
  "panels": [
    {
      "title": "Key Metrics",
      "type": "stat",
      "gridPos": {
        "h": 4,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "options": {
        "orientation": "horizontal",
        "textMode": "auto"
      }
    }
  ]
}
```

### Responsive Layout

Use flexible grid layouts:

```json
{
  "panels": [
    {
      "gridPos": {
        "h": 8,
        "w": 24,  // Full width
        "x": 0,
        "y": 0
      }
    },
    {
      "gridPos": {
        "h": 8,
        "w": 12,  // Half width
        "x": 0,
        "y": 8
      }
    },
    {
      "gridPos": {
        "h": 8,
        "w": 12,  // Half width
        "x": 12,
        "y": 8
      }
    }
  ]
}
```

## Integration with External Systems

### Annotations

Add external events as annotations:

```json
{
  "annotations": {
    "list": [
      {
        "name": "Maintenance Windows",
        "datasource": "-- Grafana --",
        "type": "tags",
        "iconColor": "rgba(255, 96, 96, 1)",
        "tags": ["maintenance"]
      },
      {
        "name": "SLURM Alerts",
        "datasource": "Prometheus-SLURM",
        "expr": "ALERTS{job=\"slurm-exporter\",alertstate=\"firing\"}",
        "iconColor": "rgba(255, 0, 0, 1)",
        "titleFormat": "{{alertname}}",
        "textFormat": "{{description}}"
      }
    ]
  }
}
```

### Data Links

Create links to external systems:

```json
{
  "panels": [
    {
      "title": "Job Details",
      "options": {
        "dataLinks": [
          {
            "title": "View in SLURM",
            "url": "https://slurm-portal.example.com/job/${__field.labels.job_id}"
          },
          {
            "title": "User Details", 
            "url": "https://user-portal.example.com/user/${__field.labels.user}"
          }
        ]
      }
    }
  ]
}
```

## Backup and Version Control

### Dashboard Backup

Automated dashboard backup:

```bash
#!/bin/bash
# backup-grafana-dashboards.sh

GRAFANA_URL="http://localhost:3000"
API_KEY="your-api-key"
BACKUP_DIR="/backups/grafana/$(date +%Y%m%d)"

mkdir -p "$BACKUP_DIR"

# Get all dashboards
curl -H "Authorization: Bearer $API_KEY" \
     "$GRAFANA_URL/api/search?type=dash-db" | \
     jq -r '.[].uri' | \
     while read uri; do
       dashboard=$(basename "$uri")
       curl -H "Authorization: Bearer $API_KEY" \
            "$GRAFANA_URL/api/dashboards/$uri" | \
            jq '.dashboard' > "$BACKUP_DIR/${dashboard}.json"
     done
```

### Version Control Integration

Store dashboards in Git:

```yaml
# .github/workflows/dashboard-sync.yml
name: Sync Grafana Dashboards

on:
  push:
    paths:
      - 'dashboards/**'

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Update Grafana Dashboards
        run: |
          for file in dashboards/*.json; do
            curl -X POST \
              -H "Authorization: Bearer ${{ secrets.GRAFANA_API_KEY }}" \
              -H "Content-Type: application/json" \
              -d @"$file" \
              "${{ secrets.GRAFANA_URL }}/api/dashboards/db"
          done
```

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

### Debug Mode

Enable debug mode for troubleshooting:

```json
{
  "panels": [
    {
      "title": "Debug Panel",
      "targets": [
        {
          "expr": "slurm_jobs_total",
          "format": "table",
          "instant": true
        }
      ],
      "options": {
        "showHeader": true,
        "sortBy": [{"displayName": "Time", "desc": false}]
      }
    }
  ]
}
```

For more detailed troubleshooting information, see the [Troubleshooting Guide](../user-guide/troubleshooting.md).