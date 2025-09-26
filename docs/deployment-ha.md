# High Availability Deployment Patterns for SLURM Exporter

This document outlines deployment patterns and best practices for achieving high availability with the SLURM Exporter in production environments.

## Table of Contents

- [Overview](#overview)
- [Architecture Patterns](#architecture-patterns)
- [Kubernetes Deployment Strategies](#kubernetes-deployment-strategies)
- [Load Balancing and Service Discovery](#load-balancing-and-service-discovery)
- [Monitoring and Health Checks](#monitoring-and-health-checks)
- [Disaster Recovery](#disaster-recovery)
- [Performance and Scaling](#performance-and-scaling)
- [Security Considerations](#security-considerations)
- [Operational Procedures](#operational-procedures)

## Overview

High availability for SLURM Exporter ensures continuous monitoring of SLURM clusters without single points of failure. This is critical for maintaining observability in production HPC environments.

### Key Requirements
- **99.9% uptime** (8.76 hours downtime per year)
- **Automatic failover** within 30 seconds
- **Zero data loss** during failover
- **Graceful handling** of SLURM API outages
- **Horizontal scaling** for large clusters

### Failure Scenarios
- Pod crashes or restarts
- Node failures or maintenance
- Network partitions
- SLURM API unavailability
- Kubernetes cluster upgrades
- Data center outages

## Architecture Patterns

### Pattern 1: Active-Active with Load Balancing

```
┌─────────────────┐    ┌─────────────────┐
│   Prometheus    │    │   Prometheus    │
│   Instance 1    │    │   Instance 2    │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          └──────────┬───────────┘
                     │
        ┌─────────────────────────┐
        │    Load Balancer        │
        └─────────┬───────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
┌───▼───┐    ┌───▼───┐    ┌───▼───┐
│Exporter│    │Exporter│    │Exporter│
│Pod 1   │    │Pod 2   │    │Pod 3   │
└────────┘    └────────┘    └────────┘
    │             │             │
    └─────────────┼─────────────┘
                  │
        ┌─────────▼─────────┐
        │   SLURM Cluster   │
        └───────────────────┘
```

**Benefits:**
- No single point of failure
- Automatic load distribution
- Horizontal scaling capability
- Real-time failover

**Implementation:**
```yaml
# High availability deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slurm-exporter
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  template:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app.kubernetes.io/name
                  operator: In
                  values: ["slurm-exporter"]
              topologyKey: kubernetes.io/hostname
```

### Pattern 2: Active-Passive with Leader Election

```
┌─────────────────┐
│   Prometheus    │
└─────────┬───────┘
          │
    ┌─────▼─────┐
    │  Service  │
    └─────┬─────┘
          │
    ┌─────▼─────┐    ┌─────────────┐    ┌─────────────┐
    │ Leader    │    │ Standby     │    │ Standby     │
    │ Exporter  │    │ Exporter    │    │ Exporter    │
    │ (Active)  │    │ (Passive)   │    │ (Passive)   │
    └───────────┘    └─────────────┘    └─────────────┘
          │                 │                    │
          └─────────────────┼────────────────────┘
                            │
                  ┌─────────▼─────────┐
                  │   SLURM Cluster   │
                  └───────────────────┘
```

**Benefits:**
- Avoids duplicate metrics
- Consistent data collection
- Resource efficiency
- Simpler Prometheus configuration

**Implementation:**
```yaml
# Leader election configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: slurm-exporter-config
data:
  config.yaml: |
    server:
      leaderElection:
        enabled: true
        leaseDuration: 15s
        renewDeadline: 10s
        retryPeriod: 2s
```

### Pattern 3: Multi-Region Deployment

```
Region A                          Region B
┌─────────────────┐               ┌─────────────────┐
│   Prometheus    │               │   Prometheus    │
│   Instance A    │               │   Instance B    │
└─────────┬───────┘               └─────────┬───────┘
          │                                 │
    ┌─────▼─────┐                     ┌─────▼─────┐
    │ Exporter  │                     │ Exporter  │
    │ Cluster A │                     │ Cluster B │
    └─────┬─────┘                     └─────┬─────┘
          │                                 │
┌─────────▼─────────┐               ┌─────────▼─────────┐
│   SLURM Cluster   │               │   SLURM Cluster   │
│      Region A     │               │      Region B     │
└───────────────────┘               └───────────────────┘
```

**Use Cases:**
- Geographically distributed SLURM clusters
- Disaster recovery requirements
- Regional compliance needs
- Network latency optimization

## Kubernetes Deployment Strategies

### Rolling Updates

**Configuration:**
```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 25%
      maxSurge: 25%
```

**Benefits:**
- Zero downtime deployments
- Gradual rollout
- Easy rollback capability
- Production-safe updates

### Blue-Green Deployments

**Implementation using Argo Rollouts:**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: slurm-exporter
spec:
  strategy:
    blueGreen:
      activeService: slurm-exporter-active
      previewService: slurm-exporter-preview
      autoPromotionEnabled: false
      scaleDownDelaySeconds: 30
      prePromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: slurm-exporter-preview
```

### Canary Deployments

**Configuration:**
```yaml
spec:
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 1m}
      - setWeight: 25
      - pause: {duration: 1m}
      - setWeight: 50
      - pause: {duration: 1m}
      - setWeight: 100
```

## Load Balancing and Service Discovery

### Service Configuration

```yaml
apiVersion: v1
kind: Service
metadata:
  name: slurm-exporter
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
    prometheus.io/path: "/metrics"
spec:
  type: ClusterIP
  sessionAffinity: None
  ports:
  - name: http-metrics
    port: 8080
    targetPort: 8080
    protocol: TCP
  selector:
    app.kubernetes.io/name: slurm-exporter
```

### Headless Service for Direct Pod Access

```yaml
apiVersion: v1
kind: Service
metadata:
  name: slurm-exporter-headless
spec:
  clusterIP: None
  ports:
  - name: http-metrics
    port: 8080
    targetPort: 8080
  selector:
    app.kubernetes.io/name: slurm-exporter
```

### Service Mesh Integration (Istio)

```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: slurm-exporter
spec:
  host: slurm-exporter
  trafficPolicy:
    loadBalancer:
      simple: LEAST_CONN
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        maxRequestsPerConnection: 10
    outlierDetection:
      consecutiveErrors: 3
      interval: 30s
      baseEjectionTime: 30s
```

## Monitoring and Health Checks

### Comprehensive Health Checks

```yaml
spec:
  template:
    spec:
      containers:
      - name: slurm-exporter
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        startupProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 30
```

### Prometheus Monitoring Rules

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: slurm-exporter-ha
spec:
  groups:
  - name: slurm-exporter.availability
    rules:
    - alert: SlurmExporterDown
      expr: up{job="slurm-exporter"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "SLURM Exporter is down"
        description: "SLURM Exporter {{ $labels.instance }} has been down for more than 1 minute."

    - alert: SlurmExporterHighErrorRate
      expr: rate(slurm_exporter_errors_total[5m]) > 0.1
      for: 2m
      labels:
        severity: warning
      annotations:
        summary: "High error rate in SLURM Exporter"
        description: "SLURM Exporter {{ $labels.instance }} error rate is {{ $value }} errors/sec."

    - alert: SlurmExporterMissingMetrics
      expr: absent(slurm_cluster_nodes_total)
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "SLURM metrics missing"
        description: "Critical SLURM metrics have been missing for more than 5 minutes."

    - alert: SlurmExporterHighMemoryUsage
      expr: process_resident_memory_bytes{job="slurm-exporter"} > 512 * 1024 * 1024
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "High memory usage in SLURM Exporter"
        description: "SLURM Exporter {{ $labels.instance }} is using {{ $value | humanize1024 }}B of memory."

    - alert: SlurmApiUnreachable
      expr: slurm_exporter_api_errors_total > 10
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "SLURM API unreachable"
        description: "SLURM API has been unreachable for more than 5 minutes."
```

### Grafana Dashboard for HA Monitoring

```json
{
  "dashboard": {
    "title": "SLURM Exporter - High Availability",
    "panels": [
      {
        "title": "Exporter Instances Availability",
        "type": "stat",
        "targets": [
          {
            "expr": "count(up{job=\"slurm-exporter\"} == 1)",
            "legendFormat": "Available Instances"
          },
          {
            "expr": "count(up{job=\"slurm-exporter\"})",
            "legendFormat": "Total Instances"
          }
        ]
      },
      {
        "title": "Request Success Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(prometheus_http_requests_total{job=\"slurm-exporter\",code=~\"2..\"}[5m]) / rate(prometheus_http_requests_total{job=\"slurm-exporter\"}[5m])",
            "legendFormat": "Success Rate - {{ $labels.instance }}"
          }
        ]
      },
      {
        "title": "Failover Events",
        "type": "graph",
        "targets": [
          {
            "expr": "increase(slurm_exporter_leader_changes_total[1h])",
            "legendFormat": "Leader Changes"
          }
        ]
      }
    ]
  }
}
```

## Disaster Recovery

### Backup Strategies

**Configuration Backup:**
```bash
#!/bin/bash
# backup-config.sh

NAMESPACE="monitoring"
BACKUP_DIR="/backup/slurm-exporter/$(date +%Y%m%d)"

mkdir -p "$BACKUP_DIR"

# Backup Helm values
helm get values slurm-exporter -n "$NAMESPACE" > "$BACKUP_DIR/helm-values.yaml"

# Backup ConfigMaps
kubectl get configmap -n "$NAMESPACE" -o yaml > "$BACKUP_DIR/configmaps.yaml"

# Backup Secrets (without sensitive data)
kubectl get secrets -n "$NAMESPACE" -o yaml | \
  grep -v "data:" > "$BACKUP_DIR/secrets-metadata.yaml"

# Backup ServiceMonitor
kubectl get servicemonitor -n "$NAMESPACE" -o yaml > "$BACKUP_DIR/servicemonitor.yaml"
```

**Automated Backup with CronJob:**
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: slurm-exporter-backup
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: bitnami/kubectl:latest
            command: ["/bin/bash", "/scripts/backup-config.sh"]
            volumeMounts:
            - name: backup-script
              mountPath: /scripts
            - name: backup-storage
              mountPath: /backup
          volumes:
          - name: backup-script
            configMap:
              name: backup-scripts
              defaultMode: 0755
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc
```

### Recovery Procedures

**Quick Recovery Script:**
```bash
#!/bin/bash
# quick-recovery.sh

set -e

NAMESPACE="monitoring"
BACKUP_DIR="/backup/slurm-exporter/latest"

echo "Starting SLURM Exporter recovery..."

# Restore namespace if needed
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Restore ConfigMaps
kubectl apply -f "$BACKUP_DIR/configmaps.yaml"

# Restore Secrets (requires manual intervention for sensitive data)
echo "⚠️  Please restore secret data manually from secure backup"

# Deploy using Helm
helm upgrade --install slurm-exporter ./charts/slurm-exporter \
  --namespace "$NAMESPACE" \
  --values "$BACKUP_DIR/helm-values.yaml" \
  --wait

# Verify deployment
kubectl rollout status deployment/slurm-exporter -n "$NAMESPACE"

echo "✅ Recovery completed successfully"
```

### Multi-Cluster Recovery

```yaml
# Cross-cluster replication using external-secrets
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "https://vault.company.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "slurm-exporter"

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: slurm-credentials
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: slurm-auth
    creationPolicy: Owner
  data:
  - secretKey: token
    remoteRef:
      key: slurm/credentials
      property: token
```

## Performance and Scaling

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: slurm-exporter-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: slurm-exporter
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: slurm_exporter_scrape_duration_seconds
      target:
        type: AverageValue
        averageValue: "5"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
```

### Vertical Pod Autoscaler

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: slurm-exporter-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: slurm-exporter
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: slurm-exporter
      maxAllowed:
        cpu: 2
        memory: 4Gi
      minAllowed:
        cpu: 100m
        memory: 128Mi
```

### Performance Tuning Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: slurm-exporter-perf-config
data:
  config.yaml: |
    # Performance optimizations for large clusters
    server:
      maxConcurrentScrapes: 10
      scrapeTimeout: 30s
      
    slurm:
      # Connection pooling
      maxConnections: 20
      connectionTimeout: 10s
      requestTimeout: 30s
      keepAlive: true
      
      # Rate limiting
      rateLimit:
        enabled: true
        requestsPerSecond: 10
        burstSize: 20
      
      # Caching
      cache:
        enabled: true
        ttl: 60s
        maxSize: 1000
        
    collectors:
      # Staggered collection intervals
      cluster:
        interval: 30s
        timeout: 20s
      nodes:
        interval: 60s
        timeout: 30s
        batchSize: 100
      jobs:
        interval: 30s
        timeout: 25s
        batchSize: 500
      partitions:
        interval: 120s
        timeout: 60s
        
    # Resource management
    resources:
      limits:
        cpu: 2
        memory: 2Gi
      requests:
        cpu: 500m
        memory: 512Mi
```

## Security Considerations

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: slurm-exporter-netpol
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: slurm-exporter
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    - podSelector:
        matchLabels:
          app.kubernetes.io/name: prometheus
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to: []  # Allow all outbound (for SLURM API)
    ports:
    - protocol: TCP
      port: 6820  # SLURM REST API
    - protocol: TCP
      port: 443   # HTTPS
    - protocol: TCP
      port: 53    # DNS
    - protocol: UDP
      port: 53    # DNS
```

### Pod Security Standards

```yaml
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 65534
    runAsGroup: 65534
    fsGroup: 65534
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: slurm-exporter
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      readOnlyRootFilesystem: true
      runAsNonRoot: true
    volumeMounts:
    - name: tmp
      mountPath: /tmp
    - name: cache
      mountPath: /cache
  volumes:
  - name: tmp
    emptyDir: {}
  - name: cache
    emptyDir: {}
```

### Secret Management

```yaml
# Using sealed-secrets for GitOps
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: slurm-credentials
spec:
  encryptedData:
    token: AgBy3i4OJSWK+PiTySYZZA9rO43cGDEQAx...
    username: AgAKAoiQm32...
    password: AgAQW2PU32...
  template:
    metadata:
      name: slurm-credentials
    type: Opaque
```

## Operational Procedures

### Pre-Deployment Checklist

```bash
#!/bin/bash
# pre-deployment-check.sh

echo "🔍 Pre-deployment checks for SLURM Exporter HA"

# Check cluster resources
echo "📊 Checking cluster resources..."
kubectl top nodes
kubectl describe nodes | grep -E "(Allocated resources|Resource)"

# Verify prerequisites
echo "🔧 Verifying prerequisites..."
helm version
kubectl version --client

# Test SLURM connectivity
echo "🌐 Testing SLURM API connectivity..."
SLURM_URL="${SLURM_URL:-http://slurm-server:6820}"
curl -f "$SLURM_URL/slurm/v0.0.40/ping" || echo "⚠️  SLURM API unreachable"

# Check monitoring stack
echo "📈 Checking monitoring stack..."
kubectl get prometheus -A
kubectl get alertmanager -A

# Verify storage
echo "💾 Checking storage classes..."
kubectl get storageclass

echo "✅ Pre-deployment checks completed"
```

### Post-Deployment Validation

```bash
#!/bin/bash
# post-deployment-validation.sh

NAMESPACE="monitoring"
RELEASE="slurm-exporter"

echo "🧪 Post-deployment validation for SLURM Exporter HA"

# Check deployment status
echo "📋 Deployment status..."
kubectl rollout status deployment/$RELEASE -n $NAMESPACE

# Verify all pods are running
echo "🔄 Pod status..."
kubectl get pods -n $NAMESPACE -l app.kubernetes.io/instance=$RELEASE

# Test health endpoints
echo "❤️  Health check..."
kubectl port-forward -n $NAMESPACE service/$RELEASE 8080:8080 &
PF_PID=$!
sleep 5

curl -f http://localhost:8080/health || echo "❌ Health check failed"
curl -f http://localhost:8080/ready || echo "❌ Readiness check failed"

# Verify metrics
echo "📊 Metrics validation..."
METRICS_COUNT=$(curl -s http://localhost:8080/metrics | grep -c "^slurm_")
echo "Found $METRICS_COUNT SLURM metrics"

kill $PF_PID

# Check Prometheus targets
echo "🎯 Prometheus target status..."
kubectl exec -n monitoring prometheus-0 -- \
  wget -qO- http://localhost:9090/api/v1/targets | \
  grep slurm-exporter

echo "✅ Post-deployment validation completed"
```

### Maintenance Procedures

**Rolling Maintenance:**
```bash
#!/bin/bash
# rolling-maintenance.sh

NAMESPACE="monitoring"
DEPLOYMENT="slurm-exporter"

echo "🔧 Starting rolling maintenance for $DEPLOYMENT"

# Scale down to minimum
kubectl scale deployment $DEPLOYMENT -n $NAMESPACE --replicas=1

# Update one pod at a time
kubectl patch deployment $DEPLOYMENT -n $NAMESPACE -p '{"spec":{"strategy":{"rollingUpdate":{"maxUnavailable":1,"maxSurge":0}}}}'

# Perform updates
kubectl set image deployment/$DEPLOYMENT -n $NAMESPACE \
  slurm-exporter=slurm-exporter:v1.1.0

# Wait for rollout
kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE

# Scale back up
kubectl scale deployment $DEPLOYMENT -n $NAMESPACE --replicas=3

echo "✅ Rolling maintenance completed"
```

**Emergency Procedures:**
```bash
#!/bin/bash
# emergency-recovery.sh

NAMESPACE="monitoring"

echo "🚨 Emergency recovery procedures"

# Quick health check
echo "🏥 Quick health assessment..."
kubectl get pods -n $NAMESPACE --field-selector=status.phase!=Running

# Force restart unhealthy pods
echo "🔄 Restarting unhealthy pods..."
kubectl delete pods -n $NAMESPACE \
  --field-selector=status.phase!=Running \
  --grace-period=0 --force

# Check critical alerts
echo "🚨 Checking critical alerts..."
kubectl logs -n monitoring alertmanager-0 | grep -i critical

# Fallback to previous version if needed
echo "⏪ Preparing rollback option..."
helm history slurm-exporter -n $NAMESPACE

echo "Manual rollback command:"
echo "helm rollback slurm-exporter <revision> -n $NAMESPACE"
```

This comprehensive high availability documentation provides the foundation for deploying and operating SLURM Exporter in production environments with enterprise-grade reliability requirements.