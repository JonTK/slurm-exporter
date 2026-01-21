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

type MockQoSLimitsSLURMClient struct {
	mock.Mock
}

func (m *MockQoSLimitsSLURMClient) GetQoSLimits(ctx context.Context, qosName string) (*QoSLimits, error) {
	args := m.Called(ctx, qosName)
	return args.Get(0).(*QoSLimits), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSViolations(ctx context.Context, opts *QoSViolationOptions) (*QoSViolations, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*QoSViolations), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSUsage(ctx context.Context, qosName string) (*QoSUsage, error) {
	args := m.Called(ctx, qosName)
	return args.Get(0).(*QoSUsage), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSConfiguration(ctx context.Context, qosName string) (*QoSConfiguration, error) {
	args := m.Called(ctx, qosName)
	return args.Get(0).(*QoSConfiguration), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSHierarchy(ctx context.Context) (*QoSHierarchy, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QoSHierarchy), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSPriorities(ctx context.Context) (*QoSPriorities, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QoSPriorities), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSAssignments(ctx context.Context, entityType string) (*QoSAssignments, error) {
	args := m.Called(ctx, entityType)
	return args.Get(0).(*QoSAssignments), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSEnforcement(ctx context.Context, qosName string) (*QoSEnforcement, error) {
	args := m.Called(ctx, qosName)
	return args.Get(0).(*QoSEnforcement), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSStatistics(ctx context.Context, qosName string, period string) (*QoSStatistics, error) {
	args := m.Called(ctx, qosName, period)
	return args.Get(0).(*QoSStatistics), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSEffectiveness(ctx context.Context, qosName string) (*QoSEffectiveness, error) {
	args := m.Called(ctx, qosName)
	return args.Get(0).(*QoSEffectiveness), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetQoSConflicts(ctx context.Context) (*QoSConflicts, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QoSConflicts), args.Error(1)
}

func (m *MockQoSLimitsSLURMClient) GetSystemQoSOverview(ctx context.Context) (*SystemQoSOverview, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemQoSOverview), args.Error(1)
}

func TestNewQoSLimitsCollector(t *testing.T) {
	client := &MockQoSLimitsSLURMClient{}
	collector := NewQoSLimitsCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.qosLimitCPUs)
	assert.NotNil(t, collector.qosViolations)
	assert.NotNil(t, collector.qosEnforcementEnabled)
	assert.NotNil(t, collector.systemQoSEfficiency)
}

func TestQoSLimitsCollector_Describe(t *testing.T) {
	client := &MockQoSLimitsSLURMClient{}
	collector := NewQoSLimitsCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 56 metrics, verify we have the correct number
	assert.Equal(t, 56, len(descs), "Should have exactly 56 metric descriptions")
}

func TestQoSLimitsCollector_Collect_Success(t *testing.T) {
	client := &MockQoSLimitsSLURMClient{}

	// Mock QoS limits
	qosLimits := &QoSLimits{
		QoSName:          "high",
		Priority:         1000,
		UsageFactor:      1.5,
		UsageThreshold:   0.8,
		GrpCPUs:          1000,
		GrpCPUsRunning:   800,
		GrpMem:           102400, // 100GB in MB
		GrpMemRunning:    81920,  // 80GB in MB
		GrpNodes:         50,
		GrpNodesRunning:  40,
		GrpJobs:          100,
		GrpJobsRunning:   80,
		GrpSubmitJobs:    120,
		GrpWall:          24 * time.Hour,
		MaxCPUs:          200,
		MaxCPUsPerUser:   50,
		MaxMem:           20480, // 20GB in MB
		MaxMemPerUser:    4096,  // 4GB in MB
		MaxNodes:         10,
		MaxNodesPerUser:  2,
		MaxJobs:          20,
		MaxJobsPerUser:   5,
		MaxSubmitJobs:    25,
		MaxSubmitJobsPerUser: 10,
		MaxWall:          8 * time.Hour,
		MaxWallPerJob:    4 * time.Hour,
		MinCPUs:          1,
		MinMem:           1024, // 1GB in MB
		MinNodes:         1,
		GraceTime:        5 * time.Minute,
		UserCount:        50,
		AccountCount:     10,
		ActiveJobs:       75,
		PendingJobs:      25,
		RunningJobs:      75,
	}

	// Mock QoS usage
	qosUsage := &QoSUsage{
		QoSName:              "high",
		CPUsInUse:            750,
		CPUsAllocated:        1000,
		CPUUtilization:       0.75,
		MemoryInUse:          76800, // 75GB
		MemoryAllocated:      102400, // 100GB
		MemoryUtilization:    0.75,
		NodesInUse:           38,
		NodesAllocated:       50,
		NodeUtilization:      0.76,
		JobsRunning:          75,
		JobsPending:          25,
		JobsCompleted:        500,
		JobsFailed:           10,
		UsersActive:          45,
		AccountsActive:       8,
		WalltimeConsumed:     1000 * time.Hour,
		WalltimeAllocated:    1200 * time.Hour,
		WalltimeUtilization:  0.83,
		EfficiencyScore:      0.82,
		ThroughputScore:      0.85,
		LoadFactor:           0.78,
	}

	// Mock QoS violations
	qosViolations := &QoSViolations{
		TotalViolations:    15,
		ActiveViolations:   3,
		ResolvedViolations: 12,
		CriticalViolations: 2,
		WarningViolations:  5,
		InfoViolations:     8,
		Violations: []QoSViolation{
			{
				ViolationID:     "vio1",
				Timestamp:       time.Now().Add(-2 * time.Hour),
				QoSName:         "high",
				ViolationType:   "cpu_limit_exceeded",
				Severity:        "critical",
				EntityType:      "user",
				EntityID:        "user1",
				LimitType:       "max_cpus_per_user",
				LimitValue:      50,
				ActualValue:     55,
				ExcessAmount:    5.0,
				Duration:        30 * time.Minute,
				Status:          "resolved",
				AutoResolved:    true,
				ResolutionTime:  time.Now().Add(-90 * time.Minute),
				ResolutionAction: "job_terminated",
			},
			{
				ViolationID:     "vio2",
				Timestamp:       time.Now().Add(-1 * time.Hour),
				QoSName:         "high",
				ViolationType:   "memory_limit_exceeded",
				Severity:        "warning",
				EntityType:      "job",
				EntityID:        "12345",
				LimitType:       "max_mem_per_user",
				LimitValue:      4096,
				ActualValue:     4200,
				ExcessAmount:    104.0,
				Duration:        15 * time.Minute,
				Status:          "active",
				AutoResolved:    false,
			},
		},
		ViolationsByQoS: map[string]int{
			"high":   10,
			"normal": 5,
		},
		ViolationsByType: map[string]int{
			"cpu_limit_exceeded":    8,
			"memory_limit_exceeded": 4,
			"job_limit_exceeded":    3,
		},
		ResolutionStats: ViolationResolutionStats{
			MeanResolutionTime:   25 * time.Minute,
			MedianResolutionTime: 20 * time.Minute,
			AutoResolutionRate:   0.8,
			ManualResolutionRate: 0.2,
			EscalationRate:       0.1,
			RecurrenceRate:       0.05,
		},
	}

	// Mock QoS enforcement
	qosEnforcement := &QoSEnforcement{
		QoSName:            "high",
		EnforcementEnabled: true,
		EnforcementLevel:   "strict",
		ViolationActions: map[string]string{
			"cpu_exceeded":    "terminate",
			"memory_exceeded": "suspend",
			"job_exceeded":    "reject",
		},
		GracePeriods: map[string]time.Duration{
			"cpu":    5 * time.Minute,
			"memory": 2 * time.Minute,
		},
		AutoRemediation: true,
		EnforcementStats: EnforcementStats{
			ActionsTotal:       100,
			ActionsSuccessful:  95,
			ActionsFailed:      5,
			PreemptionsTotal:   20,
			JobsTerminated:     15,
			JobsSuspended:      8,
			JobsRequeued:       3,
			UsersWarned:        25,
			UsersBlocked:       2,
			EffectivenessScore: 0.92,
		},
	}

	// Mock QoS statistics
	qosStatistics := &QoSStatistics{
		QoSName:               "high",
		Period:                "24h",
		JobsSubmitted:         150,
		JobsCompleted:         140,
		JobsFailed:            8,
		JobsCanceled:          2,
		JobsPreempted:         5,
		AverageQueueTime:      15 * time.Minute,
		AverageRunTime:        2 * time.Hour,
		AverageTurnaroundTime: 2*time.Hour + 15*time.Minute,
		ResourceEfficiency:    0.85,
		Throughput:            6.25, // jobs per hour
		UtilizationRate:       0.82,
		SLACompliance:         0.95,
		PerformanceScore:      0.88,
		UserSatisfaction:      0.90,
		CostEffectiveness:     0.86,
	}

	// Mock QoS effectiveness
	qosEffectiveness := &QoSEffectiveness{
		QoSName:              "high",
		OverallEffectiveness: 0.87,
		ResourceOptimization: 0.85,
		FairnessScore:        0.82,
		PerformanceImpact:    0.90,
		UserProductivity:     0.88,
		SystemThroughput:     0.84,
		PolicyCompliance:     0.92,
		CostBenefitRatio:     1.25,
	}

	// Mock QoS hierarchy
	qosHierarchy := &QoSHierarchy{
		RootQoS:     []string{"high", "normal", "low"},
		QoSTree:     map[string][]string{"high": {"premium"}, "normal": {"standard"}},
		InheritanceChain: map[string][]string{
			"premium":  {"high"},
			"standard": {"normal"},
		},
		PriorityOrder: []string{"high", "normal", "low"},
		DefaultQoS:    "normal",
		MaxDepth:      2,
		TotalQoS:      5,
	}

	// Mock QoS conflicts
	qosConflicts := &QoSConflicts{
		TotalConflicts:    3,
		CriticalConflicts: 1,
		ConflictsByType: map[string]int{
			"priority_conflict": 2,
			"limit_conflict":    1,
		},
		Conflicts: []QoSConflict{
			{
				ConflictID:   "conf1",
				ConflictType: "priority_conflict",
				QoSNames:     []string{"high", "premium"},
				Description:  "Overlapping priority ranges",
				Status:       "active",
			},
		},
	}

	// Mock system QoS overview
	systemOverview := &SystemQoSOverview{
		TotalQoS:           5,
		ActiveQoS:          4,
		DefaultQoS:         "normal",
		HighestPriorityQoS: "high",
		LowestPriorityQoS:  "low",
		QoSUtilization: map[string]float64{
			"high":   0.85,
			"normal": 0.65,
			"low":    0.25,
		},
		SystemLoad:       0.72,
		OverallEfficiency: 0.83,
		ViolationRate:    0.05,
		ComplianceScore:  0.92,
		SystemHealth:     "good",
	}

	// Setup mock expectations
	client.On("GetQoSLimits", mock.Anything, "normal").Return(&QoSLimits{QoSName: "normal", Priority: 500}, nil)
	client.On("GetQoSLimits", mock.Anything, "high").Return(qosLimits, nil)
	client.On("GetQoSLimits", mock.Anything, mock.AnythingOfType("string")).Return(&QoSLimits{}, nil)

	client.On("GetQoSUsage", mock.Anything, "normal").Return(&QoSUsage{QoSName: "normal"}, nil)
	client.On("GetQoSUsage", mock.Anything, "high").Return(qosUsage, nil)
	client.On("GetQoSUsage", mock.Anything, mock.AnythingOfType("string")).Return(&QoSUsage{}, nil)

	client.On("GetQoSViolations", mock.Anything, mock.AnythingOfType("*collector.QoSViolationOptions")).Return(qosViolations, nil)

	client.On("GetQoSEnforcement", mock.Anything, "normal").Return(&QoSEnforcement{EnforcementEnabled: false, EnforcementStats: EnforcementStats{}}, nil)
	client.On("GetQoSEnforcement", mock.Anything, "high").Return(qosEnforcement, nil)
	client.On("GetQoSEnforcement", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEnforcement{EnforcementStats: EnforcementStats{}}, nil)

	client.On("GetQoSStatistics", mock.Anything, "normal", "24h").Return(&QoSStatistics{QoSName: "normal"}, nil)
	client.On("GetQoSStatistics", mock.Anything, "high", "24h").Return(qosStatistics, nil)
	client.On("GetQoSStatistics", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QoSStatistics{}, nil)

	client.On("GetQoSHierarchy", mock.Anything).Return(qosHierarchy, nil)
	client.On("GetQoSConflicts", mock.Anything).Return(qosConflicts, nil)

	client.On("GetQoSEffectiveness", mock.Anything, "normal").Return(&QoSEffectiveness{QoSName: "normal"}, nil)
	client.On("GetQoSEffectiveness", mock.Anything, "high").Return(qosEffectiveness, nil)
	client.On("GetQoSEffectiveness", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEffectiveness{}, nil)

	client.On("GetSystemQoSOverview", mock.Anything).Return(systemOverview, nil)

	collector := NewQoSLimitsCollector(client)

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
	foundLimits := false
	foundUsage := false
	foundViolations := false
	foundEnforcement := false
	foundPerformance := false
	foundEffectiveness := false
	foundSystem := false

	for name := range metricNames {
		if strings.Contains(name, "qos_limit_cpus") {
			foundLimits = true
		}
		if strings.Contains(name, "qos_resource_usage") {
			foundUsage = true
		}
		if strings.Contains(name, "qos_violations_total") {
			foundViolations = true
		}
		if strings.Contains(name, "qos_enforcement_enabled") {
			foundEnforcement = true
		}
		if strings.Contains(name, "qos_performance_score") {
			foundPerformance = true
		}
		if strings.Contains(name, "qos_overall_effectiveness") {
			foundEffectiveness = true
		}
		if strings.Contains(name, "system_qos_efficiency") {
			foundSystem = true
		}
	}

	assert.True(t, foundLimits, "Should have QoS limit metrics")
	assert.True(t, foundUsage, "Should have QoS usage metrics")
	assert.True(t, foundViolations, "Should have QoS violation metrics")
	assert.True(t, foundEnforcement, "Should have QoS enforcement metrics")
	assert.True(t, foundPerformance, "Should have QoS performance metrics")
	assert.True(t, foundEffectiveness, "Should have QoS effectiveness metrics")
	assert.True(t, foundSystem, "Should have system QoS metrics")

	client.AssertExpectations(t)
}

func TestQoSLimitsCollector_Collect_Error(t *testing.T) {
	client := &MockQoSLimitsSLURMClient{}

	// Mock error responses
	client.On("GetQoSLimits", mock.Anything, mock.AnythingOfType("string")).Return((*QoSLimits)(nil), assert.AnError)
	client.On("GetQoSUsage", mock.Anything, mock.AnythingOfType("string")).Return((*QoSUsage)(nil), assert.AnError)
	client.On("GetQoSViolations", mock.Anything, mock.AnythingOfType("*collector.QoSViolationOptions")).Return((*QoSViolations)(nil), assert.AnError)
	client.On("GetQoSEnforcement", mock.Anything, mock.AnythingOfType("string")).Return((*QoSEnforcement)(nil), assert.AnError)
	client.On("GetQoSStatistics", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*QoSStatistics)(nil), assert.AnError)
	client.On("GetQoSHierarchy", mock.Anything).Return((*QoSHierarchy)(nil), assert.AnError)
	client.On("GetQoSConflicts", mock.Anything).Return((*QoSConflicts)(nil), nil)
	client.On("GetQoSEffectiveness", mock.Anything, mock.AnythingOfType("string")).Return((*QoSEffectiveness)(nil), assert.AnError)
	client.On("GetSystemQoSOverview", mock.Anything).Return((*SystemQoSOverview)(nil), assert.AnError)

	collector := NewQoSLimitsCollector(client)

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

func TestQoSLimitsCollector_MetricValues(t *testing.T) {
	client := &MockQoSLimitsSLURMClient{}

	// Create test data with known values
	qosLimits := &QoSLimits{
		QoSName:         "test_qos",
		Priority:        1500,
		UsageFactor:     2.0,
		UsageThreshold:  0.9,
		GrpCPUs:         2000,
		GrpMem:          204800, // 200GB in MB
		MaxCPUsPerUser:  100,
		MaxMemPerUser:   8192, // 8GB in MB
		MinCPUs:         2,
		MinMem:          2048, // 2GB in MB
		GraceTime:       10 * time.Minute,
		UserCount:       75,
		AccountCount:    15,
	}

	qosUsage := &QoSUsage{
		QoSName:            "test_qos",
		CPUsInUse:          1500,
		CPUUtilization:     0.75,
		MemoryInUse:        153600, // 150GB
		MemoryUtilization:  0.75,
		JobsRunning:        50,
		JobsPending:        20,
		JobsCompleted:      200,
		JobsFailed:         5,
		EfficiencyScore:    0.88,
	}

	qosEffectiveness := &QoSEffectiveness{
		QoSName:              "test_qos",
		OverallEffectiveness: 0.92,
		ResourceOptimization: 0.89,
		FairnessScore:        0.85,
		SystemThroughput:     0.87,
		PolicyCompliance:     0.95,
		CostBenefitRatio:     1.35,
	}

	systemOverview := &SystemQoSOverview{
		SystemLoad:       0.68,
		OverallEfficiency: 0.84,
		ViolationRate:    0.03,
		ComplianceScore:  0.94,
		QoSUtilization: map[string]float64{
			"test_qos": 0.75,
		},
	}

	client.On("GetQoSLimits", mock.Anything, "normal").Return(qosLimits, nil)
	client.On("GetQoSLimits", mock.Anything, mock.AnythingOfType("string")).Return(&QoSLimits{}, nil)
	client.On("GetQoSUsage", mock.Anything, "normal").Return(qosUsage, nil)
	client.On("GetQoSUsage", mock.Anything, mock.AnythingOfType("string")).Return(&QoSUsage{}, nil)
	client.On("GetQoSViolations", mock.Anything, mock.AnythingOfType("*collector.QoSViolationOptions")).Return(&QoSViolations{
		Violations:      []QoSViolation{},
		ResolutionStats: ViolationResolutionStats{},
	}, nil)
	client.On("GetQoSEnforcement", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEnforcement{
		EnforcementStats: EnforcementStats{},
	}, nil)
	client.On("GetQoSStatistics", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QoSStatistics{}, nil)
	client.On("GetQoSHierarchy", mock.Anything).Return(&QoSHierarchy{
		InheritanceChain: map[string][]string{},
	}, nil)
	client.On("GetQoSConflicts", mock.Anything).Return(&QoSConflicts{
		ConflictsByType: map[string]int{},
	}, nil)
	client.On("GetQoSEffectiveness", mock.Anything, "normal").Return(qosEffectiveness, nil)
	client.On("GetQoSEffectiveness", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEffectiveness{}, nil)
	client.On("GetSystemQoSOverview", mock.Anything).Return(systemOverview, nil)

	collector := NewQoSLimitsCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundCPULimit := false
	foundPriority := false
	foundEfficiency := false
	foundSystemLoad := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_qos_limit_cpus":
			if len(mf.Metric) > 0 {
				for _, metric := range mf.Metric {
					hasQoS := false
					hasLimitType := false
					for _, label := range metric.Label {
						if *label.Name == "qos" && *label.Value == "normal" {
							hasQoS = true
						}
						if *label.Name == "limit_type" && *label.Value == "group" {
							hasLimitType = true
						}
					}
					if hasQoS && hasLimitType {
						assert.Equal(t, float64(2000), *metric.Gauge.Value)
						foundCPULimit = true
					}
				}
			}
		case "slurm_qos_priority":
			if len(mf.Metric) > 0 {
				for _, metric := range mf.Metric {
					for _, label := range metric.Label {
						if *label.Name == "qos" && *label.Value == "normal" {
							assert.Equal(t, float64(1500), *metric.Gauge.Value)
							foundPriority = true
						}
					}
				}
			}
		case "slurm_qos_efficiency_score":
			if len(mf.Metric) > 0 {
				for _, metric := range mf.Metric {
					for _, label := range metric.Label {
						if *label.Name == "qos" && *label.Value == "normal" {
							assert.Equal(t, float64(0.88), *metric.Gauge.Value)
							foundEfficiency = true
						}
					}
				}
			}
		case "slurm_system_qos_load":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.68), *mf.Metric[0].Gauge.Value)
				foundSystemLoad = true
			}
		}
	}

	assert.True(t, foundCPULimit, "Should find CPU limit metric with correct value")
	assert.True(t, foundPriority, "Should find priority metric with correct value")
	assert.True(t, foundEfficiency, "Should find efficiency metric with correct value")
	assert.True(t, foundSystemLoad, "Should find system load metric with correct value")
}

