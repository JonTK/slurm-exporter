// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	slurm "github.com/jontk/slurm-client"
)

const (
	diagnosticsCollectorSubsystem = "diagnostics"
)

type DiagnosticsCollector struct {
	client  slurm.SlurmClient
	logger  *logrus.Entry
	timeout time.Duration

	// Metrics
	serverThreadCount    *prometheus.Desc
	agentCount           *prometheus.Desc
	agentThreadCount     *prometheus.Desc
	agentQueueSize       *prometheus.Desc
	dbdAgentCount        *prometheus.Desc
	jobsSubmitted        *prometheus.Desc
	jobsStarted          *prometheus.Desc
	jobsCompleted        *prometheus.Desc
	jobsCanceled         *prometheus.Desc
	jobsFailed           *prometheus.Desc
	scheduleCycleMax     *prometheus.Desc
	scheduleCycleLast    *prometheus.Desc
	scheduleCycleMean    *prometheus.Desc
	scheduleCycleCounter *prometheus.Desc
	backfillCycleMax     *prometheus.Desc
	backfillCycleLast    *prometheus.Desc
	backfillCycleMean    *prometheus.Desc
	backfillCycleCounter *prometheus.Desc
	gittosCount          *prometheus.Desc
	gittosTime           *prometheus.Desc
	scheduleQueueLen     *prometheus.Desc
}

// NewDiagnosticsCollector creates a new diagnostics collector
func NewDiagnosticsCollector(client slurm.SlurmClient, logger *logrus.Entry, timeout time.Duration) *DiagnosticsCollector {
	return &DiagnosticsCollector{
		client:  client,
		logger:  logger.WithField("collector", "diagnostics"),
		timeout: timeout,

		serverThreadCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "server_thread_count"),
			"Number of server threads",
			[]string{"cluster"},
			nil,
		),
		agentCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "agent_count"),
			"Number of agents",
			[]string{"cluster"},
			nil,
		),
		agentThreadCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "agent_thread_count"),
			"Number of agent threads",
			[]string{"cluster"},
			nil,
		),
		agentQueueSize: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "agent_queue_size"),
			"Size of agent queue",
			[]string{"cluster"},
			nil,
		),
		dbdAgentCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "dbd_agent_count"),
			"Number of database agents",
			[]string{"cluster"},
			nil,
		),
		jobsSubmitted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "jobs_submitted_total"),
			"Total number of jobs submitted",
			[]string{"cluster"},
			nil,
		),
		jobsStarted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "jobs_started_total"),
			"Total number of jobs started",
			[]string{"cluster"},
			nil,
		),
		jobsCompleted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "jobs_completed_total"),
			"Total number of jobs completed",
			[]string{"cluster"},
			nil,
		),
		jobsCanceled: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "jobs_canceled_total"),
			"Total number of jobs canceled",
			[]string{"cluster"},
			nil,
		),
		jobsFailed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "jobs_failed_total"),
			"Total number of jobs failed",
			[]string{"cluster"},
			nil,
		),
		scheduleCycleMax: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "schedule_cycle_max_microseconds"),
			"Maximum schedule cycle time in microseconds",
			[]string{"cluster"},
			nil,
		),
		scheduleCycleLast: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "schedule_cycle_last_microseconds"),
			"Last schedule cycle time in microseconds",
			[]string{"cluster"},
			nil,
		),
		scheduleCycleMean: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "schedule_cycle_mean_microseconds"),
			"Mean schedule cycle time in microseconds",
			[]string{"cluster"},
			nil,
		),
		scheduleCycleCounter: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "schedule_cycle_counter"),
			"Number of schedule cycles",
			[]string{"cluster"},
			nil,
		),
		backfillCycleMax: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "backfill_cycle_max_microseconds"),
			"Maximum backfill cycle time in microseconds",
			[]string{"cluster"},
			nil,
		),
		backfillCycleLast: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "backfill_cycle_last_microseconds"),
			"Last backfill cycle time in microseconds",
			[]string{"cluster"},
			nil,
		),
		backfillCycleMean: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "backfill_cycle_mean_microseconds"),
			"Mean backfill cycle time in microseconds",
			[]string{"cluster"},
			nil,
		),
		backfillCycleCounter: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "backfill_cycle_counter"),
			"Number of backfill cycles",
			[]string{"cluster"},
			nil,
		),
		gittosCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "gittos_count"),
			"Number of gittos operations",
			[]string{"cluster"},
			nil,
		),
		gittosTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "gittos_time_microseconds"),
			"Total gittos time in microseconds",
			[]string{"cluster"},
			nil,
		),
		scheduleQueueLen: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, diagnosticsCollectorSubsystem, "schedule_queue_length"),
			"Length of schedule queue",
			[]string{"cluster"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *DiagnosticsCollector) Name() string {
	return "diagnostics_simple"
}

