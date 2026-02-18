# Alerting Setup

This guide covers setting up comprehensive alerting for SLURM cluster monitoring using Prometheus AlertManager, covering everything from basic alerts to advanced notification workflows.

## Overview

SLURM Exporter alerting provides proactive monitoring for:

- **Cluster Health**: Node failures, controller issues, service outages
- **Resource Utilization**: High usage, bottlenecks, capacity planning
- **Job Performance**: Failed jobs, efficiency issues, queue problems
- **System Performance**: Exporter health, collection failures, performance degradation

## AlertManager Setup

### Installation

```bash
# Docker
docker run -d \
  --name alertmanager \
  -p 9093:9093 \
  -v ./alertmanager.yml:/etc/alertmanager/alertmanager.yml \
  prom/alertmanager:latest

# Kubernetes
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install alertmanager prometheus-community/alertmanager
```

### Basic Configuration

Create `alertmanager.yml`:

```yaml
global:
  smtp_smarthost: 'mail.example.com:587'
  smtp_from: 'slurm-alerts@example.com'
  smtp_auth_username: 'slurm-alerts@example.com'
  smtp_auth_password: 'your-password'

route:
  group_by: ['alertname', 'cluster', 'severity']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        severity: warning
      receiver: 'warning-alerts'

receivers:
  - name: 'default'
    email_configs:
      - to: 'admin@example.com'
        subject: 'SLURM Alert: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          Cluster: {{ .Labels.cluster }}
          Severity: {{ .Labels.severity }}
          {{ end }}

  - name: 'critical-alerts'
    email_configs:
      - to: 'oncall@example.com'
        subject: '[CRITICAL] SLURM: {{ .GroupLabels.alertname }}'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        channel: '#slurm-critical'
        title: 'Critical SLURM Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'

  - name: 'warning-alerts'
    email_configs:
      - to: 'team@example.com'
        subject: '[WARNING] SLURM: {{ .GroupLabels.alertname }}'
```

## Alert Rules

### Cluster Health Alerts

```yaml
# alerts/cluster-health.yml
groups:
  - name: slurm.cluster.health
    rules:
      - alert: SlurmControllerDown
        expr: slurm_controller_up == 0
        for: 1m
        labels:
          severity: critical
          component: controller
        annotations:
          summary: "SLURM controller is down"
          description: "SLURM controller {{ $labels.controller }} has been down for more than 1 minute"
          runbook_url: "https://docs.example.com/runbooks/slurm-controller-down"

      - alert: HighNodeFailureRate
        expr: increase(slurm_nodes_total{state="down"}[1h]) > 5
        for: 5m
        labels:
          severity: warning
          component: nodes
        annotations:
          summary: "High node failure rate detected"
          description: "More than 5 nodes have failed in the last hour"

      - alert: PartitionDown
        expr: slurm_partition_up == 0
        for: 5m
        labels:
          severity: critical
          component: partition
        annotations:
          summary: "Partition {{ $labels.partition }} is down"
          description: "Partition {{ $labels.partition }} has been unavailable for 5 minutes"
```

### Resource Utilization Alerts

```yaml
# alerts/resource-utilization.yml
groups:
  - name: slurm.resources
    rules:
      - alert: HighClusterCpuUtilization
        expr: (slurm_partition_cpus_allocated / slurm_partition_cpus_total) > 0.9
        for: 10m
        labels:
          severity: warning
          component: resources
        annotations:
          summary: "High CPU utilization in partition {{ $labels.partition }}"
          description: "CPU utilization is {{ $value | humanizePercentage }} in partition {{ $labels.partition }}"

      - alert: HighMemoryUtilization
        expr: (slurm_partition_memory_allocated_bytes / slurm_partition_memory_total_bytes) > 0.85
        for: 10m
        labels:
          severity: warning
          component: resources
        annotations:
          summary: "High memory utilization in partition {{ $labels.partition }}"
          description: "Memory utilization is {{ $value | humanizePercentage }} in partition {{ $labels.partition }}"

      - alert: NoIdleNodes
        expr: slurm_partition_nodes_idle == 0
        for: 15m
        labels:
          severity: warning
          component: capacity
        annotations:
          summary: "No idle nodes in partition {{ $labels.partition }}"
          description: "All nodes are allocated in partition {{ $labels.partition }}"
```

### Job Performance Alerts

```yaml
# alerts/job-performance.yml
groups:
  - name: slurm.jobs
    rules:
      - alert: HighJobFailureRate
        expr: rate(slurm_jobs_failed_total[1h]) > 0.1
        for: 10m
        labels:
          severity: warning
          component: jobs
        annotations:
          summary: "High job failure rate detected"
          description: "Job failure rate is {{ $value | humanize }} failures per second"

      - alert: LongQueueWaitTime
        expr: histogram_quantile(0.95, slurm_job_wait_time_seconds) > 3600
        for: 30m
        labels:
          severity: warning
          component: scheduling
        annotations:
          summary: "Long queue wait times detected"
          description: "95th percentile wait time is {{ $value | humanizeDuration }}"

      - alert: LowJobEfficiency
        expr: avg(slurm_job_cpu_efficiency_ratio) by (partition) < 0.5
        for: 1h
        labels:
          severity: warning
          component: efficiency
        annotations:
          summary: "Low job efficiency in partition {{ $labels.partition }}"
          description: "Average CPU efficiency is {{ $value | humanizePercentage }}"

      - alert: StuckJobs
        expr: increase(slurm_jobs_running[6h]) == 0 and slurm_jobs_running > 0
        for: 10m
        labels:
          severity: critical
          component: jobs
        annotations:
          summary: "Jobs appear to be stuck"
          description: "No change in running job count for 6 hours, possible stuck jobs"
```

