# SLURM Exporter Operational Runbooks

This directory contains operational runbooks for troubleshooting and incident response for the SLURM Exporter production deployment.

## Quick Reference

### Emergency Contacts
- **On-Call Engineer**: +1-555-ONCALL
- **SRE Team**: sre-team@example.com
- **Security Team**: security@example.com
- **SLURM Administrators**: slurm-admins@example.com

### Critical Alert Response Matrix

| Alert Type | Severity | Response Time | Primary Responder | Escalation |
|------------|----------|---------------|-------------------|------------|
| Service Down | P0 | < 5 minutes | On-Call SRE | Engineering Manager |
| SLA Breach | P0 | < 5 minutes | On-Call SRE | VP Engineering |
| Security Incident | P0 | < 2 minutes | Security Team | CISO |
| Performance Degradation | P1 | < 30 minutes | SRE Team | Team Lead |
| Capacity Warning | P2 | < 4 hours | SRE Team | - |

### Service Health Dashboard
- **Grafana**: https://grafana.example.com/d/slurm-exporter-ops
- **Prometheus**: https://prometheus.example.com
- **Alertmanager**: https://alertmanager.example.com
- **PagerDuty**: https://example.pagerduty.com

## Runbook Index

### Critical Incidents (P0)
1. [Service Down](./service-down.md) - Complete service outage
2. [SLA Breach](./sla-breach.md) - Availability below 99.9%
3. [Security Violation](./security-violation.md) - Security policy breach
4. [Authentication Failure](./authentication-failure.md) - SLURM API auth issues
5. [Cardinality Explosion](./cardinality-explosion.md) - Metric explosion

### Performance Issues (P1)
6. [High Response Time](./performance-slow.md) - Response time > 5s
7. [High Error Rate](./error-rate-high.md) - Error rate > 1%
8. [Memory Issues](./memory-issues.md) - High memory usage
9. [SLURM Connectivity](./slurm-connectivity.md) - SLURM API issues
10. [Data Staleness](./data-stale.md) - Outdated metrics

### Operational Procedures (P2)
11. [Configuration Drift](./configuration-drift.md) - Config inconsistencies
12. [Certificate Renewal](./certificate-renewal.md) - TLS certificate management
13. [Pod Scheduling Issues](./pod-scheduling.md) - Kubernetes scheduling problems
14. [Capacity Planning](./capacity-planning.md) - Resource scaling decisions
15. [Backup and Recovery](./backup-recovery.md) - Data protection procedures

### Kubernetes Operations
16. [Rolling Updates](./rolling-updates.md) - Safe deployment procedures
17. [HPA Scaling Issues](./hpa-scaling.md) - Auto-scaling problems
18. [Network Policy Issues](./network-policy.md) - Connectivity problems
19. [PVC Management](./pvc-management.md) - Storage issues
20. [Service Discovery](./service-discovery.md) - Discovery failures

## General Incident Response Process

### 1. Alert Acknowledgment (< 2 minutes)
```bash
# Acknowledge in PagerDuty
# Join #incident-response Slack channel
# Update incident status board
```

### 2. Initial Assessment (< 5 minutes)
```bash
# Check service health dashboard
kubectl get pods -n slurm-exporter
curl https://slurm-exporter.example.com/health

# Review recent changes
git log --oneline --since="1 hour ago"
kubectl rollout history deployment/slurm-exporter -n slurm-exporter
```

### 3. Impact Analysis (< 10 minutes)
- Determine customer impact scope
- Check dependent services
- Estimate business impact
- Decide on escalation needs

### 4. Investigation & Mitigation (Variable)
- Follow specific runbook procedures
- Document all actions taken
- Communicate progress updates
- Apply immediate fixes or workarounds

### 5. Resolution & Monitoring (< 30 minutes post-fix)
- Verify service recovery
- Monitor for regression
- Update stakeholders
- Schedule post-incident review

### 6. Post-Incident Review (Within 48 hours)
- Document timeline and root cause
- Identify improvement opportunities
- Update runbooks and monitoring
- Implement preventive measures

## Common Diagnostic Commands

### Service Health Check
```bash
# Basic service status
kubectl get all -n slurm-exporter

# Pod details and events
kubectl describe pods -n slurm-exporter
kubectl get events -n slurm-exporter --sort-by='.lastTimestamp'

# Service connectivity
kubectl port-forward svc/slurm-exporter 8080:8080 -n slurm-exporter
curl http://localhost:8080/health
curl http://localhost:8080/metrics | head -20
```

