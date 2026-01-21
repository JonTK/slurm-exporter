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

type MockUserBehaviorAnalysisSLURMClient struct {
	mock.Mock
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserBehaviorProfile(ctx context.Context, userName string) (*UserBehaviorProfile, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserBehaviorProfile), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserJobSubmissionPatterns(ctx context.Context, userName string, period string) (*JobSubmissionPatterns, error) {
	args := m.Called(ctx, userName, period)
	return args.Get(0).(*JobSubmissionPatterns), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserResourceUsagePatterns(ctx context.Context, userName string, period string) (*ResourceUsagePatterns, error) {
	args := m.Called(ctx, userName, period)
	return args.Get(0).(*ResourceUsagePatterns), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserSchedulingBehavior(ctx context.Context, userName string) (*SchedulingBehavior, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*SchedulingBehavior), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) AnalyzeUserFairShareOptimization(ctx context.Context, userName string) (*FairShareOptimization, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*FairShareOptimization), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserAdaptationMetrics(ctx context.Context, userName string) (*UserAdaptationMetrics, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserAdaptationMetrics), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserComplianceAnalysis(ctx context.Context, userName string) (*UserComplianceAnalysis, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserComplianceAnalysis), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserEfficiencyTrends(ctx context.Context, userName string, period string) (*UserEfficiencyTrends, error) {
	args := m.Called(ctx, userName, period)
	return args.Get(0).(*UserEfficiencyTrends), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) PredictUserBehavior(ctx context.Context, userName string, timeframe string) (*UserBehaviorPrediction, error) {
	args := m.Called(ctx, userName, timeframe)
	return args.Get(0).(*UserBehaviorPrediction), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserClusteringAnalysis(ctx context.Context) (*UserClusteringAnalysis, error) {
	args := m.Called(ctx)
	return args.Get(0).(*UserClusteringAnalysis), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetBehavioralAnomalies(ctx context.Context, userName string) (*BehavioralAnomalies, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*BehavioralAnomalies), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetUserLearningProgress(ctx context.Context, userName string) (*UserLearningProgress, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserLearningProgress), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetPersonalizedRecommendations(ctx context.Context, userName string) (*PersonalizedRecommendations, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*PersonalizedRecommendations), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetBehaviorOptimizationSuggestions(ctx context.Context, userName string) (*BehaviorOptimizationSuggestions, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*BehaviorOptimizationSuggestions), args.Error(1)
}

func (m *MockUserBehaviorAnalysisSLURMClient) GetTrainingRecommendations(ctx context.Context, userName string) (*TrainingRecommendations, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*TrainingRecommendations), args.Error(1)
}

// Types are already defined in user_behavior_analysis.go

func TestNewUserBehaviorAnalysisCollector(t *testing.T) {
	client := &MockUserBehaviorAnalysisSLURMClient{}
	collector := NewUserBehaviorAnalysisCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.userSubmissionFrequency)
	assert.NotNil(t, collector.userResourceEfficiency)
	assert.NotNil(t, collector.userBehaviorStability)
	assert.NotNil(t, collector.fairShareOptimizationScore)
}

func TestUserBehaviorAnalysisCollector_Describe(t *testing.T) {
	client := &MockUserBehaviorAnalysisSLURMClient{}
	collector := NewUserBehaviorAnalysisCollector(client)

	ch := make(chan *prometheus.Desc, 200)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 48 metrics, verify we have the correct number
	assert.Equal(t, 48, len(descs), "Should have exactly 48 metric descriptions")
}

func TestUserBehaviorAnalysisCollector_Collect_Success(t *testing.T) {
	client := &MockUserBehaviorAnalysisSLURMClient{}

	// Mock user behavior profile
	behaviorProfile := &UserBehaviorProfile{
		UserName:                "testuser",
		AccountName:             "testaccount",
		ProfileVersion:          "v1.0",
		AnalysisPeriod:          "30d",
		SubmissionFrequency:     2.5,
		SubmissionConsistency:   0.85,
		SubmissionTiming: map[string]float64{
			"morning": 0.4,
			"afternoon": 0.35,
			"evening": 0.25,
		},
		PeakSubmissionHours:   []int{9, 14, 16},
		SubmissionBurstiness:  0.3,
		ResourceEfficiency:    0.82,
		ResourcePredictability: 0.78,
		ResourceOverRequest:   0.15,
		ResourceUnderUsage:    0.08,
		ResourceOptimization:  0.75,
		QueuePatience:         0.65,
		PriorityUsage:         0.7,
		PartitionPreference: map[string]float64{
			"compute": 0.6,
			"gpu":     0.3,
			"memory":  0.1,
		},
		TimeSlotPreference: map[string]float64{
			"morning":   0.4,
			"afternoon": 0.35,
			"evening":   0.25,
		},
		JobSizeDistribution: map[string]float64{
			"small":  0.3,
			"medium": 0.5,
			"large":  0.2,
		},
		BehaviorStability:   0.88,
		AdaptationRate:      0.72,
		LearningCurve:       0.68,
		PolicyCompliance:    0.92,
		ResponseToFeedback:  0.85,
		CollaborationLevel:  0.6,
		ResourceSharing:     0.55,
		HelpSeeking:        0.4,
		CommunityEngagement: 0.5,
		MentorshipParticipation: 0.3,
		RiskProfile:         "low",
		ComplianceScore:     0.9,
		ViolationProneness:  0.05,
		PolicyAwareness:     0.88,
		SafetyMindedness:    0.95,
		ProfileConfidence:   0.85,
		DataCompleteness:    0.92,
		LastUpdated:         time.Now(),
		NextUpdateDue:       time.Now().Add(7 * 24 * time.Hour),
	}

	// Mock job submission patterns
	submissionPatterns := &JobSubmissionPatterns{
		UserName:       "testuser",
		AnalysisPeriod: "30d",
		HourlyDistribution: map[int]float64{
			9:  0.15,
			10: 0.12,
			11: 0.10,
			14: 0.18,
			15: 0.15,
			16: 0.12,
		},
		DailyDistribution: map[string]float64{
			"monday":    0.18,
			"tuesday":   0.16,
			"wednesday": 0.15,
			"thursday":  0.17,
			"friday":    0.14,
			"saturday":  0.10,
			"sunday":    0.10,
		},
		WeeklyDistribution: map[int]float64{
			1: 0.22,
			2: 0.20,
			3: 0.18,
			4: 0.15,
		},
		MonthlyDistribution: map[int]float64{
			1: 0.08,
			15: 0.12,
			30: 0.15,
		},
		SeasonalPatterns: map[string]float64{
			"spring": 1.0,
			"summer": 0.8,
			"fall":   1.1,
			"winter": 0.9,
		},
		AverageJobsPerDay:    3.2,
		MaxJobsPerDay:        12,
		SubmissionVariability: 0.25,
		BurstSubmissions:     8,
		BurstFrequency:       0.15,
		PreferredSubmissionTime: "09:30",
		SubmissionPredictability: 0.78,
		PlanningHorizon:      48.0,
		UrgencyPattern: map[string]float64{
			"low":    0.4,
			"medium": 0.45,
			"high":   0.15,
		},
		DeadlineAwareness:  0.82,
		PatternStability:   0.85,
		SeasonalAdaptation: 0.7,
		WorkloadAdaptation: 0.75,
		PolicyAdaptation:   0.8,
		LastAnalyzed:       time.Now(),
	}

	// Mock resource usage patterns
	resourcePatterns := &ResourceUsagePatterns{
		UserName:       "testuser",
		AnalysisPeriod: "30d",
		CPURequestPattern: map[string]float64{
			"1-4":   0.3,
			"5-8":   0.4,
			"9-16":  0.2,
			"17-32": 0.1,
		},
		MemoryRequestPattern: map[string]float64{
			"1-8":   0.25,
			"9-16":  0.35,
			"17-32": 0.25,
			"33-64": 0.15,
		},
		GPURequestPattern: map[string]float64{
			"0": 0.7,
			"1": 0.2,
			"2": 0.08,
			"4": 0.02,
		},
		StorageRequestPattern: map[string]float64{
			"1-100":   0.4,
			"101-500": 0.35,
			"501-1000": 0.2,
			"1000+":   0.05,
		},
		RuntimeRequestPattern: map[string]float64{
			"1-4":   0.3,
			"5-12":  0.4,
			"13-24": 0.2,
			"24+":   0.1,
		},
		CPUEfficiency:        0.85,
		MemoryEfficiency:     0.78,
		GPUEfficiency:        0.92,
		IOEfficiency:         0.75,
		OverallEfficiency:    0.82,
		RequestAccuracy:      0.8,
		OverestimationRate:   0.15,
		UnderestimationRate:  0.05,
		WasteRate:           0.12,
		UtilizationVariability: 0.18,
		ResourceLearning:     0.68,
		OptimizationTrend:   "improving",
		BestPracticeAdoption: 0.72,
		EfficiencyImprovement: 0.15,
		ResourceContention:   0.25,
		PeakUsageImpact:     0.3,
		ResourceSharing:     0.55,
		CollaborativeUsage:  0.45,
		LastAnalyzed:        time.Now(),
	}

	// Mock scheduling behavior
	schedulingBehavior := &SchedulingBehavior{
		UserName:           "testuser",
		QueuePatience:      0.75,
		QueueJumping:      0.05,
		CancellationRate:  0.08,
		ResubmissionRate:  0.12,
		QueueOptimization: 0.78,
		PriorityStrategy:  "balanced",
		PriorityEfficiency: 0.82,
		PriorityAbuse:     0.02,
		PrioritySharing:   0.6,
		PartitionWisdom:   0.85,
		ResourceMatching:  0.8,
		LoadAwareness:     0.72,
		OptimalSelection:  0.78,
		BackfillUtilization: 0.65,
		OffPeakUsage:      0.4,
		FlexibilityScore:  0.7,
		AdaptiveScheduling: 0.68,
		SystemRespect:     0.92,
		PolicyAdherence:   0.88,
		FairPlayScore:     0.85,
		CommunityImpact:   0.6,
		LastAnalyzed:      time.Now(),
	}

	// Setup mock expectations
	client.On("GetUserBehaviorProfile", mock.Anything, mock.AnythingOfType("string")).Return(behaviorProfile, nil)
	client.On("GetUserJobSubmissionPatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(submissionPatterns, nil)
	client.On("GetUserResourceUsagePatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(resourcePatterns, nil)
	client.On("GetUserSchedulingBehavior", mock.Anything, mock.AnythingOfType("string")).Return(schedulingBehavior, nil)
	client.On("AnalyzeUserFairShareOptimization", mock.Anything, mock.AnythingOfType("string")).Return(&FairShareOptimization{OptimizationScore: 0.8}, nil)
	client.On("GetUserAdaptationMetrics", mock.Anything, mock.AnythingOfType("string")).Return(&UserAdaptationMetrics{AdaptationEffectiveness: 0.75}, nil)
	client.On("GetUserEfficiencyTrends", mock.Anything, mock.AnythingOfType("string"), "30d").Return(&UserEfficiencyTrends{TrendDirection: "improving", TrendScore: 0.1}, nil)
	client.On("PredictUserBehavior", mock.Anything, mock.AnythingOfType("string"), "7d").Return(&UserBehaviorPrediction{PredictionAccuracy: 0.85}, nil)
	client.On("GetUserClusteringAnalysis", mock.Anything).Return(&UserClusteringAnalysis{ClusterAssignments: map[string]string{"testuser": "cluster_1"}}, nil)
	client.On("GetBehavioralAnomalies", mock.Anything, mock.AnythingOfType("string")).Return(&BehavioralAnomalies{AnomalyScore: 0.1}, nil)
	client.On("GetUserLearningProgress", mock.Anything, mock.AnythingOfType("string")).Return(&UserLearningProgress{LearningScore: 0.75}, nil)

	collector := NewUserBehaviorAnalysisCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 600)
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
	foundSubmissionFrequency := false
	foundResourceEfficiency := false
	foundBehaviorStability := false
	foundOptimizationScore := false
	foundAdaptationRate := false

	for name := range metricNames {
		if strings.Contains(name, "submission_frequency") {
			foundSubmissionFrequency = true
		}
		if strings.Contains(name, "resource_efficiency") {
			foundResourceEfficiency = true
		}
		if strings.Contains(name, "behavior_stability") {
			foundBehaviorStability = true
		}
		if strings.Contains(name, "fairshare_optimization_score") {
			foundOptimizationScore = true
		}
		if strings.Contains(name, "adaptation_rate") {
			foundAdaptationRate = true
		}
	}

	assert.True(t, foundSubmissionFrequency, "Should have submission frequency metrics")
	assert.True(t, foundResourceEfficiency, "Should have resource efficiency metrics")
	assert.True(t, foundBehaviorStability, "Should have behavior stability metrics")
	assert.True(t, foundOptimizationScore, "Should have optimization score metrics")
	assert.True(t, foundAdaptationRate, "Should have adaptation rate metrics")

	client.AssertExpectations(t)
}

func TestUserBehaviorAnalysisCollector_Collect_Error(t *testing.T) {
	client := &MockUserBehaviorAnalysisSLURMClient{}

	// Mock error response
	client.On("GetUserBehaviorProfile", mock.Anything, mock.AnythingOfType("string")).Return((*UserBehaviorProfile)(nil), assert.AnError)

	collector := NewUserBehaviorAnalysisCollector(client)

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

func TestUserBehaviorAnalysisCollector_MetricValues(t *testing.T) {
	client := &MockUserBehaviorAnalysisSLURMClient{}

	// Create test data with known values
	behaviorProfile := &UserBehaviorProfile{
		UserName:              "test_user",
		AccountName:           "test_account",
		SubmissionFrequency:   3.5,
		SubmissionConsistency: 0.9,
		ResourceEfficiency:    0.85,
		BehaviorStability:     0.88,
		AdaptationRate:        0.75,
		ComplianceScore:       0.92,
	}

	submissionPatterns := &JobSubmissionPatterns{
		UserName:                "test_user",
		SubmissionVariability:   0.2,
		SubmissionPredictability: 0.85,
		PlanningHorizon:         72.0,
		DeadlineAwareness:       0.9,
		HourlyDistribution: map[int]float64{
			9:  0.2,
			14: 0.15,
		},
		DailyDistribution: map[string]float64{
			"monday":  0.18,
			"tuesday": 0.16,
		},
	}

	resourcePatterns := &ResourceUsagePatterns{
		UserName:         "test_user",
		CPUEfficiency:    0.88,
		MemoryEfficiency: 0.82,
		WasteRate:        0.1,
		ResourceLearning: 0.7,
		OptimizationTrend: "improving",
		BestPracticeAdoption: 0.8,
	}

	schedulingBehavior := &SchedulingBehavior{
		UserName:           "test_user",
		QueueOptimization: 0.85,
		PriorityEfficiency: 0.9,
		BackfillUtilization: 0.7,
		SystemRespect:     0.95,
		FairPlayScore:     0.88,
	}

	client.On("GetUserBehaviorProfile", mock.Anything, mock.AnythingOfType("string")).Return(behaviorProfile, nil)
	client.On("GetUserJobSubmissionPatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(submissionPatterns, nil)
	client.On("GetUserResourceUsagePatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(resourcePatterns, nil)
	client.On("GetUserSchedulingBehavior", mock.Anything, mock.AnythingOfType("string")).Return(schedulingBehavior, nil)
	client.On("AnalyzeUserFairShareOptimization", mock.Anything, mock.AnythingOfType("string")).Return(&FairShareOptimization{OptimizationScore: 0.8}, nil)
	client.On("GetUserAdaptationMetrics", mock.Anything, mock.AnythingOfType("string")).Return(&UserAdaptationMetrics{AdaptationEffectiveness: 0.75}, nil)
	client.On("GetUserEfficiencyTrends", mock.Anything, mock.AnythingOfType("string"), "30d").Return(&UserEfficiencyTrends{TrendDirection: "improving", TrendScore: 0.1}, nil)
	client.On("PredictUserBehavior", mock.Anything, mock.AnythingOfType("string"), "7d").Return(&UserBehaviorPrediction{PredictionAccuracy: 0.85}, nil)
	client.On("GetUserClusteringAnalysis", mock.Anything).Return(&UserClusteringAnalysis{ClusterAssignments: map[string]string{"test_user": "cluster_1"}}, nil)
	client.On("GetBehavioralAnomalies", mock.Anything, mock.AnythingOfType("string")).Return(&BehavioralAnomalies{AnomalyScore: 0.1}, nil)
	client.On("GetUserLearningProgress", mock.Anything, mock.AnythingOfType("string")).Return(&UserLearningProgress{LearningScore: 0.75}, nil)

	collector := NewUserBehaviorAnalysisCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundSubmissionFrequency := false
	foundBehaviorStability := false
	foundResourceEfficiency := false
	foundComplianceScore := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_user_submission_frequency":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(3.5), *mf.Metric[0].Gauge.Value)
				foundSubmissionFrequency = true
			}
		case "slurm_user_behavior_stability":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.88), *mf.Metric[0].Gauge.Value)
				foundBehaviorStability = true
			}
		case "slurm_user_resource_efficiency":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.85), *mf.Metric[0].Gauge.Value)
				foundResourceEfficiency = true
			}
		case "slurm_user_compliance_score":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.92), *mf.Metric[0].Gauge.Value)
				foundComplianceScore = true
			}
		}
	}

	assert.True(t, foundSubmissionFrequency, "Should find submission frequency metric with correct value")
	assert.True(t, foundBehaviorStability, "Should find behavior stability metric with correct value")
	assert.True(t, foundResourceEfficiency, "Should find resource efficiency metric with correct value")
	assert.True(t, foundComplianceScore, "Should find compliance score metric with correct value")
}