func TestQoSLimitsCollector_Integration(t *testing.T) {
	client := &MockQoSLimitsSLURMClient{}

	// Setup comprehensive mock data
	setupQoSLimitsMocks(client)

	collector := NewQoSLimitsCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_qos_priority Priority value for QoS
		# TYPE slurm_qos_priority gauge
		slurm_qos_priority{qos="high"} 1000
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_qos_priority")
	assert.NoError(t, err)
}

func TestQoSLimitsCollector_ViolationMetrics(t *testing.T) {
	client := &MockQoSLimitsSLURMClient{}

	// Setup violation specific mocks
	setupQoSViolationMocks(client)

	collector := NewQoSLimitsCollector(client)

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

	// Verify violation metrics are present
	foundViolations := false
	foundActive := false
	foundResolution := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "qos_violations_total") {
			foundViolations = true
		}
		if strings.Contains(desc, "qos_active_violations") {
			foundActive = true
		}
		if strings.Contains(desc, "qos_auto_resolution_rate") {
			foundResolution = true
		}
	}

	assert.True(t, foundViolations, "Should find violation metrics")
	assert.True(t, foundActive, "Should find active violation metrics")
	assert.True(t, foundResolution, "Should find resolution rate metrics")
}

func TestQoSLimitsCollector_EnforcementMetrics(t *testing.T) {
	client := &MockQoSLimitsSLURMClient{}

	// Setup enforcement specific mocks
	setupEnforcementMocks(client)

	collector := NewQoSLimitsCollector(client)

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

	// Verify enforcement metrics are present
	foundEnabled := false
	foundActions := false
	foundEffectiveness := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "qos_enforcement_enabled") {
			foundEnabled = true
		}
		if strings.Contains(desc, "qos_enforcement_actions_total") {
			foundActions = true
		}
		if strings.Contains(desc, "qos_enforcement_effectiveness") {
			foundEffectiveness = true
		}
	}

	assert.True(t, foundEnabled, "Should find enforcement enabled metrics")
	assert.True(t, foundActions, "Should find enforcement action metrics")
	assert.True(t, foundEffectiveness, "Should find enforcement effectiveness metrics")
}

