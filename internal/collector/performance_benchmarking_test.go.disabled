package collector

import (
	"context"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockBenchmarkingSlurmClient for testing
type MockBenchmarkingSlurmClient struct {
	mock.Mock
}

func (m *MockBenchmarkingSlurmClient) Jobs() slurm.JobManager {
	args := m.Called()
	return args.Get(0).(slurm.JobManager)
}

func (m *MockBenchmarkingSlurmClient) Nodes() slurm.NodeManager {
	args := m.Called()
	return args.Get(0).(slurm.NodeManager)
}

func (m *MockBenchmarkingSlurmClient) Partitions() slurm.PartitionManager {
	args := m.Called()
	return args.Get(0).(slurm.PartitionManager)
}

func (m *MockBenchmarkingSlurmClient) Info() slurm.InfoManager {
	args := m.Called()
	return args.Get(0).(slurm.InfoManager)
}

// Mock slurm types that are not yet available in the client
type NodeListOption interface{}
type NodeGetOption interface{}

// MockBenchmarkingJobManager for testing
type MockBenchmarkingJobManager struct {
	mock.Mock
}

func (m *MockBenchmarkingJobManager) List(ctx context.Context, opts ...JobListOption) (*slurm.JobList, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*slurm.JobList), args.Error(1)
}

func (m *MockBenchmarkingJobManager) Get(ctx context.Context, jobID string, opts ...JobGetOption) (*slurm.Job, error) {
	args := m.Called(ctx, jobID, opts)
	return args.Get(0).(*slurm.Job), args.Error(1)
}

func (m *MockBenchmarkingJobManager) Submit(ctx context.Context, job *JobSubmissionRequest) (*JobSubmissionResponse, error) {
	args := m.Called(ctx, job)
	return args.Get(0).(*slurm.JobSubmissionResponse), args.Error(1)
}

func (m *MockBenchmarkingJobManager) Cancel(ctx context.Context, jobID string, opts ...JobCancelOption) error {
	args := m.Called(ctx, jobID, opts)
	return args.Error(0)
}

// MockBenchmarkingNodeManager for testing
type MockBenchmarkingNodeManager struct {
	mock.Mock
}

func (m *MockBenchmarkingNodeManager) List(ctx context.Context, opts ...NodeListOption) (*slurm.NodeList, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*slurm.NodeList), args.Error(1)
}

func (m *MockBenchmarkingNodeManager) Get(ctx context.Context, nodeName string, opts ...NodeGetOption) (*slurm.Node, error) {
	args := m.Called(ctx, nodeName, opts)
	return args.Get(0).(*slurm.Node), args.Error(1)
}

func TestNewPerformanceBenchmarkingCollector(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval:      60 * time.Second,
		BenchmarkRetention:      24 * time.Hour,
		EnableJobComparison:     true,
		EnableUserComparison:    true,
		EnableNodeComparison:    true,
		EnableClusterComparison: true,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)

	require.NoError(t, err)
	assert.NotNil(t, collector)
	assert.Equal(t, mockClient, collector.slurmClient)
	assert.Equal(t, config, collector.config)
	assert.NotNil(t, collector.metrics)
	assert.NotNil(t, collector.benchmarkData)
	assert.NotNil(t, collector.historicalData)
	assert.NotNil(t, collector.comparisonResults)
	assert.NotNil(t, collector.baselineManager)
	assert.NotNil(t, collector.jobComparator)
	assert.NotNil(t, collector.userComparator)
	assert.NotNil(t, collector.nodeComparator)
	assert.NotNil(t, collector.clusterComparator)
	assert.NotNil(t, collector.statAnalyzer)
	assert.NotNil(t, collector.trendAnalyzer)
}

func TestNewPerformanceBenchmarkingCollector_DefaultConfig(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, nil)

	require.NoError(t, err)
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.config)
	assert.Equal(t, 60*time.Second, collector.config.CollectionInterval)
	assert.True(t, collector.config.EnableJobComparison)
	assert.True(t, collector.config.EnableUserComparison)
	assert.True(t, collector.config.EnableNodeComparison)
	assert.True(t, collector.config.EnableClusterComparison)
}

