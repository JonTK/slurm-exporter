# Security Framework

This directory contains a comprehensive security framework for SLURM Exporter production deployments, implementing enterprise-grade security controls, compliance automation, and continuous security monitoring.

## Overview

The security framework provides:
- **Pod Security Standards**: Restricted security policies and hardened deployments
- **Network Security**: Multi-layered network policies and service mesh security
- **Secrets Management**: External secrets, automated rotation, and encryption
- **Vulnerability Scanning**: Automated container and dependency scanning
- **Compliance Automation**: Continuous compliance monitoring and reporting
- **Runtime Security**: Real-time threat detection and response

## Framework Components

### 1. Pod Security Standards (`pod-security-standards.yaml`)

#### Security Hardening Features
- **Restricted Pod Security Standards**: Enforces the most restrictive security policies
- **Non-root Execution**: All containers run as non-root user (65534)
- **Read-only Root Filesystem**: Prevents runtime modifications
- **Capability Dropping**: Removes all Linux capabilities
- **Security Context Constraints**: OpenShift-compatible security constraints

#### Deployment Security
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65534
  runAsGroup: 65534
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop: [ALL]
  seccompProfile:
    type: RuntimeDefault
```

#### Security Validation
- **Init Container Validation**: Pre-flight security checks
- **Runtime Security Verification**: Continuous security posture monitoring
- **Resource Constraints**: Security-conscious resource limits

### 2. Network Security (`network-security.yaml`)

#### Network Policies
- **Default Deny**: Explicit deny-all policy with granular allow rules
- **Ingress Controls**: Restricted access for monitoring systems only
- **Egress Controls**: Limited outbound access to SLURM APIs and essential services
- **IP Restrictions**: CIDR-based access controls

#### Service Mesh Security (Istio)
- **Mutual TLS**: Strict mTLS enforcement between services
- **Authorization Policies**: Fine-grained access control
- **Security Headers**: Comprehensive HTTP security headers
- **Access Logging**: Detailed security event logging

#### Runtime Network Monitoring
- **Falco Integration**: Real-time network anomaly detection
- **Connection Monitoring**: Tracking unexpected outbound connections
- **Data Exfiltration Detection**: Large data transfer monitoring

### 3. Secrets Management (`secrets-management.yaml`)

#### External Secrets Integration
- **HashiCorp Vault**: Enterprise secrets management
- **External Secrets Operator**: Kubernetes-native secrets sync
- **Sealed Secrets**: GitOps-compatible encrypted secrets

#### Automated Rotation
- **Monthly Rotation**: Automated credential rotation every 30 days
- **JWT Token Management**: SLURM JWT token lifecycle management
- **Certificate Management**: Automated TLS certificate renewal

#### Encryption at Rest
```yaml
annotations:
  encryption.kubernetes.io/provider: "aes-cbc"
  encryption.kubernetes.io/key-id: "slurm-exporter-key-2024"
  secrets.kubernetes.io/rotation-schedule: "0 0 1 * *"
```

### 4. Vulnerability Scanning (`vulnerability-scanning.yaml`)

#### Container Scanning
- **Daily Trivy Scans**: Comprehensive container vulnerability scanning
- **Severity Thresholds**: Policy-based vulnerability management
  - Critical: 0 allowed
  - High: 5 maximum
  - Medium: 20 maximum
  - Low: 100 maximum

#### Kubernetes Security Scanning
- **Weekly kube-bench**: CIS Kubernetes Benchmark validation
- **Polaris Configuration**: Best practice configuration scanning
- **Security Policy Validation**: Continuous policy compliance checking

#### Dependency Scanning
- **Daily Grype Scans**: Software composition analysis
- **License Compliance**: Open source license validation
- **Supply Chain Security**: Dependency vulnerability tracking

### 5. Compliance Automation (`compliance-automation.yaml`)

#### Supported Frameworks
- **SOC 2 Type II**: Comprehensive security and availability controls
- **CIS Kubernetes Benchmark**: Industry-standard security benchmarks
- **NIST Cybersecurity Framework**: Federal cybersecurity guidelines
- **Custom Policies**: Organization-specific compliance requirements

#### Automated Assessments
- **Weekly Compliance Scans**: Comprehensive compliance validation
- **Continuous Monitoring**: Real-time compliance posture tracking
- **Automated Reporting**: Stakeholder compliance reports

## Usage Instructions

### Initial Security Setup

#### 1. Deploy Security Framework
```bash
# Deploy all security components
kubectl apply -f pod-security-standards.yaml
kubectl apply -f network-security.yaml
kubectl apply -f secrets-management.yaml
kubectl apply -f vulnerability-scanning.yaml
kubectl apply -f compliance-automation.yaml

