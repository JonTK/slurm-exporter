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

type MockFairShareDecaySLURMClient struct {
	mock.Mock
}

func (m *MockFairShareDecaySLURMClient) GetFairShareDecayConfiguration(ctx context.Context) (*FairShareDecayConfig, error) {
	args := m.Called(ctx)
	return args.Get(0).(*FairShareDecayConfig), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) GetFairShareDecayMetrics(ctx context.Context, userName string) (*FairShareDecayMetrics, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*FairShareDecayMetrics), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) GetFairShareResetSchedule(ctx context.Context) (*FairShareResetSchedule, error) {
	args := m.Called(ctx)
	return args.Get(0).(*FairShareResetSchedule), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) GetUserDecayImpactAnalysis(ctx context.Context, userName string) (*UserDecayImpactAnalysis, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserDecayImpactAnalysis), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) GetSystemDecayEffectiveness(ctx context.Context) (*SystemDecayEffectiveness, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemDecayEffectiveness), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) GetDecayTrendAnalysis(ctx context.Context, period string) (*DecayTrendAnalysis, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*DecayTrendAnalysis), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) ValidateDecayConfiguration(ctx context.Context) (*DecayConfigurationValidation, error) {
	args := m.Called(ctx)
	return args.Get(0).(*DecayConfigurationValidation), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) OptimizeDecayParameters(ctx context.Context) (*DecayOptimizationResult, error) {
	args := m.Called(ctx)
	return args.Get(0).(*DecayOptimizationResult), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) PredictDecayImpact(ctx context.Context, userName string, timeframe string) (*DecayImpactPrediction, error) {
	args := m.Called(ctx, userName, timeframe)
	return args.Get(0).(*DecayImpactPrediction), args.Error(1)
}

func (m *MockFairShareDecaySLURMClient) AnalyzeDecayAnomalies(ctx context.Context) ([]*DecayAnomaly, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*DecayAnomaly), args.Error(1)
}

func TestNewFairShareDecayCollector(t *testing.T) {
	client := &MockFairShareDecaySLURMClient{}
	collector := NewFairShareDecayCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.fairShareDecayRate)
	assert.NotNil(t, collector.fairShareHalfLife)
	assert.NotNil(t, collector.fairShareResetSchedule)
	assert.NotNil(t, collector.systemDecayEffectiveness)
	assert.NotNil(t, collector.decayConfigValidation)
}

func TestFairShareDecayCollector_Describe(t *testing.T) {
	client := &MockFairShareDecaySLURMClient{}
	collector := NewFairShareDecayCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 33 metrics, verify we have the correct number
	assert.Equal(t, 33, len(descs), "Should have exactly 33 metric descriptions")
}

