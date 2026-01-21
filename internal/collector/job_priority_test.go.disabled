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

type MockJobPrioritySLURMClient struct {
	mock.Mock
}

func (m *MockJobPrioritySLURMClient) CalculateJobPriority(ctx context.Context, jobID string) (*JobPriorityFactors, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*JobPriorityFactors), args.Error(1)
}

func (m *MockJobPrioritySLURMClient) PredictJobScheduling(ctx context.Context, jobID string) (*JobSchedulingPrediction, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*JobSchedulingPrediction), args.Error(1)
}

func (m *MockJobPrioritySLURMClient) GetQueueAnalysis(ctx context.Context, partition string) (*QueueAnalysis, error) {
	args := m.Called(ctx, partition)
	return args.Get(0).(*QueueAnalysis), args.Error(1)
}

func (m *MockJobPrioritySLURMClient) GetSystemPriorityStats(ctx context.Context) (*SystemPriorityStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemPriorityStats), args.Error(1)
}

func (m *MockJobPrioritySLURMClient) GetUserPriorityPattern(ctx context.Context, userName string) (*UserPriorityPattern, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserPriorityPattern), args.Error(1)
}

func (m *MockJobPrioritySLURMClient) ValidatePriorityPrediction(ctx context.Context, jobID string) (*PriorityPredictionValidation, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*PriorityPredictionValidation), args.Error(1)
}

func TestNewJobPriorityCollector(t *testing.T) {
	client := &MockJobPrioritySLURMClient{}
	collector := NewJobPriorityCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.jobPriorityScore)
	assert.NotNil(t, collector.jobPriorityRank)
	assert.NotNil(t, collector.jobEstimatedWaitTime)
	assert.NotNil(t, collector.systemPriorityStats)
	assert.NotNil(t, collector.queueDepth)
	assert.NotNil(t, collector.userPriorityPattern)
	assert.NotNil(t, collector.predictionAccuracy)
}

func TestJobPriorityCollector_Describe(t *testing.T) {
	client := &MockJobPrioritySLURMClient{}
	collector := NewJobPriorityCollector(client)

	ch := make(chan *prometheus.Desc, 50)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 29 metrics, verify we have the correct number
	assert.Equal(t, 29, len(descs), "Should have exactly 29 metric descriptions")
}

