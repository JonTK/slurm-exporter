# SLURM Exporter Helm Chart

This Helm chart deploys the SLURM Prometheus Exporter on a Kubernetes cluster using Helm package manager.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Access to a SLURM cluster with REST API enabled

## Installation

### Install from Local Chart

```bash
# Clone the repository
git clone https://github.com/jontk/slurm-exporter.git
cd slurm-exporter

# Install the chart
helm install slurm-exporter ./charts/slurm-exporter
```

### Install with Custom Values

```bash
# Development environment
helm install slurm-exporter ./charts/slurm-exporter \
  -f charts/slurm-exporter/values-development.yaml

# Production environment
helm install slurm-exporter ./charts/slurm-exporter \
  -f charts/slurm-exporter/values-production.yaml \
  --set config.slurm.baseURL="https://your-slurm-cluster:6820" \
  --set secrets.slurmAuth.data.jwt-token="$(echo -n 'your-jwt-token' | base64)"
```

## Configuration

The following table lists the configurable parameters and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.imageRegistry` | Global Docker image registry | `""` |
| `global.imagePullSecrets` | Global Docker registry secret names | `[]` |

### Image Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.registry` | Image registry | `""` |
| `image.repository` | Image repository | `slurm-exporter` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Deployment Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `updateStrategy` | Deployment update strategy | `RollingUpdate` |

### Service Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `service.annotations` | Service annotations | `{}` |

### Monitoring Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceMonitor.enabled` | Enable ServiceMonitor for Prometheus Operator | `false` |
| `serviceMonitor.interval` | Scrape interval | `30s` |
| `prometheusRule.enabled` | Enable PrometheusRule for alerting | `false` |

### Security Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext.runAsUser` | User ID to run the container | `65534` |
| `podSecurityContext.runAsGroup` | Group ID to run the container | `65534` |
| `securityContext.readOnlyRootFilesystem` | Mount root filesystem as read-only | `true` |

### SLURM Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.slurm.baseURL` | SLURM REST API base URL | `https://slurm-api.example.com:6820` |
| `config.slurm.apiVersion` | SLURM API version | `v0.0.42` |
| `config.slurm.auth.type` | Authentication type | `none` |

### Resource Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |

## Examples

### Basic Installation

```bash
helm install my-slurm-exporter ./charts/slurm-exporter \
  --set config.slurm.baseURL="https://slurm.example.com:6820"
```

### Installation with TLS and Authentication

```bash
# Create secrets first
kubectl create secret tls slurm-exporter-tls \
  --cert=path/to/cert.pem \
  --key=path/to/key.pem

kubectl create secret generic slurm-exporter-auth \
  --from-literal=jwt-token="your-jwt-token"

# Install with TLS and auth
helm install my-slurm-exporter ./charts/slurm-exporter \
  --set config.server.tls.enabled=true \
  --set config.slurm.auth.type=jwt \
  --set config.slurm.auth.tokenFile="/etc/slurm-exporter/secrets/jwt-token" \
  --set secrets.tls.name="slurm-exporter-tls" \
  --set secrets.slurmAuth.name="slurm-exporter-auth"
```

### Installation with Prometheus Operator

```bash
helm install my-slurm-exporter ./charts/slurm-exporter \
  --set serviceMonitor.enabled=true \
  --set serviceMonitor.labels.prometheus=kube-prometheus \
  --set prometheusRule.enabled=true
```

### Installation with Autoscaling

```bash
helm install my-slurm-exporter ./charts/slurm-exporter \
  --set autoscaling.enabled=true \
  --set autoscaling.minReplicas=2 \
  --set autoscaling.maxReplicas=5 \
  --set autoscaling.targetCPUUtilizationPercentage=70
```

## Upgrading

```bash
# Upgrade to a new version
helm upgrade my-slurm-exporter ./charts/slurm-exporter

# Upgrade with new values
helm upgrade my-slurm-exporter ./charts/slurm-exporter \
  -f new-values.yaml
```

## Uninstalling

```bash
helm uninstall my-slurm-exporter
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/name=slurm-exporter
kubectl logs -l app.kubernetes.io/name=slurm-exporter
```

### Check Configuration

```bash
kubectl get configmap slurm-exporter -o yaml
```

### Check Service and Endpoints

```bash
kubectl get service slurm-exporter
kubectl get endpoints slurm-exporter
```

### Check ServiceMonitor (if using Prometheus Operator)

```bash
kubectl get servicemonitor slurm-exporter
```

### Test Metrics Endpoint

```bash
kubectl port-forward service/slurm-exporter 8080:8080
curl http://localhost:8080/metrics
```

## Values Files

The chart includes example values files for different environments:

- `values.yaml` - Default values
- `values-development.yaml` - Development environment
- `values-production.yaml` - Production environment

Copy and modify these files according to your requirements.

## Contributing

Please read the main project [CONTRIBUTING.md](../../CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](../../LICENSE) file for details.