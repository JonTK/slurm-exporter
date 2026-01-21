package collector

import (
	"context"
	"log/slog"
	"math"
	"os"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewEnergyMonitor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tests := []struct {
		name   string
		config *EnergyConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &EnergyConfig{
				MonitoringInterval:         30 * time.Second,
				MaxJobsPerCollection:       50,
				EnableCarbonTracking:       false,
				EnableEfficiencyTracking:   false,
				EnergyCalculationMethod:    "estimated",
				DefaultCPUPowerWatts:       20.0,
				DefaultMemoryPowerWatts:    2.0,
				DefaultGPUPowerWatts:       300.0,
				DefaultNodeBasePowerWatts:  60.0,
				CarbonIntensityGCO2kWh:     350.0,
				CarbonTrackingRegion:       "EU",
				EnableRealTimeCarbonAPI:    true,
				EnablePowerCapping:         true,
				PowerCapThresholdWatts:     800.0,
				EnergyBudgetPerJobWh:       500.0,
				EnergyDataRetention:        3 * 24 * time.Hour,
				CarbonDataRetention:        14 * 24 * time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, err := NewEnergyMonitor(mockClient, logger, tt.config)
			require.NoError(t, err)
			assert.NotNil(t, monitor)
			assert.NotNil(t, monitor.config)
			assert.NotNil(t, monitor.metrics)
			assert.NotNil(t, monitor.energyData)
			assert.NotNil(t, monitor.carbonData)
			assert.NotNil(t, monitor.clusterEnergy)
			assert.NotNil(t, monitor.carbonCalculator)
			assert.NotNil(t, monitor.efficiencyTracker)

			if tt.config == nil {
				assert.Equal(t, 60*time.Second, monitor.config.MonitoringInterval)
				assert.Equal(t, 100, monitor.config.MaxJobsPerCollection)
				assert.True(t, monitor.config.EnableCarbonTracking)
				assert.True(t, monitor.config.EnableEfficiencyTracking)
				assert.Equal(t, "power_model", monitor.config.EnergyCalculationMethod)
				assert.Equal(t, 15.0, monitor.config.DefaultCPUPowerWatts)
			} else {
				assert.Equal(t, tt.config.MonitoringInterval, monitor.config.MonitoringInterval)
				assert.Equal(t, tt.config.MaxJobsPerCollection, monitor.config.MaxJobsPerCollection)
				assert.Equal(t, tt.config.EnableCarbonTracking, monitor.config.EnableCarbonTracking)
				assert.Equal(t, tt.config.EnableEfficiencyTracking, monitor.config.EnableEfficiencyTracking)
				assert.Equal(t, tt.config.EnergyCalculationMethod, monitor.config.EnergyCalculationMethod)
				assert.Equal(t, tt.config.DefaultCPUPowerWatts, monitor.config.DefaultCPUPowerWatts)
			}
		})
	}
}

func TestEnergyMonitor_Describe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 100)
	monitor.Describe(ch)
	close(ch)

	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	// Verify we have the expected number of metric descriptions
	assert.GreaterOrEqual(t, len(descs), 25)
}

