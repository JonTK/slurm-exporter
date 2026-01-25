// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// AccountQuotaSLURMClient defines the interface for SLURM client operations
// needed for account quota tracking and monitoring
type AccountQuotaSLURMClient interface {
	// Account Quota Operations
	GetAccountQuotas(ctx context.Context, accountName string) (*AccountQuotas, error)
	GetAllAccountQuotas(ctx context.Context) ([]*AccountQuotas, error)
	GetAccountQuotaUsage(ctx context.Context, accountName string) (*AccountQuotaUsage, error)
	GetAccountQuotaHistory(ctx context.Context, accountName string, period string) (*AccountQuotaHistory, error)

	// Resource Limit Operations
	GetAccountResourceLimits(ctx context.Context, accountName string) (*AccountResourceLimits, error)
	GetEffectiveResourceLimits(ctx context.Context, accountName string, userName string) (*EffectiveResourceLimits, error)
	GetAccountQuotaViolations(ctx context.Context, accountName string, period string) ([]*QuotaViolation, error)
	GetQuotaEnforcementStatus(ctx context.Context, accountName string) (*QuotaEnforcementStatus, error)

	// Quota Analysis Operations
	GetAccountQuotaUtilization(ctx context.Context, accountName string) (*QuotaUtilization, error)
	GetAccountQuotaTrends(ctx context.Context, accountName string, period string) (*QuotaTrends, error)
	GetAccountQuotaAlerts(ctx context.Context, accountName string) ([]*QuotaAlert, error)
	GetAccountQuotaRecommendations(ctx context.Context, accountName string) (*QuotaRecommendations, error)
}

// AccountQuotas represents comprehensive quota information for an account
type AccountQuotas struct {
	AccountName   string `json:"account_name"`
	ParentAccount string `json:"parent_account"`
	Description   string `json:"description"`
	Organization  string `json:"organization"`

	// CPU Quotas
	CPUMinutes *ResourceQuota `json:"cpu_minutes"`
	CPUCores   *ResourceQuota `json:"cpu_cores"`
	CPUNodes   *ResourceQuota `json:"cpu_nodes"`

	// Memory Quotas
	MemoryGB        *ResourceQuota `json:"memory_gb"`
	MemoryNodeHours *ResourceQuota `json:"memory_node_hours"`

	// GPU Quotas
	GPUHours *ResourceQuota `json:"gpu_hours"`
	GPUCards *ResourceQuota `json:"gpu_cards"`

	// Storage Quotas
	StorageGB *ResourceQuota `json:"storage_gb"`
	ScratchGB *ResourceQuota `json:"scratch_gb"`
	ArchiveGB *ResourceQuota `json:"archive_gb"`

	// Job Quotas
	MaxJobs        *LimitQuota    `json:"max_jobs"`
	MaxSubmitJobs  *LimitQuota    `json:"max_submit_jobs"`
	MaxArraySize   *LimitQuota    `json:"max_array_size"`
	MaxJobDuration *DurationQuota `json:"max_job_duration"`

	// QoS Quotas
	QoSLimits  map[string]*QoSQuota `json:"qos_limits"`
	DefaultQoS string               `json:"default_qos"`
	AllowedQoS []string             `json:"allowed_qos"`

	// Partition Quotas
	PartitionQuotas   map[string]*PartitionQuota `json:"partition_quotas"`
	AllowedPartitions []string                   `json:"allowed_partitions"`

	// Time-based Quotas
	QuotaPeriod    string        `json:"quota_period"`
	QuotaResetDate time.Time     `json:"quota_reset_date"`
	GracePeriod    time.Duration `json:"grace_period"`

	// Metadata
	CreatedAt        time.Time `json:"created_at"`
	ModifiedAt       time.Time `json:"modified_at"`
	ModifiedBy       string    `json:"modified_by"`
	QuotaVersion     int       `json:"quota_version"`
	EnforcementLevel string    `json:"enforcement_level"`
	Active           bool      `json:"active"`
}

// ResourceQuota represents a quota for a specific resource type
type ResourceQuota struct {
	Limit            float64    `json:"limit"`
	Used             float64    `json:"used"`
	Reserved         float64    `json:"reserved"`
	Available        float64    `json:"available"`
	UtilizationRate  float64    `json:"utilization_rate"`
	ProjectedUsage   float64    `json:"projected_usage"`
	ThresholdPercent float64    `json:"threshold_percent"`
	IsUnlimited      bool       `json:"is_unlimited"`
	IsSoft           bool       `json:"is_soft"`
	ExpiresAt        *time.Time `json:"expires_at"`
}

// LimitQuota represents a numeric limit quota
type LimitQuota struct {
	Limit       int  `json:"limit"`
	Current     int  `json:"current"`
	Peak        int  `json:"peak"`
	Violations  int  `json:"violations"`
	IsUnlimited bool `json:"is_unlimited"`
}

// DurationQuota represents a time duration quota
type DurationQuota struct {
	Limit       time.Duration `json:"limit"`
	Average     time.Duration `json:"average"`
	Maximum     time.Duration `json:"maximum"`
	Violations  int           `json:"violations"`
	IsUnlimited bool          `json:"is_unlimited"`
}

