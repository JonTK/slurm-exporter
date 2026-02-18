# Performance Validation Framework

This directory contains a comprehensive performance validation framework for SLURM Exporter production deployments, including load testing, benchmarking, capacity planning, and performance regression detection.

## Overview

The performance validation framework provides:
- **Load Testing**: Automated testing scenarios for baseline, stress, spike, and endurance
- **Benchmarking**: Comprehensive performance benchmarks for memory, CPU, network, and scalability
- **Capacity Planning**: Predictive analysis and resource scaling recommendations
- **Regression Detection**: Automated detection of performance degradation
- **SLA Validation**: Continuous validation of service level agreements

## Framework Components

### 1. Load Testing (`load-testing.yaml`)

#### Test Scenarios
- **Baseline Test**: Normal operational load (10 VUs, 10 minutes)
- **Stress Test**: High load testing (100 VUs, 15 minutes)
- **Spike Test**: Sudden load spikes (200 VUs, 5 minutes)
- **Endurance Test**: Long-running stability (50 VUs, 60 minutes)
- **Capacity Test**: Maximum capacity determination (500 VUs, 20 minutes)

#### SLA Thresholds
- **Response Time P95**: < 5 seconds
- **Response Time P99**: < 10 seconds
- **Error Rate**: < 1%
- **Availability**: > 99.9%
- **Throughput**: > 45 RPS (baseline)

#### Test Execution
```bash
# Run baseline load test
kubectl create job --from=job/load-test-baseline \
  baseline-test-$(date +%s) -n slurm-exporter

# Run stress test
kubectl create job --from=job/load-test-stress \
  stress-test-$(date +%s) -n slurm-exporter

# Monitor test progress
kubectl logs -f job/baseline-test-$(date +%s) -n slurm-exporter
```

### 2. Performance Benchmarking (`benchmarking.yaml`)

#### Benchmark Categories

##### Memory Benchmarks
- Memory allocation performance
- JSON serialization/deserialization
- Garbage collection impact
- Concurrent memory access

##### CPU Benchmarks
- Metric processing performance
- Regex matching efficiency
- Concurrent collection overhead
- JSON parsing performance

##### Network Benchmarks
- SLURM API latency
- HTTP throughput
- Connection pooling efficiency
- Concurrent request handling

##### Scalability Benchmarks
- Large cluster simulation
- High cardinality metrics
- Concurrent scraper performance
- Memory usage under load

#### Running Benchmarks
```bash
# Run memory benchmarks
kubectl create job --from=job/benchmark-memory \
  memory-bench-$(date +%s) -n slurm-exporter

# Run scalability benchmarks
kubectl create job --from=job/benchmark-scalability \
  scalability-bench-$(date +%s) -n slurm-exporter

# View results
kubectl logs job/memory-bench-$(date +%s) -n slurm-exporter
```

### 3. Capacity Planning (`capacity-planning.yaml`)

#### Resource Prediction Models
- **CPU Utilization**: 30-day prediction with 15% growth threshold
- **Memory Utilization**: 30-day prediction with 20% growth threshold
- **Cardinality Growth**: 60-day prediction with 25% growth threshold
- **Request Volume**: 45-day prediction with 30% growth threshold
- **Storage Usage**: 90-day prediction with 10% growth threshold

#### Scaling Recommendations
```yaml
cpu:
  current_limit: "1000m"
  recommended_increment: "500m"
  max_limit: "4000m"

memory:
  current_limit: "1Gi"
  recommended_increment: "512Mi"
  max_limit: "8Gi"

replicas:
  current_min: 3
  current_max: 10
  recommended_max: 20
```

#### Capacity Analysis
```bash
# Run monthly capacity analysis
kubectl create job --from=cronjob/capacity-analysis \
  capacity-analysis-$(date +%s) -n slurm-exporter

# Run cluster growth simulation
kubectl create job --from=job/capacity-simulation \
  capacity-sim-$(date +%s) -n slurm-exporter
```

### 4. Performance Regression Detection

#### Automated Weekly Validation
- **Schedule**: Every Sunday at 1 AM
- **Comparison**: Current week vs. previous week
- **Threshold**: 20% performance degradation
- **Metrics Monitored**:
  - Response time P95/P99
  - Error rates
  - Memory usage
  - CPU utilization
  - GC duration

#### Regression Detection
```bash
# Check regression detection job
kubectl get cronjob performance-regression-detection -n slurm-exporter

# View latest regression analysis
kubectl logs job/performance-regression-detection-$(date +%Y%m%d) -n slurm-exporter
```

## Performance Baselines