func TestEnergyMonitor_CollectEnergyMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, &EnergyConfig{
		MonitoringInterval:       60 * time.Second,
		MaxJobsPerCollection:     25,
		EnableCarbonTracking:     true,
		EnableEfficiencyTracking: true,
		EnergyDataRetention:      24 * time.Hour,
		CarbonDataRetention:      7 * 24 * time.Hour,
	})
	require.NoError(t, err)

	// Setup mock expectations
	now := time.Now()
	startTime := now.Add(-3 * time.Hour)
	endTime := now.Add(-30 * time.Minute)

	testJob := &slurm.Job{
		JobID:      "energy-123",
		Name:       "energy-test-job",
		UserName:   "energyuser",
		Account:    "energyaccount",
		Partition:  "compute",
		JobState:   "COMPLETED",
		CPUs:       8,
		Memory:     16384, // 16GB in MB
		Nodes:      2,
		TimeLimit:  240, // 4 hours
		StartTime:  &startTime,
		EndTime:    &endTime,
	}

	jobList := &slurm.JobList{
		Jobs: []*slurm.Job{testJob},
	}

	mockJobManager.On("List", mock.Anything, mock.MatchedBy(func(opts *slurm.ListJobsOptions) bool {
		return len(opts.States) >= 2 && opts.MaxCount == 25
	})).Return(jobList, nil)

	// Test energy metrics collection
	ctx := context.Background()
	err = monitor.collectEnergyMetrics(ctx)
	require.NoError(t, err)

	// Verify energy data was collected
	energyData, exists := monitor.GetEnergyData("energy-123")
	assert.True(t, exists)
	assert.NotNil(t, energyData)
	assert.Equal(t, "energy-123", energyData.JobID)
	assert.WithinDuration(t, time.Now(), energyData.Timestamp, 5*time.Second)

	// Verify energy consumption metrics
	assert.Greater(t, energyData.TotalEnergyWh, 0.0)
	assert.Greater(t, energyData.CPUEnergyWh, 0.0)
	assert.Greater(t, energyData.MemoryEnergyWh, 0.0)
	assert.GreaterOrEqual(t, energyData.GPUEnergyWh, 0.0) // May be 0 if no GPU
	assert.Greater(t, energyData.StorageEnergyWh, 0.0)
	assert.Greater(t, energyData.NetworkEnergyWh, 0.0)
	assert.Greater(t, energyData.CoolingEnergyWh, 0.0)

	// Verify power metrics
	assert.Greater(t, energyData.AveragePowerW, 0.0)
	assert.GreaterOrEqual(t, energyData.PeakPowerW, energyData.AveragePowerW)
	assert.GreaterOrEqual(t, energyData.PowerEfficiency, 0.0)
	assert.LessOrEqual(t, energyData.PowerEfficiency, 1.0)

	// Verify efficiency metrics
	assert.Greater(t, energyData.EnergyPerCPUHour, 0.0)
	assert.Greater(t, energyData.EnergyPerTaskUnit, 0.0)
	assert.GreaterOrEqual(t, energyData.EnergyWasteWh, 0.0)

	// Verify cost metrics
	assert.Greater(t, energyData.EnergyCostUSD, 0.0)
	assert.Greater(t, energyData.CostPerCPUHour, 0.0)
	assert.GreaterOrEqual(t, energyData.CostEfficiency, 0.0)
	assert.LessOrEqual(t, energyData.CostEfficiency, 1.0)

	// Verify carbon data was collected
	carbonData, exists := monitor.GetCarbonData("energy-123")
	assert.True(t, exists)
	assert.NotNil(t, carbonData)
	assert.Equal(t, "energy-123", carbonData.JobID)

	// Verify carbon footprint metrics
	assert.Greater(t, carbonData.TotalCO2grams, 0.0)
	assert.Greater(t, carbonData.CO2PerCPUHour, 0.0)
	assert.Greater(t, carbonData.CO2PerTaskUnit, 0.0)
	assert.Greater(t, carbonData.CarbonIntensity, 0.0)
	assert.GreaterOrEqual(t, carbonData.CarbonEfficiency, 0.0)
	assert.LessOrEqual(t, carbonData.CarbonEfficiency, 1.0)
	assert.GreaterOrEqual(t, carbonData.CarbonWasteGrams, 0.0)
	assert.GreaterOrEqual(t, carbonData.EnvironmentalImpactScore, 0.0)
	assert.LessOrEqual(t, carbonData.EnvironmentalImpactScore, 1.0)
	assert.Contains(t, []string{"A", "B", "C", "D", "F"}, carbonData.SustainabilityGrade)
	assert.GreaterOrEqual(t, carbonData.GreenEnergyPercentage, 0.0)
	assert.LessOrEqual(t, carbonData.GreenEnergyPercentage, 1.0)

	// Verify cluster energy data
	clusterData := monitor.GetClusterEnergyData()
	assert.NotNil(t, clusterData)
	assert.Greater(t, clusterData.TotalClusterEnergyWh, 0.0)
	assert.Greater(t, clusterData.TotalClusterPowerW, 0.0)
	assert.Greater(t, clusterData.TotalClusterCO2grams, 0.0)

	mockJobManager.AssertExpectations(t)
}