// QoSQuota represents QoS-specific quotas
type QoSQuota struct {
	QoSName       string         `json:"qos_name"`
	Priority      int            `json:"priority"`
	CPULimit      *ResourceQuota `json:"cpu_limit"`
	MemoryLimit   *ResourceQuota `json:"memory_limit"`
	JobLimit      *LimitQuota    `json:"job_limit"`
	WallTimeLimit *DurationQuota `json:"walltime_limit"`
	PreemptMode   string         `json:"preempt_mode"`
}

// PartitionQuota represents partition-specific quotas
type PartitionQuota struct {
	PartitionName  string         `json:"partition_name"`
	CPUQuota       *ResourceQuota `json:"cpu_quota"`
	MemoryQuota    *ResourceQuota `json:"memory_quota"`
	NodeQuota      *ResourceQuota `json:"node_quota"`
	MaxJobsPerUser int            `json:"max_jobs_per_user"`
	MaxTimeLimit   time.Duration  `json:"max_time_limit"`
	Priority       int            `json:"priority"`
}

// AccountQuotaUsage represents current quota usage statistics
type AccountQuotaUsage struct {
	AccountName string `json:"account_name"`
	Period      string `json:"period"`

	// Resource Usage
	CPUUsage     ResourceUsageStats `json:"cpu_usage"`
	MemoryUsage  ResourceUsageStats `json:"memory_usage"`
	GPUUsage     ResourceUsageStats `json:"gpu_usage"`
	StorageUsage ResourceUsageStats `json:"storage_usage"`

	// Job Statistics
	JobsSubmitted int `json:"jobs_submitted"`
	JobsRunning   int `json:"jobs_running"`
	JobsPending   int `json:"jobs_pending"`
	JobsCompleted int `json:"jobs_completed"`
	JobsFailed    int `json:"jobs_failed"`

	// User Activity
	ActiveUsers     int                `json:"active_users"`
	TotalUsers      int                `json:"total_users"`
	UserQuotaShares map[string]float64 `json:"user_quota_shares"`

	// Efficiency Metrics
	ResourceEfficiency float64 `json:"resource_efficiency"`
	QuotaEfficiency    float64 `json:"quota_efficiency"`
	WastePercentage    float64 `json:"waste_percentage"`

	// Trend Indicators
	UsageTrend         string     `json:"usage_trend"`
	GrowthRate         float64    `json:"growth_rate"`
	ProjectedDepletion *time.Time `json:"projected_depletion"`

	LastUpdated time.Time `json:"last_updated"`
}

// ResourceUsageStats represents detailed usage statistics for a resource
type ResourceUsageStats struct {
	Total           float64 `json:"total"`
	Used            float64 `json:"used"`
	Reserved        float64 `json:"reserved"`
	Available       float64 `json:"available"`
	UtilizationRate float64 `json:"utilization_rate"`
	PeakUsage       float64 `json:"peak_usage"`
	AverageUsage    float64 `json:"average_usage"`
	StandardDev     float64 `json:"standard_deviation"`
}

// AccountQuotaCollector collects account quota metrics
type AccountQuotaCollector struct {
	client AccountQuotaSLURMClient
	mutex  sync.RWMutex

	// Quota Limit Metrics
	accountQuotaLimit       *prometheus.GaugeVec
	accountQuotaUsed        *prometheus.GaugeVec
	accountQuotaReserved    *prometheus.GaugeVec
	accountQuotaAvailable   *prometheus.GaugeVec
	accountQuotaUtilization *prometheus.GaugeVec
	accountQuotaThreshold   *prometheus.GaugeVec

	// Resource-specific Quota Metrics
	accountCPUQuotaMinutes *prometheus.GaugeVec
	accountMemoryQuotaGB   *prometheus.GaugeVec
	accountGPUQuotaHours   *prometheus.GaugeVec
	accountStorageQuotaGB  *prometheus.GaugeVec
	accountJobQuotaLimit   *prometheus.GaugeVec
	accountJobQuotaCurrent *prometheus.GaugeVec

	// QoS Quota Metrics
	accountQoSQuotaLimit *prometheus.GaugeVec
	accountQoSQuotaUsed  *prometheus.GaugeVec
	accountQoSPriority   *prometheus.GaugeVec
	accountQoSJobLimit   *prometheus.GaugeVec

	// Partition Quota Metrics
	accountPartitionQuotaLimit *prometheus.GaugeVec
	accountPartitionQuotaUsed  *prometheus.GaugeVec
	accountPartitionPriority   *prometheus.GaugeVec
	accountPartitionMaxJobs    *prometheus.GaugeVec

	// Usage Statistics Metrics
	accountResourceUsageTotal *prometheus.GaugeVec
	accountResourceUsageRate  *prometheus.GaugeVec
	accountResourceEfficiency *prometheus.GaugeVec
	accountResourceWaste      *prometheus.GaugeVec
	accountJobsSubmitted      *prometheus.CounterVec
	accountJobsCompleted      *prometheus.CounterVec
	accountJobsFailed         *prometheus.CounterVec
	accountActiveUsers        *prometheus.GaugeVec

	// Quota Violation Metrics
	accountQuotaViolations        *prometheus.CounterVec
	accountQuotaViolationSeverity *prometheus.GaugeVec
	accountQuotaEnforcementStatus *prometheus.GaugeVec
	accountQuotaGracePeriod       *prometheus.GaugeVec

	// Trend and Projection Metrics
	accountQuotaGrowthRate     *prometheus.GaugeVec
	accountQuotaTrendDirection *prometheus.GaugeVec
	accountQuotaProjectedUsage *prometheus.GaugeVec
	accountQuotaDepletionDays  *prometheus.GaugeVec

	// Alert and Recommendation Metrics
	accountQuotaAlertLevel            *prometheus.GaugeVec
	accountQuotaAlertCount            *prometheus.GaugeVec
	accountQuotaRecommendationScore   *prometheus.GaugeVec
	accountQuotaOptimizationPotential *prometheus.GaugeVec

	// Metadata Metrics
	accountQuotaLastModified *prometheus.GaugeVec
	accountQuotaVersion      *prometheus.GaugeVec
	accountQuotaActive       *prometheus.GaugeVec
	accountQuotaResetDays    *prometheus.GaugeVec
}

