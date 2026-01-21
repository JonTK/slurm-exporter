# Alerting Setup

This guide covers setting up comprehensive alerting for SLURM cluster monitoring using Prometheus AlertManager, covering everything from basic alerts to advanced notification workflows.

## Overview

SLURM Exporter alerting provides proactive monitoring for:

- **Cluster Health**: Node failures, controller issues, service outages
- **Resource Utilization**: High usage, bottlenecks, capacity planning
- **Job Performance**: Failed jobs, efficiency issues, queue problems
- **System Performance**: Exporter health, collection failures, performance degradation
- **Security**: Anomalous behavior, unauthorized access, suspicious patterns

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

      - alert: SlurmDatabaseDown
        expr: slurm_database_up == 0
        for: 2m
        labels:
          severity: critical
          component: database
        annotations:
          summary: "SLURM database is unreachable"
          description: "SLURM database connection has been down for more than 2 minutes"

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

      - alert: LowDiskSpace
        expr: (slurm_node_disk_free_bytes / slurm_node_disk_total_bytes) < 0.1
        for: 5m
        labels:
          severity: critical
          component: storage
        annotations:
          summary: "Low disk space on node {{ $labels.node }}"
          description: "Disk space is {{ $value | humanizePercentage }} on node {{ $labels.node }}"
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

      - alert: HighResourceWaste
        expr: sum(rate(slurm_job_waste_cpu_hours[1h])) by (partition) > 50
        for: 30m
        labels:
          severity: warning
          component: waste
        annotations:
          summary: "High resource waste in partition {{ $labels.partition }}"
          description: "CPU waste rate is {{ $value }} hours per hour"

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

### Security and Anomaly Alerts

```yaml
# alerts/security-anomalies.yml
groups:
  - name: slurm.security
    rules:
      - alert: AnomalousJobBehavior
        expr: slurm_anomaly_score{entity_type="job"} > 0.9
        for: 1m
        labels:
          severity: warning
          component: security
        annotations:
          summary: "Anomalous job behavior detected"
          description: "Job {{ $labels.entity_id }} has anomaly score {{ $value }}"

      - alert: SuspiciousUserActivity
        expr: slurm_anomaly_score{entity_type="user"} > 0.8
        for: 5m
        labels:
          severity: warning
          component: security
        annotations:
          summary: "Suspicious user activity"
          description: "User {{ $labels.entity_id }} showing anomalous behavior (score: {{ $value }})"

      - alert: UnauthorizedAccess
        expr: increase(slurm_exporter_api_requests_total{status=~"401|403"}[10m]) > 10
        for: 2m
        labels:
          severity: critical
          component: security
        annotations:
          summary: "Multiple unauthorized access attempts"
          description: "{{ $value }} unauthorized access attempts in 10 minutes"

      - alert: ResourceQuotaViolation
        expr: slurm_account_cpu_hours_used_total > slurm_account_cpu_hours_allocated
        for: 1m
        labels:
          severity: warning
          component: quota
        annotations:
          summary: "Account quota exceeded"
          description: "Account {{ $labels.account }} has exceeded CPU quota"
```

## Advanced Alerting Configurations

### Multi-Level Alerting

```yaml
# Progressive alert escalation
route:
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'level-1'
  routes:
    - match:
        severity: critical
      receiver: 'level-2'
      group_wait: 0s
      repeat_interval: 30m
      routes:
        - match:
            component: controller
        receiver: 'level-3'
        group_wait: 0s
        repeat_interval: 15m

receivers:
  - name: 'level-1'
    email_configs:
      - to: 'team@example.com'
  - name: 'level-2' 
    email_configs:
      - to: 'oncall@example.com'
    slack_configs:
      - channel: '#alerts'
  - name: 'level-3'
    email_configs:
      - to: 'emergency@example.com'
    pagerduty_configs:
      - service_key: 'your-pagerduty-key'
```

### Time-Based Routing

```yaml
# Different handling for business hours vs off-hours
route:
  routes:
    - match_re:
        time: '^(Mon|Tue|Wed|Thu|Fri) (09|10|11|12|13|14|15|16|17):'
      receiver: 'business-hours'
    - match_re:
        time: '^(Sat|Sun)'
      receiver: 'weekend'
    - receiver: 'after-hours'

time_intervals:
  - name: business-hours
    time_intervals:
      - times:
          - start_time: '09:00'
            end_time: '18:00'
        weekdays: ['monday:friday']
```

### Conditional Alerting

