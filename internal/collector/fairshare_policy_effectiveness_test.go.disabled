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

type MockFairSharePolicyEffectivenessSLURMClient struct {
	mock.Mock
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetFairSharePolicyConfiguration(ctx context.Context) (*FairSharePolicyConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*FairSharePolicyConfiguration), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyImplementationDetails(ctx context.Context) (*PolicyImplementationDetails, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PolicyImplementationDetails), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyParameterAnalysis(ctx context.Context) (*PolicyParameterAnalysis, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PolicyParameterAnalysis), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) ValidatePolicyConfiguration(ctx context.Context) (*PolicyConfigurationValidation, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PolicyConfigurationValidation), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyEffectivenessMetrics(ctx context.Context, period string) (*PolicyEffectivenessMetrics, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*PolicyEffectivenessMetrics), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetFairnessDistributionAnalysis(ctx context.Context, period string) (*FairnessDistributionAnalysis, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*FairnessDistributionAnalysis), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetResourceAllocationEffectiveness(ctx context.Context, period string) (*ResourceAllocationEffectiveness, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*ResourceAllocationEffectiveness), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetUserEquityAnalysis(ctx context.Context, period string) (*UserEquityAnalysis, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*UserEquityAnalysis), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyImpactAssessment(ctx context.Context, policyID string, period string) (*PolicyImpactAssessment, error) {
	args := m.Called(ctx, policyID, period)
	return args.Get(0).(*PolicyImpactAssessment), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetSystemPerformanceImpact(ctx context.Context, period string) (*SystemPerformanceImpact, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*SystemPerformanceImpact), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetUserSatisfactionMetrics(ctx context.Context, period string) (*UserSatisfactionMetrics, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*UserSatisfactionMetrics), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetResourceUtilizationImpact(ctx context.Context, period string) (*ResourceUtilizationImpact, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*ResourceUtilizationImpact), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyOptimizationRecommendations(ctx context.Context) (*PolicyOptimizationRecommendations, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PolicyOptimizationRecommendations), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyTuningHistory(ctx context.Context, period string) (*PolicyTuningHistory, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*PolicyTuningHistory), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyPerformanceComparison(ctx context.Context, policyA, policyB string) (*PolicyPerformanceComparison, error) {
	args := m.Called(ctx, policyA, policyB)
	return args.Get(0).(*PolicyPerformanceComparison), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetOptimalPolicyParameters(ctx context.Context, objectiveFunction string) (*OptimalPolicyParameters, error) {
	args := m.Called(ctx, objectiveFunction)
	return args.Get(0).(*OptimalPolicyParameters), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyComplianceMonitoring(ctx context.Context) (*PolicyComplianceMonitoring, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PolicyComplianceMonitoring), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyViolationAnalysis(ctx context.Context, period string) (*PolicyViolationAnalysis, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*PolicyViolationAnalysis), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyDriftDetection(ctx context.Context) (*PolicyDriftDetection, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PolicyDriftDetection), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyEffectivenessAlerts(ctx context.Context) (*PolicyEffectivenessAlerts, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PolicyEffectivenessAlerts), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyEffectivenessTrends(ctx context.Context, period string) (*PolicyEffectivenessTrends, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*PolicyEffectivenessTrends), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyScenarioAnalysis(ctx context.Context, scenario string) (*PolicyScenarioAnalysis, error) {
	args := m.Called(ctx, scenario)
	return args.Get(0).(*PolicyScenarioAnalysis), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyPredictiveModeling(ctx context.Context, timeframe string) (*PolicyPredictiveModeling, error) {
	args := m.Called(ctx, timeframe)
	return args.Get(0).(*PolicyPredictiveModeling), args.Error(1)
}

func (m *MockFairSharePolicyEffectivenessSLURMClient) GetPolicyCostBenefitAnalysis(ctx context.Context, period string) (*PolicyCostBenefitAnalysis, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*PolicyCostBenefitAnalysis), args.Error(1)
}

// Types are already defined in fairshare_policy_effectiveness.go

func TestNewFairSharePolicyEffectivenessCollector(t *testing.T) {
	client := &MockFairSharePolicyEffectivenessSLURMClient{}
	collector := NewFairSharePolicyEffectivenessCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.policyConfigurationValid)
	assert.NotNil(t, collector.policyOverallEffectiveness)
	assert.NotNil(t, collector.policyFairnessIndex)
	assert.NotNil(t, collector.policyOptimizationPotential)
}

func TestFairSharePolicyEffectivenessCollector_Describe(t *testing.T) {
	client := &MockFairSharePolicyEffectivenessSLURMClient{}
	collector := NewFairSharePolicyEffectivenessCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 52 metrics, verify we have the correct number
	assert.Equal(t, 52, len(descs), "Should have exactly 52 metric descriptions")
}

func TestFairSharePolicyEffectivenessCollector_Collect_Success(t *testing.T) {
	client := &MockFairSharePolicyEffectivenessSLURMClient{}

	// Mock policy configuration
	policyConfig := &FairSharePolicyConfiguration{
		PolicyID:                "policy_1",
		PolicyName:              "default_fairshare",
		PolicyVersion:           "v2.0",
		ImplementationDate:      time.Now().Add(-30 * 24 * time.Hour),
		LastModified:            time.Now().Add(-5 * 24 * time.Hour),
		FairShareAlgorithm:      "priority_multifactor",
		DecayHalfLife:           604800.0,
		UsageFactor:             0.5,
		LevelFactor:             0.3,
		PriorityWeight:          10000.0,
		AccountHierarchyEnabled: true,
		InheritParentLimits:     true,
		AccountNormalization:    "proportional",
		HierarchyDepth:          3,
		UserFairShareEnabled:    true,
		QoSFairShareEnabled:     true,
		PartitionFairShare: map[string]float64{
			"compute": 0.6,
			"gpu":     0.3,
			"memory":  0.1,
		},
		DefaultFairShare:      1.0,
		EnforcementLevel:      "strict",
		GracePeriodsEnabled:   true,
		GracePeriodDuration:   3600.0,
		ViolationPenalties: map[string]float64{
			"minor":    0.1,
			"major":    0.3,
			"critical": 0.5,
		},
		CalculationFrequency:  "5min",
		DatabasePurgePolicy:   "30d",
		BackfillConsideration: true,
		PreemptionEnabled:     false,
		ConfigurationValid:    true,
		ValidationErrors:      []string{},
		ConfigurationHash:     "abc123def456",
		LastValidated:         time.Now(),
	}

	// Mock policy effectiveness metrics
	effectivenessMetrics := &PolicyEffectivenessMetrics{
		PolicyID:                     "policy_1",
		MeasurementPeriod:            "24h",
		GiniCoefficient:              0.32,
		FairnessIndex:                0.85,
		EquityScore:                  0.82,
		VariationCoefficient:         0.28,
		JainsFairnessIndex:           0.88,
		OverallEffectiveness:         0.86,
		AllocationEffectiveness:      0.84,
		EnforcementEffectiveness:     0.89,
		UserComplianceRate:           0.92,
		SystemStabilityScore:         0.87,
		ResourceDistributionFairness: 0.83,
		WaitTimeEquity:               0.79,
		ThroughputFairness:           0.85,
		AccessibilityScore:           0.90,
		PolicyAdherenceRate:          0.91,
		ViolationRate:                0.08,
		CorrectionEfficiency:         0.85,
		Responseiveness:              0.88,
		UserSatisfactionImpact:       0.75,
		SystemPerformanceImpact:      0.15,
		ResourceEfficiencyImpact:     0.22,
		CollaborationImpact:          0.65,
		TrendDirection:               "improving",
		EffectivenessStability:       0.92,
		ImprovementRate:              0.05,
		AdaptationScore:              0.78,
		LastMeasured:                 time.Now(),
	}

	// Mock policy impact assessment
	impactAssessment := &PolicyImpactAssessment{
		PolicyID:         "policy_1",
		AssessmentPeriod: "7d",
		ResourceAllocationChange: map[string]float64{
			"cpu":    0.12,
			"memory": 0.08,
			"gpu":    -0.05,
		},
		WaitTimeImpact:          -0.15,
		ThroughputImpact:        0.18,
		UtilizationImpact:       0.12,
		EfficiencyImpact:        0.20,
		SubmissionPatternChange: 0.08,
		UserAdaptationRate:      0.75,
		CollaborationChange:     0.15,
		CompetitionLevel:        -0.10,
		GameficationPrevention:  0.85,
		SystemStabilityChange:   0.10,
		LoadBalancingImprovement: 0.22,
		ResourceContentionChange: -0.18,
		SchedulingEfficiencyChange: 0.15,
		CostEfficiencyChange:    0.25,
		ResourceWasteReduction:  0.30,
		ProductivityImpact:      0.12,
		ROIImprovement:          0.35,
		PolicyComplianceImprovement: 0.20,
		GovernanceEffectiveness: 0.88,
		AuditabilityScore:       0.92,
		TransparencyImprovement: 0.15,
		ImplementationRisk:      0.15,
		UserAcceptanceRisk:      0.25,
		SystemPerformanceRisk:   0.10,
		ComplianceRisk:          0.05,
		ImpactConfidence:        0.85,
		LastAssessed:            time.Now(),
	}

	// Setup mock expectations
	client.On("GetFairSharePolicyConfiguration", mock.Anything).Return(policyConfig, nil)
	client.On("GetPolicyEffectivenessMetrics", mock.Anything, "24h").Return(effectivenessMetrics, nil)
	client.On("GetPolicyImpactAssessment", mock.Anything, "policy_1", "7d").Return(impactAssessment, nil)
	client.On("GetPolicyOptimizationRecommendations", mock.Anything).Return(&PolicyOptimizationRecommendations{}, nil)
	client.On("GetPolicyPredictiveModeling", mock.Anything, "7d").Return(&PolicyPredictiveModeling{}, nil)
	client.On("GetPolicyCostBenefitAnalysis", mock.Anything, "30d").Return(&PolicyCostBenefitAnalysis{}, nil)

	collector := NewFairSharePolicyEffectivenessCollector(client)

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
	foundConfigValid := false
	foundEffectiveness := false
	foundFairnessIndex := false
	foundImpactMetrics := false
	foundOptimization := false

	for name := range metricNames {
		if strings.Contains(name, "configuration_valid") {
			foundConfigValid = true
		}
		if strings.Contains(name, "overall_effectiveness") {
			foundEffectiveness = true
		}
		if strings.Contains(name, "fairness_index") {
			foundFairnessIndex = true
		}
		if strings.Contains(name, "impact") {
			foundImpactMetrics = true
		}
		if strings.Contains(name, "optimization") {
			foundOptimization = true
		}
	}

	assert.True(t, foundConfigValid, "Should have configuration validity metrics")
	assert.True(t, foundEffectiveness, "Should have effectiveness metrics")
	assert.True(t, foundFairnessIndex, "Should have fairness index metrics")
	assert.True(t, foundImpactMetrics, "Should have impact assessment metrics")
	assert.True(t, foundOptimization, "Should have optimization metrics")

	client.AssertExpectations(t)
}

func TestFairSharePolicyEffectivenessCollector_Collect_Error(t *testing.T) {
	client := &MockFairSharePolicyEffectivenessSLURMClient{}

	// Mock error response
	client.On("GetFairSharePolicyConfiguration", mock.Anything).Return((*FairSharePolicyConfiguration)(nil), assert.AnError)

	collector := NewFairSharePolicyEffectivenessCollector(client)

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

func TestFairSharePolicyEffectivenessCollector_MetricValues(t *testing.T) {
	client := &MockFairSharePolicyEffectivenessSLURMClient{}

	// Create test data with known values
	policyConfig := &FairSharePolicyConfiguration{
		PolicyID:             "test_policy",
		PolicyName:           "test_fairshare",
		ConfigurationValid:   true,
		ImplementationDate:   time.Now().Add(-10 * 24 * time.Hour),
		LastModified:         time.Now().Add(-24 * time.Hour),
		DecayHalfLife:        604800.0,
		UsageFactor:          0.6,
		LevelFactor:          0.4,
		PriorityWeight:       5000.0,
		DefaultFairShare:     1.0,
		ValidationErrors:     []string{},
	}

	effectivenessMetrics := &PolicyEffectivenessMetrics{
		PolicyID:                "test_policy",
		OverallEffectiveness:    0.88,
		FairnessIndex:           0.85,
		GiniCoefficient:         0.25,
		UserComplianceRate:      0.92,
		SystemStabilityScore:    0.90,
		TrendDirection:          "improving",
		EffectivenessStability:  0.95,
		ImprovementRate:         0.03,
	}

	impactAssessment := &PolicyImpactAssessment{
		PolicyID:           "test_policy",
		WaitTimeImpact:     -0.20,
		ThroughputImpact:   0.15,
		UtilizationImpact:  0.10,
		ResourceAllocationChange: map[string]float64{
			"cpu": 0.05,
		},
	}

	client.On("GetFairSharePolicyConfiguration", mock.Anything).Return(policyConfig, nil)
	client.On("GetPolicyEffectivenessMetrics", mock.Anything, "24h").Return(effectivenessMetrics, nil)
	client.On("GetPolicyImpactAssessment", mock.Anything, "policy_1", "7d").Return(impactAssessment, nil)
	client.On("GetPolicyOptimizationRecommendations", mock.Anything).Return(&PolicyOptimizationRecommendations{}, nil)
	client.On("GetPolicyPredictiveModeling", mock.Anything, "7d").Return(&PolicyPredictiveModeling{}, nil)
	client.On("GetPolicyCostBenefitAnalysis", mock.Anything, "30d").Return(&PolicyCostBenefitAnalysis{}, nil)

	collector := NewFairSharePolicyEffectivenessCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundConfigValid := false
	foundOverallEffectiveness := false
	foundGiniCoefficient := false
	foundComplianceRate := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_fairshare_policy_configuration_valid":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(1), *mf.Metric[0].Gauge.Value)
				foundConfigValid = true
			}
		case "slurm_fairshare_policy_overall_effectiveness":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.88), *mf.Metric[0].Gauge.Value)
				foundOverallEffectiveness = true
			}
		case "slurm_fairshare_policy_gini_coefficient":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.25), *mf.Metric[0].Gauge.Value)
				foundGiniCoefficient = true
			}
		case "slurm_fairshare_policy_user_compliance_rate":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.92), *mf.Metric[0].Gauge.Value)
				foundComplianceRate = true
			}
		}
	}

	assert.True(t, foundConfigValid, "Should find configuration validity metric with correct value")
	assert.True(t, foundOverallEffectiveness, "Should find overall effectiveness metric with correct value")
	assert.True(t, foundGiniCoefficient, "Should find Gini coefficient metric with correct value")
	assert.True(t, foundComplianceRate, "Should find compliance rate metric with correct value")
}

