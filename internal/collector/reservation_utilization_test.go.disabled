package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReservationUtilizationSLURMClient struct {
	mock.Mock
}

func (m *MockReservationUtilizationSLURMClient) GetReservationUtilization(ctx context.Context, reservationName string) (*ReservationUtilizationMetrics, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationUtilizationMetrics), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationInfo(ctx context.Context, reservationName string) (*ReservationInfo, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationInfo), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationList(ctx context.Context, filter *ReservationFilter) (*ReservationList, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationList), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationStatistics(ctx context.Context, reservationName string, period string) (*ReservationStatistics, error) {
	args := m.Called(ctx, reservationName, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationStatistics), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationEfficiency(ctx context.Context, reservationName string) (*ReservationEfficiency, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationEfficiency), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationConflicts(ctx context.Context) (*ReservationConflicts, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationConflicts), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationAlerts(ctx context.Context, alertType string) (*ReservationAlerts, error) {
	args := m.Called(ctx, alertType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationAlerts), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationTrends(ctx context.Context, reservationName string, period string) (*ReservationTrends, error) {
	args := m.Called(ctx, reservationName, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationTrends), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationPredictions(ctx context.Context, reservationName string) (*ReservationPredictions, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationPredictions), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationAnalytics(ctx context.Context, reservationName string) (*ReservationAnalytics, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationAnalytics), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetSystemReservationOverview(ctx context.Context) (*SystemReservationOverview, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SystemReservationOverview), args.Error(1)
}

// Remove mock types that are no longer needed

func (m *MockReservationUtilizationSLURMClient) GetReservationOptimization(ctx context.Context, reservationName string) (*ReservationOptimization, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationOptimization), args.Error(1)
}

// Methods removed as they are not part of the interface

func TestNewReservationUtilizationCollector(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
}

func TestReservationUtilizationCollector_Describe(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 50, count)
}