func TestEnergyMonitor_EnergyCalculations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, &EnergyConfig{
		DefaultCPUPowerWatts:      20.0,
		DefaultMemoryPowerWatts:   2.0,
		DefaultGPUPowerWatts:      300.0,
		DefaultNodeBasePowerWatts: 60.0,
	})
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-2 * time.Hour)
	endTime := now.Add(-10 * time.Minute)

	testJob := &slurm.Job{
		JobID:      "calc-456",
		Name:       "calculation-job",
		UserName:   "calcuser",
		Account:    "calcaccount",
		Partition:  "compute",
		JobState:   "COMPLETED",
		CPUs:       4,
		Memory:     8192, // 8GB
		Nodes:      1,
		TimeLimit:  120, // 2 hours
		StartTime:  &startTime,
		EndTime:    &endTime,
	}

	// Test energy calculation
	energyData := monitor.calculateJobEnergyConsumption(testJob)
	assert.NotNil(t, energyData)
	assert.Equal(t, "calc-456", energyData.JobID)

	// Verify runtime calculation (should be close to 1.83 hours)
	expectedRuntimeHours := endTime.Sub(startTime).Hours()
	assert.InDelta(t, 1.83, expectedRuntimeHours, 0.1)

	// Test individual energy components
	runtimeHours := expectedRuntimeHours

	// Test CPU energy calculation
	cpuEnergy := monitor.calculateCPUEnergy(testJob, runtimeHours)
	expectedCPUEnergy := 4.0 * 20.0 * runtimeHours * monitor.estimateCPUUtilization(testJob)
	assert.InDelta(t, expectedCPUEnergy, cpuEnergy, expectedCPUEnergy*0.1) // Allow 10% tolerance

	// Test memory energy calculation
	memoryEnergy := monitor.calculateMemoryEnergy(testJob, runtimeHours)
	expectedMemoryEnergy := 8.0 * 2.0 * runtimeHours // 8GB * 2W/GB * runtime
	assert.InDelta(t, expectedMemoryEnergy, memoryEnergy, expectedMemoryEnergy*0.1)

	// Test GPU energy calculation (should be 0 for this job)
	gpuEnergy := monitor.calculateGPUEnergy(testJob, runtimeHours)
	assert.Equal(t, 0.0, gpuEnergy) // This job shouldn't have GPUs

	// Test storage and network energy
	storageEnergy := monitor.calculateStorageEnergy(testJob, runtimeHours)
	assert.Greater(t, storageEnergy, 0.0)

	networkEnergy := monitor.calculateNetworkEnergy(testJob, runtimeHours)
	assert.Greater(t, networkEnergy, 0.0)

	// Test total energy is sum of components plus base power and cooling
	basePower := 60.0 * 1.0 * runtimeHours // 60W base power for 1 node
	computeEnergy := cpuEnergy + memoryEnergy + gpuEnergy
	coolingEnergy := computeEnergy * 0.3
	expectedTotalEnergy := basePower + computeEnergy + storageEnergy + networkEnergy + coolingEnergy

	assert.InDelta(t, expectedTotalEnergy, energyData.TotalEnergyWh, expectedTotalEnergy*0.1)

	// Test power calculations
	expectedAveragePower := energyData.TotalEnergyWh / runtimeHours
	assert.InDelta(t, expectedAveragePower, energyData.AveragePowerW, expectedAveragePower*0.1)
	assert.Greater(t, energyData.PeakPowerW, energyData.AveragePowerW)

	// Test efficiency calculations
	assert.GreaterOrEqual(t, energyData.PowerEfficiency, 0.0)
	assert.LessOrEqual(t, energyData.PowerEfficiency, 1.0)
	assert.Greater(t, energyData.EnergyPerCPUHour, 0.0)
	assert.GreaterOrEqual(t, energyData.EnergyWasteWh, 0.0)

	// Test cost calculations
	expectedCost := (energyData.TotalEnergyWh / 1000.0) * 0.12 // $0.12/kWh
	assert.InDelta(t, expectedCost, energyData.EnergyCostUSD, expectedCost*0.1)
}