// Describe sends metric descriptions to the channel
func (c *DiagnosticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.serverThreadCount
	ch <- c.agentCount
	ch <- c.agentThreadCount
	// ch <- c.agentQueueSize // Not available in current API
	ch <- c.dbdAgentCount
	ch <- c.jobsSubmitted
	ch <- c.jobsStarted
	ch <- c.jobsCompleted
	ch <- c.jobsCanceled
	ch <- c.jobsFailed
	ch <- c.scheduleCycleMax
	ch <- c.scheduleCycleLast
	ch <- c.scheduleCycleMean
	ch <- c.scheduleCycleCounter
	ch <- c.backfillCycleMax
	ch <- c.backfillCycleLast
	ch <- c.backfillCycleMean
	ch <- c.backfillCycleCounter
	// ch <- c.gittosCount // Not available in current API
	// ch <- c.gittosTime // Not available in current API
	// ch <- c.scheduleQueueLen // Not available in current API
}

// Collect gathers metrics from SLURM
func (c *DiagnosticsCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.Debug("Collecting diagnostics metrics")

	// Get cluster name for labels
	clusterName := "default"
	infoManager := c.client.Info()
	if infoManager != nil {
		if info, err := infoManager.Get(ctx); err == nil && info != nil {
			clusterName = info.ClusterName
		}
	}

	// Get diagnostics
	diag, err := c.client.GetDiagnostics(ctx)
	if err != nil {
		c.logger.WithError(err).Debug("Failed to get diagnostics")
		return err
	}

	if diag == nil {
		return nil
	}

	// Export thread and agent metrics
	ch <- prometheus.MustNewConstMetric(
		c.serverThreadCount,
		prometheus.GaugeValue,
		float64(diag.ServerThreadCount),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.agentCount,
		prometheus.GaugeValue,
		float64(diag.AgentCount),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.agentThreadCount,
		prometheus.GaugeValue,
		float64(diag.AgentThreadCount),
		clusterName,
	)

	// AgentQueueSize not available in current Diagnostics struct
	// ch <- prometheus.MustNewConstMetric(
	// 	c.agentQueueSize,
	// 	prometheus.GaugeValue,
	// 	float64(diag.AgentQueueSize),
	// 	clusterName,
	// )

	ch <- prometheus.MustNewConstMetric(
		c.dbdAgentCount,
		prometheus.GaugeValue,
		float64(diag.DBDAgentCount),
		clusterName,
	)

	// Export job metrics
	ch <- prometheus.MustNewConstMetric(
		c.jobsSubmitted,
		prometheus.CounterValue,
		float64(diag.JobsSubmitted),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.jobsStarted,
		prometheus.CounterValue,
		float64(diag.JobsStarted),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.jobsCompleted,
		prometheus.CounterValue,
		float64(diag.JobsCompleted),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.jobsCanceled,
		prometheus.CounterValue,
		float64(diag.JobsCanceled),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.jobsFailed,
		prometheus.CounterValue,
		float64(diag.JobsFailed),
		clusterName,
	)

	// Export schedule cycle metrics
	ch <- prometheus.MustNewConstMetric(
		c.scheduleCycleMax,
		prometheus.GaugeValue,
		float64(diag.ScheduleCycleMax),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.scheduleCycleLast,
		prometheus.GaugeValue,
		float64(diag.ScheduleCycleLast),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.scheduleCycleMean,
		prometheus.GaugeValue,
		float64(diag.ScheduleCycleMean),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.scheduleCycleCounter,
		prometheus.CounterValue,
		float64(diag.ScheduleCycleCounter),
		clusterName,
	)

	// Export backfill cycle metrics (fields now prefixed with BF)
	ch <- prometheus.MustNewConstMetric(
		c.backfillCycleMax,
		prometheus.GaugeValue,
		float64(diag.BFCycleMax),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.backfillCycleLast,
		prometheus.GaugeValue,
		float64(diag.BFCycle),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.backfillCycleMean,
		prometheus.GaugeValue,
		float64(diag.BFCycleMean),
		clusterName,
	)

	// Note: BFCycleCounter doesn't exist, using BFCycle as substitute
	ch <- prometheus.MustNewConstMetric(
		c.backfillCycleCounter,
		prometheus.CounterValue,
		float64(diag.BFCycle),
		clusterName,
	)

	// Gittos metrics not available in current Diagnostics struct
	// ch <- prometheus.MustNewConstMetric(
	// 	c.gittosCount,
	// 	prometheus.CounterValue,
	// 	float64(diag.GittosCount),
	// 	clusterName,
	// )

	// ch <- prometheus.MustNewConstMetric(
	// 	c.gittosTime,
	// 	prometheus.CounterValue,
	// 	float64(diag.GittosTime),
	// 	clusterName,
	// )

	// Schedule queue length not available in current Diagnostics struct
	// ch <- prometheus.MustNewConstMetric(
	// 	c.scheduleQueueLen,
	// 	prometheus.GaugeValue,
	// 	float64(diag.ScheduleQueueLen),
	// 	clusterName,
	// )

	c.logger.Debug("Diagnostics metrics collected")
	return nil
}

// Close cleans up any resources
func (c *DiagnosticsCollector) Close() error {
	c.logger.Info("Closing diagnostics collector")
	return nil
}

// IsEnabled returns whether this collector is enabled
func (c *DiagnosticsCollector) IsEnabled() bool {
	return true
}

// SetEnabled sets whether this collector is enabled
func (c *DiagnosticsCollector) SetEnabled(enabled bool) {
	// For now, we don't track enabled state in these simple collectors
	// The enabled state is managed by the registry
}
