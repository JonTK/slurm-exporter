package collector

import (
	"context"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSlurmClient for testing
type MockTaskUtilizationSlurmClient struct {
	mock.Mock
}

func (m *MockTaskUtilizationSlurmClient) Jobs() slurm.JobManager {
	args := m.Called()
	return args.Get(0).(slurm.JobManager)
}

func (m *MockTaskUtilizationSlurmClient) Nodes() slurm.NodeManager {
	args := m.Called()
	return args.Get(0).(slurm.NodeManager)
}

func (m *MockTaskUtilizationSlurmClient) Partitions() slurm.PartitionManager {
	args := m.Called()
	return args.Get(0).(slurm.PartitionManager)
}

func (m *MockTaskUtilizationSlurmClient) Info() slurm.InfoManager {
	args := m.Called()
	return args.Get(0).(slurm.InfoManager)
}

// MockJobManager for testing task utilization
type MockTaskUtilizationJobManager struct {
	mock.Mock
}

func (m *MockTaskUtilizationJobManager) List(ctx context.Context, opts ...interface{}) (*slurm.JobList, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*slurm.JobList), args.Error(1)
}

func (m *MockTaskUtilizationJobManager) Get(ctx context.Context, jobID string, opts ...interface{}) (*slurm.Job, error) {
	args := m.Called(ctx, jobID, opts)
	return args.Get(0).(*slurm.Job), args.Error(1)
}

func (m *MockTaskUtilizationJobManager) Submit(ctx context.Context, job interface{}) (interface{}, error) {
	args := m.Called(ctx, job)
	return args.Get(0), args.Error(1)
}

func (m *MockTaskUtilizationJobManager) Cancel(ctx context.Context, jobID string, opts ...interface{}) error {
	args := m.Called(ctx, jobID, opts)
	return args.Error(0)
}

func TestNewTaskUtilizationMonitor(t *testing.T) {
	mockClient := &MockTaskUtilizationSlurmClient{}
	config := &TaskUtilizationConfig{
		CollectionInterval:     30 * time.Second,
		TaskCacheTTL:          300 * time.Second,
		LoadBalanceThreshold:  0.8,
		EfficiencyThreshold:   0.7,
		MaxTasksPerJob:        1000,
		MetricsBufferSize:     10000,
	}

	monitor := NewTaskUtilizationMonitor(mockClient, config, nil)

	assert.NotNil(t, monitor)
	assert.Equal(t, mockClient, monitor.slurmClient)
	assert.Equal(t, config, monitor.config)
	assert.NotNil(t, monitor.metrics)
	assert.NotNil(t, monitor.taskData)
	assert.NotNil(t, monitor.stepData)
	assert.NotNil(t, monitor.loadBalanceAnalyzer)
	assert.NotNil(t, monitor.performanceAnalyzer)
}

func TestTaskUtilizationMonitor_Describe(t *testing.T) {
	mockClient := &MockTaskUtilizationSlurmClient{}
	config := &TaskUtilizationConfig{
		CollectionInterval: 30 * time.Second,
	}

	monitor := NewTaskUtilizationMonitor(mockClient, config, nil)

	ch := make(chan *prometheus.Desc, 20)
	monitor.Describe(ch)
	close(ch)

	// Count the number of metrics described
	count := 0
	for range ch {
		count++
	}

	// Should have multiple metrics described
	assert.Greater(t, count, 10)
}

func TestTaskUtilizationMonitor_Collect(t *testing.T) {
	mockClient := &MockTaskUtilizationSlurmClient{}
	mockJobManager := &MockTaskUtilizationJobManager{}

	// Create test job data
	testJobs := &slurm.JobList{
		Jobs: []*slurm.Job{
			{
				JobID:     "12345",
				JobState:  "RUNNING",
				StartTime: time.Now().Add(-time.Hour),
				Nodes:     []string{"node001", "node002"},
				CPUs:      8,
				Memory:    "16GB",
				NodeList:  "node[001-002]",
			},
			{
				JobID:     "12346",
				JobState:  "COMPLETED",
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now().Add(-30 * time.Minute),
				Nodes:     []string{"node003"},
				CPUs:      4,
				Memory:    "8GB",
				NodeList:  "node003",
			},
		},
	}

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)

	config := &TaskUtilizationConfig{
		CollectionInterval: 30 * time.Second,
		TaskCacheTTL:      300 * time.Second,
	}

	monitor := NewTaskUtilizationMonitor(mockClient, config, nil)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(monitor)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
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
}

