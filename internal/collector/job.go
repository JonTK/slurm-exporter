// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	// Commented out as only used in commented-out parser functions
	// "strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

// JobCollector collects job-level metrics
type JobCollector struct {
	*BaseCollector
	metrics      *metrics.MetricDefinitions
	clusterName  string
	errorBuilder *ErrorBuilder
}

// NewJobCollector creates a new job collector
func NewJobCollector(
	config *config.CollectorConfig,
	opts *CollectorOptions,
	client interface{}, // Will be *slurm.Client when slurm-client build issues are fixed
	metrics *metrics.MetricDefinitions,
	clusterName string,
) *JobCollector {
	base := NewBaseCollector("job", config, opts, client, &CollectorMetrics{
		Total:    metrics.CollectionSuccess,
		Errors:   metrics.CollectionErrors,
		Duration: metrics.CollectionDuration,
		Up:       metrics.CollectorUp,
	}, nil)

	return &JobCollector{
		BaseCollector: base,
		metrics:       metrics,
		clusterName:   clusterName,
		errorBuilder:  NewErrorBuilder("job", base.logger),
	}
}

// Describe implements the Collector interface
func (jc *JobCollector) Describe(ch chan<- *prometheus.Desc) {
	// Job information and state metrics
	jc.metrics.JobInfo.Describe(ch)
	jc.metrics.JobStates.Describe(ch)
	jc.metrics.JobQueueTime.Describe(ch)
	jc.metrics.JobRunTime.Describe(ch)
	jc.metrics.JobWaitTime.Describe(ch)
	jc.metrics.JobCPURequested.Describe(ch)
	jc.metrics.JobMemoryRequested.Describe(ch)
	jc.metrics.JobNodesRequested.Describe(ch)
	jc.metrics.JobCPUAllocated.Describe(ch)
	jc.metrics.JobMemoryAllocated.Describe(ch)
	jc.metrics.JobNodesAllocated.Describe(ch)
	jc.metrics.JobPriority.Describe(ch)
	jc.metrics.JobStartTime.Describe(ch)
	jc.metrics.JobEndTime.Describe(ch)
	jc.metrics.JobExitCode.Describe(ch)
}

// Collect implements the Collector interface
func (jc *JobCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	return jc.CollectWithMetrics(ctx, ch, jc.collectJobMetrics)
}

// collectJobMetrics performs the actual job metrics collection
func (jc *JobCollector) collectJobMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	jc.logger.Debug("Starting job metrics collection")

	// Since SLURM client is not available, we'll simulate the metrics for now
	// In real implementation, this would use the SLURM client to fetch job data

	// Collect running and pending jobs
	if err := jc.collectActiveJobs(ctx, ch); err != nil {
		return jc.errorBuilder.API(err, "/slurm/v1/jobs", 500, "",
			"Check SLURM job data availability",
			"Verify job information is accessible")
	}

	// Collect job queue statistics
	if err := jc.collectJobQueueStats(ctx, ch); err != nil {
		return jc.errorBuilder.API(err, "/slurm/v1/jobs", 500, "",
			"Check SLURM job queue data",
			"Verify queue statistics are available")
	}

	// Collect job states summary
	if err := jc.collectJobStatesSummary(ctx, ch); err != nil {
		return jc.errorBuilder.API(err, "/slurm/v1/jobs", 500, "",
			"Check SLURM job states data",
			"Verify job state information is accessible")
	}

	jc.logger.Debug("Completed job metrics collection")
	return nil
}

// sendJobMetric sends a job metric with standard labels (cluster, jobID, user, account, partition)
func (jc *JobCollector) sendJobMetric(ch chan<- prometheus.Metric, metricVec *prometheus.GaugeVec, value float64, jobID, user, account, partition string) {
	jc.SendMetric(ch, jc.BuildMetric(
		metricVec.WithLabelValues(jc.clusterName, jobID, user, account, partition).Desc(),
		prometheus.GaugeValue,
		value,
		jc.clusterName, jobID, user, account, partition,
	))
}

// sendJobInfoMetric sends the job info metric with state label
func (jc *JobCollector) sendJobInfoMetric(ch chan<- prometheus.Metric, jobID, jobName, user, account, partition, state string) {
	jc.SendMetric(ch, jc.BuildMetric(
		jc.metrics.JobInfo.WithLabelValues(jc.clusterName, jobID, jobName, user, account, partition, state).Desc(),
		prometheus.GaugeValue,
		1,
		jc.clusterName, jobID, jobName, user, account, partition, state,
	))
}

