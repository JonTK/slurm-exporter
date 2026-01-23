// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

// UserCollector collects user and account-level metrics
type UserCollector struct {
	*BaseCollector
	metrics      *metrics.MetricDefinitions
	clusterName  string
	errorBuilder *ErrorBuilder
}

// NewUserCollector creates a new user collector
func NewUserCollector(
	config *config.CollectorConfig,
	opts *CollectorOptions,
	client interface{}, // Will be *slurm.Client when slurm-client build issues are fixed
	metrics *metrics.MetricDefinitions,
	clusterName string,
) *UserCollector {
	base := NewBaseCollector("user", config, opts, client, &CollectorMetrics{
		Total:    metrics.CollectionSuccess,
		Errors:   metrics.CollectionErrors,
		Duration: metrics.CollectionDuration,
		Up:       metrics.CollectorUp,
	}, nil)

	return &UserCollector{
		BaseCollector: base,
		metrics:       metrics,
		clusterName:   clusterName,
		errorBuilder:  NewErrorBuilder("user", base.logger),
	}
}

// Describe implements the Collector interface
func (uc *UserCollector) Describe(ch chan<- *prometheus.Desc) {
	// User metrics
	uc.metrics.UserJobCount.Describe(ch)
	uc.metrics.UserCPUAllocated.Describe(ch)
	uc.metrics.UserMemoryAllocated.Describe(ch)
	uc.metrics.UserJobStates.Describe(ch)

	// Account metrics
	uc.metrics.AccountInfo.Describe(ch)
	uc.metrics.AccountJobCount.Describe(ch)
	uc.metrics.AccountCPUAllocated.Describe(ch)
	uc.metrics.AccountMemoryAllocated.Describe(ch)
	uc.metrics.AccountFairShare.Describe(ch)
	uc.metrics.AccountUsage.Describe(ch)
}

// Collect implements the Collector interface
func (uc *UserCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	return uc.CollectWithMetrics(ctx, ch, uc.collectUserMetrics)
}

// collectUserMetrics performs the actual user and account metrics collection
func (uc *UserCollector) collectUserMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	uc.logger.Debug("Starting user and account metrics collection")

	// Since SLURM client is not available, we'll simulate the metrics for now
	// In real implementation, this would use the SLURM client to fetch user and account data

	// Collect user statistics
	if err := uc.collectUserStats(ctx, ch); err != nil {
		return uc.errorBuilder.API(err, "/slurm/v1/jobs", 500, "",
			"Check SLURM user data availability",
			"Verify user job information is accessible")
	}

	// Collect account information and usage
	if err := uc.collectAccountStats(ctx, ch); err != nil {
		return uc.errorBuilder.API(err, "/slurm/v1/accounts", 500, "",
			"Check SLURM account data availability",
			"Verify account information is accessible")
	}

	uc.logger.Debug("Completed user and account metrics collection")
	return nil
}

// collectUserStats collects user-level statistics
func (uc *UserCollector) collectUserStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate user statistics - in real implementation this would come from SLURM API
	// This represents aggregated data from jobs by user
	userStats := []struct {
		Username        string
		Account         string
		JobCount        int
		CPUAllocated    int
		MemoryAllocated int64 // bytes
		JobStates       map[string]int
	}{
		{
			Username:        "alice",
			Account:         "ml_team",
			JobCount:        25,
			CPUAllocated:    120,
			MemoryAllocated: 256 * 1024 * 1024 * 1024, // 256GB
			JobStates: map[string]int{
				"running":   3,
				"pending":   2,
				"completed": 20,
			},
		},
		{
			Username:        "bob",
			Account:         "physics",
			JobCount:        18,
			CPUAllocated:    96,
			MemoryAllocated: 192 * 1024 * 1024 * 1024, // 192GB
			JobStates: map[string]int{
				"running":   2,
				"pending":   1,
				"completed": 15,
			},
		},
		{
			Username:        "charlie",
			Account:         "bio_team",
			JobCount:        32,
			CPUAllocated:    64,
			MemoryAllocated: 128 * 1024 * 1024 * 1024, // 128GB
			JobStates: map[string]int{
				"running":   1,
				"pending":   3,
				"completed": 28,
			},
		},
		{
			Username:        "diana",
			Account:         "chemistry",
			JobCount:        12,
			CPUAllocated:    48,
			MemoryAllocated: 96 * 1024 * 1024 * 1024, // 96GB
			JobStates: map[string]int{
				"running":   2,
				"pending":   0,
				"completed": 10,
			},
		},
	}

	for _, user := range userStats {
		// User job count
		uc.SendMetric(ch, uc.BuildMetric(
			uc.metrics.UserJobCount.WithLabelValues(uc.clusterName, user.Username, user.Account).Desc(),
			prometheus.GaugeValue,
			float64(user.JobCount),
			uc.clusterName,
			user.Username,
			user.Account,
		))

		// User CPU allocation
		uc.SendMetric(ch, uc.BuildMetric(
			uc.metrics.UserCPUAllocated.WithLabelValues(uc.clusterName, user.Username, user.Account).Desc(),
			prometheus.GaugeValue,
			float64(user.CPUAllocated),
			uc.clusterName,
			user.Username,
			user.Account,
		))

		// User memory allocation
		uc.SendMetric(ch, uc.BuildMetric(
			uc.metrics.UserMemoryAllocated.WithLabelValues(uc.clusterName, user.Username, user.Account).Desc(),
			prometheus.GaugeValue,
			float64(user.MemoryAllocated),
			uc.clusterName,
			user.Username,
			user.Account,
		))

		// User job states
		for state, count := range user.JobStates {
			uc.SendMetric(ch, uc.BuildMetric(
				uc.metrics.UserJobStates.WithLabelValues(uc.clusterName, user.Username, user.Account, state).Desc(),
				prometheus.GaugeValue,
				float64(count),
				uc.clusterName,
				user.Username,
				user.Account,
				state,
			))
		}
	}

	uc.LogCollectionf("Collected statistics for %d users", len(userStats))
	return nil
}