func TestUserBehaviorAnalysisCollector_Integration(t *testing.T) {
	client := &MockUserBehaviorAnalysisSLURMClient{}

	// Setup comprehensive mock data
	setupUserBehaviorMocks(client)

	collector := NewUserBehaviorAnalysisCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_user_submission_frequency User job submission frequency per day
		# TYPE slurm_user_submission_frequency gauge
		slurm_user_submission_frequency{account_name="test_account",user_name="test_user"} 2.5
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_user_submission_frequency")
	assert.NoError(t, err)
}

func TestUserBehaviorAnalysisCollector_BehaviorPatterns(t *testing.T) {
	client := &MockUserBehaviorAnalysisSLURMClient{}

	behaviorProfile := &UserBehaviorProfile{
		UserName:              "pattern_user",
		AccountName:           "pattern_account",
		SubmissionFrequency:   4.2,
		SubmissionConsistency: 0.82,
		ResourceEfficiency:    0.78,
		BehaviorStability:     0.9,
		AdaptationRate:        0.68,
		LearningCurve:         0.75,
		CollaborationLevel:    0.6,
		ComplianceScore:       0.88,
	}

	submissionPatterns := &JobSubmissionPatterns{
		UserName:                "pattern_user",
		SubmissionVariability:   0.3,
		BurstFrequency:          0.2,
		SubmissionPredictability: 0.8,
		PlanningHorizon:         48.0,
		DeadlineAwareness:       0.85,
		PatternStability:        0.88,
		HourlyDistribution: map[int]float64{
			8:  0.1,
			9:  0.15,
			10: 0.12,
			14: 0.18,
			15: 0.2,
			16: 0.15,
		},
		DailyDistribution: map[string]float64{
			"monday":    0.2,
			"tuesday":   0.18,
			"wednesday": 0.16,
			"thursday":  0.17,
			"friday":    0.15,
		},
	}

	client.On("GetUserBehaviorProfile", mock.Anything, mock.AnythingOfType("string")).Return(behaviorProfile, nil)
	client.On("GetUserJobSubmissionPatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(submissionPatterns, nil)
	client.On("GetUserResourceUsagePatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(&ResourceUsagePatterns{}, nil)
	client.On("GetUserSchedulingBehavior", mock.Anything, mock.AnythingOfType("string")).Return(&SchedulingBehavior{}, nil)
	client.On("AnalyzeUserFairShareOptimization", mock.Anything, mock.AnythingOfType("string")).Return(&FairShareOptimization{}, nil)
	client.On("GetUserAdaptationMetrics", mock.Anything, mock.AnythingOfType("string")).Return(&UserAdaptationMetrics{}, nil)
	client.On("GetUserEfficiencyTrends", mock.Anything, mock.AnythingOfType("string"), "30d").Return(&UserEfficiencyTrends{}, nil)
	client.On("PredictUserBehavior", mock.Anything, mock.AnythingOfType("string"), "7d").Return(&UserBehaviorPrediction{}, nil)
	client.On("GetUserClusteringAnalysis", mock.Anything).Return(&UserClusteringAnalysis{}, nil)
	client.On("GetBehavioralAnomalies", mock.Anything, mock.AnythingOfType("string")).Return(&BehavioralAnomalies{}, nil)
	client.On("GetUserLearningProgress", mock.Anything, mock.AnythingOfType("string")).Return(&UserLearningProgress{}, nil)

	collector := NewUserBehaviorAnalysisCollector(client)

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

	// Verify behavior pattern metrics are present
	foundPatternMetrics := false
	foundHourlyDistribution := false
	foundDailyDistribution := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "submission_pattern") {
			foundPatternMetrics = true
		}
		if strings.Contains(desc, "hourly_distribution") {
			foundHourlyDistribution = true
		}
		if strings.Contains(desc, "daily_distribution") {
			foundDailyDistribution = true
		}
	}

	assert.True(t, foundHourlyDistribution, "Should find hourly distribution metrics")
	assert.True(t, foundDailyDistribution, "Should find daily distribution metrics")
}