# Verify security namespace
kubectl get namespace slurm-exporter -o yaml
```

#### 2. Configure Secrets
```bash
# Create security notification secrets
kubectl create secret generic security-notifications \
  --from-literal=slack-webhook="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
  -n slurm-exporter

kubectl create secret generic compliance-notifications \
  --from-literal=slack-webhook="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
  -n slurm-exporter

kubectl create secret generic secrets-notifications \
  --from-literal=slack-webhook="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
  -n slurm-exporter

# Configure Vault integration (if using external secrets)
kubectl create secret generic vault-token \
  --from-literal=token="YOUR_VAULT_TOKEN" \
  -n slurm-exporter
```

#### 3. Verify Security Setup
```bash
# Check pod security enforcement
kubectl get pods -n slurm-exporter -o jsonpath='{.items[*].spec.securityContext}'

# Verify network policies
kubectl get networkpolicies -n slurm-exporter

# Check vulnerability scanning jobs
kubectl get cronjobs -n slurm-exporter | grep scan

# Verify compliance automation
kubectl get cronjobs -n slurm-exporter | grep compliance
```

### Security Operations

#### Daily Security Monitoring
```bash
# Check security scan results
kubectl logs -l component=vulnerability-scanning -n slurm-exporter --tail=100

# Monitor security alerts
kubectl get events -n slurm-exporter --field-selector type=Warning

# Review runtime security events
kubectl logs -l app=falco -n falco-system | grep slurm-exporter

# Check compliance status
kubectl logs job/compliance-assessment-$(date +%Y%m%d) -n slurm-exporter
```

#### Weekly Security Reviews
```bash
#!/bin/bash
# Weekly security review script

echo "=== Weekly Security Review ==="

# 1. Vulnerability Scan Summary
echo "Vulnerability Scan Results:"
kubectl logs job/container-vulnerability-scan-$(date -d "7 days ago" +%Y%m%d) -n slurm-exporter | \
  grep "Vulnerability Summary" -A 5

# 2. Compliance Assessment
echo "Compliance Status:"
kubectl logs job/compliance-assessment-$(date -d "1 week ago" +%Y%m%d) -n slurm-exporter | \
  grep "COMPLIANCE"

# 3. Security Events
echo "Security Events (Last 7 days):"
kubectl get events -n slurm-exporter --field-selector type=Warning \
  --sort-by='.lastTimestamp' | tail -20

# 4. Network Policy Violations
echo "Network Policy Violations:"
kubectl logs -l app=falco -n falco-system --since=168h | \
  grep "network.*policy.*violation" | wc -l

# 5. Secrets Rotation Status
echo "Secrets Rotation Status:"
kubectl get secret slurm-credentials -n slurm-exporter \
  -o jsonpath='{.metadata.annotations.secrets\.kubernetes\.io/last-rotation}'
```

#### Emergency Security Response
```bash
# Immediate security lockdown
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "replicas": 0
  }
}'

# Review recent security events
kubectl get events -n slurm-exporter --sort-by='.lastTimestamp' | tail -50

# Force secrets rotation
kubectl create job --from=cronjob/secrets-rotation \
  emergency-rotation-$(date +%s) -n slurm-exporter