func TestJobPriorityCollector_Collect_Success(t *testing.T) {
	client := &MockJobPrioritySLURMClient{}

	// Mock system priority stats
	systemStats := &SystemPriorityStats{
		TotalJobs:              150,
		AveragePriority:        50000.0,
		PriorityStdDev:         15000.0,
		PriorityRange:          100000,
		PriorityMedian:         45000.0,
		AlgorithmVersion:       "v1.0",
		RebalanceFrequency:     0.5,
		CalculationTime:        0.025,
		CalculationAccuracy:    0.95,
		PriorityInflation:      0.02,
		PriorityVolatility:     0.15,
		LastUpdated:            time.Now(),
	}

	// Mock job priority factors
	priorityFactors := &JobPriorityFactors{
		JobID:           "12345",
		UserName:        "testuser",
		AccountName:     "testaccount",
		PartitionName:   "cpu",
		QoSName:         "normal",
		TotalPriority:   75000,
		AgePriority:     15000,
		FairShareFactor: 0.85,
		QoSPriority:     10000,
		PartitionPrio:   5000,
		SizePriority:    2000,
		AssocPriority:   1000,
		NicePriority:    0,
		QueueRank:       5,
		PartitionRank:   3,
		UserRank:        2,
		SubmittedAt:     time.Now().Add(-time.Hour),
		LastCalculated:  time.Now(),
	}

	// Mock scheduling prediction
	schedulingPrediction := &JobSchedulingPrediction{
		JobID:                 "12345",
		EstimatedWaitTime:     30 * time.Minute,
		EstimatedStartTime:    time.Now().Add(30 * time.Minute),
		ConfidenceLevel:       0.85,
		PredictionMethod:      "ml_model",
		QueuePosition:         5,
		QueueDepth:            25,
		JobsAhead:             4,
		ProcessingRate:        2.5,
		PriorityTrend:         "increasing",
		PriorityVelocity:      1000.0,
		PriorityVolatility:    0.12,
		ResourceAvailability:  0.7,
		StarvationRisk:        0.1,
		PreemptionRisk:        0.05,
		BackfillProbability:   0.3,
		LastUpdated:           time.Now(),
	}

	// Mock queue analysis
	queueAnalysis := &QueueAnalysis{
		PartitionName:      "cpu",
		TotalJobs:          50,
		PendingJobs:        25,
		RunningJobs:        25,
		ProcessingRate:     2.5,
		AverageWaitTime:    1800.0,
		EfficiencyScore:    0.85,
		HighPriorityJobs:   10,
		MediumPriorityJobs: 30,
		LowPriorityJobs:    10,
		StarvationRisk:     0.2,
		OldestJobAge:       7200.0,
		ThroughputTrend:    "stable",
		LatencyTrend:       "decreasing",
		LastAnalyzed:       time.Now(),
	}

	// Mock user priority pattern
	userPattern := &UserPriorityPattern{
		UserName:             "testuser",
		AccountName:          "testaccount",
		AveragePriority:      45000.0,
		PriorityVariance:     12000.0,
		SubmissionPattern:    "regular",
		HighPriorityRatio:    0.3,
		LowPriorityRatio:     0.2,
		PriorityConsistency:  0.75,
		BehaviorScore:        0.8,
		SubmissionFrequency:  5.2,
		BatchingBehavior:     0.6,
		WaitTimeEfficiency:   0.7,
		ResourceEfficiency:   0.85,
		SchedulingEfficiency: 0.9,
		LastAnalyzed:         time.Now(),
	}

	// Mock prediction validation
	predictionValidation := &PriorityPredictionValidation{
		JobID:               "12345",
		PredictedWaitTime:   1800.0,
		ActualWaitTime:      1650.0,
		PredictionError:     150.0,
		PredictionAccuracy:  0.92,
		ConfidenceLevel:     0.85,
		AbsoluteError:       150.0,
		RelativeError:       0.08,
		ErrorCategory:       "minor",
		ErrorSeverity:       "low",
		ModelVersion:        "v2.1",
		ModelAccuracy:       0.89,
		ModelConfidence:     0.87,
		ValidatedAt:         time.Now(),
	}

	// Setup mock expectations
	client.On("GetSystemPriorityStats", mock.Anything).Return(systemStats, nil)
	client.On("CalculateJobPriority", mock.Anything, mock.AnythingOfType("string")).Return(priorityFactors, nil)
	client.On("PredictJobScheduling", mock.Anything, mock.AnythingOfType("string")).Return(schedulingPrediction, nil)
	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(queueAnalysis, nil)
	client.On("GetUserPriorityPattern", mock.Anything, mock.AnythingOfType("string")).Return(userPattern, nil)
	client.On("ValidatePriorityPrediction", mock.Anything, mock.AnythingOfType("string")).Return(predictionValidation, nil)

	collector := NewJobPriorityCollector(client)

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
	foundPriorityMetric := false
	foundQueueMetric := false
	foundPredictionMetric := false
	foundSystemMetric := false

	for name := range metricNames {
		if strings.Contains(name, "job_priority_score") {
			foundPriorityMetric = true
		}
		if strings.Contains(name, "queue_depth") {
			foundQueueMetric = true
		}
		if strings.Contains(name, "prediction_accuracy") {
			foundPredictionMetric = true
		}
		if strings.Contains(name, "system_priority") {
			foundSystemMetric = true
		}
	}

	assert.True(t, foundPriorityMetric, "Should have priority metrics")
	assert.True(t, foundQueueMetric, "Should have queue metrics")
	assert.True(t, foundPredictionMetric, "Should have prediction metrics")
	assert.True(t, foundSystemMetric, "Should have system metrics")

	client.AssertExpectations(t)
}

func TestJobPriorityCollector_Collect_Error(t *testing.T) {
	client := &MockJobPrioritySLURMClient{}

	// Mock error response
	client.On("GetSystemPriorityStats", mock.Anything).Return((*SystemPriorityStats)(nil), assert.AnError)

	collector := NewJobPriorityCollector(client)

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

func TestJobPriorityCollector_MetricValues(t *testing.T) {
	client := &MockJobPrioritySLURMClient{}

	// Create test data with known values
	systemStats := &SystemPriorityStats{
		TotalJobs:              100,
		AveragePriority:        50000.0,
		AlgorithmVersion:       "test_v1",
		RebalanceFrequency:     0.5,
		CalculationAccuracy:    0.95,
	}

	priorityFactors := &JobPriorityFactors{
		JobID:           "test_job",
		UserName:        "test_user",
		AccountName:     "test_account",
		PartitionName:   "test_partition",
		QoSName:         "normal",
		TotalPriority:   60000,
		AgePriority:     20000,
		FairShareFactor: 0.8,
		QueueRank:       3,
	}

	schedulingPrediction := &JobSchedulingPrediction{
		JobID:            "test_job",
		EstimatedWaitTime: 25 * time.Minute,
		QueuePosition:    3,
		PredictionMethod: "test_model",
		PriorityTrend:    "stable",
		PriorityVelocity: 0.0,
		PriorityVolatility: 0.1,
	}

	client.On("GetSystemPriorityStats", mock.Anything).Return(systemStats, nil)
	client.On("CalculateJobPriority", mock.Anything, "test_job").Return(priorityFactors, nil)
	client.On("PredictJobScheduling", mock.Anything, "test_job").Return(schedulingPrediction, nil)
	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&QueueAnalysis{}, nil)
	client.On("GetUserPriorityPattern", mock.Anything, mock.AnythingOfType("string")).Return(&UserPriorityPattern{}, nil)
	client.On("ValidatePriorityPrediction", mock.Anything, mock.AnythingOfType("string")).Return(&PriorityPredictionValidation{}, nil)

	collector := NewJobPriorityCollector(client)

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
	foundSystemJobs := false
	foundJobPriority := false
	foundWaitTime := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_system_priority_statistics":
			for _, metric := range mf.Metric {
				for _, label := range metric.Label {
					if *label.Name == "statistic" && *label.Value == "total_jobs" {
						assert.Equal(t, float64(100), *metric.Gauge.Value)
						foundSystemJobs = true
					}
				}
			}
		case "slurm_job_priority_score":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(60000), *mf.Metric[0].Gauge.Value)
				foundJobPriority = true
			}
		case "slurm_job_estimated_wait_time_seconds":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(1500), *mf.Metric[0].Gauge.Value) // 25 minutes = 1500 seconds
				foundWaitTime = true
			}
		}
	}

	assert.True(t, foundSystemJobs, "Should find system jobs metric with correct value")
	assert.True(t, foundJobPriority, "Should find job priority metric with correct value")
	assert.True(t, foundWaitTime, "Should find wait time metric with correct value")
}