### Current Performance Baselines
```yaml
memory_allocation:
  metric: "allocations_per_second"
  baseline: 10000
  tolerance: 10%
  unit: "allocs/sec"

metric_processing:
  metric: "metrics_processed_per_second"
  baseline: 1000
  tolerance: 15%
  unit: "metrics/sec"

slurm_api_latency:
  metric: "api_response_time_p95"
  baseline: 100
  tolerance: 20%
  unit: "ms"

gc_pressure:
  metric: "gc_pause_time_p99"
  baseline: 10
  tolerance: 25%
  unit: "ms"
```

### Performance Targets by Cluster Size

| Cluster Size | Response Time P95 | Throughput | Memory Usage | Error Rate |
|--------------|-------------------|------------|--------------|------------|
| Small (< 1K nodes) | < 2s | > 100 RPS | < 512Mi | < 0.5% |
| Medium (1K-5K nodes) | < 5s | > 50 RPS | < 1Gi | < 1% |
| Large (> 5K nodes) | < 10s | > 25 RPS | < 2Gi | < 2% |

## Usage Instructions

### Setup Performance Testing Environment

#### 1. Deploy Performance Testing Infrastructure
```bash
# Apply all performance validation components
kubectl apply -f load-testing.yaml
kubectl apply -f benchmarking.yaml
kubectl apply -f capacity-planning.yaml

# Create required secrets
kubectl create secret generic performance-notifications \
  --from-literal=slack-webhook="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
  -n slurm-exporter

kubectl create secret generic capacity-notifications \
  --from-literal=slack-webhook="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK" \
  -n slurm-exporter
```

#### 2. Verify Setup
```bash
# Check CronJobs are scheduled
kubectl get cronjobs -n slurm-exporter

# Verify configuration
kubectl get configmap load-test-config -n slurm-exporter -o yaml
kubectl get configmap benchmark-config -n slurm-exporter -o yaml
kubectl get configmap capacity-planning-config -n slurm-exporter -o yaml
```

### Running Performance Tests

#### Complete Performance Validation Suite
```bash
#!/bin/bash
# Complete performance validation script

set -e

echo "Starting comprehensive performance validation"

# 1. Run baseline load test
echo "Running baseline load test..."
kubectl create job --from=job/load-test-baseline \
  baseline-$(date +%s) -n slurm-exporter

# Wait for completion
kubectl wait --for=condition=complete job/baseline-$(date +%s) \
  -n slurm-exporter --timeout=900s

# 2. Run memory benchmarks
echo "Running memory benchmarks..."
kubectl create job --from=job/benchmark-memory \
  memory-bench-$(date +%s) -n slurm-exporter

# 3. Run capacity analysis
echo "Running capacity analysis..."
kubectl create job --from=cronjob/capacity-analysis \
  capacity-analysis-$(date +%s) -n slurm-exporter

echo "Performance validation suite started"
echo "Monitor progress with: kubectl get jobs -n slurm-exporter"
```

#### Stress Testing for Production Readiness
```bash
#!/bin/bash
# Production readiness stress test

# Run stress test
kubectl create job --from=job/load-test-stress \
  stress-$(date +%s) -n slurm-exporter

# Run scalability benchmarks
kubectl create job --from=job/benchmark-scalability \
  scalability-$(date +%s) -n slurm-exporter

# Monitor resource usage during tests
kubectl top pods -n slurm-exporter --watch
```

### Performance Monitoring Dashboard

#### Key Metrics to Monitor
```promql
# Response time trends
histogram_quantile(0.95, rate(slurm_exporter_scrape_duration_seconds_bucket[5m]))

# Throughput trends
rate(prometheus_http_requests_total{job="slurm-exporter"}[5m])

# Error rate trends
rate(slurm_exporter_collection_errors_total[5m]) / rate(slurm_exporter_collections_total[5m])

# Resource utilization
container_cpu_usage_seconds_total{pod=~"slurm-exporter-.*"}
container_memory_working_set_bytes{pod=~"slurm-exporter-.*"}

# Cardinality trends
slurm_exporter_cardinality_current
```

#### Custom Grafana Dashboard
```json
{
  "dashboard": {
    "title": "SLURM Exporter Performance Validation",
    "panels": [
      {
        "title": "Load Test Results",
        "type": "stat",
        "targets": [
          {
            "expr": "slurm_load_test_response_time_p95",
            "legendFormat": "Response Time P95"
          }
        ]
      },
      {
        "title": "Benchmark Trends",
        "type": "timeseries",
        "targets": [
          {
            "expr": "slurm_benchmark_memory_allocs_per_second",
            "legendFormat": "Memory Allocations/sec"
          }
        ]
      }
    ]
  }
}
```

### Capacity Planning Workflow

#### Monthly Capacity Review Process
1. **Automated Analysis**: Monthly CronJob runs capacity analysis
2. **Review Report**: Check Slack notifications for recommendations
3. **Resource Planning**: Review scaling recommendations
4. **Implementation**: Apply recommended scaling changes
5. **Validation**: Run load tests to validate changes

