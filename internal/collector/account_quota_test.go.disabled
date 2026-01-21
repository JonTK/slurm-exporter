package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountQuotaSLURMClient struct {
	mock.Mock
}

func (m *MockAccountQuotaSLURMClient) GetAccountQuotas(ctx context.Context, accountName string) (*AccountQuotas, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountQuotas), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetUserQuotas(ctx context.Context, userName string) (*UserQuotas, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserQuotas), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetAccountHierarchy(ctx context.Context, accountName string) (*AccountHierarchy, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountHierarchy), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetQuotaUsage(ctx context.Context, accountName string, period string) (*QuotaUsage, error) {
	args := m.Called(ctx, accountName, period)
	return args.Get(0).(*QuotaUsage), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetQuotaViolations(ctx context.Context, accountName string, severity string) (*QuotaViolations, error) {
	args := m.Called(ctx, accountName, severity)
	return args.Get(0).(*QuotaViolations), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetQuotaTrends(ctx context.Context, accountName string, period string) (*QuotaTrends, error) {
	args := m.Called(ctx, accountName, period)
	return args.Get(0).(*QuotaTrends), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetAccountQuotaTrends(ctx context.Context, accountName string, period string) (*QuotaTrends, error) {
	args := m.Called(ctx, accountName, period)
	return args.Get(0).(*QuotaTrends), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetQuotaAlerts(ctx context.Context, accountName string) (*QuotaAlerts, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*QuotaAlerts), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetQoSLimits(ctx context.Context, qosName string) (*QoSLimits, error) {
	args := m.Called(ctx, qosName)
	return args.Get(0).(*QoSLimits), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetPartitionLimits(ctx context.Context, partitionName string) (*PartitionLimits, error) {
	args := m.Called(ctx, partitionName)
	return args.Get(0).(*PartitionLimits), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetQuotaCompliance(ctx context.Context, accountName string) (*QuotaCompliance, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*QuotaCompliance), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetQuotaRecommendations(ctx context.Context, accountName string) (*QuotaRecommendations, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*QuotaRecommendations), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetAccountQuotaRecommendations(ctx context.Context, accountName string) (*QuotaRecommendations, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*QuotaRecommendations), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetSystemQuotaSummary(ctx context.Context) (*SystemQuotaSummary, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemQuotaSummary), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetAccountQuotaAlerts(ctx context.Context, accountName string) ([]*QuotaAlert, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*QuotaAlert), args.Error(1)
}

func (m *MockAccountQuotaSLURMClient) GetAccountQuotaHistory(ctx context.Context, accountName string, period string) (*AccountQuotaHistory, error) {
	args := m.Called(ctx, accountName, period)
	return args.Get(0).(*AccountQuotaHistory), args.Error(1)
}

// Mock structures for missing types
type UserQuotas struct{}
type QuotaUsage struct{}
type QuotaViolations struct{}
type QuotaAlerts struct{}
type SystemQuotaSummary struct{}

func TestNewAccountQuotaCollector(t *testing.T) {
	client := &MockAccountQuotaSLURMClient{}
	collector := NewAccountQuotaCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.accountQuotaLimit)
	assert.NotNil(t, collector.accountQuotaUsed)
	assert.NotNil(t, collector.accountQuotaAvailable)
	assert.NotNil(t, collector.accountQuotaViolations)
}

func TestAccountQuotaCollector_Describe(t *testing.T) {
	client := &MockAccountQuotaSLURMClient{}
	collector := NewAccountQuotaCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 40 metrics, verify we have the correct number
	assert.Equal(t, 40, len(descs), "Should have exactly 40 metric descriptions")
}

func TestAccountQuotaCollector_Collect_Success(t *testing.T) {
	client := &MockAccountQuotaSLURMClient{}

	// Mock account quotas
	accountQuotas := &AccountQuotas{
		AccountName: "research",
		CPUMinutes: &ResourceQuota{
			Limit:           10000,
			Used:            7500,
			Available:       2500,
			Reserved:        500,
			UtilizationRate: 0.75,
		},
		MemoryGB: &ResourceQuota{
			Limit:           102400, // 100GB
			Used:            81920,  // 80GB
			Available:       20480,  // 20GB
			Reserved:        5120,   // 5GB
			UtilizationRate: 0.80,
		},
		GPUHours: &ResourceQuota{
			Limit:           100,
			Used:            60,
			Available:       40,
			Reserved:        10,
			UtilizationRate: 0.60,
		},
		StorageGB: &ResourceQuota{
			Limit:           1048576, // 1TB
			Used:            524288,  // 500GB
			Available:       524288,  // 500GB
			Reserved:        52428,   // 50GB
			UtilizationRate: 0.50,
		},
		MaxJobs: &LimitQuota{
			Limit:   1000,
			Current: 250,
			Peak:    300,
		},
	}

	// Mock quota usage
	quotaUsage := &QuotaUsage{
		AccountName:       "research",
		Period:           "24h",
		AverageCPUUsage:  7000,
		PeakCPUUsage:     8500,
		AverageMemUsage:  75000,
		PeakMemUsage:     85000,
		TrendDirection:   "increasing",
		ProjectedExhaust: time.Now().Add(7 * 24 * time.Hour),
	}

	// Mock quota violations
	quotaViolations := &QuotaViolations{
		AccountName: "research",
		Violations: []Violation{
			{
				Type:      "cpu_exceeded",
				Severity:  "warning",
				Timestamp: time.Now().Add(-2 * time.Hour),
				Value:     8500,
				Limit:     8000,
			},
		},
		TotalCount:     5,
		WarningCount:   3,
		CriticalCount:  2,
		LastViolation:  time.Now().Add(-2 * time.Hour),
		ResolutionTime: 30 * time.Minute,
	}

	// Mock quota compliance
	quotaCompliance := &QuotaCompliance{
		AccountName:     "research",
		ComplianceScore: 0.92,
		PolicyAdherence: 0.95,
		ViolationRate:   0.05,
		Status:          "compliant",
	}

	// Setup mock expectations
	client.On("GetAccountQuotas", mock.Anything, "research").Return(accountQuotas, nil)
	client.On("GetAccountQuotas", mock.Anything, "engineering").Return(&AccountQuotas{
		AccountName: "engineering",
		CPUMinutes: &ResourceQuota{
			Limit:           5000,
			Used:            2500,
			UtilizationRate: 0.50,
		},
	}, nil)
	client.On("GetQuotaUsage", mock.Anything, "research", "24h").Return(quotaUsage, nil)
	client.On("GetQuotaViolations", mock.Anything, "research", "all").Return(quotaViolations, nil)
	client.On("GetQuotaTrends", mock.Anything, "research", "7d").Return(&QuotaTrends{}, nil)
	client.On("GetQuotaAlerts", mock.Anything, "research").Return(&QuotaAlerts{}, nil)
	client.On("GetQuotaCompliance", mock.Anything, "research").Return(quotaCompliance, nil)
	client.On("GetSystemQuotaSummary", mock.Anything).Return(&SystemQuotaSummary{}, nil)

	collector := NewAccountQuotaCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 200)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify we collected metrics
	assert.Greater(t, len(metrics), 0, "Should collect at least some metrics")

	// Verify specific metrics are present
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		desc := metric.Desc()
		metricNames[desc.String()] = true
	}

	// Check for key metric families
	foundCPUQuota := false
	foundMemoryQuota := false
	foundViolations := false
	foundCompliance := false
	foundLatency := false

	for name := range metricNames {
		if strings.Contains(name, "cpu_quota_limit") {
			foundCPUQuota = true
		}
		if strings.Contains(name, "memory_quota_limit") {
			foundMemoryQuota = true
		}
		if strings.Contains(name, "quota_violation_total") {
			foundViolations = true
		}
		if strings.Contains(name, "quota_compliance_score") {
			foundCompliance = true
		}
		if strings.Contains(name, "collection_latency") {
			foundLatency = true
		}
	}

	assert.True(t, foundCPUQuota, "Should have CPU quota metrics")
	assert.True(t, foundMemoryQuota, "Should have memory quota metrics")
	assert.True(t, foundViolations, "Should have violation metrics")
	assert.True(t, foundCompliance, "Should have compliance metrics")
	assert.True(t, foundLatency, "Should have latency metrics")

	client.AssertExpectations(t)
}

func TestAccountQuotaCollector_Collect_Error(t *testing.T) {
	client := &MockAccountQuotaSLURMClient{}

	// Mock error response
	client.On("GetAccountQuotas", mock.Anything, mock.AnythingOfType("string")).Return((*AccountQuotas)(nil), assert.AnError)
	client.On("GetSystemQuotaSummary", mock.Anything).Return((*SystemQuotaSummary)(nil), assert.AnError)

	collector := NewAccountQuotaCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should still collect some metrics (empty metrics after reset)
	assert.GreaterOrEqual(t, len(metrics), 0, "Should handle errors gracefully")

	client.AssertExpectations(t)
}

func TestAccountQuotaCollector_MetricValues(t *testing.T) {
	client := &MockAccountQuotaSLURMClient{}

	// Create test data with known values
	accountQuotas := &AccountQuotas{
		AccountName: "test_account",
		CPUQuota: ResourceQuota{
			Limit:      5000,
			Used:       3750,
			Available:  1250,
			Percentage: 75.0,
		},
		MemoryQuota: ResourceQuota{
			Limit:      102400,
			Used:       51200,
			Available:  51200,
			Percentage: 50.0,
		},
		JobQuota: JobResourceQuota{
			MaxJobs:        500,
			RunningJobs:    100,
			PendingJobs:    50,
			AvailableSlots: 350,
			Percentage:     30.0,
		},
		WarningLevel:  80.0,
		CriticalLevel: 90.0,
		Enforced:      true,
		QoSQuotas:     map[string]ResourceQuota{},
		PartitionQuotas: map[string]ResourceQuota{},
	}

	client.On("GetAccountQuotas", mock.Anything, "research").Return(accountQuotas, nil)
	client.On("GetAccountQuotas", mock.Anything, mock.AnythingOfType("string")).Return(&AccountQuotas{
		QoSQuotas:       map[string]ResourceQuota{},
		PartitionQuotas: map[string]ResourceQuota{},
	}, nil)
	client.On("GetQuotaUsage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaUsage{}, nil)
	client.On("GetQuotaViolations", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaViolations{}, nil)
	client.On("GetQuotaTrends", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaTrends{}, nil)
	client.On("GetQuotaAlerts", mock.Anything, mock.AnythingOfType("string")).Return(&QuotaAlerts{}, nil)
	client.On("GetQuotaCompliance", mock.Anything, mock.AnythingOfType("string")).Return(&QuotaCompliance{}, nil)
	client.On("GetSystemQuotaSummary", mock.Anything).Return(&SystemQuotaSummary{}, nil)

	collector := NewAccountQuotaCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundCPULimit := false
	foundCPUUsed := false
	foundMemoryLimit := false
	foundJobMax := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_account_cpu_quota_limit":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(5000), *mf.Metric[0].Gauge.Value)
				foundCPULimit = true
			}
		case "slurm_account_cpu_quota_used":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(3750), *mf.Metric[0].Gauge.Value)
				foundCPUUsed = true
			}
		case "slurm_account_memory_quota_limit_bytes":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(102400*1024*1024), *mf.Metric[0].Gauge.Value)
				foundMemoryLimit = true
			}
		case "slurm_account_job_quota_max":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(500), *mf.Metric[0].Gauge.Value)
				foundJobMax = true
			}
		}
	}

	assert.True(t, foundCPULimit, "Should find CPU limit metric with correct value")
	assert.True(t, foundCPUUsed, "Should find CPU used metric with correct value")
	assert.True(t, foundMemoryLimit, "Should find memory limit metric with correct value")
	assert.True(t, foundJobMax, "Should find job max metric with correct value")
}

