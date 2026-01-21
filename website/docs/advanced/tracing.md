# Distributed Tracing

SLURM Exporter includes comprehensive distributed tracing support using OpenTelemetry, enabling detailed observability into metric collection operations and performance analysis.

## Overview

Distributed tracing provides:

- **End-to-end visibility** into metric collection workflows
- **Performance bottleneck identification** in SLURM API calls
- **Detailed timing analysis** of collection operations
- **Error tracking and debugging** across the entire stack
- **Integration** with popular tracing backends (Jaeger, Zipkin, etc.)

## Configuration

### Basic Setup

Enable tracing in your configuration:

```yaml
monitoring:
  tracing:
    enabled: true
    endpoint: "http://jaeger:14268/api/traces"  # OTLP HTTP endpoint
    sample_rate: 0.1                            # Sample 10% of traces
    insecure: true                              # For development only
    
    # Service identification
    service_name: "slurm-exporter"
    service_version: "1.0.0"
    
    # Additional attributes
    attributes:
      environment: "production"
      cluster: "hpc-cluster-1"
      datacenter: "dc1"
```

### Advanced Configuration

```yaml
monitoring:
  tracing:
    enabled: true
    
    # OTLP Configuration
    otlp:
      endpoint: "https://otlp-gateway.example.com:4318"
      headers:
        authorization: "Bearer your-token"
      timeout: 30s
      compression: "gzip"
    
    # Sampling configuration
    sampling:
      type: "probabilistic"        # probabilistic, always_on, always_off, custom
      rate: 0.05                   # 5% sampling rate
      
      # Custom sampling rules
      rules:
        - service_name: "slurm-exporter"
          operation_name: "collect.*"
          sample_rate: 0.2           # Higher rate for collection operations
        - service_name: "slurm-exporter"
          operation_name: "api.*"
          sample_rate: 0.1           # Lower rate for API calls
    
    # Resource attributes
    resource:
      service.name: "slurm-exporter"
      service.version: "1.0.0"
      service.namespace: "monitoring"
      deployment.environment: "production"
      k8s.cluster.name: "prod-k8s"
      k8s.namespace.name: "monitoring"
      k8s.pod.name: "${POD_NAME}"
      k8s.node.name: "${NODE_NAME}"
    
    # Detailed tracing options
    detailed_tracing:
      enabled: false               # Enable for debugging only
      collectors: ["jobs", "nodes"] # Specific collectors to trace in detail
      duration: "10m"              # Auto-disable after duration
      include_payloads: false      # Include API response payloads (sensitive)
```

## Supported Backends

### Jaeger

```yaml
monitoring:
  tracing:
    enabled: true
    endpoint: "http://jaeger-collector:14268/api/traces"
    service_name: "slurm-exporter"
```

Docker Compose setup:
```yaml
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # Jaeger UI
      - "14268:14268"  # HTTP collector
    environment:
      - COLLECTOR_OTLP_ENABLED=true
  
  slurm-exporter:
    image: slurm/exporter:latest
    environment:
      - SLURM_EXPORTER_MONITORING_TRACING_ENABLED=true
      - SLURM_EXPORTER_MONITORING_TRACING_ENDPOINT=http://jaeger:14268/api/traces
```

### Zipkin

```yaml
monitoring:
  tracing:
    enabled: true
    endpoint: "http://zipkin:9411/api/v2/spans"
    format: "zipkin"               # Use Zipkin format
```

### OTLP (OpenTelemetry Protocol)

```yaml
monitoring:
  tracing:
    enabled: true
    endpoint: "http://otel-collector:4318/v1/traces"
    headers:
      x-api-key: "your-api-key"
```

### Cloud Providers

**AWS X-Ray:**
```yaml
monitoring:
  tracing:
    enabled: true
    exporter: "xray"
    aws:
      region: "us-west-2"
      role_arn: "arn:aws:iam::123456789012:role/XRayRole"
```

**Google Cloud Trace:**
```yaml
monitoring:
  tracing:
    enabled: true
    exporter: "google_cloud"
    google_cloud:
      project_id: "your-project-id"
```

**Azure Monitor:**
```yaml
monitoring:
  tracing:
    enabled: true
    exporter: "azure_monitor"
    azure:
      instrumentation_key: "your-instrumentation-key"
```

## Trace Structure

### Collection Traces

Each metric collection generates a trace with the following structure:

```
collect.jobs (root span)
├── api.jobs (SLURM API call)
├── process_job (per job processing)
│   ├── api.job_details (job detail API call)
│   └── generate_job_metrics (metric generation)
└── metrics.jobs (final metric generation)
```

### Span Attributes

**Collection Spans:**
- `collector`: Name of the collector
- `operation`: Type of operation (collection, api_call, metric_generation)
- `timestamp`: Start timestamp
- `duration_ms`: Operation duration
- `success`: Success/failure status
- `error`: Error message (if failed)

**API Call Spans:**
- `api.endpoint`: SLURM API endpoint
- `api.method`: HTTP method
- `api.response_code`: HTTP response code
- `api.response_size`: Response payload size
- `api.retry_count`: Number of retries
- `slurm.version`: SLURM API version

**Metric Generation Spans:**
- `metric.count`: Number of metrics generated
- `metric.cardinality`: Estimated cardinality
- `metric.series_count`: Number of time series
- `metric.label_count`: Number of unique labels

## Advanced Features

### Detailed Tracing

Enable detailed tracing for specific collectors:

```bash
# Via API
curl -X POST http://localhost:9341/debug/tracing/enable \
  -H "Content-Type: application/json" \
  -d '{"collector": "jobs", "duration": "10m"}'

# Via configuration
monitoring:
  tracing:
    detailed_tracing:
      enabled: true
      collectors: ["jobs"]
      duration: "10m"
```

Detailed tracing includes:
- Individual job processing spans
- API request/response payloads (sanitized)
- Memory allocation tracking
- Lock contention analysis
- Database query details

### Sampling Strategies

**Probabilistic Sampling:**
```yaml
monitoring:
  tracing:
    sampling:
      type: "probabilistic"
      rate: 0.1  # 10% of traces
```

**Adaptive Sampling:**
```yaml
monitoring:
  tracing:
    sampling:
      type: "adaptive"
      target_tps: 100        # Target traces per second
      max_tps: 1000         # Maximum traces per second
      
      # Per-operation sampling
      operations:
        "collect.jobs": 0.2
        "collect.nodes": 0.1
        "api.*": 0.05
```

**Error-Based Sampling:**
```yaml
monitoring:
  tracing:
    sampling:
      type: "error_based"
      error_rate: 1.0       # Always sample errors
      success_rate: 0.05    # 5% for successful operations
```

### Trace Correlation

Correlate traces with logs and metrics:

```yaml
logging:
  format: "json"
  fields:
    trace_id: true          # Include trace ID in logs
    span_id: true           # Include span ID in logs

metrics:
  labels:
    trace_sampling: true    # Add sampling info to metrics
```

Example log entry:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "message": "Collection completed",
  "collector": "jobs",
  "duration_ms": 2450,
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "job_count": 156,
  "metrics_generated": 468
}
```

## Observability Integration

### Custom Spans

Add custom spans in collectors:

```go
// In collector implementation
func (c *JobCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
    // Start collection tracing
    ctx, finishCollection := c.tracer.TraceCollection(ctx, "jobs")
    defer finishCollection()
    
    // Add custom attributes
    c.tracer.AddSpanAttribute(ctx, "cluster", c.cluster)
    c.tracer.AddSpanAttribute(ctx, "partition_count", len(c.partitions))
    
    // Trace API calls
    ctx, finishAPI := c.tracer.TraceAPICall(ctx, "jobs", "LIST")
    jobs, err := c.client.ListJobs(ctx)
    finishAPI(err)
    
    if err != nil {
        c.tracer.RecordError(ctx, err)
        return err
    }
    
    // Process with child spans
    for _, job := range jobs {
        ctx, span := c.tracer.CreateChildSpan(ctx, "process_job")
        c.tracer.AddSpanAttribute(ctx, "job.id", job.ID)
        c.tracer.AddSpanAttribute(ctx, "job.state", job.State)
        
        // Process job...
        
        span.End()
    }
    
    return nil
}
```

### Metrics Integration

Export tracing metrics:

```promql
# Trace sampling rate
slurm_exporter_trace_sampling_rate

# Span creation rate
rate(slurm_exporter_spans_created_total[5m])

# Trace export success rate
rate(slurm_exporter_trace_exports_total{result="success"}[5m]) /
rate(slurm_exporter_trace_exports_total[5m])

