# Metrics Catalog

This page provides a catalog of metrics exposed by SLURM Exporter.

## Metric Naming Convention

All SLURM Exporter metrics follow the Prometheus naming convention:

- **Prefix**: All metrics start with `slurm_`
- **Component**: Indicates the SLURM component (`job_`, `node_`, `partition_`, etc.)
- **Measurement**: Describes what is being measured
- **Unit**: Included in the metric name when applicable (`_bytes`, `_seconds`, `_total`)

## Common Labels

Many metrics include these common labels:

| Label | Description | Example Values |
|-------|-------------|----------------|
| `cluster` | SLURM cluster name | `production`, `development` |
| `partition` | SLURM partition name | `compute`, `gpu`, `highmem` |
| `job_id` | SLURM job ID | `12345` |
| `user` | Username | `alice`, `bob` |
| `account` | SLURM account | `research`, `engineering` |
| `qos` | Quality of Service | `normal`, `high`, `low` |
| `state` | Resource state | `idle`, `allocated`, `down` |

## Job Metrics

### Basic Job Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_jobs_total` | Counter | Total number of jobs processed | `state`, `partition`, `user`, `account` |
| `slurm_jobs_running` | Gauge | Number of currently running jobs | `partition`, `user`, `account` |
| `slurm_jobs_pending` | Gauge | Number of pending jobs | `partition`, `user`, `account`, `reason` |
| `slurm_jobs_completed_total` | Counter | Total completed jobs | `partition`, `user`, `account`, `exit_code` |
| `slurm_jobs_failed_total` | Counter | Total failed jobs | `partition`, `user`, `account`, `exit_code` |

### Job Resource Usage

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_job_cpus_allocated` | Gauge | CPUs allocated to job | `job_id`, `partition`, `user` |
| `slurm_job_memory_allocated_bytes` | Gauge | Memory allocated to job | `job_id`, `partition`, `user` |
| `slurm_job_runtime_seconds` | Gauge | Job runtime in seconds | `job_id`, `partition`, `user` |
| `slurm_job_time_limit_seconds` | Gauge | Job time limit in seconds | `job_id`, `partition`, `user` |
| `slurm_job_submission_time` | Gauge | Job submission timestamp | `job_id`, `partition`, `user` |
| `slurm_job_start_time` | Gauge | Job start timestamp | `job_id`, `partition`, `user` |
| `slurm_job_end_time` | Gauge | Job end timestamp | `job_id`, `partition`, `user` |

### Job Performance Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_job_cpu_efficiency_ratio` | Gauge | CPU efficiency (0-1) | `job_id`, `partition`, `user` |
| `slurm_job_memory_efficiency_ratio` | Gauge | Memory efficiency (0-1) | `job_id`, `partition`, `user` |
| `slurm_job_wait_time_seconds` | Gauge | Time spent waiting in queue | `job_id`, `partition`, `user` |

## Node Metrics

### Node State and Health

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_nodes_total` | Gauge | Total number of nodes | `state`, `partition` |
| `slurm_node_up` | Gauge | Node availability (1=up, 0=down) | `node`, `partition` |
| `slurm_node_state` | Gauge | Node state (encoded) | `node`, `partition`, `state` |
| `slurm_node_drain_info` | Gauge | Node drain information | `node`, `reason` |

### Node Resources

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_node_cpus_total` | Gauge | Total CPUs on node | `node`, `partition` |
| `slurm_node_cpus_allocated` | Gauge | Allocated CPUs | `node`, `partition` |
| `slurm_node_cpus_idle` | Gauge | Idle CPUs | `node`, `partition` |
| `slurm_node_memory_total_bytes` | Gauge | Total memory on node | `node`, `partition` |
| `slurm_node_memory_allocated_bytes` | Gauge | Allocated memory | `node`, `partition` |
| `slurm_node_memory_free_bytes` | Gauge | Free memory | `node`, `partition` |

### Node Hardware

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_node_cores_per_socket` | Gauge | CPU cores per socket | `node` |
| `slurm_node_sockets` | Gauge | Number of CPU sockets | `node` |
| `slurm_node_threads_per_core` | Gauge | Threads per CPU core | `node` |
| `slurm_node_gpus_total` | Gauge | Total GPUs on node | `node`, `gpu_type` |
| `slurm_node_gpus_allocated` | Gauge | Allocated GPUs | `node`, `gpu_type` |

## Partition Metrics

### Partition State

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_partition_up` | Gauge | Partition availability (1=up, 0=down) | `partition` |
| `slurm_partition_state` | Gauge | Partition state (encoded) | `partition`, `state` |
| `slurm_partition_nodes_total` | Gauge | Total nodes in partition | `partition` |
| `slurm_partition_nodes_idle` | Gauge | Idle nodes in partition | `partition` |
| `slurm_partition_nodes_allocated` | Gauge | Allocated nodes in partition | `partition` |
| `slurm_partition_nodes_down` | Gauge | Down nodes in partition | `partition` |