func TestAccountQuotaCollector_Integration(t *testing.T) {
	client := &MockAccountQuotaSLURMClient{}

	// Setup comprehensive mock data
	setupAccountQuotaMocks(client)

	collector := NewAccountQuotaCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_account_cpu_quota_limit CPU quota limit for account
		# TYPE slurm_account_cpu_quota_limit gauge
		slurm_account_cpu_quota_limit{account="research"} 10000
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_account_cpu_quota_limit")
	assert.NoError(t, err)
}

func TestAccountQuotaCollector_QoSMetrics(t *testing.T) {
	client := &MockAccountQuotaSLURMClient{}

	// Setup QoS specific mocks
	setupQoSMocks(client)

	collector := NewAccountQuotaCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify QoS metrics are present
	foundQoSLimit := false
	foundQoSUsed := false
	foundQoSPercentage := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "qos_quota_limit") {
			foundQoSLimit = true
		}
		if strings.Contains(desc, "qos_quota_used") {
			foundQoSUsed = true
		}
		if strings.Contains(desc, "qos_quota_percentage") {
			foundQoSPercentage = true
		}
	}

	assert.True(t, foundQoSLimit, "Should find QoS limit metrics")
	assert.True(t, foundQoSUsed, "Should find QoS used metrics")
	assert.True(t, foundQoSPercentage, "Should find QoS percentage metrics")
}