# Collection duration with tracing overhead
histogram_quantile(0.95, 
  rate(slurm_exporter_collect_duration_seconds_bucket{tracing="enabled"}[5m])
)
```

### Alerting on Traces

Create alerts based on tracing data:

```yaml
groups:
  - name: tracing.rules
    rules:
      - alert: HighTraceErrorRate
        expr: |
          rate(slurm_exporter_trace_errors_total[5m]) /
          rate(slurm_exporter_traces_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High trace error rate detected"
          
      - alert: CollectionLatencyHigh
        expr: |
          histogram_quantile(0.95,
            rate(slurm_exporter_collect_duration_seconds_bucket[5m])
          ) > 30
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Collection latency is high"
```

## Performance Impact

### Overhead Analysis

Tracing adds minimal overhead:

- **CPU overhead**: ~2-5% with 10% sampling
- **Memory overhead**: ~10-20MB for trace buffers  
- **Network overhead**: ~1-5KB per trace
- **Latency impact**: <1ms per traced operation

### Optimization Tips

1. **Use appropriate sampling rates:**
   ```yaml
   # Production
   sample_rate: 0.01  # 1%
   
   # Development  
   sample_rate: 0.1   # 10%
   
   # Debugging
   sample_rate: 1.0   # 100%
   ```

2. **Batch trace exports:**
   ```yaml
   monitoring:
     tracing:
       batch:
         timeout: 5s
         export_timeout: 30s
         max_export_batch_size: 512
   ```

3. **Use resource limits:**
   ```yaml
   monitoring:
     tracing:
       limits:
         max_spans_per_trace: 1000
         max_attributes_per_span: 128
         max_events_per_span: 128
   ```

## Deployment Patterns

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slurm-exporter
spec:
  template:
    spec:
      containers:
      - name: slurm-exporter
        image: slurm/exporter:latest
        env:
        - name: SLURM_EXPORTER_MONITORING_TRACING_ENABLED
          value: "true"
        - name: SLURM_EXPORTER_MONITORING_TRACING_ENDPOINT
          value: "http://jaeger-collector:14268/api/traces"
        - name: K8S_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: K8S_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### High Availability Setup

```yaml
monitoring:
  tracing:
    enabled: true
    
    # Multiple endpoints for HA
    endpoints:
      - "http://jaeger-1:14268/api/traces"
      - "http://jaeger-2:14268/api/traces"
    
    # Load balancing strategy
    load_balancer: "round_robin"  # round_robin, random, least_connections
    
    # Retry configuration
    retry:
      max_attempts: 3
      backoff: "exponential"
      initial_delay: "1s"
      max_delay: "30s"
```

## Troubleshooting

### Common Issues

**No traces appearing:**
```bash
# Check exporter status
curl -s http://localhost:9341/debug/tracing/status

# Verify endpoint connectivity
curl -s http://jaeger:14268/api/traces

# Check sampling configuration
curl -s http://localhost:9341/debug/tracing/config
```

**High memory usage:**
```yaml
# Reduce sampling rate
monitoring:
  tracing:
    sample_rate: 0.01  # 1% instead of 10%

# Enable batch processing
monitoring:
  tracing:
    batch:
      max_export_batch_size: 256
      timeout: 2s
```

**Export failures:**
```bash
# Check exporter logs
journalctl -u slurm-exporter -f | grep -i trace

# Test endpoint manually
curl -X POST http://jaeger:14268/api/traces \
  -H "Content-Type: application/json" \
  -d '{"spans":[{"traceID":"test"}]}'
```

### Debug Commands

```bash
# Enable detailed tracing for 5 minutes
curl -X POST http://localhost:9341/debug/tracing/enable \
  -d '{"collector": "jobs", "duration": "5m"}'

# Get tracing statistics  
curl -s http://localhost:9341/debug/tracing/stats | jq .

# Export current traces
curl -s http://localhost:9341/debug/tracing/export > traces.json

# Reset tracing state
curl -X POST http://localhost:9341/debug/tracing/reset
```

## Best Practices

1. **Start with low sampling rates** in production (1-5%)
2. **Use detailed tracing sparingly** for debugging only
3. **Monitor tracing overhead** and adjust accordingly
4. **Correlate traces with logs and metrics** for full observability
5. **Set up alerting** on trace export failures
6. **Regular cleanup** of old traces to manage storage
7. **Use trace context propagation** for end-to-end visibility

<!-- For more observability features, see the [Performance Monitoring](../user-guide/performance.md) guide. -->