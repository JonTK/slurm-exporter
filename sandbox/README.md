# SLURM Exporter Sandbox Environment

A containerized environment for experimenting with the SLURM Exporter and a local monitoring stack.

> **Status**: This sandbox is a work in progress. The `docker-compose.yml` defines the full architecture below, but several required directories (`slurm-configs/`, `sample-jobs/`, `web/`, `grafana/`) have not yet been created. The sandbox cannot be started as-is.

## Architecture

The `docker-compose.yml` defines the following services:

| Service | Port | Description |
|---------|------|-------------|
| **SLURM Controller** | 6820, 6817 | SLURM controller with REST API (slurmrestd) |
| **SLURM Compute Node** | - | Simulated worker node |
| **Job Generator** | - | Submits sample jobs continuously |
| **SLURM Exporter** | 10341 | Collects metrics from the SLURM REST API |
| **Prometheus** | 9090 | Scrapes and stores metrics |
| **Alertmanager** | 9093 | Alert routing |
| **Grafana** | 3000 | Dashboards (admin/admin) |
| **Node Exporter** | 9100 | Host system metrics |
| **Tutorial Server** | 8888 | Nginx-based tutorial portal |

## Files

| File | Description |
|------|-------------|
| `docker-compose.yml` | Service definitions for the full sandbox stack |
| `sandbox-config.yaml` | Exporter configuration for sandbox use |
| `prometheus.yml` | Prometheus scrape configuration |
| `alert-rules.yml` | Prometheus alerting rules |
| `alertmanager.yml` | Alertmanager routing configuration |

## Missing Prerequisites

The following directories are referenced by `docker-compose.yml` but do not yet exist:

- `slurm-configs/` — SLURM configuration files (slurm.conf, etc.)
- `sample-jobs/` — Job submission scripts for the job generator
- `web/` — Static files and nginx.conf for the tutorial server
- `grafana/provisioning/` and `grafana/dashboards/` — Grafana auto-provisioning

The SLURM container image (`slurm-docker/slurm:latest`) must also be built or available locally.

## Usage (once prerequisites are in place)

```bash
cd sandbox
docker-compose up -d

# Check service health
docker-compose ps

# View exporter metrics
curl http://localhost:10341/metrics

# Access Grafana
open http://localhost:3000

# View logs
docker-compose logs slurm-exporter

# Tear down
docker-compose down -v
```