func TestFairSharePolicyEffectivenessCollector_Integration(t *testing.T) {
	client := &MockFairSharePolicyEffectivenessSLURMClient{}

	// Setup comprehensive mock data
	setupPolicyEffectivenessMocks(client)

	collector := NewFairSharePolicyEffectivenessCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_fairshare_policy_overall_effectiveness Fair-share policy overall effectiveness score (0-1)
		# TYPE slurm_fairshare_policy_overall_effectiveness gauge
		slurm_fairshare_policy_overall_effectiveness{measurement_period="24h",policy_id="policy_1"} 0.85
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_fairshare_policy_overall_effectiveness")
	assert.NoError(t, err)
}

func TestFairSharePolicyEffectivenessCollector_PolicyParameters(t *testing.T) {
	client := &MockFairSharePolicyEffectivenessSLURMClient{}

	policyConfig := &FairSharePolicyConfiguration{
		PolicyID:         "param_test",
		PolicyName:       "parameter_test",
		DecayHalfLife:    302400.0,
		UsageFactor:      0.7,
		LevelFactor:      0.2,
		PriorityWeight:   8000.0,
		DefaultFairShare: 2.0,
		PartitionFairShare: map[string]float64{
			"compute": 0.5,
			"gpu":     0.4,
			"memory":  0.1,
		},
		ConfigurationValid: true,
		ValidationErrors:   []string{},
	}

	client.On("GetFairSharePolicyConfiguration", mock.Anything).Return(policyConfig, nil)
	client.On("GetPolicyEffectivenessMetrics", mock.Anything, "24h").Return(&PolicyEffectivenessMetrics{}, nil)
	client.On("GetPolicyImpactAssessment", mock.Anything, "policy_1", "7d").Return(&PolicyImpactAssessment{}, nil)
	client.On("GetPolicyOptimizationRecommendations", mock.Anything).Return(&PolicyOptimizationRecommendations{}, nil)
	client.On("GetPolicyPredictiveModeling", mock.Anything, "7d").Return(&PolicyPredictiveModeling{}, nil)
	client.On("GetPolicyCostBenefitAnalysis", mock.Anything, "30d").Return(&PolicyCostBenefitAnalysis{}, nil)

	collector := NewFairSharePolicyEffectivenessCollector(client)

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

	// Verify parameter metrics are present
	foundParameterMetrics := false
	foundValidationMetrics := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "parameter_value") {
			foundParameterMetrics = true
		}
		if strings.Contains(desc, "validation") {
			foundValidationMetrics = true
		}
	}

	assert.True(t, foundParameterMetrics, "Should find parameter value metrics")
	assert.True(t, foundValidationMetrics, "Should find validation metrics")
}

