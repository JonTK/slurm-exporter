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
	licensesCollectorSubsystem = "licenses"
)

type LicensesCollector struct {
	client  slurm.SlurmClient
	logger  *logrus.Entry
	timeout time.Duration

	// Metrics
	licensesTotal     *prometheus.Desc
	licensesUsed      *prometheus.Desc
	licensesAvailable *prometheus.Desc
	licensesFree      *prometheus.Desc
	licensesReserved  *prometheus.Desc
}

// NewLicensesCollector creates a new licenses collector
func NewLicensesCollector(client slurm.SlurmClient, logger *logrus.Entry, timeout time.Duration) *LicensesCollector {
	return &LicensesCollector{
		client:  client,
		logger:  logger.WithField("collector", "licenses"),
		timeout: timeout,

		licensesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, licensesCollectorSubsystem, "total"),
			"Total number of licenses for a feature",
			[]string{"feature", "cluster"},
			nil,
		),
		licensesUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, licensesCollectorSubsystem, "used"),
			"Number of licenses currently in use",
			[]string{"feature", "cluster"},
			nil,
		),
		licensesAvailable: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, licensesCollectorSubsystem, "available"),
			"Number of licenses available",
			[]string{"feature", "cluster"},
			nil,
		),
		licensesFree: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, licensesCollectorSubsystem, "free"),
			"Number of free licenses",
			[]string{"feature", "cluster"},
			nil,
		),
		licensesReserved: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, licensesCollectorSubsystem, "reserved"),
			"Number of reserved licenses",
			[]string{"feature", "cluster"},
			nil,
		),
	}
}

// Name returns the collector name
func (c *LicensesCollector) Name() string {
	return "licenses_simple"
}

// Describe sends metric descriptions to the channel
func (c *LicensesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.licensesTotal
	ch <- c.licensesUsed
	ch <- c.licensesAvailable
	ch <- c.licensesFree
	ch <- c.licensesReserved
}

// Collect gathers metrics from SLURM
func (c *LicensesCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	c.logger.Debug("Collecting license metrics")

	// Get licenses information
	licenses, err := c.client.GetLicenses(timeoutCtx)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get licenses")
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

	// Process each license
	if licenses != nil && licenses.Licenses != nil {
		for _, license := range licenses.Licenses {
			feature := license.Name
			if feature == "" {
				continue
			}

			// Total licenses
			ch <- prometheus.MustNewConstMetric(
				c.licensesTotal,
				prometheus.GaugeValue,
				float64(license.Total),
				feature, clusterName,
			)

			// Used licenses
			ch <- prometheus.MustNewConstMetric(
				c.licensesUsed,
				prometheus.GaugeValue,
				float64(license.Used),
				feature, clusterName,
			)

			// Available licenses
			ch <- prometheus.MustNewConstMetric(
				c.licensesAvailable,
				prometheus.GaugeValue,
				float64(license.Free),
				feature, clusterName,
			)

			// Free licenses (same as available)
			ch <- prometheus.MustNewConstMetric(
				c.licensesFree,
				prometheus.GaugeValue,
				float64(license.Free),
				feature, clusterName,
			)

			// Reserved licenses
			ch <- prometheus.MustNewConstMetric(
				c.licensesReserved,
				prometheus.GaugeValue,
				float64(license.Reserved),
				feature, clusterName,
			)
		}
	}

	c.logger.WithField("license_count", len(licenses.Licenses)).Debug("License metrics collected")
	return nil
}

// Close cleans up any resources
func (c *LicensesCollector) Close() error {
	c.logger.Info("Closing licenses collector")
	return nil
}

// IsEnabled returns whether this collector is enabled
func (c *LicensesCollector) IsEnabled() bool {
	return true
}

// SetEnabled sets whether this collector is enabled
func (c *LicensesCollector) SetEnabled(enabled bool) {
	// For now, we don't track enabled state in these simple collectors
	// The enabled state is managed by the registry
}
