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

type MockAccountCostTrackingSLURMClient struct {
	mock.Mock
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostMetrics(ctx context.Context, account string) (*AccountCostMetrics, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*AccountCostMetrics), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountBudgetInfo(ctx context.Context, account string) (*AccountBudgetInfo, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*AccountBudgetInfo), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostHistory(ctx context.Context, account string, startTime, endTime time.Time) ([]*AccountCostHistoryEntry, error) {
	args := m.Called(ctx, account, startTime, endTime)
	return args.Get(0).([]*AccountCostHistoryEntry), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostBreakdown(ctx context.Context, account string) (*AccountCostBreakdown, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*AccountCostBreakdown), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostForecasts(ctx context.Context, account string) (*AccountCostForecasts, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*AccountCostForecasts), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostAlerts(ctx context.Context, account string) ([]*AccountCostAlert, error) {
	args := m.Called(ctx, account)
	return args.Get(0).([]*AccountCostAlert), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostOptimizations(ctx context.Context, account string) ([]*AccountCostOptimization, error) {
	args := m.Called(ctx, account)
	return args.Get(0).([]*AccountCostOptimization), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostComparisons(ctx context.Context, account string) (*AccountCostComparisons, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*AccountCostComparisons), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountBudgetUtilization(ctx context.Context, account string) (*AccountBudgetUtilization, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*AccountBudgetUtilization), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostTrends(ctx context.Context, account string) (*AccountCostTrends, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*AccountCostTrends), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostAnalytics(ctx context.Context, account string) (*AccountCostAnalytics, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(*AccountCostAnalytics), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostReports(ctx context.Context, account string) ([]*AccountCostReport, error) {
	args := m.Called(ctx, account)
	return args.Get(0).([]*AccountCostReport), args.Error(1)
}

func (m *MockAccountCostTrackingSLURMClient) GetAccountCostPolicies(ctx context.Context, account string) ([]*AccountCostPolicy, error) {
	args := m.Called(ctx, account)
	return args.Get(0).([]*AccountCostPolicy), args.Error(1)
}

func TestAccountCostTrackingCollector_CostMetrics(t *testing.T) {
	mockClient := new(MockAccountCostTrackingSLURMClient)
	collector := NewAccountCostTrackingCollector(mockClient)

	costMetrics := &AccountCostMetrics{
		AccountName:          "test-account",
		TotalCost:            15000.50,
		CostToDate:           12000.25,
		EstimatedMonthlyCost: 18000.75,
		CostPerCPUHour:       0.12,
		CostPerGPUHour:       2.50,
		CostPerMemoryGBHour:  0.05,
		CostPerStorageGBHour: 0.02,
		AverageDailyCost:     500.25,
		PeakDailyCost:        800.50,
		CostEfficiency:       0.85,
		CostPerJob:           25.75,
		CostPerUser:          300.50,
		LastUpdated:          time.Now(),
	}

	budgetInfo := &AccountBudgetInfo{
		AccountName:        "test-account",
		TotalBudget:        20000.00,
		RemainingBudget:    8000.00,
		BudgetPeriod:       "monthly",
		BudgetStartDate:    time.Now().AddDate(0, -1, 0),
		BudgetEndDate:      time.Now().AddDate(0, 1, 0),
		BudgetUtilization:  0.60,
		DaysRemaining:      15,
		BurnRate:           533.33,
		ProjectedOverrun:   2000.00,
		BudgetStatus:       "warning",
		LastBudgetReset:    time.Now().AddDate(0, -1, 0),
		BudgetAlertLevel:   "medium",
	}

	costBreakdown := &AccountCostBreakdown{
		AccountName:       "test-account",
		ComputeCost:       8000.00,
		StorageCost:       2000.00,
		NetworkCost:       500.00,
		LicenseCost:       1000.00,
		SupportCost:       800.00,
		OverheadCost:      700.00,
		CPUCost:           6000.00,
		GPUCost:           2000.00,
		MemoryCost:        1500.00,
		LocalStorageCost:  1000.00,
		SharedStorageCost: 1000.00,
		BackupCost:        200.00,
		MonitoringCost:    300.00,
		SecurityCost:      500.00,
	}

	costForecasts := &AccountCostForecasts{
		AccountName:             "test-account",
		WeeklyForecast:          3500.00,
		MonthlyForecast:         15000.00,
		QuarterlyForecast:       45000.00,
		AnnualForecast:          180000.00,
		TrendDirection:          "up",
		ConfidenceLevel:         0.85,
		SeasonalFactors:         map[string]float64{"Q1": 1.1, "Q2": 0.9, "Q3": 1.0, "Q4": 1.2},
		GrowthRate:              0.15,
		ProjectedBudgetOverrun:  5000.00,
		OptimisticForecast:      12000.00,
		PessimisticForecast:     18000.00,
		ForecastAccuracy:        0.92,
		LastForecastUpdate:      time.Now(),
	}

	costAlerts := []*AccountCostAlert{
		{
			AlertID:          "alert-001",
			AccountName:      "test-account",
			AlertType:        "budget_threshold",
			Severity:         "high",
			Message:          "Budget threshold exceeded",
			CurrentValue:     12000.00,
			ThresholdValue:   10000.00,
			AlertTime:        time.Now(),
			Status:           "active",
			AcknowledgedBy:   "",
			AcknowledgedTime: time.Time{},
			ResolvedTime:     time.Time{},
			EscalationLevel:  1,
			AutoResolution:   false,
		},
		{
			AlertID:          "alert-002",
			AccountName:      "test-account",
			AlertType:        "cost_spike",
			Severity:         "medium",
			Message:          "Unusual cost spike detected",
			CurrentValue:     800.00,
			ThresholdValue:   600.00,
			AlertTime:        time.Now().Add(-time.Hour),
			Status:           "resolved",
			AcknowledgedBy:   "admin",
			AcknowledgedTime: time.Now().Add(-30 * time.Minute),
			ResolvedTime:     time.Now().Add(-10 * time.Minute),
			EscalationLevel:  0,
			AutoResolution:   true,
		},
	}

	costOptimizations := []*AccountCostOptimization{
		{
			OptimizationID:       "opt-001",
			AccountName:          "test-account",
			OptimizationType:     "resource_rightsizing",
			Description:          "Optimize resource allocation",
			EstimatedSavings:     2000.00,
			ImplementationEffort: "medium",
			Priority:             "high",
			Status:               "active",
			CreatedTime:          time.Now().AddDate(0, 0, -7),
			ImplementedTime:      time.Time{},
			ActualSavings:        0.00,
			ROI:                  0.00,
			Recommendation:       "Reduce oversized instances",
		},
		{
			OptimizationID:       "opt-002",
			AccountName:          "test-account",
			OptimizationType:     "scheduling_optimization",
			Description:          "Optimize job scheduling",
			EstimatedSavings:     1500.00,
			ImplementationEffort: "low",
			Priority:             "medium",
			Status:               "implemented",
			CreatedTime:          time.Now().AddDate(0, 0, -14),
			ImplementedTime:      time.Now().AddDate(0, 0, -3),
			ActualSavings:        1200.00,
			ROI:                  4.5,
			Recommendation:       "Use spot instances for non-critical jobs",
		},
	}

	costComparisons := &AccountCostComparisons{
		AccountName:             "test-account",
		PeerAccountsAvgCost:     14500.00,
		SimilarWorkloadsAvgCost: 13800.00,
		IndustryBenchmark:       16000.00,
		CostRanking:             8,
		TotalAccounts:           25,
		CostPercentile:          0.68,
		EfficiencyRanking:       5,
		CostVariance:            1200.00,
		ComparisonPeriod:        "monthly",
		LastComparison:          time.Now(),
	}

	budgetUtilization := &AccountBudgetUtilization{
		AccountName:           "test-account",
		CurrentUtilization:    0.60,
		ProjectedUtilization:  0.90,
		UtilizationTrend:      "increasing",
		DailyBurnRate:         533.33,
		OptimalBurnRate:       500.00,
		BurnRateVariance:      0.15,
		TimeToDepletion:       15,
		BudgetHealth:          "warning",
		RecommendedAdjustment: -50.00,
		UtilizationHistory:    []float64{0.45, 0.50, 0.55, 0.60},
		LastUtilizationUpdate: time.Now(),
	}

	costTrends := &AccountCostTrends{
		AccountName:         "test-account",
		ShortTermTrend:      "increasing",
		LongTermTrend:       "stable",
		SeasonalPattern:     "quarterly_peaks",
		CostVolatility:      0.25,
		TrendStrength:       0.75,
		CyclicalPatterns:    []string{"weekly", "monthly"},
		AnomalyCount:        3,
		TrendChangePoints:   []time.Time{time.Now().AddDate(0, -2, 0)},
		PredictiveAccuracy:  0.88,
		TrendConfidence:     0.92,
		LastTrendAnalysis:   time.Now(),
	}

	costAnalytics := &AccountCostAnalytics{
		AccountName:               "test-account",
		CostElasticity:            0.65,
		UsageCorrelation:          0.82,
		SeasonalityIndex:          0.15,
		CostPredictability:        0.78,
		ResourceUtilizationImpact: 0.45,
		UserBehaviorImpact:        0.25,
		WorkloadPatternImpact:     0.30,
		ExternalFactorImpact:      0.10,
		CostDrivers:               []string{"compute", "storage", "licenses"},
		CostInhibitors:            []string{"efficiency_programs", "automation"},
		AnalyticsScore:            0.85,
		LastAnalyticsUpdate:       time.Now(),
	}

	costPolicies := []*AccountCostPolicy{
		{
			PolicyID:        "policy-001",
			AccountName:     "test-account",
			PolicyType:      "budget_limit",
			PolicyName:      "Monthly Budget Limit",
			Description:     "Enforce monthly budget limits",
			CostLimit:       20000.00,
			TimeWindow:      "monthly",
			EnforcementMode: "strict",
			Violations:      2,
			LastViolation:   time.Now().AddDate(0, 0, -5),
			CreatedTime:     time.Now().AddDate(0, -6, 0),
			ModifiedTime:    time.Now().AddDate(0, -1, 0),
			Active:          true,
		},
		{
			PolicyID:        "policy-002",
			AccountName:     "test-account",
			PolicyType:      "cost_spike",
			PolicyName:      "Cost Spike Detection",
			Description:     "Detect and alert on cost spikes",
			CostLimit:       1000.00,
			TimeWindow:      "daily",
			EnforcementMode: "warning",
			Violations:      1,
			LastViolation:   time.Now().AddDate(0, 0, -2),
			CreatedTime:     time.Now().AddDate(0, -3, 0),
			ModifiedTime:    time.Now().AddDate(0, 0, -10),
			Active:          true,
		},
	}

	mockClient.On("GetAccountCostMetrics", mock.Anything, "account1").Return(costMetrics, nil)
	mockClient.On("GetAccountCostMetrics", mock.Anything, "account2").Return(costMetrics, nil)
	mockClient.On("GetAccountCostMetrics", mock.Anything, "account3").Return(costMetrics, nil)

	mockClient.On("GetAccountBudgetInfo", mock.Anything, "account1").Return(budgetInfo, nil)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, "account2").Return(budgetInfo, nil)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, "account3").Return(budgetInfo, nil)

	mockClient.On("GetAccountCostBreakdown", mock.Anything, "account1").Return(costBreakdown, nil)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, "account2").Return(costBreakdown, nil)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, "account3").Return(costBreakdown, nil)

	mockClient.On("GetAccountCostForecasts", mock.Anything, "account1").Return(costForecasts, nil)
	mockClient.On("GetAccountCostForecasts", mock.Anything, "account2").Return(costForecasts, nil)
	mockClient.On("GetAccountCostForecasts", mock.Anything, "account3").Return(costForecasts, nil)

	mockClient.On("GetAccountCostAlerts", mock.Anything, "account1").Return(costAlerts, nil)
	mockClient.On("GetAccountCostAlerts", mock.Anything, "account2").Return(costAlerts, nil)
	mockClient.On("GetAccountCostAlerts", mock.Anything, "account3").Return(costAlerts, nil)

	mockClient.On("GetAccountCostOptimizations", mock.Anything, "account1").Return(costOptimizations, nil)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, "account2").Return(costOptimizations, nil)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, "account3").Return(costOptimizations, nil)

	mockClient.On("GetAccountCostComparisons", mock.Anything, "account1").Return(costComparisons, nil)
	mockClient.On("GetAccountCostComparisons", mock.Anything, "account2").Return(costComparisons, nil)
	mockClient.On("GetAccountCostComparisons", mock.Anything, "account3").Return(costComparisons, nil)

	mockClient.On("GetAccountBudgetUtilization", mock.Anything, "account1").Return(budgetUtilization, nil)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, "account2").Return(budgetUtilization, nil)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, "account3").Return(budgetUtilization, nil)

	mockClient.On("GetAccountCostTrends", mock.Anything, "account1").Return(costTrends, nil)
	mockClient.On("GetAccountCostTrends", mock.Anything, "account2").Return(costTrends, nil)
	mockClient.On("GetAccountCostTrends", mock.Anything, "account3").Return(costTrends, nil)

	mockClient.On("GetAccountCostAnalytics", mock.Anything, "account1").Return(costAnalytics, nil)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, "account2").Return(costAnalytics, nil)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, "account3").Return(costAnalytics, nil)

	mockClient.On("GetAccountCostPolicies", mock.Anything, "account1").Return(costPolicies, nil)
	mockClient.On("GetAccountCostPolicies", mock.Anything, "account2").Return(costPolicies, nil)
	mockClient.On("GetAccountCostPolicies", mock.Anything, "account3").Return(costPolicies, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountCostTrackingCollector_CostMetricsValues(t *testing.T) {
	mockClient := new(MockAccountCostTrackingSLURMClient)
	collector := NewAccountCostTrackingCollector(mockClient)

	costMetrics := &AccountCostMetrics{
		AccountName:          "test-account",
		TotalCost:            15000.50,
		CostToDate:           12000.25,
		EstimatedMonthlyCost: 18000.75,
		CostPerCPUHour:       0.12,
		CostPerGPUHour:       2.50,
		CostEfficiency:       0.85,
		CostPerJob:           25.75,
		CostPerUser:          300.50,
	}

	mockClient.On("GetAccountCostMetrics", mock.Anything, "account1").Return(costMetrics, nil)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, mock.Anything).Return(&AccountBudgetInfo{}, nil)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, mock.Anything).Return(&AccountCostBreakdown{}, nil)
	mockClient.On("GetAccountCostForecasts", mock.Anything, mock.Anything).Return(&AccountCostForecasts{TrendDirection: "stable"}, nil)
	mockClient.On("GetAccountCostAlerts", mock.Anything, mock.Anything).Return([]*AccountCostAlert{}, nil)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, mock.Anything).Return([]*AccountCostOptimization{}, nil)
	mockClient.On("GetAccountCostComparisons", mock.Anything, mock.Anything).Return(&AccountCostComparisons{}, nil)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, mock.Anything).Return(&AccountBudgetUtilization{}, nil)
	mockClient.On("GetAccountCostTrends", mock.Anything, mock.Anything).Return(&AccountCostTrends{}, nil)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, mock.Anything).Return(&AccountCostAnalytics{}, nil)
	mockClient.On("GetAccountCostPolicies", mock.Anything, mock.Anything).Return([]*AccountCostPolicy{}, nil)

	// Test total cost metric
	expectedCost := `
		# HELP slurm_account_total_cost Total cost for account
		# TYPE slurm_account_total_cost gauge
		slurm_account_total_cost{account="account1"} 15000.5
	`
	err := testutil.CollectAndCompare(collector, strings.NewReader(expectedCost), "slurm_account_total_cost")
	assert.NoError(t, err)

	// Test cost efficiency metric
	expectedEfficiency := `
		# HELP slurm_account_cost_efficiency Cost efficiency score for account
		# TYPE slurm_account_cost_efficiency gauge
		slurm_account_cost_efficiency{account="account1"} 0.85
	`
	err = testutil.CollectAndCompare(collector, strings.NewReader(expectedEfficiency), "slurm_account_cost_efficiency")
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestAccountCostTrackingCollector_BudgetMetrics(t *testing.T) {
	mockClient := new(MockAccountCostTrackingSLURMClient)
	collector := NewAccountCostTrackingCollector(mockClient)

	budgetInfo := &AccountBudgetInfo{
		AccountName:        "test-account",
		TotalBudget:        20000.00,
		RemainingBudget:    8000.00,
		BudgetPeriod:       "monthly",
		BudgetUtilization:  0.60,
		DaysRemaining:      15,
		BurnRate:           533.33,
		ProjectedOverrun:   2000.00,
		BudgetStatus:       "warning",
		BudgetAlertLevel:   "medium",
	}

	mockClient.On("GetAccountCostMetrics", mock.Anything, mock.Anything).Return(&AccountCostMetrics{}, nil)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, "account1").Return(budgetInfo, nil)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, mock.Anything).Return(&AccountCostBreakdown{}, nil)
	mockClient.On("GetAccountCostForecasts", mock.Anything, mock.Anything).Return(&AccountCostForecasts{TrendDirection: "stable"}, nil)
	mockClient.On("GetAccountCostAlerts", mock.Anything, mock.Anything).Return([]*AccountCostAlert{}, nil)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, mock.Anything).Return([]*AccountCostOptimization{}, nil)
	mockClient.On("GetAccountCostComparisons", mock.Anything, mock.Anything).Return(&AccountCostComparisons{}, nil)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, mock.Anything).Return(&AccountBudgetUtilization{}, nil)
	mockClient.On("GetAccountCostTrends", mock.Anything, mock.Anything).Return(&AccountCostTrends{}, nil)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, mock.Anything).Return(&AccountCostAnalytics{}, nil)
	mockClient.On("GetAccountCostPolicies", mock.Anything, mock.Anything).Return([]*AccountCostPolicy{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect budget metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountCostTrackingCollector_ForecastMetrics(t *testing.T) {
	mockClient := new(MockAccountCostTrackingSLURMClient)
	collector := NewAccountCostTrackingCollector(mockClient)

	costForecasts := &AccountCostForecasts{
		AccountName:             "test-account",
		WeeklyForecast:          3500.00,
		MonthlyForecast:         15000.00,
		QuarterlyForecast:       45000.00,
		AnnualForecast:          180000.00,
		TrendDirection:          "up",
		ConfidenceLevel:         0.85,
		GrowthRate:              0.15,
		ForecastAccuracy:        0.92,
	}

	mockClient.On("GetAccountCostMetrics", mock.Anything, mock.Anything).Return(&AccountCostMetrics{}, nil)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, mock.Anything).Return(&AccountBudgetInfo{}, nil)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, mock.Anything).Return(&AccountCostBreakdown{}, nil)
	mockClient.On("GetAccountCostForecasts", mock.Anything, "account1").Return(costForecasts, nil)
	mockClient.On("GetAccountCostAlerts", mock.Anything, mock.Anything).Return([]*AccountCostAlert{}, nil)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, mock.Anything).Return([]*AccountCostOptimization{}, nil)
	mockClient.On("GetAccountCostComparisons", mock.Anything, mock.Anything).Return(&AccountCostComparisons{}, nil)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, mock.Anything).Return(&AccountBudgetUtilization{}, nil)
	mockClient.On("GetAccountCostTrends", mock.Anything, mock.Anything).Return(&AccountCostTrends{}, nil)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, mock.Anything).Return(&AccountCostAnalytics{}, nil)
	mockClient.On("GetAccountCostPolicies", mock.Anything, mock.Anything).Return([]*AccountCostPolicy{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect forecast metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountCostTrackingCollector_AlertMetrics(t *testing.T) {
	mockClient := new(MockAccountCostTrackingSLURMClient)
	collector := NewAccountCostTrackingCollector(mockClient)

	costAlerts := []*AccountCostAlert{
		{
			AlertID:         "alert-001",
			AccountName:     "test-account",
			AlertType:       "budget_threshold",
			Severity:        "high",
			Status:          "active",
			EscalationLevel: 1,
		},
		{
			AlertID:         "alert-002",
			AccountName:     "test-account",
			AlertType:       "cost_spike",
			Severity:        "medium",
			Status:          "resolved",
			EscalationLevel: 0,
		},
	}

	mockClient.On("GetAccountCostMetrics", mock.Anything, mock.Anything).Return(&AccountCostMetrics{}, nil)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, mock.Anything).Return(&AccountBudgetInfo{}, nil)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, mock.Anything).Return(&AccountCostBreakdown{}, nil)
	mockClient.On("GetAccountCostForecasts", mock.Anything, mock.Anything).Return(&AccountCostForecasts{TrendDirection: "stable"}, nil)
	mockClient.On("GetAccountCostAlerts", mock.Anything, "account1").Return(costAlerts, nil)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, mock.Anything).Return([]*AccountCostOptimization{}, nil)
	mockClient.On("GetAccountCostComparisons", mock.Anything, mock.Anything).Return(&AccountCostComparisons{}, nil)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, mock.Anything).Return(&AccountBudgetUtilization{}, nil)
	mockClient.On("GetAccountCostTrends", mock.Anything, mock.Anything).Return(&AccountCostTrends{}, nil)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, mock.Anything).Return(&AccountCostAnalytics{}, nil)
	mockClient.On("GetAccountCostPolicies", mock.Anything, mock.Anything).Return([]*AccountCostPolicy{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect alert metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountCostTrackingCollector_OptimizationMetrics(t *testing.T) {
	mockClient := new(MockAccountCostTrackingSLURMClient)
	collector := NewAccountCostTrackingCollector(mockClient)

	costOptimizations := []*AccountCostOptimization{
		{
			OptimizationID:   "opt-001",
			AccountName:      "test-account",
			OptimizationType: "resource_rightsizing",
			EstimatedSavings: 2000.00,
			Priority:         "high",
			Status:           "active",
			ActualSavings:    0.00,
			ROI:              0.00,
		},
		{
			OptimizationID:   "opt-002",
			AccountName:      "test-account",
			OptimizationType: "scheduling_optimization",
			EstimatedSavings: 1500.00,
			Priority:         "medium",
			Status:           "implemented",
			ActualSavings:    1200.00,
			ROI:              4.5,
		},
	}

	mockClient.On("GetAccountCostMetrics", mock.Anything, mock.Anything).Return(&AccountCostMetrics{}, nil)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, mock.Anything).Return(&AccountBudgetInfo{}, nil)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, mock.Anything).Return(&AccountCostBreakdown{}, nil)
	mockClient.On("GetAccountCostForecasts", mock.Anything, mock.Anything).Return(&AccountCostForecasts{TrendDirection: "stable"}, nil)
	mockClient.On("GetAccountCostAlerts", mock.Anything, mock.Anything).Return([]*AccountCostAlert{}, nil)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, "account1").Return(costOptimizations, nil)
	mockClient.On("GetAccountCostComparisons", mock.Anything, mock.Anything).Return(&AccountCostComparisons{}, nil)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, mock.Anything).Return(&AccountBudgetUtilization{}, nil)
	mockClient.On("GetAccountCostTrends", mock.Anything, mock.Anything).Return(&AccountCostTrends{}, nil)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, mock.Anything).Return(&AccountCostAnalytics{}, nil)
	mockClient.On("GetAccountCostPolicies", mock.Anything, mock.Anything).Return([]*AccountCostPolicy{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect optimization metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountCostTrackingCollector_PolicyMetrics(t *testing.T) {
	mockClient := new(MockAccountCostTrackingSLURMClient)
	collector := NewAccountCostTrackingCollector(mockClient)

	costPolicies := []*AccountCostPolicy{
		{
			PolicyID:        "policy-001",
			AccountName:     "test-account",
			PolicyType:      "budget_limit",
			CostLimit:       20000.00,
			Violations:      2,
			Active:          true,
		},
		{
			PolicyID:        "policy-002",
			AccountName:     "test-account",
			PolicyType:      "cost_spike",
			CostLimit:       1000.00,
			Violations:      1,
			Active:          true,
		},
	}

	mockClient.On("GetAccountCostMetrics", mock.Anything, mock.Anything).Return(&AccountCostMetrics{}, nil)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, mock.Anything).Return(&AccountBudgetInfo{}, nil)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, mock.Anything).Return(&AccountCostBreakdown{}, nil)
	mockClient.On("GetAccountCostForecasts", mock.Anything, mock.Anything).Return(&AccountCostForecasts{TrendDirection: "stable"}, nil)
	mockClient.On("GetAccountCostAlerts", mock.Anything, mock.Anything).Return([]*AccountCostAlert{}, nil)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, mock.Anything).Return([]*AccountCostOptimization{}, nil)
	mockClient.On("GetAccountCostComparisons", mock.Anything, mock.Anything).Return(&AccountCostComparisons{}, nil)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, mock.Anything).Return(&AccountBudgetUtilization{}, nil)
	mockClient.On("GetAccountCostTrends", mock.Anything, mock.Anything).Return(&AccountCostTrends{}, nil)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, mock.Anything).Return(&AccountCostAnalytics{}, nil)
	mockClient.On("GetAccountCostPolicies", mock.Anything, "account1").Return(costPolicies, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect policy metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountCostTrackingCollector_ErrorHandling(t *testing.T) {
	mockClient := new(MockAccountCostTrackingSLURMClient)
	collector := NewAccountCostTrackingCollector(mockClient)

	mockClient.On("GetAccountCostMetrics", mock.Anything, mock.Anything).Return((*AccountCostMetrics)(nil), assert.AnError)
	mockClient.On("GetAccountBudgetInfo", mock.Anything, mock.Anything).Return((*AccountBudgetInfo)(nil), assert.AnError)
	mockClient.On("GetAccountCostBreakdown", mock.Anything, mock.Anything).Return((*AccountCostBreakdown)(nil), assert.AnError)
	mockClient.On("GetAccountCostForecasts", mock.Anything, mock.Anything).Return((*AccountCostForecasts)(nil), assert.AnError)
	mockClient.On("GetAccountCostAlerts", mock.Anything, mock.Anything).Return(([]*AccountCostAlert)(nil), assert.AnError)
	mockClient.On("GetAccountCostOptimizations", mock.Anything, mock.Anything).Return(([]*AccountCostOptimization)(nil), assert.AnError)
	mockClient.On("GetAccountCostComparisons", mock.Anything, mock.Anything).Return((*AccountCostComparisons)(nil), assert.AnError)
	mockClient.On("GetAccountBudgetUtilization", mock.Anything, mock.Anything).Return((*AccountBudgetUtilization)(nil), assert.AnError)
	mockClient.On("GetAccountCostTrends", mock.Anything, mock.Anything).Return((*AccountCostTrends)(nil), assert.AnError)
	mockClient.On("GetAccountCostAnalytics", mock.Anything, mock.Anything).Return((*AccountCostAnalytics)(nil), assert.AnError)
	mockClient.On("GetAccountCostPolicies", mock.Anything, mock.Anything).Return(([]*AccountCostPolicy)(nil), assert.AnError)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Equal(t, 0, metricCount, "Should not collect metrics on error")

	mockClient.AssertExpectations(t)
}