```yaml
# Alert only during high activity periods
- alert: HighLatencyDuringPeak
  expr: slurm_job_wait_time_seconds > 1800 and slurm_jobs_pending > 100
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "High latency during peak usage"

# Alert with context-aware thresholds
- alert: AdaptiveHighUtilization
  expr: |
    (
      slurm_partition_cpus_allocated / slurm_partition_cpus_total > 0.9 and
      hour() >= 9 and hour() <= 17
    ) or (
      slurm_partition_cpus_allocated / slurm_partition_cpus_total > 0.7 and
      (hour() < 9 or hour() > 17)
    )
  for: 15m
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
    headers:
      X-Priority: '1'
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
    actions:
      - type: button
        text: 'View in Prometheus'
        url: '{{ .GeneratorURL }}'
      - type: button
        text: 'Silence Alert'
        url: '{{ .SilenceURL }}'
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

### Microsoft Teams

```yaml
webhook_configs:
  - url: 'https://outlook.office.com/webhook/YOUR-TEAMS-WEBHOOK'
    title: 'SLURM Alert: {{ .GroupLabels.alertname }}'
    text: |
      {{ range .Alerts }}
      **{{ .Annotations.summary }}**
      
      - **Severity:** {{ .Labels.severity }}
      - **Cluster:** {{ .Labels.cluster }}
      - **Description:** {{ .Annotations.description }}
      {{ end }}
```

## Alert Templates

### Custom Templates

Create reusable alert templates:

```yaml
# templates/slurm-alerts.tmpl
{{ define "slurm.title" }}
[{{ .Status | toUpper }}] SLURM {{ .GroupLabels.cluster }}: {{ .GroupLabels.alertname }}
{{ end }}

{{ define "slurm.description" }}
{{ range .Alerts }}
Alert: {{ .Annotations.summary }}
Severity: {{ .Labels.severity }}
Cluster: {{ .Labels.cluster }}
Component: {{ .Labels.component }}
Description: {{ .Annotations.description }}
Time: {{ .StartsAt.Format "2006-01-02 15:04:05" }}
{{ if .Annotations.runbook_url }}
Runbook: {{ .Annotations.runbook_url }}
{{ end }}
---
{{ end }}
{{ end }}

{{ define "slurm.resolved" }}
The following alerts have been resolved:
{{ range .Alerts }}
- {{ .Annotations.summary }} ({{ .Labels.cluster }})
{{ end }}
{{ end }}
```

Use templates in receivers:

```yaml
receivers:
  - name: 'slurm-team'
    email_configs:
      - to: 'team@example.com'
        subject: '{{ template "slurm.title" . }}'
        body: '{{ template "slurm.description" . }}'
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

# Silence via AlertManager UI
# Navigate to http://alertmanager:9093/#/silences
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

### Maintenance Windows

```bash
# Create maintenance silence script
#!/bin/bash
ALERTMANAGER_URL="http://alertmanager:9093"
START_TIME=$(date -d "+5 minutes" -u +"%Y-%m-%dT%H:%M:%SZ")
END_TIME=$(date -d "+2 hours" -u +"%Y-%m-%dT%H:%M:%SZ")

curl -X POST "$ALERTMANGER_URL/api/v1/silences" \
  -H "Content-Type: application/json" \
  -d "{
    \"matchers\": [
      {\"name\": \"cluster\", \"value\": \"production\"}
    ],
    \"startsAt\": \"$START_TIME\",
    \"endsAt\": \"$END_TIME\",
    \"createdBy\": \"maintenance-script\",
    \"comment\": \"Scheduled maintenance window\"
  }"
```

## Testing and Validation

### Alert Testing

```bash
# Test alert rules
promtool query instant \
  'slurm_controller_up == 0' \
  --url=http://prometheus:9090

# Validate alert rules syntax
promtool check rules alerts/*.yml

# Test AlertManager configuration
amtool config check alertmanager.yml
```

### Synthetic Alerts

Create test alerts for validation:

```yaml
# test-alerts.yml
- alert: TestAlert
  expr: vector(1)
  labels:
    severity: warning
    component: test
  annotations:
    summary: "Test alert for validation"
    description: "This is a test alert"
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

## Monitoring Alerting Health

### AlertManager Metrics

Monitor AlertManager itself:

```yaml
- alert: AlertManagerDown
  expr: up{job="alertmanager"} == 0
  for: 1m
  labels:
    severity: critical

- alert: AlertManagerHighMemory
  expr: process_resident_memory_bytes{job="alertmanager"} > 500000000
  for: 10m
  labels:
    severity: warning

- alert: AlertsNotFiring
  expr: absent(ALERTS{alertstate="firing"}) == 1
  for: 1h
  labels:
    severity: warning
  annotations:
    summary: "No alerts firing for 1 hour - possible monitoring issue"
```

### Notification Delivery Monitoring

```yaml
- alert: NotificationFailures
  expr: increase(alertmanager_notifications_failed_total[1h]) > 5
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High notification failure rate"

- alert: AlertGrouping
  expr: alertmanager_alerts > 100
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Large number of alerts may indicate grouping issues"
```

## Best Practices

### Alert Design

1. **Clear, actionable alerts** - Every alert should have a clear action
2. **Context-rich descriptions** - Include enough detail for quick diagnosis
3. **Appropriate severity levels** - Reserve critical for service-affecting issues
4. **Runbook links** - Include links to resolution procedures
5. **Avoid alert fatigue** - Tune thresholds to minimize false positives

### Alert Naming

```yaml
# Good naming convention
SlurmControllerDown         # Clear, specific
HighNodeFailureRate        # Describes condition
LowJobEfficiency           # Indicates performance issue

# Avoid generic names
SystemAlert                # Too vague
Error                      # Not descriptive
Problem                    # Unclear severity
```

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