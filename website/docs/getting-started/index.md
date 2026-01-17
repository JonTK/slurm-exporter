# Getting Started

Welcome to SLURM Exporter! This guide will help you get up and running quickly with monitoring your SLURM cluster.

## Overview

SLURM Exporter provides comprehensive monitoring for SLURM workload managers by exposing metrics in Prometheus format. Whether you're running a small research cluster or a large-scale HPC facility, SLURM Exporter scales to meet your monitoring needs.

## Prerequisites

Before installing SLURM Exporter, ensure you have:

### SLURM Environment
- **SLURM version**: 20.02+ (recommended: 22.05+)
- **SLURM REST API**: Enabled and accessible
- **Database access**: Read access to SLURM database (optional but recommended)

### Infrastructure
- **Operating System**: Linux (RHEL/CentOS 7+, Ubuntu 18.04+, SLES 15+)
- **Architecture**: x86_64, arm64
- **Memory**: Minimum 512MB RAM (recommended: 2GB+)
- **CPU**: 1 core minimum (recommended: 2+ cores)
- **Network**: HTTP/HTTPS access to SLURM REST API

### Monitoring Stack
- **Prometheus**: 2.30+ for metric collection
- **Grafana**: 8.0+ for visualization (optional)
- **AlertManager**: For alerting (optional)

## Installation Methods

Choose the installation method that best fits your environment:

### :material-kubernetes: Kubernetes (Recommended)
Perfect for containerized environments with automatic scaling and management.

[→ Kubernetes Installation](installation.md#kubernetes){ .md-button .md-button--primary }

### :material-docker: Docker
Quick setup for testing and development environments.

[→ Docker Installation](installation.md#docker){ .md-button }

### :material-package-variant: Package Manager
System packages for RHEL/CentOS, Ubuntu, and SUSE distributions.

[→ Package Installation](installation.md#packages){ .md-button }

### :material-hammer-wrench: From Source
Build from source for custom configurations or development.

[→ Build from Source](installation.md#source){ .md-button }

## Quick Start Paths

### :material-rocket-launch: 5-Minute Setup

Get basic monitoring running immediately:

1. **Deploy** with single command
2. **Configure** SLURM connection
3. **Verify** metrics collection
4. **View** in Prometheus

[→ Quick Start Guide](quickstart.md){ .md-button .md-button--primary }

### :material-cog: Production Setup

Complete production deployment with high availability:

1. **Plan** your architecture
2. **Configure** for production
3. **Deploy** with redundancy
4. **Monitor** the monitoring

[→ Production Guide](../deployment/production.md){ .md-button }

## Configuration Overview

SLURM Exporter uses a flexible YAML configuration system:

```yaml title="Basic Configuration"
# SLURM connection settings
slurm:
  host: "slurm-controller.example.com"
  port: 6820
  auth:
    type: "jwt"
    token: "your-jwt-token"

# Metric collection settings
metrics:
  enabled_collectors:
    - jobs
    - nodes
    - partitions
    - accounts
  collection_interval: 30s

# Server settings
server:
  port: 9341
  path: "/metrics"
```

[→ Full Configuration Reference](configuration.md){ .md-button }

## Key Concepts

### Collectors
Collectors gather specific types of metrics:
- **Jobs**: Job execution, queuing, resource usage
- **Nodes**: Node health, utilization, hardware
- **Partitions**: Queue status, limits, availability
- **Accounts**: User activity, fairshare, quotas

### Job Analytics
Advanced analytics engine providing:
- **Efficiency Scoring**: Automatic resource efficiency calculation
- **Waste Detection**: Identification of underutilized resources
- **Bottleneck Analysis**: Real-time cluster bottleneck identification
- **Trend Analysis**: Historical performance tracking

### Performance Features
- **Smart Caching**: Reduces SLURM API load
- **Batch Processing**: Efficient bulk metric collection
- **Connection Pooling**: Optimized network usage
- **Cardinality Control**: Prevents metric explosion

## Architecture Patterns

### Single Exporter
Simple deployment for small to medium clusters:

```mermaid
graph LR
    SE[SLURM Exporter] --> SLURM[SLURM REST API]
    SE --> PROM[Prometheus]
```

### High Availability
Redundant deployment for production environments:

```mermaid
graph TB
    LB[Load Balancer] --> SE1[SLURM Exporter 1]
    LB --> SE2[SLURM Exporter 2]
    SE1 --> SLURM[SLURM REST API]
    SE2 --> SLURM
    PROM[Prometheus] --> LB
```

### Multi-Cluster
Federated monitoring across multiple SLURM clusters:

```mermaid
graph TB
    PROM[Prometheus Federation]
    PROM --> SE1[Cluster A Exporter]
    PROM --> SE2[Cluster B Exporter]
    SE1 --> SLURM1[SLURM Cluster A]
    SE2 --> SLURM2[SLURM Cluster B]
```

## What's Next?

Choose your next step based on your needs:

<div class="grid cards" markdown>

-   :material-timer-outline:{ .lg .middle } **New to SLURM Exporter?**

    ---

    Start with our quick start guide to get basic monitoring running in 5 minutes.

    [:octicons-arrow-right-24: Quick Start](quickstart.md)

-   :material-server:{ .lg .middle } **Production Deployment?**

    ---

    Jump to our production deployment guide for enterprise-grade setup.

    [:octicons-arrow-right-24: Production Guide](../deployment/production.md)

-   :material-chart-line:{ .lg .middle } **Advanced Features?**

    ---

    Explore job analytics, custom metrics, and advanced configuration options.

    [:octicons-arrow-right-24: Advanced Features](../advanced/index.md)

-   :material-help-circle:{ .lg .middle } **Need Help?**

    ---

    Check our troubleshooting guide or connect with the community.

    [:octicons-arrow-right-24: Troubleshooting](../user-guide/troubleshooting.md)

</div>

## Community & Support

- **GitHub**: [Report issues and contribute](https://github.com/jontk/slurm-exporter)
- **Documentation**: You're reading it! 📚
- **Community**: Join discussions on SLURM community forums

---

Ready to start monitoring your SLURM cluster? Let's begin with the [installation guide](installation.md)!