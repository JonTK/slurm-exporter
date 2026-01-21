# Quick Start

Get SLURM Exporter running in your environment in under 5 minutes! This guide provides the fastest path to basic SLURM monitoring.

## Prerequisites Check

Before starting, ensure you have:

- [ ] SLURM cluster with REST API enabled
- [ ] Access credentials (JWT token or user/password)
- [ ] Network connectivity to SLURM controller
- [ ] Docker, Kubernetes, or Linux system for deployment

## Step 1: Choose Your Deployment Method

=== "Docker (Fastest)"

    Perfect for development and testing:

    ```bash
    docker run -d \
      --name slurm-exporter \
      -p 9341:9341 \
      -e SLURM_EXPORTER_SLURM_HOST=your-slurm-controller.example.com \
      -e SLURM_EXPORTER_SLURM_TOKEN=your-jwt-token \
      slurm/exporter:latest
    ```

=== "Kubernetes"

    For production environments:

    ```bash
    helm repo add slurm-exporter https://jontk.github.io/slurm-exporter
    helm install slurm-exporter slurm-exporter/slurm-exporter \
      --set config.slurm.host=your-slurm-controller.example.com \
      --set config.slurm.token=your-jwt-token
    ```

=== "Binary"

    Direct installation on Linux:

    ```bash
    # Download and install
    curl -LO https://github.com/jontk/slurm-exporter/releases/latest/download/slurm-exporter-linux-amd64.tar.gz
    tar xzf slurm-exporter-linux-amd64.tar.gz
    chmod +x slurm-exporter
    
    # Configure and run
    ./slurm-exporter \
      --slurm.host=your-slurm-controller.example.com \
      --slurm.token=your-jwt-token
    ```

## Step 2: Verify Installation

Check that SLURM Exporter is running and collecting metrics:

```bash
# Check health endpoint
curl http://localhost:9341/health

# Expected response:
# {"status":"ok","timestamp":"2024-01-15T10:30:00Z","version":"1.0.0"}
```

```bash
# Check metrics endpoint
curl http://localhost:9341/metrics | head -20

# Expected output should include SLURM metrics like:
# slurm_up 1
# slurm_jobs_pending 42
# slurm_nodes_total 128
```

## Step 3: Test Basic Functionality

Verify key metrics are being collected:

```bash
# Check job metrics
curl -s http://localhost:9341/metrics | grep slurm_jobs

# Check node metrics  
curl -s http://localhost:9341/metrics | grep slurm_nodes

# Check partition metrics
curl -s http://localhost:9341/metrics | grep slurm_partitions
```

Expected metrics include:
- `slurm_jobs_pending` - Jobs waiting to run
- `slurm_jobs_running` - Currently executing jobs
- `slurm_nodes_idle` - Available compute nodes
- `slurm_nodes_down` - Unavailable nodes

## Step 4: Quick Prometheus Setup

Set up basic Prometheus monitoring:

=== "Docker Compose"

    Create `docker-compose.yml`:

    ```yaml
    version: '3.8'
    services:
      slurm-exporter:
        image: slurm/exporter:latest
        ports:
          - "9341:9341"
        environment:
          SLURM_EXPORTER_SLURM_HOST: your-slurm-controller.example.com
          SLURM_EXPORTER_SLURM_TOKEN: your-jwt-token

      prometheus:
        image: prom/prometheus:latest
        ports:
          - "9090:9090"
        volumes:
          - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ```

    Create `prometheus.yml`:

    ```yaml
    global:
      scrape_interval: 30s

    scrape_configs:
      - job_name: 'slurm-exporter'
        static_configs:
          - targets: ['slurm-exporter:9341']
        scrape_interval: 30s
        metrics_path: /metrics
    ```

    Start the stack:

    ```bash
    docker-compose up -d
    ```

=== "Kubernetes"

    Apply Prometheus configuration:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: prometheus-config
    data:
      prometheus.yml: |
        global:
          scrape_interval: 30s
        scrape_configs:
          - job_name: 'slurm-exporter'
            kubernetes_sd_configs:
            - role: endpoints
            relabel_configs:
            - source_labels: [__meta_kubernetes_service_name]
              action: keep
              regex: slurm-exporter
    ```

## Step 5: View Your Metrics

Access Prometheus to see your SLURM metrics:

1. **Open Prometheus**: http://localhost:9090
2. **Go to Graph tab**
3. **Try these queries**:

```promql
# Total jobs in the system
sum(slurm_jobs_pending + slurm_jobs_running)

# Node utilization percentage
(slurm_nodes_allocated / slurm_nodes_total) * 100

# Job completion rate (5-minute average)
rate(slurm_jobs_completed_total[5m])
```

## What's Working Now? ✅

After completing these steps, you have:

- [x] **SLURM Exporter** collecting basic metrics
- [x] **Prometheus** scraping and storing metrics
- [x] **Core monitoring** of jobs, nodes, and partitions
- [x] **HTTP endpoints** for health checks and metrics
- [x] **Foundation** for alerting and dashboards

## Common Quick Fixes

### Connection Issues

```bash
# Test SLURM API connectivity
curl -H "X-SLURM-USER-TOKEN: your-jwt-token" \
     http://your-slurm-controller.example.com:6820/slurm/v0.0.40/ping

# Should return: {"ping": "UP"}
```

### No Metrics

```bash
# Check exporter logs
docker logs slurm-exporter

# Enable debug logging
docker run --environment SLURM_EXPORTER_LOG_LEVEL=debug slurm/exporter:latest
```

### Permission Denied

```bash
# Verify token has correct permissions
scontrol show config | grep AccountingStorageType

# Test with sinfo command
sinfo --format="%P %.5a %.10l %.6D %.6t %N"
```

## Next Steps

Now that you have basic monitoring working, consider these enhancements:

### 📊 **Add Dashboards**
Import pre-built Grafana dashboards for visualization:

[→ Grafana Integration Guide](../integration/grafana.md){ .md-button }

### 🚨 **Set Up Alerting**  
Configure alerts for job failures and node issues:

[→ Alerting Setup Guide](../user-guide/alerting.md){ .md-button }

### ⚡ **Enable Advanced Features**
Turn on job analytics, efficiency monitoring, and more:

[→ Distributed Tracing](../advanced/tracing.md){ .md-button }

### 🏭 **Production Deployment**
Scale up for production with high availability:

<!-- [→ Production Guide](../deployment/production.md){ .md-button } -->

## Quick Reference

### Default Ports
- **SLURM Exporter**: 9341
- **Prometheus**: 9090
- **Grafana**: 3000
- **SLURM REST API**: 6820

### Key Endpoints
- **Metrics**: `http://localhost:9341/metrics`
- **Health**: `http://localhost:9341/health`
- **Debug**: `http://localhost:9341/debug/vars`

### Essential Metrics
```promql
# System health
slurm_up

# Workload overview  
slurm_jobs_pending
slurm_jobs_running
slurm_jobs_completed_total

# Resource utilization
slurm_nodes_idle
slurm_nodes_allocated
slurm_nodes_down

# Partition status
slurm_partition_jobs_pending
slurm_partition_nodes_idle
```

### Configuration Files
- **Docker**: Environment variables
- **Kubernetes**: Helm values
- **Binary**: `/etc/slurm-exporter/config.yaml`

---

🎉 **Congratulations!** You now have SLURM monitoring up and running. 

Need help? Check our [Troubleshooting Guide](../user-guide/troubleshooting.md) or explore [Job Analytics](../user-guide/job-analytics.md) for advanced features.