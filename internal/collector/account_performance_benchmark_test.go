package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountPerformanceBenchmarkSLURMClient struct {
	mock.Mock
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountPerformanceMetrics(ctx context.Context, accountName string) (*AccountPerformanceMetrics, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountPerformanceMetrics), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountBenchmarkResults(ctx context.Context, accountName string) ([]*AccountBenchmarkResult, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*AccountBenchmarkResult), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountPerformanceComparisons(ctx context.Context, accountName string) (*AccountPerformanceComparisons, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountPerformanceComparisons), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountPerformanceTrends(ctx context.Context, accountName string) (*AccountPerformanceTrends, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountPerformanceTrends), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountEfficiencyAnalysis(ctx context.Context, accountName string) (*AccountEfficiencyAnalysis, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountEfficiencyAnalysis), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountResourceUtilization(ctx context.Context, accountName string) (*AccountResourceUtilization, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountResourceUtilization), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountWorkloadCharacteristics(ctx context.Context, accountName string) (*AccountWorkloadCharacteristics, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountWorkloadCharacteristics), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountPerformanceBaselines(ctx context.Context, accountName string) ([]*AccountPerformanceBaseline, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*AccountPerformanceBaseline), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountSLACompliance(ctx context.Context, accountName string) (*AccountSLACompliance, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountSLACompliance), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountPerformanceOptimization(ctx context.Context, accountName string) ([]*AccountPerformanceOptimization, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*AccountPerformanceOptimization), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountCapacityPredictions(ctx context.Context, accountName string) (*AccountCapacityPredictions, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountCapacityPredictions), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountPerformanceAlerts(ctx context.Context, accountName string) ([]*AccountPerformanceAlert, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*AccountPerformanceAlert), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetAccountBenchmarkHistory(ctx context.Context, accountName string) ([]*AccountBenchmarkHistory, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*AccountBenchmarkHistory), args.Error(1)
}

func (m *MockAccountPerformanceBenchmarkSLURMClient) GetSystemPerformanceOverview(ctx context.Context) (*SystemPerformanceOverview, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemPerformanceOverview), args.Error(1)
}

