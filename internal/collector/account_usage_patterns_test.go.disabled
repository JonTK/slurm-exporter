package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountUsagePatternsSLURMClient struct {
	mock.Mock
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountUsagePatterns(ctx context.Context, accountName string) (*AccountUsagePatterns, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountUsagePatterns), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) ListAccounts(ctx context.Context) ([]*AccountInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AccountInfo), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountAnomalies(ctx context.Context, accountName string) ([]*AccountAnomaly, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AccountAnomaly), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountBehaviorProfile(ctx context.Context, accountName string) (*AccountBehaviorProfile, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountBehaviorProfile), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountTrendAnalysis(ctx context.Context, accountName string) (*AccountTrendAnalysis, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountTrendAnalysis), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountSeasonality(ctx context.Context, accountName string) (*AccountSeasonality, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountSeasonality), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountPredictions(ctx context.Context, accountName string) (*AccountPredictions, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountPredictions), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountClassification(ctx context.Context, accountName string) (*AccountClassification, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountClassification), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountComparison(ctx context.Context, accountA, accountB string) (*AccountComparison, error) {
	args := m.Called(ctx, accountA, accountB)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountComparison), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountRiskAnalysis(ctx context.Context, accountName string) (*AccountRiskAnalysis, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountRiskAnalysis), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountOptimizationSuggestions(ctx context.Context, accountName string) (*AccountOptimizationSuggestions, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountOptimizationSuggestions), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetAccountBenchmarks(ctx context.Context, accountName string) (*AccountBenchmarks, error) {
	args := m.Called(ctx, accountName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccountBenchmarks), args.Error(1)
}

func (m *MockAccountUsagePatternsSLURMClient) GetSystemUsageOverview(ctx context.Context) (*SystemUsageOverview, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SystemUsageOverview), args.Error(1)
}

func TestNewAccountUsagePatternsCollector(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
}

func TestAccountUsagePatternsCollector_Describe(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 57, count)
}