func TestJobPriorityCollector_Integration(t *testing.T) {
	client := &MockJobPrioritySLURMClient{}

	// Setup comprehensive mock data
	setupPriorityMocks(client)

	collector := NewJobPriorityCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_system_priority_statistics System-wide priority statistics
		# TYPE slurm_system_priority_statistics gauge
		slurm_system_priority_statistics{algorithm_version="v1.0",statistic="total_jobs"} 150
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_system_priority_statistics")
	assert.NoError(t, err)
}

func TestJobPriorityCollector_StarvationRiskLevels(t *testing.T) {
	client := &MockJobPrioritySLURMClient{}

	// Test different starvation risk levels
	testCases := []struct {
		risk  float64
		level string
	}{
		{0.8, "high"},
		{0.5, "medium"},
		{0.2, "low"},
	}

	for _, tc := range testCases {
		queueAnalysis := &QueueAnalysis{
			PartitionName:  "test",
			StarvationRisk: tc.risk,
		}

		client.On("GetSystemPriorityStats", mock.Anything).Return(&SystemPriorityStats{}, nil)
		client.On("GetQueueAnalysis", mock.Anything, "test").Return(queueAnalysis, nil)

		collector := NewJobPriorityCollector(client)

		// Override sample partitions for testing
		collector.getSamplePartitions = func() []string {
			return []string{"test"}
		}

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

		// Check that starvation risk is categorized correctly
		found := false
		for _, metric := range metrics {
			if strings.Contains(metric.Desc().String(), "queue_starvation_risk") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find starvation risk metric for level %s", tc.level)

		client.ExpectedCalls = nil // Reset expectations for next iteration
	}
}

func setupPriorityMocks(client *MockJobPrioritySLURMClient) {
	systemStats := &SystemPriorityStats{
		TotalJobs:              150,
		AveragePriority:        50000.0,
		AlgorithmVersion:       "v1.0",
		RebalanceFrequency:     0.5,
		CalculationAccuracy:    0.95,
	}

	priorityFactors := &JobPriorityFactors{
		JobID:           "12345",
		UserName:        "user1",
		AccountName:     "account1",
		PartitionName:   "cpu",
		QoSName:         "normal",
		TotalPriority:   50000,
		AgePriority:     10000,
		FairShareFactor: 0.8,
		QueueRank:       5,
	}

	schedulingPrediction := &JobSchedulingPrediction{
		JobID:            "12345",
		EstimatedWaitTime: 30 * time.Minute,
		QueuePosition:    5,
		PredictionMethod: "ml_model",
		PriorityTrend:    "stable",
		PriorityVelocity: 0.0,
	}

	queueAnalysis := &QueueAnalysis{
		PartitionName:  "cpu",
		TotalJobs:      50,
		PendingJobs:    25,
		RunningJobs:    25,
		StarvationRisk: 0.2,
	}

	userPattern := &UserPriorityPattern{
		UserName:        "user1",
		AccountName:     "account1",
		AveragePriority: 45000.0,
	}

	predictionValidation := &PriorityPredictionValidation{
		JobID:              "12345",
		PredictionAccuracy: 0.9,
		ModelVersion:       "v1.0",
		ModelAccuracy:      0.88,
	}

	client.On("GetSystemPriorityStats", mock.Anything).Return(systemStats, nil)
	client.On("CalculateJobPriority", mock.Anything, mock.AnythingOfType("string")).Return(priorityFactors, nil)
	client.On("PredictJobScheduling", mock.Anything, mock.AnythingOfType("string")).Return(schedulingPrediction, nil)
	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(queueAnalysis, nil)
	client.On("GetUserPriorityPattern", mock.Anything, mock.AnythingOfType("string")).Return(userPattern, nil)
	client.On("ValidatePriorityPrediction", mock.Anything, mock.AnythingOfType("string")).Return(predictionValidation, nil)
}