# SLURM Exporter Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the SLURM Prometheus Exporter in a Kubernetes cluster.

## Overview

The deployment includes:
- **Namespace**: Isolated environment for SLURM monitoring
- **Deployment**: Main exporter application with health checks
- **Service**: ClusterIP service for metrics exposure
- **ServiceMonitor**: Prometheus Operator integration
- **ConfigMap**: Exporter configuration
- **HPA**: Horizontal Pod Autoscaler for scaling
- **PDB**: Pod Disruption Budget for availability
- **NetworkPolicy**: Security policies for network access

## Prerequisites

- Kubernetes 1.21+
- kubectl configured to access your cluster
- SLURM REST API accessible from the cluster
- (Optional) Prometheus Operator for ServiceMonitor
- (Optional) Metrics Server for HPA

## Quick Start

### 1. Basic Deployment

```bash
# Create namespace and deploy all resources
kubectl apply -k kubernetes/

# Or manually apply individual resources
kubectl apply -f kubernetes/namespace.yaml
kubectl apply -f kubernetes/serviceaccount.yaml
kubectl apply -f kubernetes/configmap.yaml
kubectl apply -f kubernetes/deployment.yaml
kubectl apply -f kubernetes/service.yaml
```

### 2. Configure Authentication

Choose one of the following authentication methods:

#### JWT Token (Recommended)
```bash
# Create secret with JWT token
kubectl create secret generic slurm-jwt-token \
  --from-file=token=/path/to/jwt/token \
  -n slurm-monitoring
```

#### Basic Authentication
```bash
# Create secret with username/password
kubectl create secret generic slurm-exporter-auth \
  --from-literal=username=your-username \
  --from-literal=password=your-password \
  -n slurm-monitoring
```

#### API Token
```bash
# Create secret with API token
kubectl create secret generic slurm-exporter-auth \
  --from-literal=token=your-api-token \
  -n slurm-monitoring
```

### 3. Update Configuration

Edit the ConfigMap to match your environment:

```bash
# Edit the configmap
kubectl edit configmap slurm-exporter-config -n slurm-monitoring

# Or create a custom config file and replace
kubectl create configmap slurm-exporter-config \
  --from-file=config.yaml=my-config.yaml \
  -n slurm-monitoring \
  --dry-run=client -o yaml | kubectl apply -f -
```

### 4. Verify Deployment

```bash
# Check pod status
kubectl get pods -n slurm-monitoring

# Check logs
kubectl logs -n slurm-monitoring -l app.kubernetes.io/name=slurm-exporter

# Test metrics endpoint
kubectl port-forward -n slurm-monitoring svc/slurm-exporter 10341:10341
curl http://localhost:10341/metrics
```

## Environment-Specific Deployments

### Using Kustomize Overlays

Deploy to different environments using overlays:

```bash
# Development
kubectl apply -k kubernetes/overlays/development/

# Production
kubectl apply -k kubernetes/overlays/production/
```

### Creating Custom Overlays

1. Create a new overlay directory:
```bash
mkdir -p kubernetes/overlays/staging
```

2. Create kustomization.yaml:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
  - ../../

namePrefix: staging-

commonLabels:
  environment: staging

configMapGenerator:
  - name: slurm-exporter-config
    behavior: merge
    literals:
      - ENVIRONMENT=staging
      - CLUSTER_NAME=staging-cluster
```

## Configuration

### Environment Variables

Key environment variables in the deployment:

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | info |
| `CLUSTER_NAME` | SLURM cluster name for labels | production |
| `ENVIRONMENT` | Environment label | production |
| `SLURM_BASE_URL` | SLURM REST API URL | - |
| `SLURM_AUTH_TYPE` | Authentication type (jwt, basic, token) | jwt |
| `SLURM_JWT_PATH` | Path to JWT token file | /var/run/secrets/slurm-jwt/token |

### Resource Limits

Default resource allocations:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

Adjust based on your cluster size and metric volume.

### Scaling

The HPA is configured to scale based on:
- CPU utilization > 80%
- Memory utilization > 80%

Modify HPA for custom metrics:
```bash
kubectl edit hpa slurm-exporter -n slurm-monitoring
```

## Security Considerations

### Network Policies

The included NetworkPolicy:
- Allows ingress from Prometheus only
- Allows egress to SLURM API and DNS
- Blocks all other traffic

Adjust for your security requirements:
```bash
kubectl edit networkpolicy slurm-exporter -n slurm-monitoring
```

### Pod Security

The deployment runs with:
- Non-root user (65534)
- Read-only root filesystem
- No privilege escalation
- All capabilities dropped

### Secret Management

**Never commit secrets to version control!**

Consider using:
- Kubernetes Secrets with encryption at rest
- External secret managers (Vault, AWS Secrets Manager)
- Sealed Secrets for GitOps

Example with Sealed Secrets:
```bash
# Install sealed-secrets controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.18.0/controller.yaml