### Log Analysis
```bash
# Application logs
kubectl logs deployment/slurm-exporter -n slurm-exporter --tail=100
kubectl logs deployment/slurm-exporter -n slurm-exporter --since=1h

# Previous container logs (if pod restarted)
kubectl logs deployment/slurm-exporter -n slurm-exporter --previous

# Multiple containers
kubectl logs deployment/slurm-exporter -c slurm-exporter -n slurm-exporter
```

### Resource Monitoring
```bash
# Resource usage
kubectl top pods -n slurm-exporter
kubectl top nodes

# Resource quotas and limits
kubectl describe quota -n slurm-exporter
kubectl describe limitrange -n slurm-exporter

# HPA status
kubectl get hpa -n slurm-exporter
kubectl describe hpa slurm-exporter -n slurm-exporter
```

### SLURM Integration Check
```bash
# Test SLURM API connectivity
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  curl -k "$SLURM_REST_URL/slurm/v0.0.39/ping"

# Check authentication
kubectl get secret slurm-credentials -n slurm-exporter -o yaml

# JWT token validation
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  sh -c 'echo $SLURM_JWT_TOKEN | base64 -d | head -c 100'
```

### Network Diagnostics
```bash
# Check service endpoints
kubectl get endpoints -n slurm-exporter

# Network policies
kubectl get networkpolicy -n slurm-exporter
kubectl describe networkpolicy -n slurm-exporter

# DNS resolution
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  nslookup slurm-head.cluster.local

# Connectivity tests
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  nc -zv slurm-head.cluster.local 6820
```

### Prometheus & Monitoring
```bash
# Check ServiceMonitor
kubectl get servicemonitor -n slurm-exporter
kubectl describe servicemonitor slurm-exporter -n slurm-exporter

# Prometheus targets
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
# Browse to http://localhost:9090/targets

# Alert status
kubectl exec prometheus-kube-prometheus-stack-prometheus-0 -n monitoring -- \
  promtool query instant 'ALERTS{alertname=~".*SlurmExporter.*"}'
```

## Escalation Procedures

### Level 1 - Initial Response (SRE On-Call)
- Acknowledge alerts within 5 minutes
- Perform initial troubleshooting
- Apply immediate fixes/workarounds
- Escalate to L2 if unresolved in 30 minutes

### Level 2 - Engineering Team
- Deep technical investigation
- Code-level debugging
- Complex configuration changes
- Escalate to L3 if unresolved in 2 hours

### Level 3 - Senior Engineering/Management
- Architecture decisions
- Major incident coordination
- External vendor engagement
- Executive communication

### Special Escalations
- **Security Incidents**: Immediate security team notification
- **SLA Breaches**: VP Engineering notification
- **Data Loss**: CTO notification
- **Regulatory Issues**: Legal/Compliance team notification

## Communication Templates

### Initial Incident Report
```
INCIDENT: [Brief Description]
STATUS: Investigating
IMPACT: [Customer/Business Impact]
ETA: [Estimated Resolution Time]
ASSIGNED: [Incident Commander]
NEXT UPDATE: [Time for next update]
```

### Status Update
```
INCIDENT UPDATE: [Brief Description]
STATUS: [Investigating/Mitigating/Resolved]
PROGRESS: [What has been done]
NEXT STEPS: [What will be done next]
ETA: [Updated resolution time]
NEXT UPDATE: [Time for next update]
```

### Resolution Notice
```
INCIDENT RESOLVED: [Brief Description]
RESOLUTION TIME: [Total incident duration]
ROOT CAUSE: [Brief root cause]
PREVENTION: [Steps to prevent recurrence]
PIR SCHEDULED: [Post-incident review meeting time]
```

## Tools and Access

### Required Access
- Kubernetes cluster admin access
- Grafana dashboard access
- Prometheus query access
- PagerDuty incident management
- Slack #incident-response channel
- SLURM cluster monitoring access

### Emergency Access Procedures
If normal access is unavailable:
1. Contact IT helpdesk: +1-555-HELP
2. Use emergency break-glass procedures
3. Document all emergency access usage
4. Review access logs post-incident

### Useful Tools
- `kubectl` - Kubernetes operations
- `curl` - API testing
- `jq` - JSON parsing
- `promtool` - Prometheus queries
- `htop` - Resource monitoring
- `tcpdump` - Network analysis

## Continuous Improvement

### Runbook Maintenance
- Review runbooks monthly
- Update after each incident
- Test procedures quarterly
- Gather feedback from users

### Metrics and KPIs
- Mean Time to Detection (MTTD)
- Mean Time to Recovery (MTTR)
- Alert accuracy and noise
- Runbook effectiveness
- Team response times

### Training and Drills
- Monthly incident response drills
- Quarterly disaster recovery tests
- Annual security incident simulations
- New team member onboarding

This runbook collection provides comprehensive operational guidance for maintaining production SLURM Exporter deployments with minimal downtime and maximum reliability.