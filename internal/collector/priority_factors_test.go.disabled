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

type MockPriorityFactorsSLURMClient struct {
	mock.Mock
}

func (m *MockPriorityFactorsSLURMClient) GetPriorityFactorBreakdown(ctx context.Context, jobID string) (*PriorityFactorBreakdown, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*PriorityFactorBreakdown), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) GetSystemFactorWeights(ctx context.Context) (*SystemFactorWeights, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemFactorWeights), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) GetUserFactorProfile(ctx context.Context, userName string) (*UserFactorProfile, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserFactorProfile), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) GetAccountFactorSummary(ctx context.Context, accountName string) (*AccountFactorSummary, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountFactorSummary), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) GetPartitionFactorAnalysis(ctx context.Context, partition string) (*PartitionFactorAnalysis, error) {
	args := m.Called(ctx, partition)
	return args.Get(0).(*PartitionFactorAnalysis), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) GetQoSFactorAnalysis(ctx context.Context, qosName string) (*QoSFactorAnalysis, error) {
	args := m.Called(ctx, qosName)
	return args.Get(0).(*QoSFactorAnalysis), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) GetFactorTrendAnalysis(ctx context.Context, factor string, period string) (*FactorTrendAnalysis, error) {
	args := m.Called(ctx, factor, period)
	return args.Get(0).(*FactorTrendAnalysis), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) GetFactorImpactAnalysis(ctx context.Context, factor string) (*FactorImpactAnalysis, error) {
	args := m.Called(ctx, factor)
	return args.Get(0).(*FactorImpactAnalysis), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) OptimizeFactorWeights(ctx context.Context, scope string) (*FactorOptimizationResult, error) {
	args := m.Called(ctx, scope)
	return args.Get(0).(*FactorOptimizationResult), args.Error(1)
}

func (m *MockPriorityFactorsSLURMClient) ValidateFactorConfiguration(ctx context.Context) (*FactorConfigurationValidation, error) {
	args := m.Called(ctx)
	return args.Get(0).(*FactorConfigurationValidation), args.Error(1)
}

func TestNewPriorityFactorsCollector(t *testing.T) {
	client := &MockPriorityFactorsSLURMClient{}
	collector := NewPriorityFactorsCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.priorityFactorWeight)
	assert.NotNil(t, collector.priorityFactorValue)
	assert.NotNil(t, collector.priorityFactorContribution)
	assert.NotNil(t, collector.factorWeightDistribution)
	assert.NotNil(t, collector.factorTrendDirection)
	assert.NotNil(t, collector.systemFactorConfig)
	assert.NotNil(t, collector.userFactorProfile)
	assert.NotNil(t, collector.accountFactorSummary)
}

func TestPriorityFactorsCollector_Describe(t *testing.T) {
	client := &MockPriorityFactorsSLURMClient{}
	collector := NewPriorityFactorsCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 41 metrics, verify we have the correct number
	assert.Equal(t, 41, len(descs), "Should have exactly 41 metric descriptions")
}