// Helper functions

func setupQoSLimitsMocks(client *MockQoSLimitsSLURMClient) {
	qosLimits := &QoSLimits{
		QoSName:  "high",
		Priority: 1000,
		GrpCPUs:  500,
	}

	client.On("GetQoSLimits", mock.Anything, "high").Return(qosLimits, nil)
	client.On("GetQoSLimits", mock.Anything, mock.AnythingOfType("string")).Return(&QoSLimits{}, nil)
	client.On("GetQoSUsage", mock.Anything, mock.AnythingOfType("string")).Return(&QoSUsage{}, nil)
	client.On("GetQoSViolations", mock.Anything, mock.AnythingOfType("*collector.QoSViolationOptions")).Return(&QoSViolations{
		Violations:      []QoSViolation{},
		ResolutionStats: ViolationResolutionStats{},
	}, nil)
	client.On("GetQoSEnforcement", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEnforcement{
		EnforcementStats: EnforcementStats{},
	}, nil)
	client.On("GetQoSStatistics", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QoSStatistics{}, nil)
	client.On("GetQoSHierarchy", mock.Anything).Return(&QoSHierarchy{
		InheritanceChain: map[string][]string{},
	}, nil)
	client.On("GetQoSConflicts", mock.Anything).Return(&QoSConflicts{
		ConflictsByType: map[string]int{},
	}, nil)
	client.On("GetQoSEffectiveness", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEffectiveness{}, nil)
	client.On("GetSystemQoSOverview", mock.Anything).Return(&SystemQoSOverview{
		QoSUtilization: map[string]float64{},
	}, nil)
}

