# SLURM Exporter

<div class="grid cards" markdown>

-   :material-rocket-launch-outline:{ .lg .middle } __Enterprise-Grade SLURM Monitoring__

    ---

    A comprehensive Prometheus exporter providing deep insights into SLURM workload manager performance, job analytics, and cluster health with enterprise-grade reliability and scalability.

    [:octicons-arrow-right-24: Getting started](getting-started/index.md)

-   :material-chart-line:{ .lg .middle } __Advanced Job Analytics__

    ---

    Built-in job analytics engine with SLURM-specific timing metrics, resource utilization analysis, efficiency scoring, and waste detection following Unix philosophy principles.

    [:octicons-arrow-right-24: Job Analytics](user-guide/job-analytics.md)

-   :material-kubernetes:{ .lg .middle } __Production Ready__

    ---

    Battle-tested deployment configurations for Kubernetes, Docker, and bare metal with high availability, security hardening, and performance optimization.

    [:octicons-arrow-right-24: Deployment Guide](deployment/index.md)

-   :material-monitor-dashboard:{ .lg .middle } __Rich Observability__

    ---

    Pre-built Grafana dashboards, Prometheus alerts, and comprehensive metrics covering jobs, nodes, partitions, fairshare, QoS, and cluster health.

    [:octicons-arrow-right-24: Metrics Catalog](user-guide/metrics.md)

</div>

## What is SLURM Exporter?

SLURM Exporter is a high-performance Prometheus exporter designed specifically for SLURM (Simple Linux Utility for Resource Management) workload managers. It provides comprehensive monitoring, analytics, and observability for HPC clusters, research computing environments, and large-scale batch processing systems.

### Key Features

#### :material-speedometer: **High Performance**
- Efficient data collection with configurable intervals
- Smart caching and connection pooling
- Low memory footprint with cardinality optimization
- Batch processing for high-frequency metrics

#### :material-chart-analytics: **Advanced Analytics**
- **Job Analytics Engine**: SLURM-specific timing metrics, resource utilization analysis
- **Efficiency Calculator**: Automatic efficiency scoring and waste detection
- **Bottleneck Analyzer**: Real-time identification of cluster bottlenecks
- **Trend Analysis**: Historical performance tracking and forecasting

#### :material-security: **Enterprise Security**
- Multiple authentication methods (JWT, API keys, certificates)
- TLS encryption with certificate validation
- RBAC integration and audit logging
- Security scanning and vulnerability monitoring

#### :material-scale-balance: **Scalability & Reliability**
- Horizontal scaling with multiple exporter instances
- Circuit breaker patterns for fault tolerance
- Graceful degradation under load
- Comprehensive health checks and self-monitoring

#### :material-kubernetes: **Cloud Native**
- Native Kubernetes support with Helm charts
- Container images for multiple architectures
- Service mesh integration (Istio, Linkerd)
- GitOps-friendly configuration management

### Architecture Overview

```mermaid
graph TB
    subgraph "SLURM Cluster"
        SLURM[SLURM REST API]
        SLURMDB[SLURM Database]
    end
    
    subgraph "SLURM Exporter"
        SE[SLURM Exporter]
        CACHE[Intelligent Cache]
        ANALYTICS[Analytics Engine]
    end
    
    subgraph "Monitoring Stack"
        PROM[Prometheus]
        GRAF[Grafana]
        AM[AlertManager]
    end
    
    SLURM --> SE
    SLURMDB --> SE
    SE --> CACHE
    SE --> ANALYTICS
    SE --> PROM
    PROM --> GRAF
    PROM --> AM
```

### Quick Start

Get up and running in minutes:

=== "Docker"

    ```bash
    docker run -d \
      --name slurm-exporter \
      -p 9341:9341 \
      -e SLURM_EXPORTER_SLURM_HOST=your-slurm-host \
      -e SLURM_EXPORTER_SLURM_TOKEN=your-jwt-token \
      slurm/exporter:latest
    ```

=== "Kubernetes"

    ```bash
    helm repo add slurm-exporter https://jontk.github.io/slurm-exporter
    helm repo update
    helm install slurm-exporter slurm-exporter/slurm-exporter \
      --set config.slurm.host=your-slurm-host \
      --set config.slurm.token=your-jwt-token
    ```

=== "Binary"

    ```bash
    # Download latest release
    curl -LO https://github.com/jontk/slurm-exporter/releases/latest/download/slurm-exporter-linux-amd64.tar.gz
    tar xzf slurm-exporter-linux-amd64.tar.gz
    
    # Configure and run
    ./slurm-exporter --config.file=config.yaml
    ```

## Why Choose SLURM Exporter?

### :material-target: **Purpose-Built for SLURM**
Unlike generic exporters, SLURM Exporter is specifically designed for SLURM environments with deep understanding of SLURM concepts like fairshare, QoS policies, and resource allocation patterns.

### :material-database-eye: **Comprehensive Coverage**
Monitor every aspect of your SLURM cluster:
- **Jobs**: Execution, scheduling, resource usage, efficiency
- **Nodes**: Health, utilization, hardware metrics
- **Partitions**: Availability, limits, queue depths
- **Users & Accounts**: Fairshare, quotas, usage patterns
- **System**: Cluster health, performance, capacity planning

### :material-tools: **Production Proven**
Deployed in production environments managing:
- **1000+** compute nodes
- **100,000+** daily jobs
- **Multi-petabyte** storage systems
- **24/7** mission-critical workloads

### :material-heart-handshake: **Community Driven**
Active open-source community with regular releases, comprehensive documentation, and responsive support from HPC professionals worldwide.

---

## Next Steps

<div class="grid cards" markdown>

-   :material-play-circle:{ .lg .middle } **Try it now**

    ---

    Follow our quick start guide to get SLURM Exporter running in your environment in under 5 minutes.

    [:octicons-arrow-right-24: Quick Start](getting-started/quickstart.md)

-   :material-book-open:{ .lg .middle } **Learn more**

    ---

    Explore our comprehensive user guide covering configuration, metrics, and best practices.

    [:octicons-arrow-right-24: User Guide](user-guide/index.md)

-   :material-github:{ .lg .middle } **Contribute**

    ---

    Join our community! Report bugs, request features, or contribute code to make SLURM Exporter even better.

    [:octicons-arrow-right-24: Contributing](development/contributing.md)

-   :material-chat:{ .lg .middle } **Get help**

    ---

    Need support? Check our troubleshooting guide or connect with the community.

    [:octicons-arrow-right-24: Troubleshooting](user-guide/troubleshooting.md)

</div>