func TestTaskUtilizationMonitor_collectTaskData(t *testing.T) {
	mockClient := &MockTaskUtilizationSlurmClient{}
	config := &TaskUtilizationConfig{
		CollectionInterval: 30 * time.Second,
		MaxTasksPerJob:     10,
	}

	monitor := NewTaskUtilizationMonitor(mockClient, config, nil)

	// Test job
	job := &slurm.Job{
		JobID:     "12345",
		JobState:  "RUNNING",
		StartTime: time.Now().Add(-time.Hour),
		Nodes:     []string{"node001", "node002"},
		CPUs:      8,
		Memory:    "16GB",
	}

	// Mock the task data collection (this would normally come from SLURM)
	taskData := monitor.collectTaskData(job)

	assert.NotNil(t, taskData)
	assert.Equal(t, "12345", taskData.JobID)
	assert.Greater(t, len(taskData.Tasks), 0)

	// Verify task data structure
	for _, task := range taskData.Tasks {
		assert.NotEmpty(t, task.TaskID)
		assert.NotEmpty(t, task.NodeID)
		assert.GreaterOrEqual(t, task.CPUUtilization, 0.0)
		assert.LessOrEqual(t, task.CPUUtilization, 1.0)
		assert.GreaterOrEqual(t, task.MemoryUtilization, 0.0)
		assert.LessOrEqual(t, task.MemoryUtilization, 1.0)
	}
}

func TestTaskUtilizationMonitor_analyzeLoadBalance(t *testing.T) {
	monitor := &TaskUtilizationMonitor{
		config: &TaskUtilizationConfig{
			LoadBalanceThreshold: 0.8,
		},
		loadBalanceAnalyzer: &LoadBalanceAnalyzer{},
	}

	// Create test step data with unbalanced tasks
	stepData := &JobStepData{
		JobID:  "12345",
		StepID: "0",
		Tasks: []*JobStepTaskData{
			{TaskID: "0", CPUUtilization: 0.9, MemoryUtilization: 0.8},
			{TaskID: "1", CPUUtilization: 0.3, MemoryUtilization: 0.4},
			{TaskID: "2", CPUUtilization: 0.7, MemoryUtilization: 0.6},
			{TaskID: "3", CPUUtilization: 0.2, MemoryUtilization: 0.3},
		},
	}

	analysis := monitor.analyzeLoadBalance(stepData)

	assert.NotNil(t, analysis)
	assert.Equal(t, "12345", analysis.JobID)
	assert.Equal(t, "0", analysis.StepID)
	assert.Greater(t, analysis.CPUImbalance, 0.0)
	assert.Greater(t, analysis.MemoryImbalance, 0.0)
	assert.False(t, analysis.IsBalanced)
}

func TestTaskUtilizationMonitor_analyzeTaskPerformance(t *testing.T) {
	monitor := &TaskUtilizationMonitor{
		config: &TaskUtilizationConfig{
			EfficiencyThreshold: 0.7,
		},
		performanceAnalyzer: &TaskPerformanceAnalyzer{},
	}

	// Create test step data
	stepData := &JobStepData{
		JobID:  "12345",
		StepID: "0",
		Tasks: []*JobStepTaskData{
			{TaskID: "0", CPUUtilization: 0.9, MemoryUtilization: 0.8},
			{TaskID: "1", CPUUtilization: 0.6, MemoryUtilization: 0.5},
			{TaskID: "2", CPUUtilization: 0.8, MemoryUtilization: 0.7},
		},
	}

	analysis := monitor.analyzeTaskPerformance(stepData)

	assert.NotNil(t, analysis)
	assert.Equal(t, "12345", analysis.JobID)
	assert.Equal(t, "0", analysis.StepID)
	assert.Greater(t, len(analysis.TaskEfficiencies), 0)
	assert.GreaterOrEqual(t, analysis.AverageEfficiency, 0.0)
	assert.LessOrEqual(t, analysis.AverageEfficiency, 1.0)

	// Check task efficiency calculations
	for taskID, efficiency := range analysis.TaskEfficiencies {
		assert.NotEmpty(t, taskID)
		assert.GreaterOrEqual(t, efficiency, 0.0)
		assert.LessOrEqual(t, efficiency, 1.0)
	}
}