#### Growth Scenario Planning
```bash
# Run different growth scenarios
kubectl create job --from=job/capacity-simulation \
  growth-conservative-$(date +%s) -n slurm-exporter

# Analyze results
kubectl logs job/growth-conservative-$(date +%s) -n slurm-exporter | \
  jq '.recommendations'
```

### Performance Regression Handling

#### When Regression is Detected
1. **Immediate Assessment**: Check regression severity and impact
2. **Root Cause Analysis**: Compare metrics with previous baseline
3. **Mitigation**: Apply immediate fixes or rollback if necessary
4. **Long-term Fix**: Implement performance optimizations
5. **Baseline Update**: Update performance baselines if changes are intentional

#### Manual Regression Testing
```bash
# Run regression detection manually
kubectl create job --from=cronjob/performance-regression-detection \
  regression-check-$(date +%s) -n slurm-exporter

# Compare specific metrics
kubectl exec prometheus-0 -n monitoring -- \
  promtool query instant 'rate(slurm_exporter_scrape_duration_seconds[1w] offset 1w)'
```

## Troubleshooting

### Common Performance Issues

#### High Response Times
```bash
# Check current response times
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  curl -w "%{time_total}" http://localhost:10341/metrics

# Check resource constraints
kubectl top pods -n slurm-exporter
kubectl describe pods -n slurm-exporter

# Review SLURM API performance
kubectl exec deployment/slurm-exporter -n slurm-exporter -- \
  curl -w "%{time_total}" "$SLURM_REST_URL/slurm/v0.0.44/ping"
```

#### Memory Issues
```bash
# Check memory usage trends
kubectl exec prometheus-0 -n monitoring -- \
  promtool query instant 'container_memory_working_set_bytes{pod=~"slurm-exporter-.*"}'

# Enable memory profiling
kubectl port-forward deployment/slurm-exporter 6060:6060 -n slurm-exporter
go tool pprof http://localhost:6060/debug/pprof/heap
```

#### Load Test Failures
```bash
# Check test job logs
kubectl logs job/baseline-$(date +%s) -n slurm-exporter

# Verify target service health
kubectl get endpoints slurm-exporter -n slurm-exporter
curl http://slurm-exporter.slurm-exporter:10341/health

# Check resource limits for test jobs
kubectl describe job/baseline-$(date +%s) -n slurm-exporter
```

### Performance Optimization

#### Immediate Optimizations
```bash
# Increase resource limits
kubectl patch deployment slurm-exporter -n slurm-exporter -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "slurm-exporter",
          "resources": {
            "limits": {
              "cpu": "2000m",
              "memory": "2Gi"
            }
          }
        }]
      }
    }
  }
}'

# Scale horizontally
kubectl scale deployment slurm-exporter --replicas=5 -n slurm-exporter

# Enable performance features
kubectl patch configmap slurm-exporter-config -n slurm-exporter -p '
{
  "data": {
    "config.yaml": "updated-config-with-performance-optimizations"
  }
}'
```

#### Long-term Optimizations
1. **Cardinality Management**: Implement metric filtering and sampling
2. **Connection Pooling**: Enable SLURM API connection pooling
3. **Caching**: Implement intelligent caching strategies
4. **Collection Optimization**: Adjust collection intervals based on data volatility
5. **Resource Tuning**: Optimize GC settings and memory allocation

## Integration with CI/CD

### Pre-deployment Performance Testing
```yaml
# .github/workflows/performance-test.yml
name: Performance Testing
on:
  pull_request:
    paths: ['internal/**', 'cmd/**']

jobs:
  performance-test:
    runs-on: ubuntu-latest
    steps:
    - name: Run Load Test
      run: |
        kubectl apply -f deployment/production/performance/load-testing.yaml
        kubectl create job --from=job/load-test-baseline baseline-ci-${{ github.run_id }}
        kubectl wait --for=condition=complete job/baseline-ci-${{ github.run_id }} --timeout=600s
    
    - name: Check Performance SLA
      run: |
        kubectl logs job/baseline-ci-${{ github.run_id }} | grep -q "✅ LOAD TEST PASSED"
```

### Performance Gating
```bash
# Performance gate script for deployments
#!/bin/bash

# Run baseline test
kubectl create job --from=job/load-test-baseline baseline-gate-$(date +%s)
kubectl wait --for=condition=complete job/baseline-gate-$(date +%s) --timeout=600s

# Check if SLA passed
if kubectl logs job/baseline-gate-$(date +%s) | grep -q "❌ LOAD TEST FAILED"; then
  echo "Performance gate failed - deployment blocked"
  exit 1
else
  echo "Performance gate passed - deployment approved"
fi
```

This performance validation framework provides comprehensive testing, monitoring, and optimization capabilities to ensure SLURM Exporter maintains excellent performance characteristics in production environments at any scale.