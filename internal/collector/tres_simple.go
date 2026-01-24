// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	slurm "github.com/jontk/slurm-client"
)

const (
	tresCollectorSubsystem = "tres"
)

type TRESCollector struct {
	client  slurm.SlurmClient
	logger  *logrus.Entry
	timeout time.Duration

	// Metrics
	tresInfo       *prometheus.Desc
	tresCount      *prometheus.Desc
	tresAllocated  *prometheus.Desc
	tresAvailable  *prometheus.Desc
	tresConfigured *prometheus.Desc
}

// NewTRESCollector creates a new TRES (Trackable Resources) collector
func NewTRESCollector(client slurm.SlurmClient, logger *logrus.Entry, timeout time.Duration) *TRESCollector {
	return &TRESCollector{
		client:  client,
		logger:  logger.WithField("collector", "tres"),
		timeout: timeout,

		tresInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, tresCollectorSubsystem, "info"),
			"TRES information (always 1)",
			[]string{"tres_id", "tres_type", "tres_name", "cluster"},
			nil,
		),
		tresCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, tresCollectorSubsystem, "count"),
			"Number of TRES of this type",
			[]string{"tres_type", "tres_name", "cluster"},
			nil,
		),
		tresAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, tresCollectorSubsystem, "allocated"),
			"Amount of TRES currently allocated",
			[]string{"tres_type", "tres_name", "node", "cluster"},
			nil,
		),
		tresAvailable: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, tresCollectorSubsystem, "available"),
			"Amount of TRES available",
			[]string{"tres_type", "tres_name", "node", "cluster"},
			nil,
		),
		tresConfigured: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, tresCollectorSubsystem, "configured"),
			"Amount of TRES configured",
			[]string{"tres_type", "tres_name", "node", "cluster"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *TRESCollector) Name() string {
	return "tres_simple"
}

// Describe sends metric descriptions to the channel
func (c *TRESCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.tresInfo
	ch <- c.tresCount
	ch <- c.tresAllocated
	ch <- c.tresAvailable
	ch <- c.tresConfigured
}

// Collect gathers metrics from SLURM
func (c *TRESCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.Debug("Collecting TRES metrics")

	// Get cluster name for labels
	clusterName := "default"
	infoManager := c.client.Info()
	if infoManager != nil {
		if info, err := infoManager.Get(ctx); err == nil && info != nil {
			clusterName = info.ClusterName
		}
	}

	// Get TRES information
	tresList, err := c.client.GetTRES(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get TRES")
		return err
	}

	if tresList == nil || tresList.TRES == nil {
		return nil
	}

	// Track TRES counts by type
	tresCounts := make(map[string]int)

	// Process each TRES
	for _, tres := range tresList.TRES {
		tresID := fmt.Sprintf("%d", tres.ID)
		tresType := tres.Type
		tresName := tres.Name

		if tresName == "" {
			tresName = tresType // Use type as name if name is empty
		}

		// Export TRES info metric
		ch <- prometheus.MustNewConstMetric(
			c.tresInfo,
			prometheus.GaugeValue,
			1,
			tresID, tresType, tresName, clusterName,
		)

		// Count TRES by type
		key := fmt.Sprintf("%s:%s", tresType, tresName)
		tresCounts[key]++
	}

	// Export TRES counts
	for key, count := range tresCounts {
		var tresType, tresName string
		_, _ = fmt.Sscanf(key, "%s:%s", &tresType, &tresName)

		ch <- prometheus.MustNewConstMetric(
			c.tresCount,
			prometheus.GaugeValue,
			float64(count),
			tresType, tresName, clusterName,
		)
	}

	// Also collect TRES allocation data from nodes if available
	c.collectNodeTRES(ctx, ch, clusterName)

	c.logger.WithField("tres_count", len(tresList.TRES)).Debug("TRES metrics collected")
	return nil
}

func (c *TRESCollector) collectNodeTRES(ctx context.Context, ch chan<- prometheus.Metric, clusterName string) {
	_ = ch
	_ = clusterName
	// Get node information which contains TRES allocation data
	nodeManager := c.client.Nodes()
	if nodeManager == nil {
		return
	}
	nodes, err := nodeManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Debug("Failed to get nodes for TRES data")
		return
	}

	if nodes == nil || nodes.Nodes == nil {
		return
	}

	// Node-level TRES data not available in current Node struct
	// TODO: Update when Node struct includes TRES fields
	/*
		for _, node := range nodes.Nodes {
			nodeName := node.Name

			// Process allocated TRES
			if node.TRESAllocated != nil {
				for tresType, value := range node.TRESAllocated {
					ch <- prometheus.MustNewConstMetric(
						c.tresAllocated,
						prometheus.GaugeValue,
						float64(value),
						tresType, tresType, nodeName, clusterName,
					)
				}
			}

			// Process configured TRES
			if node.TRESConfigured != nil {
				for tresType, value := range node.TRESConfigured {
					ch <- prometheus.MustNewConstMetric(
						c.tresConfigured,
						prometheus.GaugeValue,
						float64(value),
						tresType, tresType, nodeName, clusterName,
					)
				}
			}

			// Calculate available TRES (configured - allocated)
			if node.TRESConfigured != nil && node.TRESAllocated != nil {
				for tresType, configured := range node.TRESConfigured {
					allocated := node.TRESAllocated[tresType]
					available := configured - allocated

					ch <- prometheus.MustNewConstMetric(
						c.tresAvailable,
						prometheus.GaugeValue,
						float64(available),
						tresType, tresType, nodeName, clusterName,
					)
				}
			}
		}
	*/
}

// Close cleans up any resources
func (c *TRESCollector) Close() error {
	c.logger.Info("Closing TRES collector")
	return nil
}

// IsEnabled returns whether this collector is enabled
func (c *TRESCollector) IsEnabled() bool {
	return true
}

// SetEnabled sets whether this collector is enabled
func (c *TRESCollector) SetEnabled(enabled bool) {
	// For now, we don't track enabled state in these simple collectors
	// The enabled state is managed by the registry
}
