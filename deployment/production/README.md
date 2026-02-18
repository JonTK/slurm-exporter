# Production Deployment Guide

This directory contains production deployment configurations and operational guidance for the SLURM Exporter.

## Directory Structure

| Directory | Description |
|-----------|-------------|
| `kubernetes-ha/` | High-availability Kubernetes deployment manifests |
| `monitoring/` | Prometheus rules, Alertmanager config, Grafana dashboards |
| `environments/` | Environment-specific config overlays (production, staging, development) |
| `runbooks/` | Operational runbooks for incident response |
| `backup/` | Backup strategy and disaster recovery definitions |
| `performance/` | Load testing and benchmarking definitions |
| `security/` | Security policies and compliance definitions |
| `incident-response/` | Incident response procedures and automation |

## Quick Start (Kubernetes)

### Prerequisites

- Kubernetes 1.19+
- SLURM cluster with REST API enabled (slurmrestd)
- Prometheus monitoring stack
- Network access to SLURM head node

### Deploy

```bash
cd kubernetes-ha
kubectl apply -f namespace.yaml
kubectl apply -f .
```

### Configure Monitoring

```bash
kubectl apply -f monitoring/prometheus-rules.yaml
kubectl apply -f monitoring/alertmanager-config.yaml
```

### Verify

```bash
# Check pods
kubectl get pods -n slurm-exporter

# Check health
kubectl port-forward svc/slurm-exporter 10341:10341 -n slurm-exporter
curl http://localhost:10341/health

# Check metrics
curl http://localhost:10341/metrics | grep slurm_
```

## Security

- Use distroless or Alpine base images (both are available)
- Run as non-root user (UID 65534)
- Set resource limits and requests
- Enable network policies (see `kubernetes-ha/networkpolicy.yaml`)
- Use Kubernetes secrets for JWT tokens
- Configure TLS for production endpoints

See [SECURITY.md](../../SECURITY.md) for the full security policy.

## Monitoring

### Key Metrics

- Exporter up/down status
- Collection duration per collector
- Collection error rate
- Process CPU and memory usage

### Alerting

Pre-built alerting rules are in `monitoring/prometheus-rules.yaml`. See also `../../prometheus/alerts/` for additional alert definitions.

## Troubleshooting

```bash
# Check exporter logs
kubectl logs deployment/slurm-exporter -n slurm-exporter

# Test SLURM API connectivity
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  curl http://slurm-api:6820/slurm/v0.0.44/ping

# Check exporter health
kubectl port-forward svc/slurm-exporter 10341:10341 -n slurm-exporter
curl http://localhost:10341/health
curl http://localhost:10341/ready
```

## Support

- **Issue Tracking**: https://github.com/jontk/slurm-exporter/issues
- **Documentation**: See [docs/](../../docs/) for configuration and troubleshooting guides