# Create sealed secret
echo -n mypassword | kubectl create secret generic slurm-auth \
  --dry-run=client \
  --from-file=password=/dev/stdin \
  -o yaml | kubeseal -o yaml > slurm-auth-sealed.yaml
```

## Monitoring

### Prometheus Configuration

If not using Prometheus Operator, add scrape config:

```yaml
scrape_configs:
  - job_name: 'slurm-exporter'
    kubernetes_sd_configs:
      - role: service
        namespaces:
          names:
            - slurm-monitoring
    relabel_configs:
      - source_labels: [__meta_kubernetes_service_name]
        regex: slurm-exporter
        action: keep
      - source_labels: [__meta_kubernetes_namespace]
        target_label: namespace
      - source_labels: [__meta_kubernetes_service_name]
        target_label: service
```

### Grafana Dashboards

Import dashboards from `../dashboards/`:
1. Configure Prometheus data source
2. Import dashboard JSON files
3. Select appropriate variables

## Troubleshooting

### Pod Won't Start

1. Check pod status:
```bash
kubectl describe pod -n slurm-monitoring -l app.kubernetes.io/name=slurm-exporter
```

2. Common issues:
   - Missing secrets
   - Incorrect SLURM API URL
   - Network connectivity
   - Invalid configuration

### No Metrics

1. Check exporter logs:
```bash
kubectl logs -n slurm-monitoring -l app.kubernetes.io/name=slurm-exporter
```

2. Verify connectivity:
```bash
# Test from within pod
kubectl exec -n slurm-monitoring -it deployment/slurm-exporter -- wget -O- http://localhost:10341/metrics
```

3. Check SLURM API access:
```bash
# Test API connectivity
kubectl exec -n slurm-monitoring -it deployment/slurm-exporter -- wget -O- $SLURM_BASE_URL/slurm/v0.0.40/ping
```

### High Memory Usage

1. Check cardinality:
```bash
curl http://localhost:10341/metrics | grep "^slurm_" | cut -d'{' -f1 | sort | uniq -c | sort -nr
```

2. Enable filtering in ConfigMap
3. Adjust cardinality limits
4. Increase resource limits if needed

## Maintenance

### Updating the Exporter

1. Build new image:
```bash
docker build -t your-registry/slurm-exporter:v1.1.0 .
docker push your-registry/slurm-exporter:v1.1.0
```

2. Update deployment:
```bash
kubectl set image deployment/slurm-exporter \
  slurm-exporter=your-registry/slurm-exporter:v1.1.0 \
  -n slurm-monitoring
```

3. Monitor rollout:
```bash
kubectl rollout status deployment/slurm-exporter -n slurm-monitoring
```

### Configuration Updates

The ConfigMap supports hot-reload:
```bash
# Edit config
kubectl edit configmap slurm-exporter-config -n slurm-monitoring

# Config will be reloaded automatically (may take up to 2 minutes)
```

### Backup and Restore

Backup configuration:
```bash
# Backup all resources
kubectl get all,cm,secret,sa -n slurm-monitoring -o yaml > slurm-exporter-backup.yaml

# Restore
kubectl apply -f slurm-exporter-backup.yaml
```

## Advanced Topics

### Multi-Cluster Monitoring

Deploy multiple instances with different configurations:

```yaml
# In each overlay, set unique cluster name
configMapGenerator:
  - name: slurm-exporter-config
    behavior: merge
    literals:
      - CLUSTER_NAME=cluster-1
```

### High Availability

For HA deployment:
1. Increase replicas in production
2. Use PodDisruptionBudget
3. Spread across availability zones
4. Consider using separate Prometheus instances

### Custom Metrics

Add custom metrics by:
1. Extending the exporter code
2. Using recording rules in Prometheus
3. Creating custom collectors

## Support

For issues:
1. Check exporter logs
2. Verify configuration
3. Test connectivity
4. Review security policies
5. Open an issue with:
   - Kubernetes version
   - Exporter version
   - Configuration (sanitized)
   - Error logs