# Run immediate vulnerability scan
kubectl create job --from=cronjob/container-vulnerability-scan \
  emergency-scan-$(date +%s) -n slurm-exporter
```

### Security Configuration Management

#### Updating Security Policies
```bash
# Update pod security policy
kubectl patch psp slurm-exporter-restricted -p '
{
  "spec": {
    "requiredDropCapabilities": ["ALL", "SYS_TIME"]
  }
}'

# Update network policies
kubectl apply -f network-security.yaml

# Update vulnerability thresholds
kubectl patch configmap vulnerability-scanning-config -n slurm-exporter -p '
{
  "data": {
    "scanning-policy.yaml": "updated-policy-content"
  }
}'
```

#### Security Baseline Management
```bash
# Export current security configuration
kubectl get psp,networkpolicy,secret -n slurm-exporter -o yaml > security-baseline.yaml

# Compare with previous baseline
diff security-baseline-previous.yaml security-baseline.yaml

# Update security documentation
git add security-baseline.yaml
git commit -m "Update security baseline $(date +%Y-%m-%d)"
```

## Security Monitoring and Alerting

### Prometheus Metrics
```promql
# Security scan failure rate
rate(slurm_security_scan_failures_total[5m])

# Compliance violations
slurm_compliance_violations_current

# Vulnerability counts by severity
slurm_vulnerabilities_total{severity="critical"}
slurm_vulnerabilities_total{severity="high"}

# Secrets rotation status
slurm_secrets_rotation_overdue

# Network policy violations
rate(falco_events_total{rule_name=~".*network.*policy.*"}[5m])

# Runtime security events
rate(falco_events_total{k8s_ns="slurm-exporter"}[5m])
```

### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "SLURM Exporter Security Dashboard",
    "panels": [
      {
        "title": "Security Scan Status",
        "type": "stat",
        "targets": [
          {
            "expr": "slurm_security_scan_last_success",
            "legendFormat": "Last Successful Scan"
          }
        ]
      },
      {
        "title": "Vulnerability Trends",
        "type": "timeseries",
        "targets": [
          {
            "expr": "slurm_vulnerabilities_total",
            "legendFormat": "{{severity}} vulnerabilities"
          }
        ]
      },
      {
        "title": "Compliance Status",
        "type": "stat",
        "targets": [
          {
            "expr": "slurm_compliance_score",
            "legendFormat": "Compliance Score"
          }
        ]
      }
    ]
  }
}
```

### Alert Manager Rules
```yaml
groups:
- name: slurm-exporter-security
  rules:
  - alert: CriticalVulnerabilityDetected
    expr: slurm_vulnerabilities_total{severity="critical"} > 0
    for: 0m
    labels:
      severity: critical
    annotations:
      summary: "Critical vulnerabilities detected in SLURM Exporter"
      runbook_url: "https://runbooks.example.com/security/critical-vulnerabilities"
  
  - alert: ComplianceViolation
    expr: slurm_compliance_score < 80
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Security compliance score below threshold"
      runbook_url: "https://runbooks.example.com/security/compliance-violations"
```

## Security Incident Response

### Incident Classification
- **Critical**: Active security breach, malware detection, data exfiltration
- **High**: Multiple critical vulnerabilities, compliance violations, failed security controls
- **Medium**: Non-critical security alerts, policy violations, configuration drift
- **Low**: Information security events, routine security notifications

### Response Procedures

#### Critical Security Incident
1. **Immediate Isolation**: Scale deployment to zero replicas
2. **Evidence Collection**: Preserve logs, capture forensic data
3. **Stakeholder Notification**: Alert security team and management
4. **Root Cause Analysis**: Investigate attack vectors and impact
5. **Remediation**: Apply security patches and controls
6. **Recovery**: Restore service with enhanced security

#### High Severity Incident
1. **Assessment**: Evaluate threat level and potential impact
2. **Containment**: Apply temporary security controls
3. **Investigation**: Analyze security events and logs
4. **Remediation**: Address vulnerabilities and misconfigurations
5. **Monitoring**: Enhanced monitoring for related threats