func TestAccountUsagePatternsCollector_Collect_Success(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	// Mock account data
	accounts := []*AccountInfo{
		{
			Name:        "research",
			Description: "Research account",
			Status:      "ACTIVE",
			CreatedAt:   time.Now().Add(-365 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
			UserCount:   25,
			JobCount:    1500,
		},
		{
			Name:        "development",
			Description: "Development account",
			Status:      "ACTIVE",
			CreatedAt:   time.Now().Add(-180 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
			UserCount:   12,
			JobCount:    650,
		},
	}

	// Mock usage patterns
	usagePatterns := &AccountUsagePatterns{
		AccountName: "research",
		UsageTimeSeries: []*TimeSeriesPoint{
			{
				Timestamp:       time.Now().Add(-24 * time.Hour),
				CPUHours:        1250.5,
				MemoryGB:        2048.0,
				GPUHours:        85.2,
				JobCount:        45,
				UserCount:       18,
				SubmissionCount: 52,
				WaitTime:        120.5,
				RunTime:         3600.0,
				CostUSD:         425.75,
			},
		},
		DailyPatterns: []*DailyUsagePattern{
			{
				Hour:              9,
				AverageCPUHours:   145.2,
				AverageMemoryGB:   256.0,
				AverageGPUHours:   12.5,
				AverageJobCount:   8.5,
				PeakUsageHour:     true,
				OffPeakHour:       false,
				UtilizationScore:  0.85,
			},
		},
		ResourceDistribution: &ResourceDistribution{
			CPUDistribution: &ResourceUsageDistribution{
				Percentile25: 50.0,
				Percentile50: 100.0,
				Percentile75: 200.0,
				Percentile90: 350.0,
				Percentile95: 500.0,
				Percentile99: 750.0,
				Mean:         180.5,
				StdDev:       125.3,
				Min:          10.0,
				Max:          800.0,
			},
		},
		UserActivityProfile: &UserActivityProfile{
			ActiveUsers:     25,
			PowerUsers:      8,
			OccasionalUsers: 12,
			NewUsers:        5,
			UserRetention:   0.92,
			UserGrowthRate:  0.15,
		},
		CostAnalysis: &CostAnalysis{
			TotalCost:      12500.75,
			CostPerCPUHour: 0.045,
			CostPerGPUHour: 2.15,
			CostPerJob:     8.33,
			CostPerUser:    500.03,
		},
		EfficiencyMetrics: &EfficiencyMetrics{
			OverallEfficiency: 0.78,
			CPUEfficiency:     0.82,
			MemoryEfficiency:  0.75,
			GPUEfficiency:     0.88,
			QueueEfficiency:   0.72,
			OptimizationScore: 0.85,
		},
		LastUpdated: time.Now(),
	}

	// Mock anomalies
	anomalies := []*AccountAnomaly{
		{
			AccountName:     "research",
			AnomalyType:     "USAGE_SPIKE",
			Severity:        "HIGH",
			DetectedAt:      time.Now().Add(-2 * time.Hour),
			Description:     "Unusual spike in CPU usage detected",
			AffectedMetrics: []string{"cpu_usage", "job_count"},
			DeviationScore:  2.5,
			Impact:          "Performance degradation",
			Recommendation:  "Investigate job scheduling",
			RootCause:       "Large batch job submission",
			IsResolved:      true,
			ResolvedAt:      &[]time.Time{time.Now().Add(-30 * time.Minute)}[0],
		},
		{
			AccountName:     "research",
			AnomalyType:     "COST_ANOMALY",
			Severity:        "MEDIUM",
			DetectedAt:      time.Now().Add(-4 * time.Hour),
			Description:     "Cost increase beyond expected range",
			AffectedMetrics: []string{"total_cost"},
			DeviationScore:  1.8,
			Impact:          "Budget impact",
			Recommendation:  "Review resource allocation",
			RootCause:       "Extended GPU usage",
			IsResolved:      false,
		},
	}

	// Mock behavior profile
	behaviorProfile := &AccountBehaviorProfile{
		AccountName:         "research",
		BehaviorType:        "RESEARCH_INTENSIVE",
		ActivityLevel:       "HIGH",
		ResourcePreference:  "GPU_HEAVY",
		SchedulingPattern:   "BATCH_ORIENTED",
		CollaborationLevel:  "HIGH",
		AdaptabilityScore:   0.78,
		ConsistencyScore:    0.85,
		EfficiencyScore:     0.82,
		InnovationScore:     0.90,
		RiskTolerance:       "MEDIUM",
		LearningRate:        0.75,
		BehaviorStability:   0.88,
		PredictabilityScore: 0.72,
	}

	// Mock trend analysis
	trendAnalysis := &AccountTrendAnalysis{
		AccountName:     "research",
		TrendDirection:  "INCREASING",
		GrowthRate:      0.15,
		TrendConfidence: 0.85,
		ForecastAccuracy: 0.78,
		Seasonality: &Seasonality{
			HasSeasonality:   true,
			SeasonalPeriod:   7,
			SeasonalStrength: 0.65,
			SeasonalFactors:  []float64{0.8, 0.9, 1.0, 1.1, 1.2, 0.7, 0.6},
			PeakSeason:       "WEEKDAYS",
			LowSeason:        "WEEKENDS",
		},
		OutlierDetection: &OutlierDetection{
			OutlierCount:        5,
			OutlierRate:         0.02,
			OutlierTypes:        []string{"usage_spike", "cost_spike"},
			LastOutlierDetected: time.Now().Add(-2 * time.Hour),
			OutlierTrend:        "STABLE",
		},
		ChangePointAnalysis: &ChangePointAnalysis{
			ChangePointCount:     3,
			LastChangePoint:      time.Now().Add(-30 * 24 * time.Hour),
			ChangePointFrequency: 0.1,
			ChangePointTypes:     []string{"usage_shift", "pattern_change"},
			ChangePointImpact:    0.25,
		},
	}

	// Mock seasonality
	seasonality := &AccountSeasonality{
		AccountName:     "research",
		SeasonalityScore: 0.75,
		DailySeasonality: &DailySeasonality{
			PeakHours:           []int{9, 10, 11, 14, 15, 16},
			OffPeakHours:        []int{0, 1, 2, 3, 4, 5, 6, 22, 23},
			SeasonalityStrength: 0.68,
			RegularityScore:     0.82,
		},
		WeeklySeasonality: &WeeklySeasonality{
			PeakDays:            []int{1, 2, 3, 4, 5},
			OffPeakDays:         []int{0, 6},
			WeekendPattern:      true,
			SeasonalityStrength: 0.75,
			RegularityScore:     0.88,
		},
		MonthlySeasonality: &MonthlySeasonality{
			PeakMonths:          []int{3, 4, 9, 10, 11},
			OffPeakMonths:       []int{6, 7, 8, 12},
			QuarterlyPattern:    true,
			SeasonalityStrength: 0.52,
			RegularityScore:     0.65,
		},
		PredictablePeriods: []string{"weekdays", "academic_term"},
		VolatilePeriods:    []string{"weekends", "holidays"},
	}

	// Mock predictions
	predictions := &AccountPredictions{
		AccountName: "research",
		ShortTermPredictions: &ShortTermPredictions{
			NextDayUsage: &UsagePrediction{
				PredictedCPUHours:  1500.0,
				PredictedMemoryGB:  2400.0,
				PredictedGPUHours:  95.0,
				PredictedJobCount:  55,
				PredictedUserCount: 22,
				PredictedCost:      485.25,
			},
			PredictionHorizon: 1,
			ConfidenceLevel:   0.85,
		},
		PredictionAccuracy: &PredictionAccuracy{
			MAE:              15.2,
			RMSE:             22.8,
			MAPE:             0.08,
			R2Score:          0.82,
			AccuracyScore:    0.78,
			ModelPerformance: "GOOD",
		},
		UncertaintyAnalysis: &UncertaintyAnalysis{
			ConfidenceIntervals: &ConfidenceIntervals{
				LowerBound:      1200.0,
				UpperBound:      1800.0,
				IntervalWidth:   600.0,
				ConfidenceLevel: 0.95,
			},
			PredictionVariance: 125.5,
			UncertaintyScore:   0.15,
			RiskFactors:        []string{"seasonal_variation", "external_events"},
		},
	}

	// Mock classification
	classification := &AccountClassification{
		AccountName:         "research",
		PrimaryClass:        "RESEARCH_INTENSIVE",
		SecondaryClasses:    []string{"GPU_HEAVY", "COLLABORATIVE"},
		ClassificationScore: 0.88,
		ClassificationModel: "RANDOM_FOREST",
		Confidence:          0.85,
		SimilarAccounts:     []string{"ai_lab", "physics_dept"},
		ClassTrends: &ClassTrends{
			ClassStability:     0.92,
			TrendDirection:     "STABLE",
			EvolutionSpeed:     0.05,
			ClassMigrationRisk: 0.12,
		},
	}

	// Mock risk analysis
	riskAnalysis := &AccountRiskAnalysis{
		AccountName:      "research",
		OverallRiskScore: 0.35,
		RiskLevel:        "MEDIUM",
		OperationalRisks: &OperationalRisks{
			ResourceExhaustion:     0.25,
			QueueBottlenecks:       0.15,
			PerformanceDegradation: 0.20,
			ServiceDisruption:      0.10,
			CapacityOverflow:       0.30,
		},
		FinancialRisks: &FinancialRisks{
			BudgetOverrun:        0.40,
			CostSpikes:           0.25,
			UncontrolledSpending: 0.15,
			ResourceWaste:        0.20,
			ROIDeterioration:     0.10,
		},
		SecurityRisks: &SecurityRisks{
			UnauthorizedAccess:  0.15,
			DataBreach:          0.10,
			PrivilegeEscalation: 0.05,
			MaliciousActivity:   0.08,
			CompromisedAccounts: 0.12,
		},
		ComplianceRisks: &ComplianceRisks{
			PolicyViolations: 0.18,
			RegulatoryBreach: 0.05,
			AuditFailures:    0.12,
			DataGovernance:   0.15,
			ComplianceDrift:  0.10,
		},
		RiskTrends: &RiskTrends{
			RiskDirection:    "STABLE",
			RiskVelocity:     0.02,
			RiskVolatility:   0.15,
			EmergingRisks:    []string{"ai_compliance", "data_privacy"},
			DiminishingRisks: []string{"legacy_systems"},
		},
	}

	// Mock optimization suggestions
	optimization := &AccountOptimizationSuggestions{
		AccountName:          "research",
		OverallScore:         0.72,
		ImprovementPotential: 0.28,
		ResourceOptimization: &ResourceOptimization{
			RightSizing: &RightSizing{
				OversizedJobs:    15,
				UndersizedJobs:   8,
				OptimalSizing:    []string{"reduce_memory_requests", "increase_cpu_allocation"},
				SavingsPotential: 1250.50,
			},
			ResourcePooling: &ResourcePooling{
				PoolingOpportunities: []string{"gpu_sharing", "memory_pooling"},
				SharedResources:      []string{"large_memory_nodes", "gpu_nodes"},
				PoolingBenefits:      0.20,
			},
		},
		PerformanceOptimization: &PerformanceOptimization{
			PerformanceGaps:     []string{"queue_time", "resource_utilization"},
			OptimizationTargets: []string{"reduce_wait_time", "improve_efficiency"},
			PerformanceActions:  []string{"priority_tuning", "scheduling_optimization"},
			ExpectedGains:       0.25,
		},
		PriorityActions: []string{"implement_right_sizing", "optimize_gpu_usage"},
		QuickWins:       []string{"update_job_templates", "user_training"},
		LongTermGoals:   []string{"workflow_automation", "resource_prediction"},
	}

	// Mock benchmarks
	benchmarks := &AccountBenchmarks{
		AccountName:      "research",
		BenchmarkSuite:   "HPC_RESEARCH",
		OverallRanking:   15,
		PeerGroupRanking: 8,
		IndustryRanking:  45,
		BenchmarkScores: &BenchmarkScores{
			EfficiencyScore:        0.78,
			PerformanceScore:       0.82,
			CostEffectivenessScore: 0.75,
			InnovationScore:        0.88,
			SustainabilityScore:    0.72,
			UserSatisfactionScore:  0.85,
		},
		ImprovementAreas: []string{"cost_optimization", "resource_efficiency"},
		BestPractices:    []string{"gpu_sharing", "batch_scheduling"},
	}

	// Mock system overview
	systemOverview := &SystemUsageOverview{
		TotalAccounts:   125,
		ActiveAccounts:  98,
		InactiveAccounts: 27,
		SystemUtilization: &SystemUtilization{
			CPUUtilization:     0.68,
			MemoryUtilization:  0.72,
			GPUUtilization:     0.85,
			StorageUtilization: 0.45,
			NetworkUtilization: 0.35,
			OverallUtilization: 0.65,
		},
		SystemHealth: &SystemHealth{
			HealthScore:       0.88,
			PerformanceScore:  0.82,
			StabilityScore:    0.90,
			AvailabilityScore: 0.95,
			ReliabilityScore:  0.87,
		},
		CapacityStatus: &CapacityStatus{
			CapacityUtilization: 0.72,
		},
		SystemAnomalies: []*SystemAnomaly{
			{
				AnomalyType:        "RESOURCE_SPIKE",
				Severity:           "MEDIUM",
				DetectedAt:         time.Now().Add(-1 * time.Hour),
				AffectedComponents: []string{"gpu_cluster"},
				Impact:             "Performance impact",
				RootCause:          "Large job submission",
				IsResolved:         false,
			},
		},
		SystemAlerts: []*SystemAlert{
			{
				AlertType:        "CAPACITY_WARNING",
				Priority:         "HIGH",
				Message:          "GPU utilization approaching capacity",
				TriggeredAt:      time.Now().Add(-30 * time.Minute),
				AffectedEntities: []string{"gpu_partition"},
				ActionRequired:   "Monitor and potentially scale",
				IsAcknowledged:   false,
			},
		},
	}

	// Set up mock expectations
	client.On("ListAccounts", mock.Anything).Return(accounts, nil)
	client.On("GetAccountUsagePatterns", mock.Anything, "research").Return(usagePatterns, nil)
	client.On("GetAccountUsagePatterns", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetAccountAnomalies", mock.Anything, "research").Return(anomalies, nil)
	client.On("GetAccountAnomalies", mock.Anything, "development").Return([]*AccountAnomaly{}, nil)
	client.On("GetAccountBehaviorProfile", mock.Anything, "research").Return(behaviorProfile, nil)
	client.On("GetAccountBehaviorProfile", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetAccountTrendAnalysis", mock.Anything, "research").Return(trendAnalysis, nil)
	client.On("GetAccountTrendAnalysis", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetAccountSeasonality", mock.Anything, "research").Return(seasonality, nil)
	client.On("GetAccountSeasonality", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetAccountPredictions", mock.Anything, "research").Return(predictions, nil)
	client.On("GetAccountPredictions", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetAccountClassification", mock.Anything, "research").Return(classification, nil)
	client.On("GetAccountClassification", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetAccountRiskAnalysis", mock.Anything, "research").Return(riskAnalysis, nil)
	client.On("GetAccountRiskAnalysis", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetAccountOptimizationSuggestions", mock.Anything, "research").Return(optimization, nil)
	client.On("GetAccountOptimizationSuggestions", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetAccountBenchmarks", mock.Anything, "research").Return(benchmarks, nil)
	client.On("GetAccountBenchmarks", mock.Anything, "development").Return(nil, errors.New("not found"))
	client.On("GetSystemUsageOverview", mock.Anything).Return(systemOverview, nil)

	ch := make(chan prometheus.Metric, 200)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	assert.Greater(t, len(metrics), 0)

	client.AssertExpectations(t)
}

func TestAccountUsagePatternsCollector_Collect_ListAccountsError(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	client.On("ListAccounts", mock.Anything).Return(nil, errors.New("API error"))

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	assert.Equal(t, 0, len(metrics))

	client.AssertExpectations(t)
}

func TestAccountUsagePatternsCollector_CollectUsagePatternsMetrics(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	patterns := &AccountUsagePatterns{
		AccountName: "test_account",
		CostAnalysis: &CostAnalysis{
			TotalCost:      5000.0,
			CostPerCPUHour: 0.05,
			CostPerGPUHour: 2.0,
			CostPerJob:     10.0,
			CostPerUser:    200.0,
		},
		EfficiencyMetrics: &EfficiencyMetrics{
			OverallEfficiency: 0.75,
			CPUEfficiency:     0.80,
			MemoryEfficiency:  0.70,
			GPUEfficiency:     0.85,
			QueueEfficiency:   0.65,
			OptimizationScore: 0.78,
		},
		UserActivityProfile: &UserActivityProfile{
			ActiveUsers:     20,
			PowerUsers:      5,
			OccasionalUsers: 10,
			NewUsers:        5,
			UserRetention:   0.90,
			UserGrowthRate:  0.10,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectUsagePatternsMetrics(ch, "test_account", patterns)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 10, count)
}

func TestAccountUsagePatternsCollector_CollectAnomalyMetrics(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	resolvedTime := time.Now()
	anomalies := []*AccountAnomaly{
		{
			AccountName:     "test_account",
			AnomalyType:     "USAGE_SPIKE",
			Severity:        "HIGH",
			DetectedAt:      time.Now().Add(-2 * time.Hour),
			DeviationScore:  2.5,
			IsResolved:      true,
			ResolvedAt:      &resolvedTime,
		},
		{
			AccountName:     "test_account",
			AnomalyType:     "COST_ANOMALY",
			Severity:        "MEDIUM",
			DetectedAt:      time.Now().Add(-1 * time.Hour),
			DeviationScore:  1.8,
			IsResolved:      false,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectAnomalyMetrics(ch, "test_account", anomalies)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestAccountUsagePatternsCollector_CollectBehaviorProfileMetrics(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	profile := &AccountBehaviorProfile{
		AccountName:         "test_account",
		BehaviorType:        "RESEARCH_INTENSIVE",
		ActivityLevel:       "HIGH",
		AdaptabilityScore:   0.78,
		ConsistencyScore:    0.85,
		InnovationScore:     0.90,
		PredictabilityScore: 0.72,
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectBehaviorProfileMetrics(ch, "test_account", profile)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 4, count)
}

func TestAccountUsagePatternsCollector_CollectTrendAnalysisMetrics(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	trends := &AccountTrendAnalysis{
		AccountName:     "test_account",
		GrowthRate:      0.15,
		TrendConfidence: 0.85,
		ForecastAccuracy: 0.78,
		OutlierDetection: &OutlierDetection{
			OutlierRate: 0.02,
		},
		ChangePointAnalysis: &ChangePointAnalysis{
			ChangePointFrequency: 0.1,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectTrendAnalysisMetrics(ch, "test_account", trends)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 5, count)
}

func TestAccountUsagePatternsCollector_CollectSeasonalityMetrics(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	seasonality := &AccountSeasonality{
		AccountName:     "test_account",
		SeasonalityScore: 0.75,
		DailySeasonality: &DailySeasonality{
			SeasonalityStrength: 0.68,
			RegularityScore:     0.82,
		},
		WeeklySeasonality: &WeeklySeasonality{
			SeasonalityStrength: 0.75,
			RegularityScore:     0.88,
		},
		MonthlySeasonality: &MonthlySeasonality{
			SeasonalityStrength: 0.52,
			RegularityScore:     0.65,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectSeasonalityMetrics(ch, "test_account", seasonality)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 4, count)
}

func TestAccountUsagePatternsCollector_CollectSystemOverviewMetrics(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	overview := &SystemUsageOverview{
		TotalAccounts:   100,
		ActiveAccounts:  85,
		InactiveAccounts: 15,
		SystemUtilization: &SystemUtilization{
			CPUUtilization:     0.68,
			MemoryUtilization:  0.72,
			GPUUtilization:     0.85,
			OverallUtilization: 0.75,
		},
		SystemHealth: &SystemHealth{
			HealthScore: 0.88,
		},
		CapacityStatus: &CapacityStatus{
			CapacityUtilization: 0.72,
		},
		SystemAnomalies: []*SystemAnomaly{
			{AnomalyType: "RESOURCE_SPIKE"},
		},
		SystemAlerts: []*SystemAlert{
			{AlertType: "CAPACITY_WARNING"},
		},
	}

	client.On("GetSystemUsageOverview", mock.Anything).Return(overview, nil)

	ch := make(chan prometheus.Metric, 50)
	collector.collectSystemOverviewMetrics(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 9, count)

	client.AssertExpectations(t)
}

func TestAccountUsagePatternsCollector_DescribeAll(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	gatherer := prometheus.Gatherer(registry)
	metricFamilies, err := gatherer.Gather()

	assert.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
}

func TestAccountUsagePatternsCollector_MetricCount(t *testing.T) {
	client := &MockAccountUsagePatternsSLURMClient{}
	collector := NewAccountUsagePatternsCollector(client)

	expectedMetrics := []string{
		"slurm_account_usage_score",
		"slurm_account_cpu_utilization",
		"slurm_account_memory_utilization",
		"slurm_account_gpu_utilization",
		"slurm_account_job_submissions",
		"slurm_account_active_users",
		"slurm_account_cost_analysis",
		"slurm_account_efficiency_score",
		"slurm_account_anomalies_total",
		"slurm_account_anomaly_severity",
		"slurm_account_anomaly_deviation_score",
		"slurm_account_anomaly_resolution_time_seconds",
		"slurm_account_behavior_type",
		"slurm_account_activity_level",
		"slurm_account_adaptability_score",
		"slurm_account_consistency_score",
		"slurm_account_innovation_score",
		"slurm_account_predictability_score",
		"slurm_account_growth_rate",
		"slurm_account_trend_confidence",
		"slurm_account_forecast_accuracy",
		"slurm_account_seasonality_strength",
		"slurm_account_outlier_rate",
		"slurm_account_change_point_frequency",
		"slurm_account_daily_seasonality",
		"slurm_account_weekly_seasonality",
		"slurm_account_monthly_seasonality",
		"slurm_account_yearly_seasonality",
		"slurm_account_seasonality_score",
		"slurm_account_prediction_accuracy",
		"slurm_account_prediction_confidence",
		"slurm_account_uncertainty_score",
		"slurm_account_prediction_variance",
		"slurm_account_classification_score",
		"slurm_account_class_confidence",
		"slurm_account_class_stability",
		"slurm_account_class_migration_risk",
		"slurm_account_similarity_score",
		"slurm_account_performance_ratio",
		"slurm_account_resource_ratio",
		"slurm_account_benchmark_score",
		"slurm_account_risk_score",
		"slurm_account_operational_risk",
		"slurm_account_financial_risk",
		"slurm_account_security_risk",
		"slurm_account_compliance_risk",
		"slurm_account_risk_velocity",
		"slurm_account_optimization_score",
		"slurm_account_improvement_potential",
		"slurm_account_right_sizing_savings",
		"slurm_account_resource_pooling_benefits",
		"slurm_account_performance_gain_potential",
		"slurm_account_overall_ranking",
		"slurm_account_peer_group_ranking",
		"slurm_account_industry_ranking",
		"slurm_account_benchmark_trend",
		"slurm_account_competitive_position",
		"slurm_system_total_accounts",
		"slurm_system_active_accounts",
		"slurm_system_utilization",
		"slurm_system_health_score",
		"slurm_system_capacity_utilization",
		"slurm_system_anomalies_total",
		"slurm_system_alerts_total",
	}

	assert.Equal(t, len(expectedMetrics), 57)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 57, count)
}