func TestEnergyMonitor_CarbonFootprintCalculation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, &EnergyConfig{
		CarbonIntensityGCO2kWh: 400.0, // 400g CO2/kWh
		CarbonTrackingRegion:   "US",
	})
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "carbon-789",
		Name:       "carbon-job",
		UserName:   "carbonuser",
		Account:    "carbonaccount",
		Partition:  "compute",
		JobState:   "COMPLETED",
		CPUs:       8,
		Memory:     16384,
		Nodes:      1,
		StartTime:  &startTime,
		EndTime:    &now,
	}

	// Create sample energy data
	energyData := &JobEnergyData{
		JobID:         "carbon-789",
		TotalEnergyWh: 1000.0, // 1kWh
		PowerEfficiency: 0.8,
	}

	// Test carbon footprint calculation
	carbonData := monitor.carbonCalculator.calculateCarbonFootprint(testJob, energyData)
	assert.NotNil(t, carbonData)
	assert.Equal(t, "carbon-789", carbonData.JobID)

	// Test basic carbon calculations
	expectedCO2 := (energyData.TotalEnergyWh / 1000.0) * 400.0 // 1kWh * 400g/kWh = 400g
	// Account for regional and time factors
	regionalFactor := monitor.carbonCalculator.regionalFactors["US"]
	timeFactor := monitor.carbonCalculator.timeFactors[now.Hour()]
	adjustedCO2 := expectedCO2 * regionalFactor * timeFactor

	assert.InDelta(t, adjustedCO2, carbonData.TotalCO2grams, adjustedCO2*0.1)

	// Test per-unit calculations
	runtimeHours := 1.0
	cpuHours := 8.0 * runtimeHours
	expectedCO2PerCPUHour := carbonData.TotalCO2grams / cpuHours
	assert.InDelta(t, expectedCO2PerCPUHour, carbonData.CO2PerCPUHour, expectedCO2PerCPUHour*0.1)

	// Test carbon intensity
	assert.Equal(t, 400.0*regionalFactor*timeFactor, carbonData.CarbonIntensity)

	// Test carbon efficiency
	assert.GreaterOrEqual(t, carbonData.CarbonEfficiency, 0.0)
	assert.LessOrEqual(t, carbonData.CarbonEfficiency, 1.0)

	// Test carbon waste
	assert.GreaterOrEqual(t, carbonData.CarbonWasteGrams, 0.0)
	assert.LessOrEqual(t, carbonData.CarbonWasteGrams, carbonData.TotalCO2grams)

	// Test sustainability metrics
	assert.GreaterOrEqual(t, carbonData.EnvironmentalImpactScore, 0.0)
	assert.LessOrEqual(t, carbonData.EnvironmentalImpactScore, 1.0)
	assert.Contains(t, []string{"A", "B", "C", "D", "F"}, carbonData.SustainabilityGrade)
	assert.GreaterOrEqual(t, carbonData.GreenEnergyPercentage, 0.0)
	assert.LessOrEqual(t, carbonData.GreenEnergyPercentage, 1.0)
}

func TestEnergyMonitor_UtilizationEstimation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name                 string
		cpus                 int
		memory               int
		nodes                int
		expectedCPUUtil      float64
		expectedGPUCount     int
		expectedIOIntensity  float64
		expectedNetIntensity float64
	}{
		{
			name:                 "small CPU job",
			cpus:                 4,
			memory:               8192,
			nodes:                1,
			expectedCPUUtil:      0.7, // Base utilization
			expectedGPUCount:     0,   // No GPU for small job
			expectedIOIntensity:  1.0, // Base I/O intensity
			expectedNetIntensity: 1.0, // Base network intensity
		},
		{
			name:                 "large CPU job",
			cpus:                 32,
			memory:               65536,
			nodes:                1,
			expectedCPUUtil:      0.8, // Higher utilization for large jobs
			expectedGPUCount:     4,   // Should estimate 4 GPUs (32/8)
			expectedIOIntensity:  1.5, // Higher I/O intensity
			expectedNetIntensity: 1.6, // Higher network intensity (1 + 0.3*2)
		},
		{
			name:                 "multi-node job",
			cpus:                 16,
			memory:               32768,
			nodes:                4,
			expectedCPUUtil:      0.7, // Base utilization
			expectedGPUCount:     2,   // Should estimate 2 GPUs
			expectedIOIntensity: 1.3,  // Higher I/O for multi-node
			expectedNetIntensity: 1.6, // Much higher network intensity (1 + 3*0.2)
		},
		{
			name:                 "memory-intensive job",
			cpus:                 8,
			memory:               131072, // 128GB - very high memory
			nodes:                1,
			expectedCPUUtil:      0.6, // Lower CPU utilization
			expectedGPUCount:     16,  // High GPU count due to memory
			expectedIOIntensity:  1.5, // Higher I/O intensity
			expectedNetIntensity: 1.0, // Base network intensity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJob := &slurm.Job{
				JobID:     "util-test-" + tt.name,
				CPUs:      tt.cpus,
				Memory:    tt.memory,
				Nodes:     tt.nodes,
			}

			// Test CPU utilization estimation
			cpuUtil := monitor.estimateCPUUtilization(testJob)
			assert.InDelta(t, tt.expectedCPUUtil, cpuUtil, 0.2, "CPU utilization should be within range")
			assert.GreaterOrEqual(t, cpuUtil, 0.1, "CPU utilization should be at least 10%")
			assert.LessOrEqual(t, cpuUtil, 1.0, "CPU utilization should not exceed 100%")

			// Test GPU count estimation
			gpuCount := monitor.estimateGPUCount(testJob)
			assert.Equal(t, tt.expectedGPUCount, gpuCount, "GPU count should match expectation")

			// Test GPU utilization estimation (if GPUs are present)
			if gpuCount > 0 {
				gpuUtil := monitor.estimateGPUUtilization(testJob)
				assert.GreaterOrEqual(t, gpuUtil, 0.0)
				assert.LessOrEqual(t, gpuUtil, 1.0)
				assert.LessOrEqual(t, gpuUtil, cpuUtil, "GPU utilization should not exceed CPU utilization")
			}

			// Test I/O intensity estimation
			ioIntensity := monitor.estimateIOIntensity(testJob)
			assert.InDelta(t, tt.expectedIOIntensity, ioIntensity, 0.3, "I/O intensity should be within range")
			assert.GreaterOrEqual(t, ioIntensity, 1.0, "I/O intensity should be at least base level")
			assert.LessOrEqual(t, ioIntensity, 2.0, "I/O intensity should not exceed maximum")

			// Test network intensity estimation
			netIntensity := monitor.estimateNetworkIntensity(testJob)
			assert.InDelta(t, tt.expectedNetIntensity, netIntensity, 0.3, "Network intensity should be within range")
			assert.GreaterOrEqual(t, netIntensity, 1.0, "Network intensity should be at least base level")
			assert.LessOrEqual(t, netIntensity, 3.0, "Network intensity should not exceed maximum")
		})
	}
}

