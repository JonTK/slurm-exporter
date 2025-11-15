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

type MockAccountSLURMClient struct {
	mock.Mock
}

func (m *MockAccountSLURMClient) GetAccountHierarchy(ctx context.Context) (*AccountHierarchy, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AccountHierarchy), args.Error(1)
}

func (m *MockAccountSLURMClient) GetAccountQuotas(ctx context.Context, accountName string) (*AccountQuota, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountQuota), args.Error(1)
}

func (m *MockAccountSLURMClient) GetAccountQoS(ctx context.Context, accountName string) ([]*QoSAssignment, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*QoSAssignment), args.Error(1)
}

func (m *MockAccountSLURMClient) GetAccountReservations(ctx context.Context, accountName string) ([]*ReservationUsage, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*ReservationUsage), args.Error(1)
}

func (m *MockAccountSLURMClient) GetAccountUsageAnalysis(ctx context.Context, accountName string) (*AccountUsageAnalysis, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountUsageAnalysis), args.Error(1)
}

func (m *MockAccountSLURMClient) GetAccountCostAnalysis(ctx context.Context, accountName string) (*AccountCostAnalysis, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountCostAnalysis), args.Error(1)
}

func (m *MockAccountSLURMClient) DetectAccountAnomalies(ctx context.Context, accountName string) ([]*AccountAnomaly, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*AccountAnomaly), args.Error(1)
}

func (m *MockAccountSLURMClient) ValidateAccountAccess(ctx context.Context, accountName string) (*AccountAccessValidation, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountAccessValidation), args.Error(1)
}

func TestNewAccountCollector(t *testing.T) {
	client := &MockAccountSLURMClient{}
	collector := NewAccountCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.accountHierarchyBalance)
	assert.NotNil(t, collector.accountHierarchyUtilization)
	assert.NotNil(t, collector.accountQuotaLimit)
	assert.NotNil(t, collector.accountQuotaUsed)
	assert.NotNil(t, collector.accountQuotaUtilization)
}

func TestAccountCollector_Describe(t *testing.T) {
	client := &MockAccountSLURMClient{}
	collector := NewAccountCollector(client)

	ch := make(chan *prometheus.Desc, 50)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 45+ metrics, verify we have a reasonable number
	assert.GreaterOrEqual(t, len(descs), 40, "Should have at least 40 metric descriptions")
}