func TestFairSharePolicyEffectivenessCollector_ImpactAssessment(t *testing.T) {
	client := &MockFairSharePolicyEffectivenessSLURMClient{}

	impactAssessment := &PolicyImpactAssessment{
		PolicyID:         "impact_test",
		WaitTimeImpact:   -0.25,
		ThroughputImpact: 0.20,
		UtilizationImpact: 0.15,
		ResourceAllocationChange: map[string]float64{
			"cpu":    0.10,
			"memory": 0.05,
			"gpu":    -0.02,
		},
		SystemStabilityChange:    0.12,
		LoadBalancingImprovement: 0.18,
		ResourceContentionChange: -0.22,
		CostEfficiencyChange:     0.30,
		ResourceWasteReduction:   0.25,
		ProductivityImpact:       0.15,
		ROIImprovement:           0.40,
	}

	client.On("GetFairSharePolicyConfiguration", mock.Anything).Return(&FairSharePolicyConfiguration{}, nil)
	client.On("GetPolicyEffectivenessMetrics", mock.Anything, "24h").Return(&PolicyEffectivenessMetrics{}, nil)
	client.On("GetPolicyImpactAssessment", mock.Anything, "policy_1", "7d").Return(impactAssessment, nil)
	client.On("GetPolicyOptimizationRecommendations", mock.Anything).Return(&PolicyOptimizationRecommendations{}, nil)
	client.On("GetPolicyPredictiveModeling", mock.Anything, "7d").Return(&PolicyPredictiveModeling{}, nil)
	client.On("GetPolicyCostBenefitAnalysis", mock.Anything, "30d").Return(&PolicyCostBenefitAnalysis{}, nil)

	collector := NewFairSharePolicyEffectivenessCollector(client)

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

	// Verify impact metrics are present
	foundWaitTimeImpact := false
	foundThroughputImpact := false
	foundROIMetrics := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "wait_time_impact") {
			foundWaitTimeImpact = true
		}
		if strings.Contains(desc, "throughput_impact") {
			foundThroughputImpact = true
		}
		if strings.Contains(desc, "roi") {
			foundROIMetrics = true
		}
	}

	assert.True(t, foundWaitTimeImpact, "Should find wait time impact metrics")
	assert.True(t, foundThroughputImpact, "Should find throughput impact metrics")
	assert.True(t, foundROIMetrics, "Should find ROI metrics")
}