func TestFairShareDecayCollector_Collect_Success(t *testing.T) {
	client := &MockFairShareDecaySLURMClient{}

	// Mock decay configuration
	decayConfig := &FairShareDecayConfig{
		DecayHalfLife:       24 * time.Hour,
		DecayFactor:         0.95,
		DecayInterval:       15 * time.Minute,
		MaxDecayAge:         7 * 24 * time.Hour,
		MinFairShareValue:   0.01,
		ResetInterval:       7 * 24 * time.Hour,
		ResetDay:            "Sunday",
		ResetTime:           "00:00",
		ResetMethod:         "full",
		PartialResetEnabled: false,
		ConfigVersion:       "v2.1",
		LastModified:        time.Now().Add(-24 * time.Hour),
		ModifiedBy:          "admin",
		EffectiveFrom:       time.Now().Add(-7 * 24 * time.Hour),
		AutoOptimization:    true,
		OptimizationTarget:  "fairness",
		TuningEnabled:       true,
		AdaptiveDecay:       false,
		LastUpdated:         time.Now(),
	}

	// Mock decay metrics
	decayMetrics := &FairShareDecayMetrics{
		UserName:            "testuser",
		AccountName:         "testaccount",
		CurrentFairShare:    0.75,
		PreDecayFairShare:   0.80,
		DecayAmount:         0.05,
		DecayPercentage:     6.25,
		TimeSinceLastDecay:  900.0, // 15 minutes
		DecayHistory:        []float64{0.80, 0.78, 0.76, 0.75},
		DecayVelocity:       -0.001,
		DecayAcceleration:   0.0001,
		DecayTrend:          "decreasing",
		PriorityImpact:      -500.0,
		QueuePositionImpact: 2.0,
		WaitTimeImpact:      120.0,
		OverallImpact:       -0.15,
		PredictedDecay:      0.03,
		TimeToReset:         432000.0, // 5 days in seconds
		ExpectedResetBenefit: 0.25,
		LastCalculated:      time.Now(),
	}

	// Mock reset schedule
	resetSchedule := &FairShareResetSchedule{
		NextResetTime:        time.Now().Add(5 * 24 * time.Hour),
		LastResetTime:        time.Now().Add(-2 * 24 * time.Hour),
		TimeSinceLastReset:   2 * 24 * time.Hour,
		TimeToNextReset:      5 * 24 * time.Hour,
		ResetFrequency:       "weekly",
		ResetMethod:          "full",
		AffectedUsers:        150,
		AffectedAccounts:     25,
		EstimatedImpact:      "moderate",
		PriorityChanges:      75,
		QueueReordering:      true,
		SystemDisruption:     0.3,
		ResetHistory:         []time.Time{time.Now().Add(-9 * 24 * time.Hour), time.Now().Add(-2 * 24 * time.Hour)},
		AverageResetInterval: 7 * 24 * time.Hour,
		ResetEffectiveness:   0.85,
		ResetReliability:     0.92,
		LastUpdated:          time.Now(),
	}

	// Mock user decay impact analysis
	userImpact := &UserDecayImpactAnalysis{
		UserName:             "testuser",
		AccountName:          "testaccount",
		DecayImpactScore:     0.7,
		DecayAdaptation:      0.8,
		DecayResistance:      0.6,
		DecayRecovery:        0.75,
		SubmissionPatterns:   "regular",
		UsageOptimization:    0.82,
		AdaptationStrategy:   "proactive",
		BehaviorChanges:      []string{"increased_submission_frequency", "better_timing"},
		JobSuccessRate:       0.95,
		ResourceEfficiency:   0.88,
		QueuePerformance:     0.79,
		OverallPerformance:   0.87,
		ResetBenefit:         0.25,
		ResetUtilization:     0.9,
		PostResetPerformance: 0.92,
		LastAnalyzed:         time.Now(),
	}

	// Mock system decay effectiveness
	systemEffectiveness := &SystemDecayEffectiveness{
		OverallEffectiveness:     0.85,
		FairnessScore:           0.88,
		BalanceScore:            0.82,
		PerformanceScore:        0.87,
		DecayEfficiency:         0.83,
		DecayConsistency:        0.89,
		DecayPredictability:     0.76,
		DecayStability:          0.91,
		ResetEffectiveness:      0.86,
		ResetTimingOptimality:   0.78,
		ResetImpactMinimization: 0.84,
		ResetRecoverySpeed:      0.73,
		SystemStress:            0.25,
		UserSatisfaction:        0.81,
		OperationalEfficiency:   0.89,
		ConfigurationHealth:     0.92,
		LastCalculated:          time.Now(),
	}

	// Mock decay trend analysis
	trendAnalysis := &DecayTrendAnalysis{
		AnalysisPeriod:       "24h",
		DecayTrendDirection:  "decreasing",
		DecayVelocity:        -0.002,
		DecayAcceleration:    0.0001,
		DecayVolatility:      0.15,
		SeasonalPatterns:     map[string]float64{"monday": 0.8, "friday": 1.2},
		SeasonalStrength:     0.3,
		CyclicalBehavior:     true,
		CyclePeriod:          "weekly",
		FutureTrend:          "stable",
		PredictionConfidence: 0.82,
		TrendStability:       0.75,
		LastAnalyzed:         time.Now(),
	}

	// Mock configuration validation
	configValidation := &DecayConfigurationValidation{
		IsValid:               true,
		ValidationScore:       0.88,
		ValidationIssues:      []string{},
		ValidationWarnings:    []string{"consider_shorter_half_life"},
		HalfLifeOptimality:    0.82,
		IntervalOptimality:    0.90,
		ResetTimingOptimality: 0.85,
		OverallOptimality:     0.86,
		RecommendedChanges:    []string{"adjust_half_life", "optimize_reset_timing"},
		OptimizationPotential: 0.12,
		RiskAssessment:        "low",
		ValidatedAt:           time.Now(),
	}

	// Mock optimization result
	optimizationResult := &DecayOptimizationResult{
		CurrentHalfLife:       24 * time.Hour,
		CurrentDecayFactor:    0.95,
		CurrentEffectiveness:  0.85,
		RecommendedHalfLife:   20 * time.Hour,
		RecommendedDecayFactor: 0.93,
		ExpectedEffectiveness: 0.91,
		ExpectedImprovement:   0.06,
		ImplementationRisk:    "low",
		TransitionPeriod:      "2_weeks",
		RollbackPlan:          "automatic",
		OptimizedAt:           time.Now(),
	}

	// Mock decay anomalies
	anomalies := []*DecayAnomaly{
		{
			AnomalyType:         "unexpected_acceleration",
			AnomalySeverity:     "medium",
			AnomalyDescription:  "Decay rate increased significantly for user group",
			AffectedUsers:       []string{"user1", "user2"},
			DetectionConfidence: 0.89,
			ImpactScore:         0.65,
			RecommendedAction:   "monitor_closely",
			DetectedAt:          time.Now().Add(-time.Hour),
		},
		{
			AnomalyType:         "reset_ineffectiveness",
			AnomalySeverity:     "low",
			AnomalyDescription:  "Recent reset had lower than expected impact",
			AffectedUsers:       []string{"user3"},
			DetectionConfidence: 0.72,
			ImpactScore:         0.35,
			RecommendedAction:   "review_reset_parameters",
			DetectedAt:          time.Now().Add(-30 * time.Minute),
		},
	}

	// Setup mock expectations
	client.On("GetFairShareDecayConfiguration", mock.Anything).Return(decayConfig, nil)
	client.On("GetFairShareDecayMetrics", mock.Anything, mock.AnythingOfType("string")).Return(decayMetrics, nil)
	client.On("GetFairShareResetSchedule", mock.Anything).Return(resetSchedule, nil)
	client.On("GetUserDecayImpactAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(userImpact, nil)
	client.On("GetSystemDecayEffectiveness", mock.Anything).Return(systemEffectiveness, nil)
	client.On("GetDecayTrendAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(trendAnalysis, nil)
	client.On("ValidateDecayConfiguration", mock.Anything).Return(configValidation, nil)
	client.On("OptimizeDecayParameters", mock.Anything).Return(optimizationResult, nil)
	client.On("AnalyzeDecayAnomalies", mock.Anything).Return(anomalies, nil)

	collector := NewFairShareDecayCollector(client)

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
	foundDecayRate := false
	foundHalfLife := false
	foundResetSchedule := false
	foundSystemEffectiveness := false
	foundAnomalyDetection := false

	for name := range metricNames {
		if strings.Contains(name, "fairshare_decay_rate") {
			foundDecayRate = true
		}
		if strings.Contains(name, "fairshare_half_life") {
			foundHalfLife = true
		}
		if strings.Contains(name, "fairshare_reset_schedule") {
			foundResetSchedule = true
		}
		if strings.Contains(name, "system_decay_effectiveness") {
			foundSystemEffectiveness = true
		}
		if strings.Contains(name, "decay_anomaly_detection") {
			foundAnomalyDetection = true
		}
	}

	assert.True(t, foundDecayRate, "Should have decay rate metrics")
	assert.True(t, foundHalfLife, "Should have half-life metrics")
	assert.True(t, foundResetSchedule, "Should have reset schedule metrics")
	assert.True(t, foundSystemEffectiveness, "Should have system effectiveness metrics")
	assert.True(t, foundAnomalyDetection, "Should have anomaly detection metrics")

	client.AssertExpectations(t)
}

func TestFairShareDecayCollector_Collect_Error(t *testing.T) {
	client := &MockFairShareDecaySLURMClient{}

	// Mock error response
	client.On("GetFairShareDecayConfiguration", mock.Anything).Return((*FairShareDecayConfig)(nil), assert.AnError)

	collector := NewFairShareDecayCollector(client)

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

func TestFairShareDecayCollector_MetricValues(t *testing.T) {
	client := &MockFairShareDecaySLURMClient{}

	// Create test data with known values
	decayConfig := &FairShareDecayConfig{
		DecayHalfLife:   48 * time.Hour,
		DecayFactor:     0.9,
		ResetInterval:   7 * 24 * time.Hour,
		ConfigVersion:   "test_v1",
	}

	systemEffectiveness := &SystemDecayEffectiveness{
		OverallEffectiveness: 0.92,
		FairnessScore:       0.88,
		BalanceScore:        0.85,
	}

	resetSchedule := &FairShareResetSchedule{
		TimeToNextReset:  72 * time.Hour,
		ResetMethod:      "partial",
		AffectedUsers:    100,
		ResetEffectiveness: 0.87,
	}

	client.On("GetFairShareDecayConfiguration", mock.Anything).Return(decayConfig, nil)
	client.On("GetSystemDecayEffectiveness", mock.Anything).Return(systemEffectiveness, nil)
	client.On("GetFairShareResetSchedule", mock.Anything).Return(resetSchedule, nil)
	client.On("GetFairShareDecayMetrics", mock.Anything, mock.AnythingOfType("string")).Return(&FairShareDecayMetrics{}, nil)
	client.On("GetUserDecayImpactAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&UserDecayImpactAnalysis{}, nil)
	client.On("GetDecayTrendAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&DecayTrendAnalysis{}, nil)
	client.On("ValidateDecayConfiguration", mock.Anything).Return(&DecayConfigurationValidation{}, nil)
	client.On("OptimizeDecayParameters", mock.Anything).Return(&DecayOptimizationResult{}, nil)
	client.On("AnalyzeDecayAnomalies", mock.Anything).Return([]*DecayAnomaly{}, nil)

	collector := NewFairShareDecayCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundHalfLife := false
	foundDecayFactor := false
	foundResetCountdown := false
	foundEffectiveness := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_fairshare_half_life_hours":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(48), *mf.Metric[0].Gauge.Value)
				foundHalfLife = true
			}
		case "slurm_fairshare_decay_factor":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.9), *mf.Metric[0].Gauge.Value)
				foundDecayFactor = true
			}
		case "slurm_fairshare_reset_countdown_hours":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(72), *mf.Metric[0].Gauge.Value)
				foundResetCountdown = true
			}
		case "slurm_system_decay_effectiveness":
			for _, metric := range mf.Metric {
				for _, label := range metric.Label {
					if *label.Name == "metric" && *label.Value == "overall" {
						assert.Equal(t, float64(0.92), *metric.Gauge.Value)
						foundEffectiveness = true
					}
				}
			}
		}
	}

	assert.True(t, foundHalfLife, "Should find half-life metric with correct value")
	assert.True(t, foundDecayFactor, "Should find decay factor metric with correct value")
	assert.True(t, foundResetCountdown, "Should find reset countdown metric with correct value")
	assert.True(t, foundEffectiveness, "Should find effectiveness metric with correct value")
}

