# SLURM Exporter Grafana Dashboards

This directory contains pre-built Grafana dashboards for monitoring SLURM clusters using the SLURM Prometheus Exporter.

## Available Dashboards

### 1. SLURM Cluster Overview (`slurm-overview.json`)

**Purpose:** High-level overview of SLURM cluster health and utilization

**Key Panels:**
- Cluster health status (exporter, controllers, database)
- Node states distribution (idle, allocated, down, drain)
- Queue sizes by partition
- CPU and memory utilization gauges
- Job states over time
- Job queue wait time histograms

**Use Cases:**
- Quick cluster health assessment
- Capacity planning
- Identifying resource bottlenecks
- Monitoring job throughput

### 2. SLURM Exporter Performance (`slurm-exporter-performance.json`)

**Purpose:** Monitor the performance and reliability of the SLURM exporter itself

**Key Panels:**
- Overall collector success rate
- Collection duration by collector (P95)
- SLA violations tracking
- Error type distribution
- Memory usage by collector
- Process CPU and memory usage
- Goroutines and file descriptors

**Use Cases:**
- Troubleshooting exporter issues
- Performance tuning
- Capacity planning for monitoring infrastructure
- SLA compliance monitoring

### 3. SLURM Job Analysis (`slurm-job-analysis.json`)

**Purpose:** Detailed analysis of job patterns and resource usage

**Key Panels:**
- Current job states pie chart
- Job failure rate trends
- Queue wait time statistics
- Queue depth by partition
- CPU/Memory allocation by partition
- Top users by job count and resource usage
- Top accounts by resource consumption

**Use Cases:**
- User behavior analysis
- Fair share policy effectiveness
- Queue optimization
- Resource allocation patterns
- Identifying heavy users

## Installation

### Prerequisites

- Grafana 8.0 or higher
- Prometheus data source configured
- SLURM exporter metrics being collected

### Import Process

1. **Via Grafana UI:**
   ```
   1. Navigate to Dashboards → Import
   2. Upload JSON file or paste contents
   3. Select your Prometheus data source
   4. Adjust dashboard UID if needed
   5. Click Import
   ```

2. **Via Grafana API:**
   ```bash
   curl -X POST -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
        -d @slurm-overview.json \
        http://localhost:3000/api/dashboards/db
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

## Configuration

### Dashboard Variables

All dashboards include templating variables for flexibility:

- **`datasource`**: Select Prometheus instance
- **`cluster`**: Filter by SLURM cluster name
- **`partition`**: Filter by partition(s) (multi-select)
- **`user`**: Filter by user(s) (multi-select)

### Customization

Common modifications:

1. **Adjust time ranges:**
   - Default: 6 hours
   - Modify in dashboard settings or use time picker

2. **Change thresholds:**
   - Edit panel → Field tab → Thresholds
   - Adjust colors and values for your environment

3. **Add custom labels:**
   - If using custom labels, add them to queries
   - Example: `{environment="$environment"}`

4. **Modify refresh interval:**
   - Default: 30 seconds
   - Change via dashboard settings

## Panel Descriptions

### Cluster Overview Dashboard

#### Health Status Row
- **Exporter Status**: Shows if the exporter is running
- **Active Controllers**: Number of SLURM controllers (should be ≥1)
- **Database Status**: SLURM database connectivity
- **CPU Utilization**: Overall cluster CPU usage percentage

#### Job Metrics Row
- **Job States Over Time**: Stacked time series of job states
- **Queue Wait Times**: Histogram showing P50 and P95 wait times

### Performance Dashboard

#### Collector Performance Row
- **Success Rate**: Overall percentage of successful collections
- **Collection Duration**: Average time to collect metrics
- **SLA Violations**: Count of SLA breaches in last hour

#### Resource Monitoring Row
- **Process CPU**: CPU usage of the exporter process
- **Process Memory**: RSS memory of the exporter
- **Goroutines & FDs**: Resource leak indicators

### Job Analysis Dashboard

#### Queue Analysis Row
- **Queue Depth**: Pending jobs per partition over time
- **Wait Time Distribution**: Percentile-based wait times

#### Top Users Row
- **By Job Count**: Users with most running jobs
- **By CPU Usage**: Users consuming most CPU cores
- **By Account**: Resource distribution by account

## Best Practices

### 1. Dashboard Organization

Create a dedicated folder for SLURM dashboards:
```
SLURM/
├── Overview
├── Performance
├── Job Analysis
├── Node Details (custom)
└── User Reports (custom)
```

### 2. Alerting Integration

Link dashboards to alerts:
- Add alert annotations to panels
- Use alert list panels for active issues
- Create drill-down links to detailed views

### 3. Performance Optimization

For large clusters:
- Increase Prometheus scrape intervals
- Use recording rules for complex queries
- Implement variable filters to limit data
- Consider dashboard caching

### 4. Access Control

Configure appropriate permissions:
- Read-only for general users
- Edit access for administrators
- Folder-level permissions for teams

## Troubleshooting

### Missing Data

1. **Check metric existence:**
   ```promql
   up{job="slurm-exporter"}
   ```

2. **Verify labels:**
   ```promql
   slurm_job_state{cluster_name="production"}
   ```

3. **Check time range:**
   - Ensure data exists for selected period

### Performance Issues

1. **Optimize queries:**
   - Add label selectors early
   - Use recording rules for heavy queries
   - Limit time ranges

2. **Reduce panel count:**
   - Split into multiple dashboards
   - Use tabs or rows to organize

3. **Enable caching:**
   - Dashboard settings → Time options → Cache timeout

### Variable Issues

1. **Empty dropdowns:**
   - Check variable query syntax
   - Verify label existence
   - Refresh variable values

2. **Slow loading:**
   - Use metric-specific queries
   - Add label filters to variable queries

## Contributing

To contribute new dashboards:

1. **Follow naming convention:**
   - `slurm-<purpose>.json`
   - Use lowercase with hyphens

2. **Include metadata:**
   - Description
   - Tags: `slurm`, specific tags
   - UID: `slurm-<purpose>`

3. **Test thoroughly:**
   - Multiple cluster sizes
   - Different time ranges
   - Various filter combinations

4. **Document panels:**
   - Add descriptions to complex panels
   - Include query comments

## Additional Resources

- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Guide](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [SLURM Exporter Metrics](../docs/metrics.md)
- [Example Alerts](../prometheus/alerts/)

## Support

For dashboard issues:
1. Check exporter logs for metric collection errors
2. Verify Prometheus is scraping successfully
3. Review Grafana server logs
4. Open an issue with dashboard export and error details