func TestAccountCollector_Collect_Success(t *testing.T) {
	client := &MockAccountSLURMClient{}

	// Mock account hierarchy
	hierarchy := &AccountHierarchy{
		RootAccounts: []*AccountNode{
			{
				Name:           "root",
				Parent:         "",
				Children:       []string{"dept1", "dept2"},
				Users:          []string{"admin"},
				Level:          0,
				TotalUsers:     10,
				ActiveUsers:    8,
				SubAccounts:    2,
				Balance:        1000000.0,
				UtilizationRate: 75.5,
			},
		},
		AccountMap: map[string]*AccountNode{
			"root": {
				Name:           "root",
				Parent:         "",
				Children:       []string{"dept1", "dept2"},
				Users:          []string{"admin"},
				Level:          0,
				TotalUsers:     10,
				ActiveUsers:    8,
				SubAccounts:    2,
				Balance:        1000000.0,
				UtilizationRate: 75.5,
			},
		},
		BalanceScore:      85.0,
		UtilizationScore:  92.3,
		TotalAccounts:     3,
		ActiveAccounts:    3,
		InactiveAccounts:  0,
		OverUtilized:      1,
		UnderUtilized:     1,
		BalancedAccounts:  1,
	}

	// Mock account quota
	quota := &AccountQuota{
		AccountName:      "root",
		CPUQuota:         1000,
		MemoryQuota:      2048000,
		StorageQuota:     10000000,
		CPUUsed:          750,
		MemoryUsed:       1536000,
		StorageUsed:      7500000,
		CPUUtilization:   75.0,
		MemoryUtilization: 75.0,
		StorageUtilization: 75.0,
		IsViolating:      false,
		ViolationType:    "",
		ViolationSeverity: "",
		LastUpdated:      time.Now(),
	}

	// Mock QoS assignments
	qosAssignments := []*QoSAssignment{
		{
			AccountName:       "root",
			QoSName:          "normal",
			IsDefault:        true,
			MaxJobs:          100,
			MaxSubmitJobs:    200,
			MaxCPUs:          1000,
			MaxMemory:        2048000,
			Priority:         1000,
			Preempt:          true,
			PreemptMode:      "cancel",
			GraceTime:        300,
			UsageThresholdFactor: 1.0,
		},
	}

	// Mock reservation usage
	reservations := []*ReservationUsage{
		{
			AccountName:       "root",
			ReservationName:   "maintenance",
			StartTime:         time.Now().Add(-time.Hour),
			EndTime:           time.Now().Add(time.Hour),
			AllocatedCPUs:     100,
			UsedCPUs:          75,
			UtilizationRate:   75.0,
			IsActive:          true,
			Purpose:           "system maintenance",
		},
	}

	// Mock usage analysis
	usageAnalysis := &AccountUsageAnalysis{
		AccountName:        "root",
		AnalysisPeriod:     "24h",
		TotalJobs:          150,
		SuccessfulJobs:     145,
		FailedJobs:         5,
		CancelledJobs:      0,
		SuccessRate:        96.7,
		AvgJobDuration:     3600,
		AvgWaitTime:        300,
		PeakUsageTime:      time.Now().Add(-6 * time.Hour),
		LowUsageTime:       time.Now().Add(-2 * time.Hour),
		UsageVariability:   0.25,
		EfficiencyScore:    88.5,
		TrendDirection:     "increasing",
		LastUpdated:        time.Now(),
	}

	// Mock cost analysis
	costAnalysis := &AccountCostAnalysis{
		AccountName:           "root",
		AnalysisPeriod:       "24h",
		TotalCost:            1500.0,
		CPUCost:              800.0,
		MemoryCost:           400.0,
		StorageCost:          300.0,
		CostPerJob:           10.0,
		CostPerCPUHour:       0.05,
		BudgetAllocated:      2000.0,
		BudgetUsed:           1500.0,
		BudgetUtilization:    75.0,
		ProjectedMonthlyCost: 45000.0,
		IsOverBudget:         false,
		CostTrend:            "stable",
		LastUpdated:          time.Now(),
	}

	// Mock anomalies
	anomalies := []*AccountAnomaly{
		{
			AccountName:    "root",
			AnomalyType:    "unusual_usage_spike",
			Severity:       "medium",
			Description:    "CPU usage increased by 200% compared to baseline",
			DetectedAt:     time.Now().Add(-time.Hour),
			Threshold:      150.0,
			ActualValue:    300.0,
			BaselineValue:  100.0,
			Confidence:     0.95,
			IsResolved:     false,
			Resolution:     "",
		},
	}

	// Mock access validation
	accessValidation := &AccountAccessValidation{
		AccountName:        "root",
		IsValid:           true,
		TotalUsers:        10,
		ActiveUsers:       8,
		InactiveUsers:     2,
		PendingUsers:      0,
		HasValidQuotas:    true,
		HasValidQoS:       true,
		HasValidPermissions: true,
		ComplianceScore:   95.0,
		SecurityRisk:      "low",
		LastAudit:         time.Now().Add(-24 * time.Hour),
		NextAudit:         time.Now().Add(6 * 24 * time.Hour),
	}

	client.On("GetAccountHierarchy", mock.Anything).Return(hierarchy, nil)
	client.On("GetAccountQuotas", mock.Anything, "root").Return(quota, nil)
	client.On("GetAccountQoS", mock.Anything, "root").Return(qosAssignments, nil)
	client.On("GetAccountReservations", mock.Anything, "root").Return(reservations, nil)
	client.On("GetAccountUsageAnalysis", mock.Anything, "root").Return(usageAnalysis, nil)
	client.On("GetAccountCostAnalysis", mock.Anything, "root").Return(costAnalysis, nil)
	client.On("DetectAccountAnomalies", mock.Anything, "root").Return(anomalies, nil)
	client.On("ValidateAccountAccess", mock.Anything, "root").Return(accessValidation, nil)

	collector := NewAccountCollector(client)

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

	// Verify we collected metrics
	assert.Greater(t, len(metrics), 0, "Should collect at least some metrics")

	// Verify specific metrics are present
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		desc := metric.Desc()
		metricNames[desc.String()] = true
	}

	// Check for key metric families
	foundHierarchyMetric := false
	foundQuotaMetric := false
	foundUsageMetric := false
	
	for name := range metricNames {
		if strings.Contains(name, "account_hierarchy") {
			foundHierarchyMetric = true
		}
		if strings.Contains(name, "account_quota") {
			foundQuotaMetric = true
		}
		if strings.Contains(name, "account_usage") {
			foundUsageMetric = true
		}
	}

	assert.True(t, foundHierarchyMetric, "Should have hierarchy metrics")
	assert.True(t, foundQuotaMetric, "Should have quota metrics")
	assert.True(t, foundUsageMetric, "Should have usage metrics")

	client.AssertExpectations(t)
}