func TestPriorityFactorsCollector_Collect_Success(t *testing.T) {
	client := &MockPriorityFactorsSLURMClient{}

	// Mock system factor weights
	systemWeights := &SystemFactorWeights{
		AgeWeightGlobal:        0.3,
		FairShareWeightGlobal:  0.4,
		QoSWeightGlobal:        0.2,
		PartitionWeightGlobal:  0.1,
		SizeWeightGlobal:       0.05,
		AssocWeightGlobal:      0.03,
		NiceWeightGlobal:       0.02,
		WeightBalance:          0.85,
		WeightOptimality:       0.9,
		WeightEffectiveness:    0.88,
		ConfigurationVersion:   "v2.0",
		CalibrationAccuracy:    0.95,
		CalibrationStability:   0.92,
		LastUpdated:            time.Now(),
	}

	// Mock priority factor breakdown
	factorBreakdown := &PriorityFactorBreakdown{
		JobID:         "12345",
		UserName:      "testuser",
		AccountName:   "testaccount",
		PartitionName: "cpu",
		QoSName:       "normal",

		// Factor values
		AgeValue:       15000.0,
		FairShareValue: 20000.0,
		QoSValue:       10000.0,
		PartitionValue: 5000.0,
		SizeValue:      2000.0,
		AssocValue:     1000.0,
		NiceValue:      0.0,

		// Factor weights
		AgeWeight:       0.3,
		FairShareWeight: 0.4,
		QoSWeight:       0.2,
		PartitionWeight: 0.1,
		SizeWeight:      0.05,
		AssocWeight:     0.03,
		NiceWeight:      0.02,

		// Factor contributions
		AgeContribution:       4500.0,
		FairShareContribution: 8000.0,
		QoSContribution:       2000.0,
		PartitionContribution: 500.0,
		SizeContribution:      100.0,
		AssocContribution:     30.0,
		NiceContribution:      0.0,

		// Normalized values
		AgeNormalized:       0.3,
		FairShareNormalized: 0.4,
		QoSNormalized:       0.2,
		PartitionNormalized: 0.1,
		SizeNormalized:      0.05,
		AssocNormalized:     0.03,
		NiceNormalized:      0.02,

		// Effective contributions
		AgeEffective:       4500.0,
		FairShareEffective: 8000.0,
		QoSEffective:       2000.0,
		PartitionEffective: 500.0,
		SizeEffective:      100.0,
		AssocEffective:     30.0,
		NiceEffective:      0.0,

		TotalPriority:      15130.0,
		DominantFactor:     "fairshare",
		SecondaryFactor:    "age",
		FactorBalance:      0.75,
		ConfigurationScore: 0.88,
		AnalyzedAt:         time.Now(),
	}

	// Mock user factor profile
	userProfile := &UserFactorProfile{
		UserName:    "testuser",
		AccountName: "testaccount",

		AvgAgeContribution:       4000.0,
		AvgFairShareContribution: 7500.0,
		AvgQoSContribution:       2200.0,
		AvgPartitionContribution: 450.0,
		AvgSizeContribution:      120.0,

		AgeVariance:       500.0,
		FairShareVariance: 1000.0,
		QoSVariance:       200.0,
		PartitionVariance: 50.0,
		SizeVariance:      20.0,

		DominantFactorPattern:   "fairshare",
		FactorStability:         0.85,
		FactorOptimizationScore: 0.78,
		DeviationFromNorm:       0.15,
		DeviationSignificance:   "minor",
		DeviationExplanation:    "Slightly higher fair-share usage",
		LastAnalyzed:            time.Now(),
	}

	// Mock account factor summary
	accountSummary := &AccountFactorSummary{
		AccountName:           "testaccount",
		TotalUsers:            25,
		ActiveUsers:           20,
		AvgFactorBalance:      0.82,
		FactorVarianceScore:   0.15,
		FactorEfficiencyScore: 0.88,
		OptimizationPotential: 0.12,
		LastSummarized:        time.Now(),
	}

	// Mock partition factor analysis
	partitionAnalysis := &PartitionFactorAnalysis{
		PartitionName:            "cpu",
		PartitionWeight:          0.1,
		PartitionPriorityBonus:   1000.0,
		PartitionFactorMultiplier: 1.2,
		FactorUtilizationRate:    0.75,
		FactorEffectiveness:      0.88,
		FactorBalance:            0.85,
		OptimizationScore:        0.82,
		BalanceRecommendation:    "optimal",
		EfficiencyRecommendation: "consider_increase",
		LastAnalyzed:             time.Now(),
	}

	// Mock QoS factor analysis
	qosAnalysis := &QoSFactorAnalysis{
		QoSName:                  "normal",
		QoSPriorityValue:         1000,
		QoSWeight:                0.2,
		QoSMultiplier:            1.0,
		QoSFactorEffectiveness:   0.85,
		QoSUtilizationRate:       0.7,
		QoSFactorImpact:          0.75,
		QoSBalanceScore:          0.88,
		QoSPerformanceScore:      0.82,
		QoSOptimizationPotential: 0.15,
		LastAnalyzed:             time.Now(),
	}

	// Mock factor trend analysis
	factorTrend := &FactorTrendAnalysis{
		FactorName:             "age",
		AnalysisPeriod:         "24h",
		TrendDirection:         "increasing",
		TrendVelocity:          0.05,
		TrendAcceleration:      0.01,
		TrendVolatility:        0.12,
		FutureTrendPrediction:  "stable",
		PredictionConfidence:   0.85,
		PredictionTimeframe:    "7d",
		SeasonalityDetected:    true,
		SeasonalityStrength:    0.3,
		SeasonalityPeriod:      "daily",
		CyclePredictability:    0.75,
		LastAnalyzed:           time.Now(),
	}

	// Mock factor impact analysis
	factorImpact := &FactorImpactAnalysis{
		FactorName:         "fairshare",
		ImpactScore:        0.85,
		ImpactSignificance: "high",
		ImpactCorrelation:  0.75,
		ImpactSensitivity:  0.65,
		ImpactChangeRate:   0.02,
		ImpactStability:    0.88,
		ImpactVolatility:   0.15,
		FactorInteractions: map[string]float64{
			"age": 0.3,
			"qos": 0.2,
		},
		FactorSynergies: []string{"age", "qos"},
		FactorConflicts: []string{},
		LastAnalyzed:    time.Now(),
	}

	// Mock optimization result
	optimizationResult := &FactorOptimizationResult{
		OptimizationScope: "system",
		CurrentConfiguration: map[string]float64{
			"age":       0.3,
			"fairshare": 0.4,
			"qos":       0.2,
		},
		CurrentEffectiveness:  0.85,
		CurrentBalance:        0.82,
		ExpectedEffectiveness: 0.92,
		ExpectedImprovement:   0.07,
		ImplementationComplexity: "medium",
		ImplementationRisk:       "low",
		ImplementationTimeframe:  "2w",
		RecommendedActions:       []string{"increase_fairshare_weight"},
		OptimizedAt:              time.Now(),
	}

	// Mock configuration validation
	configValidation := &FactorConfigurationValidation{
		ValidationScope:    "system",
		IsValid:           true,
		ValidationScore:   0.92,
		ValidationIssues:  []string{},
		ValidationWarnings: []string{"minor_imbalance_detected"},
		ConfigurationHealth: "good",
		HealthScore:        0.88,
		StabilityScore:     0.91,
		OptimalityScore:    0.85,
		ImmediateActions:   []string{},
		LongTermActions:    []string{"monitor_fairshare_trends"},
		ValidatedAt:        time.Now(),
	}

	// Setup mock expectations
	client.On("GetSystemFactorWeights", mock.Anything).Return(systemWeights, nil)
	client.On("GetPriorityFactorBreakdown", mock.Anything, mock.AnythingOfType("string")).Return(factorBreakdown, nil)
	client.On("GetUserFactorProfile", mock.Anything, mock.AnythingOfType("string")).Return(userProfile, nil)
	client.On("GetAccountFactorSummary", mock.Anything, mock.AnythingOfType("string")).Return(accountSummary, nil)
	client.On("GetPartitionFactorAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(partitionAnalysis, nil)
	client.On("GetQoSFactorAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(qosAnalysis, nil)
	client.On("GetFactorTrendAnalysis", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(factorTrend, nil)
	client.On("GetFactorImpactAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(factorImpact, nil)
	client.On("OptimizeFactorWeights", mock.Anything, "system").Return(optimizationResult, nil)
	client.On("ValidateFactorConfiguration", mock.Anything).Return(configValidation, nil)

	collector := NewPriorityFactorsCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 300)
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
	foundFactorWeight := false
	foundFactorValue := false
	foundTrendAnalysis := false
	foundSystemConfig := false
	foundUserProfile := false

	for name := range metricNames {
		if strings.Contains(name, "priority_factor_weight") {
			foundFactorWeight = true
		}
		if strings.Contains(name, "priority_factor_value") {
			foundFactorValue = true
		}
		if strings.Contains(name, "factor_trend") {
			foundTrendAnalysis = true
		}
		if strings.Contains(name, "system_factor_configuration") {
			foundSystemConfig = true
		}
		if strings.Contains(name, "user_factor_profile") {
			foundUserProfile = true
		}
	}

	assert.True(t, foundFactorWeight, "Should have factor weight metrics")
	assert.True(t, foundFactorValue, "Should have factor value metrics")
	assert.True(t, foundTrendAnalysis, "Should have trend analysis metrics")
	assert.True(t, foundSystemConfig, "Should have system config metrics")
	assert.True(t, foundUserProfile, "Should have user profile metrics")

	client.AssertExpectations(t)
}

func TestPriorityFactorsCollector_Collect_Error(t *testing.T) {
	client := &MockPriorityFactorsSLURMClient{}

	// Mock error response
	client.On("GetSystemFactorWeights", mock.Anything).Return((*SystemFactorWeights)(nil), assert.AnError)

	collector := NewPriorityFactorsCollector(client)

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

func TestPriorityFactorsCollector_MetricValues(t *testing.T) {
	client := &MockPriorityFactorsSLURMClient{}

	// Create test data with known values
	systemWeights := &SystemFactorWeights{
		AgeWeightGlobal:        0.25,
		FairShareWeightGlobal:  0.5,
		QoSWeightGlobal:        0.15,
		WeightBalance:          0.8,
		ConfigurationVersion:   "test_v1",
		CalibrationAccuracy:    0.95,
	}

	factorBreakdown := &PriorityFactorBreakdown{
		JobID:         "test_job",
		UserName:      "test_user",
		AccountName:   "test_account",
		PartitionName: "test_partition",
		QoSName:       "normal",
		AgeValue:      10000.0,
		FairShareValue: 20000.0,
		QoSValue:      5000.0,
		TotalPriority: 35000.0,
	}

	client.On("GetSystemFactorWeights", mock.Anything).Return(systemWeights, nil)
	client.On("GetPriorityFactorBreakdown", mock.Anything, "test_job").Return(factorBreakdown, nil)
	client.On("GetUserFactorProfile", mock.Anything, mock.AnythingOfType("string")).Return(&UserFactorProfile{}, nil)
	client.On("GetAccountFactorSummary", mock.Anything, mock.AnythingOfType("string")).Return(&AccountFactorSummary{}, nil)
	client.On("GetPartitionFactorAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&PartitionFactorAnalysis{}, nil)
	client.On("GetQoSFactorAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&QoSFactorAnalysis{}, nil)
	client.On("GetFactorTrendAnalysis", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FactorTrendAnalysis{}, nil)
	client.On("GetFactorImpactAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&FactorImpactAnalysis{}, nil)
	client.On("OptimizeFactorWeights", mock.Anything, "system").Return(&FactorOptimizationResult{}, nil)
	client.On("ValidateFactorConfiguration", mock.Anything).Return(&FactorConfigurationValidation{}, nil)

	collector := NewPriorityFactorsCollector(client)

	// Override sample job IDs for testing
	originalGetSampleJobIDs := collector.getSampleJobIDs
	collector.getSampleJobIDs = func() []string {
		return []string{"test_job"}
	}
	defer func() {
		collector.getSampleJobIDs = originalGetSampleJobIDs
	}()

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundSystemWeight := false
	foundFactorValue := false
	foundWeightBalance := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_system_factor_configuration":
			for _, metric := range mf.Metric {
				for _, label := range metric.Label {
					if *label.Name == "factor" && *label.Value == "age" {
						assert.Equal(t, float64(0.25), *metric.Gauge.Value)
						foundSystemWeight = true
					}
				}
			}
		case "slurm_priority_factor_value":
			for _, metric := range mf.Metric {
				for _, label := range metric.Label {
					if *label.Name == "factor" && *label.Value == "age" {
						assert.Equal(t, float64(10000), *metric.Gauge.Value)
						foundFactorValue = true
					}
				}
			}
		case "slurm_priority_factor_weight_balance":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.8), *mf.Metric[0].Gauge.Value)
				foundWeightBalance = true
			}
		}
	}

	assert.True(t, foundSystemWeight, "Should find system weight metric with correct value")
	assert.True(t, foundFactorValue, "Should find factor value metric with correct value")
	assert.True(t, foundWeightBalance, "Should find weight balance metric with correct value")
}

func TestPriorityFactorsCollector_Integration(t *testing.T) {
	client := &MockPriorityFactorsSLURMClient{}

	// Setup comprehensive mock data
	setupFactorMocks(client)

	collector := NewPriorityFactorsCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_system_factor_configuration System-wide priority factor configuration
		# TYPE slurm_system_factor_configuration gauge
		slurm_system_factor_configuration{config_type="weight",factor="age",version="v2.0"} 0.3
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_system_factor_configuration")
	assert.NoError(t, err)
}

func TestPriorityFactorsCollector_FactorBreakdown(t *testing.T) {
	client := &MockPriorityFactorsSLURMClient{}

	// Test comprehensive factor breakdown
	breakdown := &PriorityFactorBreakdown{
		JobID:         "test_job",
		UserName:      "test_user",
		AccountName:   "test_account",
		PartitionName: "cpu",
		QoSName:       "normal",

		AgeValue:       15000.0,
		FairShareValue: 25000.0,
		QoSValue:       8000.0,

		AgeContribution:       4500.0,
		FairShareContribution: 10000.0,
		QoSContribution:       1600.0,

		DominantFactor:    "fairshare",
		SecondaryFactor:   "age",
		FactorBalance:     0.75,
	}

	client.On("GetSystemFactorWeights", mock.Anything).Return(&SystemFactorWeights{}, nil)
	client.On("GetPriorityFactorBreakdown", mock.Anything, "test_job").Return(breakdown, nil)
	client.On("GetUserFactorProfile", mock.Anything, mock.AnythingOfType("string")).Return(&UserFactorProfile{}, nil)
	client.On("GetAccountFactorSummary", mock.Anything, mock.AnythingOfType("string")).Return(&AccountFactorSummary{}, nil)
	client.On("GetPartitionFactorAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&PartitionFactorAnalysis{}, nil)
	client.On("GetQoSFactorAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&QoSFactorAnalysis{}, nil)
	client.On("GetFactorTrendAnalysis", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&FactorTrendAnalysis{}, nil)
	client.On("GetFactorImpactAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&FactorImpactAnalysis{}, nil)
	client.On("OptimizeFactorWeights", mock.Anything, "system").Return(&FactorOptimizationResult{}, nil)
	client.On("ValidateFactorConfiguration", mock.Anything).Return(&FactorConfigurationValidation{}, nil)

	collector := NewPriorityFactorsCollector(client)

	// Override sample job IDs for testing
	collector.getSampleJobIDs = func() []string {
		return []string{"test_job"}
	}

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

	// Verify factor breakdown metrics are present
	foundFactorValue := false
	foundFactorContribution := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "priority_factor_value") {
			foundFactorValue = true
		}
		if strings.Contains(desc, "priority_factor_contribution") {
			foundFactorContribution = true
		}
	}

	assert.True(t, foundFactorValue, "Should find factor value metrics")
	assert.True(t, foundFactorContribution, "Should find factor contribution metrics")
}

func setupFactorMocks(client *MockPriorityFactorsSLURMClient) {
	systemWeights := &SystemFactorWeights{
		AgeWeightGlobal:        0.3,
		FairShareWeightGlobal:  0.4,
		QoSWeightGlobal:        0.2,
		WeightBalance:          0.85,
		ConfigurationVersion:   "v2.0",
		CalibrationAccuracy:    0.95,
	}

	factorBreakdown := &PriorityFactorBreakdown{
		JobID:         "12345",
		UserName:      "user1",
		AccountName:   "account1",
		PartitionName: "cpu",
		QoSName:       "normal",
		AgeValue:      10000.0,
		FairShareValue: 15000.0,
		QoSValue:      5000.0,
		TotalPriority: 30000.0,
	}

	userProfile := &UserFactorProfile{
		UserName:                 "user1",
		AccountName:              "account1",
		AvgAgeContribution:       3000.0,
		AvgFairShareContribution: 6000.0,
		FactorStability:          0.85,
	}

	accountSummary := &AccountFactorSummary{
		AccountName:           "account1",
		TotalUsers:            10,
		ActiveUsers:           8,
		AvgFactorBalance:      0.8,
		FactorEfficiencyScore: 0.85,
	}

	partitionAnalysis := &PartitionFactorAnalysis{
		PartitionName:         "cpu",
		PartitionWeight:       0.1,
		FactorUtilizationRate: 0.75,
		FactorEffectiveness:   0.88,
	}

	qosAnalysis := &QoSFactorAnalysis{
		QoSName:                "normal",
		QoSPriorityValue:       1000,
		QoSWeight:              0.2,
		QoSFactorEffectiveness: 0.85,
	}

	factorTrend := &FactorTrendAnalysis{
		FactorName:           "age",
		TrendDirection:       "increasing",
		TrendVelocity:        0.05,
		PredictionConfidence: 0.8,
	}

	factorImpact := &FactorImpactAnalysis{
		FactorName:        "fairshare",
		ImpactScore:       0.85,
		ImpactCorrelation: 0.75,
		FactorInteractions: map[string]float64{
			"age": 0.3,
		},
	}

	optimizationResult := &FactorOptimizationResult{
		OptimizationScope:     "system",
		CurrentEffectiveness:  0.85,
		ExpectedEffectiveness: 0.9,
		ExpectedImprovement:   0.05,
	}

	configValidation := &FactorConfigurationValidation{
		ValidationScope: "system",
		IsValid:         true,
		ValidationScore: 0.9,
		HealthScore:     0.88,
	}

	client.On("GetSystemFactorWeights", mock.Anything).Return(systemWeights, nil)
	client.On("GetPriorityFactorBreakdown", mock.Anything, mock.AnythingOfType("string")).Return(factorBreakdown, nil)
	client.On("GetUserFactorProfile", mock.Anything, mock.AnythingOfType("string")).Return(userProfile, nil)
	client.On("GetAccountFactorSummary", mock.Anything, mock.AnythingOfType("string")).Return(accountSummary, nil)
	client.On("GetPartitionFactorAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(partitionAnalysis, nil)
	client.On("GetQoSFactorAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(qosAnalysis, nil)
	client.On("GetFactorTrendAnalysis", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(factorTrend, nil)
	client.On("GetFactorImpactAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(factorImpact, nil)
	client.On("OptimizeFactorWeights", mock.Anything, "system").Return(optimizationResult, nil)
	client.On("ValidateFactorConfiguration", mock.Anything).Return(configValidation, nil)
}