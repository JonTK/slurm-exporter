# Environment Configurations

This directory contains environment-specific configuration files for SLURM Exporter deployments across different environments.

## Overview

Environment configurations provide tailored settings optimized for specific deployment scenarios, from development to production, ensuring appropriate resource allocation, security levels, and operational characteristics.

## Available Environments

### Production (`production.yaml`)
- **Target**: High-availability production workloads
- **Security**: Maximum security hardening and compliance
- **Performance**: Optimized for large-scale SLURM clusters
- **Monitoring**: Comprehensive observability and alerting
- **Resources**: High resource limits with auto-scaling
- **Features**: Full feature set with production-grade reliability

### Staging (`staging.yaml`)
- **Target**: Pre-production testing and validation
- **Security**: Standard security with testing flexibility
- **Performance**: Moderate optimization for testing loads
- **Monitoring**: Enhanced debugging with tracing enabled
- **Resources**: Medium resource allocation
- **Features**: All features enabled for validation

### Development (`development.yaml`)
- **Target**: Local development and feature testing
- **Security**: Minimal security for development ease
- **Performance**: Basic optimization with debug features
- **Monitoring**: Verbose logging and full tracing
- **Resources**: Minimal resource requirements
- **Features**: All features plus experimental capabilities

## Configuration Structure

Each environment configuration includes:

```yaml
# Environment metadata
environment:
  name: "environment-name"
  cluster: "cluster-identifier"
  region: "deployment-region"
  compliance_level: "security-compliance-level"
  security_tier: "security-classification"

# Core service configurations
slurm: {}          # SLURM cluster connection settings
server: {}         # HTTP server configuration
logging: {}        # Logging and output settings
collectors: {}     # Metric collection configuration
performance: {}    # Performance optimization settings
health: {}         # Health check configuration
security: {}       # Security and access control
monitoring: {}     # Observability and metrics
features: {}       # Feature flags and capabilities
resources: {}      # Resource limits and requests
overrides: {}      # Environment-specific overrides
integrations: {}   # External system integrations
```

## Usage

### 1. Direct Configuration
```bash
# Use environment-specific config directly
./slurm-exporter --config=deployment/production/environments/production.yaml
```

### 2. Kubernetes ConfigMap
```bash
# Create ConfigMap from environment config
kubectl create configmap slurm-exporter-config \
  --from-file=config.yaml=deployment/production/environments/production.yaml \
  -n slurm-exporter
```

### 3. Helm Values Override
```bash
# Use with Helm chart
helm install slurm-exporter ./charts/slurm-exporter \
  --values deployment/production/environments/production.yaml
```

### 4. Docker Compose
```yaml
# docker-compose.yml
services:
  slurm-exporter:
    image: ghcr.io/jontk/slurm-exporter:latest
    volumes:
      - ./deployment/production/environments/production.yaml:/etc/config.yaml
    command: ["--config=/etc/config.yaml"]
```

## Environment Comparison

| Feature | Development | Staging | Production |
|---------|------------|---------|------------|
| Security Level | Minimal | Standard | Maximum |
| Resource Limits | 256Mi RAM | 512Mi RAM | 1Gi RAM |
| Logging Level | Debug | Debug | Info |
| TLS Enabled | No | No | Yes |
| Rate Limiting | Disabled | Enabled | Enabled |
| Circuit Breaker | Disabled | Enabled | Enabled |
| High Availability | No | No | Yes |
| Auto-scaling | No | Limited | Full |
| Compliance | None | Medium | High |
| Monitoring | Basic | Enhanced | Comprehensive |
| Backup | None | None | Enabled |

## Environment-Specific Features

### Production Features
- **High Availability**: Multi-replica deployment with anti-affinity
- **Security Hardening**: TLS, security contexts, network policies
- **Compliance**: Audit logging, data retention, encryption
- **Performance Optimization**: Connection pooling, caching, cardinality control
- **Disaster Recovery**: Backup, monitoring, incident response

### Staging Features  
- **Testing Focus**: All features enabled for validation
- **Debug Capabilities**: Enhanced logging, tracing, profiling
- **Moderate Security**: Balanced security for testing flexibility
- **Integration Testing**: Full integration with external systems

### Development Features
- **Developer Productivity**: Hot reload, verbose logging, debug endpoints
- **Experimental Features**: Early access to new capabilities
- **Minimal Security**: No authentication or rate limiting
- **Local Development**: Support for local development workflows

## Configuration Customization

### Environment Variables
All configurations support environment variable substitution:

```yaml
slurm:
  rest_url: "${SLURM_REST_URL}"
  jwt_token: "${SLURM_JWT_TOKEN}"

logging:
  level: "${LOG_LEVEL:-info}"

custom_labels:
  cluster_name: "${CLUSTER_NAME}"
  region: "${AWS_REGION}"
```

### Configuration Overrides
Override specific settings for your deployment:

```bash
# Override via environment variables
export SLURM_REST_URL="https://your-slurm-cluster:6820"
export LOG_LEVEL="debug"
export CLUSTER_NAME="your-cluster"

# Override via command-line flags
./slurm-exporter \
  --config=production.yaml \
  --log-level=debug \
  --slurm.rest-url="https://your-cluster:6820"
```

### Custom Environments
Create custom environment configurations by copying and modifying existing ones:

```bash
# Create custom environment
cp production.yaml custom.yaml

# Edit custom.yaml for your specific requirements
vim custom.yaml
```

## Security Considerations

### Production Security
- Enable TLS for all communications
- Use proper authentication and authorization
- Implement network policies and security contexts
- Enable audit logging and monitoring
- Regular security updates and vulnerability scanning

### Development Security
- Never use production credentials in development
- Use separate SLURM clusters for development/testing
- Implement proper secret management
- Regular security reviews of development practices

## Performance Tuning

### Resource Allocation
Adjust resource limits based on cluster size:

```yaml
# Small clusters (< 1000 nodes)
resources:
  requests: { cpu: "100m", memory: "128Mi" }
  limits: { cpu: "500m", memory: "512Mi" }

# Medium clusters (1000-5000 nodes)  
resources:
  requests: { cpu: "200m", memory: "256Mi" }
  limits: { cpu: "1000m", memory: "1Gi" }

# Large clusters (> 5000 nodes)
resources:
  requests: { cpu: "500m", memory: "512Mi" }
  limits: { cpu: "2000m", memory: "2Gi" }
```

### Collection Intervals
Optimize collection frequencies based on workload:

```yaml
# High-frequency monitoring
collectors:
  jobs: { collection_interval: "15s" }
  nodes: { collection_interval: "30s" }

# Standard monitoring  
collectors:
  jobs: { collection_interval: "30s" }
  nodes: { collection_interval: "60s" }

# Low-frequency monitoring
collectors:
  jobs: { collection_interval: "60s" }
  nodes: { collection_interval: "120s" }
```

## Monitoring and Alerting

### Environment-Specific Alerts
Each environment has different alerting thresholds:

| Metric | Development | Staging | Production |
|--------|------------|---------|------------|
| Error Rate | 50% | 10% | 1% |
| Response Time | N/A | 10s | 5s |
| Memory Usage | 95% | 85% | 80% |
| Availability | N/A | 95% | 99.9% |

### Dashboard Configuration
Import environment-specific Grafana dashboards:

```bash
# Production dashboard with SLA tracking
kubectl apply -f monitoring/grafana-production-dashboard.json

# Development dashboard with debug info
kubectl apply -f monitoring/grafana-development-dashboard.json
```

## Migration Between Environments

### Development to Staging
1. Update resource limits and security settings
2. Enable authentication and rate limiting  
3. Add health checks and monitoring
4. Test integration with external systems

### Staging to Production
1. Enable TLS and security hardening
2. Configure high availability and auto-scaling
3. Enable compliance and audit logging
4. Set up backup and disaster recovery
5. Configure production monitoring and alerting

## Troubleshooting

### Common Issues

#### Configuration Validation Errors
```bash
# Validate configuration syntax
./slurm-exporter --config=production.yaml --validate

# Check for missing environment variables
env | grep SLURM
```

#### Environment Variable Substitution
```bash
# Test variable substitution
envsubst < production.yaml > resolved-config.yaml
./slurm-exporter --config=resolved-config.yaml --dry-run
```

#### Resource Constraints
```bash
# Check resource usage
kubectl top pods -n slurm-exporter
kubectl describe pod -n slurm-exporter

# Adjust limits if needed
kubectl patch deployment slurm-exporter -p '{"spec":{"template":{"spec":{"containers":[{"name":"slurm-exporter","resources":{"limits":{"memory":"2Gi"}}}]}}}}'
```

## Best Practices

### Configuration Management
- Store configurations in version control
- Use environment-specific branches or tags
- Implement configuration validation in CI/CD
- Document all environment-specific changes

### Security
- Use separate credentials for each environment
- Implement least-privilege access
- Regular security reviews and updates
- Monitor for configuration drift

### Operations
- Automate configuration deployment
- Test configurations in lower environments first
- Monitor configuration changes
- Maintain environment parity where possible

This environment configuration system provides a robust foundation for deploying SLURM Exporter across different environments while maintaining consistency, security, and operational excellence.