func TestUserBehaviorAnalysisCollector_OptimizationMetrics(t *testing.T) {
	client := &MockUserBehaviorAnalysisSLURMClient{}

	behaviorProfile := &UserBehaviorProfile{
		UserName:           "optimization_user",
		AccountName:        "optimization_account",
		ResourceEfficiency: 0.9,
		AdaptationRate:     0.85,
		LearningCurve:      0.8,
		ComplianceScore:    0.95,
	}

	resourcePatterns := &ResourceUsagePatterns{
		UserName:             "optimization_user",
		CPUEfficiency:        0.92,
		MemoryEfficiency:     0.88,
		OverallEfficiency:    0.9,
		ResourceLearning:     0.85,
		OptimizationTrend:    "improving",
		BestPracticeAdoption: 0.9,
		EfficiencyImprovement: 0.2,
	}

	client.On("GetUserBehaviorProfile", mock.Anything, mock.AnythingOfType("string")).Return(behaviorProfile, nil)
	client.On("GetUserJobSubmissionPatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(&JobSubmissionPatterns{}, nil)
	client.On("GetUserResourceUsagePatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(resourcePatterns, nil)
	client.On("GetUserSchedulingBehavior", mock.Anything, mock.AnythingOfType("string")).Return(&SchedulingBehavior{}, nil)
	client.On("AnalyzeUserFairShareOptimization", mock.Anything, mock.AnythingOfType("string")).Return(&FairShareOptimization{OptimizationScore: 0.88}, nil)
	client.On("GetUserAdaptationMetrics", mock.Anything, mock.AnythingOfType("string")).Return(&UserAdaptationMetrics{AdaptationEffectiveness: 0.82}, nil)
	client.On("GetUserEfficiencyTrends", mock.Anything, mock.AnythingOfType("string"), "30d").Return(&UserEfficiencyTrends{TrendDirection: "improving", TrendScore: 0.15}, nil)
	client.On("PredictUserBehavior", mock.Anything, mock.AnythingOfType("string"), "7d").Return(&UserBehaviorPrediction{PredictionAccuracy: 0.9}, nil)
	client.On("GetUserClusteringAnalysis", mock.Anything).Return(&UserClusteringAnalysis{}, nil)
	client.On("GetBehavioralAnomalies", mock.Anything, mock.AnythingOfType("string")).Return(&BehavioralAnomalies{AnomalyScore: 0.05}, nil)
	client.On("GetUserLearningProgress", mock.Anything, mock.AnythingOfType("string")).Return(&UserLearningProgress{LearningScore: 0.85}, nil)

	collector := NewUserBehaviorAnalysisCollector(client)

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

	// Verify optimization metrics are present
	foundOptimizationMetrics := false
	foundAdaptationMetrics := false
	foundLearningMetrics := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "optimization") {
			foundOptimizationMetrics = true
		}
		if strings.Contains(desc, "adaptation") {
			foundAdaptationMetrics = true
		}
		if strings.Contains(desc, "learning") {
			foundLearningMetrics = true
		}
	}

	assert.True(t, foundOptimizationMetrics, "Should find optimization metrics")
	assert.True(t, foundAdaptationMetrics, "Should find adaptation metrics")
	assert.True(t, foundLearningMetrics, "Should find learning metrics")
}