func TestPerformanceBenchmarkingCollector_Describe(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 50)
	collector.Describe(ch)
	close(ch)

	// Count the number of metrics described
	count := 0
	for range ch {
		count++
	}

	// Should have multiple metrics described
	assert.Greater(t, count, 20)
}

func TestPerformanceBenchmarkingCollector_Collect(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	mockJobManager := &MockBenchmarkingJobManager{}
	mockNodeManager := &MockBenchmarkingNodeManager{}

	// Create test data
	testJobs := &slurm.JobList{
		Jobs: []*slurm.Job{
			{
				JobID:     "12345",
				JobState:  "COMPLETED",
				UserName:  "user1",
				Account:   "account1",
				Partition: "partition1",
				StartTime: time.Now().Add(-time.Hour),
				EndTime:   time.Now().Add(-30 * time.Minute),
				CPUs:      8,
				Memory:    "16GB",
			},
			{
				JobID:     "12346",
				JobState:  "RUNNING",
				UserName:  "user2",
				Account:   "account2",
				Partition: "partition1",
				StartTime: time.Now().Add(-30 * time.Minute),
				CPUs:      4,
				Memory:    "8GB",
			},
		},
	}

	testNodes := &slurm.NodeList{
		Nodes: []*slurm.Node{
			{
				Name:       "node001",
				State:      "IDLE",
				Partitions: []string{"partition1"},
				CPUTotal:   16,
				MemoryTotal: 32768,
			},
			{
				Name:       "node002",
				State:      "ALLOCATED",
				Partitions: []string{"partition1"},
				CPUTotal:   16,
				MemoryTotal: 32768,
			},
		},
	}

	mockClient.On("Jobs").Return(mockJobManager)
	mockClient.On("Nodes").Return(mockNodeManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(testNodes, nil)

	config := &BenchmarkingConfig{
		CollectionInterval:   60 * time.Second,
		MaxJobsPerBenchmark:  100,
		MaxNodesPerBenchmark: 50,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Collect metrics
	ch := make(chan prometheus.Metric, 200)
	collector.Collect(ch)
	close(ch)

	// Count collected metrics
	metricCount := 0
	for range ch {
		metricCount++
	}

	// Should collect multiple metrics
	assert.Greater(t, metricCount, 0)

	mockClient.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestPerformanceBenchmarkingCollector_createJobPerformanceSnapshot(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	job := &slurm.Job{
		JobID:     "12345",
		UserName:  "testuser",
		Account:   "testaccount",
		Partition: "testpartition",
		JobState:  "COMPLETED",
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(-30 * time.Minute),
		CPUs:      8,
		Memory:    "16GB",
	}

	snapshot := collector.createJobPerformanceSnapshot(job)

	assert.NotNil(t, snapshot)
	assert.Equal(t, "12345", snapshot.JobID)
	assert.Equal(t, "testuser", snapshot.JobMetadata.UserName)
	assert.Equal(t, "testaccount", snapshot.JobMetadata.Account)
	assert.Equal(t, "testpartition", snapshot.JobMetadata.Partition)
	assert.Equal(t, 8, snapshot.JobMetadata.JobSize)
	assert.NotEmpty(t, snapshot.PerformanceData)
	assert.NotEmpty(t, snapshot.ResourceData)
	assert.NotEmpty(t, snapshot.EfficiencyData)
	assert.Greater(t, snapshot.DataCompleteness, 0.0)
	assert.Greater(t, snapshot.MeasurementQuality, 0.0)
}

func TestPerformanceBenchmarkingCollector_createNodePerformanceProfile(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	node := &slurm.Node{
		Name:       "node001",
		State:      "IDLE",
		Partitions: []string{"partition1", "partition2"},
		CPUTotal:   16,
		MemoryTotal: 32768,
	}

	profile := collector.createNodePerformanceProfile(node)

	assert.NotNil(t, profile)
	assert.Equal(t, "node001", profile.NodeName)
	assert.Equal(t, "partition1", profile.Partition) // Should take first partition
	assert.NotEmpty(t, profile.ThroughputMetrics)
	assert.NotEmpty(t, profile.UtilizationMetrics)
	assert.Greater(t, profile.UptimePercentage, 0.0)
	assert.Equal(t, 1.0, profile.RelativePerformance)
	assert.Equal(t, 1, profile.PerformanceRank)
	assert.Equal(t, "A", profile.PerformanceGrade)
}

func TestPerformanceBenchmarkingCollector_createClusterPerformanceSnapshot(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	ctx := context.Background()
	snapshot := collector.createClusterPerformanceSnapshot(ctx)

	assert.NotNil(t, snapshot)
	assert.Equal(t, "default", snapshot.ClusterName)
	assert.Greater(t, snapshot.TotalThroughput, 0.0)
	assert.Greater(t, snapshot.AverageLatency, 0.0)
	assert.Greater(t, snapshot.OverallEfficiency, 0.0)
	assert.NotEmpty(t, snapshot.ResourceUtilization)
	assert.Greater(t, snapshot.SystemUptime, 0.0)
	assert.Greater(t, snapshot.DataQuality, 0.0)
	assert.Equal(t, time.Hour, snapshot.MeasurementPeriod)
}

func TestPerformanceBenchmarkingCollector_calculatePerformanceScores(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	// Test job performance score calculation
	jobSnapshot := &JobPerformanceSnapshot{
		JobID:          "12345",
		PerformanceData: map[string]float64{"throughput": 5.0},
		ResourceData:    map[string]float64{"cpu_utilization": 0.8},
	}
	jobScore := collector.calculateJobPerformanceScore(jobSnapshot)
	assert.GreaterOrEqual(t, jobScore, 0.0)
	assert.LessOrEqual(t, jobScore, 100.0)

	// Test user performance score calculation
	userProfile := &UserPerformanceProfile{
		UserName:        "testuser",
		AverageMetrics:  map[string]float64{"efficiency": 0.85},
		ConsistencyScore: 0.9,
	}
	userScore := collector.calculateUserPerformanceScore(userProfile)
	assert.GreaterOrEqual(t, userScore, 0.0)
	assert.LessOrEqual(t, userScore, 100.0)

	// Test node performance score calculation
	nodeProfile := &NodePerformanceProfile{
		NodeName:            "node001",
		UptimePercentage:    99.5,
		RelativePerformance: 1.2,
	}
	nodeScore := collector.calculateNodePerformanceScore(nodeProfile)
	assert.GreaterOrEqual(t, nodeScore, 0.0)
	assert.LessOrEqual(t, nodeScore, 100.0)

	// Test cluster performance score calculation
	clusterSnapshot := &ClusterPerformanceSnapshot{
		ClusterName:       "default",
		TotalThroughput:   120.0,
		OverallEfficiency: 0.85,
		SystemUptime:      99.9,
	}
	clusterScore := collector.calculateClusterPerformanceScore(clusterSnapshot)
	assert.GreaterOrEqual(t, clusterScore, 0.0)
	assert.LessOrEqual(t, clusterScore, 100.0)
}

func TestBenchmarkingConfig_DefaultValues(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, nil)
	require.NoError(t, err)

	config := collector.config

	// Test default configuration values
	assert.Equal(t, 60*time.Second, config.CollectionInterval)
	assert.Equal(t, 7*24*time.Hour, config.BenchmarkRetention)
	assert.Equal(t, 30*24*time.Hour, config.HistoricalDataRetention)
	assert.Equal(t, 24*time.Hour, config.BaselineCalculationInterval)
	assert.Equal(t, 100, config.BaselineMinSamples)
	assert.Equal(t, 0.95, config.BaselineConfidenceLevel)
	assert.True(t, config.EnableJobComparison)
	assert.True(t, config.EnableUserComparison)
	assert.True(t, config.EnableNodeComparison)
	assert.True(t, config.EnableClusterComparison)
	assert.True(t, config.EnableStatisticalAnalysis)
	assert.Equal(t, "modified_zscore", config.OutlierDetectionMethod)
	assert.Equal(t, 3.5, config.OutlierThreshold)
	assert.Contains(t, config.ConfidenceIntervals, 0.95)
	assert.Contains(t, config.ConfidenceIntervals, 0.99)
	assert.Contains(t, config.BenchmarkCategories, "cpu_intensive")
	assert.Contains(t, config.BenchmarkCategories, "memory_intensive")
	assert.Contains(t, config.PerformanceMetrics, "throughput")
	assert.Contains(t, config.PerformanceMetrics, "latency")
	assert.Contains(t, config.ComparisonDimensions, "temporal")
	assert.Contains(t, config.ComparisonDimensions, "user")
	assert.Equal(t, 1000, config.MaxJobsPerBenchmark)
	assert.Equal(t, 100, config.MaxNodesPerBenchmark)
	assert.True(t, config.EnableParallelProcessing)
	assert.Equal(t, 5, config.MaxConcurrentComparisons)
	assert.True(t, config.GenerateReports)
	assert.Equal(t, 24*time.Hour, config.ReportGenerationInterval)
	assert.Equal(t, 7*24*time.Hour, config.ReportRetention)
}

func TestBenchmarkingConfig_ComparisonThresholds(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, nil)
	require.NoError(t, err)

	thresholds := collector.config.ComparisonThresholds

	assert.Equal(t, 0.1, thresholds.PerformanceDegradeThreshold)
	assert.Equal(t, 0.1, thresholds.PerformanceImproveThreshold)
	assert.Equal(t, 0.2, thresholds.EfficiencyVarianceThreshold)
	assert.Equal(t, 0.15, thresholds.ResourceWasteThreshold)
	assert.Equal(t, 2.0, thresholds.AnomalyThreshold)
}

func TestPerformanceBenchmarkingCollector_collectJobBenchmarks(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	mockJobManager := &MockBenchmarkingJobManager{}

	testJobs := &slurm.JobList{
		Jobs: []*slurm.Job{
			{
				JobID:     "12345",
				JobState:  "COMPLETED",
				UserName:  "user1",
				Account:   "account1",
				Partition: "partition1",
				StartTime: time.Now().Add(-time.Hour),
				EndTime:   time.Now().Add(-30 * time.Minute),
				CPUs:      8,
				Memory:    "16GB",
			},
		},
	}

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)

	config := &BenchmarkingConfig{
		CollectionInterval:  60 * time.Second,
		MaxJobsPerBenchmark: 100,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	ctx := context.Background()
	err = collector.collectJobBenchmarks(ctx)

	assert.NoError(t, err)
	assert.Len(t, collector.jobComparator.jobMetrics, 1)
	assert.Contains(t, collector.jobComparator.jobMetrics, "12345")

	mockClient.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
}

func TestPerformanceBenchmarkingCollector_collectNodeBenchmarks(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	mockNodeManager := &MockBenchmarkingNodeManager{}

	testNodes := &slurm.NodeList{
		Nodes: []*slurm.Node{
			{
				Name:       "node001",
				State:      "IDLE",
				Partitions: []string{"partition1"},
				CPUTotal:   16,
				MemoryTotal: 32768,
			},
		},
	}

	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(testNodes, nil)

	config := &BenchmarkingConfig{
		CollectionInterval:   60 * time.Second,
		MaxNodesPerBenchmark: 50,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	ctx := context.Background()
	err = collector.collectNodeBenchmarks(ctx)

	assert.NoError(t, err)
	assert.Len(t, collector.nodeComparator.nodeMetrics, 1)
	assert.Contains(t, collector.nodeComparator.nodeMetrics, "node001")

	mockClient.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestPerformanceBenchmarkingCollector_collectClusterBenchmarks(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	ctx := context.Background()
	err = collector.collectClusterBenchmarks(ctx)

	assert.NoError(t, err)
	assert.Len(t, collector.clusterComparator.clusterMetrics, 1)
	assert.Contains(t, collector.clusterComparator.clusterMetrics, "default")
}

func TestPerformanceBenchmarkingCollector_updateUserMetricsWithJob(t *testing.T) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	profile := &UserPerformanceProfile{
		UserName:       "testuser",
		AverageMetrics: make(map[string]float64),
	}

	snapshot := &JobPerformanceSnapshot{
		PerformanceData: map[string]float64{
			"throughput": 5.0,
			"latency":    0.1,
		},
	}

	collector.updateUserMetricsWithJob(profile, snapshot)

	assert.Equal(t, 5.0, profile.AverageMetrics["throughput"])
	assert.Equal(t, 0.1, profile.AverageMetrics["latency"])

	// Test averaging with second job
	snapshot2 := &JobPerformanceSnapshot{
		PerformanceData: map[string]float64{
			"throughput": 7.0,
			"latency":    0.2,
		},
	}

	collector.updateUserMetricsWithJob(profile, snapshot2)

	assert.Equal(t, 6.0, profile.AverageMetrics["throughput"]) // (5.0 + 7.0) / 2
	assert.Equal(t, 0.15, profile.AverageMetrics["latency"])   // (0.1 + 0.2) / 2
}

func TestPerformanceBenchmarkingCollector_Integration(t *testing.T) {
	// This test verifies the complete workflow
	mockClient := &MockBenchmarkingSlurmClient{}
	mockJobManager := &MockBenchmarkingJobManager{}
	mockNodeManager := &MockBenchmarkingNodeManager{}

	testJobs := &slurm.JobList{
		Jobs: []*slurm.Job{
			{
				JobID:     "12345",
				JobState:  "COMPLETED",
				UserName:  "user1",
				Account:   "account1",
				Partition: "partition1",
				StartTime: time.Now().Add(-time.Hour),
				EndTime:   time.Now().Add(-30 * time.Minute),
				CPUs:      8,
				Memory:    "16GB",
			},
		},
	}

	testNodes := &slurm.NodeList{
		Nodes: []*slurm.Node{
			{
				Name:       "node001",
				State:      "IDLE",
				Partitions: []string{"partition1"},
				CPUTotal:   16,
				MemoryTotal: 32768,
			},
		},
	}

	mockClient.On("Jobs").Return(mockJobManager)
	mockClient.On("Nodes").Return(mockNodeManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(testNodes, nil)

	config := &BenchmarkingConfig{
		CollectionInterval:         30 * time.Second,
		BenchmarkRetention:         24 * time.Hour,
		EnableJobComparison:        true,
		EnableUserComparison:       true,
		EnableNodeComparison:       true,
		EnableClusterComparison:    true,
		EnableStatisticalAnalysis:  true,
		MaxJobsPerBenchmark:        100,
		MaxNodesPerBenchmark:       50,
		EnableParallelProcessing:   true,
		MaxConcurrentComparisons:   3,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	require.NoError(t, err)

	// Test metric collection
	registry := prometheus.NewRegistry()
	err = registry.Register(collector)
	require.NoError(t, err)

	// Collect metrics to verify no errors
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)
	assert.Greater(t, len(metricFamilies), 0)

	mockClient.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestJobPerformanceSnapshot_DataStructure(t *testing.T) {
	snapshot := &JobPerformanceSnapshot{
		JobID:     "12345",
		Timestamp: time.Now(),
		JobMetadata: &JobMetadata{
			UserName:  "testuser",
			Account:   "testaccount",
			Partition: "testpartition",
			JobSize:   8,
			Runtime:   time.Hour,
		},
		PerformanceData: map[string]float64{
			"throughput": 5.0,
			"latency":    0.1,
		},
		ResourceData: map[string]float64{
			"cpu_utilization":    0.8,
			"memory_utilization": 0.6,
		},
		EfficiencyData: map[string]float64{
			"overall_efficiency": 0.75,
		},
		DataCompleteness:   0.95,
		MeasurementQuality: 0.9,
	}

	assert.Equal(t, "12345", snapshot.JobID)
	assert.Equal(t, "testuser", snapshot.JobMetadata.UserName)
	assert.Equal(t, 5.0, snapshot.PerformanceData["throughput"])
	assert.Equal(t, 0.8, snapshot.ResourceData["cpu_utilization"])
	assert.Equal(t, 0.75, snapshot.EfficiencyData["overall_efficiency"])
	assert.Equal(t, 0.95, snapshot.DataCompleteness)
	assert.Equal(t, 0.9, snapshot.MeasurementQuality)
}

func TestUserPerformanceProfile_DataStructure(t *testing.T) {
	profile := &UserPerformanceProfile{
		UserName:    "testuser",
		Account:     "testaccount",
		LastUpdated: time.Now(),
		AverageMetrics: map[string]float64{
			"throughput": 6.0,
			"efficiency": 0.8,
		},
		BestMetrics: map[string]float64{
			"throughput": 10.0,
			"efficiency": 0.95,
		},
		WorstMetrics: map[string]float64{
			"throughput": 2.0,
			"efficiency": 0.6,
		},
		ConsistencyScore:    0.85,
		VariabilityScore:    0.2,
		ResourceEfficiency:  0.78,
		CostEfficiency:      0.82,
		QueueEfficiency:     0.9,
		ImprovementTrend:    "improving",
		ImprovementRate:     0.05,
		RelativeRanking:     15,
		PercentileScore:     85.0,
	}

	assert.Equal(t, "testuser", profile.UserName)
	assert.Equal(t, "testaccount", profile.Account)
	assert.Equal(t, 6.0, profile.AverageMetrics["throughput"])
	assert.Equal(t, 10.0, profile.BestMetrics["throughput"])
	assert.Equal(t, 2.0, profile.WorstMetrics["throughput"])
	assert.Equal(t, 0.85, profile.ConsistencyScore)
	assert.Equal(t, 0.2, profile.VariabilityScore)
	assert.Equal(t, "improving", profile.ImprovementTrend)
	assert.Equal(t, 85.0, profile.PercentileScore)
}

func TestNodePerformanceProfile_DataStructure(t *testing.T) {
	profile := &NodePerformanceProfile{
		NodeName:    "node001",
		Partition:   "partition1",
		LastUpdated: time.Now(),
		ThroughputMetrics: map[string]float64{
			"job_throughput": 8.5,
		},
		UtilizationMetrics: map[string]float64{
			"cpu_utilization": 0.75,
		},
		UptimePercentage:     99.5,
		FailureRate:          0.001,
		MaintenanceEvents:    2,
		JobSuccess:           0.98,
		AverageJobRuntime:    45 * time.Minute,
		ResourceWasteRate:    0.12,
		RelativePerformance:  1.15,
		PerformanceRank:      3,
		PerformanceGrade:     "A",
		PerformanceIssues:    []string{"none"},
		MaintenanceNeeds:     []string{"routine_check"},
		UpgradeRecommendations: []string{"memory_upgrade"},
	}

	assert.Equal(t, "node001", profile.NodeName)
	assert.Equal(t, "partition1", profile.Partition)
	assert.Equal(t, 8.5, profile.ThroughputMetrics["job_throughput"])
	assert.Equal(t, 0.75, profile.UtilizationMetrics["cpu_utilization"])
	assert.Equal(t, 99.5, profile.UptimePercentage)
	assert.Equal(t, 0.001, profile.FailureRate)
	assert.Equal(t, 0.98, profile.JobSuccess)
	assert.Equal(t, 1.15, profile.RelativePerformance)
	assert.Equal(t, 3, profile.PerformanceRank)
	assert.Equal(t, "A", profile.PerformanceGrade)
}

func TestClusterPerformanceSnapshot_DataStructure(t *testing.T) {
	snapshot := &ClusterPerformanceSnapshot{
		ClusterName:     "test-cluster",
		Timestamp:       time.Now(),
		TotalThroughput: 150.0,
		AverageLatency:  0.05,
		OverallEfficiency: 0.82,
		ResourceUtilization: map[string]float64{
			"cpu":    0.70,
			"memory": 0.65,
			"gpu":    0.45,
		},
		TotalCapacity: map[string]float64{
			"cpu":    1000.0,
			"memory": 2048.0,
		},
		AvailableCapacity: map[string]float64{
			"cpu":    300.0,
			"memory": 716.8,
		},
		UtilizationRate:  0.70,
		AverageQueueTime: 5 * time.Minute,
		QueueThroughput:  25.0,
		QueueEfficiency:  0.88,
		PowerEfficiency:  0.75,
		CarbonFootprint:  125.5,
		SystemUptime:     99.8,
		FailureRate:      0.002,
		MaintenanceTime:  2 * time.Hour,
		PerformanceTrends: map[string]string{
			"throughput": "improving",
			"efficiency": "stable",
		},
		DataQuality:       0.95,
		MeasurementPeriod: time.Hour,
	}

	assert.Equal(t, "test-cluster", snapshot.ClusterName)
	assert.Equal(t, 150.0, snapshot.TotalThroughput)
	assert.Equal(t, 0.05, snapshot.AverageLatency)
	assert.Equal(t, 0.82, snapshot.OverallEfficiency)
	assert.Equal(t, 0.70, snapshot.ResourceUtilization["cpu"])
	assert.Equal(t, 0.70, snapshot.UtilizationRate)
	assert.Equal(t, 99.8, snapshot.SystemUptime)
	assert.Equal(t, "improving", snapshot.PerformanceTrends["throughput"])
	assert.Equal(t, 0.95, snapshot.DataQuality)
}

// Benchmark tests
func BenchmarkPerformanceBenchmarkingCollector_Collect(b *testing.B) {
	mockClient := &MockBenchmarkingSlurmClient{}
	mockJobManager := &MockBenchmarkingJobManager{}
	mockNodeManager := &MockBenchmarkingNodeManager{}

	// Create large datasets for benchmarking
	jobs := make([]*slurm.Job, 100)
	for i := 0; i < 100; i++ {
		jobs[i] = &slurm.Job{
			JobID:     fmt.Sprintf("job%d", i),
			JobState:  "COMPLETED",
			UserName:  fmt.Sprintf("user%d", i%10),
			Account:   fmt.Sprintf("account%d", i%5),
			Partition: "partition1",
			StartTime: time.Now().Add(-time.Hour),
			EndTime:   time.Now().Add(-30 * time.Minute),
			CPUs:      8,
			Memory:    "16GB",
		}
	}

	nodes := make([]*slurm.Node, 50)
	for i := 0; i < 50; i++ {
		nodes[i] = &slurm.Node{
			Name:       fmt.Sprintf("node%03d", i),
			State:      "IDLE",
			Partitions: []string{"partition1"},
			CPUTotal:   16,
			MemoryTotal: 32768,
		}
	}

	testJobs := &slurm.JobList{Jobs: jobs}
	testNodes := &slurm.NodeList{Nodes: nodes}

	mockClient.On("Jobs").Return(mockJobManager)
	mockClient.On("Nodes").Return(mockNodeManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(testNodes, nil)

	config := &BenchmarkingConfig{
		CollectionInterval:   60 * time.Second,
		MaxJobsPerBenchmark:  1000,
		MaxNodesPerBenchmark: 100,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 1000)
		collector.Collect(ch)
		close(ch)
		// Drain the channel
		for range ch {
		}
	}
}

func BenchmarkPerformanceBenchmarkingCollector_createJobPerformanceSnapshot(b *testing.B) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	if err != nil {
		b.Fatal(err)
	}

	job := &slurm.Job{
		JobID:     "12345",
		UserName:  "testuser",
		Account:   "testaccount",
		Partition: "testpartition",
		JobState:  "COMPLETED",
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(-30 * time.Minute),
		CPUs:      8,
		Memory:    "16GB",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.createJobPerformanceSnapshot(job)
	}
}

func BenchmarkPerformanceBenchmarkingCollector_calculatePerformanceScore(b *testing.B) {
	mockClient := &MockBenchmarkingSlurmClient{}
	config := &BenchmarkingConfig{
		CollectionInterval: 60 * time.Second,
	}

	collector, err := NewPerformanceBenchmarkingCollector(mockClient, nil, config)
	if err != nil {
		b.Fatal(err)
	}

	snapshot := &JobPerformanceSnapshot{
		JobID:          "12345",
		PerformanceData: map[string]float64{"throughput": 5.0},
		ResourceData:    map[string]float64{"cpu_utilization": 0.8},
		EfficiencyData:  map[string]float64{"overall_efficiency": 0.75},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.calculateJobPerformanceScore(snapshot)
	}
}