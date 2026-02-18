# Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the SLURM Prometheus Exporter.

## Quick Deployment

### Using kubectl
```bash
# Apply all manifests
kubectl apply -f k8s/

# Or use kustomize
kubectl apply -k k8s/
```

### Using kustomize
```bash
# Build and view the configuration
kustomize build k8s/

# Apply directly
kustomize build k8s/ | kubectl apply -f -
```

## Components

### Core Components
- **Namespace** (`namespace.yaml`) - Dedicated namespace for monitoring components
- **ServiceAccount & RBAC** (`rbac.yaml`) - Service account with minimal required permissions
- **Deployment** (`deployment.yaml`) - Main application deployment with resource limits and health checks
- **Service** (`service.yaml`) - Service to expose the exporter
- **ConfigMap** (`configmap.yaml`) - Configuration management

### Security & Policy
- **NetworkPolicy** (`networkpolicy.yaml`) - Network security policies
- **PodDisruptionBudget** (`poddisruptionbudget.yaml`) - Availability guarantees during disruptions

### Scaling & Performance
- **HorizontalPodAutoscaler** (`hpa.yaml`) - Automatic scaling based on CPU/memory usage

### Management
- **Kustomization** (`kustomization.yaml`) - Kustomize configuration for easy customization

## Configuration

### Environment-specific Customization

Create overlays for different environments:

```bash
# Create environment-specific directories
mkdir -p k8s/overlays/{dev,staging,prod}
```

Example overlay for production (`k8s/overlays/prod/kustomization.yaml`):
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: slurm-monitoring

resources:
- ../../base

patchesStrategicMerge:
- deployment-prod.yaml

images:
- name: slurm-exporter
  newTag: v1.0.0

replicas:
- name: slurm-exporter
  count: 2
```

### Configuration Management

1. **ConfigMap Generation**: The kustomization automatically generates a ConfigMap from `../configs/config.yaml`

2. **Manual ConfigMap**: Alternatively, create your own ConfigMap:
   ```bash
   kubectl create configmap slurm-exporter-config \
     --from-file=config.yaml=configs/config.yaml \
     -n monitoring
   ```

3. **Secret for TLS** (optional):
   ```bash
   kubectl create secret tls slurm-exporter-tls \
     --cert=path/to/cert.pem \
     --key=path/to/key.pem \
     -n monitoring
   ```

## Resource Requirements

### Default Resource Limits
- **CPU**: 100m (request) / 500m (limit)
- **Memory**: 128Mi (request) / 512Mi (limit)

### Scaling Configuration
- **Min Replicas**: 1
- **Max Replicas**: 3
- **CPU Threshold**: 70%
- **Memory Threshold**: 80%

## Security Features

### Pod Security
- Runs as non-root user (UID 65534)
- Read-only root filesystem
- No privilege escalation
- Drops all capabilities
- Security context constraints

### Network Security
- NetworkPolicy restricts ingress/egress traffic
- Only allows necessary communication:
  - Prometheus scraping on port 10341
  - DNS resolution
  - SLURM API access
  - Kubernetes API (if needed)

### RBAC
- Minimal permissions
- ServiceAccount with least privilege access
- Only necessary cluster role bindings

## Health Checks

### Probe Configuration
- **Startup Probe**: Initial health check with extended timeout
- **Liveness Probe**: Continuous health monitoring
- **Readiness Probe**: Service readiness validation

### Endpoints
- **Health**: `/health` - Basic application health
- **Readiness**: `/ready` - Service readiness including collector status

## Monitoring Integration

### Prometheus Discovery
The deployment includes annotations for automatic Prometheus discovery:
```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "10341"
  prometheus.io/path: "/metrics"
```

### ServiceMonitor (Prometheus Operator)
If using Prometheus Operator, create a ServiceMonitor:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: slurm-exporter
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: slurm-exporter
  endpoints:
  - port: http-metrics
    interval: 30s
    path: /metrics
```

## Troubleshooting

### Common Issues

1. **Pod not starting**:
   ```bash
   kubectl describe pod -l app=slurm-exporter -n monitoring
   kubectl logs -l app=slurm-exporter -n monitoring
   ```

2. **Configuration issues**:
   ```bash
   kubectl get configmap slurm-exporter-config -n monitoring -o yaml
   ```

3. **Network connectivity**:
   ```bash
   kubectl exec -it deployment/slurm-exporter -n monitoring -- /bin/sh
   # Test SLURM API connectivity from inside the pod
   ```

4. **RBAC permissions**:
   ```bash
   kubectl auth can-i --list --as=system:serviceaccount:monitoring:slurm-exporter
   ```

### Debugging Commands

```bash
# View all resources
kubectl get all -n monitoring -l app=slurm-exporter

# Check pod logs
kubectl logs -f deployment/slurm-exporter -n monitoring

# Exec into pod
kubectl exec -it deployment/slurm-exporter -n monitoring -- /bin/sh

# View events
kubectl get events -n monitoring --sort-by='.lastTimestamp'

# Check network policies
kubectl describe networkpolicy slurm-exporter -n monitoring
```

## Cleanup

```bash
# Remove all resources
kubectl delete -k k8s/

# Or remove specific components
kubectl delete -f k8s/
```