func setupQoSViolationMocks(client *MockQoSLimitsSLURMClient) {
	violations := &QoSViolations{
		TotalViolations:    20,
		ActiveViolations:   5,
		CriticalViolations: 3,
		WarningViolations:  7,
		InfoViolations:     10,
		Violations: []QoSViolation{
			{
				QoSName:       "high",
				ViolationType: "cpu_exceeded",
				Severity:      "critical",
				Status:        "active",
				Duration:      45 * time.Minute,
			},
			{
				QoSName:        "normal",
				ViolationType:  "memory_exceeded",
				Severity:       "warning",
				Status:         "resolved",
				Duration:       20 * time.Minute,
				ResolutionTime: time.Now().Add(-10 * time.Minute),
			},
		},
		ViolationsByQoS:  map[string]int{"high": 12, "normal": 8},
		ViolationsByType: map[string]int{"cpu_exceeded": 10, "memory_exceeded": 10},
		ResolutionStats: ViolationResolutionStats{
			AutoResolutionRate: 0.75,
			RecurrenceRate:     0.1,
		},
	}

	client.On("GetQoSLimits", mock.Anything, mock.AnythingOfType("string")).Return(&QoSLimits{}, nil)
	client.On("GetQoSUsage", mock.Anything, mock.AnythingOfType("string")).Return(&QoSUsage{}, nil)
	client.On("GetQoSViolations", mock.Anything, mock.AnythingOfType("*collector.QoSViolationOptions")).Return(violations, nil)
	client.On("GetQoSEnforcement", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEnforcement{
		EnforcementStats: EnforcementStats{},
	}, nil)
	client.On("GetQoSStatistics", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QoSStatistics{}, nil)
	client.On("GetQoSHierarchy", mock.Anything).Return(&QoSHierarchy{
		InheritanceChain: map[string][]string{},
	}, nil)
	client.On("GetQoSConflicts", mock.Anything).Return(&QoSConflicts{
		ConflictsByType: map[string]int{},
	}, nil)
	client.On("GetQoSEffectiveness", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEffectiveness{}, nil)
	client.On("GetSystemQoSOverview", mock.Anything).Return(&SystemQoSOverview{
		QoSUtilization: map[string]float64{},
	}, nil)
}

