# Backup and Recovery Documentation

This directory contains comprehensive backup and disaster recovery procedures for the SLURM Exporter production deployment.

## Overview

The backup and recovery system provides:
- **Automated backups** of configurations, monitoring data, and deployment state
- **Multi-tier storage** with primary, secondary, and local backup destinations
- **Disaster recovery procedures** for various failure scenarios
- **Backup validation** and integrity testing
- **Recovery automation** with time-bound objectives

## Backup Strategy

### 3-2-1 Backup Rule
- **3 copies** of important data
- **2 different storage media** (cloud and local)
- **1 offsite location** (different region)

### Recovery Objectives
- **RPO (Recovery Point Objective)**: 1 hour maximum data loss
- **RTO (Recovery Time Objective)**: 30 minutes maximum recovery time
- **MTTR (Mean Time To Recovery)**: 15 minutes average recovery time

## Backup Components

### 1. Configuration Backups
- **Schedule**: Every 6 hours
- **Retention**: 90 days
- **Contents**: Kubernetes manifests, ConfigMaps, Secrets, NetworkPolicies
- **Encryption**: AES-256-CBC encryption
- **Compression**: gzip compression

### 2. Monitoring Data Backups
- **Schedule**: Daily at 2 AM
- **Retention**: 30 days
- **Contents**: Prometheus metrics, alerting rules, targets configuration
- **Format**: JSON exports and time-series data

### 3. Deployment State Backups
- **Schedule**: Every 12 hours
- **Retention**: 30 days
- **Contents**: Complete Kubernetes deployment state
- **Purpose**: Full environment reconstruction

### 4. Certificate Backups
- **Schedule**: Weekly (Sundays)
- **Retention**: 365 days
- **Contents**: TLS certificates and keys
- **Security**: Highest encryption standards

### 5. Log Backups
- **Schedule**: Daily at 3 AM
- **Retention**: 7 days
- **Contents**: Application and system logs
- **Purpose**: Troubleshooting and audit trails

## Storage Destinations

### Primary Storage (us-west-2)
```
S3 Bucket: slurm-exporter-backup-primary
├── configuration/
│   ├── config-backup-20240101-120000.tar.gz.enc
│   └── config-backup-20240101-180000.tar.gz.enc
├── monitoring/
│   ├── monitoring-backup-20240101-020000.tar.gz
│   └── monitoring-backup-20240102-020000.tar.gz
├── deployment/
│   ├── deployment-backup-20240101-000000.tar.gz.enc
│   └── deployment-backup-20240101-120000.tar.gz.enc
├── certificates/
│   └── cert-backup-20240101-010000.tar.gz.enc
└── logs/
    ├── log-backup-20240101-030000.tar.gz
    └── log-backup-20240102-030000.tar.gz
```

### Secondary Storage (us-east-1)
- **Purpose**: Cross-region disaster recovery
- **Content**: Mirror of primary storage
- **Access**: Automatic failover if primary unavailable

### Local Storage
- **Purpose**: Fast local recovery
- **Retention**: 7 days
- **Storage Class**: Fast SSD PersistentVolume
- **Size**: 100Gi

## Backup Operations

### Automated Backup Jobs

#### Configuration Backup CronJob
```bash
# View backup schedule
kubectl get cronjob configuration-backup -n slurm-exporter

# Check backup history
kubectl get jobs -n slurm-exporter -l backup-type=configuration

# View latest backup logs
kubectl logs job/configuration-backup-$(date +%Y%m%d) -n slurm-exporter
```

#### Monitoring Data Backup CronJob
```bash
# View monitoring backup status
kubectl get cronjob monitoring-data-backup -n slurm-exporter

# Check backup completion
kubectl get jobs -n slurm-exporter -l backup-type=monitoring
```

#### Backup Validation CronJob
```bash
# View validation schedule
kubectl get cronjob backup-validation -n slurm-exporter

# Check validation results
kubectl logs job/backup-validation-$(date +%Y%m%d) -n slurm-exporter
```

### Manual Backup Operations

#### Immediate Configuration Backup
```bash
# Trigger immediate backup
kubectl create job --from=cronjob/configuration-backup \
  manual-config-backup-$(date +%s) \
  -n slurm-exporter

# Monitor backup progress
kubectl logs -f job/manual-config-backup-$(date +%s) -n slurm-exporter
```

#### Emergency Backup Before Maintenance
```bash
# Create pre-maintenance backup
kubectl create job --from=cronjob/configuration-backup \
  pre-maintenance-backup-$(date +%s) \
  -n slurm-exporter

# Wait for completion
kubectl wait --for=condition=complete \
  job/pre-maintenance-backup-$(date +%s) \
  -n slurm-exporter --timeout=600s
```