func TestAccountPerformanceBenchmarkCollector_PerformanceMetrics(t *testing.T) {
	mockClient := new(MockAccountPerformanceBenchmarkSLURMClient)
	collector := NewAccountPerformanceBenchmarkCollector(mockClient)

	performanceMetrics := &AccountPerformanceMetrics{
		AccountName:              "test-account",
		TotalJobsCompleted:       15420,
		AverageJobDuration:       2*time.Hour + 30*time.Minute,
		MedianJobDuration:        1*time.Hour + 45*time.Minute,
		JobThroughput:            12.5,
		JobSuccessRate:           0.94,
		JobFailureRate:           0.06,
		QueueWaitTime:            15 * time.Minute,
		AverageQueueTime:         25 * time.Minute,
		MedianQueueTime:          18 * time.Minute,
		TurnaroundTime:           3 * time.Hour,
		ResourceUtilizationCPU:   0.78,
		ResourceUtilizationGPU:   0.85,
		ResourceUtilizationMemory: 0.72,
		ResourceUtilizationStorage: 0.68,
		NetworkUtilization:       0.55,
		EnergyEfficiency:         0.82,
		CostEfficiency:           0.76,
		PerformanceScore:         0.85,
		ReliabilityScore:         0.92,
		AvailabilityScore:        0.98,
		ScalabilityScore:         0.75,
		PerformanceGrade:         "B",
		LastUpdated:              time.Now(),
		MeasurementPeriod:        "monthly",
	}

	benchmarkResults := []*AccountBenchmarkResult{
		{
			BenchmarkID:          "bench-001",
			AccountName:          "test-account",
			BenchmarkType:        "performance",
			BenchmarkName:        "cpu_intensive_workload",
			ExecutionTime:        time.Now(),
			Duration:             30 * time.Minute,
			Score:                850.5,
			ReferenceScore:       800.0,
			PerformanceRatio:     1.063,
			Percentile:           75.5,
			Ranking:              15,
			TotalParticipants:    120,
			BenchmarkCategory:    "compute",
			WorkloadType:         "cpu_intensive",
			ResourceProfile:      "high_cpu",
			TestConditions:       map[string]interface{}{"nodes": 4, "cores": 32},
			Metrics:              map[string]float64{"throughput": 850.5, "latency": 125.3},
			ComparisonBaseline:   "industry_standard",
			PerformanceClass:     "high",
			AccuracyLevel:        0.95,
			ConfidenceInterval:   0.92,
			StandardDeviation:    45.2,
			Variance:             2043.04,
			OptimizationPotential: 0.15,
			RecommendedActions:   []string{"optimize_memory_allocation", "tune_cpu_affinity"},
		},
		{
			BenchmarkID:          "bench-002",
			AccountName:          "test-account",
			BenchmarkType:        "throughput",
			BenchmarkName:        "data_processing_pipeline",
			ExecutionTime:        time.Now().Add(-time.Hour),
			Duration:             45 * time.Minute,
			Score:                1250.8,
			ReferenceScore:       1200.0,
			PerformanceRatio:     1.042,
			Percentile:           68.2,
			Ranking:              25,
			TotalParticipants:    120,
			BenchmarkCategory:    "data_processing",
			WorkloadType:         "mixed",
			ResourceProfile:      "balanced",
			TestConditions:       map[string]interface{}{"dataset_size": "10GB", "partitions": 8},
			Metrics:              map[string]float64{"records_per_second": 1250.8, "memory_efficiency": 0.78},
			ComparisonBaseline:   "peer_average",
			PerformanceClass:     "medium",
			AccuracyLevel:        0.88,
			ConfidenceInterval:   0.85,
			StandardDeviation:    62.1,
			Variance:             3856.41,
			OptimizationPotential: 0.22,
			RecommendedActions:   []string{"increase_parallelism", "optimize_io_operations"},
		},
	}

	performanceComparisons := &AccountPerformanceComparisons{
		AccountName:               "test-account",
		PeerAccounts:              []string{"peer1", "peer2", "peer3"},
		PeerAveragePerformance:    0.78,
		IndustryBenchmark:         0.82,
		TopPerformerBenchmark:     0.95,
		PerformanceRank:           45,
		TotalAccountsCompared:     150,
		PerformancePercentile:     70.0,
		RelativePerformance:       1.09,
		PerformanceGap:            0.10,
		CompetitivePosition:       "above_average",
		StrengthAreas:             []string{"reliability", "cost_efficiency"},
		ImprovementAreas:          []string{"throughput", "scalability"},
		BenchmarkMetrics:          map[string]float64{"cpu_performance": 0.85, "memory_efficiency": 0.72},
		ComparisonPeriod:          "monthly",
		ComparisonMethodology:     "statistical_analysis",
		StatisticalSignificance:   0.95,
		ConfidenceLevel:           0.90,
		LastComparison:            time.Now(),
		TrendDirection:            "improving",
		YearOverYearChange:        0.12,
		QuarterOverQuarterChange:  0.05,
	}

	performanceTrends := &AccountPerformanceTrends{
		AccountName:              "test-account",
		TrendPeriod:              "quarterly",
		PerformanceTrend:         "improving",
		TrendStrength:            0.75,
		TrendDirection:           "upward",
		LinearRegression:         map[string]float64{"slope": 0.05, "r_squared": 0.82},
		SeasonalPatterns:         map[string]float64{"Q1": 1.1, "Q2": 0.9, "Q3": 1.0, "Q4": 1.2},
		CyclicalPatterns:         []string{"quarterly_peaks", "monthly_dips"},
		AnomalyCount:             3,
		TrendChangePoints:        []time.Time{time.Now().AddDate(0, -6, 0)},
		ForecastAccuracy:         0.88,
		PredictiveConfidence:     0.85,
		HistoricalVariability:    0.15,
		PerformanceVolatility:    0.18,
		GrowthRate:               0.08,
		DeclineRate:              0.02,
		StabilityIndex:           0.82,
		ConsistencyScore:         0.78,
		ReliabilityTrend:         "stable",
		QualityTrend:             "improving",
		EfficiencyTrend:          "improving",
		ScalabilityTrend:         "declining",
		TrendFactors:             []string{"workload_changes", "infrastructure_upgrades"},
		ExternalInfluences:       []string{"market_conditions", "seasonal_demand"},
	}

	efficiencyAnalysis := &AccountEfficiencyAnalysis{
		AccountName:               "test-account",
		OverallEfficiencyScore:    0.78,
		CPUEfficiency:             0.82,
		GPUEfficiency:             0.75,
		MemoryEfficiency:          0.68,
		StorageEfficiency:         0.72,
		NetworkEfficiency:         0.85,
		EnergyEfficiency:          0.76,
		CostEfficiency:            0.81,
		TimeEfficiency:            0.79,
		ResourceWaste:             0.22,
		IdleTime:                  0.18,
		OverprovisioningRatio:     0.15,
		UnderprovisioningRatio:    0.08,
		OptimalResourceAllocation: map[string]float64{"cpu": 0.85, "memory": 0.75, "storage": 0.70},
		BottleneckAnalysis:        map[string]float64{"memory": 0.68, "io": 0.55},
		PerformanceConstraints:    []string{"memory_bandwidth", "storage_iops"},
		EfficiencyFactors:         map[string]float64{"workload_balance": 0.75, "resource_contention": 0.25},
		WasteReductionPotential:   0.15,
		OptimizationOpportunities: []string{"memory_tuning", "io_optimization"},
		EfficiencyTrends:          map[string]float64{"cpu": 0.05, "memory": -0.02},
		BenchmarkComparison:       map[string]float64{"peer_average": 0.72, "industry_best": 0.90},
		TargetEfficiency:          0.85,
		EfficiencyGap:             0.07,
		ImprovementPotential:      0.20,
	}

	workloadCharacteristics := &AccountWorkloadCharacteristics{
		AccountName:              "test-account",
		WorkloadTypes:            []string{"batch", "interactive", "long_running"},
		JobSizeDistribution:      map[string]float64{"small": 0.40, "medium": 0.45, "large": 0.15},
		JobDurationDistribution:  map[string]float64{"short": 0.30, "medium": 0.50, "long": 0.20},
		ResourceDemandPatterns:   map[string]float64{"cpu_intensive": 0.35, "memory_intensive": 0.25, "io_intensive": 0.20, "balanced": 0.20},
		TemporalPatterns:         map[string]float64{"peak_hours": 0.65, "off_hours": 0.35},
		ConcurrencyPatterns:      map[string]float64{"low": 0.25, "medium": 0.50, "high": 0.25},
		BatchJobCharacteristics:  map[string]float64{"average_duration": 3600, "resource_efficiency": 0.75},
		InteractiveJobCharacteristics: map[string]float64{"response_time": 2.5, "user_satisfaction": 0.85},
		LongRunningJobCharacteristics: map[string]float64{"stability": 0.90, "resource_consumption": 0.65},
		CPUIntensiveRatio:        0.35,
		GPUIntensiveRatio:        0.15,
		MemoryIntensiveRatio:     0.25,
		IOIntensiveRatio:         0.20,
		NetworkIntensiveRatio:    0.05,
		MixedWorkloadRatio:       0.20,
		WorkloadComplexity:       0.65,
		WorkloadVariability:      0.45,
		WorkloadPredictability:   0.70,
		SeasonalityIndex:         0.25,
		WorkloadEfficiency:       0.78,
		ResourceDiversity:        0.60,
		ApplicationProfiles:      map[string]float64{"scientific": 0.40, "data_analytics": 0.35, "web_services": 0.25},
		UserBehaviorPatterns:     map[string]float64{"regular": 0.60, "bursty": 0.25, "sporadic": 0.15},
		WorkloadTrends:           map[string]float64{"growth": 0.08, "complexity_increase": 0.12},
	}

	slaCompliance := &AccountSLACompliance{
		AccountName:              "test-account",
		SLATargets:               map[string]float64{"availability": 0.99, "performance": 0.95, "response_time": 2.0},
		CurrentPerformance:       map[string]float64{"availability": 0.985, "performance": 0.92, "response_time": 2.3},
		ComplianceStatus:         map[string]string{"availability": "compliant", "performance": "warning", "response_time": "violation"},
		CompliancePercentage:     map[string]float64{"availability": 99.5, "performance": 96.8, "response_time": 85.2},
		ViolationCount:           map[string]int{"availability": 0, "performance": 2, "response_time": 5},
		ViolationSeverity:        map[string]string{"performance": "medium", "response_time": "high"},
		TimeToResolution:         map[string]time.Duration{"performance": 4 * time.Hour, "response_time": 2 * time.Hour},
		SLACredits:               map[string]float64{"response_time": 150.0},
		PerformanceMargin:        map[string]float64{"availability": 0.5, "performance": -2.1, "response_time": -15.0},
		RiskLevel:                "medium",
		ComplianceScore:          0.89,
		AvailabilityCompliance:   99.5,
		PerformanceCompliance:    96.8,
		ReliabilityCompliance:    98.2,
		SLAPeriod:                "monthly",
		MonitoringFrequency:      "real_time",
		ReportingPeriod:          "weekly",
		EscalationTriggers:       map[string]interface{}{"response_time": map[string]interface{}{"threshold": 3.0, "duration": "5m"}},
		RemediationActions:       []string{"scale_resources", "optimize_algorithms"},
		SLAReviewDate:            time.Now().AddDate(0, 1, 0),
		ContractualObligations:   map[string]interface{}{"penalty_rate": 0.1, "maximum_credits": 1000.0},
		PenaltyExposure:          350.0,
		ComplianceHistory:        []string{"compliant", "compliant", "warning", "violation"},
		TrendAnalysis:            map[string]string{"availability": "stable", "performance": "declining", "response_time": "improving"},
	}

	optimizations := []*AccountPerformanceOptimization{
		{
			OptimizationID:           "opt-001",
			AccountName:              "test-account",
			OptimizationType:         "resource_allocation",
			OptimizationArea:         "memory_management",
			CurrentPerformance:       0.68,
			TargetPerformance:        0.82,
			PotentialImprovement:     0.20,
			ImplementationComplexity: "medium",
			ImplementationCost:       5000.0,
			ExpectedROI:              2.5,
			TimeToImplement:          4 * time.Hour,
			RiskLevel:                "low",
			Priority:                 "high",
			Status:                   "pending",
			Recommendation:           "Implement memory pooling and optimize garbage collection",
			DetailedAnalysis:         "Memory fragmentation is causing performance degradation",
			Prerequisites:            []string{"memory_profiling", "baseline_establishment"},
			Dependencies:             []string{"infrastructure_upgrade"},
			ResourceRequirements:     map[string]interface{}{"engineering_hours": 40, "testing_resources": "dedicated_cluster"},
			ExpectedBenefits:         []string{"improved_throughput", "reduced_latency", "better_resource_utilization"},
			PotentialRisks:           []string{"temporary_performance_degradation", "compatibility_issues"},
			MitigationStrategies:     []string{"phased_rollout", "rollback_plan"},
			SuccessMetrics:           []string{"memory_efficiency_improvement", "response_time_reduction"},
			MonitoringPlan:           "Real-time performance monitoring with alerting",
			ReviewSchedule:           "weekly",
			ApprovalRequired:         true,
			BusinessImpact:           "high",
		},
	}

	capacityPredictions := &AccountCapacityPredictions{
		AccountName:              "test-account",
		PredictionPeriod:         "quarterly",
		PredictionConfidence:     0.85,
		CPUCapacityForecast:      map[string]float64{"1_month": 120.5, "3_months": 145.2, "6_months": 168.7},
		GPUCapacityForecast:      map[string]float64{"1_month": 45.2, "3_months": 52.8, "6_months": 61.3},
		MemoryCapacityForecast:   map[string]float64{"1_month": 2048.0, "3_months": 2560.0, "6_months": 3072.0},
		StorageCapacityForecast:  map[string]float64{"1_month": 5120.0, "3_months": 6144.0, "6_months": 7168.0},
		NetworkCapacityForecast:  map[string]float64{"1_month": 10.5, "3_months": 12.8, "6_months": 15.2},
		WorkloadGrowthPrediction: 0.15,
		UserGrowthPrediction:     0.08,
		ResourceDemandGrowth:     map[string]float64{"cpu": 0.12, "memory": 0.18, "storage": 0.25},
		CapacityUtilizationForecast: map[string]float64{"cpu": 0.85, "memory": 0.78, "storage": 0.65},
		BottleneckPredictions:    []string{"memory", "storage_io"},
		ScalingRequirements:      map[string]interface{}{"nodes": 5, "memory_per_node": "256GB"},
		CapacityShortfalls:       map[string]float64{"memory": 512.0, "storage": 1024.0},
		CapacityExcess:           map[string]float64{"cpu": 25.5, "network": 5.2},
		OptimalCapacityPlan:      map[string]interface{}{"target_utilization": 0.80, "buffer_capacity": 0.20},
		InvestmentRequirements:   map[string]float64{"hardware": 150000.0, "software": 25000.0},
		CostProjections:          map[string]float64{"1_month": 12000.0, "3_months": 38000.0, "6_months": 78000.0},
		RiskFactors:              []string{"budget_constraints", "vendor_delays", "technology_changes"},
		ScenarioAnalysis:         map[string]interface{}{"conservative": 0.85, "aggressive": 1.25},
		ContingencyPlanning:      map[string]interface{}{"cloud_burst": true, "temporary_resources": "available"},
		CapacityMilestones:       []string{"Q1_expansion", "Q2_optimization", "Q3_evaluation"},
		ReviewPoints:             []time.Time{time.Now().AddDate(0, 1, 0), time.Now().AddDate(0, 3, 0)},
		AccuracyTracking:         map[string]float64{"historical_accuracy": 0.88, "trend_accuracy": 0.82},
		ModelValidation:          map[string]float64{"r_squared": 0.85, "mean_absolute_error": 0.12},
	}

	alerts := []*AccountPerformanceAlert{
		{
			AlertID:              "alert-001",
			AccountName:          "test-account",
			AlertType:            "performance_degradation",
			Severity:             "high",
			Metric:               "response_time",
			CurrentValue:         3.5,
			ThresholdValue:       2.0,
			DeviationPercentage:  75.0,
			AlertTime:            time.Now().Add(-time.Hour),
			Status:               "active",
			Description:          "Response time exceeds SLA threshold",
			Impact:               "User experience degradation",
			RootCause:            "Memory bottleneck detected",
			RecommendedActions:   []string{"scale_memory", "optimize_queries"},
			EscalationLevel:      2,
			NotificationsSent:    3,
			AcknowledgedBy:       "ops_team",
			AcknowledgedTime:     time.Now().Add(-30 * time.Minute),
			ResolvedTime:         time.Time{},
			ResolutionNotes:      "",
			RelatedAlerts:        []string{"alert-002", "alert-003"},
			TrendData:            map[string]float64{"1h_avg": 3.2, "24h_avg": 2.8},
			HistoricalContext:    "Similar pattern observed last month during peak usage",
			BusinessImpact:       "medium",
			TechnicalImpact:      "high",
			SLAImpact:            "violation",
			AutoRemediationApplied: true,
			ManualInterventionRequired: true,
			AlertFrequency:       "recurring",
			SuppressUntil:        time.Time{},
		},
	}

	systemOverview := &SystemPerformanceOverview{
		OverviewDate:             time.Now(),
		TotalAccounts:            150,
		AveragePerformanceScore:  0.78,
		MedianPerformanceScore:   0.82,
		PerformanceDistribution:  map[string]int{"A": 25, "B": 45, "C": 55, "D": 20, "F": 5},
		TopPerformingAccounts:    []string{"account_premium_1", "account_premium_2"},
		UnderperformingAccounts:  []string{"account_basic_1", "account_basic_2"},
		PerformanceTrends:        map[string]float64{"month_over_month": 0.05, "quarter_over_quarter": 0.12},
		SystemUtilization:        map[string]float64{"cpu": 0.75, "memory": 0.68, "storage": 0.55, "network": 0.45},
		SystemEfficiency:         0.78,
		SystemReliability:        0.92,
		SystemAvailability:       0.98,
		SystemScalability:        0.75,
		ResourceContention:       map[string]float64{"cpu": 0.15, "memory": 0.25, "storage": 0.10},
		CapacityUtilization:      map[string]float64{"cpu": 0.75, "memory": 0.68, "storage": 0.55},
		PerformanceBottlenecks:   []string{"memory_bandwidth", "storage_iops"},
		SystemAlerts:             25,
		SystemViolations:         8,
		ComplianceScore:          0.89,
		OptimizationOpportunities: 42,
		SystemHealthScore:        0.85,
		OverallSystemGrade:       "B+",
		BenchmarkComparisons:     map[string]float64{"industry_average": 0.72, "best_in_class": 0.95},
		IndustryRanking:          15,
		SystemMetrics:            map[string]float64{"throughput": 1250.5, "latency": 125.3},
		QualityMetrics:           map[string]float64{"availability": 0.98, "reliability": 0.92},
		OperationalMetrics:       map[string]float64{"cost_per_job": 2.5, "energy_efficiency": 0.82},
	}

	mockClient.On("GetAccountPerformanceMetrics", mock.Anything, "account1").Return(performanceMetrics, nil)
	mockClient.On("GetAccountPerformanceMetrics", mock.Anything, "account2").Return(performanceMetrics, nil)
	mockClient.On("GetAccountPerformanceMetrics", mock.Anything, "account3").Return(performanceMetrics, nil)

	mockClient.On("GetAccountBenchmarkResults", mock.Anything, "account1").Return(benchmarkResults, nil)
	mockClient.On("GetAccountBenchmarkResults", mock.Anything, "account2").Return(benchmarkResults, nil)
	mockClient.On("GetAccountBenchmarkResults", mock.Anything, "account3").Return(benchmarkResults, nil)

	mockClient.On("GetAccountPerformanceComparisons", mock.Anything, "account1").Return(performanceComparisons, nil)
	mockClient.On("GetAccountPerformanceComparisons", mock.Anything, "account2").Return(performanceComparisons, nil)
	mockClient.On("GetAccountPerformanceComparisons", mock.Anything, "account3").Return(performanceComparisons, nil)

	mockClient.On("GetAccountPerformanceTrends", mock.Anything, "account1").Return(performanceTrends, nil)
	mockClient.On("GetAccountPerformanceTrends", mock.Anything, "account2").Return(performanceTrends, nil)
	mockClient.On("GetAccountPerformanceTrends", mock.Anything, "account3").Return(performanceTrends, nil)

	mockClient.On("GetAccountEfficiencyAnalysis", mock.Anything, "account1").Return(efficiencyAnalysis, nil)
	mockClient.On("GetAccountEfficiencyAnalysis", mock.Anything, "account2").Return(efficiencyAnalysis, nil)
	mockClient.On("GetAccountEfficiencyAnalysis", mock.Anything, "account3").Return(efficiencyAnalysis, nil)

	mockClient.On("GetAccountResourceUtilization", mock.Anything, mock.Anything).Return(&AccountResourceUtilization{}, nil)

	mockClient.On("GetAccountWorkloadCharacteristics", mock.Anything, "account1").Return(workloadCharacteristics, nil)
	mockClient.On("GetAccountWorkloadCharacteristics", mock.Anything, "account2").Return(workloadCharacteristics, nil)
	mockClient.On("GetAccountWorkloadCharacteristics", mock.Anything, "account3").Return(workloadCharacteristics, nil)

	mockClient.On("GetAccountSLACompliance", mock.Anything, "account1").Return(slaCompliance, nil)
	mockClient.On("GetAccountSLACompliance", mock.Anything, "account2").Return(slaCompliance, nil)
	mockClient.On("GetAccountSLACompliance", mock.Anything, "account3").Return(slaCompliance, nil)

	mockClient.On("GetAccountPerformanceOptimization", mock.Anything, "account1").Return(optimizations, nil)
	mockClient.On("GetAccountPerformanceOptimization", mock.Anything, "account2").Return(optimizations, nil)
	mockClient.On("GetAccountPerformanceOptimization", mock.Anything, "account3").Return(optimizations, nil)

	mockClient.On("GetAccountCapacityPredictions", mock.Anything, "account1").Return(capacityPredictions, nil)
	mockClient.On("GetAccountCapacityPredictions", mock.Anything, "account2").Return(capacityPredictions, nil)
	mockClient.On("GetAccountCapacityPredictions", mock.Anything, "account3").Return(capacityPredictions, nil)

	mockClient.On("GetAccountPerformanceAlerts", mock.Anything, "account1").Return(alerts, nil)
	mockClient.On("GetAccountPerformanceAlerts", mock.Anything, "account2").Return(alerts, nil)
	mockClient.On("GetAccountPerformanceAlerts", mock.Anything, "account3").Return(alerts, nil)

	mockClient.On("GetSystemPerformanceOverview", mock.Anything).Return(systemOverview, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountPerformanceBenchmarkCollector_BenchmarkResults(t *testing.T) {
	mockClient := new(MockAccountPerformanceBenchmarkSLURMClient)
	collector := NewAccountPerformanceBenchmarkCollector(mockClient)

	benchmarkResults := []*AccountBenchmarkResult{
		{
			BenchmarkID:          "bench-001",
			AccountName:          "test-account",
			BenchmarkType:        "cpu_performance",
			BenchmarkName:        "LINPACK",
			Score:                1250.5,
			ReferenceScore:       1200.0,
			PerformanceRatio:     1.042,
			Percentile:           75.2,
			Ranking:              15,
			AccuracyLevel:        0.95,
			ConfidenceInterval:   0.90,
		},
		{
			BenchmarkID:          "bench-002",
			AccountName:          "test-account",
			BenchmarkType:        "memory_bandwidth",
			BenchmarkName:        "STREAM",
			Score:                850.3,
			ReferenceScore:       800.0,
			PerformanceRatio:     1.063,
			Percentile:           68.7,
			Ranking:             22,
			AccuracyLevel:        0.92,
			ConfidenceInterval:   0.88,
		},
	}

	mockClient.On("GetAccountPerformanceMetrics", mock.Anything, mock.Anything).Return(&AccountPerformanceMetrics{}, nil)
	mockClient.On("GetAccountBenchmarkResults", mock.Anything, "account1").Return(benchmarkResults, nil)
	mockClient.On("GetAccountPerformanceComparisons", mock.Anything, mock.Anything).Return(&AccountPerformanceComparisons{}, nil)
	mockClient.On("GetAccountPerformanceTrends", mock.Anything, mock.Anything).Return(&AccountPerformanceTrends{}, nil)
	mockClient.On("GetAccountEfficiencyAnalysis", mock.Anything, mock.Anything).Return(&AccountEfficiencyAnalysis{}, nil)
	mockClient.On("GetAccountResourceUtilization", mock.Anything, mock.Anything).Return(&AccountResourceUtilization{}, nil)
	mockClient.On("GetAccountWorkloadCharacteristics", mock.Anything, mock.Anything).Return(&AccountWorkloadCharacteristics{}, nil)
	mockClient.On("GetAccountSLACompliance", mock.Anything, mock.Anything).Return(&AccountSLACompliance{ViolationCount: map[string]int{}, SLACredits: map[string]float64{}}, nil)
	mockClient.On("GetAccountPerformanceOptimization", mock.Anything, mock.Anything).Return([]*AccountPerformanceOptimization{}, nil)
	mockClient.On("GetAccountCapacityPredictions", mock.Anything, mock.Anything).Return(&AccountCapacityPredictions{}, nil)
	mockClient.On("GetAccountPerformanceAlerts", mock.Anything, mock.Anything).Return([]*AccountPerformanceAlert{}, nil)
	mockClient.On("GetSystemPerformanceOverview", mock.Anything).Return(&SystemPerformanceOverview{SystemUtilization: map[string]float64{}}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect benchmark metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountPerformanceBenchmarkCollector_AlertMetrics(t *testing.T) {
	mockClient := new(MockAccountPerformanceBenchmarkSLURMClient)
	collector := NewAccountPerformanceBenchmarkCollector(mockClient)

	alerts := []*AccountPerformanceAlert{
		{
			AlertID:              "alert-001",
			AccountName:          "test-account",
			AlertType:            "response_time",
			Severity:             "high",
			CurrentValue:         3.5,
			ThresholdValue:       2.0,
			AlertTime:            time.Now().Add(-time.Hour),
			ResolvedTime:         time.Now().Add(-30 * time.Minute),
			EscalationLevel:      2,
			AutoRemediationApplied: true,
		},
		{
			AlertID:              "alert-002",
			AccountName:          "test-account",
			AlertType:            "throughput",
			Severity:             "medium",
			CurrentValue:         850.0,
			ThresholdValue:       1000.0,
			AlertTime:            time.Now().Add(-2 * time.Hour),
			ResolvedTime:         time.Now().Add(-time.Hour),
			EscalationLevel:      1,
			AutoRemediationApplied: false,
		},
	}

	mockClient.On("GetAccountPerformanceMetrics", mock.Anything, mock.Anything).Return(&AccountPerformanceMetrics{}, nil)
	mockClient.On("GetAccountBenchmarkResults", mock.Anything, mock.Anything).Return([]*AccountBenchmarkResult{}, nil)
	mockClient.On("GetAccountPerformanceComparisons", mock.Anything, mock.Anything).Return(&AccountPerformanceComparisons{}, nil)
	mockClient.On("GetAccountPerformanceTrends", mock.Anything, mock.Anything).Return(&AccountPerformanceTrends{}, nil)
	mockClient.On("GetAccountEfficiencyAnalysis", mock.Anything, mock.Anything).Return(&AccountEfficiencyAnalysis{}, nil)
	mockClient.On("GetAccountResourceUtilization", mock.Anything, mock.Anything).Return(&AccountResourceUtilization{}, nil)
	mockClient.On("GetAccountWorkloadCharacteristics", mock.Anything, mock.Anything).Return(&AccountWorkloadCharacteristics{}, nil)
	mockClient.On("GetAccountSLACompliance", mock.Anything, mock.Anything).Return(&AccountSLACompliance{ViolationCount: map[string]int{}, SLACredits: map[string]float64{}}, nil)
	mockClient.On("GetAccountPerformanceOptimization", mock.Anything, mock.Anything).Return([]*AccountPerformanceOptimization{}, nil)
	mockClient.On("GetAccountCapacityPredictions", mock.Anything, mock.Anything).Return(&AccountCapacityPredictions{}, nil)
	mockClient.On("GetAccountPerformanceAlerts", mock.Anything, "account1").Return(alerts, nil)
	mockClient.On("GetSystemPerformanceOverview", mock.Anything).Return(&SystemPerformanceOverview{SystemUtilization: map[string]float64{}}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect alert metrics")

	mockClient.AssertExpectations(t)
}

func TestAccountPerformanceBenchmarkCollector_ErrorHandling(t *testing.T) {
	mockClient := new(MockAccountPerformanceBenchmarkSLURMClient)
	collector := NewAccountPerformanceBenchmarkCollector(mockClient)

	mockClient.On("GetAccountPerformanceMetrics", mock.Anything, mock.Anything).Return((*AccountPerformanceMetrics)(nil), assert.AnError)
	mockClient.On("GetAccountBenchmarkResults", mock.Anything, mock.Anything).Return(([]*AccountBenchmarkResult)(nil), assert.AnError)
	mockClient.On("GetAccountPerformanceComparisons", mock.Anything, mock.Anything).Return((*AccountPerformanceComparisons)(nil), assert.AnError)
	mockClient.On("GetAccountPerformanceTrends", mock.Anything, mock.Anything).Return((*AccountPerformanceTrends)(nil), assert.AnError)
	mockClient.On("GetAccountEfficiencyAnalysis", mock.Anything, mock.Anything).Return((*AccountEfficiencyAnalysis)(nil), assert.AnError)
	mockClient.On("GetAccountResourceUtilization", mock.Anything, mock.Anything).Return((*AccountResourceUtilization)(nil), assert.AnError)
	mockClient.On("GetAccountWorkloadCharacteristics", mock.Anything, mock.Anything).Return((*AccountWorkloadCharacteristics)(nil), assert.AnError)
	mockClient.On("GetAccountSLACompliance", mock.Anything, mock.Anything).Return((*AccountSLACompliance)(nil), assert.AnError)
	mockClient.On("GetAccountPerformanceOptimization", mock.Anything, mock.Anything).Return(([]*AccountPerformanceOptimization)(nil), assert.AnError)
	mockClient.On("GetAccountCapacityPredictions", mock.Anything, mock.Anything).Return((*AccountCapacityPredictions)(nil), assert.AnError)
	mockClient.On("GetAccountPerformanceAlerts", mock.Anything, mock.Anything).Return(([]*AccountPerformanceAlert)(nil), assert.AnError)
	mockClient.On("GetSystemPerformanceOverview", mock.Anything).Return((*SystemPerformanceOverview)(nil), assert.AnError)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Equal(t, 0, metricCount, "Should not collect metrics on error")

	mockClient.AssertExpectations(t)
}