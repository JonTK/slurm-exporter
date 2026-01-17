package collector

import (
	"context"
	"fmt"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	clusterCollectorSubsystem = "cluster"
)

// ClusterSimpleCollector collects cluster-wide metrics
type ClusterSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Cluster info metrics
	clusterInfo        *prometheus.Desc
	clusterNodes       *prometheus.Desc
	clusterCPUs        *prometheus.Desc
	clusterJobs        *prometheus.Desc
	clusterVersion     *prometheus.Desc
	clusterControllers *prometheus.Desc
}

// NewClusterSimpleCollector creates a new Cluster collector
func NewClusterSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *ClusterSimpleCollector {
	c := &ClusterSimpleCollector{
		client:  client,
		logger:  logger.WithField("collector", "cluster"),
		enabled: true,
	}

	// Initialize metrics
	c.clusterInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "info"),
		"Cluster information with all labels",
		[]string{"cluster", "control_host", "control_port", "rpc_version", "plugin_version"},
		nil,
	)

	c.clusterNodes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "nodes_total"),
		"Total number of nodes in the cluster",
		[]string{"cluster"},
		nil,
	)

	c.clusterCPUs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "cpus_total"),
		"Total number of CPUs in the cluster",
		[]string{"cluster"},
		nil,
	)

	c.clusterJobs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "jobs_total"),
		"Total number of jobs in the cluster",
		[]string{"cluster"},
		nil,
	)

	c.clusterVersion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "version_info"),
		"SLURM version information",
		[]string{"cluster", "version", "major", "minor", "patch"},
		nil,
	)

	c.clusterControllers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, clusterCollectorSubsystem, "controllers_total"),
		"Number of controllers in the cluster",
		[]string{"cluster"},
		nil,
	)

	return c
}

// Name returns the collector name
func (c *ClusterSimpleCollector) Name() string {
	return "cluster"
}

// IsEnabled returns whether this collector is enabled
func (c *ClusterSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *ClusterSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *ClusterSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.clusterInfo
	ch <- c.clusterNodes
	ch <- c.clusterCPUs
	ch <- c.clusterJobs
	ch <- c.clusterVersion
	ch <- c.clusterControllers
}

// Collect implements the Collector interface
func (c *ClusterSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ch)
}

// collect gathers metrics from SLURM
func (c *ClusterSimpleCollector) collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get Info manager from client
	infoManager := c.client.Info()
	if infoManager == nil {
		return fmt.Errorf("info manager not available")
	}

	// Get cluster information
	info, err := infoManager.Get(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get cluster info")
		return err
	}
	if info == nil {
		err := fmt.Errorf("cluster info is nil")
		c.logger.WithError(err).Error("Invalid cluster info response")
		return err
	}

	// Also get cluster stats
	stats, err := infoManager.Stats(ctx)
	if err != nil {
		c.logger.WithError(err).Warn("Failed to get cluster stats, continuing with info only")
		// Continue with info only
	}

	c.logger.WithField("cluster", info.ClusterName).Info("Collected cluster info")

	// Extract cluster name with default
	clusterName := "default"
	if info.ClusterName != "" {
		clusterName = info.ClusterName
	}

	// Control information (use defaults as this info isn't in the interface)
	controlHost := "localhost"
	controlPort := "6817"

	// Version information
	version := info.Version
	if version == "" {
		version = "unknown"
	}

	// Parse version components (simple parsing)
	// Version might be like "23.02.1"
	var maj, min, pat int
	if _, err := fmt.Sscanf(info.Version, "%d.%d.%d", &maj, &min, &pat); err != nil {
		// If parsing fails, use defaults
		maj, min, pat = 0, 0, 0
	}
	major := fmt.Sprintf("%d", maj)
	minor := fmt.Sprintf("%d", min)
	patch := fmt.Sprintf("%d", pat)

	// RPC/API version
	rpcVersion := info.APIVersion
	if rpcVersion == "" {
		rpcVersion = "unknown"
	}
	pluginVersion := info.Release

	// Cluster info metric
	ch <- prometheus.MustNewConstMetric(
		c.clusterInfo,
		prometheus.GaugeValue,
		1,
		clusterName, controlHost, controlPort, rpcVersion, pluginVersion,
	)

	// Total nodes (if available in stats)
	totalNodes := float64(0)
	if stats != nil {
		totalNodes = float64(stats.TotalNodes)
	}
	ch <- prometheus.MustNewConstMetric(
		c.clusterNodes,
		prometheus.GaugeValue,
		totalNodes,
		clusterName,
	)

	// Total CPUs
	totalCPUs := float64(0)
	if stats != nil {
		totalCPUs = float64(stats.TotalCPUs)
	}
	ch <- prometheus.MustNewConstMetric(
		c.clusterCPUs,
		prometheus.GaugeValue,
		totalCPUs,
		clusterName,
	)

	// Total jobs
	totalJobs := float64(0)
	if stats != nil {
		totalJobs = float64(stats.TotalJobs)
	}
	ch <- prometheus.MustNewConstMetric(
		c.clusterJobs,
		prometheus.GaugeValue,
		totalJobs,
		clusterName,
	)

	// Version info
	ch <- prometheus.MustNewConstMetric(
		c.clusterVersion,
		prometheus.GaugeValue,
		1,
		clusterName, version, major, minor, patch,
	)

	// Number of controllers (default to 1 since we don't have this info)
	ch <- prometheus.MustNewConstMetric(
		c.clusterControllers,
		prometheus.GaugeValue,
		float64(1),
		clusterName,
	)

	return nil
}