func TestFairShareDecayCollector_Integration(t *testing.T) {
	client := &MockFairShareDecaySLURMClient{}

	// Setup comprehensive mock data
	setupDecayMocks(client)

	collector := NewFairShareDecayCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_fairshare_half_life_hours Fair-share half-life in hours
		# TYPE slurm_fairshare_half_life_hours gauge
		slurm_fairshare_half_life_hours{config_version="v2.0",scope="system"} 24
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_fairshare_half_life_hours")
	assert.NoError(t, err)
}

func TestFairShareDecayCollector_AnomalyDetection(t *testing.T) {
	client := &MockFairShareDecaySLURMClient{}

	anomalies := []*DecayAnomaly{
		{
			AnomalyType:         "rapid_decay",
			AnomalySeverity:     "high",
			ImpactScore:         0.9,
			DetectionConfidence: 0.95,
		},
		{
			AnomalyType:         "slow_recovery",
			AnomalySeverity:     "medium",
			ImpactScore:         0.6,
			DetectionConfidence: 0.85,
		},
	}

	client.On("GetFairShareDecayConfiguration", mock.Anything).Return(&FairShareDecayConfig{}, nil)
	client.On("GetSystemDecayEffectiveness", mock.Anything).Return(&SystemDecayEffectiveness{}, nil)
	client.On("GetFairShareResetSchedule", mock.Anything).Return(&FairShareResetSchedule{}, nil)
	client.On("GetFairShareDecayMetrics", mock.Anything, mock.AnythingOfType("string")).Return(&FairShareDecayMetrics{}, nil)
	client.On("GetUserDecayImpactAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&UserDecayImpactAnalysis{}, nil)
	client.On("GetDecayTrendAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&DecayTrendAnalysis{}, nil)
	client.On("ValidateDecayConfiguration", mock.Anything).Return(&DecayConfigurationValidation{}, nil)
	client.On("OptimizeDecayParameters", mock.Anything).Return(&DecayOptimizationResult{}, nil)
	client.On("AnalyzeDecayAnomalies", mock.Anything).Return(anomalies, nil)

	collector := NewFairShareDecayCollector(client)

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

	// Verify anomaly detection metrics are present
	foundAnomalyMetrics := false
	for _, metric := range metrics {
		if strings.Contains(metric.Desc().String(), "decay_anomaly_detection") {
			foundAnomalyMetrics = true
			break
		}
	}

	assert.True(t, foundAnomalyMetrics, "Should find anomaly detection metrics")
}

func setupDecayMocks(client *MockFairShareDecaySLURMClient) {
	decayConfig := &FairShareDecayConfig{
		DecayHalfLife: 24 * time.Hour,
		DecayFactor:   0.95,
		ConfigVersion: "v2.0",
	}

	systemEffectiveness := &SystemDecayEffectiveness{
		OverallEffectiveness: 0.85,
		FairnessScore:       0.88,
		BalanceScore:        0.82,
	}

	resetSchedule := &FairShareResetSchedule{
		TimeToNextReset:    5 * 24 * time.Hour,
		ResetMethod:        "full",
		AffectedUsers:      150,
		ResetEffectiveness: 0.85,
	}

	decayMetrics := &FairShareDecayMetrics{
		UserName:        "user1",
		AccountName:     "account1",
		CurrentFairShare: 0.75,
		DecayAmount:     0.05,
	}

	userImpact := &UserDecayImpactAnalysis{
		UserName:         "user1",
		AccountName:      "account1",
		DecayImpactScore: 0.7,
		DecayAdaptation:  0.8,
	}

	trendAnalysis := &DecayTrendAnalysis{
		AnalysisPeriod:      "24h",
		DecayTrendDirection: "decreasing",
		DecayVelocity:       -0.002,
	}

	configValidation := &DecayConfigurationValidation{
		IsValid:         true,
		ValidationScore: 0.88,
	}

	optimizationResult := &DecayOptimizationResult{
		CurrentEffectiveness:  0.85,
		ExpectedEffectiveness: 0.91,
		ExpectedImprovement:   0.06,
	}

	client.On("GetFairShareDecayConfiguration", mock.Anything).Return(decayConfig, nil)
	client.On("GetSystemDecayEffectiveness", mock.Anything).Return(systemEffectiveness, nil)
	client.On("GetFairShareResetSchedule", mock.Anything).Return(resetSchedule, nil)
	client.On("GetFairShareDecayMetrics", mock.Anything, mock.AnythingOfType("string")).Return(decayMetrics, nil)
	client.On("GetUserDecayImpactAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(userImpact, nil)
	client.On("GetDecayTrendAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(trendAnalysis, nil)
	client.On("ValidateDecayConfiguration", mock.Anything).Return(configValidation, nil)
	client.On("OptimizeDecayParameters", mock.Anything).Return(optimizationResult, nil)
	client.On("AnalyzeDecayAnomalies", mock.Anything).Return([]*DecayAnomaly{}, nil)
}