// collectAccountStats collects account-level information and usage
func (uc *UserCollector) collectAccountStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate account information and usage - in real implementation this would come from SLURM API
	// This represents data from account associations and usage tracking
	accountStats := []struct {
		AccountName          string
		Organization         string
		Description          string
		ParentAccount        string
		Priority             int
		MaxJobs              int
		MaxCPUs              int
		MaxMemory            int64 // bytes
		TotalJobCount        int
		TotalCPUAllocated    int
		TotalMemoryAllocated int64 // bytes
		FairShareScore       float64
		RawUsage             float64 // CPU hours
		NormalizedUsage      float64 // normalized units
		EffectiveUsage       float64 // with fair-share applied
	}{
		{
			AccountName:          "ml_team",
			Organization:         "Computer Science",
			Description:          "Machine Learning Research Team",
			ParentAccount:        "cs_dept",
			Priority:             1000,
			MaxJobs:              50,
			MaxCPUs:              500,
			MaxMemory:            1024 * 1024 * 1024 * 1024, // 1TB
			TotalJobCount:        45,
			TotalCPUAllocated:    240,
			TotalMemoryAllocated: 512 * 1024 * 1024 * 1024, // 512GB
			FairShareScore:       0.85,
			RawUsage:             1200.5,
			NormalizedUsage:      1.2,
			EffectiveUsage:       1.02,
		},
		{
			AccountName:          "physics",
			Organization:         "Physics Department",
			Description:          "Physics Research Computing",
			ParentAccount:        "physics_dept",
			Priority:             800,
			MaxJobs:              40,
			MaxCPUs:              400,
			MaxMemory:            800 * 1024 * 1024 * 1024, // 800GB
			TotalJobCount:        32,
			TotalCPUAllocated:    192,
			TotalMemoryAllocated: 384 * 1024 * 1024 * 1024, // 384GB
			FairShareScore:       0.92,
			RawUsage:             980.3,
			NormalizedUsage:      0.98,
			EffectiveUsage:       0.90,
		},
		{
			AccountName:          "bio_team",
			Organization:         "Biology Department",
			Description:          "Bioinformatics Research",
			ParentAccount:        "bio_dept",
			Priority:             600,
			MaxJobs:              30,
			MaxCPUs:              300,
			MaxMemory:            600 * 1024 * 1024 * 1024, // 600GB
			TotalJobCount:        38,
			TotalCPUAllocated:    152,
			TotalMemoryAllocated: 304 * 1024 * 1024 * 1024, // 304GB
			FairShareScore:       0.78,
			RawUsage:             760.8,
			NormalizedUsage:      0.76,
			EffectiveUsage:       0.59,
		},
		{
			AccountName:          "chemistry",
			Organization:         "Chemistry Department",
			Description:          "Computational Chemistry",
			ParentAccount:        "chem_dept",
			Priority:             700,
			MaxJobs:              25,
			MaxCPUs:              250,
			MaxMemory:            500 * 1024 * 1024 * 1024, // 500GB
			TotalJobCount:        18,
			TotalCPUAllocated:    72,
			TotalMemoryAllocated: 144 * 1024 * 1024 * 1024, // 144GB
			FairShareScore:       1.15,
			RawUsage:             420.2,
			NormalizedUsage:      0.42,
			EffectiveUsage:       0.48,
		},
	}

	for _, account := range accountStats {
		// Account info metric
		uc.SendMetric(ch, uc.BuildMetric(
			uc.metrics.AccountInfo.WithLabelValues(
				uc.clusterName,
				account.AccountName,
				account.Organization,
				account.Description,
				account.ParentAccount,
			).Desc(),
			prometheus.GaugeValue,
			1,
			uc.clusterName,
			account.AccountName,
			account.Organization,
			account.Description,
			account.ParentAccount,
		))

		// Account job count
		uc.SendMetric(ch, uc.BuildMetric(
			uc.metrics.AccountJobCount.WithLabelValues(uc.clusterName, account.AccountName).Desc(),
			prometheus.GaugeValue,
			float64(account.TotalJobCount),
			uc.clusterName,
			account.AccountName,
		))

		// Account CPU allocation
		uc.SendMetric(ch, uc.BuildMetric(
			uc.metrics.AccountCPUAllocated.WithLabelValues(uc.clusterName, account.AccountName).Desc(),
			prometheus.GaugeValue,
			float64(account.TotalCPUAllocated),
			uc.clusterName,
			account.AccountName,
		))

		// Account memory allocation
		uc.SendMetric(ch, uc.BuildMetric(
			uc.metrics.AccountMemoryAllocated.WithLabelValues(uc.clusterName, account.AccountName).Desc(),
			prometheus.GaugeValue,
			float64(account.TotalMemoryAllocated),
			uc.clusterName,
			account.AccountName,
		))

		// Account fair-share score
		uc.SendMetric(ch, uc.BuildMetric(
			uc.metrics.AccountFairShare.WithLabelValues(uc.clusterName, account.AccountName).Desc(),
			prometheus.GaugeValue,
			account.FairShareScore,
			uc.clusterName,
			account.AccountName,
		))

		// Account usage metrics (different types)
		usageTypes := map[string]float64{
			"raw":        account.RawUsage,
			"normalized": account.NormalizedUsage,
			"effective":  account.EffectiveUsage,
		}

		for usageType, usage := range usageTypes {
			uc.SendMetric(ch, uc.BuildMetric(
				uc.metrics.AccountUsage.WithLabelValues(uc.clusterName, account.AccountName, usageType).Desc(),
				prometheus.GaugeValue,
				usage,
				uc.clusterName,
				account.AccountName,
				usageType,
			))
		}
	}

	uc.LogCollectionf("Collected statistics for %d accounts", len(accountStats))
	return nil
}