### System Performance Alerts

```yaml
# alerts/system-performance.yml
groups:
  - name: slurm.exporter
    rules:
      - alert: SlurmExporterDown
        expr: up{job="slurm-exporter"} == 0
        for: 1m
        labels:
          severity: critical
          component: exporter
        annotations:
          summary: "SLURM Exporter is down"
          description: "SLURM Exporter instance {{ $labels.instance }} is down"

      - alert: HighCollectionDuration
        expr: slurm_exporter_collect_duration_seconds > 30
        for: 5m
        labels:
          severity: warning
          component: performance
        annotations:
          summary: "Slow metric collection"
          description: "Metric collection taking {{ $value }}s for collector {{ $labels.collector }}"

      - alert: CollectionErrors
        expr: increase(slurm_exporter_collect_errors_total[5m]) > 5
        for: 2m
        labels:
          severity: warning
          component: reliability
        annotations:
          summary: "Collection errors detected"
          description: "{{ $value }} collection errors in the last 5 minutes"

      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes{job="slurm-exporter"} > 1000000000
        for: 10m
        labels:
          severity: warning
          component: performance
        annotations:
          summary: "SLURM Exporter high memory usage"
          description: "Memory usage is {{ $value | humanizeBytes }}"

      - alert: StaleScrape
        expr: time() - slurm_exporter_last_collect_timestamp > 300
        for: 5m
        labels:
          severity: warning
          component: freshness
        annotations:
          summary: "Stale metrics detected"
          description: "Last successful collection was {{ $value }}s ago"
```

## Notification Channels

### Email Configuration

```yaml
email_configs:
  - to: 'team@example.com'
    from: 'slurm-alerts@example.com'
    subject: '[{{ .Status | toUpper }}] {{ .GroupLabels.alertname }}'
    html: |
      <h2>SLURM Alert: {{ .GroupLabels.alertname }}</h2>
      <table>
        <tr><th>Alert</th><th>Severity</th><th>Description</th></tr>
        {{ range .Alerts }}
        <tr>
          <td>{{ .Annotations.summary }}</td>
          <td>{{ .Labels.severity }}</td>
          <td>{{ .Annotations.description }}</td>
        </tr>
        {{ end }}
      </table>
```

### Slack Integration

```yaml
slack_configs:
  - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
    channel: '#slurm-alerts'
    title: '{{ .Status | title }} - {{ .GroupLabels.alertname }}'
    text: |
      {{ range .Alerts }}
      *Alert:* {{ .Annotations.summary }}
      *Severity:* {{ .Labels.severity }}
      *Cluster:* {{ .Labels.cluster }}
      *Description:* {{ .Annotations.description }}
      {{ if .Annotations.runbook_url }}*Runbook:* {{ .Annotations.runbook_url }}{{ end }}
      {{ end }}
    color: '{{ if eq .Status "firing" }}danger{{ else }}good{{ end }}'
```

### PagerDuty Integration

```yaml
pagerduty_configs:
  - service_key: 'your-pagerduty-integration-key'
    severity: '{{ .GroupLabels.severity }}'
    description: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
    details:
      cluster: '{{ .GroupLabels.cluster }}'
      component: '{{ .GroupLabels.component }}'
      runbook: '{{ range .Alerts }}{{ .Annotations.runbook_url }}{{ end }}'
```

## Alert Management

### Silencing Alerts

```bash
# Silence alerts via API
curl -X POST http://alertmanager:9093/api/v1/silences \
  -H "Content-Type: application/json" \
  -d '{
    "matchers": [
      {"name": "alertname", "value": "SlurmControllerDown"},
      {"name": "cluster", "value": "production"}
    ],
    "startsAt": "2024-01-15T10:00:00Z",
    "endsAt": "2024-01-15T12:00:00Z",
    "createdBy": "admin@example.com",
    "comment": "Planned maintenance window"
  }'
```

### Alert Inhibition

Prevent alert spam by suppressing related alerts:

```yaml
inhibit_rules:
  - source_match:
      alertname: 'SlurmControllerDown'
    target_match:
      alertname: 'HighJobFailureRate'
    equal: ['cluster']

  - source_match:
      alertname: 'PartitionDown'
    target_match_re:
      alertname: '(HighCpuUtilization|NoIdleNodes)'
    equal: ['cluster', 'partition']
```

## Testing and Validation

### Alert Testing

```bash
# Validate alert rules syntax
promtool check rules alerts/*.yml

# Test AlertManager configuration
amtool config check alertmanager.yml
```

### End-to-End Testing

```bash
# Send test notification
amtool alert add \
  --alertmanager.url=http://alertmanager:9093 \
  --annotation=summary="Test Alert" \
  --annotation=description="End-to-end test" \
  TestAlert severity=warning
```

## Best Practices

### Alert Design

1. **Clear, actionable alerts** - Every alert should have a clear action
2. **Context-rich descriptions** - Include enough detail for quick diagnosis
3. **Appropriate severity levels** - Reserve critical for service-affecting issues
4. **Runbook links** - Include links to resolution procedures
5. **Avoid alert fatigue** - Tune thresholds to minimize false positives

### Threshold Tuning

```yaml
# Progressive thresholds
- alert: ModerateResourceUsage
  expr: cpu_usage > 0.7
  labels:
    severity: info

- alert: HighResourceUsage
  expr: cpu_usage > 0.85
  labels:
    severity: warning

- alert: CriticalResourceUsage
  expr: cpu_usage > 0.95
  labels:
    severity: critical
```

For more advanced monitoring configurations, see the [Prometheus Integration](../integration/prometheus.md) and [Grafana Dashboards](../integration/grafana.md) guides.
