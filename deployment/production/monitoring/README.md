# Production Monitoring Setup

This directory contains comprehensive monitoring configurations for the SLURM Exporter production deployment.

## Overview

The monitoring setup provides:
- **Service Level Objectives (SLOs)** with automated SLA tracking
- **Multi-tier alerting** with escalation policies
- **Operational dashboards** for real-time visibility
- **Capacity planning** and trend analysis
- **Security monitoring** and incident response

## Components

### 1. Prometheus Rules (`prometheus-rules.yaml`)

Comprehensive alerting rules organized by priority and component:

#### Alert Categories
- **SLA & SLO Monitoring**: Availability, performance, and reliability metrics
- **Operational Health**: Service discovery, certificates, configuration drift
- **Business Logic**: Data quality, cardinality, metric freshness
- **Dependencies**: SLURM cluster health, authentication, API connectivity
- **Capacity Planning**: Resource trends, queue buildup, scaling limits
- **Security**: Access patterns, policy violations, unusual behavior

#### Alert Priorities
- **P0 (Critical)**: Immediate response required, page on-call
- **P1 (Warning)**: Standard operational workflow, business hours
- **P2 (Info)**: Capacity planning, non-urgent maintenance

### 2. Alertmanager Configuration (`alertmanager-config.yaml`)

Sophisticated alert routing with multi-channel notifications:

#### Notification Channels
- **PagerDuty**: Critical alerts with immediate escalation
- **Slack**: Team notifications with contextual information
- **Email**: Executive notifications and detailed incident reports
- **Security Team**: Dedicated security incident workflow

#### Routing Rules
- Critical alerts → Immediate PagerDuty + Slack + Email
- SLA breaches → Executive notification + Operations team
- Security alerts → Security team PagerDuty + SOC notification
- Capacity alerts → Business hours only notification

#### Inhibition Rules
- Prevent alert spam during outages
- Suppress component alerts when service is down
- Intelligent alert correlation and deduplication

### 3. Grafana Dashboard (`grafana-dashboards.json`)

Operational dashboard featuring:
- **SLA Status**: Real-time availability tracking with 99.9% target
- **Service Health**: Instance status and health monitoring
- **Performance Metrics**: Response times with 5-second SLO
- **Error Tracking**: Error rates with 1% SLO threshold
- **Cardinality Monitoring**: Metric explosion prevention

## Setup Instructions

### Prerequisites

```bash
# Verify monitoring stack is installed
kubectl get pods -n monitoring
kubectl get servicemonitor -A
kubectl get prometheusrule -A
```

### 1. Deploy Prometheus Rules

```bash
# Apply operational alerting rules
kubectl apply -f prometheus-rules.yaml

# Verify rules are loaded
kubectl get prometheusrule -n monitoring
```

### 2. Configure Alertmanager

```bash
# Create Alertmanager configuration secret
kubectl create secret generic alertmanager-config \
  --from-file=alertmanager.yml=alertmanager-config.yaml \
  -n monitoring

# Update Alertmanager to use new config
kubectl patch prometheus kube-prometheus-stack-prometheus \
  -n monitoring \
  --type='merge' \
  -p='{"spec":{"alerting":{"alertmanagers":[{"name":"alertmanager-operated","namespace":"monitoring","port":"web"}]}}}'

# Restart Alertmanager
kubectl rollout restart statefulset/alertmanager-kube-prometheus-stack-alertmanager -n monitoring
```

### 3. Import Grafana Dashboard

```bash
# Create dashboard ConfigMap
kubectl create configmap slurm-exporter-dashboard \
  --from-file=dashboard.json=grafana-dashboards.json \
  -n monitoring

# Label for Grafana auto-discovery
kubectl label configmap slurm-exporter-dashboard \
  grafana_dashboard=1 \
  -n monitoring

# Or import via Grafana UI
# 1. Open Grafana
# 2. Go to Dashboards → Import
# 3. Upload grafana-dashboards.json
# 4. Configure data source: Prometheus
```

### 4. Configure External Integrations

#### PagerDuty Setup
```bash
# Get PagerDuty integration key
PAGERDUTY_KEY="your-integration-key"

# Update Alertmanager config with PagerDuty key
sed -i "s/YOUR_PAGERDUTY_INTEGRATION_KEY/$PAGERDUTY_KEY/g" alertmanager-config.yaml
```

#### Slack Setup
```bash
# Get Slack webhook URL
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

# Update Alertmanager config with Slack webhook
sed -i "s|YOUR/SLACK/WEBHOOK|$SLACK_WEBHOOK|g" alertmanager-config.yaml
```

#### Email Setup
```bash
# Configure SMTP settings
SMTP_HOST="smtp.example.com:587"
SMTP_USER="alerts@example.com"
SMTP_PASS="your-smtp-password"

# Update global SMTP configuration in alertmanager-config.yaml
```