// collectActiveJobs collects metrics for individual active jobs
func (jc *JobCollector) collectActiveJobs(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate active job data - in real implementation this would come from SLURM API
	// This represents what we might get from /slurm/v1/jobs
	activeJobs := []struct {
		JobID           string
		JobName         string
		User            string
		Account         string
		Partition       string
		State           string
		CPUsRequested   int
		MemoryRequested int64 // bytes
		NodesRequested  int
		CPUsAllocated   int
		MemoryAllocated int64 // bytes
		NodesAllocated  int
		Priority        int
		SubmitTime      int64 // unix timestamp
		StartTime       int64 // unix timestamp
		EndTime         int64 // unix timestamp
		ExitCode        int
	}{
		{
			JobID:           "12345",
			JobName:         "training_job",
			User:            "alice",
			Account:         "ml_team",
			Partition:       "gpu",
			State:           "RUNNING",
			CPUsRequested:   24,
			MemoryRequested: 64 * 1024 * 1024 * 1024, // 64GB
			NodesRequested:  1,
			CPUsAllocated:   24,
			MemoryAllocated: 64 * 1024 * 1024 * 1024, // 64GB
			NodesAllocated:  1,
			Priority:        1000,
			SubmitTime:      time.Now().Add(-2 * time.Hour).Unix(),
			StartTime:       time.Now().Add(-90 * time.Minute).Unix(),
			EndTime:         0, // Still running
			ExitCode:        0,
		},
		{
			JobID:           "12346",
			JobName:         "simulation",
			User:            "bob",
			Account:         "physics",
			Partition:       "compute",
			State:           "PENDING",
			CPUsRequested:   48,
			MemoryRequested: 128 * 1024 * 1024 * 1024, // 128GB
			NodesRequested:  2,
			CPUsAllocated:   0,
			MemoryAllocated: 0,
			NodesAllocated:  0,
			Priority:        500,
			SubmitTime:      time.Now().Add(-30 * time.Minute).Unix(),
			StartTime:       0, // Not started yet
			EndTime:         0,
			ExitCode:        0,
		},
		{
			JobID:           "12344",
			JobName:         "analysis",
			User:            "charlie",
			Account:         "bio_team",
			Partition:       "compute",
			State:           "COMPLETED",
			CPUsRequested:   16,
			MemoryRequested: 32 * 1024 * 1024 * 1024, // 32GB
			NodesRequested:  1,
			CPUsAllocated:   16,
			MemoryAllocated: 32 * 1024 * 1024 * 1024, // 32GB
			NodesAllocated:  1,
			Priority:        750,
			SubmitTime:      time.Now().Add(-4 * time.Hour).Unix(),
			StartTime:       time.Now().Add(-3 * time.Hour).Unix(),
			EndTime:         time.Now().Add(-1 * time.Hour).Unix(),
			ExitCode:        0,
		},
	}

	for _, job := range activeJobs {
		normalizedState := jc.parseJobState(job.State)

		// Job info metric
		jc.sendJobInfoMetric(ch, job.JobID, job.JobName, job.User, job.Account, job.Partition, normalizedState)

		// Resource request metrics
		jc.sendJobMetric(ch, jc.metrics.JobCPURequested, float64(job.CPUsRequested), job.JobID, job.User, job.Account, job.Partition)
		jc.sendJobMetric(ch, jc.metrics.JobMemoryRequested, float64(job.MemoryRequested), job.JobID, job.User, job.Account, job.Partition)
		jc.sendJobMetric(ch, jc.metrics.JobNodesRequested, float64(job.NodesRequested), job.JobID, job.User, job.Account, job.Partition)

		// Resource allocation metrics
		jc.sendJobMetric(ch, jc.metrics.JobCPUAllocated, float64(job.CPUsAllocated), job.JobID, job.User, job.Account, job.Partition)
		jc.sendJobMetric(ch, jc.metrics.JobMemoryAllocated, float64(job.MemoryAllocated), job.JobID, job.User, job.Account, job.Partition)
		jc.sendJobMetric(ch, jc.metrics.JobNodesAllocated, float64(job.NodesAllocated), job.JobID, job.User, job.Account, job.Partition)

		// Priority metric
		jc.sendJobMetric(ch, jc.metrics.JobPriority, float64(job.Priority), job.JobID, job.User, job.Account, job.Partition)

		// Timing metrics
		if job.StartTime > 0 {
			jc.sendJobMetric(ch, jc.metrics.JobStartTime, float64(job.StartTime), job.JobID, job.User, job.Account, job.Partition)
		}

		if job.EndTime > 0 {
			jc.sendJobMetric(ch, jc.metrics.JobEndTime, float64(job.EndTime), job.JobID, job.User, job.Account, job.Partition)
			jc.sendJobMetric(ch, jc.metrics.JobExitCode, float64(job.ExitCode), job.JobID, job.User, job.Account, job.Partition)
		}
	}

	jc.LogCollectionf("Collected metrics for %d active jobs", len(activeJobs))
	return nil
}

