# SLURM Prometheus Exporter

[![Go Reference](https://pkg.go.dev/badge/github.com/jontk/slurm-exporter.svg)](https://pkg.go.dev/github.com/jontk/slurm-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/jontk/slurm-exporter)](https://goreportcard.com/report/github.com/jontk/slurm-exporter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive Prometheus exporter for SLURM (Simple Linux Utility for Resource Management) clusters, providing extensive monitoring capabilities for HPC environments.

## Overview

The SLURM Prometheus Exporter is designed to provide complete visibility into SLURM cluster operations, including:

- **Cluster Health**: Real-time monitoring of cluster status, node states, and resource utilization
- **Job Analytics**: Comprehensive job lifecycle tracking, queue analysis, and performance metrics
- **User & Account Monitoring**: Usage tracking, quota monitoring, and fair-share analysis
- **Partition Management**: Partition configuration and utilization statistics
- **Performance Insights**: Efficiency metrics, throughput analysis, and bottleneck identification

## Features

### 🔍 Comprehensive Monitoring
- **50+ Prometheus metrics** covering all aspects of SLURM operations
- **Multi-dimensional labeling** for detailed analysis and alerting
- **Real-time data collection** with configurable intervals
- **Enterprise-scale support** for clusters with 1000+ nodes and 10000+ jobs

### 🚀 Production Ready
- **High availability** deployment patterns
- **Graceful degradation** when SLURM services are unavailable
- **Comprehensive error handling** with structured logging
- **Performance optimized** for minimal cluster impact

### 🐳 Flexible Deployment
- **Docker containerization** with multi-stage builds
- **Kubernetes manifests** and Helm charts included
- **Multiple authentication methods** (JWT tokens, API keys, service accounts)
- **TLS support** for secure metrics exposition

### 📊 Rich Metrics Categories

#### Cluster Overview
- Total cluster capacity (nodes, CPUs, memory, GPUs)
- Node state distribution (idle, allocated, mixed, down, draining)
- Overall utilization percentages
- SLURM version and configuration information

#### Node-Level Metrics
- Hardware specifications and capabilities
- Resource utilization (CPU, memory, GPU)
- Health indicators and uptime statistics
- Feature detection and constraints

#### Job Analytics
- Job state distribution and lifecycle tracking
- Queue lengths and wait times by partition
- Resource efficiency (requested vs. actual usage)
- Completion rates and failure analysis

#### User & Account Monitoring
- Per-user and per-account resource usage
- Fair-share calculations and priority tracking
- Quota monitoring and enforcement
- Active session tracking

#### Performance Metrics
- Job throughput rates (jobs/hour)
- Queue time vs. execution time analysis
- Resource waste identification
- Scheduling efficiency metrics

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Access to SLURM REST API (v0.0.40-v0.0.43 supported)
- SLURM cluster with REST API enabled

### Installation

```bash
# Clone the repository
git clone https://github.com/jontk/slurm-exporter.git
cd slurm-exporter

# Build the binary
make build

# Or install directly
make install
```

### Configuration

Create a configuration file:

```yaml
# config.yaml
slurm:
  base_url: "https://your-slurm-server:6820"
  auth:
    type: "jwt"
    token: "your-jwt-token"
  timeout: "30s"

server:
  address: ":8080"
  metrics_path: "/metrics"
  
collectors:
  cluster:
    enabled: true
    interval: "30s"
  nodes:
    enabled: true
    interval: "30s"
  jobs:
    enabled: true
    interval: "15s"
  users:
    enabled: true
    interval: "60s"
  partitions:
    enabled: true
    interval: "60s"
  performance:
    enabled: true
    interval: "30s"

logging:
  level: "info"
  format: "json"
```

### Running the Exporter

```bash
# Using the binary
./bin/slurm-exporter --config config.yaml

# Using Go run
go run . --config config.yaml

# Using Make
make run
```

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Run with Docker
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml localhost/slurm-exporter:latest

# Or use Docker Compose
make docker-compose-up
```

### Kubernetes Deployment

```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/

# Or use Helm
helm install slurm-exporter ./charts/slurm-exporter/
```

## Accessing Metrics

Once running, metrics are available at:

- **Metrics endpoint**: `http://localhost:8080/metrics`
- **Health check**: `http://localhost:8080/health`
- **Readiness check**: `http://localhost:8080/ready`

## Development

### Building from Source

```bash
# Install dependencies
go mod tidy

# Run tests
make test

# Run linting
make lint

# Run all checks
make check
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires SLURM cluster)
export SLURM_REST_URL="https://your-cluster:6820"
make integration-test

# Coverage report
make test-coverage
```

## Documentation

- [Installation Guide](docs/installation.md) - Detailed installation and setup instructions
- [Configuration Reference](docs/configuration.md) - Complete configuration options
- [Metrics Documentation](docs/metrics.md) - All available metrics with descriptions
- [Alerting Guide](docs/alerting.md) - Pre-built alerting rules and best practices

## Monitoring Dashboard

Import the included Grafana dashboard templates:

```bash
# Import dashboard JSON
kubectl create configmap grafana-dashboard --from-file=docs/grafana-dashboards.json
```

## Alerting Rules

Example Prometheus alerting rules:

```yaml
groups:
  - name: slurm-cluster
    rules:
      - alert: SlurmNodeDown
        expr: slurm_node_state{state="down"} > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "SLURM node {{ $labels.node }} is down"
          
      - alert: SlurmHighQueueLength
        expr: slurm_partition_queue_length > 100
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High queue length in partition {{ $labels.partition }}"
```

## Contributing

We welcome contributions! Please see our [contributing guidelines](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make check` to ensure all tests pass
6. Submit a pull request

## Architecture

The exporter is built with a modular architecture:

- **Collectors**: Individual metric collection modules for different SLURM components
- **SLURM Client**: Integration with the enterprise-grade SLURM REST API client
- **Configuration**: Flexible YAML-based configuration with environment variable overrides
- **Server**: HTTP server with health checks and metrics exposition
- **Metrics**: Prometheus metric definitions and registration

## Performance

- **Memory Usage**: < 100MB under normal operation
- **CPU Impact**: < 5% on SLURM head node
- **Collection Time**: < 30 seconds for 1000+ node clusters
- **Concurrent Safe**: Supports multiple collectors running in parallel

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/jontk/slurm-exporter/issues)
- **Discussions**: [GitHub Discussions](https://github.com/jontk/slurm-exporter/discussions)
- **Documentation**: [docs/](docs/)

## Acknowledgments

- Built on the [SLURM Client Library](https://github.com/jontk/slurm-client)
- Inspired by the Prometheus ecosystem and monitoring best practices
- Thanks to the SLURM community for the robust REST API

---

**Enterprise-Grade SLURM Monitoring** • **Production Ready** • **Comprehensive Metrics** • **Flexible Deployment**