## Disaster Recovery Scenarios

### 1. Single Pod Failure
- **Probability**: High
- **Impact**: Low
- **Detection**: 1 minute
- **Recovery**: 5 minutes (automatic)
- **Procedure**: Kubernetes automatic restart

### 2. Deployment Failure
- **Probability**: Medium
- **Impact**: Medium
- **Detection**: 2 minutes
- **Recovery**: 10 minutes (semi-automatic)
- **Procedure**: Automated configuration restore

### 3. Namespace Deletion
- **Probability**: Low
- **Impact**: High
- **Detection**: 1 minute
- **Recovery**: 30 minutes (manual)
- **Procedure**: Full namespace restoration

### 4. Cluster Failure
- **Probability**: Very Low
- **Impact**: Critical
- **Detection**: 5 minutes
- **Recovery**: 2 hours (manual)
- **Procedure**: New cluster setup + restoration

### 5. Region Failure
- **Probability**: Very Low
- **Impact**: Critical
- **Detection**: 10 minutes
- **Recovery**: 4 hours (manual)
- **Procedure**: Cross-region failover

## Recovery Procedures

### Quick Recovery (Deployment Failure)
```bash
# Start disaster recovery
kubectl apply -f disaster-recovery.yaml

# Or use quick script
./deployment/production/backup/scripts/quick-restore.sh deployment_failure latest

# Monitor recovery
kubectl logs -f job/dr-restore-$(date +%s) -n slurm-exporter
```

### Manual Recovery Steps

#### 1. Assess the Situation
```bash
# Check current state
kubectl get all -n slurm-exporter
kubectl get events -n slurm-exporter --sort-by='.lastTimestamp'

# Review recent changes
kubectl rollout history deployment/slurm-exporter -n slurm-exporter
```

#### 2. Download Latest Backup
```bash
# List available backups
aws s3 ls s3://slurm-exporter-backup-primary/configuration/

# Download specific backup
BACKUP_FILE="config-backup-20240101-120000.tar.gz.enc"
aws s3 cp "s3://slurm-exporter-backup-primary/configuration/$BACKUP_FILE" ./

# Decrypt and extract
openssl enc -aes-256-cbc -d -salt \
  -in "$BACKUP_FILE" \
  -out "backup.tar.gz" \
  -pass pass:"$ENCRYPTION_KEY"
tar -xzf backup.tar.gz
```

#### 3. Restore Configuration
```bash
# Restore ConfigMaps
kubectl apply -f backup-*/configmaps.yaml

# Restore Secrets (be careful)
kubectl apply -f backup-*/secrets.yaml

# Restore Deployment
kubectl apply -f backup-*/kubernetes-resources.yaml
```

#### 4. Verify Recovery
```bash
# Check deployment status
kubectl get deployment slurm-exporter -n slurm-exporter
kubectl rollout status deployment/slurm-exporter -n slurm-exporter

# Test service health
kubectl port-forward svc/slurm-exporter 10341:10341 -n slurm-exporter &
curl http://localhost:10341/health
curl http://localhost:10341/metrics | head -10
```

### Cross-Region Failover
```bash
# Switch to backup region
export AWS_DEFAULT_REGION=us-east-1

# Download from secondary backup
aws s3 cp "s3://slurm-exporter-backup-secondary/configuration/$BACKUP_FILE" ./

# Deploy to new cluster
kubectl config use-context new-cluster-context
kubectl create namespace slurm-exporter
kubectl apply -f backup-*/
```

## Backup Validation

### Automated Validation
The backup validation job runs daily and performs:
- **Existence Check**: Verifies recent backups exist
- **Integrity Test**: Downloads and tests backup decryption
- **Restoration Test**: Performs test restore operations
- **Retention Check**: Validates retention policy compliance

### Manual Validation
```bash
# Test backup download and decryption
./scripts/validate-backup.sh config-backup-20240101-120000.tar.gz.enc

# Test complete restore procedure (dry-run)
./scripts/test-restore.sh --dry-run --backup=latest

# Verify backup integrity
./scripts/integrity-check.sh --all-backups
```

## Monitoring and Alerting

### Backup Monitoring Metrics
```promql
# Backup job success rate
rate(slurm_backup_jobs_completed_total[24h])

# Backup size trends
slurm_backup_size_bytes

# Backup duration
slurm_backup_duration_seconds

# Validation success rate
rate(slurm_backup_validation_success_total[24h])
```

