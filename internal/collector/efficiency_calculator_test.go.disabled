package collector

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEfficiencyCalculator(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tests := []struct {
		name   string
		config *EfficiencyConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &EfficiencyConfig{
				CPUIdleThreshold:        0.05,
				CPUOptimalThreshold:     0.85,
				MemoryWasteThreshold:    0.4,
				MemoryPressureThreshold: 0.95,
				CPUWeight:               0.5,
				MemoryWeight:            0.3,
				IOWeight:                0.15,
				NetworkWeight:           0.05,
				OptimalUtilization:      0.75,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator := NewEfficiencyCalculator(logger, tt.config)
			require.NotNil(t, calculator)
			assert.NotNil(t, calculator.config)
			assert.NotNil(t, calculator.logger)

			if tt.config == nil {
				assert.Equal(t, 0.1, calculator.config.CPUIdleThreshold)
				assert.Equal(t, 0.8, calculator.config.CPUOptimalThreshold)
				assert.Equal(t, 0.4, calculator.config.CPUWeight)
			} else {
				assert.Equal(t, tt.config.CPUIdleThreshold, calculator.config.CPUIdleThreshold)
				assert.Equal(t, tt.config.CPUOptimalThreshold, calculator.config.CPUOptimalThreshold)
				assert.Equal(t, tt.config.CPUWeight, calculator.config.CPUWeight)
			}
		})
	}
}

func TestEfficiencyCalculator_CalculateEfficiency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	calculator := NewEfficiencyCalculator(logger, nil)

	tests := []struct {
		name     string
		data     *ResourceUtilizationData
		wantErr  bool
		validate func(*testing.T, *EfficiencyMetrics)
	}{
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name: "high efficiency job",
			data: &ResourceUtilizationData{
				CPURequested:    4.0,
				CPUAllocated:    4.0,
				CPUUsed:         3.2, // 80% utilization
				CPUTimeTotal:    3600.0,
				CPUTimeUser:     3200.0,
				CPUTimeSystem:   400.0,
				WallTime:        3600.0, // 1 hour
				MemoryRequested: 8 * 1024 * 1024 * 1024, // 8GB
				MemoryAllocated: 8 * 1024 * 1024 * 1024,
				MemoryUsed:      6 * 1024 * 1024 * 1024, // 75% utilization
				MemoryPeak:      7 * 1024 * 1024 * 1024, // 87.5% peak
				IOReadBytes:     100 * 1024 * 1024,     // 100MB
				IOWriteBytes:    50 * 1024 * 1024,      // 50MB
				IOReadOps:       1000,
				IOWriteOps:      500,
				IOWaitTime:      36.0, // 1% of wall time
				NetworkRxBytes:  1024 * 1024 * 1024, // 1GB
				NetworkTxBytes:  512 * 1024 * 1024,  // 512MB
				JobState:        "COMPLETED",
			},
			wantErr: false,
			validate: func(t *testing.T, metrics *EfficiencyMetrics) {
				assert.Greater(t, metrics.CPUEfficiency, 0.7)
				assert.Greater(t, metrics.MemoryEfficiency, 0.6)
				assert.Greater(t, metrics.OverallEfficiency, 0.6)
				assert.Equal(t, "B", metrics.EfficiencyGrade)
				assert.Contains(t, []string{"Good", "Excellent"}, metrics.EfficiencyCategory)
				assert.NotEmpty(t, metrics.Recommendations)
			},
		},
		{
			name: "low efficiency job",
			data: &ResourceUtilizationData{
				CPURequested:    8.0,
				CPUAllocated:    8.0,
				CPUUsed:         0.8, // 10% utilization - very low
				CPUTimeTotal:    360.0,
				CPUTimeUser:     320.0,
				CPUTimeSystem:   40.0,
				WallTime:        3600.0, // 1 hour
				MemoryRequested: 16 * 1024 * 1024 * 1024, // 16GB
				MemoryAllocated: 16 * 1024 * 1024 * 1024,
				MemoryUsed:      2 * 1024 * 1024 * 1024, // 12.5% utilization - very low
				MemoryPeak:      3 * 1024 * 1024 * 1024, // 18.75% peak
				IOReadBytes:     10 * 1024 * 1024, // 10MB
				IOWriteBytes:    5 * 1024 * 1024,  // 5MB
				IOWaitTime:      360.0,            // 10% of wall time - high
				JobState:        "COMPLETED",
			},
			wantErr: false,
			validate: func(t *testing.T, metrics *EfficiencyMetrics) {
				assert.Less(t, metrics.CPUEfficiency, 0.4)
				assert.Less(t, metrics.MemoryEfficiency, 0.4)
				assert.Less(t, metrics.OverallEfficiency, 0.4)
				assert.Contains(t, []string{"D", "F"}, metrics.EfficiencyGrade)
				assert.Contains(t, []string{"Poor", "Critical"}, metrics.EfficiencyCategory)
				assert.Greater(t, metrics.WasteRatio, 0.5)
				assert.Greater(t, metrics.ImprovementPotential, 0.3)
				assert.Contains(t, metrics.Recommendations[0], "reducing CPU allocation")
			},
		},
		{
			name: "memory pressure job",
			data: &ResourceUtilizationData{
				CPURequested:    4.0,
				CPUAllocated:    4.0,
				CPUUsed:         3.6, // 90% utilization
				CPUTimeTotal:    3600.0,
				WallTime:        3600.0,
				MemoryRequested: 4 * 1024 * 1024 * 1024, // 4GB
				MemoryAllocated: 4 * 1024 * 1024 * 1024,
				MemoryUsed:      3.8 * 1024 * 1024 * 1024, // 95% utilization - high pressure
				MemoryPeak:      3.9 * 1024 * 1024 * 1024, // 97.5% peak
				JobState:        "COMPLETED",
			},
			wantErr: false,
			validate: func(t *testing.T, metrics *EfficiencyMetrics) {
				assert.Greater(t, metrics.CPUEfficiency, 0.8)
				// Memory efficiency might be lower due to pressure
				assert.Contains(t, metrics.Recommendations, "Consider increasing memory allocation - memory pressure detected")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, err := calculator.CalculateEfficiency(tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, metrics)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, metrics)

			// Validate basic metric ranges
			assert.GreaterOrEqual(t, metrics.CPUEfficiency, 0.0)
			assert.LessOrEqual(t, metrics.CPUEfficiency, 1.0)
			assert.GreaterOrEqual(t, metrics.MemoryEfficiency, 0.0)
			assert.LessOrEqual(t, metrics.MemoryEfficiency, 1.0)
			assert.GreaterOrEqual(t, metrics.OverallEfficiency, 0.0)
			assert.LessOrEqual(t, metrics.OverallEfficiency, 1.0)
			assert.NotEmpty(t, metrics.EfficiencyGrade)
			assert.NotEmpty(t, metrics.EfficiencyCategory)

			if tt.validate != nil {
				tt.validate(t, metrics)
			}
		})
	}
}

