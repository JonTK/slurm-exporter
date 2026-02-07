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
	clustersCollectorSubsystem = "clusters"
)

type ClustersCollector struct {
	client  slurm.SlurmClient
	logger  *logrus.Entry
	timeout time.Duration

	// Metrics
	clusterInfo           *prometheus.Desc
	clusterRPCVersion     *prometheus.Desc
	clusterPluginInfo     *prometheus.Desc
	clusterFederationInfo *prometheus.Desc
	clusterTRES           *prometheus.Desc
	clusterFeatures       *prometheus.Desc
}

// NewClustersCollector creates a new clusters collector
func NewClustersCollector(client slurm.SlurmClient, logger *logrus.Entry, timeout time.Duration) *ClustersCollector {
	return &ClustersCollector{
		client:  client,
		logger:  logger.WithField("collector", "clusters"),
		timeout: timeout,

		clusterInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clustersCollectorSubsystem, "info"),
			"Cluster information (always 1)",
			[]string{"cluster", "control_host"},
			nil,
		),
		clusterRPCVersion: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clustersCollectorSubsystem, "rpc_version"),
			"RPC version of the cluster",
			[]string{"cluster"},
			nil,
		),
		clusterPluginInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clustersCollectorSubsystem, "plugin_info"),
			"Plugin information (always 1)",
			[]string{"cluster", "plugin_type", "plugin_id"},
			nil,
		),
		clusterFederationInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clustersCollectorSubsystem, "federation_info"),
			"Federation state information (always 1)",
			[]string{"cluster", "federation_state"},
			nil,
		),
		clusterTRES: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clustersCollectorSubsystem, "tres_configured"),
			"Configured TRES resources (always 1)",
			[]string{"cluster", "tres"},
			nil,
		),
		clusterFeatures: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, clustersCollectorSubsystem, "features_configured"),
			"Configured features (always 1)",
			[]string{"cluster", "feature"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *ClustersCollector) Name() string {
	return "clusters_simple"
}

// Describe sends metric descriptions to the channel
func (c *ClustersCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.clusterInfo
	ch <- c.clusterRPCVersion
	ch <- c.clusterPluginInfo
	ch <- c.clusterFederationInfo
	ch <- c.clusterTRES
	ch <- c.clusterFeatures
}

// Collect gathers metrics from SLURM
func (c *ClustersCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.Debug("Collecting cluster metrics")

	// Get clusters manager
	clustersManager := c.client.Clusters()
	if clustersManager == nil {
		c.logger.Debug("Clusters manager not available")
		return nil
	}

	// Get clusters information
	clusters, err := clustersManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get clusters")
		return err
	}

	if clusters == nil || clusters.Clusters == nil {
		return nil
	}

	// Process each cluster
	for _, cluster := range clusters.Clusters {
		// Safely extract cluster name
		clusterName := "unknown"
		if cluster.Name != nil {
			clusterName = *cluster.Name
		}

		// Safely extract control host
		controlHost := ""
		if cluster.Controller != nil && cluster.Controller.Host != nil {
			controlHost = *cluster.Controller.Host
		}

		// Export cluster info
		ch <- prometheus.MustNewConstMetric(
			c.clusterInfo,
			prometheus.GaugeValue,
			1,
			clusterName, controlHost,
		)

		// RPC version
		if cluster.RpcVersion != nil && *cluster.RpcVersion > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.clusterRPCVersion,
				prometheus.GaugeValue,
				float64(*cluster.RpcVersion),
				clusterName,
			)
		}

		// Plugin information (SelectPlugin is now a string, not an ID)
		if cluster.SelectPlugin != nil && *cluster.SelectPlugin != "" {
			ch <- prometheus.MustNewConstMetric(
				c.clusterPluginInfo,
				prometheus.GaugeValue,
				1,
				clusterName, "select", *cluster.SelectPlugin,
			)
		}

		// Note: PluginIDAuth and PluginIDAcct fields no longer exist in the new API
		// Skipping acct plugin metric

		// Note: FederationState field no longer exists in the new API
		// Skipping federation state metric

		// TRES list (field is now TRES, not TRESList)
		for _, tres := range cluster.TRES {
			// TRES is now a complex struct with Type (string) and other fields
			if tres.Type != "" {
				ch <- prometheus.MustNewConstMetric(
					c.clusterTRES,
					prometheus.GaugeValue,
					1,
					clusterName, tres.Type,
				)
			}
		}

		// Note: Features and FederationFeatures fields no longer exist in the new API
		// Skipping cluster features metrics
	}

	c.logger.WithField("cluster_count", len(clusters.Clusters)).Debug("Cluster metrics collected")
	return nil
}

// Close cleans up any resources
func (c *ClustersCollector) Close() error {
	c.logger.Info("Closing clusters collector")
	return nil
}

// IsEnabled returns whether this collector is enabled
func (c *ClustersCollector) IsEnabled() bool {
	return true
}

// SetEnabled sets whether this collector is enabled
func (c *ClustersCollector) SetEnabled(enabled bool) {
	// For now, we don't track enabled state in these simple collectors
	// The enabled state is managed by the registry
}
