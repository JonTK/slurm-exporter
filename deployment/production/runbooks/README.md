# SLURM Exporter Operational Runbooks

Operational runbooks for troubleshooting and incident response for SLURM Exporter production deployments.

## Available Runbooks

| Runbook | Description |
|---------|-------------|
| [Service Down](./service-down.md) | Complete service outage troubleshooting |
| [Performance Slow](./performance-slow.md) | High response time diagnosis and resolution |
| [Security Violation](./security-violation.md) | Security policy breach response |

## General Incident Response

### 1. Initial Assessment

```bash
# Check service status
kubectl get pods -n slurm-exporter
kubectl get events -n slurm-exporter --sort-by='.lastTimestamp'

# Check health
kubectl port-forward svc/slurm-exporter 10341:10341 -n slurm-exporter
curl http://localhost:10341/health
```

### 2. Check Logs

```bash
# Application logs
kubectl logs deployment/slurm-exporter -n slurm-exporter --tail=100

# Previous container logs (if restarted)
kubectl logs deployment/slurm-exporter -n slurm-exporter --previous
```

### 3. Check SLURM API Connectivity

```bash
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  curl http://slurm-api:6820/slurm/v0.0.44/ping
```

### 4. Check Resource Usage

```bash
kubectl top pods -n slurm-exporter
kubectl get hpa -n slurm-exporter
```

### 5. Check Prometheus Targets

```bash
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
# Browse to http://localhost:9090/targets
```

## Escalation

Customize these contacts for your organization:

1. **Level 1**: On-call engineer - initial triage and immediate fixes
2. **Level 2**: Engineering team - deep investigation and code-level debugging
3. **Level 3**: Senior engineering - architecture decisions and major incident coordination

## Post-Incident

- Document timeline and root cause
- Update runbooks with lessons learned
- Implement preventive measures
- Schedule post-incident review within 48 hours