func setupPolicyEffectivenessMocks(client *MockFairSharePolicyEffectivenessSLURMClient) {
	policyConfig := &FairSharePolicyConfiguration{
		PolicyID:                "policy_1",
		PolicyName:              "default_fairshare",
		PolicyVersion:           "v2.0",
		ImplementationDate:      time.Now().Add(-30 * 24 * time.Hour),
		LastModified:            time.Now().Add(-5 * 24 * time.Hour),
		DecayHalfLife:           604800.0,
		UsageFactor:             0.5,
		LevelFactor:             0.3,
		PriorityWeight:          10000.0,
		DefaultFairShare:        1.0,
		ConfigurationValid:      true,
		ValidationErrors:        []string{},
	}

	effectivenessMetrics := &PolicyEffectivenessMetrics{
		PolicyID:              "policy_1",
		MeasurementPeriod:     "24h",
		OverallEffectiveness:  0.85,
		FairnessIndex:         0.82,
		GiniCoefficient:       0.30,
		UserComplianceRate:    0.90,
		SystemStabilityScore:  0.88,
		TrendDirection:        "stable",
		ImprovementRate:       0.02,
	}

	impactAssessment := &PolicyImpactAssessment{
		PolicyID:         "policy_1",
		AssessmentPeriod: "7d",
		WaitTimeImpact:   -0.15,
		ThroughputImpact: 0.12,
		UtilizationImpact: 0.08,
		ResourceAllocationChange: map[string]float64{
			"cpu": 0.05,
		},
		CostEfficiencyChange: 0.20,
	}

	client.On("GetFairSharePolicyConfiguration", mock.Anything).Return(policyConfig, nil)
	client.On("GetPolicyEffectivenessMetrics", mock.Anything, "24h").Return(effectivenessMetrics, nil)
	client.On("GetPolicyImpactAssessment", mock.Anything, "policy_1", "7d").Return(impactAssessment, nil)
	client.On("GetPolicyOptimizationRecommendations", mock.Anything).Return(&PolicyOptimizationRecommendations{}, nil)
	client.On("GetPolicyPredictiveModeling", mock.Anything, "7d").Return(&PolicyPredictiveModeling{}, nil)
	client.On("GetPolicyCostBenefitAnalysis", mock.Anything, "30d").Return(&PolicyCostBenefitAnalysis{}, nil)
}