func TestEnergyMonitor_EfficiencyTracking(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, &EnergyConfig{
		EnableEfficiencyTracking: true,
	})
	require.NoError(t, err)

	jobID := "efficiency-test"
	tracker := monitor.efficiencyTracker

	// Test initial efficiency tracking
	tracker.trackEnergyEfficiency(jobID, 0.7)
	assert.Contains(t, tracker.efficiencyData, jobID)

	trend := tracker.efficiencyData[jobID]
	assert.Equal(t, jobID, trend.JobID)
	assert.Len(t, trend.Timestamps, 1)
	assert.Len(t, trend.EfficiencyScores, 1)
	assert.Equal(t, 0.7, trend.EfficiencyScores[0])
	assert.Equal(t, "stable", trend.Trend)

	// Test multiple efficiency recordings
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	tracker.trackEnergyEfficiency(jobID, 0.75)
	tracker.trackEnergyEfficiency(jobID, 0.8)
	tracker.trackEnergyEfficiency(jobID, 0.85)

	trend = tracker.efficiencyData[jobID]
	assert.Len(t, trend.EfficiencyScores, 4)
	assert.Equal(t, "improving", trend.Trend)
	assert.Greater(t, trend.TrendStrength, 0.0)

	// Test declining trend
	tracker.trackEnergyEfficiency(jobID, 0.8)
	tracker.trackEnergyEfficiency(jobID, 0.75)
	tracker.trackEnergyEfficiency(jobID, 0.7)

	trend = tracker.efficiencyData[jobID]
	trendDirection, trendStrength := tracker.calculateTrend([]float64{0.85, 0.8, 0.75, 0.7})
	assert.Equal(t, "declining", trendDirection)
	assert.Greater(t, trendStrength, 0.0)
}

