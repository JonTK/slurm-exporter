# Security Violation Runbook

**Alert**: `SlurmExporterSecurityViolation`  
**Severity**: P0 Critical  
**Response Time**: < 2 minutes  

## Symptoms
- Security policy violations detected
- Unauthorized access attempts
- Privilege escalation attempts
- Suspicious network traffic patterns
- Container security context violations

## Immediate Actions (< 2 minutes)

### 1. STOP AND ASSESS - Do Not Take Actions That Destroy Evidence
```bash
# DO NOT restart pods or change configurations immediately
# Preserve logs and state for forensic analysis

# Acknowledge security alert
# Notify security team immediately: security@example.com
# Join #security-incident Slack channel
```

### 2. Isolate Affected Components
```bash
# Scale down potentially compromised instances (but keep 1 for analysis)
kubectl scale deployment slurm-exporter --replicas=1 -n slurm-exporter

# Apply network isolation (if network policy violation)
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: slurm-exporter-isolation
  namespace: slurm-exporter
spec:
  podSelector:
    matchLabels:
      app: slurm-exporter
  policyTypes:
  - Ingress
  - Egress
  # Deny all traffic except to logging/monitoring
EOF
```

### 3. Secure Evidence
```bash
# Capture current state
kubectl get all -n slurm-exporter -o yaml > incident-state-$(date +%s).yaml

# Capture logs immediately (before they rotate)
kubectl logs deployment/slurm-exporter -n slurm-exporter --all-containers=true > security-logs-$(date +%s).log
kubectl logs deployment/slurm-exporter -n slurm-exporter --previous --all-containers=true > security-logs-previous-$(date +%s).log

# Capture pod details
kubectl describe pods -n slurm-exporter > pod-details-$(date +%s).txt
```

## Investigation Procedures

### Security Context Analysis
```bash
# Check for privilege escalation
kubectl get pods -n slurm-exporter -o jsonpath='{.items[*].spec.securityContext}' | jq

# Check container security contexts
kubectl get pods -n slurm-exporter -o jsonpath='{.items[*].spec.containers[*].securityContext}' | jq

# Check for privileged containers
kubectl get pods -n slurm-exporter -o yaml | grep -A 5 -B 5 "privileged\|allowPrivilegeEscalation\|runAsRoot"

# Check capabilities
kubectl get pods -n slurm-exporter -o yaml | grep -A 10 -B 5 capabilities
```

### Access Pattern Analysis
```bash
# Check recent access to secrets
kubectl get events -n slurm-exporter --sort-by='.lastTimestamp' | grep -i "secret\|credential"

# Review RBAC permissions
kubectl describe rolebinding -n slurm-exporter
kubectl describe clusterrolebinding | grep slurm-exporter

# Check service account usage
kubectl get serviceaccount slurm-exporter -n slurm-exporter -o yaml
```

### Network Security Analysis
```bash
# Check network policies
kubectl get networkpolicy -n slurm-exporter -o yaml

# Review ingress/egress rules
kubectl describe networkpolicy -n slurm-exporter

# Check for unexpected network connections
kubectl exec deployment/slurm-exporter -n slurm-exporter -- netstat -tuln
kubectl exec deployment/slurm-exporter -n slurm-exporter -- ss -tuln
```

### Container Image Analysis
```bash
# Check image integrity
kubectl get pods -n slurm-exporter -o jsonpath='{.items[*].spec.containers[*].image}'

# Verify image signatures (if using cosign)
cosign verify ghcr.io/jontk/slurm-exporter:v1.0.0

# Check for image vulnerabilities
trivy image ghcr.io/jontk/slurm-exporter:v1.0.0
```

### Audit Log Analysis
```bash
# Search Kubernetes audit logs for suspicious activity
kubectl logs -n kube-system -l component=kube-apiserver | grep slurm-exporter | grep -E "(create|delete|patch|update)"

# Check for unauthorized API access
kubectl logs -n kube-system -l component=kube-apiserver | grep "user.*slurm" | grep -v "system:serviceaccount"

# Review authentication events
kubectl get events --all-namespaces --sort-by='.lastTimestamp' | grep -i "auth\|login\|token"
```

## Common Security Violations

### 1. Privilege Escalation
**Symptoms**: Pod running as root, privileged containers

