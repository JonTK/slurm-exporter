# SLURM Exporter Grafana Dashboards

Pre-built Grafana dashboards for monitoring SLURM clusters with the SLURM Prometheus Exporter.

## Available Dashboards

This directory contains 12 dashboard JSON files:

| File | Description |
|------|-------------|
| `slurm-overview.json` | High-level cluster health and utilization |
| `slurm-exporter-performance.json` | Exporter self-monitoring (collection times, errors, SLA) |
| `slurm-job-analysis.json` | Job patterns, queue depths, wait times, top users |
| `slurm-user-view.json` | Per-user resource usage and job counts |
| `slurm-capacity-planning.json` | Capacity trends and forecasting |
| `slurm-operations.json` | Operational metrics for administrators |
| `slurm-executive-overview.json` | Management-level summary |
| `cluster-overview.json` | Cluster-wide resource overview |
| `job-analytics.json` | Detailed job analytics |
| `resource-utilization.json` | CPU, memory, and GPU utilization |
| `node-health.json` | Node state and hardware health |
| `operational-dashboard.json` | Day-to-day operations view |

## Installation

### Prerequisites

- Grafana 8.0 or higher
- Prometheus data source configured and scraping the SLURM exporter

### Import via Grafana UI

1. Navigate to **Dashboards > Import**
2. Click **Upload JSON file**
3. Select a dashboard JSON file from this directory
4. Choose your Prometheus data source
5. Click **Import**

### Import via Grafana API

```bash
curl -X POST -H "Content-Type: application/json" \
     -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
     -d @slurm-overview.json \
     http://localhost:3000/api/dashboards/db
```

### Import via Provisioning

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

Copy the JSON files to the provisioning path and restart Grafana.

## Dashboard Variables

All dashboards include templating variables:

- **`datasource`**: Select Prometheus instance
- **`cluster`**: Filter by SLURM cluster name
- **`partition`**: Filter by partition (multi-select)
- **`user`**: Filter by user (multi-select)

## See Also

- [Metrics Guide](../website/docs/user-guide/metrics.md) — Full list of available metrics
- [Prometheus Alert Rules](../prometheus/alerts/) — Alerting rules for SLURM metrics
