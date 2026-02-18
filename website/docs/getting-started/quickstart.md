# Quick Start

Get SLURM Exporter running in your environment in under 5 minutes! This guide provides the fastest path to basic SLURM monitoring.

## Prerequisites Check

Before starting, ensure you have:

- [ ] SLURM cluster with REST API enabled (slurmrestd)
- [ ] Access credentials (JWT token)
- [ ] Network connectivity to SLURM controller
- [ ] Docker or Linux system for deployment

## Step 1: Create a Configuration File

Create a `config.yaml` file with your SLURM connection details:

```yaml title="config.yaml"
server:
  address: ":10341"
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"

slurm:
  base_url: "http://your-slurm-controller.example.com:6820"
  api_version: "v0.0.44"
  auth:
    type: "jwt"
    username: "root"
    token: "your-jwt-token"
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s

collectors:
  jobs:
    enabled: true
  nodes:
    enabled: true
  partitions:
    enabled: true
  cluster:
    enabled: true

logging:
  level: "info"
  format: "json"

validation:
  allow_insecure_connections: true
```

## Step 2: Choose Your Deployment Method

=== "Docker (Fastest)"

    Perfect for development and testing:

    ```bash
    docker run -d \
      --name slurm-exporter \
      -p 10341:10341 \
      -v $(pwd)/config.yaml:/etc/slurm-exporter/config.yaml:ro \
      ghcr.io/jontk/slurm-exporter:latest
    ```

=== "Binary"

    Direct installation on Linux:

    ```bash
    # Download and install
    curl -LO https://github.com/jontk/slurm-exporter/releases/latest/download/slurm-exporter-linux-amd64.tar.gz
    tar xzf slurm-exporter-linux-amd64.tar.gz
    chmod +x slurm-exporter

    # Configure and run
    ./slurm-exporter --config=config.yaml
    ```

## Step 3: Verify Installation

Check that SLURM Exporter is running and collecting metrics:

```bash
# Check health endpoint
curl http://localhost:10341/health
```

```bash
# Check readiness endpoint
curl http://localhost:10341/ready
```

```bash
# Check metrics endpoint
curl http://localhost:10341/metrics | head -20

# Expected output should include SLURM metrics like:
# slurm_exporter_... metrics
```

## Step 4: Test Basic Functionality

Verify key metrics are being collected:

```bash
# Check job metrics
curl -s http://localhost:10341/metrics | grep slurm_job

# Check node metrics
curl -s http://localhost:10341/metrics | grep slurm_node

# Check partition metrics
curl -s http://localhost:10341/metrics | grep slurm_partition
```

## Step 5: Quick Prometheus Setup

Set up basic Prometheus monitoring:

=== "Docker Compose"

    Create `docker-compose.yml`:

    ```yaml
    version: '3.8'
    services:
      slurm-exporter:
        image: ghcr.io/jontk/slurm-exporter:latest
        ports:
          - "10341:10341"
        volumes:
          - ./config.yaml:/etc/slurm-exporter/config.yaml:ro

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
          - targets: ['slurm-exporter:10341']
        scrape_interval: 30s
        metrics_path: /metrics
    ```

    Start the stack:

    ```bash
    docker-compose up -d
    ```

=== "Standalone Prometheus"

    Add to your existing `prometheus.yml`:

    ```yaml
    scrape_configs:
      - job_name: 'slurm-exporter'
        static_configs:
          - targets: ['localhost:10341']
        scrape_interval: 30s
        metrics_path: /metrics
    ```

## Step 6: View Your Metrics

Access Prometheus to see your SLURM metrics:

1. **Open Prometheus**: http://localhost:9090
2. **Go to Graph tab**
3. **Try these queries**:

```promql
# Check exporter is up
up{job="slurm-exporter"}

# Node utilization percentage
(slurm_partition_cpus_allocated / slurm_partition_cpus_total) * 100

# Job completion rate (5-minute average)
rate(slurm_jobs_completed_total[5m])
```

## What's Working Now?

After completing these steps, you have:

- [x] **SLURM Exporter** collecting basic metrics
- [x] **Prometheus** scraping and storing metrics
- [x] **Core monitoring** of jobs, nodes, and partitions
- [x] **HTTP endpoints** for health checks and metrics (`/health`, `/ready`, `/metrics`)
- [x] **Foundation** for alerting and dashboards

## Common Quick Fixes

### Connection Issues

```bash
# Test SLURM API connectivity
curl -H "X-SLURM-USER-TOKEN: your-jwt-token" \
     http://your-slurm-controller.example.com:6820/slurm/v0.0.44/ping
```

### No Metrics

```bash
# Check exporter logs
docker logs slurm-exporter

# Enable debug logging via config file
# Set logging.level to "debug" in config.yaml and restart
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

### **Add Dashboards**
Import pre-built Grafana dashboards for visualization:

[-> Grafana Integration Guide](../integration/grafana.md){ .md-button }

### **Set Up Alerting**
Configure alerts for job failures and node issues:

[-> Alerting Setup Guide](../user-guide/alerting.md){ .md-button }

### **Full Configuration**
Explore the complete configuration reference to tune the exporter for your environment:

[-> Configuration Reference](../user-guide/configuration.md){ .md-button }

### **Production Deployment**
Scale up for production with high availability:

<!-- [-> Production Guide](../deployment/production.md){ .md-button } -->

## Quick Reference

### Default Ports
- **SLURM Exporter**: 10341
- **Prometheus**: 9090
- **Grafana**: 3000
- **SLURM REST API**: 6820

### Key Endpoints
- **Metrics**: `http://localhost:10341/metrics`
- **Health**: `http://localhost:10341/health`
- **Ready**: `http://localhost:10341/ready`

### CLI Flags
```bash
slurm-exporter --help

  --config         Path to configuration file (default: "configs/config.yaml")
  --log-level      Log level: debug, info, warn, error (default: "info")
  --addr           Listen address (default: ":10341")
  --metrics-path   Metrics endpoint path (default: "/metrics")
  --version        Show version information and exit
  --health-check   Perform health check and exit
```

### Configuration Files
- **Docker**: Mount config file to `/etc/slurm-exporter/config.yaml`
- **Binary**: Pass with `--config=/path/to/config.yaml`
- **Package**: `/etc/slurm-exporter/config.yaml`

---

**Congratulations!** You now have SLURM monitoring up and running.

Need help? Check our [Troubleshooting Guide](../user-guide/troubleshooting.md) or explore [Job Analytics](../user-guide/job-analytics.md) for advanced features.