func TestTaskUtilizationMonitor_calculateStandardDeviation(t *testing.T) {
	monitor := &TaskUtilizationMonitor{}

	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "uniform values",
			values:   []float64{1.0, 1.0, 1.0, 1.0},
			expected: 0.0,
		},
		{
			name:     "varied values",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 1.5811388300841898, // sqrt(2.5)
		},
		{
			name:     "empty values",
			values:   []float64{},
			expected: 0.0,
		},
		{
			name:     "single value",
			values:   []float64{5.0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.calculateStandardDeviation(tt.values)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestTaskUtilizationMonitor_calculateMean(t *testing.T) {
	monitor := &TaskUtilizationMonitor{}

	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "positive values",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "mixed values",
			values:   []float64{-1.0, 0.0, 1.0},
			expected: 0.0,
		},
		{
			name:     "empty values",
			values:   []float64{},
			expected: 0.0,
		},
		{
			name:     "single value",
			values:   []float64{7.5},
			expected: 7.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := monitor.calculateMean(tt.values)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTaskUtilizationMonitor_updateMetrics(t *testing.T) {
	mockClient := &MockTaskUtilizationSlurmClient{}
	config := &TaskUtilizationConfig{
		CollectionInterval: 30 * time.Second,
	}

	monitor := NewTaskUtilizationMonitor(mockClient, config, nil)

	// Create test step data
	stepData := &JobStepData{
		JobID:     "12345",
		StepID:    "0",
		Timestamp: time.Now(),
		Tasks: []*JobStepTaskData{
			{
				JobID:              "12345",
				StepID:             "0",
				TaskID:             "0",
				NodeID:             "node001",
				CPUUtilization:     0.8,
				MemoryUtilization:  0.7,
				IOUtilization:      0.6,
				NetworkUtilization: 0.5,
			},
		},
	}

	// Create test analyses
	loadBalance := &LoadBalanceAnalysis{
		JobID:           "12345",
		StepID:          "0",
		CPUImbalance:    0.2,
		MemoryImbalance: 0.3,
		IsBalanced:      false,
	}

	taskPerformance := &TaskPerformanceAnalysis{
		JobID:             "12345",
		StepID:            "0",
		AverageEfficiency: 0.75,
		TaskEfficiencies:  map[string]float64{"0": 0.75},
	}

	// Update metrics
	monitor.updateMetrics(stepData, loadBalance, taskPerformance)

	// Verify metrics were updated (basic test that no errors occurred)
	assert.NotNil(t, monitor.metrics)
}

func TestTaskUtilizationMonitor_Integration(t *testing.T) {
	// This test verifies the complete workflow
	mockClient := &MockTaskUtilizationSlurmClient{}
	mockJobManager := &MockTaskUtilizationJobManager{}

	testJobs := &slurm.JobList{
		Jobs: []*slurm.Job{
			{
				JobID:     "12345",
				JobState:  "RUNNING",
				StartTime: time.Now().Add(-time.Hour),
				Nodes:     []string{"node001"},
				CPUs:      4,
				Memory:    "8GB",
			},
		},
	}

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)

	config := &TaskUtilizationConfig{
		CollectionInterval:     10 * time.Second,
		TaskCacheTTL:          60 * time.Second,
		LoadBalanceThreshold:  0.8,
		EfficiencyThreshold:   0.7,
		MaxTasksPerJob:        100,
		MetricsBufferSize:     1000,
	}

	monitor := NewTaskUtilizationMonitor(mockClient, config, nil)

	// Test metric collection
	registry := prometheus.NewRegistry()
	err := registry.Register(monitor)
	require.NoError(t, err)

	// Collect metrics to verify no errors
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)
	assert.Greater(t, len(metricFamilies), 0)

	mockClient.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
}

func TestTaskUtilizationConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *TaskUtilizationConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &TaskUtilizationConfig{
				CollectionInterval:     30 * time.Second,
				TaskCacheTTL:          300 * time.Second,
				LoadBalanceThreshold:  0.8,
				EfficiencyThreshold:   0.7,
				MaxTasksPerJob:        1000,
				MetricsBufferSize:     10000,
			},
			wantErr: false,
		},
		{
			name: "invalid collection interval",
			config: &TaskUtilizationConfig{
				CollectionInterval: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid load balance threshold",
			config: &TaskUtilizationConfig{
				CollectionInterval:    30 * time.Second,
				LoadBalanceThreshold: 1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid efficiency threshold",
			config: &TaskUtilizationConfig{
				CollectionInterval:   30 * time.Second,
				LoadBalanceThreshold: 0.8,
				EfficiencyThreshold:  -0.1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkTaskUtilizationMonitor_Collect(b *testing.B) {
	mockClient := &MockTaskUtilizationSlurmClient{}
	mockJobManager := &MockTaskUtilizationJobManager{}

	// Create large job list for benchmarking
	jobs := make([]*slurm.Job, 100)
	for i := 0; i < 100; i++ {
		jobs[i] = &slurm.Job{
			JobID:     string(rune(10000 + i)),
			JobState:  "RUNNING",
			StartTime: time.Now().Add(-time.Hour),
			Nodes:     []string{"node001"},
			CPUs:      4,
			Memory:    "8GB",
		}
	}

	testJobs := &slurm.JobList{Jobs: jobs}

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)

	config := &TaskUtilizationConfig{
		CollectionInterval: 30 * time.Second,
		MaxTasksPerJob:     10,
	}

	monitor := NewTaskUtilizationMonitor(mockClient, config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 1000)
		monitor.Collect(ch)
		close(ch)
		// Drain the channel
		for range ch {
		}
	}
}

func BenchmarkTaskUtilizationMonitor_analyzeLoadBalance(b *testing.B) {
	monitor := &TaskUtilizationMonitor{
		config: &TaskUtilizationConfig{
			LoadBalanceThreshold: 0.8,
		},
		loadBalanceAnalyzer: &LoadBalanceAnalyzer{},
	}

	// Create test data with many tasks
	tasks := make([]*JobStepTaskData, 100)
	for i := 0; i < 100; i++ {
		tasks[i] = &JobStepTaskData{
			TaskID:             string(rune(i)),
			CPUUtilization:     float64(i%10) / 10.0,
			MemoryUtilization:  float64(i%8) / 8.0,
		}
	}

	stepData := &JobStepData{
		JobID:  "12345",
		StepID: "0",
		Tasks:  tasks,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.analyzeLoadBalance(stepData)
	}
}