package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	slurm "github.com/jontk/slurm-client"
)

const (
	wckeysCollectorSubsystem = "wckeys"
)

type WCKeysCollector struct {
	client  slurm.SlurmClient
	logger  *logrus.Entry
	timeout time.Duration

	// Metrics
	wckeysTotal    *prometheus.Desc
	wckeysActive   *prometheus.Desc
	wckeysJobCount *prometheus.Desc
	wckeysUsage    *prometheus.Desc
	wckeysInfo     *prometheus.Desc
}

// NewWCKeysCollector creates a new WCKeys (Workload Characterization Keys) collector
func NewWCKeysCollector(client slurm.SlurmClient, logger *logrus.Entry, timeout time.Duration) *WCKeysCollector {
	return &WCKeysCollector{
		client:  client,
		logger:  logger.WithField("collector", "wckeys"),
		timeout: timeout,

		wckeysTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, wckeysCollectorSubsystem, "total"),
			"Total number of WCKeys defined",
			[]string{"cluster"},
			nil,
		),
		wckeysActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, wckeysCollectorSubsystem, "active"),
			"Number of active WCKeys",
			[]string{"cluster"},
			nil,
		),
		wckeysJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, wckeysCollectorSubsystem, "job_count"),
			"Number of jobs using this WCKey",
			[]string{"wckey", "user", "cluster"},
			nil,
		),
		wckeysUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, wckeysCollectorSubsystem, "usage_seconds"),
			"Total usage time in seconds for this WCKey",
			[]string{"wckey", "user", "cluster"},
			nil,
		),
		wckeysInfo: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, wckeysCollectorSubsystem, "info"),
			"WCKey information (always 1)",
			[]string{"wckey", "user", "cluster", "flags"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *WCKeysCollector) Name() string {
	return "wckeys_simple"
}

// Describe sends metric descriptions to the channel
func (c *WCKeysCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.wckeysTotal
	ch <- c.wckeysActive
	ch <- c.wckeysJobCount
	ch <- c.wckeysUsage
	ch <- c.wckeysInfo
}

// Collect gathers metrics from SLURM
func (c *WCKeysCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.Debug("Collecting WCKeys metrics")

	// Get cluster name for labels
	clusterName := "default"
	infoManager := c.client.Info()
	if infoManager != nil {
		if info, err := infoManager.Get(ctx); err == nil && info != nil {
			clusterName = info.ClusterName
		}
	}

	// Get WCKeys manager
	wckeysManager := c.client.WCKeys()
	if wckeysManager == nil {
		c.logger.Debug("WCKeys manager not available")
		return nil
	}

	// Get WCKeys information
	wckeys, err := wckeysManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get WCKeys")
		return err
	}

	if wckeys == nil || wckeys.WCKeys == nil {
		return nil
	}

	totalWCKeys := len(wckeys.WCKeys)
	activeWCKeys := 0

	// Process each WCKey
	for _, wckey := range wckeys.WCKeys {
		wckeyName := wckey.Name
		userName := wckey.User
		if userName == "" {
			userName = "all"
		}

		// Flags not available in current WCKey struct
		flags := "none"

		// Export WCKey info
		ch <- prometheus.MustNewConstMetric(
			c.wckeysInfo,
			prometheus.GaugeValue,
			1,
			wckeyName, userName, clusterName, flags,
		)

		activeWCKeys++

		// Job count and usage time not available in current WCKey struct
		// TODO: Update when WCKey struct includes additional fields
	}

	// Export totals
	ch <- prometheus.MustNewConstMetric(
		c.wckeysTotal,
		prometheus.GaugeValue,
		float64(totalWCKeys),
		clusterName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.wckeysActive,
		prometheus.GaugeValue,
		float64(activeWCKeys),
		clusterName,
	)

	c.logger.WithField("wckey_count", totalWCKeys).Debug("WCKeys metrics collected")
	return nil
}

// Close cleans up any resources
func (c *WCKeysCollector) Close() error {
	c.logger.Info("Closing WCKeys collector")
	return nil
}

// IsEnabled returns whether this collector is enabled
func (c *WCKeysCollector) IsEnabled() bool {
	return true
}

// SetEnabled sets whether this collector is enabled
func (c *WCKeysCollector) SetEnabled(enabled bool) {
	// For now, we don't track enabled state in these simple collectors
	// The enabled state is managed by the registry
}