## Service Level Objectives (SLOs)

### Availability SLO
- **Target**: 99.9% uptime (8.77 hours downtime per year)
- **Measurement**: 24-hour rolling window
- **Alert**: < 99.9% for 5 minutes triggers critical alert

### Performance SLO
- **Target**: 95% of requests complete in < 5 seconds
- **Measurement**: 95th percentile response time
- **Alert**: > 5 seconds for 10 minutes triggers warning

### Reliability SLO
- **Target**: < 1% error rate
- **Measurement**: Error rate over 5-minute window
- **Alert**: > 1% for 15 minutes triggers warning

## Alert Playbooks

### Critical Alert Response (P0)

1. **Immediate Actions**
   - Acknowledge PagerDuty alert
   - Join incident response channel
   - Check service status dashboard

2. **Investigation**
   - Review Grafana dashboard for anomalies
   - Check recent deployments or changes
   - Examine application and infrastructure logs

3. **Communication**
   - Update incident status in Slack
   - Notify stakeholders if customer impact
   - Document investigation steps

4. **Resolution**
   - Apply immediate fix or rollback
   - Monitor service recovery
   - Conduct post-incident review

### Warning Alert Response (P1)

1. **Assessment**
   - Evaluate business impact
   - Check if issue is trending worse
   - Review related alerts

2. **Investigation**
   - Follow runbook procedures
   - Check capacity and performance metrics
   - Analyze trends and patterns

3. **Action**
   - Apply preventive measures
   - Schedule maintenance if needed
   - Update monitoring thresholds if appropriate

## Monitoring Best Practices

### Alert Hygiene
- **Actionable**: Every alert must have a clear response action
- **Contextual**: Include runbook links and troubleshooting steps
- **Prioritized**: Use consistent priority levels across all alerts
- **Correlated**: Prevent alert storms with inhibition rules

### Dashboard Design
- **User-Focused**: Show business impact, not just technical metrics
- **Layered**: High-level overview with drill-down capability
- **Consistent**: Standard colors, thresholds, and formatting
- **Automated**: Self-updating with proper templating

### SLO Management
- **Realistic**: Set achievable targets based on business needs
- **Measurable**: Use objective, quantifiable metrics
- **Business-Aligned**: Connect technical SLOs to user experience
- **Iterative**: Regularly review and adjust based on data

## Troubleshooting

### Common Issues

#### Alerts Not Firing
```bash
# Check Prometheus rules
kubectl get prometheusrule -n monitoring -o yaml

# Verify rule syntax
promtool check rules prometheus-rules.yaml

# Check Prometheus targets
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
# Browse to http://localhost:9090/targets
```

#### Alertmanager Not Sending Notifications
```bash
# Check Alertmanager config
kubectl get secret alertmanager-config -n monitoring -o yaml

# Verify Alertmanager status
kubectl logs statefulset/alertmanager-kube-prometheus-stack-alertmanager -n monitoring

# Test webhook endpoints
curl -X POST "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
  -H "Content-Type: application/json" \
  -d '{"text":"Test message"}'
```

#### Dashboard Not Loading
```bash
# Check Grafana pod status
kubectl get pods -n monitoring | grep grafana

# Check dashboard ConfigMap
kubectl get configmap slurm-exporter-dashboard -n monitoring

# Verify Grafana datasource
kubectl exec -it deployment/grafana -n monitoring -- \
  curl http://localhost:3000/api/datasources
```

### Debug Commands

```bash
# Check SLURM Exporter metrics
kubectl port-forward svc/slurm-exporter 10341:10341 -n slurm-exporter
curl http://localhost:10341/metrics | grep slurm_exporter

# Check Prometheus scraping
kubectl exec -it prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'up{job="slurm-exporter"}'

# Check alert status
kubectl exec -it prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'ALERTS{alertname=~".*SlurmExporter.*"}'
```

## Capacity Planning

### Resource Monitoring
Monitor these key metrics for capacity planning:
- **CPU Usage**: Target < 70% average utilization
- **Memory Usage**: Target < 80% to allow for spikes
- **Network I/O**: Monitor for bandwidth saturation
- **Disk Usage**: Prometheus retention and log storage

### Scaling Triggers
Automatic scaling based on:
- **Request Rate**: Scale up if > 100 req/min sustained
- **Response Time**: Scale up if 95th percentile > 5s
- **Error Rate**: Scale up if error rate > 1%
- **Resource Utilization**: Scale up if CPU > 70% or Memory > 80%

### Growth Planning
Review monthly:
- **Metric Cardinality Growth**: Plan for 20% monthly growth
- **Request Volume Trends**: Monitor cluster size correlation
- **Storage Requirements**: Prometheus and log retention needs
- **Network Bandwidth**: Inter-service communication growth

This monitoring setup provides enterprise-grade observability with proactive alerting and comprehensive operational visibility for production SLURM Exporter deployments.