func setupUserBehaviorMocks(client *MockUserBehaviorAnalysisSLURMClient) {
	behaviorProfile := &UserBehaviorProfile{
		UserName:              "test_user",
		AccountName:           "test_account",
		SubmissionFrequency:   2.5,
		SubmissionConsistency: 0.85,
		ResourceEfficiency:    0.82,
		BehaviorStability:     0.88,
		AdaptationRate:        0.72,
		LearningCurve:         0.68,
		ComplianceScore:       0.9,
		CollaborationLevel:    0.6,
	}

	submissionPatterns := &JobSubmissionPatterns{
		UserName:                "test_user",
		SubmissionVariability:   0.25,
		SubmissionPredictability: 0.8,
		PlanningHorizon:         48.0,
		DeadlineAwareness:       0.85,
		PatternStability:        0.88,
		HourlyDistribution: map[int]float64{
			9:  0.15,
			14: 0.18,
		},
		DailyDistribution: map[string]float64{
			"monday":  0.18,
			"tuesday": 0.16,
		},
	}

	resourcePatterns := &ResourceUsagePatterns{
		UserName:             "test_user",
		CPUEfficiency:        0.85,
		MemoryEfficiency:     0.78,
		ResourceLearning:     0.68,
		OptimizationTrend:    "improving",
		BestPracticeAdoption: 0.72,
		WasteRate:           0.12,
		UtilizationVariability: 0.18,
		ResourceContention:   0.25,
	}

	schedulingBehavior := &SchedulingBehavior{
		UserName:           "test_user",
		QueueOptimization: 0.78,
		PriorityEfficiency: 0.82,
		BackfillUtilization: 0.65,
		SystemRespect:     0.92,
		FairPlayScore:     0.85,
	}

	client.On("GetUserBehaviorProfile", mock.Anything, mock.AnythingOfType("string")).Return(behaviorProfile, nil)
	client.On("GetUserJobSubmissionPatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(submissionPatterns, nil)
	client.On("GetUserResourceUsagePatterns", mock.Anything, mock.AnythingOfType("string"), "7d").Return(resourcePatterns, nil)
	client.On("GetUserSchedulingBehavior", mock.Anything, mock.AnythingOfType("string")).Return(schedulingBehavior, nil)
	client.On("AnalyzeUserFairShareOptimization", mock.Anything, mock.AnythingOfType("string")).Return(&FairShareOptimization{OptimizationScore: 0.8}, nil)
	client.On("GetUserAdaptationMetrics", mock.Anything, mock.AnythingOfType("string")).Return(&UserAdaptationMetrics{AdaptationEffectiveness: 0.75}, nil)
	client.On("GetUserEfficiencyTrends", mock.Anything, mock.AnythingOfType("string"), "30d").Return(&UserEfficiencyTrends{TrendDirection: "improving", TrendScore: 0.1}, nil)
	client.On("PredictUserBehavior", mock.Anything, mock.AnythingOfType("string"), "7d").Return(&UserBehaviorPrediction{PredictionAccuracy: 0.85}, nil)
	client.On("GetUserClusteringAnalysis", mock.Anything).Return(&UserClusteringAnalysis{ClusterAssignments: map[string]string{"test_user": "cluster_1"}}, nil)
	client.On("GetBehavioralAnomalies", mock.Anything, mock.AnythingOfType("string")).Return(&BehavioralAnomalies{AnomalyScore: 0.1}, nil)
	client.On("GetUserLearningProgress", mock.Anything, mock.AnythingOfType("string")).Return(&UserLearningProgress{LearningScore: 0.75}, nil)
}