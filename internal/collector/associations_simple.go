// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	associationsCollectorSubsystem = "association"
)

// AssociationsSimpleCollector collects association-related metrics
type AssociationsSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Association metrics
	associationInfo        *prometheus.Desc
	associationCPULimit    *prometheus.Desc
	associationMemoryLimit *prometheus.Desc
	associationTimeLimit   *prometheus.Desc
	associationPriority    *prometheus.Desc
	associationSharesRaw   *prometheus.Desc
	associationSharesNorm  *prometheus.Desc
}

// NewAssociationsSimpleCollector creates a new Associations collector
func NewAssociationsSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *AssociationsSimpleCollector {
	c := &AssociationsSimpleCollector{
		client:  client,
		logger:  logger.WithField("collector", "associations"),
		enabled: true,
	}

	// Initialize metrics
	c.associationInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, associationsCollectorSubsystem, "info"),
		"Association information with all labels",
		[]string{"user", "account", "cluster", "partition", "qos"},
		nil,
	)

	c.associationCPULimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, associationsCollectorSubsystem, "cpu_limit"),
		"CPU limit for the association",
		[]string{"user", "account", "cluster", "partition"},
		nil,
	)

	c.associationMemoryLimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, associationsCollectorSubsystem, "memory_limit_bytes"),
		"Memory limit for the association in bytes",
		[]string{"user", "account", "cluster", "partition"},
		nil,
	)

	c.associationTimeLimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, associationsCollectorSubsystem, "time_limit_minutes"),
		"Time limit for the association in minutes",
		[]string{"user", "account", "cluster", "partition"},
		nil,
	)

	c.associationPriority = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, associationsCollectorSubsystem, "priority"),
		"Priority for the association",
		[]string{"user", "account", "cluster", "partition"},
		nil,
	)

	c.associationSharesRaw = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, associationsCollectorSubsystem, "shares_raw"),
		"Raw shares for the association",
		[]string{"user", "account", "cluster", "partition"},
		nil,
	)

	c.associationSharesNorm = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, associationsCollectorSubsystem, "shares_normalized"),
		"Normalized shares for the association",
		[]string{"user", "account", "cluster", "partition"},
		nil,
	)

	return c
}

// Name returns the collector name
func (c *AssociationsSimpleCollector) Name() string {
	return "associations"
}

// IsEnabled returns whether this collector is enabled
func (c *AssociationsSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *AssociationsSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *AssociationsSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.associationInfo
	ch <- c.associationCPULimit
	ch <- c.associationMemoryLimit
	ch <- c.associationTimeLimit
	ch <- c.associationPriority
	ch <- c.associationSharesRaw
	ch <- c.associationSharesNorm
}

// Collect implements the Collector interface
func (c *AssociationsSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ch)
}

// collect gathers metrics from SLURM
func (c *AssociationsSimpleCollector) collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get Associations manager from client
	associationsManager := c.client.Associations()
	if associationsManager == nil {
		return fmt.Errorf("associations manager not available")
	}

	// List all associations
	assocList, err := associationsManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list associations")
		return err
	}

	c.logger.WithField("count", len(assocList.Associations)).Info("Collected association entries")

	for _, assoc := range assocList.Associations {
		// Get association details
		user := assoc.User
		if user == "" {
			user = "unknown"
		}

		account := assoc.Account
		if account == "" {
			account = "default"
		}

		cluster := assoc.Cluster
		if cluster == "" {
			cluster = "default"
		}

		partition := assoc.Partition
		if partition == "" {
			partition = "default"
		}

		qos := assoc.DefaultQoS
		if qos == "" {
			qos = "normal"
		}

		// Association info metric
		ch <- prometheus.MustNewConstMetric(
			c.associationInfo,
			prometheus.GaugeValue,
			1,
			user, account, cluster, partition, qos,
		)

		// Resource limits from TRES
		if assoc.MaxTRESPerJob != nil {
			// CPU limit
			if cpuStr, ok := assoc.MaxTRESPerJob["cpu"]; ok {
				var cpuLimit float64
				if _, err := fmt.Sscanf(cpuStr, "%f", &cpuLimit); err == nil && cpuLimit > 0 {
					ch <- prometheus.MustNewConstMetric(
						c.associationCPULimit,
						prometheus.GaugeValue,
						cpuLimit,
						user, account, cluster, partition,
					)
				}
			}

			// Memory limit
			if memStr, ok := assoc.MaxTRESPerJob["mem"]; ok {
				var memLimit float64
				_, _ = fmt.Sscanf(memStr, "%f", &memLimit)
				if memLimit > 0 {
					ch <- prometheus.MustNewConstMetric(
						c.associationMemoryLimit,
						prometheus.GaugeValue,
						memLimit*1024*1024, // Convert MB to bytes
						user, account, cluster, partition,
					)
				}
			}
		}

		if assoc.MaxWallDuration != nil && *assoc.MaxWallDuration > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.associationTimeLimit,
				prometheus.GaugeValue,
				float64(*assoc.MaxWallDuration),
				user, account, cluster, partition,
			)
		}

		// Priority and shares
		if assoc.Priority > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.associationPriority,
				prometheus.GaugeValue,
				float64(assoc.Priority),
				user, account, cluster, partition,
			)
		}

		if assoc.SharesRaw > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.associationSharesRaw,
				prometheus.GaugeValue,
				float64(assoc.SharesRaw),
				user, account, cluster, partition,
			)
		}

		// Note: SharesNorm and usage metrics are not in the Association struct
		// They would need to be obtained from other API calls or calculated
	}

	return nil
}
