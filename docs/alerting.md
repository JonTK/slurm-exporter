# SLURM Exporter Alerting Guide

This guide provides production-ready alerting rules for monitoring SLURM clusters with Prometheus, organized by severity and use case.

## Table of Contents

- [Alert Severity Levels](#alert-severity-levels)
- [Critical Alerts](#critical-alerts)
- [Warning Alerts](#warning-alerts)
- [Informational Alerts](#informational-alerts)
- [Alert Routing](#alert-routing)
- [Integration Examples](#integration-examples)
- [Best Practices](#best-practices)
- [Troubleshooting Alerts](#troubleshooting-alerts)

## Alert Severity Levels

### Critical (P1)
- **Response Time**: Immediate (within 5 minutes)
- **Examples**: Controller down, mass node failures, data loss risk
- **Notification**: Page on-call engineer, create incident

### Warning (P2)
- **Response Time**: Within 1 hour
- **Examples**: High utilization, degraded performance, approaching limits
- **Notification**: Email team, Slack alert

### Info (P3)
- **Response Time**: Next business day
- **Examples**: Maintenance reminders, efficiency concerns, trends
- **Notification**: Dashboard, daily summary

## Critical Alerts

### SLURM Controller Availability

```yaml
groups:
  - name: slurm_critical
    interval: 30s
    rules:
      - alert: SlurmControllerDown
        expr: up{job="slurm-exporter"} == 0 or slurm_cluster_info == 0
        for: 5m
        labels:
          severity: critical
          team: hpc-ops
          component: slurm-controller
        annotations:
          summary: "SLURM controller is not responding"
          description: |
            The SLURM controller has been unreachable for 5 minutes.
            Cluster: {{ $labels.cluster_name }}
            Impact: No new jobs can be scheduled
            Action: Check controller services and network connectivity
          runbook_url: "https://wiki.company.com/slurm/controller-down"
          dashboard_url: "https://grafana.company.com/d/slurm-health"

      - alert: SlurmControllerFailover
        expr: |
          slurm_cluster_info{control_host!=""} 
          and 
          (slurm_cluster_info{control_host=""} offset 5m) != 0
        for: 1m
        labels:
          severity: critical
          team: hpc-ops
          component: slurm-controller
        annotations:
          summary: "SLURM controller failover detected"
          description: |
            SLURM controller has failed over from primary to backup.
            New controller: {{ $labels.control_host }}
            Action: Investigate primary controller failure
          runbook_url: "https://wiki.company.com/slurm/controller-failover"
```

### Mass Node Failures

```yaml
      - alert: MassNodeFailure
        expr: |
          (
            sum(slurm_node_state{state="down"} == 1) 
            / 
            sum(slurm_node_state == 1)
          ) > 0.1
        for: 10m
        labels:
          severity: critical
          team: hpc-ops
          component: compute-nodes
        annotations:
          summary: "More than 10% of compute nodes are down"
          description: |
            {{ $value | humanizePercentage }} of cluster nodes are in DOWN state.
            Total nodes down: {{ printf "sum(slurm_node_state{state='down'} == 1)" | query | first | value }}
            Impact: Severe capacity reduction
            Action: Check for infrastructure issues (power, network, storage)
          runbook_url: "https://wiki.company.com/slurm/mass-node-failure"

      - alert: EntirePartitionDown
        expr: |
          sum by (partition) (slurm_partition_nodes_total{state="down"}) 
          == 
          sum by (partition) (slurm_partition_nodes_total{state="total"})
        for: 5m
        labels:
          severity: critical
          team: hpc-ops
          component: partition
        annotations:
          summary: "Entire partition {{ $labels.partition }} is down"
          description: |
            All nodes in partition {{ $labels.partition }} are down.
            Total nodes affected: {{ $value }}
            Impact: No jobs can run in this partition
            Action: Check partition-specific infrastructure
```

### Scheduler Failures

```yaml
      - alert: SchedulerStalled
        expr: |
          time() - slurm_exporter_last_collection_timestamp{collector="scheduler"} > 600
          or
          increase(slurm_scheduler_cycle_duration_seconds_count[5m]) == 0
        for: 5m
        labels:
          severity: critical
          team: hpc-ops
          component: scheduler
        annotations:
          summary: "SLURM scheduler is not processing jobs"
          description: |
            The SLURM scheduler has not completed a cycle in over 10 minutes.
            Last update: {{ $value | humanizeDuration }} ago
            Impact: No new jobs will start
            Action: Check slurmctld logs and restart if necessary

      - alert: SchedulerPerformanceCritical
        expr: |
          histogram_quantile(0.95, 
            rate(slurm_scheduler_cycle_duration_seconds_bucket[5m])
          ) > 30
        for: 10m
        labels:
          severity: critical
          team: hpc-ops
          component: scheduler
        annotations:
          summary: "SLURM scheduler cycle time critically high"
          description: |
            95th percentile scheduler cycle time is {{ $value }}s (threshold: 30s).
            Impact: Severe job start delays
            Action: Review scheduler configuration and job queue depth
```

### Data Loss Risk

```yaml
      - alert: AccountingDatabaseDown
        expr: |
          slurm_exporter_api_requests_total{endpoint="/slurmdb",status!="200"} 
          / 
          slurm_exporter_api_requests_total{endpoint="/slurmdb"} > 0.95
        for: 15m
        labels:
          severity: critical
          team: hpc-ops
          component: accounting
        annotations:
          summary: "SLURM accounting database is unavailable"
          description: |
            Over 95% of requests to SLURM accounting database are failing.
            Error rate: {{ $value | humanizePercentage }}
            Impact: Job accounting data may be lost
            Action: Check slurmdbd service and MySQL database

      - alert: ExporterDataCollectionFailure
        expr: |
          sum(rate(slurm_exporter_collection_errors_total[5m])) by (collector) > 1
        for: 10m
        labels:
          severity: critical
          team: hpc-ops
          component: monitoring
        annotations:
          summary: "SLURM exporter failing to collect {{ $labels.collector }} metrics"
          description: |
            The {{ $labels.collector }} collector is experiencing {{ $value }} errors per second.
            Impact: Missing metrics for monitoring and alerting
            Action: Check exporter logs and SLURM API connectivity
```

## Warning Alerts

### Resource Utilization

```yaml
groups:
  - name: slurm_warning
    interval: 60s
    rules:
      - alert: HighCPUUtilization
        expr: |
          (
            slurm_cluster_cpus_total{state="allocated"} 
            / 
            slurm_cluster_cpus_total{state="total"}
          ) > 0.90
        for: 30m
        labels:
          severity: warning
          team: hpc-ops
          component: capacity
        annotations:
          summary: "Cluster CPU utilization above 90%"
          description: |
            CPU utilization has been {{ $value | humanizePercentage }} for 30 minutes.
            Allocated CPUs: {{ printf "slurm_cluster_cpus_total{state='allocated'}" | query | first | value }}
            Total CPUs: {{ printf "slurm_cluster_cpus_total{state='total'}" | query | first | value }}
            Impact: Limited capacity for new jobs
            Action: Consider scaling or job prioritization

      - alert: HighMemoryUtilization
        expr: |
          (
            slurm_cluster_memory_total_bytes{state="allocated"} 
            / 
            slurm_cluster_memory_total_bytes{state="total"}
          ) > 0.85
        for: 30m
        labels:
          severity: warning
          team: hpc-ops
          component: capacity
        annotations:
          summary: "Cluster memory utilization above 85%"
          description: |
            Memory utilization: {{ $value | humanizePercentage }}
            Allocated: {{ printf "slurm_cluster_memory_total_bytes{state='allocated'}" | query | first | value | humanize1024 }}
            Total: {{ printf "slurm_cluster_memory_total_bytes{state='total'}" | query | first | value | humanize1024 }}

      - alert: GPUShortage
        expr: |
          (
            sum(slurm_node_gpus_allocated) 
            / 
            sum(slurm_node_gpus_total)
          ) > 0.95
        for: 1h
        labels:
          severity: warning
          team: hpc-ops
          component: gpu-resources
        annotations:
          summary: "GPU utilization above 95%"
          description: |
            {{ $value | humanizePercentage }} of GPUs are allocated.
            Available GPUs: {{ printf "sum(slurm_node_gpus_total - slurm_node_gpus_allocated)" | query | first | value }}
            Impact: GPU jobs may experience long wait times
```

### Queue Depth

```yaml
      - alert: ExcessivePendingJobs
        expr: slurm_cluster_jobs_total{state="pending"} > 1000
        for: 1h
        labels:
          severity: warning
          team: hpc-ops
          component: scheduler
        annotations:
          summary: "Over 1000 jobs pending for 1 hour"
          description: |
            Current pending jobs: {{ $value }}
            Running jobs: {{ printf "slurm_cluster_jobs_total{state='running'}" | query | first | value }}
            Action: Review job priorities and resource availability

      - alert: LongJobWaitTimes
        expr: |
          histogram_quantile(0.90, 
            rate(slurm_job_wait_time_seconds_bucket[1h])
          ) > 14400  # 4 hours
        for: 30m
        labels:
          severity: warning
          team: hpc-ops
          component: scheduler
        annotations:
          summary: "90% of jobs waiting more than 4 hours"
          description: |
            90th percentile wait time: {{ $value | humanizeDuration }}
            Action: Review scheduler configuration and fairshare policies

      - alert: JobBacklogGrowing
        expr: |
          predict_linear(slurm_cluster_jobs_total{state="pending"}[2h], 3600) 
          > 
          slurm_cluster_jobs_total{state="pending"} * 1.5
        for: 30m
        labels:
          severity: warning
          team: hpc-ops
          component: capacity
        annotations:
          summary: "Job backlog growing rapidly"
          description: |
            Pending jobs projected to increase 50% in next hour.
            Current: {{ printf "slurm_cluster_jobs_total{state='pending'}" | query | first | value }}
            Projected: {{ $value }}
```

### Node Health

```yaml
      - alert: NodesInDrainState
        expr: |
          sum(slurm_node_state{state="drain"} == 1) > 10
        for: 24h
        labels:
          severity: warning
          team: hpc-ops
          component: maintenance
        annotations:
          summary: "{{ $value }} nodes in drain state for 24 hours"
          description: |
            Nodes have been draining for extended period.
            May indicate forgotten maintenance or issues preventing drain completion.
            Action: Review drain reasons and complete maintenance

      - alert: HighNodeTemperature
        expr: slurm_node_temperature_celsius{sensor="cpu0"} > 85
        for: 15m
        labels:
          severity: warning
          team: hpc-ops
          component: infrastructure
        annotations:
          summary: "High temperature on node {{ $labels.node }}"
          description: |
            CPU temperature: {{ $value }}°C (threshold: 85°C)
            Sensor: {{ $labels.sensor }}
            Action: Check cooling system and node workload

      - alert: NodeMemoryPressure
        expr: |
          (
            slurm_node_memory_allocated_bytes 
            / 
            slurm_node_memory_total_bytes
          ) > 0.95
        for: 30m
        labels:
          severity: warning
          team: hpc-ops
          component: node-health
        annotations:
          summary: "High memory usage on node {{ $labels.node }}"
          description: |
            Memory utilization: {{ $value | humanizePercentage }}
            Allocated: {{ printf "slurm_node_memory_allocated_bytes{node='%s'}" $labels.node | query | first | value | humanize1024 }}
            Action: Monitor for OOM kills and memory leaks
```

### Performance Degradation

```yaml
      - alert: SchedulerSlowdown
        expr: |
          histogram_quantile(0.95, 
            rate(slurm_scheduler_cycle_duration_seconds_bucket[10m])
          ) > 10
        for: 20m
        labels:
          severity: warning
          team: hpc-ops
          component: scheduler
        annotations:
          summary: "SLURM scheduler performance degraded"
          description: |
            95th percentile cycle time: {{ $value }}s (normal: <5s)
            May cause job start delays
            Action: Review scheduler logs and configuration

      - alert: BackfillIneffective
        expr: |
          rate(slurm_backfill_jobs_total[1h]) < 1 
          and 
          slurm_cluster_jobs_total{state="pending"} > 100
        for: 2h
        labels:
          severity: warning
          team: hpc-ops
          component: scheduler
        annotations:
          summary: "Backfill scheduler not filling gaps"
          description: |
            Less than 1 job per hour via backfill despite {{ printf "slurm_cluster_jobs_total{state='pending'}" | query | first | value }} pending jobs.
            Action: Review backfill configuration and job constraints

      - alert: LowJobEfficiency
        expr: |
          histogram_quantile(0.50, 
            rate(slurm_job_efficiency_percentage_bucket[2h])
          ) < 50
        for: 4h
        labels:
          severity: warning
          team: hpc-users
          component: efficiency
        annotations:
          summary: "Median job efficiency below 50%"
          description: |
            Half of all jobs are using less than 50% of allocated resources.
            Current median efficiency: {{ $value }}%
            Impact: Wasted resources and longer queue times
            Action: User education and policy enforcement
```

### Resource Quotas

```yaml
      - alert: AccountApproachingQuota
        expr: |
          (
            slurm_account_usage_cpu_hours 
            / 
            slurm_account_quota_cpu_hours
          ) > 0.80
        for: 1h
        labels:
          severity: warning
          team: hpc-admin
          component: accounting
        annotations:
          summary: "Account {{ $labels.account }} at {{ $value | humanizePercentage }} of CPU quota"
          description: |
            Usage: {{ printf "slurm_account_usage_cpu_hours{account='%s'}" $labels.account | query | first | value | humanize }}
            Quota: {{ printf "slurm_account_quota_cpu_hours{account='%s'}" $labels.account | query | first | value | humanize }}
            Period: {{ $labels.period }}
            Action: Notify account owner

      - alert: UserExcessiveResourceUse
        expr: |
          slurm_user_cpus_allocated > 1000 
          or 
          slurm_user_memory_allocated_bytes > 10*1024^4  # 10TB
        for: 4h
        labels:
          severity: warning
          team: hpc-admin
          component: fairshare
        annotations:
          summary: "User {{ $labels.user }} using excessive resources"
          description: |
            CPUs allocated: {{ printf "slurm_user_cpus_allocated{user='%s'}" $labels.user | query | first | value }}
            Memory allocated: {{ printf "slurm_user_memory_allocated_bytes{user='%s'}" $labels.user | query | first | value | humanize1024 }}
            Duration: 4+ hours
            Action: Review user's jobs and fairshare impact
```

## Informational Alerts

### Maintenance and Lifecycle

```yaml
groups:
  - name: slurm_info
    interval: 300s
    rules:
      - alert: NodeMaintenanceReminder
        expr: |
          (time() - slurm_node_drain_timestamp) > 86400 * 7  # 7 days
          and
          slurm_node_state{state="drain"} == 1
        labels:
          severity: info
          team: hpc-ops
          component: maintenance
        annotations:
          summary: "Node {{ $labels.node }} in drain state for over 7 days"
          description: |
            Node has been draining for {{ $value | humanizeDuration }}.
            Action: Complete maintenance or undrain if no longer needed

      - alert: SlurmVersionMismatch
        expr: |
          count(count by (version) (slurm_cluster_info)) > 1
        labels:
          severity: info
          team: hpc-ops
          component: configuration
        annotations:
          summary: "Multiple SLURM versions detected"
          description: |
            Different SLURM versions running in environment.
            Versions: {{ range query "count by (version) (slurm_cluster_info)" }}{{ .Labels.version }} {{ end }}
            Action: Plan upgrade to consistent version

      - alert: CertificateExpiringSoon
        expr: |
          (slurm_tls_cert_expiry_timestamp - time()) < 86400 * 30  # 30 days
        labels:
          severity: info
          team: hpc-ops
          component: security
        annotations:
          summary: "TLS certificate expiring in {{ $value | humanizeDuration }}"
          description: |
            Certificate for {{ $labels.subject }} expires on {{ $labels.expiry_date }}.
            Action: Renew certificate before expiration
```

### Capacity Planning

```yaml
      - alert: ProjectedCapacityShortage
        expr: |
          predict_linear(
            slurm_cluster_cpus_total{state="allocated"}[7d], 
            86400 * 30  # 30 days
          ) 
          > 
          slurm_cluster_cpus_total{state="total"} * 0.95
        labels:
          severity: info
          team: hpc-planning
          component: capacity
        annotations:
          summary: "CPU capacity projected to reach 95% in 30 days"
          description: |
            Based on 7-day trend, CPU allocation will reach critical levels.
            Current utilization: {{ printf "(slurm_cluster_cpus_total{state='allocated'} / slurm_cluster_cpus_total{state='total'})" | query | first | value | humanizePercentage }}
            Action: Plan capacity expansion

      - alert: PartitionUnderutilized
        expr: |
          avg_over_time(
            (slurm_partition_cpus_total{state="allocated"} / slurm_partition_cpus_total{state="total"})
            [7d]
          ) < 0.30
        labels:
          severity: info
          team: hpc-planning
          component: efficiency
        annotations:
          summary: "Partition {{ $labels.partition }} averaging {{ $value | humanizePercentage }} utilization"
          description: |
            7-day average CPU utilization is below 30%.
            Consider consolidating partitions or adjusting policies.
```

### User Behavior

```yaml
      - alert: FrequentJobFailures
        expr: |
          (
            sum by (user) (increase(slurm_user_jobs_total{state="failed"}[24h]))
            /
            sum by (user) (increase(slurm_user_jobs_total[24h]))
          ) > 0.25
          and
          sum by (user) (increase(slurm_user_jobs_total[24h])) > 10
        labels:
          severity: info
          team: hpc-support
          component: user-support
        annotations:
          summary: "User {{ $labels.user }} has {{ $value | humanizePercentage }} job failure rate"
          description: |
            More than 25% of jobs failing in last 24 hours.
            Total jobs: {{ printf "sum(increase(slurm_user_jobs_total{user='%s'}[24h]))" $labels.user | query | first | value }}
            Failed jobs: {{ printf "sum(increase(slurm_user_jobs_total{user='%s',state='failed'}[24h]))" $labels.user | query | first | value }}
            Action: Reach out to offer assistance

      - alert: IneffientResourceRequests
        expr: |
          avg by (user) (
            slurm_job_cpus_requested / slurm_job_cpus_used
          ) > 4
          and
          count by (user) (slurm_job_info) > 5
        labels:
          severity: info
          team: hpc-support
          component: efficiency
        annotations:
          summary: "User {{ $labels.user }} overestimating resource needs"
          description: |
            Average CPU request is {{ $value }}x actual usage.
            Impact: Longer wait times for all users
            Action: Provide guidance on resource estimation
```

## Alert Routing

### Prometheus Alertmanager Configuration

```yaml
# alertmanager.yml
global:
  resolve_timeout: 5m
  slack_api_url: 'YOUR_SLACK_WEBHOOK_URL'
  pagerduty_url: 'https://events.pagerduty.com/v2/enqueue'

templates:
  - '/etc/alertmanager/templates/*.tmpl'

route:
  group_by: ['cluster', 'alertname']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'default'
  
  routes:
    # Critical alerts - immediate paging
    - match:
        severity: critical
      receiver: pagerduty
      group_wait: 10s
      repeat_interval: 1h
      continue: true
    
    # Critical alerts also go to Slack
    - match:
        severity: critical
      receiver: slack-critical
      
    # Warning alerts - team channel
    - match:
        severity: warning
      receiver: slack-warnings
      group_wait: 5m
      
    # Info alerts - daily digest
    - match:
        severity: info
      receiver: email-daily
      group_interval: 24h
      repeat_interval: 24h

receivers:
  - name: 'default'
    # Fallback - should not receive alerts
    
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
        description: '{{ template "pagerduty.default.description" . }}'
        details:
          cluster: '{{ .GroupLabels.cluster }}'
          component: '{{ .CommonLabels.component }}'
          
  - name: 'slack-critical'
    slack_configs:
      - channel: '#hpc-critical'
        title: '🚨 {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        send_resolved: true
        
  - name: 'slack-warnings'
    slack_configs:
      - channel: '#hpc-alerts'
        title: '⚠️ {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        send_resolved: true
        
  - name: 'email-daily'
    email_configs:
      - to: 'hpc-team@company.com'
        from: 'alertmanager@company.com'
        headers:
          Subject: 'HPC Daily Alert Summary'
        html: '{{ template "email.default.html" . }}'

inhibit_rules:
  # Don't alert on node failures if controller is down
  - source_match:
      alertname: 'SlurmControllerDown'
    target_match_re:
      alertname: '(NodeDown|SchedulerStalled)'
    equal: ['cluster']
    
  # Suppress utilization warnings during maintenance
  - source_match:
      alertname: 'PlannedMaintenance'
    target_match_re:
      severity: '(warning|info)'
    equal: ['cluster']
```

### Alert Templates

```yaml
# templates/slack.tmpl
{{ define "slack.default.title" }}
[{{ .Status | toUpper }}{{ if eq .Status "firing" }}:{{ .Alerts.Firing | len }}{{ end }}] {{ .GroupLabels.alertname }}
{{ end }}

{{ define "slack.default.text" }}
{{ range .Alerts }}
*Alert:* {{ .Labels.alertname }}
*Cluster:* {{ .Labels.cluster }}
*Summary:* {{ .Annotations.summary }}
*Description:* {{ .Annotations.description }}
*Severity:* {{ .Labels.severity }}
{{ if .Labels.node }}*Node:* {{ .Labels.node }}{{ end }}
{{ if .Labels.partition }}*Partition:* {{ .Labels.partition }}{{ end }}
{{ if .Labels.user }}*User:* {{ .Labels.user }}{{ end }}
*Source:* <{{ .GeneratorURL }}|Prometheus>
{{ if .Annotations.runbook_url }}*Runbook:* <{{ .Annotations.runbook_url }}|View>{{ end }}
{{ if .Annotations.dashboard_url }}*Dashboard:* <{{ .Annotations.dashboard_url }}|View>{{ end }}
{{ end }}
{{ end }}

# templates/email.tmpl
{{ define "email.default.subject" }}
[{{ .Status | toUpper }}] HPC Cluster Alerts - {{ .GroupLabels.cluster }}
{{ end }}

{{ define "email.default.html" }}
<!DOCTYPE html>
<html>
<head>
    <style>
        .alert-critical { color: #d73027; }
        .alert-warning { color: #fee08b; }
        .alert-info { color: #3288bd; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h2>HPC Cluster Alert Summary</h2>
    <p>Generated: {{ .ExternalURL }}</p>
    
    <table>
        <tr>
            <th>Alert</th>
            <th>Severity</th>
            <th>Component</th>
            <th>Summary</th>
            <th>Duration</th>
        </tr>
        {{ range .Alerts }}
        <tr>
            <td>{{ .Labels.alertname }}</td>
            <td class="alert-{{ .Labels.severity }}">{{ .Labels.severity }}</td>
            <td>{{ .Labels.component }}</td>
            <td>{{ .Annotations.summary }}</td>
            <td>{{ .StartsAt.Sub .EndsAt }}</td>
        </tr>
        {{ end }}
    </table>
    
    <h3>Alert Details</h3>
    {{ range .Alerts }}
    <h4>{{ .Labels.alertname }}</h4>
    <p><strong>Description:</strong> {{ .Annotations.description }}</p>
    {{ if .Annotations.runbook_url }}
    <p><strong>Runbook:</strong> <a href="{{ .Annotations.runbook_url }}">{{ .Annotations.runbook_url }}</a></p>
    {{ end }}
    <hr>
    {{ end }}
</body>
</html>
{{ end }}
```

## Integration Examples

### PagerDuty Integration

```yaml
# Additional PagerDuty configuration
pagerduty_configs:
  - service_key: 'YOUR_SERVICE_KEY'
    client: 'Prometheus'
    client_url: '{{ template "pagerduty.default.clientURL" . }}'
    description: '{{ template "pagerduty.default.description" . }}'
    severity: '{{ if eq .GroupLabels.severity "critical" }}critical{{ else }}warning{{ end }}'
    details:
      firing: '{{ template "pagerduty.default.instances" .Alerts.Firing }}'
      resolved: '{{ template "pagerduty.default.instances" .Alerts.Resolved }}'
      num_firing: '{{ .Alerts.Firing | len }}'
      num_resolved: '{{ .Alerts.Resolved | len }}'
    class: '{{ .GroupLabels.component }}'
    component: 'slurm-{{ .GroupLabels.cluster }}'
    group: '{{ .GroupLabels.job }}'
```

### Webhook Integration

```python
# webhook_handler.py
from flask import Flask, request, jsonify
import json
import logging

app = Flask(__name__)
logger = logging.getLogger(__name__)

@app.route('/webhook/alerts', methods=['POST'])
def handle_alert():
    data = request.json
    
    for alert in data.get('alerts', []):
        alert_name = alert['labels'].get('alertname')
        severity = alert['labels'].get('severity')
        status = alert['status']
        
        if severity == 'critical' and status == 'firing':
            # Trigger automatic remediation
            handle_critical_alert(alert)
        elif severity == 'warning':
            # Create ticket in service desk
            create_ticket(alert)
        
        # Log all alerts
        logger.info(f"Alert {alert_name}: {status}")
    
    return jsonify({'status': 'processed'}), 200

def handle_critical_alert(alert):
    """Automatic remediation for critical alerts"""
    alert_name = alert['labels'].get('alertname')
    
    if alert_name == 'NodeDown':
        node = alert['labels'].get('node')
        # Attempt node restart
        restart_node(node)
    elif alert_name == 'SchedulerStalled':
        # Restart scheduler service
        restart_scheduler()

def create_ticket(alert):
    """Create service desk ticket for warnings"""
    # Integration with ServiceNow, Jira, etc.
    pass
```

### Grafana Annotations

```yaml
# Prometheus recording rules for Grafana annotations
groups:
  - name: annotations
    interval: 60s
    rules:
      - record: deployment_info
        expr: |
          max by (cluster, version) (
            slurm_cluster_info
          )
        labels:
          __annotation__: "deployment"
          description: "SLURM version {{ $labels.version }} deployed"
          
      - record: maintenance_windows
        expr: |
          count by (cluster) (
            slurm_node_state{state="drain"} == 1
          ) > 0
        labels:
          __annotation__: "maintenance"
          description: "Maintenance in progress"
```

## Best Practices

### Alert Design Principles

1. **Actionable Alerts**
   - Every alert should have a clear action
   - Include runbook links for complex issues
   - Avoid alerts that require no action

2. **Appropriate Severity**
   - Critical: Immediate action required
   - Warning: Proactive intervention needed
   - Info: Awareness and planning

3. **Reduce Alert Fatigue**
   ```yaml
   # Use appropriate time windows
   for: 5m   # Critical - quick detection
   for: 30m  # Warning - avoid transient issues
   for: 2h   # Info - established trends
   ```

4. **Meaningful Thresholds**
   ```yaml
   # Based on historical data
   - record: cpu_utilization_p95_7d
     expr: |
       quantile_over_time(0.95,
         slurm_cluster_cpus_total{state="allocated"} 
         / 
         slurm_cluster_cpus_total{state="total"}
       [7d:1h])
   ```

### Testing Alerts

1. **Unit Testing with promtool**
   ```bash
   # Validate alert syntax
   promtool check rules alerts/*.yml
   
   # Test specific scenarios
   promtool test rules tests/*.yml
   ```

2. **Test File Example**
   ```yaml
   # tests/node_failure.yml
   rule_files:
     - ../alerts/critical.yml
   
   evaluation_interval: 1m
   
   tests:
     - interval: 1m
       input_series:
         - series: 'slurm_node_state{node="compute001",state="down"}'
           values: '0 0 0 1 1 1 1 1 1 1'
         - series: 'slurm_node_state{node="compute001",state="idle"}'  
           values: '1 1 1 0 0 0 0 0 0 0'
           
       alert_rule_test:
         - eval_time: 10m
           alertname: NodeDown
           exp_alerts:
             - exp_labels:
                 node: compute001
                 severity: warning
   ```

3. **Staging Environment Testing**
   ```yaml
   # Deploy alerts to staging first
   - name: slurm_alerts_staging
     interval: 30s
     rules:
       - alert: TEST_HighCPUUtilization
         expr: |
           (slurm_cluster_cpus_total{state="allocated",cluster="staging"} 
            / 
            slurm_cluster_cpus_total{state="total",cluster="staging"}) > 0.50
         for: 5m
         labels:
           severity: info
           environment: staging
   ```

### Maintenance Mode

```yaml
# Silence alerts during maintenance
- record: maintenance_mode
  expr: |
    max_over_time(
      slurm_maintenance_window{cluster="prod"}[5m]
    ) == 1
    
# Modify alerts to check maintenance mode
- alert: NodeDown
  expr: |
    slurm_node_state{state="down"} == 1
    and
    maintenance_mode != 1
  for: 10m
```

### SLO-based Alerts

```yaml
# Define SLOs
- record: job_start_time_slo
  expr: |
    histogram_quantile(0.95,
      rate(slurm_job_wait_time_seconds_bucket[1h])
    ) < 3600  # 95% of jobs start within 1 hour
    
# Alert on SLO violations  
- alert: JobStartTimeSLOViolation
  expr: |
    avg_over_time(job_start_time_slo[1h]) < 0.95
  for: 30m
  labels:
    severity: warning
    slo: job_start_time
  annotations:
    summary: "Job start time SLO violation"
    description: |
      Less than 95% of jobs starting within 1 hour target.
      Current success rate: {{ $value | humanizePercentage }}
```

## Troubleshooting Alerts

### Common Issues

1. **Too Many Alerts**
   - Increase `for` duration
   - Adjust thresholds based on percentiles
   - Use inhibition rules
   - Group related alerts

2. **Missing Alerts**
   - Check metric availability
   - Verify label matching
   - Review time windows
   - Test with lower thresholds

3. **Flapping Alerts**
   ```yaml
   # Add hysteresis
   - alert: HighCPUUtilization
     expr: |
       (
         slurm_cluster_cpus_total{state="allocated"} 
         / 
         slurm_cluster_cpus_total{state="total"}
       ) > 0.90
     for: 30m
     # Don't resolve until below 85%
     # Implementation varies by Prometheus version
   ```

4. **Alert Delays**
   - Check evaluation_interval
   - Review group_wait settings
   - Verify time synchronization
   - Monitor Prometheus performance

### Debugging Commands

```bash
# Check alert state
curl -s http://prometheus:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname=="NodeDown")'

# Test alert routing
amtool config routes test --config.file=alertmanager.yml \
  severity=critical cluster=prod alertname=NodeDown

# Verify metric presence
curl -s http://prometheus:9090/api/v1/query?query=slurm_node_state | jq '.data.result | length'

# Check for silence rules
amtool silence query --alertmanager.url=http://alertmanager:9093

# Validate configuration
promtool check config prometheus.yml
amtool check-config alertmanager.yml
```

This comprehensive alerting guide provides production-ready monitoring for SLURM clusters with appropriate severity levels, routing, and integration options.