**Investigation**:
```bash
# Check current security context
kubectl get pods -n slurm-exporter -o yaml | grep -A 20 securityContext

# Check for privilege escalation attempts in logs
kubectl logs deployment/slurm-exporter -n slurm-exporter | grep -i "permission\|denied\|unauthorized\|sudo\|su\|root"

# Check process list in container
kubectl exec deployment/slurm-exporter -n slurm-exporter -- ps auxf
```

**Immediate Actions**:
```bash
# Apply strict security context
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "securityContext": {
          "runAsNonRoot": true,
          "runAsUser": 65534,
          "fsGroup": 65534
        },
        "containers": [{
          "name": "slurm-exporter",
          "securityContext": {
            "allowPrivilegeEscalation": false,
            "readOnlyRootFilesystem": true,
            "runAsNonRoot": true,
            "runAsUser": 65534,
            "capabilities": {
              "drop": ["ALL"]
            }
          }
        }]
      }
    }
  }
}'
```

### 2. Unauthorized Network Access
**Symptoms**: Connections to unexpected hosts, policy violations

**Investigation**:
```bash
# Check active connections
kubectl exec deployment/slurm-exporter -n slurm-exporter -- netstat -tan | grep ESTABLISHED

# Check DNS queries
kubectl exec deployment/slurm-exporter -n slurm-exporter -- cat /etc/resolv.conf

# Review network policy logs (if supported)
# Check CNI logs for policy violations
```

**Immediate Actions**:
```bash
# Apply restrictive network policy
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: slurm-exporter-restricted
  namespace: slurm-exporter
spec:
  podSelector:
    matchLabels:
      app: slurm-exporter
  policyTypes:
  - Egress
  egress:
  # Only allow DNS
  - to: []
    ports:
    - protocol: UDP
      port: 53
  # Only allow SLURM API
  - to:
    - namespaceSelector:
        matchLabels:
          name: slurm-cluster
    ports:
    - protocol: TCP
      port: 6820
EOF
```

### 3. Credential Theft Attempts
**Symptoms**: Unauthorized secret access, token manipulation

**Investigation**:
```bash
# Check secret access patterns
kubectl get events -n slurm-exporter | grep -i secret

# Review mounted secrets
kubectl describe pods -n slurm-exporter | grep -A 10 -B 10 "Mounts:\|Volumes:"

# Check for credential exposure in logs
kubectl logs deployment/slurm-exporter -n slurm-exporter | grep -E "(token|password|key|secret)" | head -20
```

**Immediate Actions**:
```bash
# Rotate all credentials immediately
# Generate new SLURM token
NEW_TOKEN=$(scontrol token username=slurm-exporter-new)

# Update secret with new credentials
kubectl create secret generic slurm-credentials-new \
  --from-literal=rest-url="$SLURM_REST_URL" \
  --from-literal=auth-type="jwt" \
  --from-literal=jwt-token="$NEW_TOKEN" \
  -n slurm-exporter

# Update deployment to use new secret
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "envFrom": [{
            "secretRef": {
              "name": "slurm-credentials-new"
            }
          }]
        }]
      }
    }
  }
}'

# Delete old secret
kubectl delete secret slurm-credentials -n slurm-exporter
```

### 4. Container Breakout Attempts
**Symptoms**: Attempts to access host filesystem, privileged operations

**Investigation**:
```bash
# Check for host mounts
kubectl get pods -n slurm-exporter -o yaml | grep -A 10 -B 5 "hostPath\|/proc\|/sys\|/dev"

# Check for privileged operations in logs
kubectl logs deployment/slurm-exporter -n slurm-exporter | grep -E "(mount|chroot|pivot_root|unshare)"

# Check container capabilities
kubectl describe pods -n slurm-exporter | grep -A 5 -B 5 capabilities
```

**Immediate Actions**:
```bash
# Remove all host mounts and enforce read-only filesystem
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "securityContext": {
            "readOnlyRootFilesystem": true,
            "allowPrivilegeEscalation": false,
            "capabilities": {
              "drop": ["ALL"]
            }
          }
        }],
        "volumes": [
          {
            "name": "tmp",
            "emptyDir": {}
          }
        ]
      }
    }
  }
}'
```

## Forensic Evidence Collection

