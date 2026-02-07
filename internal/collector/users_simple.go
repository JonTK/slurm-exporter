// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"

	slurm "github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	usersCollectorSubsystem = "user"
)

// UsersSimpleCollector collects user-related metrics
type UsersSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// User metrics
	userInfo         *prometheus.Desc
	userJobsRunning  *prometheus.Desc
	userJobsPending  *prometheus.Desc
	userCPUsUsed     *prometheus.Desc
	userMemoryUsed   *prometheus.Desc
	userAssociations *prometheus.Desc
}

// NewUsersSimpleCollector creates a new Users collector
func NewUsersSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *UsersSimpleCollector {
	c := &UsersSimpleCollector{
		client:  client,
		logger:  logger.WithField("collector", "users"),
		enabled: true,
	}

	// Initialize metrics
	c.userInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, usersCollectorSubsystem, "info"),
		"User information with all labels",
		[]string{"user", "default_account", "admin_level"},
		nil,
	)

	c.userJobsRunning = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, usersCollectorSubsystem, "jobs_running"),
		"Number of running jobs for the user",
		[]string{"user", "account", "partition"},
		nil,
	)

	c.userJobsPending = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, usersCollectorSubsystem, "jobs_pending"),
		"Number of pending jobs for the user",
		[]string{"user", "account", "partition"},
		nil,
	)

	c.userCPUsUsed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, usersCollectorSubsystem, "cpus_used"),
		"Number of CPUs currently used by the user",
		[]string{"user", "account", "partition"},
		nil,
	)

	c.userMemoryUsed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, usersCollectorSubsystem, "memory_used_bytes"),
		"Memory currently used by the user in bytes",
		[]string{"user", "account", "partition"},
		nil,
	)

	c.userAssociations = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, usersCollectorSubsystem, "associations_total"),
		"Number of associations for the user",
		[]string{"user"},
		nil,
	)

	return c
}

// Name returns the collector name
func (c *UsersSimpleCollector) Name() string {
	return "users"
}

// IsEnabled returns whether this collector is enabled
func (c *UsersSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *UsersSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *UsersSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.userInfo
	ch <- c.userJobsRunning
	ch <- c.userJobsPending
	ch <- c.userCPUsUsed
	ch <- c.userMemoryUsed
	ch <- c.userAssociations
}

// Collect implements the Collector interface
func (c *UsersSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ctx, ch)
}

// collect gathers metrics from SLURM
func (c *UsersSimpleCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Get Users manager from client
	usersManager := c.client.Users()
	if usersManager == nil {
		return fmt.Errorf("users manager not available")
	}

	// List all users
	userList, err := usersManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list users")
		return err
	}

	c.logger.WithField("count", len(userList.Users)).Info("Collected user entries")

	// Also collect job statistics per user
	jobsManager := c.client.Jobs()
	var jobStats map[string]userJobStats
	if jobsManager != nil {
		jobStats = c.collectJobStatsByUser(ctx, jobsManager)
	}

	for _, user := range userList.Users {
		// User info metric
		// AdminLevel is now AdministratorLevel (array of enums)
		adminLevel := "none"
		if len(user.AdministratorLevel) > 0 {
			adminLevel = string(user.AdministratorLevel[0])
		}

		// DefaultAccount is now nested under Default.Account
		defaultAccount := "default"
		if user.Default != nil && user.Default.Account != nil {
			defaultAccount = *user.Default.Account
		}

		ch <- prometheus.MustNewConstMetric(
			c.userInfo,
			prometheus.GaugeValue,
			1,
			user.Name, defaultAccount, adminLevel,
		)

		// Number of associations
		associationCount := float64(len(user.Associations))
		ch <- prometheus.MustNewConstMetric(
			c.userAssociations,
			prometheus.GaugeValue,
			associationCount,
			user.Name,
		)

		// Add job statistics if available
		if stats, ok := jobStats[user.Name]; ok {
			for key, count := range stats.runningJobs {
				ch <- prometheus.MustNewConstMetric(
					c.userJobsRunning,
					prometheus.GaugeValue,
					float64(count),
					user.Name, key.account, key.partition,
				)
			}

			for key, count := range stats.pendingJobs {
				ch <- prometheus.MustNewConstMetric(
					c.userJobsPending,
					prometheus.GaugeValue,
					float64(count),
					user.Name, key.account, key.partition,
				)
			}

			for key, cpus := range stats.cpusUsed {
				ch <- prometheus.MustNewConstMetric(
					c.userCPUsUsed,
					prometheus.GaugeValue,
					float64(cpus),
					user.Name, key.account, key.partition,
				)
			}

			for key, memory := range stats.memoryUsed {
				ch <- prometheus.MustNewConstMetric(
					c.userMemoryUsed,
					prometheus.GaugeValue,
					float64(memory),
					user.Name, key.account, key.partition,
				)
			}
		}
	}

	return nil
}

// userJobKey represents a unique user/account/partition combination
type userJobKey struct {
	account   string
	partition string
}

// userJobStats holds job statistics for a user
type userJobStats struct {
	runningJobs map[userJobKey]int
	pendingJobs map[userJobKey]int
	cpusUsed    map[userJobKey]int
	memoryUsed  map[userJobKey]int64
}

// collectJobStatsByUser collects job statistics grouped by user
func (c *UsersSimpleCollector) collectJobStatsByUser(ctx context.Context, jobsManager slurm.JobManager) map[string]userJobStats {
	stats := make(map[string]userJobStats)
	// List all jobs
	jobList, err := jobsManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Warn("Failed to list jobs for user stats")
		return stats
	}

	for _, job := range jobList.Jobs {
		// UserID is now a pointer
		userName := ""
		if job.UserID != nil {
			userName = fmt.Sprintf("%d", *job.UserID)
		}
		if userName == "" {
			continue
		}

		// Initialize user stats if not exists
		if _, ok := stats[userName]; !ok {
			stats[userName] = userJobStats{
				runningJobs: make(map[userJobKey]int),
				pendingJobs: make(map[userJobKey]int),
				cpusUsed:    make(map[userJobKey]int),
				memoryUsed:  make(map[userJobKey]int64),
			}
		}

		// Partition is now a pointer
		partition := ""
		if job.Partition != nil {
			partition = *job.Partition
		}

		// Create key for grouping
		key := userJobKey{
			account:   "default", // Job doesn't have account field in slurm.Job
			partition: partition,
		}

		// State is now an array of JobState enums
		state := ""
		if len(job.JobState) > 0 {
			state = string(job.JobState[0])
		}

		// Count jobs by state
		switch state {
		case "RUNNING", "COMPLETING":
			stats[userName].runningJobs[key]++

			// CPU/memory tracking would require parsing TRES, skip for now
			// TODO: Add TRES parsing when needed

		case "PENDING":
			stats[userName].pendingJobs[key]++
		}
	}

	return stats
}
