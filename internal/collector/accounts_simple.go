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
		// Account info metric
		organization := account.Organization
		if organization == "" {
			organization = "default"
		}

		description := account.Description
		if description == "" {
			description = "No description"
		}

		parentAccount := account.ParentAccount
		if parentAccount == "" {
			parentAccount = "root"
		}

		ch <- prometheus.MustNewConstMetric(
			c.accountInfo,
			prometheus.GaugeValue,
			1,
			account.Name, organization, description, parentAccount,
		)

		// Number of users in account
		userCount := float64(userCounts[account.Name])
		ch <- prometheus.MustNewConstMetric(
			c.accountUsers,
			prometheus.GaugeValue,
			userCount,
			account.Name,
		)

		// Resource limits (account-wide, not per partition)
		defaultPartition := account.DefaultPartition
		if defaultPartition == "" {
			defaultPartition = "default"
		}

		// CPU limit
		if account.CPULimit > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.accountCPULimit,
				prometheus.GaugeValue,
				float64(account.CPULimit),
				account.Name, defaultPartition,
			)
		}

		// Job limit
		if account.MaxJobs > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.accountJobLimit,
				prometheus.GaugeValue,
				float64(account.MaxJobs),
				account.Name, defaultPartition,
			)
		}

		// Node limit
		if account.MaxNodes > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.accountNodeLimit,
				prometheus.GaugeValue,
				float64(account.MaxNodes),
				account.Name, defaultPartition,
			)
		}

		// Check TRES limits for memory
		if account.MaxTRES != nil {
			if memLimit, ok := account.MaxTRES["mem"]; ok && memLimit > 0 {
				ch <- prometheus.MustNewConstMetric(
					c.accountMemoryLimit,
					prometheus.GaugeValue,
					float64(memLimit)*1024*1024, // Convert MB to bytes
					account.Name, defaultPartition,
				)
			}
		}
	}

	return nil
}
