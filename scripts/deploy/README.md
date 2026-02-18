# SLURM Exporter Deployment Scripts

This directory contains deployment scripts for managing the SLURM Exporter across different environments.

## Scripts Overview

### Core Deployment Scripts

- **`deploy.sh`** - Deploy SLURM Exporter to Kubernetes using Helm
- **`undeploy.sh`** - Remove SLURM Exporter from Kubernetes 
- **`upgrade.sh`** - Upgrade existing deployment to new version
- **`rollback.sh`** - Rollback deployment to previous version
- **`build-and-deploy.sh`** - Build Docker image and deploy in one step
- **`status.sh`** - Check deployment status and health

## Quick Start

### Deploy to Development
```bash
./deploy.sh -e development -u "http://slurm-dev:6820"
```

### Deploy to Production
```bash
./deploy.sh -e production -t "v1.0.0" -u "https://slurm-prod:6820"
```

### Check Status
```bash
./status.sh -e production --logs
```

### Upgrade Deployment
```bash
./upgrade.sh -e production -t "v1.1.0"
```

## Environment Configuration

The scripts support three environments with different defaults:

### Development
- Namespace: `monitoring`
- Values file: `values-development.yaml`
- Timeout: 300s
- No confirmation required

### Staging  
- Namespace: `monitoring`
- Values file: `values.yaml`
- Timeout: 300s
- No confirmation required

### Production
- Namespace: `monitoring`  
- Values file: `values-production.yaml`
- Timeout: 600s
- Confirmation required for destructive operations
- Waits for rollout completion

## Script Details

### deploy.sh

Deploy the SLURM Exporter to Kubernetes using Helm.

**Usage:**
```bash
./deploy.sh [OPTIONS]
```

**Key Options:**
- `-e, --environment` - Target environment (development|staging|production)
- `-u, --slurm-url` - SLURM REST API base URL
- `-t, --image-tag` - Docker image tag
- `-f, --values-file` - Custom Helm values file
- `-k, --context` - Kubectl context
- `--dry-run` - Preview changes without applying

**Examples:**
```bash
# Development deployment
./deploy.sh -e development -u "http://slurm-dev:6820"

# Production with custom image
./deploy.sh -e production -t "v1.0.0" -u "https://slurm-prod:6820"

# Custom values file
./deploy.sh -f custom-values.yaml -k prod-cluster

# Dry run
./deploy.sh -e staging --dry-run
```

### undeploy.sh

Remove the SLURM Exporter from Kubernetes.

**Usage:**
```bash
./undeploy.sh [OPTIONS]
```

**Key Options:**
- `-e, --environment` - Target environment
- `--delete-namespace` - Also delete the namespace
- `--force` - Skip confirmation prompts
- `--dry-run` - Preview changes

**Examples:**
```bash
# Remove from development
./undeploy.sh -e development

# Remove and delete namespace
./undeploy.sh -e staging --delete-namespace

# Force removal without confirmation
./undeploy.sh -e production --force
```

### upgrade.sh

Upgrade an existing SLURM Exporter deployment.

**Usage:**
```bash
./upgrade.sh [OPTIONS]
```

**Key Options:**
- `-e, --environment` - Target environment
- `-t, --image-tag` - New image tag
- `--no-wait` - Don't wait for rollout completion

**Examples:**
```bash
# Upgrade to new version
./upgrade.sh -e production -t "v1.2.0"

# Upgrade with custom values
./upgrade.sh -f custom-values.yaml -t "latest"
```

### rollback.sh

Rollback deployment to a previous version.

**Usage:**
```bash
./rollback.sh [OPTIONS]
```

**Key Options:**
- `-e, --environment` - Target environment
- `--revision` - Specific revision to rollback to
- `--no-wait` - Don't wait for rollout completion

**Examples:**
```bash
# Rollback to previous version
./rollback.sh -e production

# Rollback to specific revision
./rollback.sh -e staging --revision 3
```

### build-and-deploy.sh

Build Docker image and deploy in a single operation.

**Usage:**
```bash
./build-and-deploy.sh [OPTIONS]
```