### Partition Resources

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_partition_cpus_total` | Gauge | Total CPUs in partition | `partition` |
| `slurm_partition_cpus_allocated` | Gauge | Allocated CPUs in partition | `partition` |
| `slurm_partition_cpus_idle` | Gauge | Idle CPUs in partition | `partition` |
| `slurm_partition_memory_total_bytes` | Gauge | Total memory in partition | `partition` |
| `slurm_partition_memory_allocated_bytes` | Gauge | Allocated memory in partition | `partition` |

### Partition Limits

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_partition_max_time_seconds` | Gauge | Maximum job time limit | `partition` |
| `slurm_partition_default_time_seconds` | Gauge | Default job time limit | `partition` |
| `slurm_partition_max_nodes` | Gauge | Maximum nodes per job | `partition` |
| `slurm_partition_max_cpus_per_node` | Gauge | Maximum CPUs per node | `partition` |

## QoS (Quality of Service) Metrics

### QoS Limits

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_qos_max_jobs_per_user` | Gauge | Max jobs per user | `qos` |
| `slurm_qos_max_submit_jobs_per_user` | Gauge | Max submit jobs per user | `qos` |
| `slurm_qos_max_cpus_per_user` | Gauge | Max CPUs per user | `qos` |
| `slurm_qos_max_nodes_per_user` | Gauge | Max nodes per user | `qos` |
| `slurm_qos_max_wall_duration_seconds` | Gauge | Max wall time | `qos` |

### QoS Usage

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_qos_jobs_running` | Gauge | Running jobs with QoS | `qos`, `user` |
| `slurm_qos_jobs_pending` | Gauge | Pending jobs with QoS | `qos`, `user` |
| `slurm_qos_priority_factor` | Gauge | QoS priority factor | `qos` |

## Reservation Metrics

### Reservation State

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_reservation_active` | Gauge | Reservation is active (1/0) | `reservation` |
| `slurm_reservation_nodes_total` | Gauge | Total nodes in reservation | `reservation` |
| `slurm_reservation_nodes_allocated` | Gauge | Allocated nodes in reservation | `reservation` |
| `slurm_reservation_nodes_idle` | Gauge | Idle nodes in reservation | `reservation` |

### Reservation Time

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_reservation_start_time` | Gauge | Reservation start timestamp | `reservation` |
| `slurm_reservation_end_time` | Gauge | Reservation end timestamp | `reservation` |
| `slurm_reservation_duration_seconds` | Gauge | Reservation duration | `reservation` |

## TRES (Trackable Resources) Metrics

### TRES Allocation

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_tres_allocated_total` | Gauge | Total allocated TRES units | `tres_type`, `tres_name` |
| `slurm_tres_configured_total` | Gauge | Total configured TRES units | `tres_type`, `tres_name` |
| `slurm_tres_usage_total` | Counter | Cumulative TRES usage | `tres_type`, `tres_name` |

## System Metrics

### Cluster Health

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_controller_up` | Gauge | Controller availability (1=up, 0=down) | `controller` |
| `slurm_version_info` | Gauge | SLURM version information | `version`, `build` |

### Scheduler Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_scheduler_queue_length` | Gauge | Number of jobs in queue | `priority` |
| `slurm_scheduler_backfill_jobs_since_start_total` | Counter | Backfilled jobs since start | `cluster` |
| `slurm_scheduler_cycle_duration_seconds` | Gauge | Last scheduling cycle duration | `cluster` |

## Exporter Self-Monitoring

### Collection Performance

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_exporter_collect_duration_seconds` | Histogram | Time spent collecting metrics | `collector` |
| `slurm_exporter_collect_errors_total` | Counter | Collection errors | `collector`, `error_type` |
| `slurm_exporter_last_collect_timestamp` | Gauge | Last successful collection time | `collector` |

### API Performance

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_exporter_api_request_duration_seconds` | Histogram | SLURM API request duration | `endpoint`, `method` |
| `slurm_exporter_api_requests_total` | Counter | Total API requests | `endpoint`, `method`, `status` |

## Usage Examples

### Basic Resource Monitoring

```promql
# CPU utilization across all nodes
slurm_node_cpus_allocated / slurm_node_cpus_total

# Memory utilization by partition
sum(slurm_partition_memory_allocated_bytes) by (partition) /
sum(slurm_partition_memory_total_bytes) by (partition)

# Job completion rate
rate(slurm_jobs_completed_total[5m])
```

### Advanced Queries

```promql
# Average job efficiency by user
avg(slurm_job_cpu_efficiency_ratio) by (user)

# Queue wait time distribution
histogram_quantile(0.95, slurm_job_wait_time_seconds)
```

### Alert Examples

```promql
# High queue length
slurm_scheduler_queue_length > 1000

# Low partition efficiency
(slurm_partition_cpus_allocated / slurm_partition_cpus_total) < 0.6

# Node failures
increase(slurm_node_up == 0[5m]) > 0
```

For more detailed examples and monitoring strategies, see the [Alerting Setup](alerting.md) and [Integration](../integration/prometheus.md) sections.