### System State Capture
```bash
# Create forensic snapshot
INCIDENT_ID="SEC-$(date +%Y%m%d-%H%M%S)"
mkdir -p /tmp/forensics/$INCIDENT_ID

# Capture all Kubernetes state
kubectl get all -A -o yaml > /tmp/forensics/$INCIDENT_ID/k8s-state.yaml
kubectl describe all -n slurm-exporter > /tmp/forensics/$INCIDENT_ID/descriptions.txt

# Capture events
kubectl get events --all-namespaces --sort-by='.lastTimestamp' > /tmp/forensics/$INCIDENT_ID/events.log

# Capture audit logs (if available)
kubectl logs -n kube-system -l component=kube-apiserver > /tmp/forensics/$INCIDENT_ID/audit.log
```

### Container Forensics
```bash
# Create container filesystem snapshot
kubectl exec deployment/slurm-exporter -n slurm-exporter -- tar czf - / > /tmp/forensics/$INCIDENT_ID/container-fs.tar.gz

# Capture process information
kubectl exec deployment/slurm-exporter -n slurm-exporter -- ps auxf > /tmp/forensics/$INCIDENT_ID/processes.txt

# Capture network state
kubectl exec deployment/slurm-exporter -n slurm-exporter -- netstat -tuln > /tmp/forensics/$INCIDENT_ID/network.txt

# Capture environment
kubectl exec deployment/slurm-exporter -n slurm-exporter -- env > /tmp/forensics/$INCIDENT_ID/environment.txt
```

### Log Preservation
```bash
# Collect all logs with timestamps
kubectl logs deployment/slurm-exporter -n slurm-exporter --timestamps=true > /tmp/forensics/$INCIDENT_ID/application.log
kubectl logs deployment/slurm-exporter -n slurm-exporter --previous --timestamps=true > /tmp/forensics/$INCIDENT_ID/application-previous.log

# System logs from nodes
kubectl get nodes -o jsonpath='{.items[*].metadata.name}' | xargs -I {} kubectl logs -n kube-system -l component=kubelet > /tmp/forensics/$INCIDENT_ID/node-{}.log
```

## Containment and Recovery

### Immediate Containment
```bash
# Isolate namespace
kubectl label namespace slurm-exporter security-incident=true

# Apply pod security policy
kubectl apply -f - <<EOF
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: slurm-exporter-restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  runAsUser:
    rule: 'MustRunAsNonRoot'
  readOnlyRootFilesystem: true
  volumes:
    - 'configMap'
    - 'secret'
    - 'emptyDir'
EOF
```

### Clean Recovery
```bash
# Deploy clean configuration from known-good state
kubectl delete deployment slurm-exporter -n slurm-exporter
kubectl delete secret slurm-credentials -n slurm-exporter

# Redeploy from secure baseline
kubectl apply -f deployment/production/kubernetes-ha/

# Verify security posture
kubectl get pods -n slurm-exporter -o yaml | grep -A 20 securityContext
```

## Communication and Escalation

### Security Team Notification
```
🚨 SECURITY INCIDENT: SLURM Exporter
SEVERITY: P0 Critical
INCIDENT ID: SEC-20240101-1200
NATURE: [Brief description of violation]
STATUS: Contained and investigating
IMPACT: [Scope of potential compromise]
EVIDENCE: Preserved in /tmp/forensics/SEC-20240101-1200/
NEXT UPDATE: [Time]
```

### Executive Communication
```
SECURITY ALERT - IMMEDIATE ATTENTION REQUIRED

A security violation has been detected in the SLURM Exporter service.

WHAT: [Nature of security violation]
WHEN: [Detection time]
STATUS: Incident contained, investigation ongoing
IMPACT: [Business impact assessment]
ACTIONS: [Immediate containment actions taken]

Security team is leading the investigation.
Next update in [timeframe].
```

## Post-Incident Actions

### Security Review
1. Complete forensic analysis
2. Assess scope of compromise
3. Review security policies and controls
4. Update security monitoring
5. Conduct security training

### Preventive Measures
1. Implement additional security controls
2. Enhance monitoring and alerting
3. Review and update RBAC policies
4. Implement pod security standards
5. Regular security assessments

### Compliance Reporting
1. Document incident timeline
2. Report to regulatory bodies (if required)
3. Update security documentation
4. Review compliance requirements
5. Implement corrective actions

## Related Runbooks
- [Authentication Failure](./authentication-failure.md)
- [Network Policy Issues](./network-policy.md)
- [Certificate Renewal](./certificate-renewal.md)
- [Backup and Recovery](./backup-recovery.md)

## Security Contacts
- **Security Team**: security@example.com / +1-555-SECURITY
- **SOC**: soc@example.com / +1-555-SOC
- **Legal/Compliance**: legal@example.com
- **CISO**: ciso@example.com