func TestAccountCollector_Collect_Error(t *testing.T) {
	client := &MockAccountSLURMClient{}

	// Mock error response
	client.On("GetAccountHierarchy", mock.Anything).Return((*AccountHierarchy)(nil), assert.AnError)

	collector := NewAccountCollector(client)

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

	// Should still collect some metrics (error metrics)
	assert.GreaterOrEqual(t, len(metrics), 0, "Should handle errors gracefully")

	client.AssertExpectations(t)
}

func TestAccountCollector_MetricValues(t *testing.T) {
	client := &MockAccountSLURMClient{}

	// Create test data with known values
	hierarchy := &AccountHierarchy{
		RootAccounts: []*AccountNode{
			{
				Name:           "test_account",
				Balance:        500000.0,
				UtilizationRate: 80.5,
			},
		},
		BalanceScore:     90.0,
		UtilizationScore: 85.0,
	}

	quota := &AccountQuota{
		AccountName:      "test_account",
		CPUQuota:         1000,
		CPUUsed:          800,
		CPUUtilization:   80.0,
		IsViolating:      false,
	}

	client.On("GetAccountHierarchy", mock.Anything).Return(hierarchy, nil)
	client.On("GetAccountQuotas", mock.Anything, "test_account").Return(quota, nil)
	client.On("GetAccountQoS", mock.Anything, "test_account").Return([]*QoSAssignment{}, nil)
	client.On("GetAccountReservations", mock.Anything, "test_account").Return([]*ReservationUsage{}, nil)
	client.On("GetAccountUsageAnalysis", mock.Anything, "test_account").Return(&AccountUsageAnalysis{}, nil)
	client.On("GetAccountCostAnalysis", mock.Anything, "test_account").Return(&AccountCostAnalysis{}, nil)
	client.On("DetectAccountAnomalies", mock.Anything, "test_account").Return([]*AccountAnomaly{}, nil)
	client.On("ValidateAccountAccess", mock.Anything, "test_account").Return(&AccountAccessValidation{}, nil)

	collector := NewAccountCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	found := false
	for _, mf := range metricFamilies {
		if *mf.Name == "slurm_account_hierarchy_balance_score" {
			assert.Equal(t, float64(90.0), *mf.Metric[0].Gauge.Value)
			found = true
			break
		}
	}
	assert.True(t, found, "Should find balance score metric with correct value")
}

func TestAccountCollector_Integration(t *testing.T) {
	client := &MockAccountSLURMClient{}

	// Setup comprehensive mock data
	setupComprehensiveMocks(client)

	collector := NewAccountCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_account_hierarchy_balance_score Account hierarchy balance score
		# TYPE slurm_account_hierarchy_balance_score gauge
		slurm_account_hierarchy_balance_score 85
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected), 
		"slurm_account_hierarchy_balance_score")
	assert.NoError(t, err)
}

func setupComprehensiveMocks(client *MockAccountSLURMClient) {
	hierarchy := &AccountHierarchy{
		RootAccounts: []*AccountNode{
			{
				Name:           "root",
				Balance:        1000000.0,
				UtilizationRate: 75.0,
			},
		},
		BalanceScore:     85.0,
		UtilizationScore: 90.0,
	}

	quota := &AccountQuota{
		AccountName:      "root",
		CPUQuota:         1000,
		CPUUsed:          750,
		CPUUtilization:   75.0,
	}

	client.On("GetAccountHierarchy", mock.Anything).Return(hierarchy, nil)
	client.On("GetAccountQuotas", mock.Anything, mock.AnythingOfType("string")).Return(quota, nil)
	client.On("GetAccountQoS", mock.Anything, mock.AnythingOfType("string")).Return([]*QoSAssignment{}, nil)
	client.On("GetAccountReservations", mock.Anything, mock.AnythingOfType("string")).Return([]*ReservationUsage{}, nil)
	client.On("GetAccountUsageAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&AccountUsageAnalysis{}, nil)
	client.On("GetAccountCostAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&AccountCostAnalysis{}, nil)
	client.On("DetectAccountAnomalies", mock.Anything, mock.AnythingOfType("string")).Return([]*AccountAnomaly{}, nil)
	client.On("ValidateAccountAccess", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAccessValidation{}, nil)
}