### Critical Alerts
- **Backup Job Failed**: Any backup job fails
- **Validation Failed**: Backup validation detects issues
- **Missing Backups**: No recent backups found
- **Storage Full**: Backup storage approaching limits

### Notification Channels
- **Slack**: Real-time notifications to #backup-alerts
- **Email**: Daily backup reports to operations team
- **PagerDuty**: Critical backup failures page on-call engineer

## Security Considerations

### Encryption
- **Algorithm**: AES-256-CBC with salt
- **Key Management**: Stored in Kubernetes secrets
- **Key Rotation**: Monthly key rotation recommended
- **Access Control**: Restricted to backup service account

### Access Control
```bash
# Backup service account permissions
kubectl describe serviceaccount slurm-exporter-backup -n slurm-exporter
kubectl describe clusterrole backup-operator
kubectl describe clusterrolebinding backup-operator
```

### Compliance
- **Data Retention**: Compliant with organizational policies
- **Audit Logging**: All backup operations logged
- **Geographic Distribution**: Cross-region compliance
- **Encryption Standards**: Industry-standard encryption

## Cost Optimization

### Storage Classes
- **Standard-IA**: For configuration backups (infrequent access)
- **Glacier**: For long-term retention (certificates)
- **Standard**: For monitoring data (frequent access)

### Lifecycle Policies
```json
{
  "Rules": [{
    "ID": "BackupLifecycle",
    "Status": "Enabled",
    "Transitions": [{
      "Days": 30,
      "StorageClass": "STANDARD_IA"
    }, {
      "Days": 90,
      "StorageClass": "GLACIER"
    }]
  }]
}
```

## Troubleshooting

### Common Issues

#### Backup Job Failures
```bash
# Check job status
kubectl describe job configuration-backup-$(date +%Y%m%d) -n slurm-exporter

# View job logs
kubectl logs job/configuration-backup-$(date +%Y%m%d) -n slurm-exporter

# Check service account permissions
kubectl auth can-i create secrets --as=system:serviceaccount:slurm-exporter:slurm-exporter-backup
```

#### Encryption Issues
```bash
# Test encryption/decryption
echo "test data" | openssl enc -aes-256-cbc -salt -pass pass:"$ENCRYPTION_KEY" | \
  openssl enc -aes-256-cbc -d -salt -pass pass:"$ENCRYPTION_KEY"

# Verify encryption key
kubectl get secret backup-encryption -n slurm-exporter -o jsonpath='{.data.encryption-key}' | base64 -d
```

#### Storage Access Issues
```bash
# Test S3 access
aws s3 ls s3://slurm-exporter-backup-primary/

# Check IAM permissions
aws iam get-user
aws iam list-attached-user-policies --user-name backup-user
```

### Recovery Troubleshooting

#### Failed Restoration
```bash
# Check restoration logs
kubectl logs job/dr-restore-$(date +%s) -n slurm-exporter

# Verify backup integrity
tar -tzf backup.tar.gz

# Test individual component restoration
kubectl apply -f backup-*/configmaps.yaml --dry-run=client
```

#### Health Check Failures
```bash
# Check pod status
kubectl describe pods -n slurm-exporter

# Check service connectivity
kubectl exec deployment/slurm-exporter -n slurm-exporter -- curl -v http://localhost:10341/health

# Check SLURM connectivity
kubectl exec deployment/slurm-exporter -n slurm-exporter -- curl -v "$SLURM_REST_URL/slurm/v0.0.44/ping"
```

## Best Practices

### Backup Best Practices
1. **Regular Testing**: Test restore procedures monthly
2. **Multiple Destinations**: Always maintain multiple backup copies
3. **Encryption**: Encrypt all sensitive backup data
4. **Monitoring**: Monitor backup success and failure rates
5. **Documentation**: Keep backup procedures up to date

### Recovery Best Practices
1. **Practice Drills**: Regular disaster recovery exercises
2. **Time Objectives**: Measure and improve recovery times
3. **Communication**: Clear communication during incidents
4. **Post-Mortem**: Learn from every recovery event
5. **Automation**: Automate common recovery scenarios

### Security Best Practices
1. **Least Privilege**: Minimal required permissions for backup operations
2. **Key Rotation**: Regular encryption key rotation
3. **Access Auditing**: Log and monitor backup access
4. **Compliance**: Regular compliance reviews and audits
5. **Incident Response**: Defined procedures for backup security incidents

This backup and recovery system provides enterprise-grade data protection and disaster recovery capabilities for production SLURM Exporter deployments, ensuring minimal downtime and data loss in various failure scenarios.