# Service Down Runbook

**Alert**: `SlurmExporterDown`  
**Severity**: P0 Critical  
**Response Time**: < 5 minutes  

## Symptoms
- SLURM Exporter instances are not responding to health checks
- Prometheus cannot scrape metrics from any instances
- Complete loss of SLURM monitoring data
- Dashboard shows all instances as down

## Immediate Actions (< 5 minutes)

### 1. Acknowledge and Assess
```bash
# Acknowledge PagerDuty alert
# Join #incident-response Slack channel

# Quick health check
kubectl get pods -n slurm-exporter
kubectl get svc -n slurm-exporter
curl -f https://slurm-exporter.example.com/health
```

### 2. Check Service Status
```bash
# Pod status and events
kubectl describe pods -n slurm-exporter
kubectl get events -n slurm-exporter --sort-by='.lastTimestamp' | tail -20

# Service endpoints
kubectl get endpoints -n slurm-exporter
kubectl describe svc slurm-exporter -n slurm-exporter
```

### 3. Quick Recovery Actions
```bash
# If pods are failing, try restart
kubectl rollout restart deployment/slurm-exporter -n slurm-exporter

# If service issues, check ingress
kubectl get ingress -n slurm-exporter
kubectl describe ingress slurm-exporter -n slurm-exporter
```

## Investigation Procedures

### Check Pod Health
```bash
# Pod details
kubectl describe pod -l app=slurm-exporter -n slurm-exporter

# Container logs (last 100 lines)
kubectl logs deployment/slurm-exporter -n slurm-exporter --tail=100

# Previous container logs if restarted
kubectl logs deployment/slurm-exporter -n slurm-exporter --previous
```

### Check Resource Constraints
```bash
# Resource usage
kubectl top pods -n slurm-exporter
kubectl top nodes

# Resource quotas
kubectl describe quota -n slurm-exporter
kubectl describe limitrange -n slurm-exporter

# Node capacity
kubectl describe node | grep -A 5 "Allocated resources"
```

### Check Kubernetes Health
```bash
# Node status
kubectl get nodes

# System pods
kubectl get pods -n kube-system | grep -E "(dns|proxy|flannel|calico)"

# Cluster events
kubectl get events --all-namespaces --sort-by='.lastTimestamp' | tail -20
```

### Check Network Connectivity
```bash
# Network policies
kubectl get networkpolicy -n slurm-exporter
kubectl describe networkpolicy -n slurm-exporter

# Service discovery
kubectl exec -it deployment/slurm-exporter -n slurm-exporter -- \
  nslookup slurm-exporter.slurm-exporter.svc.cluster.local

# External connectivity
kubectl exec -it deployment/slurm-exporter -n slurm-exporter -- \
  curl -k "$SLURM_REST_URL/slurm/v0.0.44/ping"
```

## Common Causes and Solutions

### 1. Pod Crashes
**Symptoms**: Pods in CrashLoopBackOff or Error state

**Investigation**:
```bash
# Check exit codes and restart counts
kubectl get pods -n slurm-exporter -o wide

# Check container logs for errors
kubectl logs deployment/slurm-exporter -n slurm-exporter --tail=200 | grep -i error

# Check readiness/liveness probe failures
kubectl describe pod -l app=slurm-exporter -n slurm-exporter | grep -A 10 "Events:"
```

**Solutions**:
```bash
# If memory issues - increase limits
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "resources": {
            "limits": {
              "memory": "1Gi"
            }
          }
        }]
      }
    }
  }
}'

# If startup issues - increase probe timeouts
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "startupProbe": {
            "failureThreshold": 30,
            "periodSeconds": 10
          }
        }]
      }
    }
  }
}'
```

### 2. Image Pull Failures
**Symptoms**: Pods stuck in ImagePullBackOff

**Investigation**:
```bash
# Check image and pull policy
kubectl describe pod -l app=slurm-exporter -n slurm-exporter | grep -A 5 "Failed"

# Check image registry connectivity
docker pull ghcr.io/jontk/slurm-exporter:v1.0.0

# Check registry credentials
kubectl get secret registry-credentials -n slurm-exporter -o yaml
```

**Solutions**:
```bash
# Update to known good image
kubectl set image deployment/slurm-exporter \
  slurm-exporter=ghcr.io/jontk/slurm-exporter:v0.9.0 \
  -n slurm-exporter

# Fix registry credentials
kubectl create secret docker-registry registry-credentials \
  --docker-server=ghcr.io \
  --docker-username=username \
  --docker-password=token \
  -n slurm-exporter

kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "imagePullSecrets": [{"name": "registry-credentials"}]
      }
    }
  }
}'
```