func TestEnergyMonitor_UpdateEnergyMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	testJob := &slurm.Job{
		JobID:      "metrics-123",
		Name:       "metrics-test-job",
		UserName:   "metricsuser",
		Account:    "metricsaccount",
		Partition:  "compute",
	}

	energyData := &JobEnergyData{
		JobID:              "metrics-123",
		TotalEnergyWh:      1500.0,
		CPUEnergyWh:        800.0,
		MemoryEnergyWh:     300.0,
		GPUEnergyWh:        200.0,
		StorageEnergyWh:    100.0,
		NetworkEnergyWh:    50.0,
		CoolingEnergyWh:    50.0,
		AveragePowerW:      750.0,
		PowerEfficiency:    0.75,
		EnergyWasteWh:      375.0,
		EnergyCostUSD:      0.18,
		CostEfficiency:     0.8,
	}

	carbonData := &CarbonFootprintData{
		JobID:                    "metrics-123",
		TotalCO2grams:           600.0,
		CarbonEfficiency:        0.7,
		CarbonWasteGrams:        180.0,
		EnvironmentalImpactScore: 0.6,
		CarbonIntensity:         400.0,
		GreenEnergyPercentage:   0.4,
	}

	// Update metrics
	monitor.updateJobEnergyMetrics(testJob, energyData, carbonData)

	// Verify energy metrics
	energyMetric := testutil.ToFloat64(monitor.metrics.JobEnergyConsumption.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 1500.0, energyMetric)

	powerMetric := testutil.ToFloat64(monitor.metrics.JobPowerConsumption.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 750.0, powerMetric)

	efficiencyMetric := testutil.ToFloat64(monitor.metrics.JobEnergyEfficiency.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.75, efficiencyMetric)

	wasteMetric := testutil.ToFloat64(monitor.metrics.JobEnergyWaste.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 375.0, wasteMetric)

	// Verify per-resource energy metrics
	cpuEnergyMetric := testutil.ToFloat64(monitor.metrics.CPUEnergyConsumption.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 800.0, cpuEnergyMetric)

	memoryEnergyMetric := testutil.ToFloat64(monitor.metrics.MemoryEnergyConsumption.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 300.0, memoryEnergyMetric)

	gpuEnergyMetric := testutil.ToFloat64(monitor.metrics.GPUEnergyConsumption.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 200.0, gpuEnergyMetric)

	// Verify carbon metrics
	carbonMetric := testutil.ToFloat64(monitor.metrics.JobCarbonEmissions.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 600.0, carbonMetric)

	carbonEfficiencyMetric := testutil.ToFloat64(monitor.metrics.JobCarbonEfficiency.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.7, carbonEfficiencyMetric)

	carbonWasteMetric := testutil.ToFloat64(monitor.metrics.JobCarbonWaste.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 180.0, carbonWasteMetric)

	// Verify cost metrics
	costMetric := testutil.ToFloat64(monitor.metrics.JobEnergyCost.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.18, costMetric)

	costEfficiencyMetric := testutil.ToFloat64(monitor.metrics.JobCostEfficiency.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.8, costEfficiencyMetric)

	// Verify ROI calculation
	expectedROI := 0.75 / 0.18 * 100 // efficiency / cost * 100
	roiMetric := testutil.ToFloat64(monitor.metrics.JobEnergyROI.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.InDelta(t, expectedROI, roiMetric, 0.1)

	// Verify savings potential
	savingsMetric := testutil.ToFloat64(monitor.metrics.EnergySavingsPotential.WithLabelValues(
		"metrics-123", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 375.0, savingsMetric)
}

func TestEnergyMonitor_ClusterMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	// Add some test energy data
	monitor.energyData["job1"] = &JobEnergyData{
		JobID:           "job1",
		TotalEnergyWh:   1000.0,
		AveragePowerW:   500.0,
		PowerEfficiency: 0.8,
	}
	monitor.energyData["job2"] = &JobEnergyData{
		JobID:           "job2",
		TotalEnergyWh:   1500.0,
		AveragePowerW:   750.0,
		PowerEfficiency: 0.7,
	}

	monitor.carbonData["job1"] = &CarbonFootprintData{
		JobID:         "job1",
		TotalCO2grams: 400.0,
	}
	monitor.carbonData["job2"] = &CarbonFootprintData{
		JobID:         "job2",
		TotalCO2grams: 600.0,
	}

	// Test cluster metrics update
	totalEnergy := 2500.0  // 1000 + 1500
	totalPower := 1250.0   // 500 + 750
	totalCarbon := 1000.0  // 400 + 600

	monitor.updateClusterEnergyMetrics(totalEnergy, totalPower, totalCarbon)

	// Verify cluster data
	clusterData := monitor.GetClusterEnergyData()
	assert.Equal(t, totalEnergy, clusterData.TotalClusterEnergyWh)
	assert.Equal(t, totalPower, clusterData.TotalClusterPowerW)
	assert.Equal(t, totalCarbon, clusterData.TotalClusterCO2grams)
	assert.Equal(t, 625.0, clusterData.AverageNodePowerW) // 1250/2 jobs
	assert.Equal(t, 0.75, clusterData.ClusterPowerEfficiency) // (0.8+0.7)/2

	// Verify cluster metrics
	clusterEnergyMetric := testutil.ToFloat64(monitor.metrics.ClusterTotalEnergy.WithLabelValues("default"))
	assert.Equal(t, totalEnergy, clusterEnergyMetric)

	clusterPowerMetric := testutil.ToFloat64(monitor.metrics.ClusterTotalPower.WithLabelValues("default"))
	assert.Equal(t, totalPower, clusterPowerMetric)

	clusterEfficiencyMetric := testutil.ToFloat64(monitor.metrics.ClusterEnergyEfficiency.WithLabelValues("default"))
	assert.Equal(t, 0.75, clusterEfficiencyMetric)

	clusterCarbonMetric := testutil.ToFloat64(monitor.metrics.ClusterCarbonEmissions.WithLabelValues("default"))
	assert.Equal(t, totalCarbon, clusterCarbonMetric)
}

