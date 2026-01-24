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
	accountsCollectorSubsystem = "account"
)

// AccountsSimpleCollector collects account-related metrics
type AccountsSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Account metrics
	accountInfo        *prometheus.Desc
	accountUsers       *prometheus.Desc
	accountCPULimit    *prometheus.Desc
	accountMemoryLimit *prometheus.Desc
	accountJobLimit    *prometheus.Desc
	accountNodeLimit   *prometheus.Desc
}

// NewAccountsSimpleCollector creates a new Accounts collector
func NewAccountsSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *AccountsSimpleCollector {
	c := &AccountsSimpleCollector{
		client:  client,
		logger:  logger.WithField("collector", "accounts"),
		enabled: true,
	}

	// Initialize metrics
	c.accountInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, accountsCollectorSubsystem, "info"),
		"Account information with all labels",
		[]string{"account", "organization", "description", "parent_account"},
		nil,
	)

	c.accountUsers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, accountsCollectorSubsystem, "users_total"),
		"Total number of users in the account",
		[]string{"account"},
		nil,
	)

	c.accountCPULimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, accountsCollectorSubsystem, "cpu_limit"),
		"CPU limit for the account",
		[]string{"account", "partition"},
		nil,
	)

	c.accountMemoryLimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, accountsCollectorSubsystem, "memory_limit_bytes"),
		"Memory limit for the account in bytes",
		[]string{"account", "partition"},
		nil,
	)

	c.accountJobLimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, accountsCollectorSubsystem, "job_limit"),
		"Job limit for the account",
		[]string{"account", "partition"},
		nil,
	)

	c.accountNodeLimit = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, accountsCollectorSubsystem, "node_limit"),
		"Node limit for the account",
		[]string{"account", "partition"},
		nil,
	)

	return c
}

// Name returns the collector name
func (c *AccountsSimpleCollector) Name() string {
	return "accounts"
}

// IsEnabled returns whether this collector is enabled
func (c *AccountsSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *AccountsSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *AccountsSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.accountInfo
	ch <- c.accountUsers
	ch <- c.accountCPULimit
	ch <- c.accountMemoryLimit
	ch <- c.accountJobLimit
	ch <- c.accountNodeLimit
}

// Collect implements the Collector interface
func (c *AccountsSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ch)
}

// accountContext holds normalized account information
type accountContext struct {
	name             string
	organization     string
	description      string
	parentAccount    string
	defaultPartition string
}

// extractAccountContext extracts and normalizes account fields with safe defaults
func extractAccountContext(account slurm.Account) accountContext {
	ctx := accountContext{
		name:             account.Name,
		organization:     account.Organization,
		description:      account.Description,
		parentAccount:    account.ParentAccount,
		defaultPartition: account.DefaultPartition,
	}

	// Apply safe defaults
	if ctx.organization == "" {
		ctx.organization = "default"
	}
	if ctx.description == "" {
		ctx.description = "No description"
	}
	if ctx.parentAccount == "" {
		ctx.parentAccount = "root"
	}
	if ctx.defaultPartition == "" {
		ctx.defaultPartition = "default"
	}

	return ctx
}

// sendAccountInfoMetric sends the account info metric
func (c *AccountsSimpleCollector) sendAccountInfoMetric(ch chan<- prometheus.Metric, ctx accountContext) {
	ch <- prometheus.MustNewConstMetric(
		c.accountInfo,
		prometheus.GaugeValue,
		1,
		ctx.name, ctx.organization, ctx.description, ctx.parentAccount,
	)
}

// sendAccountUserCountMetric sends the user count metric
func (c *AccountsSimpleCollector) sendAccountUserCountMetric(ch chan<- prometheus.Metric, accountName string, userCount int) {
	ch <- prometheus.MustNewConstMetric(
		c.accountUsers,
		prometheus.GaugeValue,
		float64(userCount),
		accountName,
	)
}

// sendAccountLimitMetric sends a resource limit metric if the limit is greater than zero
func (c *AccountsSimpleCollector) sendAccountLimitMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, limit int, accountName, partition string) {
	if limit > 0 {
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(limit),
			accountName, partition,
		)
	}
}

// sendAccountMemoryLimit sends memory limit metric from TRES if available
func (c *AccountsSimpleCollector) sendAccountMemoryLimit(ch chan<- prometheus.Metric, account slurm.Account, partition string) {
	if account.MaxTRES != nil {
		if memLimit, ok := account.MaxTRES["mem"]; ok && memLimit > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.accountMemoryLimit,
				prometheus.GaugeValue,
				float64(memLimit)*1024*1024, // Convert MB to bytes
				account.Name, partition,
			)
		}
	}
}

// collect gathers metrics from SLURM
func (c *AccountsSimpleCollector) collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get Accounts manager from client
	accountsManager := c.client.Accounts()
	if accountsManager == nil {
		return fmt.Errorf("accounts manager not available")
	}

	// List all accounts
	accountList, err := accountsManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list accounts")
		return err
	}

	c.logger.WithField("count", len(accountList.Accounts)).Info("Collected account entries")

	// Also get associations to count users per account
	associationsManager := c.client.Associations()
	userCounts := make(map[string]int)
	if associationsManager != nil {
		assocList, err := associationsManager.List(ctx, nil)
		if err == nil {
			for _, assoc := range assocList.Associations {
				if assoc.Account != "" {
					userCounts[assoc.Account]++
				}
			}
		}
	}

	for _, account := range accountList.Accounts {
		// Extract normalized account context
		accCtx := extractAccountContext(account)

		// Send account info metric
		c.sendAccountInfoMetric(ch, accCtx)

		// Send user count metric
		c.sendAccountUserCountMetric(ch, account.Name, userCounts[account.Name])

		// Send resource limit metrics
		c.sendAccountLimitMetric(ch, c.accountCPULimit, account.CPULimit, account.Name, accCtx.defaultPartition)
		c.sendAccountLimitMetric(ch, c.accountJobLimit, account.MaxJobs, account.Name, accCtx.defaultPartition)
		c.sendAccountLimitMetric(ch, c.accountNodeLimit, account.MaxNodes, account.Name, accCtx.defaultPartition)

		// Send memory limit from TRES
		c.sendAccountMemoryLimit(ch, account, accCtx.defaultPartition)
	}

	return nil
}