func TestEfficiencyCalculator_CPUEfficiency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	calculator := NewEfficiencyCalculator(logger, nil)

	tests := []struct {
		name           string
		cpuAllocated   float64
		cpuUsed        float64
		wallTime       float64
		cpuTimeTotal   float64
		expectedRange  [2]float64 // min, max expected efficiency
	}{
		{
			name:          "optimal CPU usage",
			cpuAllocated:  4.0,
			cpuUsed:       3.2, // 80% utilization
			wallTime:      3600.0,
			cpuTimeTotal:  3200.0,
			expectedRange: [2]float64{0.8, 1.0},
		},
		{
			name:          "low CPU usage",
			cpuAllocated:  8.0,
			cpuUsed:       0.8, // 10% utilization
			wallTime:      3600.0,
			cpuTimeTotal:  800.0,
			expectedRange: [2]float64{0.0, 0.3},
		},
		{
			name:          "very high CPU usage",
			cpuAllocated:  2.0,
			cpuUsed:       1.9, // 95% utilization
			wallTime:      3600.0,
			cpuTimeTotal:  1900.0,
			expectedRange: [2]float64{0.7, 1.0},
		},
		{
			name:          "zero allocation",
			cpuAllocated:  0.0,
			cpuUsed:       0.0,
			wallTime:      3600.0,
			cpuTimeTotal:  0.0,
			expectedRange: [2]float64{0.0, 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &ResourceUtilizationData{
				CPUAllocated: tt.cpuAllocated,
				CPUUsed:      tt.cpuUsed,
				WallTime:     tt.wallTime,
				CPUTimeTotal: tt.cpuTimeTotal,
			}

			efficiency := calculator.calculateCPUEfficiency(data)
			assert.GreaterOrEqual(t, efficiency, tt.expectedRange[0], "CPU efficiency below expected minimum")
			assert.LessOrEqual(t, efficiency, tt.expectedRange[1], "CPU efficiency above expected maximum")
		})
	}
}