### 3. Configuration Issues
**Symptoms**: Pods start but fail health checks

**Investigation**:
```bash
# Check ConfigMap
kubectl get configmap slurm-exporter-config -n slurm-exporter -o yaml

# Check secrets
kubectl get secret slurm-credentials -n slurm-exporter -o yaml

# Test configuration locally
kubectl exec -it deployment/slurm-exporter -n slurm-exporter -- \
  /bin/slurm-exporter --config=/etc/slurm-exporter/config.yaml --dry-run
```

**Solutions**:
```bash
# Rollback to previous config
kubectl rollout undo deployment/slurm-exporter -n slurm-exporter

# Fix configuration and restart
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "corrected-config-content"
  }
}'
kubectl rollout restart deployment/slurm-exporter -n slurm-exporter
```

### 4. SLURM API Connectivity
**Symptoms**: Pods healthy but no metrics collected

**Investigation**:
```bash
# Test SLURM API directly
kubectl exec -it deployment/slurm-exporter -n slurm-exporter -- \
  curl -v -k "$SLURM_REST_URL/slurm/v0.0.44/ping"

# Check authentication
kubectl exec -it deployment/slurm-exporter -n slurm-exporter -- env | grep SLURM

# Check network policies
kubectl get networkpolicy -n slurm-exporter
```

**Solutions**:
```bash
# Update SLURM credentials
kubectl create secret generic slurm-credentials \
  --from-literal=rest-url="https://slurm-head:6820" \
  --from-literal=jwt-token="$(scontrol token username=exporter)" \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods to pick up new credentials
kubectl rollout restart deployment/slurm-exporter -n slurm-exporter
```

### 5. Resource Exhaustion
**Symptoms**: Pods pending or evicted

**Investigation**:
```bash
# Check node resources
kubectl describe nodes | grep -A 10 "Allocated resources"

# Check resource quotas
kubectl describe quota -n slurm-exporter

# Check pod resource requests
kubectl describe deployment slurm-exporter -n slurm-exporter | grep -A 10 "Requests"
```

**Solutions**:
```bash
# Scale down temporarily
kubectl scale deployment slurm-exporter --replicas=1 -n slurm-exporter

# Increase resource quotas
kubectl patch resourcequota slurm-exporter-quota -n slurm-exporter -p '
{
  "spec": {
    "hard": {
      "requests.cpu": "4",
      "requests.memory": "8Gi",
      "limits.cpu": "8",
      "limits.memory": "16Gi"
    }
  }
}'

# Add more nodes (if possible)
# Contact infrastructure team for node scaling
```

## Escalation Criteria

### Escalate to L2 if:
- Service not restored within 15 minutes
- Multiple failure modes occurring simultaneously  
- Infrastructure-level issues identified
- Root cause requires code changes

### Escalate to L3 if:
- Service not restored within 1 hour
- Data loss or corruption suspected
- External vendor assistance required
- Executive communication needed

## Communication Template

### Initial Alert
```
🚨 CRITICAL: SLURM Exporter service is completely down
STATUS: Investigating
IMPACT: Complete loss of SLURM monitoring
ETA: Investigating (target 15 minutes)
ASSIGNED: [Your name]
```

### Investigation Update
```
🔍 UPDATE: SLURM Exporter down - investigating [specific issue]
PROGRESS: [Actions taken]
NEXT: [Next steps]
ETA: [Updated estimate]
```

### Resolution Notice
```
✅ RESOLVED: SLURM Exporter service restored
DURATION: [Total downtime]
CAUSE: [Brief root cause]
MONITORING: Watching for stability
```

## Prevention Actions

### Short-term
- Increase health check frequency
- Add additional monitoring for early detection
- Review resource quotas and limits
- Test failover procedures

### Long-term
- Implement circuit breakers for SLURM API calls
- Add chaos engineering tests
- Improve observability and logging
- Automate common recovery procedures

## Related Runbooks
- [Authentication Failure](./authentication-failure.md)
- [Performance Issues](./performance-slow.md)
- [Pod Scheduling Issues](./pod-scheduling.md)
- [Network Policy Issues](./network-policy.md)

## Post-Incident Actions
1. Update incident timeline
2. Schedule post-incident review
3. Update monitoring and alerting
4. Improve automation and runbooks
5. Conduct team retrospective