func TestReservationUtilizationCollector_Collect_Success(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	reservations := []*ReservationInfo{
		{
			Name:      "gpu_reservation",
			State:     "ACTIVE",
			StartTime: startTime,
			EndTime:   endTime,
			Nodes:     []string{"gpu01", "gpu02"},
			Users:     []string{"user1", "user2"},
			Accounts:  []string{"research", "ml"},
			Features:  []string{"gpu", "high_memory"},
			NodeCount: 2,
		},
		{
			Name:      "cpu_reservation",
			State:     "INACTIVE",
			StartTime: startTime,
			EndTime:   endTime,
			Nodes:     []string{"cpu01", "cpu02", "cpu03"},
			Users:     []string{"user3"},
			Accounts:  []string{"compute"},
			Features:  []string{"cpu", "standard"},
			NodeCount: 3,
		},
	}

	utilization := &ReservationUtilization{
		ReservationName:  "gpu_reservation",
		TotalCPUs:        128,
		UsedCPUs:         96,
		TotalMemoryMB:    512000,
		UsedMemoryMB:     384000,
		TotalGPUs:        8,
		UsedGPUs:         6,
		TotalNodes:       2,
		UsedNodes:        2,
		ActiveJobs:       12,
		QueuedJobs:       3,
		UtilizationRate:  0.75,
		EfficiencyScore:  0.85,
		LastUpdated:      time.Now(),
	}

	statistics := &ReservationStatistics{
		ReservationName:    "gpu_reservation",
		TotalJobsRun:       245,
		SuccessfulJobs:     230,
		FailedJobs:         15,
		CancelledJobs:      8,
		AverageJobDuration: 3600,
		AverageWaitTime:    300,
		PeakUtilization:    0.95,
		LowestUtilization:  0.15,
		AverageUtilization: 0.72,
		TotalCPUHours:      15680,
		TotalGPUHours:      3920,
		WastedCPUHours:     1240,
		WastedGPUHours:     310,
		OversubscriptionEvents: 5,
		UndersubscriptionEvents: 23,
	}

	efficiency := &ReservationEfficiency{
		ReservationName:        "gpu_reservation",
		OverallEfficiency:      0.85,
		ResourceEfficiency:     0.78,
		TimeEfficiency:         0.75,
		EnergyEfficiency:       0.68,
		CostEfficiency:         0.71,
		OptimizationPotential:  0.12,
		EfficiencyBreakdown:    map[string]float64{"cpu": 0.88, "memory": 0.75, "gpu": 0.92},
	}

	conflicts := &ReservationConflicts{
		TotalConflicts:     5,
		ActiveConflicts:    2,
		ResolvedConflicts:  3,
		Conflicts: []ReservationConflict{
			{
				ConflictID:          "conflict-001",
				Type:                "RESOURCE_OVERLAP",
				Severity:            "HIGH",
				AffectedReservations: []string{"gpu_reservation", "maintenance"},
				Description:         "GPU resources overlap with maintenance window",
				DetectedTime:        time.Now().Add(-2 * time.Hour),
				ResolutionTime:      time.Now(),
				Status:              "resolved",
			},
		},
		ConflictsByType: map[string]int{
			"RESOURCE_OVERLAP": 3,
			"TIME_CONFLICT":    2,
		},
		LastAnalyzed: time.Now(),
	}

	alerts := &ReservationAlerts{
		AlertType:       "all",
		TotalAlerts:     10,
		ActiveAlerts:    3,
		ResolvedAlerts:  7,
		Alerts: []ReservationAlert{
			{
				AlertID:         "alert-001",
				ReservationName: "gpu_reservation",
				AlertType:       "LOW_UTILIZATION",
				Severity:        "WARNING",
				Timestamp:       time.Now().Add(-time.Hour),
				Message:         "Reservation utilization below threshold",
				Threshold:       0.5,
				ActualValue:     0.35,
				Status:          "active",
			},
			{
				AlertID:         "alert-002",
				ReservationName: "gpu_reservation",
				AlertType:       "HIGH_COST",
				Severity:        "INFO",
				Timestamp:       time.Now().Add(-2 * time.Hour),
				ResolvedAt:      time.Now().Add(-30 * time.Minute),
				Status:          "resolved",
			},
		},
		ResolutionStats: AlertResolutionStats{
			FalsePositiveRate: 0.1,
		},
	}

	trends := &ReservationTrends{
		ReservationName:      "gpu_reservation",
		Period:               "7d",
		UtilizationTrend:     []float64{0.70, 0.72, 0.74, 0.75, 0.76, 0.78, 0.80},
		EfficiencyTrend:      []float64{0.82, 0.83, 0.83, 0.84, 0.85, 0.85, 0.86},
		ForecastAccuracy:     0.85,
		ConfidenceInterval:   0.1,
	}

	predictions := &ReservationPredictions{
		ReservationName:       "gpu_reservation",
		PredictionHorizon:     24 * time.Hour,
		UtilizationForecast:   []float64{0.80, 0.82, 0.83},
		EfficiencyForecast:    []float64{0.86, 0.87, 0.88},
		ConfidenceLevel:       0.85,
		ModelAccuracy:         0.82,
		LastModelUpdate:       time.Now().Add(-time.Hour),
	}

	optimization := &ReservationOptimization{
		ReservationName: "gpu_reservation",
		OptimizationPotential: OptimizationPotential{
			EfficiencyGain:   0.10,
			CostSavings:      1250.50,
			PerformanceGain:  0.15,
		},
	}

	// Mock analytics
	analytics := &ReservationAnalytics{
		ReservationName:    "gpu_reservation",
		AnalysisDate:       time.Now(),
		CostAnalysis: ReservationCostAnalysis{
			TotalCost: 10000.0,
		},
	}

	// Mock system overview
	systemOverview := &SystemReservationOverview{
		TotalReservations:    10,
		ActiveReservations:   6,
		PlannedReservations:  2,
		ExpiredReservations:  2,
		CancelledReservations: 0,
		TotalNodes:           100,
		ReservedNodes:        60,
		AvailableNodes:       40,
		ReservationUtilization: 0.75,
		SystemEfficiency:     0.85,
		TotalCost:            50000.0,
		CostPerHour:          250.0,
		OptimizationScore:    0.8,
		SystemHealth:         "good",
		LastUpdated:          time.Now(),
	}

	client.On("GetReservationList", mock.Anything, mock.Anything).Return(reservationList, nil)
	client.On("GetReservationUtilization", mock.Anything, "gpu_reservation").Return(utilization, nil)
	client.On("GetReservationUtilization", mock.Anything, "bigmem").Return(nil, errors.New("not found"))
	client.On("GetReservationUtilization", mock.Anything, "debug").Return(nil, errors.New("not found"))
	client.On("GetReservationUtilization", mock.Anything, "maintenance").Return(nil, errors.New("not found"))
	client.On("GetReservationStatistics", mock.Anything, "gpu_reservation", "24h").Return(statistics, nil)
	client.On("GetReservationStatistics", mock.Anything, "bigmem", "24h").Return(nil, errors.New("not found"))
	client.On("GetReservationStatistics", mock.Anything, "debug", "24h").Return(nil, errors.New("not found"))
	client.On("GetReservationStatistics", mock.Anything, "maintenance", "24h").Return(nil, errors.New("not found"))
	client.On("GetReservationEfficiency", mock.Anything, "gpu_reservation").Return(efficiency, nil)
	client.On("GetReservationEfficiency", mock.Anything, "bigmem").Return(nil, errors.New("not found"))
	client.On("GetReservationEfficiency", mock.Anything, "debug").Return(nil, errors.New("not found"))
	client.On("GetReservationEfficiency", mock.Anything, "maintenance").Return(nil, errors.New("not found"))
	client.On("GetReservationConflicts", mock.Anything).Return(conflicts, nil)
	client.On("GetReservationAlerts", mock.Anything, "all").Return(alerts, nil)
	client.On("GetReservationTrends", mock.Anything, "gpu_reservation", "7d").Return(trends, nil)
	client.On("GetReservationTrends", mock.Anything, "bigmem", "7d").Return(nil, errors.New("not found"))
	client.On("GetReservationTrends", mock.Anything, "debug", "7d").Return(nil, errors.New("not found"))
	client.On("GetReservationPredictions", mock.Anything, "gpu_reservation").Return(predictions, nil)
	client.On("GetReservationPredictions", mock.Anything, "bigmem").Return(nil, errors.New("not found"))
	client.On("GetReservationPredictions", mock.Anything, "debug").Return(nil, errors.New("not found"))
	client.On("GetReservationAnalytics", mock.Anything, "gpu_reservation").Return(analytics, nil)
	client.On("GetReservationAnalytics", mock.Anything, "bigmem").Return(nil, errors.New("not found"))
	client.On("GetReservationAnalytics", mock.Anything, "debug").Return(nil, errors.New("not found"))
	client.On("GetReservationOptimization", mock.Anything, "gpu_reservation").Return(optimization, nil)
	client.On("GetReservationOptimization", mock.Anything, "bigmem").Return(nil, errors.New("not found"))
	client.On("GetReservationOptimization", mock.Anything, "debug").Return(nil, errors.New("not found"))
	client.On("GetSystemReservationOverview", mock.Anything).Return(systemOverview, nil)

	ch := make(chan prometheus.Metric, 100)
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

func TestReservationUtilizationCollector_Collect_ListReservationsError(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	client.On("GetReservationList", mock.Anything, mock.Anything).Return(nil, errors.New("API error"))

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

func TestReservationUtilizationCollector_CollectUtilizationMetrics(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	utilization := &ReservationUtilization{
		ReservationName:  "test_reservation",
		TotalCPUs:        64,
		UsedCPUs:        48,
		TotalMemoryMB:   256000,
		UsedMemoryMB:    192000,
		TotalGPUs:       4,
		UsedGPUs:        3,
		TotalNodes:      4,
		UsedNodes:       3,
		ActiveJobs:      10,
		QueuedJobs:      2,
		UtilizationRate: 0.75,
		EfficiencyScore: 0.85,
		LastUpdated:     time.Now(),
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectUtilizationMetrics(ch, "test_reservation", utilization)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 12, count)
}

func TestReservationUtilizationCollector_CollectStatisticsMetrics(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	statistics := &ReservationStatistics{
		ReservationName:    "test_reservation",
		TotalJobsRun:       100,
		SuccessfulJobs:     95,
		FailedJobs:         5,
		CancelledJobs:      2,
		AverageJobDuration: 3600,
		AverageWaitTime:    300,
		PeakUtilization:    0.95,
		LowestUtilization:  0.15,
		AverageUtilization: 0.75,
		TotalCPUHours:      7200,
		TotalGPUHours:      1800,
		WastedCPUHours:     720,
		WastedGPUHours:     180,
		OversubscriptionEvents: 3,
		UndersubscriptionEvents: 15,
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectStatisticsMetrics(ch, "test_reservation", statistics)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 16, count)
}

func TestReservationUtilizationCollector_CollectEfficiencyMetrics(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	efficiency := &ReservationEfficiency{
		ReservationName:        "test_reservation",
		CPUEfficiency:          0.88,
		MemoryEfficiency:       0.75,
		GPUEfficiency:          0.92,
		NetworkEfficiency:      0.68,
		StorageEfficiency:      0.71,
		OverallEfficiency:      0.85,
		ResourceWastePercent:   0.15,
		OptimizationPotential:  0.12,
		BottleneckResources:    []string{"memory", "network"},
		OptimizationSuggestions: []string{"increase_memory", "optimize_network"},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectEfficiencyMetrics(ch, "test_reservation", efficiency)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 8, count)
}

func TestReservationUtilizationCollector_CollectConflictMetrics(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	conflicts := []*ReservationConflict{
		{
			ReservationName:      "test_reservation",
			ConflictType:         "RESOURCE_OVERLAP",
			ConflictDescription:  "Resource overlap detected",
			ConflictSeverity:     "HIGH",
			ConflictStartTime:    time.Now(),
			ConflictEndTime:      time.Now().Add(2 * time.Hour),
			AffectedResources:    []string{"node01", "node02"},
			ResolutionRequired:   true,
			ResolutionSuggestion: "Reschedule conflicting reservation",
		},
		{
			ReservationName:      "test_reservation",
			ConflictType:         "TIME_OVERLAP",
			ConflictDescription:  "Time overlap with maintenance",
			ConflictSeverity:     "MEDIUM",
			ConflictStartTime:    time.Now().Add(4 * time.Hour),
			ConflictEndTime:      time.Now().Add(6 * time.Hour),
			AffectedResources:    []string{"node03"},
			ResolutionRequired:   false,
			ResolutionSuggestion: "Monitor during maintenance window",
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectConflictMetrics(ch, "test_reservation", conflicts)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 4, count)
}

func TestReservationUtilizationCollector_DescribeAll(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	gatherer := prometheus.Gatherer(registry)
	metricFamilies, err := gatherer.Gather()

	assert.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
}

func TestReservationUtilizationCollector_MetricCount(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	expectedMetrics := []string{
		"slurm_reservation_total_cpus",
		"slurm_reservation_used_cpus",
		"slurm_reservation_total_memory_bytes",
		"slurm_reservation_used_memory_bytes",
		"slurm_reservation_total_gpus",
		"slurm_reservation_used_gpus",
		"slurm_reservation_total_nodes",
		"slurm_reservation_used_nodes",
		"slurm_reservation_active_jobs",
		"slurm_reservation_queued_jobs",
		"slurm_reservation_utilization_rate",
		"slurm_reservation_efficiency_score",
		"slurm_reservation_jobs_total",
		"slurm_reservation_jobs_successful",
		"slurm_reservation_jobs_failed",
		"slurm_reservation_jobs_cancelled",
		"slurm_reservation_job_duration_seconds",
		"slurm_reservation_job_wait_time_seconds",
		"slurm_reservation_utilization_peak",
		"slurm_reservation_utilization_lowest",
		"slurm_reservation_utilization_average",
		"slurm_reservation_cpu_hours_total",
		"slurm_reservation_gpu_hours_total",
		"slurm_reservation_cpu_hours_wasted",
		"slurm_reservation_gpu_hours_wasted",
		"slurm_reservation_oversubscription_events",
		"slurm_reservation_undersubscription_events",
		"slurm_reservation_cpu_efficiency",
		"slurm_reservation_memory_efficiency",
		"slurm_reservation_gpu_efficiency",
		"slurm_reservation_network_efficiency",
		"slurm_reservation_storage_efficiency",
		"slurm_reservation_overall_efficiency",
		"slurm_reservation_resource_waste_percent",
		"slurm_reservation_optimization_potential",
		"slurm_reservation_conflicts_total",
		"slurm_reservation_conflicts_resolution_required",
		"slurm_reservation_conflicts_by_severity",
		"slurm_reservation_conflicts_by_type",
		"slurm_reservation_alerts_total",
		"slurm_reservation_alerts_active",
		"slurm_reservation_predicted_utilization",
		"slurm_reservation_predicted_efficiency",
		"slurm_reservation_trend_confidence",
		"slurm_reservation_forecast_accuracy",
		"slurm_reservation_anomaly_detected",
		"slurm_reservation_optimal_size",
		"slurm_reservation_overlaps_total",
		"slurm_reservation_overlap_conflict_potential",
		"slurm_reservation_optimization_potential_score",
		"slurm_reservation_estimated_cost_savings",
	}

	assert.Equal(t, len(expectedMetrics), 50)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 50, count)
}