// NewAccountQuotaCollector creates a new account quota collector
func NewAccountQuotaCollector(client AccountQuotaSLURMClient) *AccountQuotaCollector {
	return &AccountQuotaCollector{
		client: client,

		// Quota Limit Metrics
		accountQuotaLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_limit",
				Help: "Account quota limit for various resources",
			},
			[]string{"account", "resource_type", "quota_type"},
		),
		accountQuotaUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_used",
				Help: "Account quota used for various resources",
			},
			[]string{"account", "resource_type", "quota_type"},
		),
		accountQuotaReserved: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_reserved",
				Help: "Account quota reserved for various resources",
			},
			[]string{"account", "resource_type", "quota_type"},
		),
		accountQuotaAvailable: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_available",
				Help: "Account quota available for various resources",
			},
			[]string{"account", "resource_type", "quota_type"},
		),
		accountQuotaUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_utilization_rate",
				Help: "Account quota utilization rate (0-1)",
			},
			[]string{"account", "resource_type", "quota_type"},
		),
		accountQuotaThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_threshold_percent",
				Help: "Account quota threshold percentage for alerts",
			},
			[]string{"account", "resource_type"},
		),

		// Resource-specific Quota Metrics
		accountCPUQuotaMinutes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cpu_quota_minutes",
				Help: "Account CPU quota in minutes",
			},
			[]string{"account", "quota_state"},
		),
		accountMemoryQuotaGB: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_memory_quota_gb",
				Help: "Account memory quota in GB",
			},
			[]string{"account", "quota_state"},
		),
		accountGPUQuotaHours: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_gpu_quota_hours",
				Help: "Account GPU quota in hours",
			},
			[]string{"account", "quota_state"},
		),
		accountStorageQuotaGB: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_storage_quota_gb",
				Help: "Account storage quota in GB",
			},
			[]string{"account", "storage_type", "quota_state"},
		),
		accountJobQuotaLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_job_quota_limit",
				Help: "Account job quota limit",
			},
			[]string{"account", "job_type"},
		),
		accountJobQuotaCurrent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_job_quota_current",
				Help: "Account current job count against quota",
			},
			[]string{"account", "job_type"},
		),

		// QoS Quota Metrics
		accountQoSQuotaLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_qos_quota_limit",
				Help: "Account QoS-specific quota limit",
			},
			[]string{"account", "qos", "resource_type"},
		),
		accountQoSQuotaUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_qos_quota_used",
				Help: "Account QoS-specific quota used",
			},
			[]string{"account", "qos", "resource_type"},
		),
		accountQoSPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_qos_priority",
				Help: "Account QoS priority level",
			},
			[]string{"account", "qos"},
		),
		accountQoSJobLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_qos_job_limit",
				Help: "Account QoS job limit",
			},
			[]string{"account", "qos"},
		),

		// Partition Quota Metrics
		accountPartitionQuotaLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_partition_quota_limit",
				Help: "Account partition-specific quota limit",
			},
			[]string{"account", "partition", "resource_type"},
		),
		accountPartitionQuotaUsed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_partition_quota_used",
				Help: "Account partition-specific quota used",
			},
			[]string{"account", "partition", "resource_type"},
		),
		accountPartitionPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_partition_priority",
				Help: "Account partition priority",
			},
			[]string{"account", "partition"},
		),
		accountPartitionMaxJobs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_partition_max_jobs",
				Help: "Account partition maximum jobs per user",
			},
			[]string{"account", "partition"},
		),

		// Usage Statistics Metrics
		accountResourceUsageTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_usage_total",
				Help: "Account total resource usage",
			},
			[]string{"account", "resource_type", "period"},
		),
		accountResourceUsageRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_usage_rate",
				Help: "Account resource usage rate",
			},
			[]string{"account", "resource_type"},
		),
		accountResourceEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_efficiency",
				Help: "Account resource usage efficiency (0-1)",
			},
			[]string{"account", "resource_type"},
		),
		accountResourceWaste: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_waste_percent",
				Help: "Account resource waste percentage",
			},
			[]string{"account", "resource_type"},
		),
		accountJobsSubmitted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_jobs_submitted_total",
				Help: "Total number of jobs submitted by account",
			},
			[]string{"account"},
		),
		accountJobsCompleted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_jobs_completed_total",
				Help: "Total number of jobs completed by account",
			},
			[]string{"account"},
		),
		accountJobsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_jobs_failed_total",
				Help: "Total number of jobs failed by account",
			},
			[]string{"account"},
		),
		accountActiveUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_active_users",
				Help: "Number of active users in account",
			},
			[]string{"account"},
		),

		// Quota Violation Metrics
		accountQuotaViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_quota_violations_total",
				Help: "Total number of quota violations by account",
			},
			[]string{"account", "resource_type", "violation_type"},
		),
		accountQuotaViolationSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_violation_severity",
				Help: "Account quota violation severity (0-10)",
			},
			[]string{"account", "resource_type"},
		),
		accountQuotaEnforcementStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_enforcement_status",
				Help: "Account quota enforcement status (0=disabled, 1=soft, 2=hard)",
			},
			[]string{"account", "resource_type"},
		),
		accountQuotaGracePeriod: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_grace_period_hours",
				Help: "Account quota grace period in hours",
			},
			[]string{"account"},
		),

		// Trend and Projection Metrics
		accountQuotaGrowthRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_growth_rate",
				Help: "Account quota usage growth rate",
			},
			[]string{"account", "resource_type"},
		),
		accountQuotaTrendDirection: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_trend_direction",
				Help: "Account quota usage trend direction (-1=decreasing, 0=stable, 1=increasing)",
			},
			[]string{"account", "resource_type"},
		),
		accountQuotaProjectedUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_projected_usage",
				Help: "Account quota projected usage",
			},
			[]string{"account", "resource_type", "projection_period"},
		),
		accountQuotaDepletionDays: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_depletion_days",
				Help: "Days until account quota depletion at current rate",
			},
			[]string{"account", "resource_type"},
		),

		// Alert and Recommendation Metrics
		accountQuotaAlertLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_alert_level",
				Help: "Account quota alert level (0=none, 1=info, 2=warning, 3=critical)",
			},
			[]string{"account", "resource_type"},
		),
		accountQuotaAlertCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_alert_count",
				Help: "Number of active quota alerts for account",
			},
			[]string{"account", "alert_type"},
		),
		accountQuotaRecommendationScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_recommendation_score",
				Help: "Account quota recommendation score (0-1)",
			},
			[]string{"account", "recommendation_type"},
		),
		accountQuotaOptimizationPotential: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_optimization_potential",
				Help: "Account quota optimization potential percentage",
			},
			[]string{"account", "resource_type"},
		),

		// Metadata Metrics
		accountQuotaLastModified: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_last_modified_timestamp",
				Help: "Timestamp of last quota modification",
			},
			[]string{"account"},
		),
		accountQuotaVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_version",
				Help: "Account quota configuration version",
			},
			[]string{"account"},
		),
		accountQuotaActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_active",
				Help: "Account quota active status (0=inactive, 1=active)",
			},
			[]string{"account"},
		),
		accountQuotaResetDays: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quota_reset_days",
				Help: "Days until account quota reset",
			},
			[]string{"account"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *AccountQuotaCollector) Describe(ch chan<- *prometheus.Desc) {
	c.accountQuotaLimit.Describe(ch)
	c.accountQuotaUsed.Describe(ch)
	c.accountQuotaReserved.Describe(ch)
	c.accountQuotaAvailable.Describe(ch)
	c.accountQuotaUtilization.Describe(ch)
	c.accountQuotaThreshold.Describe(ch)
	c.accountCPUQuotaMinutes.Describe(ch)
	c.accountMemoryQuotaGB.Describe(ch)
	c.accountGPUQuotaHours.Describe(ch)
	c.accountStorageQuotaGB.Describe(ch)
	c.accountJobQuotaLimit.Describe(ch)
	c.accountJobQuotaCurrent.Describe(ch)
	c.accountQoSQuotaLimit.Describe(ch)
	c.accountQoSQuotaUsed.Describe(ch)
	c.accountQoSPriority.Describe(ch)
	c.accountQoSJobLimit.Describe(ch)
	c.accountPartitionQuotaLimit.Describe(ch)
	c.accountPartitionQuotaUsed.Describe(ch)
	c.accountPartitionPriority.Describe(ch)
	c.accountPartitionMaxJobs.Describe(ch)
	c.accountResourceUsageTotal.Describe(ch)
	c.accountResourceUsageRate.Describe(ch)
	c.accountResourceEfficiency.Describe(ch)
	c.accountResourceWaste.Describe(ch)
	c.accountJobsSubmitted.Describe(ch)
	c.accountJobsCompleted.Describe(ch)
	c.accountJobsFailed.Describe(ch)
	c.accountActiveUsers.Describe(ch)
	c.accountQuotaViolations.Describe(ch)
	c.accountQuotaViolationSeverity.Describe(ch)
	c.accountQuotaEnforcementStatus.Describe(ch)
	c.accountQuotaGracePeriod.Describe(ch)
	c.accountQuotaGrowthRate.Describe(ch)
	c.accountQuotaTrendDirection.Describe(ch)
	c.accountQuotaProjectedUsage.Describe(ch)
	c.accountQuotaDepletionDays.Describe(ch)
	c.accountQuotaAlertLevel.Describe(ch)
	c.accountQuotaAlertCount.Describe(ch)
	c.accountQuotaRecommendationScore.Describe(ch)
	c.accountQuotaOptimizationPotential.Describe(ch)
	c.accountQuotaLastModified.Describe(ch)
	c.accountQuotaVersion.Describe(ch)
	c.accountQuotaActive.Describe(ch)
	c.accountQuotaResetDays.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *AccountQuotaCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx := context.Background()

	// Get all account quotas
	allQuotas, err := c.client.GetAllAccountQuotas(ctx)
	if err != nil {
		return
	}

	for _, quota := range allQuotas {
		c.collectAccountQuota(ctx, quota)
		c.collectAccountUsage(ctx, quota.AccountName)
		c.collectQuotaViolations(ctx, quota.AccountName)
		c.collectQuotaTrends(ctx, quota.AccountName)
		c.collectQuotaAlerts(ctx, quota.AccountName)
	}

	// Collect all metrics
	c.accountQuotaLimit.Collect(ch)
	c.accountQuotaUsed.Collect(ch)
	c.accountQuotaReserved.Collect(ch)
	c.accountQuotaAvailable.Collect(ch)
	c.accountQuotaUtilization.Collect(ch)
	c.accountQuotaThreshold.Collect(ch)
	c.accountCPUQuotaMinutes.Collect(ch)
	c.accountMemoryQuotaGB.Collect(ch)
	c.accountGPUQuotaHours.Collect(ch)
	c.accountStorageQuotaGB.Collect(ch)
	c.accountJobQuotaLimit.Collect(ch)
	c.accountJobQuotaCurrent.Collect(ch)
	c.accountQoSQuotaLimit.Collect(ch)
	c.accountQoSQuotaUsed.Collect(ch)
	c.accountQoSPriority.Collect(ch)
	c.accountQoSJobLimit.Collect(ch)
	c.accountPartitionQuotaLimit.Collect(ch)
	c.accountPartitionQuotaUsed.Collect(ch)
	c.accountPartitionPriority.Collect(ch)
	c.accountPartitionMaxJobs.Collect(ch)
	c.accountResourceUsageTotal.Collect(ch)
	c.accountResourceUsageRate.Collect(ch)
	c.accountResourceEfficiency.Collect(ch)
	c.accountResourceWaste.Collect(ch)
	c.accountJobsSubmitted.Collect(ch)
	c.accountJobsCompleted.Collect(ch)
	c.accountJobsFailed.Collect(ch)
	c.accountActiveUsers.Collect(ch)
	c.accountQuotaViolations.Collect(ch)
	c.accountQuotaViolationSeverity.Collect(ch)
	c.accountQuotaEnforcementStatus.Collect(ch)
	c.accountQuotaGracePeriod.Collect(ch)
	c.accountQuotaGrowthRate.Collect(ch)
	c.accountQuotaTrendDirection.Collect(ch)
	c.accountQuotaProjectedUsage.Collect(ch)
	c.accountQuotaDepletionDays.Collect(ch)
	c.accountQuotaAlertLevel.Collect(ch)
	c.accountQuotaAlertCount.Collect(ch)
	c.accountQuotaRecommendationScore.Collect(ch)
	c.accountQuotaOptimizationPotential.Collect(ch)
	c.accountQuotaLastModified.Collect(ch)
	c.accountQuotaVersion.Collect(ch)
	c.accountQuotaActive.Collect(ch)
	c.accountQuotaResetDays.Collect(ch)
}