**Key Options:**
- `-e, --environment` - Target environment
- `-r, --registry` - Docker registry
- `-t, --tag` - Image tag (defaults to git commit hash)
- `--skip-build` - Skip Docker build
- `--skip-deploy` - Skip deployment
- `--no-push` - Don't push to registry

**Examples:**
```bash
# Build and deploy to development
./build-and-deploy.sh -e development

# Build with registry and deploy to production
./build-and-deploy.sh -e production -r myregistry.com -t v1.0.0

# Only build, don't deploy
./build-and-deploy.sh --skip-deploy -t latest
```

### status.sh

Check the status of a deployed SLURM Exporter.

**Usage:**
```bash
./status.sh [OPTIONS]
```

**Key Options:**
- `-e, --environment` - Target environment
- `--logs` - Show recent logs
- `--tail` - Number of log lines to show

**Examples:**
```bash
# Check status
./status.sh -e production

# Check with logs
./status.sh -e development --logs --tail 100
```

## Prerequisites

All scripts require:
- `kubectl` - Kubernetes CLI
- `helm` - Helm package manager
- `docker` - Docker CLI (for build-and-deploy.sh)
- `git` - Git CLI (for build-and-deploy.sh tag generation)

## Environment Variables

Scripts support environment variables for common settings:

```bash
export ENVIRONMENT="production"
export NAMESPACE="monitoring"
export HELM_RELEASE_NAME="slurm-exporter"
export DOCKER_REGISTRY="myregistry.com"
export SLURM_BASE_URL="https://slurm-prod:6820"
```

## File Structure

```
scripts/deploy/
├── deploy.sh           # Main deployment script
├── undeploy.sh         # Removal script
├── upgrade.sh          # Upgrade script  
├── rollback.sh         # Rollback script
├── build-and-deploy.sh # Build and deploy script
├── status.sh           # Status checking script
└── README.md           # This documentation
```

## Security Considerations

### Production Deployments
- Always use specific image tags (not `latest`)
- Verify SLURM endpoints are accessible
- Use proper authentication credentials
- Review changes with `--dry-run` first
- Production scripts require explicit confirmation

### Authentication
- Scripts inherit kubectl authentication
- Ensure proper RBAC permissions
- Store sensitive values in Kubernetes secrets
- Use secure registries for production images

## Troubleshooting

### Common Issues

**Deployment fails with timeout:**
```bash
# Check pod status
kubectl get pods -n monitoring -l app.kubernetes.io/instance=slurm-exporter

# Check logs
kubectl logs -n monitoring -l app.kubernetes.io/instance=slurm-exporter
```

**Image pull errors:**
```bash
# Verify image exists
docker pull myregistry.com/slurm-exporter:v1.0.0

# Check registry credentials
kubectl get secrets -n monitoring
```

**SLURM connection issues:**
```bash
# Test SLURM API connectivity
kubectl exec -n monitoring deployment/slurm-exporter -- wget -q -O - http://slurm-server:6820/slurm/v0.0.40/ping

# Check configuration
kubectl get configmap -n monitoring slurm-exporter-config -o yaml
```

### Debugging Commands

```bash
# Get detailed deployment info
kubectl describe deployment -n monitoring slurm-exporter

# Check events
kubectl get events -n monitoring --sort-by='.lastTimestamp'

# Port forward for local testing
kubectl port-forward -n monitoring service/slurm-exporter 10341:10341
curl http://localhost:10341/health
curl http://localhost:10341/metrics
```

## Integration with CI/CD

### GitHub Actions Example
```yaml
- name: Deploy to Production
  run: |
    ./scripts/deploy/build-and-deploy.sh \
      -e production \
      -r ${{ env.REGISTRY }} \
      -t ${{ github.sha }} \
      -u ${{ secrets.SLURM_URL }}
```

### GitLab CI Example
```yaml
deploy:production:
  script:
    - ./scripts/deploy/build-and-deploy.sh -e production -r $REGISTRY -t $CI_COMMIT_SHA
  when: manual
  only:
    - main
```

For more information, see the main project documentation.