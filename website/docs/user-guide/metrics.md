# Metrics Catalog

This page provides a comprehensive catalog of all metrics exposed by SLURM Exporter.

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
| `slurm_job_gpu_utilization_ratio` | Gauge | GPU utilization (0-1) | `job_id`, `partition`, `user`, `gpu_id` |
| `slurm_job_io_read_bytes_total` | Counter | Total bytes read | `job_id`, `partition`, `user` |
| `slurm_job_io_write_bytes_total` | Counter | Total bytes written | `job_id`, `partition`, `user` |

### Job Step Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_job_step_cpu_time_seconds` | Gauge | CPU time consumed by step | `job_id`, `step_id`, `user` |
| `slurm_job_step_memory_max_bytes` | Gauge | Maximum memory used by step | `job_id`, `step_id`, `user` |
| `slurm_job_step_runtime_seconds` | Gauge | Step runtime | `job_id`, `step_id`, `user` |

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

### Node Performance

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_node_load_average` | Gauge | Node load average | `node`, `period` |
| `slurm_node_cpu_utilization_ratio` | Gauge | CPU utilization (0-1) | `node` |
| `slurm_node_memory_utilization_ratio` | Gauge | Memory utilization (0-1) | `node` |
| `slurm_node_boot_time` | Gauge | Node boot timestamp | `node` |
| `slurm_node_slurmd_start_time` | Gauge | Slurmd start timestamp | `node` |

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

## Account Metrics

### Account Usage

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_account_jobs_running` | Gauge | Running jobs for account | `account`, `user` |
| `slurm_account_jobs_pending` | Gauge | Pending jobs for account | `account`, `user` |
| `slurm_account_cpu_hours_used_total` | Counter | Total CPU hours consumed | `account`, `user` |
| `slurm_account_cpu_hours_allocated` | Gauge | Allocated CPU hours | `account` |
| `slurm_account_cpu_hours_remaining` | Gauge | Remaining CPU hours | `account` |

### Account Limits

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_account_max_jobs` | Gauge | Maximum concurrent jobs | `account` |
| `slurm_account_max_submit_jobs` | Gauge | Maximum submittable jobs | `account` |
| `slurm_account_max_cpus` | Gauge | Maximum CPUs | `account` |
| `slurm_account_max_nodes` | Gauge | Maximum nodes | `account` |
| `slurm_account_max_wall_duration_seconds` | Gauge | Maximum wall time | `account` |

## Fairshare Metrics

### Fairshare Values

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_fairshare_shares` | Gauge | Fairshare allocation | `account`, `user` |
| `slurm_fairshare_usage_raw` | Gauge | Raw usage value | `account`, `user` |
| `slurm_fairshare_usage_normalized` | Gauge | Normalized usage (0-1) | `account`, `user` |
| `slurm_fairshare_factor` | Gauge | Current fairshare factor | `account`, `user` |
| `slurm_fairshare_level_factor` | Gauge | Level fairshare factor | `account`, `user` |

### Fairshare Trends

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_fairshare_usage_trend` | Gauge | Usage trend indicator | `account`, `user`, `period` |
| `slurm_fairshare_priority_trend` | Gauge | Priority trend | `account`, `user`, `period` |
| `slurm_fairshare_decay_factor` | Gauge | Usage decay factor | `cluster` |

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

## License Metrics

### License Usage

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_licenses_total` | Gauge | Total available licenses | `license`, `feature` |
| `slurm_licenses_used` | Gauge | Currently used licenses | `license`, `feature` |
| `slurm_licenses_free` | Gauge | Free licenses | `license`, `feature` |
| `slurm_licenses_reserved` | Gauge | Reserved licenses | `license`, `feature` |

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
| `slurm_database_up` | Gauge | Database availability (1=up, 0=down) | `database` |
| `slurm_version_info` | Gauge | SLURM version information | `version`, `build` |

### Scheduler Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_scheduler_queue_length` | Gauge | Number of jobs in queue | `priority` |
| `slurm_scheduler_backfill_jobs_since_start_total` | Counter | Backfilled jobs since start | `cluster` |
| `slurm_scheduler_backfill_jobs_since_cycle_total` | Counter | Backfilled jobs in last cycle | `cluster` |
| `slurm_scheduler_cycle_duration_seconds` | Gauge | Last scheduling cycle duration | `cluster` |

## Advanced Analytics Metrics

### Job Efficiency Analytics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_job_efficiency_score` | Gauge | Overall efficiency score (0-1) | `job_id`, `user` |
| `slurm_job_waste_cpu_hours` | Gauge | Wasted CPU hours | `job_id`, `user` |
| `slurm_job_waste_memory_gb_hours` | Gauge | Wasted memory GB-hours | `job_id`, `user` |
| `slurm_job_bottleneck_score` | Gauge | Bottleneck severity (0-1) | `job_id`, `resource_type` |

### Trend Analytics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_usage_trend_hourly` | Gauge | Hourly usage trend | `resource_type`, `hour` |
| `slurm_usage_trend_daily` | Gauge | Daily usage trend | `resource_type`, `day` |
| `slurm_usage_trend_weekly` | Gauge | Weekly usage trend | `resource_type`, `week` |

### Anomaly Detection

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_anomaly_score` | Gauge | Anomaly detection score | `entity_type`, `entity_id` |
| `slurm_anomaly_threshold` | Gauge | Anomaly detection threshold | `entity_type` |

## Exporter Self-Monitoring

### Collection Performance

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_exporter_collect_duration_seconds` | Histogram | Time spent collecting metrics | `collector` |
| `slurm_exporter_collect_errors_total` | Counter | Collection errors | `collector`, `error_type` |
| `slurm_exporter_last_collect_timestamp` | Gauge | Last successful collection time | `collector` |

### Cache Performance

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_exporter_cache_hits_total` | Counter | Cache hit count | `cache_type` |
| `slurm_exporter_cache_misses_total` | Counter | Cache miss count | `cache_type` |
| `slurm_exporter_cache_size` | Gauge | Current cache size | `cache_type` |

### API Performance

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_exporter_api_request_duration_seconds` | Histogram | SLURM API request duration | `endpoint`, `method` |
| `slurm_exporter_api_requests_total` | Counter | Total API requests | `endpoint`, `method`, `status` |
| `slurm_exporter_connection_pool_active` | Gauge | Active connections | `pool` |
| `slurm_exporter_connection_pool_idle` | Gauge | Idle connections | `pool` |

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

### Advanced Analytics Queries

```promql
# Average job efficiency by user
avg(slurm_job_cpu_efficiency_ratio) by (user)

# Wasted resources by partition
sum(slurm_job_waste_cpu_hours) by (partition)

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