func (c *AccountQuotaCollector) collectAccountQuota(ctx context.Context, quota *AccountQuotas) {
	_ = ctx
	if quota.CPUMinutes != nil {
		c.accountQuotaLimit.WithLabelValues(quota.AccountName, "cpu", "minutes").Set(quota.CPUMinutes.Limit)
		c.accountQuotaUsed.WithLabelValues(quota.AccountName, "cpu", "minutes").Set(quota.CPUMinutes.Used)
		c.accountQuotaReserved.WithLabelValues(quota.AccountName, "cpu", "minutes").Set(quota.CPUMinutes.Reserved)
		c.accountQuotaAvailable.WithLabelValues(quota.AccountName, "cpu", "minutes").Set(quota.CPUMinutes.Available)
		c.accountQuotaUtilization.WithLabelValues(quota.AccountName, "cpu", "minutes").Set(quota.CPUMinutes.UtilizationRate)
		c.accountQuotaThreshold.WithLabelValues(quota.AccountName, "cpu").Set(quota.CPUMinutes.ThresholdPercent)
		c.accountCPUQuotaMinutes.WithLabelValues(quota.AccountName, "limit").Set(quota.CPUMinutes.Limit)
		c.accountCPUQuotaMinutes.WithLabelValues(quota.AccountName, "used").Set(quota.CPUMinutes.Used)
		c.accountCPUQuotaMinutes.WithLabelValues(quota.AccountName, "available").Set(quota.CPUMinutes.Available)
	}
	if quota.MemoryGB != nil {
		c.accountQuotaLimit.WithLabelValues(quota.AccountName, "memory", "gb").Set(quota.MemoryGB.Limit)
		c.accountQuotaUsed.WithLabelValues(quota.AccountName, "memory", "gb").Set(quota.MemoryGB.Used)
		c.accountQuotaReserved.WithLabelValues(quota.AccountName, "memory", "gb").Set(quota.MemoryGB.Reserved)
		c.accountQuotaAvailable.WithLabelValues(quota.AccountName, "memory", "gb").Set(quota.MemoryGB.Available)
		c.accountQuotaUtilization.WithLabelValues(quota.AccountName, "memory", "gb").Set(quota.MemoryGB.UtilizationRate)
		c.accountMemoryQuotaGB.WithLabelValues(quota.AccountName, "limit").Set(quota.MemoryGB.Limit)
		c.accountMemoryQuotaGB.WithLabelValues(quota.AccountName, "used").Set(quota.MemoryGB.Used)
		c.accountMemoryQuotaGB.WithLabelValues(quota.AccountName, "available").Set(quota.MemoryGB.Available)
	}
	c.publishSimpleResourceQuotas(quota)
	for qosName, qosQuota := range quota.QoSLimits {
		c.accountQoSPriority.WithLabelValues(quota.AccountName, qosName).Set(float64(qosQuota.Priority))
		if qosQuota.CPULimit != nil {
			c.accountQoSQuotaLimit.WithLabelValues(quota.AccountName, qosName, "cpu").Set(qosQuota.CPULimit.Limit)
			c.accountQoSQuotaUsed.WithLabelValues(quota.AccountName, qosName, "cpu").Set(qosQuota.CPULimit.Used)
		}
		if qosQuota.JobLimit != nil {
			c.accountQoSJobLimit.WithLabelValues(quota.AccountName, qosName).Set(float64(qosQuota.JobLimit.Limit))
		}
	}
	for partName, partQuota := range quota.PartitionQuotas {
		c.accountPartitionPriority.WithLabelValues(quota.AccountName, partName).Set(float64(partQuota.Priority))
		c.accountPartitionMaxJobs.WithLabelValues(quota.AccountName, partName).Set(float64(partQuota.MaxJobsPerUser))
		if partQuota.CPUQuota != nil {
			c.accountPartitionQuotaLimit.WithLabelValues(quota.AccountName, partName, "cpu").Set(partQuota.CPUQuota.Limit)
			c.accountPartitionQuotaUsed.WithLabelValues(quota.AccountName, partName, "cpu").Set(partQuota.CPUQuota.Used)
		}
	}
	c.accountQuotaActive.WithLabelValues(quota.AccountName).Set(boolToFloat64(quota.Active))
	c.accountQuotaVersion.WithLabelValues(quota.AccountName).Set(float64(quota.QuotaVersion))
	c.accountQuotaLastModified.WithLabelValues(quota.AccountName).Set(float64(quota.ModifiedAt.Unix()))
	c.accountQuotaGracePeriod.WithLabelValues(quota.AccountName).Set(quota.GracePeriod.Hours())
	c.accountQuotaResetDays.WithLabelValues(quota.AccountName).Set(time.Until(quota.QuotaResetDate).Hours() / 24)
}

