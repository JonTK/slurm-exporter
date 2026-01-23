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
	sharesCollectorSubsystem = "shares"
)

type SharesCollector struct {
	client  slurm.SlurmClient
	logger  *logrus.Entry
	timeout time.Duration

	// Metrics
	fairshareRawShares        *prometheus.Desc
	fairshareNormalizedShares *prometheus.Desc
	fairshareRawUsage         *prometheus.Desc
	fairshareEffectiveUsage   *prometheus.Desc
	fairshareUsageFactor      *prometheus.Desc
	fairshareLevel            *prometheus.Desc
	fairshareTreeDepth        *prometheus.Desc
}

// NewSharesCollector creates a new shares (fairshare) collector
func NewSharesCollector(client slurm.SlurmClient, logger *logrus.Entry, timeout time.Duration) *SharesCollector {
	return &SharesCollector{
		client:  client,
		logger:  logger.WithField("collector", "shares"),
		timeout: timeout,

		fairshareRawShares: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, sharesCollectorSubsystem, "raw_shares"),
			"Raw share value for the user/account",
			[]string{"user", "account", "partition", "cluster"},
			nil,
		),
		fairshareNormalizedShares: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, sharesCollectorSubsystem, "normalized_shares"),
			"Normalized share value (0-1)",
			[]string{"user", "account", "partition", "cluster"},
			nil,
		),
		fairshareRawUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, sharesCollectorSubsystem, "raw_usage"),
			"Raw usage value",
			[]string{"user", "account", "partition", "cluster"},
			nil,
		),
		fairshareEffectiveUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, sharesCollectorSubsystem, "effective_usage"),
			"Effective usage after decay (0-1)",
			[]string{"user", "account", "partition", "cluster"},
			nil,
		),
		fairshareUsageFactor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, sharesCollectorSubsystem, "usage_factor"),
			"Fairshare factor (0-1, higher is better priority)",
			[]string{"user", "account", "partition", "cluster"},
			nil,
		),
		fairshareLevel: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, sharesCollectorSubsystem, "level"),
			"Level in the fairshare tree",
			[]string{"user", "account", "partition", "cluster"},
			nil,
		),
		fairshareTreeDepth: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, sharesCollectorSubsystem, "tree_depth"),
			"Maximum depth of the fairshare tree",
			[]string{"partition", "cluster"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *SharesCollector) Name() string {
	return "shares_simple"
}

// Describe sends metric descriptions to the channel
func (c *SharesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.fairshareRawShares
	ch <- c.fairshareNormalizedShares
	ch <- c.fairshareRawUsage
	ch <- c.fairshareEffectiveUsage
	ch <- c.fairshareUsageFactor
	ch <- c.fairshareLevel
	ch <- c.fairshareTreeDepth
}

// Collect gathers metrics from SLURM
func (c *SharesCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.Debug("Collecting fairshare metrics")

	// Get shares information
	shares, err := c.client.GetShares(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get shares")
		return err
	}

	// Get cluster name for labels
	clusterName := "default"
	infoManager := c.client.Info()
	if infoManager != nil {
		if info, err := infoManager.Get(ctx); err == nil && info != nil {
			clusterName = info.ClusterName
		}
	}

	// Process shares data
	if shares != nil && shares.Shares != nil {
		// Track tree depth per partition
		depthMap := make(map[string]uint32)

		for _, share := range shares.Shares {
			user := share.User
			if user == "" {
				user = "root"
			}

			account := share.Account
			if account == "" {
				account = "root"
			}

			partition := share.Partition
			if partition == "" {
				partition = "all"
			}

			// Track maximum depth
			if uint32(share.Level) > depthMap[partition] {
				depthMap[partition] = uint32(share.Level)
			}

			// Raw shares
			ch <- prometheus.MustNewConstMetric(
				c.fairshareRawShares,
				prometheus.GaugeValue,
				float64(share.RawShares),
				user, account, partition, clusterName,
			)

			// Normalized shares
			ch <- prometheus.MustNewConstMetric(
				c.fairshareNormalizedShares,
				prometheus.GaugeValue,
				share.NormShares,
				user, account, partition, clusterName,
			)

			// Raw usage
			ch <- prometheus.MustNewConstMetric(
				c.fairshareRawUsage,
				prometheus.GaugeValue,
				float64(share.RawUsage),
				user, account, partition, clusterName,
			)

			// Effective usage
			ch <- prometheus.MustNewConstMetric(
				c.fairshareEffectiveUsage,
				prometheus.GaugeValue,
				share.EffectUsage,
				user, account, partition, clusterName,
			)

			// Usage factor
			ch <- prometheus.MustNewConstMetric(
				c.fairshareUsageFactor,
				prometheus.GaugeValue,
				share.FairShare,
				user, account, partition, clusterName,
			)

			// Level in tree
			ch <- prometheus.MustNewConstMetric(
				c.fairshareLevel,
				prometheus.GaugeValue,
				float64(share.Level),
				user, account, partition, clusterName,
			)
		}

		// Export tree depth metrics
		for partition, depth := range depthMap {
			ch <- prometheus.MustNewConstMetric(
				c.fairshareTreeDepth,
				prometheus.GaugeValue,
				float64(depth),
				partition, clusterName,
			)
		}
	}

	c.logger.WithField("share_count", len(shares.Shares)).Debug("Fairshare metrics collected")
	return nil
}

// Close cleans up any resources
func (c *SharesCollector) Close() error {
	c.logger.Info("Closing shares collector")
	return nil
}

// IsEnabled returns whether this collector is enabled
func (c *SharesCollector) IsEnabled() bool {
	return true
}

// SetEnabled sets whether this collector is enabled
func (c *SharesCollector) SetEnabled(enabled bool) {
	// For now, we don't track enabled state in these simple collectors
	// The enabled state is managed by the registry
}