// collectJobQueueStats collects queue time statistics
func (jc *JobCollector) collectJobQueueStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate queue statistics - in real implementation this would be calculated from job data
	queueStats := []struct {
		Partition string
		Account   string
		QueueTime float64 // seconds
		RunTime   float64 // seconds
		WaitTime  float64 // seconds
	}{
		{"gpu", "ml_team", 600, 5400, 600},      // 10 min queue, 90 min run, 10 min wait
		{"compute", "physics", 1800, 0, 1800},   // 30 min queue (pending), 0 run, 30 min wait
		{"compute", "bio_team", 300, 7200, 300}, // 5 min queue, 2 hour run, 5 min wait
	}

	for _, stat := range queueStats {
		// Queue time histogram
		jc.metrics.JobQueueTime.WithLabelValues(
			jc.clusterName,
			stat.Partition,
			stat.Account,
		).Observe(stat.QueueTime)

		// Run time histogram (if job has started)
		if stat.RunTime > 0 {
			jc.metrics.JobRunTime.WithLabelValues(
				jc.clusterName,
				stat.Partition,
				stat.Account,
			).Observe(stat.RunTime)
		}

		// Wait time histogram
		jc.metrics.JobWaitTime.WithLabelValues(
			jc.clusterName,
			stat.Partition,
			stat.Account,
		).Observe(stat.WaitTime)
	}

	jc.LogCollectionf("Collected queue statistics for %d partition/account combinations", len(queueStats))
	return nil
}

// collectJobStatesSummary collects job states summary
func (jc *JobCollector) collectJobStatesSummary(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate job states summary
	jobStates := map[string]int{
		"pending":    45,
		"running":    120,
		"completed":  850,
		"cancelled":  15,
		StatusFailed: 8,
		"timeout":    3,
	}

	for state, count := range jobStates {
		jc.SendMetric(ch, jc.BuildMetric(
			jc.metrics.JobStates.WithLabelValues(jc.clusterName, state).Desc(),
			prometheus.GaugeValue,
			float64(count),
			jc.clusterName,
			state,
		))
	}

	jc.LogCollectionf("Collected job states summary: %d states", len(jobStates))
	return nil
}

// parseJobState converts SLURM job state string to normalized state
func (jc *JobCollector) parseJobState(slurmState string) string {
	// SLURM job states: PENDING, RUNNING, SUSPENDED, COMPLETED, CANCELLED, FAILED, TIMEOUT, etc.
	switch slurmState {
	case "PENDING", "PD":
		return "pending"
	case "RUNNING", "R":
		return "running"
	case "SUSPENDED", "S":
		return "suspended"
	case "COMPLETED", "CD":
		return "completed"
	case "CANCELLED", "CA":
		return "cancelled"
	case "FAILED", "F":
		return StatusFailed
	case "TIMEOUT", "TO":
		return "timeout"
	case "NODE_FAIL", "NF":
		return "node_fail"
	case "PREEMPTED", "PR":
		return "preempted"
	case "BOOT_FAIL", "BF":
		return "boot_fail"
	case "DEADLINE", "DL":
		return "deadline"
	case "OUT_OF_MEMORY", "OOM":
		return "out_of_memory"
	default:
		return "unknown"
	}
}

// TODO: Following parser functions are unused - preserved for future data parsing needs
/*
// parseTimeString converts SLURM time string to Unix timestamp
func (jc *JobCollector) parseTimeString(slurmTime string) (int64, error) {
	if slurmTime == "" || slurmTime == "Unknown" {
		return 0, nil
	}

	// SLURM typically uses format: "2023-02-15T10:30:25"
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, slurmTime); err == nil {
			return t.Unix(), nil
		}
	}

	return 0, jc.errorBuilder.Parsing(nil, "parse_time", "timestamp",
		"Check SLURM time format configuration",
		"Verify time zone settings")
}

// parseMemorySize converts SLURM memory specification to bytes
func (jc *JobCollector) parseMemorySize(slurmMemory string) (int64, error) {
	// SLURM memory can be in various formats: "1024M", "2G", "500000K", etc.
	if slurmMemory == "" {
		return 0, nil
	}

	// Extract numeric part and unit
	var value float64
	var unit string

	// Simple parsing - in real implementation, use more robust parsing
	if len(slurmMemory) > 0 {
		lastChar := slurmMemory[len(slurmMemory)-1]
		switch lastChar {
		case 'K', 'k':
			unit = "K"
			if val, err := strconv.ParseFloat(slurmMemory[:len(slurmMemory)-1], 64); err == nil {
				value = val
			}
		case 'M', 'm':
			unit = "M"
			if val, err := strconv.ParseFloat(slurmMemory[:len(slurmMemory)-1], 64); err == nil {
				value = val
			}
		case 'G', 'g':
			unit = "G"
			if val, err := strconv.ParseFloat(slurmMemory[:len(slurmMemory)-1], 64); err == nil {
				value = val
			}
		case 'T', 't':
			unit = "T"
			if val, err := strconv.ParseFloat(slurmMemory[:len(slurmMemory)-1], 64); err == nil {
				value = val
			}
		default:
			// Assume MB if no unit
			unit = "M"
			if val, err := strconv.ParseFloat(slurmMemory, 64); err == nil {
				value = val
			}
		}
	}

	// Convert to bytes
	switch unit {
	case "K":
		return int64(value * 1024), nil
	case "M":
		return int64(value * 1024 * 1024), nil
	case "G":
		return int64(value * 1024 * 1024 * 1024), nil
	case "T":
		return int64(value * 1024 * 1024 * 1024 * 1024), nil
	default:
		return int64(value), nil
	}
}
*/
