# Kubernetes High Availability Deployment

This directory contains production-ready Kubernetes manifests for deploying SLURM Exporter in a high-availability configuration.

## Overview

The HA deployment provides:
- **Multi-replica deployment** with pod anti-affinity
- **Horizontal and vertical auto-scaling**
- **Network policies** for security
- **Service mesh integration** (Istio support)
- **Comprehensive monitoring** with Prometheus
- **Load balancing** with multiple service types
- **Security hardening** with RBAC and security contexts

## Quick Start

### Prerequisites

```bash
# Verify cluster requirements
kubectl version --short
kubectl get nodes

# Required Kubernetes version: 1.19+
# Required nodes: 3+ for proper HA distribution
```

### 1. Prepare Secrets

```bash
# Create SLURM credentials secret
kubectl create secret generic slurm-credentials \
  --namespace=slurm-exporter \
  --from-literal=rest-url="https://your-slurm-head:6820" \
  --from-literal=auth-type="jwt" \
  --from-literal=jwt-token="$(scontrol token username=exporter)"

# Or for API key authentication
kubectl create secret generic slurm-credentials \
  --namespace=slurm-exporter \
  --from-literal=rest-url="https://your-slurm-head:6820" \
  --from-literal=auth-type="api_key" \
  --from-literal=api-key="your-api-key"
```

### 2. Deploy with Kustomize

```bash
# Deploy everything at once
kubectl apply -k .

# Or deploy step by step
kubectl apply -f namespace.yaml
kubectl apply -f secrets.yaml
kubectl apply -f configmap.yaml
kubectl apply -f serviceaccount.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f hpa.yaml
kubectl apply -f networkpolicy.yaml
kubectl apply -f monitoring.yaml
kubectl apply -f ingress.yaml
```

### 3. Verify Deployment

```bash
# Check deployment status
kubectl get all -n slurm-exporter

# Check pod distribution
kubectl get pods -n slurm-exporter -o wide

# Check auto-scaling
kubectl get hpa -n slurm-exporter

# Check service endpoints
kubectl get endpoints -n slurm-exporter

# Test connectivity
kubectl port-forward svc/slurm-exporter 10341:10341 -n slurm-exporter
curl http://localhost:10341/health
curl http://localhost:10341/metrics | head -20
```

## Architecture

### Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Ingress/LB    │    │   Prometheus    │    │   Grafana       │
│                 │    │   Monitoring    │    │   Dashboards    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          ▼                      ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Services                         │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────────────┐ │
│  │   ClusterIP   │ │   Headless    │ │    LoadBalancer       │ │
│  │   (Internal)  │ │  (Discovery)  │ │     (External)        │ │
│  └───────┬───────┘ └───────┬───────┘ └───────┬───────────────┘ │
└──────────┼─────────────────┼─────────────────┼─────────────────┘
           │                 │                 │
           ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                     SLURM Exporter Pods                        │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐          │
│  │   Pod-1     │   │   Pod-2     │   │   Pod-3     │          │
│  │   Node-A    │   │   Node-B    │   │   Node-C    │          │
│  └─────┬───────┘   └─────┬───────┘   └─────┬───────┘          │
└────────┼───────────────────┼───────────────────┼─────────────────┘
         │                 │                 │
         ▼                 ▼                 ▼
    ┌─────────────────────────────────────────────────┐
    │              SLURM Cluster                      │
    │         ┌─────────────────┐                     │
    │         │   REST API      │                     │
    │         │   Port 6820     │                     │
    │         └─────────────────┘                     │
    └─────────────────────────────────────────────────┘
```

### High Availability Features

1. **Pod Distribution**
   - Anti-affinity rules spread pods across nodes and zones
   - Minimum 3 replicas for quorum
   - Graceful handling of node failures

2. **Auto-scaling**
   - Horizontal Pod Autoscaler (HPA) based on CPU/memory/custom metrics
   - Vertical Pod Autoscaler (VPA) for resource optimization
   - Pod Disruption Budget (PDB) for maintenance windows

3. **Load Balancing**
   - ClusterIP for internal communication
   - LoadBalancer for external access
   - Headless service for direct pod discovery

4. **Health Monitoring**
   - Startup, liveness, and readiness probes
   - ServiceMonitor for Prometheus scraping
   - PrometheusRule for alerting

## Configuration

### Environment Variables

The deployment uses these environment sources:

1. **ConfigMaps**:
   - `slurm-exporter-config`: Main application configuration
   - `slurm-exporter-env`: Environment-specific settings
   - `slurm-exporter-scripts`: Health check and lifecycle scripts

2. **Secrets**:
   - `slurm-credentials`: SLURM authentication credentials
   - `slurm-exporter-tls`: TLS certificates for ingress

### Customization

#### Resource Requirements

Edit `deployment.yaml` to adjust resource allocation:

```yaml
resources:
  requests:
    cpu: 200m      # Increase for larger clusters
    memory: 256Mi  # Increase for high cardinality
  limits:
    cpu: 1000m     # Max CPU burst
    memory: 1Gi    # Memory limit