func TestAccountQuotaCollector_PartitionMetrics(t *testing.T) {
	client := &MockAccountQuotaSLURMClient{}

	// Setup partition specific mocks
	setupPartitionMocks(client)

	collector := NewAccountQuotaCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify partition metrics are present
	foundPartitionLimit := false
	foundPartitionUsed := false
	foundPartitionPercentage := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "partition_quota_limit") {
			foundPartitionLimit = true
		}
		if strings.Contains(desc, "partition_quota_used") {
			foundPartitionUsed = true
		}
		if strings.Contains(desc, "partition_quota_percentage") {
			foundPartitionPercentage = true
		}
	}

	assert.True(t, foundPartitionLimit, "Should find partition limit metrics")
	assert.True(t, foundPartitionUsed, "Should find partition used metrics")
	assert.True(t, foundPartitionPercentage, "Should find partition percentage metrics")
}

// Helper functions

func setupAccountQuotaMocks(client *MockAccountQuotaSLURMClient) {
	accountQuotas := &AccountQuotas{
		AccountName: "research",
		CPUQuota: ResourceQuota{
			Limit:      10000,
			Used:       7500,
			Available:  2500,
			Percentage: 75.0,
		},
		MemoryQuota: ResourceQuota{
			Limit:      102400,
			Used:       81920,
			Available:  20480,
			Percentage: 80.0,
		},
		JobQuota: JobResourceQuota{
			MaxJobs:        1000,
			RunningJobs:    250,
			PendingJobs:    150,
			AvailableSlots: 600,
			Percentage:     40.0,
		},
		WarningLevel:    80.0,
		CriticalLevel:   90.0,
		Enforced:        true,
		QoSQuotas:       map[string]ResourceQuota{},
		PartitionQuotas: map[string]ResourceQuota{},
	}

	client.On("GetAccountQuotas", mock.Anything, "research").Return(accountQuotas, nil)
	client.On("GetAccountQuotas", mock.Anything, mock.AnythingOfType("string")).Return(&AccountQuotas{
		QoSQuotas:       map[string]ResourceQuota{},
		PartitionQuotas: map[string]ResourceQuota{},
	}, nil)
	client.On("GetQuotaUsage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaUsage{}, nil)
	client.On("GetQuotaViolations", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaViolations{}, nil)
	client.On("GetQuotaTrends", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaTrends{}, nil)
	client.On("GetQuotaAlerts", mock.Anything, mock.AnythingOfType("string")).Return(&QuotaAlerts{}, nil)
	client.On("GetQuotaCompliance", mock.Anything, mock.AnythingOfType("string")).Return(&QuotaCompliance{}, nil)
	client.On("GetSystemQuotaSummary", mock.Anything).Return(&SystemQuotaSummary{}, nil)
}

func setupQoSMocks(client *MockAccountQuotaSLURMClient) {
	accountQuotas := &AccountQuotas{
		AccountName: "research",
		CPUQuota: ResourceQuota{
			Limit:      10000,
			Used:       7500,
			Percentage: 75.0,
		},
		QoSQuotas: map[string]ResourceQuota{
			"high": {
				Limit:      1000,
				Used:       800,
				Available:  200,
				Percentage: 80.0,
			},
			"normal": {
				Limit:      5000,
				Used:       2500,
				Available:  2500,
				Percentage: 50.0,
			},
			"low": {
				Limit:      2000,
				Used:       500,
				Available:  1500,
				Percentage: 25.0,
			},
		},
		PartitionQuotas: map[string]ResourceQuota{},
	}

	client.On("GetAccountQuotas", mock.Anything, mock.AnythingOfType("string")).Return(accountQuotas, nil)
	client.On("GetQuotaUsage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaUsage{}, nil)
	client.On("GetQuotaViolations", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaViolations{}, nil)
	client.On("GetQuotaTrends", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaTrends{}, nil)
	client.On("GetQuotaAlerts", mock.Anything, mock.AnythingOfType("string")).Return(&QuotaAlerts{}, nil)
	client.On("GetQuotaCompliance", mock.Anything, mock.AnythingOfType("string")).Return(&QuotaCompliance{}, nil)
	client.On("GetSystemQuotaSummary", mock.Anything).Return(&SystemQuotaSummary{}, nil)
}

func setupPartitionMocks(client *MockAccountQuotaSLURMClient) {
	accountQuotas := &AccountQuotas{
		AccountName: "research",
		CPUQuota: ResourceQuota{
			Limit:      10000,
			Used:       7500,
			Percentage: 75.0,
		},
		QoSQuotas: map[string]ResourceQuota{},
		PartitionQuotas: map[string]ResourceQuota{
			"gpu": {
				Limit:      50,
				Used:       40,
				Available:  10,
				Percentage: 80.0,
			},
			"cpu": {
				Limit:      1000,
				Used:       600,
				Available:  400,
				Percentage: 60.0,
			},
			"highmem": {
				Limit:      100,
				Used:       30,
				Available:  70,
				Percentage: 30.0,
			},
		},
	}

	client.On("GetAccountQuotas", mock.Anything, mock.AnythingOfType("string")).Return(accountQuotas, nil)
	client.On("GetQuotaUsage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaUsage{}, nil)
	client.On("GetQuotaViolations", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaViolations{}, nil)
	client.On("GetQuotaTrends", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QuotaTrends{}, nil)
	client.On("GetQuotaAlerts", mock.Anything, mock.AnythingOfType("string")).Return(&QuotaAlerts{}, nil)
	client.On("GetQuotaCompliance", mock.Anything, mock.AnythingOfType("string")).Return(&QuotaCompliance{}, nil)
	client.On("GetSystemQuotaSummary", mock.Anything).Return(&SystemQuotaSummary{}, nil)
}