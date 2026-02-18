# SLURM Exporter Alerting Rules

This directory contains comprehensive Prometheus alerting rules for monitoring SLURM clusters using the SLURM exporter.

## 📋 Overview

The alerting rules are organized into six files:

1. **Health Monitoring** (`alerts/slurm_health.yaml`)
   - Exporter availability and health
   - SLURM API connectivity
   - Controller and database health
   - Configuration changes

2. **Job Monitoring** (`alerts/slurm_jobs.yaml`)
   - Job failure rates
   - Queue backlogs and wait times
   - Resource usage by jobs
   - Inefficient job detection

3. **Resource Monitoring** (`alerts/slurm_resources.yaml`)
   - CPU and memory utilization
   - Node availability
   - Resource fragmentation
   - Capacity planning

4. **SLA Monitoring** (`alerts/slurm_sla.yaml`)
   - Job start time SLAs
   - Availability SLAs
   - Performance metrics
   - Data quality

5. **Exporter Performance** (`alerts/slurm_exporter_performance.yaml`)
   - Collection duration and errors
   - Exporter resource usage

6. **Cluster Health** (`alerts/slurm_cluster_health.yaml`)
   - Cluster-wide health indicators

## 🚀 Quick Start

### 1. Deploy to Prometheus

Copy the alert files to your Prometheus configuration directory:

```bash
# Copy all alert files
cp -r prometheus/alerts /etc/prometheus/

# Add to prometheus.yml
cat >> /etc/prometheus/prometheus.yml <<EOF
rule_files:
  - "alerts/slurm_health.yaml"
  - "alerts/slurm_jobs.yaml"
  - "alerts/slurm_resources.yaml"
  - "alerts/slurm_sla.yaml"
  - "alerts/slurm_exporter_performance.yaml"
  - "alerts/slurm_cluster_health.yaml"
EOF

# Reload Prometheus
curl -X POST http://localhost:9090/-/reload
```

### 2. Configure Alertmanager

Use the example configuration as a starting point:

```bash
cp prometheus/alertmanager-example.yaml /etc/alertmanager/alertmanager.yml
# Edit to add your notification endpoints
vim /etc/alertmanager/alertmanager.yml
```

### 3. Verify Alerts

Check that alerts are loaded correctly:

```bash
# View all alerts
curl http://localhost:9090/api/v1/rules | jq '.data.groups[].rules[].name'

# Check alert state
curl http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | {name: .labels.alertname, state: .state}'
```

## 🎯 Alert Severity Levels

- **🔴 Critical**: Immediate action required
  - Service outages
  - Data loss risk
  - SLA breaches
  
- **🟡 Warning**: Investigation needed
  - Performance degradation
  - Approaching limits
  - Failed jobs
  
- **🔵 Info**: Informational only
  - Capacity planning
  - Usage trends
  - Configuration changes

## 🏷️ Alert Labels

All alerts include standard labels for routing:

- `severity`: critical, warning, info
- `component`: Part of system affected
- `team`: Responsible team
- `cluster_name`: Affected SLURM cluster
- `environment`: prod, staging, dev

## 📊 Key Alerts

### Critical Alerts

| Alert | Description | Action Required |
|-------|-------------|-----------------|
| `SLURMExporterDown` | Exporter not responding | Check exporter service |
| `SLURMAPIUnreachable` | Cannot connect to SLURM | Check slurmctld service |
| `SLURMControllerDown` | No active controllers | Restart SLURM controller |
| `SLURMQueueStalled` | Jobs not starting | Check scheduler and resources |

### Warning Alerts

| Alert | Description | Investigation Needed |
|-------|-------------|---------------------|
| `SLURMHighJobFailureRate` | >25% jobs failing | Check job logs and system health |
| `SLURMCPUExhaustion` | >95% CPU used | Consider adding nodes |
| `SLURMExcessiveQueueWaitTime` | Long queue times | Review priorities and resources |
| `SLURMNodeOffline` | Nodes unavailable | Check node health |

## 🔧 Customization

### Adjusting Thresholds

Edit the alert expressions to match your environment:

```yaml
# Example: Change CPU exhaustion threshold from 95% to 90%
- alert: SLURMCPUExhaustion
  expr: |
    (
      sum by (cluster_name, partition) (slurm_node_cpus_allocated)
      /
      sum by (cluster_name, partition) (slurm_node_cpus_total)
    ) > 0.90  # Changed from 0.95
```

### Adding Custom Alerts

Create new alerts following the pattern:

```yaml
- alert: YourCustomAlert
  expr: your_metric > threshold
  for: 5m
  labels:
    severity: warning
    component: your_component
    team: your_team
  annotations:
    summary: "Brief description"
    description: |
      Detailed description with context.
      Value: {{ $value }}
      Labels: {{ $labels }}
    runbook_url: https://wiki.example.com/runbooks/your-alert
```

### Disabling Alerts

Comment out or remove alerts that don't apply:

```yaml
# Disabled - not using reservations
# - alert: SLURMReservationConflicts
#   expr: ...
```

## 📚 Runbooks

Each alert includes a `runbook_url` annotation pointing to documentation for resolving the issue. Create runbooks that include:

1. **Alert meaning**: What the alert indicates
2. **Impact**: Business/user impact
3. **Diagnosis**: How to investigate
4. **Resolution**: Steps to fix
5. **Prevention**: How to avoid recurrence

## 🔗 Integration

### Grafana

Import dashboard annotations:

```json
{
  "datasource": "Prometheus",
  "enable": true,
  "expr": "ALERTS{alertname=~\"SLURM.*\",alertstate=\"firing\"}"
}
```

### PagerDuty

Alerts with `pagerduty_key` label will create incidents:

```yaml
labels:
  pagerduty_key: slurm-critical-issue
```

### Slack

Configure channels in Alertmanager:

```yaml
slack_configs:
  - channel: '#slurm-alerts'
    title: '{{ .GroupLabels.alertname }}'
```

## 📈 Metrics Used

Key metrics referenced in alerts:

- `slurm_job_state`: Job states and counts
- `slurm_node_cpus_*`: CPU allocation
- `slurm_job_queue_time_seconds`: Queue wait times
- `slurm_performance_*`: Efficiency metrics
- `slurm_system_*`: System health metrics

## 🧪 Testing Alerts

Test alerts without triggering notifications:

```bash
# Use promtool to check syntax
promtool check rules alerts/*.yaml

# Test specific alert
promtool test rules test_alerts.yaml
```

## 📝 Best Practices

1. **Start Conservative**: Begin with higher thresholds and tune down
2. **Avoid Alert Fatigue**: Only alert on actionable issues
3. **Include Context**: Use labels and annotations effectively
4. **Test Changes**: Validate alerts in staging first
5. **Document Everything**: Maintain runbooks for all alerts

## 🤝 Contributing

To add new alerts:

1. Identify the monitoring gap
2. Write the alert rule
3. Test in development
4. Document in runbook
5. Submit PR with rationale

## 📞 Support

For questions or issues:
- Check existing runbooks
- Review Prometheus logs
- Contact the monitoring team
- File an issue in the repository