### Recovery Procedures
```bash
# Security incident recovery checklist

# 1. Force vulnerability scan
kubectl create job --from=cronjob/container-vulnerability-scan \
  incident-scan-$(date +%s) -n slurm-exporter

# 2. Rotate all secrets immediately
kubectl create job --from=cronjob/secrets-rotation \
  incident-rotation-$(date +%s) -n slurm-exporter

# 3. Update security baseline
kubectl apply -f pod-security-standards.yaml
kubectl apply -f network-security.yaml

# 4. Verify compliance
kubectl create job --from=cronjob/compliance-assessment \
  incident-compliance-$(date +%s) -n slurm-exporter

# 5. Enhanced monitoring
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "security.kubernetes.io/incident-recovery": "'$(date)'"
        }
      }
    }
  }
}'
```

## Troubleshooting

### Common Security Issues

#### Pod Security Policy Violations
```bash
# Check PSP events
kubectl get events -n slurm-exporter --field-selector reason=FailedCreate

# Verify security context
kubectl get pods -n slurm-exporter -o yaml | grep -A 10 securityContext

# Test security context manually
kubectl run test-security --image=alpine:3.18 \
  --dry-run=server -n slurm-exporter \
  --overrides='{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":65534}}}'
```

#### Network Policy Issues
```bash
# Test network connectivity
kubectl run test-network --image=alpine:3.18 -n slurm-exporter \
  --rm -it -- sh -c "wget -qO- http://slurm-exporter:10341/health"

# Check network policy logs
kubectl logs -l app=calico-node -n calico-system | grep DENY

# Verify policy application
kubectl describe networkpolicy -n slurm-exporter
```

#### Vulnerability Scan Failures
```bash
# Check scan job logs
kubectl logs job/container-vulnerability-scan-$(date +%Y%m%d) -n slurm-exporter

# Verify image accessibility
kubectl run test-scan --image=aquasec/trivy:latest \
  --rm -it -- trivy image ghcr.io/jontk/slurm-exporter:v1.0.0

# Check resource constraints
kubectl describe job/container-vulnerability-scan-$(date +%Y%m%d) -n slurm-exporter
```

#### Secrets Management Issues
```bash
# Check external secrets status
kubectl get externalsecret -n slurm-exporter
kubectl describe externalsecret slurm-credentials-external -n slurm-exporter

# Verify Vault connectivity
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  wget -qO- --timeout=5 https://vault.example.com/v1/sys/health

# Check secrets rotation
kubectl get job/secrets-rotation-$(date +%Y%m%d) -n slurm-exporter
```

### Security Validation Tests
```bash
#!/bin/bash
# Comprehensive security validation script

echo "Running security validation tests..."

# Test 1: Verify non-root execution
echo "Test 1: Non-root execution"
kubectl exec deployment/slurm-exporter -n slurm-exporter -- id

# Test 2: Verify read-only filesystem
echo "Test 2: Read-only filesystem"
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  sh -c "touch /test-file 2>&1 || echo 'Filesystem is read-only (PASS)'"

# Test 3: Verify capabilities dropped
echo "Test 3: Capabilities verification"
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  grep Cap /proc/self/status

# Test 4: Network policy enforcement
echo "Test 4: Network policy test"
kubectl run test-network --image=alpine:3.18 --rm -n default \
  -- sh -c "timeout 5 wget -qO- http://slurm-exporter.slurm-exporter:10341/health || echo 'Network policy enforced (PASS)'"

# Test 5: Secrets encryption
echo "Test 5: Secrets encryption"
kubectl get secret slurm-credentials -n slurm-exporter \
  -o jsonpath='{.metadata.annotations.encryption\.kubernetes\.io/provider}'

echo "Security validation tests completed"
```

This comprehensive security framework provides defense-in-depth protection for SLURM Exporter deployments, ensuring robust security posture through automated controls, continuous monitoring, and proactive threat detection.