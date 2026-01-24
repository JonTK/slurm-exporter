// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"

	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-client/interfaces"
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
	return c.collect(ctx, ch)
}

// associationContext holds normalized association information
type associationContext struct {
	user      string
	account   string
	cluster   string
	partition string
	qos       string
}

// extractAssociationContext extracts and normalizes association fields with safe defaults
func extractAssociationContext(assoc *interfaces.Association) associationContext {
	ctx := associationContext{
		user:      assoc.User,
		account:   assoc.Account,
		cluster:   assoc.Cluster,
		partition: assoc.Partition,
		qos:       assoc.DefaultQoS,
	}

	// Apply safe defaults
	if ctx.user == "" {
		ctx.user = "unknown"
	}
	if ctx.account == "" {
		ctx.account = "default"
	}
	if ctx.cluster == "" {
		ctx.cluster = "default"
	}
	if ctx.partition == "" {
		ctx.partition = "default"
	}
	if ctx.qos == "" {
		ctx.qos = "normal"
	}

	return ctx
}

// sendAssociationInfoMetric sends the association info metric
func (c *AssociationsSimpleCollector) sendAssociationInfoMetric(ch chan<- prometheus.Metric, ctx associationContext) {
	ch <- prometheus.MustNewConstMetric(
		c.associationInfo,
		prometheus.GaugeValue,
		1,
		ctx.user, ctx.account, ctx.cluster, ctx.partition, ctx.qos,
	)
}

// sendTRESLimitsMetrics sends CPU and memory limit metrics from TRES
func (c *AssociationsSimpleCollector) sendTRESLimitsMetrics(ch chan<- prometheus.Metric, assoc *interfaces.Association, ctx associationContext) {
	if assoc.MaxTRESPerJob == nil {
		return
	}

	// CPU limit
	if cpuStr, ok := assoc.MaxTRESPerJob["cpu"]; ok {
		var cpuLimit float64
		if _, err := fmt.Sscanf(cpuStr, "%f", &cpuLimit); err == nil && cpuLimit > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.associationCPULimit,
				prometheus.GaugeValue,
				cpuLimit,
				ctx.user, ctx.account, ctx.cluster, ctx.partition,
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
				ctx.user, ctx.account, ctx.cluster, ctx.partition,
			)
		}
	}
}

// sendAssociationLimitMetrics sends time limit, priority, and shares metrics
func (c *AssociationsSimpleCollector) sendAssociationLimitMetrics(ch chan<- prometheus.Metric, assoc *interfaces.Association, ctx associationContext) {
	// Time limit
	if assoc.MaxWallDuration != nil && *assoc.MaxWallDuration > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.associationTimeLimit,
			prometheus.GaugeValue,
			float64(*assoc.MaxWallDuration),
			ctx.user, ctx.account, ctx.cluster, ctx.partition,
		)
	}

	// Priority
	if assoc.Priority > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.associationPriority,
			prometheus.GaugeValue,
			float64(assoc.Priority),
			ctx.user, ctx.account, ctx.cluster, ctx.partition,
		)
	}

	// Shares
	if assoc.SharesRaw > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.associationSharesRaw,
			prometheus.GaugeValue,
			float64(assoc.SharesRaw),
			ctx.user, ctx.account, ctx.cluster, ctx.partition,
		)
	}
}

// collect gathers metrics from SLURM
func (c *AssociationsSimpleCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
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
		// Extract normalized association context
		assocCtx := extractAssociationContext(assoc)

		// Send association info metric
		c.sendAssociationInfoMetric(ch, assocCtx)

		// Send TRES resource limits
		c.sendTRESLimitsMetrics(ch, assoc, assocCtx)

		// Send other limit metrics (time, priority, shares)
		c.sendAssociationLimitMetrics(ch, assoc, assocCtx)
	}

	return nil
}