func TestEfficiencyCalculator_MemoryEfficiency(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	calculator := NewEfficiencyCalculator(logger, nil)

	tests := []struct {
		name            string
		memoryAllocated int64
		memoryUsed      int64
		memoryPeak      int64
		expectedRange   [2]float64
	}{
		{
			name:            "optimal memory usage",
			memoryAllocated: 8 * 1024 * 1024 * 1024, // 8GB
			memoryUsed:      6 * 1024 * 1024 * 1024, // 6GB (75%)
			memoryPeak:      7 * 1024 * 1024 * 1024, // 7GB (87.5%)
			expectedRange:   [2]float64{0.7, 1.0},
		},
		{
			name:            "memory waste",
			memoryAllocated: 16 * 1024 * 1024 * 1024, // 16GB
			memoryUsed:      2 * 1024 * 1024 * 1024,  // 2GB (12.5%)
			memoryPeak:      3 * 1024 * 1024 * 1024,  // 3GB (18.75%)
			expectedRange:   [2]float64{0.0, 0.4},
		},
		{
			name:            "memory pressure",
			memoryAllocated: 4 * 1024 * 1024 * 1024,  // 4GB
			memoryUsed:      3.8 * 1024 * 1024 * 1024, // 3.8GB (95%)
			memoryPeak:      3.9 * 1024 * 1024 * 1024, // 3.9GB (97.5%)
			expectedRange:   [2]float64{0.6, 1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &ResourceUtilizationData{
				MemoryAllocated: tt.memoryAllocated,
				MemoryUsed:      tt.memoryUsed,
				MemoryPeak:      tt.memoryPeak,
			}

			efficiency := calculator.calculateMemoryEfficiency(data)
			assert.GreaterOrEqual(t, efficiency, tt.expectedRange[0], "Memory efficiency below expected minimum")
			assert.LessOrEqual(t, efficiency, tt.expectedRange[1], "Memory efficiency above expected maximum")
		})
	}
}

func TestEfficiencyCalculator_Grades(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	calculator := NewEfficiencyCalculator(logger, nil)

	tests := []struct {
		efficiency       float64
		expectedGrade    string
		expectedCategory string
	}{
		{0.95, "A", "Excellent"},
		{0.85, "B", "Good"},
		{0.75, "C", "Good"},
		{0.65, "D", "Fair"},
		{0.45, "F", "Poor"},
		{0.25, "F", "Critical"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedGrade, func(t *testing.T) {
			grade := calculator.determineEfficiencyGrade(tt.efficiency)
			category := calculator.determineEfficiencyCategory(tt.efficiency)

			assert.Equal(t, tt.expectedGrade, grade)
			assert.Equal(t, tt.expectedCategory, category)
		})
	}
}

func TestCreateResourceUtilizationDataFromJob(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-2 * time.Hour)
	endTime := now.Add(-1 * time.Hour)

	job := &slurm.Job{
		JobID:      "12345",
		Name:       "test-job",
		UserName:   "testuser",
		Account:    "testaccount",
		Partition:  "compute",
		JobState:   "COMPLETED",
		CPUs:       4,
		Memory:     8192, // 8GB in MB
		Nodes:      1,
		StartTime:  &startTime,
		EndTime:    &endTime,
	}

	data := CreateResourceUtilizationDataFromJob(job)

	assert.Equal(t, 4.0, data.CPURequested)
	assert.Equal(t, 4.0, data.CPUAllocated)
	assert.Equal(t, 3.0, data.CPUUsed) // 75% of allocated
	assert.Equal(t, int64(8192*1024*1024), data.MemoryRequested)
	assert.Equal(t, int64(8192*1024*1024), data.MemoryAllocated)
	assert.Equal(t, int64(8192*1024*1024*65/100), data.MemoryUsed) // 65% of allocated
	assert.Equal(t, "COMPLETED", data.JobState)
	assert.Equal(t, startTime, data.StartTime)
	assert.Equal(t, endTime, data.EndTime)
	assert.Equal(t, 3600.0, data.WallTime) // 1 hour
	assert.Greater(t, data.CPUTimeTotal, 0.0)
}

func TestEfficiencyCalculator_Recommendations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	calculator := NewEfficiencyCalculator(logger, nil)

	// Test low CPU utilization scenario
	lowCPUData := &ResourceUtilizationData{
		CPUAllocated: 8.0,
		CPUUsed:      0.4, // 5% utilization
		WallTime:     3600.0,
		MemoryAllocated: 8 * 1024 * 1024 * 1024,
		MemoryUsed:      6 * 1024 * 1024 * 1024,
	}

	metrics, err := calculator.CalculateEfficiency(lowCPUData)
	require.NoError(t, err)

	found := false
	for _, rec := range metrics.Recommendations {
		if testContains(rec, "reducing CPU allocation") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should recommend reducing CPU allocation for low utilization")

	// Test memory pressure scenario
	memoryPressureData := &ResourceUtilizationData{
		CPUAllocated: 4.0,
		CPUUsed:      3.2,
		WallTime:     3600.0,
		MemoryAllocated: 4 * 1024 * 1024 * 1024,
		MemoryUsed:      3.8 * 1024 * 1024 * 1024, // 95% utilization
	}

	metrics, err = calculator.CalculateEfficiency(memoryPressureData)
	require.NoError(t, err)

	found = false
	for _, rec := range metrics.Recommendations {
		if testContains(rec, "increasing memory allocation") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should recommend increasing memory allocation for high pressure")
}

// Helper function to check if a string contains a substring
func testContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			testHasSubstring(s, substr))))
}

func testHasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}