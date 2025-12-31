# Production Deployment Guide

This directory contains production-ready deployment configurations and operational guidance for the SLURM Exporter.

## Deployment Strategies

### 1. High Availability Kubernetes Deployment
- **File**: `kubernetes-ha/`
- **Use Case**: Large-scale production environments
- **Features**: Multiple replicas, load balancing, auto-scaling
- **Requirements**: Kubernetes 1.19+, 3+ nodes

### 2. Docker Swarm Deployment
- **File**: `docker-swarm/`
- **Use Case**: Docker-native environments
- **Features**: Service discovery, rolling updates, secrets management
- **Requirements**: Docker Swarm mode

### 3. Standalone Production Deployment
- **File**: `standalone/`
- **Use Case**: Bare metal or VM deployments
- **Features**: systemd integration, log rotation, monitoring
- **Requirements**: Linux system with systemd

### 4. Cloud-Native Deployment
- **File**: `cloud-native/`
- **Use Case**: AWS/Azure/GCP environments
- **Features**: Cloud-specific integrations, managed services
- **Requirements**: Cloud provider account and CLI tools

## Quick Start

### Prerequisites
- SLURM cluster with REST API enabled
- Prometheus monitoring stack
- Network access to SLURM head node
- TLS certificates (recommended)

### Basic Production Setup

1. **Choose deployment method**:
   ```bash
   # Kubernetes (recommended)
   cd kubernetes-ha
   kubectl apply -f .
   
   # Docker Swarm
   cd docker-swarm
   docker stack deploy -c docker-compose.yml slurm-exporter
   
   # Standalone
   cd standalone
   sudo ./install.sh
   ```

2. **Configure monitoring**:
   ```bash
   # Add to Prometheus configuration
   kubectl apply -f monitoring/prometheus-config.yaml
   
   # Deploy Grafana dashboards
   kubectl apply -f monitoring/grafana-dashboards.yaml
   
   # Set up alerting rules
   kubectl apply -f monitoring/alert-rules.yaml
   ```

3. **Verify deployment**:
   ```bash
   # Check health
   curl https://slurm-exporter.example.com/health
   
   # Check metrics
   curl https://slurm-exporter.example.com/metrics
   
   # Verify in Prometheus
   kubectl port-forward svc/prometheus 9090:9090
   # Browse to http://localhost:9090
   ```

## Security Considerations

### Network Security
- Deploy behind a reverse proxy (nginx, HAProxy)
- Use TLS termination at load balancer
- Implement network policies in Kubernetes
- Restrict access to SLURM API endpoints

### Authentication & Authorization
- Enable authentication for metrics endpoint
- Use service accounts in Kubernetes
- Implement RBAC for cluster access
- Rotate credentials regularly

### Container Security
- Use distroless or minimal base images
- Run as non-root user
- Set resource limits and requests
- Enable security contexts and policies

## Monitoring & Observability

### Key Metrics to Monitor
- **Exporter Health**: Up/down status, response time
- **Collection Performance**: Scrape duration, error rate
- **Resource Usage**: CPU, memory, network
- **SLURM Connectivity**: API response time, error rate

### Alerting Rules
- Exporter down for > 5 minutes
- High error rate (> 5% over 10 minutes)
- High collection latency (> 30s)
- Memory usage > 80%

### Log Management
- Structured JSON logging
- Centralized log aggregation (ELK, Loki)
- Log retention policies
- Error log alerting

## Performance Tuning

### Scaling Guidelines
- **Small clusters** (< 1000 nodes): Single instance
- **Medium clusters** (1000-10000 nodes): 2-3 instances with load balancing
- **Large clusters** (> 10000 nodes): Horizontal scaling with sharding

### Optimization Settings
```yaml
# High-performance configuration
collectors:
  jobs:
    collection_interval: 30s
    timeout: 15s
    cardinality_limit: 10000
  
  nodes:
    collection_interval: 60s
    timeout: 30s
    enable_detailed_metrics: false

caching:
  enabled: true
  ttl: 30s
  max_size: 1000

performance:
  gc_percent: 100
  max_memory: "512Mi"
  connection_pool_size: 10
```

## Backup & Recovery

### Configuration Backup
- Store configurations in version control
- Backup Kubernetes manifests
- Document environment-specific settings

### Disaster Recovery
- Multi-region deployment options
- Database backup procedures (if applicable)
- Recovery time objectives (RTO < 30 minutes)

## Maintenance Procedures

### Regular Maintenance
- Weekly security updates
- Monthly performance reviews
- Quarterly capacity planning
- Annual security audits

### Update Procedures
- Blue-green deployments
- Canary releases for major updates
- Rollback procedures
- Change management process

## Troubleshooting

### Common Issues
1. **Connection timeouts**: Check network connectivity to SLURM API
2. **High memory usage**: Review cardinality limits and caching settings
3. **Missing metrics**: Verify collector configuration and SLURM permissions
4. **Authentication failures**: Check credentials and token expiration

### Debugging Tools
```bash
# Check exporter status
kubectl logs deployment/slurm-exporter

# Test SLURM connectivity
kubectl exec deployment/slurm-exporter -- curl http://slurm-api:6820/slurm/v0.0.39/ping

# Performance profiling
kubectl port-forward deployment/slurm-exporter 6060:6060
go tool pprof http://localhost:6060/debug/pprof/profile
```

## Support & Escalation

### Support Tiers
1. **Level 1**: Basic troubleshooting, configuration issues
2. **Level 2**: Performance tuning, advanced configuration
3. **Level 3**: Core development team, critical issues

### Escalation Procedures
- **Severity 1**: Immediate escalation, 24/7 response
- **Severity 2**: Business hours escalation, 4-hour response
- **Severity 3**: Standard support, 24-hour response

### Contact Information
- **Operations Team**: ops@example.com
- **On-call Engineer**: +1-555-ONCALL
- **Documentation**: https://docs.example.com/slurm-exporter
- **Issue Tracking**: https://github.com/jontk/slurm-exporter/issues