// publishSimpleResourceQuotas publishes GPU, Storage, and Job quotas
func (c *AccountQuotaCollector) publishSimpleResourceQuotas(quota *AccountQuotas) {
	if quota.GPUHours != nil {
		c.accountQuotaLimit.WithLabelValues(quota.AccountName, "gpu", "hours").Set(quota.GPUHours.Limit)
		c.accountQuotaUsed.WithLabelValues(quota.AccountName, "gpu", "hours").Set(quota.GPUHours.Used)
		c.accountQuotaUtilization.WithLabelValues(quota.AccountName, "gpu", "hours").Set(quota.GPUHours.UtilizationRate)
		c.accountGPUQuotaHours.WithLabelValues(quota.AccountName, "limit").Set(quota.GPUHours.Limit)
		c.accountGPUQuotaHours.WithLabelValues(quota.AccountName, "used").Set(quota.GPUHours.Used)
	}
	if quota.StorageGB != nil {
		c.accountStorageQuotaGB.WithLabelValues(quota.AccountName, "home", "limit").Set(quota.StorageGB.Limit)
		c.accountStorageQuotaGB.WithLabelValues(quota.AccountName, "home", "used").Set(quota.StorageGB.Used)
	}
	if quota.ScratchGB != nil {
		c.accountStorageQuotaGB.WithLabelValues(quota.AccountName, "scratch", "limit").Set(quota.ScratchGB.Limit)
		c.accountStorageQuotaGB.WithLabelValues(quota.AccountName, "scratch", "used").Set(quota.ScratchGB.Used)
	}
	if quota.MaxJobs != nil {
		c.accountJobQuotaLimit.WithLabelValues(quota.AccountName, "max_jobs").Set(float64(quota.MaxJobs.Limit))
		c.accountJobQuotaCurrent.WithLabelValues(quota.AccountName, "max_jobs").Set(float64(quota.MaxJobs.Current))
	}
	if quota.MaxSubmitJobs != nil {
		c.accountJobQuotaLimit.WithLabelValues(quota.AccountName, "max_submit").Set(float64(quota.MaxSubmitJobs.Limit))
		c.accountJobQuotaCurrent.WithLabelValues(quota.AccountName, "max_submit").Set(float64(quota.MaxSubmitJobs.Current))
	}
}