```

#### Scaling Parameters

Edit `hpa.yaml` to adjust scaling behavior:

```yaml
spec:
  minReplicas: 3      # Minimum instances
  maxReplicas: 10     # Maximum instances
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 70  # Scale up threshold
```

#### Network Security

Edit `networkpolicy.yaml` to customize network access:

```yaml
# Allow specific monitoring systems
- from:
  - namespaceSelector:
      matchLabels:
        name: monitoring
  ports:
  - protocol: TCP
    port: 10341
```

## Monitoring Integration

### Prometheus Setup

The deployment includes ServiceMonitor for automatic Prometheus discovery:

```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    scrape_configs:
    - job_name: 'slurm-exporter'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names: [slurm-exporter]
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        action: keep
        regex: slurm-exporter
```

### Grafana Dashboards

Import the included dashboards:

```bash
# Import dashboards
kubectl apply -f ../../grafana/dashboards/
```

### Alerting Rules

Configure Alertmanager for notifications:

```yaml
# alertmanager-config.yaml
route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'slurm-exporter-alerts'

receivers:
- name: 'slurm-exporter-alerts'
  slack_configs:
  - api_url: 'YOUR_SLACK_WEBHOOK'
    channel: '#slurm-alerts'
    title: 'SLURM Exporter Alert'
```

## Security

### RBAC Configuration

The deployment uses minimal RBAC permissions:

- **ClusterRole**: Read access to nodes, services, endpoints
- **Role**: Namespace-specific access to configs and secrets
- **ServiceAccount**: Dedicated service account with token auto-mount

### Network Policies

Network access is restricted to:
- Prometheus monitoring systems
- Load balancer health checks
- SLURM API endpoints
- Kubernetes API server

### Security Contexts

All containers run with security hardening:
- Non-root user (UID 65534)
- Read-only root filesystem
- Dropped capabilities
- seccomp profile

### Pod Security Standards

The deployment is compatible with Pod Security Standards:
```yaml
# namespace.yaml
metadata:
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

## Troubleshooting

### Common Issues

1. **Pods not starting**
   ```bash
   # Check pod status
   kubectl describe pods -n slurm-exporter
   
   # Check events
   kubectl get events -n slurm-exporter --sort-by='.lastTimestamp'
   
   # Check logs
   kubectl logs -f deployment/slurm-exporter -n slurm-exporter
   ```

2. **SLURM connectivity issues**
   ```bash
   # Test connectivity from pod
   kubectl exec -it deployment/slurm-exporter -n slurm-exporter -- \
     curl -k https://slurm-head:6820/slurm/v0.0.44/ping
   
   # Check secret configuration
   kubectl get secret slurm-credentials -n slurm-exporter -o yaml
   ```

3. **Metrics not being scraped**
   ```bash
   # Check ServiceMonitor
   kubectl get servicemonitor -n slurm-exporter
   
   # Check Prometheus targets
   kubectl port-forward svc/prometheus 9090:9090 -n monitoring
   # Browse to http://localhost:9090/targets
   ```

### Debug Mode

Enable debug logging:
```bash
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "args": [
            "--config=/etc/slurm-exporter/config.yaml",
            "--log-level=debug"
          ]
        }]
      }
    }
  }
}'
```

### Performance Debugging

Enable pprof endpoint:
```bash
kubectl port-forward deployment/slurm-exporter 6060:6060 -n slurm-exporter

# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Maintenance

### Rolling Updates

Update the deployment image:
```bash
kubectl set image deployment/slurm-exporter \
  slurm-exporter=ghcr.io/jontk/slurm-exporter:v1.1.0 \
  -n slurm-exporter
```

### Configuration Updates

Update configuration without restart:
```bash
# Update ConfigMap
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "new-config-content"
  }
}'

# Restart deployment to pick up changes
kubectl rollout restart deployment/slurm-exporter -n slurm-exporter
```

### Scaling Operations

Manual scaling:
```bash
# Scale up
kubectl scale deployment slurm-exporter --replicas=5 -n slurm-exporter

# Scale down
kubectl scale deployment slurm-exporter --replicas=2 -n slurm-exporter
```

## Backup and Recovery

### Configuration Backup

```bash
# Backup all configurations
kubectl get all,configmaps,secrets,networkpolicies,servicemonitors,prometheusrules \
  -n slurm-exporter -o yaml > slurm-exporter-backup.yaml
```

### Disaster Recovery

```bash
# Restore from backup
kubectl apply -f slurm-exporter-backup.yaml

# Verify restoration
kubectl get pods -n slurm-exporter
kubectl get svc -n slurm-exporter
```

This high-availability deployment provides enterprise-grade reliability, security, and observability for production SLURM monitoring environments.