// AccountInfo represents account information structure
type UserAccountInfo struct {
	Name         string
	Organization string
	Description  string
	Parent       string
	Priority     int
	Limits       AccountLimits
	Usage        AccountUsage
}

// AccountLimits represents resource limits for an account
type AccountLimits struct {
	MaxJobs   int
	MaxCPUs   int
	MaxMemory int64
	MaxNodes  int
}

// AccountUsage represents usage statistics for an account
type UserAccountUsage struct {
	JobCount    int
	CPUHours    float64
	MemoryHours float64
	NodeHours   float64
	FairShare   float64
	RawUsage    float64
	NormUsage   float64
	EffUsage    float64
}

// UserInfo represents user information and statistics
type UserInfo struct {
	Username       string
	FullName       string
	Email          string
	DefaultAccount string
	Accounts       []string
	Active         bool
	AdminLevel     string
	Usage          UserUsage
}

// UserUsage represents usage statistics for a user
type UserUsage struct {
	JobCount        int
	CPUAllocated    int
	MemoryAllocated int64
	JobStates       map[string]int
	RecentActivity  time.Time
}

// TODO: Following parser methods are unused - preserved for future user association parsing
/*
// parseAccountAssociation parses SLURM account association data
func (uc *UserCollector) parseAccountAssociation(data interface{}) (*AccountInfo, error) {
	// In real implementation, this would parse actual SLURM API response
	// For now, return a mock account info
	return &AccountInfo{
		Name:        "example_account",
		Description: "Example account description",
		Status:      "active",
		CreatedAt:   time.Now().Add(-365 * 24 * time.Hour), // 1 year ago
		UpdatedAt:   time.Now(),
		UserCount:   10,
		JobCount:    25,
	}, nil
}

// parseUserAssociation parses SLURM user association data
func (uc *UserCollector) parseUserAssociation(data interface{}) (*UserInfo, error) {
	// In real implementation, this would parse actual SLURM API response
	// For now, return a mock user info
	return &UserInfo{
		Username:       "example_user",
		FullName:       "Example User",
		Email:          "user@example.com",
		DefaultAccount: "default_account",
		Accounts:       []string{"account1", "account2"},
		Active:         true,
		AdminLevel:     "None",
		Usage: UserUsage{
			JobCount:        15,
			CPUAllocated:    48,
			MemoryAllocated: 96 * 1024 * 1024 * 1024, // 96GB
			JobStates: map[string]int{
				"running":   2,
				"pending":   1,
				"completed": 12,
			},
			RecentActivity: time.Now().Add(-24 * time.Hour),
		},
	}, nil
}
*/