func (c *AccountQuotaCollector) collectAccountUsage(ctx context.Context, accountName string) {
	usage, err := c.client.GetAccountQuotaUsage(ctx, accountName)
	if err != nil {
		return
	}

	// Resource usage statistics
	c.accountResourceUsageTotal.WithLabelValues(accountName, "cpu", usage.Period).Set(usage.CPUUsage.Total)
	c.accountResourceUsageRate.WithLabelValues(accountName, "cpu").Set(usage.CPUUsage.UtilizationRate)
	c.accountResourceEfficiency.WithLabelValues(accountName, "cpu").Set(usage.ResourceEfficiency)
	c.accountResourceWaste.WithLabelValues(accountName, "cpu").Set(usage.WastePercentage)

	// Job statistics - Set values for counters (these would typically increment)
	// In a real implementation, these would be incremented based on events
	c.accountJobsSubmitted.WithLabelValues(accountName).Add(float64(usage.JobsSubmitted))
	c.accountJobsCompleted.WithLabelValues(accountName).Add(float64(usage.JobsCompleted))
	c.accountJobsFailed.WithLabelValues(accountName).Add(float64(usage.JobsFailed))

	// User activity
	c.accountActiveUsers.WithLabelValues(accountName).Set(float64(usage.ActiveUsers))

	// Growth and trends
	c.accountQuotaGrowthRate.WithLabelValues(accountName, "cpu").Set(usage.GrowthRate)

	var trendValue float64
	switch usage.UsageTrend {
	case "increasing":
		trendValue = 1.0
	case "decreasing":
		trendValue = -1.0
	default:
		trendValue = 0.0
	}
	c.accountQuotaTrendDirection.WithLabelValues(accountName, "cpu").Set(trendValue)

	// Projected depletion
	if usage.ProjectedDepletion != nil {
		daysToDepletion := time.Until(*usage.ProjectedDepletion).Hours() / 24
		c.accountQuotaDepletionDays.WithLabelValues(accountName, "cpu").Set(daysToDepletion)
	}
}

