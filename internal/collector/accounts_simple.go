// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"

	slurm "github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// AccountsSimpleCollector collects account-related metrics
// TODO: This collector needs to be updated for the new Account API
// The Account struct no longer has ParentAccount, DefaultPartition, MaxTRES, CPULimit, MaxJobs, MaxNodes
// These fields have moved to Association.Max which has a complex nested structure
type AccountsSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool
}

// NewAccountsSimpleCollector creates a new accounts collector
func NewAccountsSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *AccountsSimpleCollector {
	return &AccountsSimpleCollector{
		logger:  logger,
		client:  client,
		enabled: false, // Disabled until API migration complete
	}
}

// Describe implements prometheus.Collector
func (c *AccountsSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	// TODO: Implement after API migration
}

// Collect implements prometheus.Collector
func (c *AccountsSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		c.logger.Debug("Accounts collector disabled - needs API migration")
		return nil
	}
	// TODO: Implement after API migration
	return nil
}

// String returns the name of the collector
func (c *AccountsSimpleCollector) String() string {
	return "accounts_simple"
}