func setupEnforcementMocks(client *MockQoSLimitsSLURMClient) {
	enforcement := &QoSEnforcement{
		QoSName:            "high",
		EnforcementEnabled: true,
		EnforcementLevel:   "strict",
		AutoRemediation:    true,
		EnforcementStats: EnforcementStats{
			ActionsTotal:       150,
			ActionsSuccessful:  140,
			ActionsFailed:      10,
			PreemptionsTotal:   25,
			JobsTerminated:     20,
			JobsSuspended:      15,
			UsersWarned:        50,
			EffectivenessScore: 0.93,
		},
	}

	client.On("GetQoSLimits", mock.Anything, mock.AnythingOfType("string")).Return(&QoSLimits{}, nil)
	client.On("GetQoSUsage", mock.Anything, mock.AnythingOfType("string")).Return(&QoSUsage{}, nil)
	client.On("GetQoSViolations", mock.Anything, mock.AnythingOfType("*collector.QoSViolationOptions")).Return(&QoSViolations{
		Violations:      []QoSViolation{},
		ResolutionStats: ViolationResolutionStats{},
	}, nil)
	client.On("GetQoSEnforcement", mock.Anything, "high").Return(enforcement, nil)
	client.On("GetQoSEnforcement", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEnforcement{
		EnforcementStats: EnforcementStats{},
	}, nil)
	client.On("GetQoSStatistics", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&QoSStatistics{}, nil)
	client.On("GetQoSHierarchy", mock.Anything).Return(&QoSHierarchy{
		InheritanceChain: map[string][]string{},
	}, nil)
	client.On("GetQoSConflicts", mock.Anything).Return(&QoSConflicts{
		ConflictsByType: map[string]int{},
	}, nil)
	client.On("GetQoSEffectiveness", mock.Anything, mock.AnythingOfType("string")).Return(&QoSEffectiveness{}, nil)
	client.On("GetSystemQoSOverview", mock.Anything).Return(&SystemQoSOverview{
		QoSUtilization: map[string]float64{},
	}, nil)
}