func TestEnergyMonitor_DataManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, &EnergyConfig{
		EnergyDataRetention: 2 * time.Hour,
		CarbonDataRetention: 4 * time.Hour,
	})
	require.NoError(t, err)

	now := time.Now()

	// Add current energy data
	monitor.energyData["current-job"] = &JobEnergyData{
		JobID:     "current-job",
		Timestamp: now,
		TotalEnergyWh: 1000.0,
	}

	// Add old energy data
	monitor.energyData["old-energy-job"] = &JobEnergyData{
		JobID:     "old-energy-job",
		Timestamp: now.Add(-3 * time.Hour), // Older than retention
		TotalEnergyWh: 500.0,
	}

	// Add current carbon data
	monitor.carbonData["current-job"] = &CarbonFootprintData{
		JobID:         "current-job",
		Timestamp:     now,
		TotalCO2grams: 400.0,
	}

	// Add old carbon data
	monitor.carbonData["old-carbon-job"] = &CarbonFootprintData{
		JobID:         "old-carbon-job",
		Timestamp:     now.Add(-5 * time.Hour), // Older than retention
		TotalCO2grams: 200.0,
	}

	// Test data retrieval before cleanup
	energyData, exists := monitor.GetEnergyData("current-job")
	assert.True(t, exists)
	assert.Equal(t, "current-job", energyData.JobID)

	energyData, exists = monitor.GetEnergyData("old-energy-job")
	assert.True(t, exists)
	assert.Equal(t, "old-energy-job", energyData.JobID)

	carbonData, exists := monitor.GetCarbonData("old-carbon-job")
	assert.True(t, exists)
	assert.Equal(t, "old-carbon-job", carbonData.JobID)

	// Clean old data
	monitor.cleanOldEnergyData()

	// Verify old energy data was removed
	_, exists = monitor.GetEnergyData("old-energy-job")
	assert.False(t, exists)

	// Verify old carbon data was removed
	_, exists = monitor.GetCarbonData("old-carbon-job")
	assert.False(t, exists)

	// Verify current data remains
	energyData, exists = monitor.GetEnergyData("current-job")
	assert.True(t, exists)
	assert.Equal(t, "current-job", energyData.JobID)

	carbonData, exists = monitor.GetCarbonData("current-job")
	assert.True(t, exists)
	assert.Equal(t, "current-job", carbonData.JobID)

	// Test energy stats
	stats := monitor.GetEnergyStats()
	assert.Equal(t, 1, stats["monitored_jobs"])
	assert.Equal(t, 1, stats["carbon_tracked"])
	assert.NotNil(t, stats["cluster_energy"])
	assert.NotNil(t, stats["config"])
}

func TestEnergyMonitor_CarbonIntensityFactors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tests := []struct {
		name           string
		region         string
		hour           int
		baseIntensity  float64
		expectedFactor float64
	}{
		{
			name:           "US peak solar hours",
			region:         "US",
			hour:           14, // 2 PM
			baseIntensity:  400.0,
			expectedFactor: 0.8, // Lower during solar hours
		},
		{
			name:           "EU peak demand hours",
			region:         "EU",
			hour:           20, // 8 PM
			baseIntensity:  400.0,
			expectedFactor: 0.8 * 1.2, // EU factor * peak demand factor
		},
		{
			name:           "ASIA off-peak hours",
			region:         "ASIA",
			hour:           3, // 3 AM
			baseIntensity:  400.0,
			expectedFactor: 1.2 * 1.0, // ASIA factor * off-peak factor
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, err := NewEnergyMonitor(mockClient, logger, &EnergyConfig{
				CarbonIntensityGCO2kWh: tt.baseIntensity,
				CarbonTrackingRegion:   tt.region,
			})
			require.NoError(t, err)

			timestamp := time.Date(2023, 6, 15, tt.hour, 0, 0, 0, time.UTC)
			actualIntensity := monitor.carbonCalculator.getCurrentCarbonIntensity(timestamp)
			expectedIntensity := tt.baseIntensity * tt.expectedFactor

			assert.InDelta(t, expectedIntensity, actualIntensity, expectedIntensity*0.05)
		})
	}
}