func (c *AccountQuotaCollector) collectQuotaViolations(ctx context.Context, accountName string) {
	violations, err := c.client.GetAccountQuotaViolations(ctx, accountName, "7d")
	if err != nil {
		return
	}

	// Count violations by type
	violationCounts := make(map[string]map[string]int)
	for _, violation := range violations {
		if _, ok := violationCounts[violation.ResourceType]; !ok {
			violationCounts[violation.ResourceType] = make(map[string]int)
		}
		violationCounts[violation.ResourceType][violation.ViolationType]++
	}

	// Set violation metrics
	for resourceType, typeCounts := range violationCounts {
		for violationType, count := range typeCounts {
			c.accountQuotaViolations.WithLabelValues(accountName, resourceType, violationType).Add(float64(count))
		}
	}

	// Get enforcement status
	_, err = c.client.GetQuotaEnforcementStatus(ctx, accountName)
	if err == nil {
		// Mock enforcement status values
		c.accountQuotaEnforcementStatus.WithLabelValues(accountName, "cpu").Set(2.0) // hard enforcement
		c.accountQuotaViolationSeverity.WithLabelValues(accountName, "cpu").Set(3.0) // severity level
	}
}

func (c *AccountQuotaCollector) collectQuotaTrends(ctx context.Context, accountName string) {
	_, err := c.client.GetAccountQuotaTrends(ctx, accountName, "30d")
	if err != nil {
		return
	}

	// Mock trend data
	c.accountQuotaProjectedUsage.WithLabelValues(accountName, "cpu", "7d").Set(1500.0)
	c.accountQuotaProjectedUsage.WithLabelValues(accountName, "cpu", "30d").Set(6000.0)
}

