// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"strings"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	namespace             = "slurm"
	qosCollectorSubsystem = "qos"
)

// QoSCollector collects QoS-related metrics
type QoSCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// QoS priority metrics
	qosPriority    *prometheus.Desc
	qosUsageFactor *prometheus.Desc

	// QoS job limits
	qosMaxJobs           *prometheus.Desc
	qosMaxJobsPerUser    *prometheus.Desc
	qosMaxJobsPerAccount *prometheus.Desc
	qosMaxSubmitJobs     *prometheus.Desc

	// QoS resource limits
	qosMaxCPUs        *prometheus.Desc
	qosMaxCPUsPerUser *prometheus.Desc
	qosMaxNodes       *prometheus.Desc
	qosMaxWallTime    *prometheus.Desc
	qosMinCPUs        *prometheus.Desc
	qosMinNodes       *prometheus.Desc

	// QoS info
	qosInfo *prometheus.Desc
}

// NewQoSCollector creates a new QoS collector
func NewQoSCollector(client slurm.SlurmClient, logger *logrus.Entry) *QoSCollector {
	c := &QoSCollector{
		client:  client,
		logger:  logger.WithField("collector", "qos"),
		enabled: true,
	}

	// Initialize metrics
	c.qosPriority = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "priority"),
		"Priority of the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosUsageFactor = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "usage_factor"),
		"Usage factor for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMaxJobs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "max_jobs"),
		"Maximum number of jobs for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMaxJobsPerUser = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "max_jobs_per_user"),
		"Maximum number of jobs per user for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMaxJobsPerAccount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "max_jobs_per_account"),
		"Maximum number of jobs per account for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMaxSubmitJobs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "max_submit_jobs"),
		"Maximum number of submitted jobs for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMaxCPUs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "max_cpus"),
		"Maximum number of CPUs for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMaxCPUsPerUser = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "max_cpus_per_user"),
		"Maximum number of CPUs per user for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMaxNodes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "max_nodes"),
		"Maximum number of nodes for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMaxWallTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "max_wall_time_seconds"),
		"Maximum wall time for jobs in the QoS (seconds)",
		[]string{"qos"},
		nil,
	)

	c.qosMinCPUs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "min_cpus"),
		"Minimum number of CPUs for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosMinNodes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "min_nodes"),
		"Minimum number of nodes for the QoS",
		[]string{"qos"},
		nil,
	)

	c.qosInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, qosCollectorSubsystem, "info"),
		"QoS information with all labels",
		[]string{"qos", "description", "preempt_mode", "flags"},
		nil,
	)

	return c
}

// Name returns the collector name
func (c *QoSCollector) Name() string {
	return "qos"
}

// IsEnabled returns whether this collector is enabled
func (c *QoSCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *QoSCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *QoSCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.qosPriority
	ch <- c.qosUsageFactor
	ch <- c.qosMaxJobs
	ch <- c.qosMaxJobsPerUser
	ch <- c.qosMaxJobsPerAccount
	ch <- c.qosMaxSubmitJobs
	ch <- c.qosMaxCPUs
	ch <- c.qosMaxCPUsPerUser
	ch <- c.qosMaxNodes
	ch <- c.qosMaxWallTime
	ch <- c.qosMinCPUs
	ch <- c.qosMinNodes
	ch <- c.qosInfo
}

// Collect implements the Collector interface
func (c *QoSCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ch)
}

// collect gathers metrics from SLURM
func (c *QoSCollector) collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get QoS manager from client
	qosManager := c.client.QoS()
	if qosManager == nil {
		return fmt.Errorf("QoS manager not available")
	}

	// List all QoS
	qosList, err := qosManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list QoS")
		return err
	}

	c.logger.WithField("count", len(qosList.QoS)).Debug("Collected QoS entries")

	for _, qos := range qosList.QoS {
		// Priority and usage factor
		ch <- prometheus.MustNewConstMetric(
			c.qosPriority,
			prometheus.GaugeValue,
			float64(qos.Priority),
			qos.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.qosUsageFactor,
			prometheus.GaugeValue,
			qos.UsageFactor,
			qos.Name,
		)

		// Job limits
		c.sendMetricIfNotInfinite(ch, c.qosMaxJobs, qos.Name, qos.MaxJobs)
		c.sendMetricIfNotInfinite(ch, c.qosMaxJobsPerUser, qos.Name, qos.MaxJobsPerUser)
		c.sendMetricIfNotInfinite(ch, c.qosMaxJobsPerAccount, qos.Name, qos.MaxJobsPerAccount)
		c.sendMetricIfNotInfinite(ch, c.qosMaxSubmitJobs, qos.Name, qos.MaxSubmitJobs)

		// Resource limits
		c.sendMetricIfNotInfinite(ch, c.qosMaxCPUs, qos.Name, qos.MaxCPUs)
		c.sendMetricIfNotInfinite(ch, c.qosMaxCPUsPerUser, qos.Name, qos.MaxCPUsPerUser)
		c.sendMetricIfNotInfinite(ch, c.qosMaxNodes, qos.Name, qos.MaxNodes)

		// Wall time (convert minutes to seconds)
		if qos.MaxWallTime > 0 && qos.MaxWallTime < 365*24*60 { // Less than a year
			ch <- prometheus.MustNewConstMetric(
				c.qosMaxWallTime,
				prometheus.GaugeValue,
				float64(qos.MaxWallTime*60),
				qos.Name,
			)
		} else if qos.MaxWallTime >= 365*24*60 {
			ch <- prometheus.MustNewConstMetric(
				c.qosMaxWallTime,
				prometheus.GaugeValue,
				-1, // Infinite
				qos.Name,
			)
		}

		// Minimum requirements
		ch <- prometheus.MustNewConstMetric(
			c.qosMinCPUs,
			prometheus.GaugeValue,
			float64(qos.MinCPUs),
			qos.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.qosMinNodes,
			prometheus.GaugeValue,
			float64(qos.MinNodes),
			qos.Name,
		)

		// QoS info
		preemptMode := qos.PreemptMode // PreemptMode is already a string
		flags := strings.Join(qos.Flags, ",")
		ch <- prometheus.MustNewConstMetric(
			c.qosInfo,
			prometheus.GaugeValue,
			1,
			qos.Name, qos.Description, preemptMode, flags,
		)
	}

	return nil
}

// sendMetricIfNotInfinite sends a metric value if it's not representing "infinite"
func (c *QoSCollector) sendMetricIfNotInfinite(ch chan<- prometheus.Metric, desc *prometheus.Desc, label string, value int) {
	// SLURM often uses very large numbers to represent "infinite" or "unlimited"
	// We'll use -1 to represent infinite in metrics
	metricValue := float64(0)
	if value > 0 && value < 1000000 {
		metricValue = float64(value)
	} else if value >= 1000000 {
		metricValue = -1 // Infinite
	}

	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		metricValue,
		label,
	)
}