func TestEnergyMonitor_ConfigurationHelpers(t *testing.T) {
	// Test regional carbon factors
	factors := getRegionalCarbonFactors()
	assert.Contains(t, factors, "US")
	assert.Contains(t, factors, "EU")
	assert.Contains(t, factors, "ASIA")
	assert.Contains(t, factors, "AU")
	assert.Equal(t, 1.0, factors["US"])
	assert.Equal(t, 0.8, factors["EU"])

	// Test time-of-day factors
	timeFactors := getTimeOfDayFactors()
	assert.Len(t, timeFactors, 24)
	assert.Equal(t, 0.8, timeFactors[12]) // Peak solar hours
	assert.Equal(t, 1.2, timeFactors[20]) // Peak demand hours
	assert.Equal(t, 1.0, timeFactors[3])  // Off-peak hours

	// Test energy benchmarks
	benchmarks := getEnergyBenchmarks()
	assert.Contains(t, benchmarks, "cpu_efficiency")
	assert.Contains(t, benchmarks, "memory_efficiency")
	assert.Contains(t, benchmarks, "gpu_efficiency")
	assert.Contains(t, benchmarks, "overall_efficiency")
	assert.Equal(t, 0.75, benchmarks["cpu_efficiency"])

	// Test optimization rules
	rules := getEnergyOptimizationRules()
	assert.Greater(t, len(rules), 0)
	for _, rule := range rules {
		assert.NotEmpty(t, rule.RuleID)
		assert.NotEmpty(t, rule.Condition)
		assert.NotEmpty(t, rule.Recommendation)
		assert.Greater(t, rule.ExpectedSavings, 0.0)
		assert.Contains(t, []string{"low", "medium", "high"}, rule.Priority)
	}
}

func TestEnergyMonitor_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	// Test when job manager returns error
	mockJobManager.On("List", mock.Anything, mock.Anything).Return((*slurm.JobList)(nil), assert.AnError)

	ctx := context.Background()
	err = monitor.collectEnergyMetrics(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list jobs for energy monitoring")

	mockJobManager.AssertExpectations(t)
}

func TestEnergyMonitor_EfficiencyScoreCalculations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewEnergyMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name              string
		totalEnergy       float64
		runtimeHours      float64
		cpus              int
		memory            int
		nodes             int
		expectedEfficiency float64
	}{
		{
			name:              "efficient job",
			totalEnergy:       100.0,
			runtimeHours:      1.0,
			cpus:              4,
			memory:            8192,
			nodes:             1,
			expectedEfficiency: 0.8, // Should be high efficiency
		},
		{
			name:              "inefficient job",
			totalEnergy:       1000.0,
			runtimeHours:      1.0,
			cpus:              4,
			memory:            8192,
			nodes:             1,
			expectedEfficiency: 0.2, // Should be low efficiency
		},
		{
			name:              "zero runtime",
			totalEnergy:       100.0,
			runtimeHours:      0.0,
			cpus:              4,
			memory:            8192,
			nodes:             1,
			expectedEfficiency: 0.0, // Should handle zero runtime
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testJob := &slurm.Job{
				CPUs:   tt.cpus,
				Memory: tt.memory,
				Nodes:  tt.nodes,
			}

			efficiency := monitor.calculatePowerEfficiency(testJob, tt.totalEnergy, tt.runtimeHours)

			assert.GreaterOrEqual(t, efficiency, 0.0, "Efficiency should be non-negative")
			assert.LessOrEqual(t, efficiency, 1.0, "Efficiency should not exceed 1.0")

			if tt.runtimeHours > 0 {
				assert.InDelta(t, tt.expectedEfficiency, efficiency, 0.3, "Efficiency should be within expected range")
			} else {
				assert.Equal(t, 0.0, efficiency, "Efficiency should be 0 for zero runtime")
			}
		})
	}
}