func (c *AccountQuotaCollector) collectQuotaAlerts(ctx context.Context, accountName string) {
	alerts, err := c.client.GetAccountQuotaAlerts(ctx, accountName)
	if err != nil {
		return
	}

	// Count alerts by type
	alertCounts := make(map[string]int)
	maxAlertLevel := 0
	for _, alert := range alerts {
		alertCounts[alert.Type]++
		if alert.Level > maxAlertLevel {
			maxAlertLevel = alert.Level
		}
	}

	// Set alert metrics
	for alertType, count := range alertCounts {
		c.accountQuotaAlertCount.WithLabelValues(accountName, alertType).Set(float64(count))
	}
	c.accountQuotaAlertLevel.WithLabelValues(accountName, "overall").Set(float64(maxAlertLevel))

	// Get recommendations
	_, err = c.client.GetAccountQuotaRecommendations(ctx, accountName)
	if err == nil {
		// Mock recommendation scores
		c.accountQuotaRecommendationScore.WithLabelValues(accountName, "increase_quota").Set(0.8)
		c.accountQuotaOptimizationPotential.WithLabelValues(accountName, "cpu").Set(25.0)
	}
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// Additional types needed for complete implementation

// AccountQuotaHistory represents historical quota data
type AccountQuotaHistory struct {
	AccountName string                  `json:"account_name"`
	Period      string                  `json:"period"`
	DataPoints  []AccountQuotaDataPoint `json:"data_points"`
}

// AccountQuotaDataPoint represents a single historical data point
type AccountQuotaDataPoint struct {
	Timestamp   time.Time          `json:"timestamp"`
	Quotas      map[string]float64 `json:"quotas"`
	Usage       map[string]float64 `json:"usage"`
	Utilization map[string]float64 `json:"utilization"`
}

// AccountResourceLimits represents resource limits for an account
type AccountResourceLimits struct {
	AccountName   string           `json:"account_name"`
	CPULimits     ResourceLimitSet `json:"cpu_limits"`
	MemoryLimits  ResourceLimitSet `json:"memory_limits"`
	GPULimits     ResourceLimitSet `json:"gpu_limits"`
	StorageLimits ResourceLimitSet `json:"storage_limits"`
}

// ResourceLimitSet represents a set of resource limits
type ResourceLimitSet struct {
	HardLimit    float64 `json:"hard_limit"`
	SoftLimit    float64 `json:"soft_limit"`
	BurstLimit   float64 `json:"burst_limit"`
	CurrentUsage float64 `json:"current_usage"`
}

// EffectiveResourceLimits represents effective limits for a user in an account
type EffectiveResourceLimits struct {
	AccountName string   `json:"account_name"`
	UserName    string   `json:"user_name"`
	CPULimit    float64  `json:"cpu_limit"`
	MemoryLimit float64  `json:"memory_limit"`
	JobLimit    int      `json:"job_limit"`
	Sources     []string `json:"sources"`
}

// Note: QuotaViolation type is defined in common_types.go

// QuotaEnforcementStatus represents enforcement status
type QuotaEnforcementStatus struct {
	AccountName       string     `json:"account_name"`
	EnforcementMode   string     `json:"enforcement_mode"`
	GracePeriodActive bool       `json:"grace_period_active"`
	GracePeriodEnd    *time.Time `json:"grace_period_end"`
	BlockedOperations []string   `json:"blocked_operations"`
}

// QuotaUtilization represents detailed utilization information
type QuotaUtilization struct {
	AccountName    string     `json:"account_name"`
	ResourceType   string     `json:"resource_type"`
	CurrentUsage   float64    `json:"current_usage"`
	QuotaLimit     float64    `json:"quota_limit"`
	Utilization    float64    `json:"utilization"`
	TrendDirection string     `json:"trend_direction"`
	ProjectedFull  *time.Time `json:"projected_full"`
}

// QuotaTrends represents quota usage trends
type QuotaTrends struct {
	AccountName string               `json:"account_name"`
	Period      string               `json:"period"`
	TrendData   map[string][]float64 `json:"trend_data"`
	Predictions map[string]float64   `json:"predictions"`
	Seasonality map[string]bool      `json:"seasonality"`
}

// QuotaAlert represents a quota-related alert
type QuotaAlert struct {
	AlertID      string    `json:"alert_id"`
	AccountName  string    `json:"account_name"`
	Type         string    `json:"type"`
	Level        int       `json:"level"`
	Message      string    `json:"message"`
	ResourceType string    `json:"resource_type"`
	CurrentUsage float64   `json:"current_usage"`
	Threshold    float64   `json:"threshold"`
	Timestamp    time.Time `json:"timestamp"`
}

// QuotaRecommendations represents optimization recommendations
type QuotaRecommendations struct {
	AccountName       string                `json:"account_name"`
	Recommendations   []QuotaRecommendation `json:"recommendations"`
	OptimizationScore float64               `json:"optimization_score"`
	PotentialSavings  float64               `json:"potential_savings"`
}

// Note: QuotaRecommendation type is defined in common_types.go
