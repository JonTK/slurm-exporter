// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
)

func newTestCalculator() *EfficiencyCalculator {
	return NewEfficiencyCalculator(slog.New(slog.NewTextHandler(os.Stdout, nil)), nil)
}

func newTestCalculatorWithConfig(config *EfficiencyConfig) *EfficiencyCalculator {
	return NewEfficiencyCalculator(slog.New(slog.NewTextHandler(os.Stdout, nil)), config)
}

func TestNewEfficiencyCalculator(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("WithConfig", func(t *testing.T) {
		config := &EfficiencyConfig{
			CPUIdleThreshold:     0.1,
			CPUOptimalThreshold:  0.8,
			MemoryWasteThreshold: 0.5,
			CPUWeight:            0.5,
			MemoryWeight:         0.5,
			MinEfficiencyScore:   0.0,
			MaxEfficiencyScore:   1.0,
			OptimalUtilization:   0.8,
		}
		calc := NewEfficiencyCalculator(logger, config)

		if calc.config.CPUIdleThreshold != 0.1 {
			t.Errorf("Expected CPUIdleThreshold 0.1, got %f", calc.config.CPUIdleThreshold)
		}
		if calc.config.CPUWeight != 0.5 {
			t.Errorf("Expected CPUWeight 0.5, got %f", calc.config.CPUWeight)
		}
	})

	t.Run("DefaultConfig", func(t *testing.T) {
		calc := newTestCalculator()

		if calc.config.MinEfficiencyScore != 0.0 {
			t.Errorf("Expected MinEfficiencyScore 0.0, got %f", calc.config.MinEfficiencyScore)
		}
		if calc.config.MaxEfficiencyScore != 1.0 {
			t.Errorf("Expected MaxEfficiencyScore 1.0, got %f", calc.config.MaxEfficiencyScore)
		}
		if calc.config.OptimalUtilization != 0.8 {
			t.Errorf("Expected OptimalUtilization 0.8, got %f", calc.config.OptimalUtilization)
		}
	})
}

func TestCalculateEfficiency(t *testing.T) {
	t.Run("NilData", func(t *testing.T) {
		calc := newTestCalculator()
		_, err := calc.CalculateEfficiency(nil)

		if err == nil {
			t.Error("Expected error for nil data")
		}
	})

	t.Run("ValidData", func(t *testing.T) {
		calc := newTestCalculator()
		data := &ResourceUtilizationData{
			CPURequested:    10.0,
			CPUAllocated:    10.0,
			CPUUsed:         8.0,
			CPUTimeTotal:    800.0,
			WallTime:        100.0,
			MemoryRequested: 10 * 1024 * 1024 * 1024,
			MemoryAllocated: 10 * 1024 * 1024 * 1024,
			MemoryUsed:      8 * 1024 * 1024 * 1024,
		}

		metrics, err := calc.CalculateEfficiency(data)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if metrics == nil {
			t.Error("Expected metrics to be returned")
		}
		if metrics.OverallEfficiency <= 0.0 || metrics.OverallEfficiency > 1.0 {
			t.Errorf("Overall efficiency out of range: %f", metrics.OverallEfficiency)
		}
		if metrics.CPUEfficiency <= 0.0 || metrics.CPUEfficiency > 1.0 {
			t.Errorf("CPU efficiency out of range: %f", metrics.CPUEfficiency)
		}
		if metrics.MemoryEfficiency <= 0.0 || metrics.MemoryEfficiency > 1.0 {
			t.Errorf("Memory efficiency out of range: %f", metrics.MemoryEfficiency)
		}
		if metrics.EfficiencyGrade == "" {
			t.Error("Expected efficiency grade to be set")
		}
		if metrics.EfficiencyCategory == "" {
			t.Error("Expected efficiency category to be set")
		}
	})

	t.Run("NoIOActivity", func(t *testing.T) {
		calc := newTestCalculator()
		data := &ResourceUtilizationData{
			CPUAllocated:    10.0,
			CPUUsed:         8.0,
			CPUTimeTotal:    800.0,
			WallTime:        100.0,
			MemoryAllocated: 10 * 1024 * 1024 * 1024,
			MemoryUsed:      8 * 1024 * 1024 * 1024,
			IOReadBytes:     0,
			IOWriteBytes:    0,
			IOReadOps:       0,
			IOWriteOps:      0,
		}

		metrics, err := calc.CalculateEfficiency(data)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if metrics.IOEfficiency != 1.0 {
			t.Errorf("Expected IO efficiency 1.0 for no I/O, got %f", metrics.IOEfficiency)
		}
	})
}

func TestCalculateCPUEfficiency(t *testing.T) {
	calc := newTestCalculator()

	t.Run("ZeroAllocated", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated: 0.0,
			WallTime:     100.0,
		}

		efficiency := calc.calculateCPUEfficiency(data)

		if efficiency != 0.0 {
			t.Errorf("Expected 0.0 for zero allocated, got %f", efficiency)
		}
	})

	t.Run("ZeroWallTime", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated: 10.0,
			WallTime:     0.0,
		}

		efficiency := calc.calculateCPUEfficiency(data)

		if efficiency != 0.0 {
			t.Errorf("Expected 0.0 for zero wall time, got %f", efficiency)
		}
	})

	t.Run("OptimalUsage", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated: 10.0,
			CPUUsed:      8.0,
			CPUTimeTotal: 800.0,
			WallTime:     100.0,
		}

		efficiency := calc.calculateCPUEfficiency(data)

		if efficiency <= 0.0 || efficiency > 1.0 {
			t.Errorf("Expected efficiency between 0 and 1, got %f", efficiency)
		}
		if efficiency < 0.6 {
			t.Errorf("Expected high efficiency for optimal usage, got %f", efficiency)
		}
	})

	t.Run("LowUsage", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated: 10.0,
			CPUUsed:      1.0,
			CPUTimeTotal: 100.0,
			WallTime:     1000.0,
		}

		efficiency := calc.calculateCPUEfficiency(data)

		if efficiency >= 0.5 {
			t.Errorf("Expected low efficiency for low usage, got %f", efficiency)
		}
	})
}

func TestCalculateMemoryEfficiency(t *testing.T) {
	calc := newTestCalculator()

	t.Run("ZeroAllocated", func(t *testing.T) {
		data := &ResourceUtilizationData{
			MemoryAllocated: 0,
		}

		efficiency := calc.calculateMemoryEfficiency(data)

		if efficiency != 0.0 {
			t.Errorf("Expected 0.0 for zero allocated, got %f", efficiency)
		}
	})

	t.Run("OptimalUsage", func(t *testing.T) {
		data := &ResourceUtilizationData{
			MemoryAllocated: 10 * 1024 * 1024 * 1024,
			MemoryUsed:      9 * 1024 * 1024 * 1024,
			MemoryPeak:      9 * 1024 * 1024 * 1024,
		}

		efficiency := calc.calculateMemoryEfficiency(data)

		if efficiency <= 0.0 || efficiency > 1.0 {
			t.Errorf("Expected efficiency between 0 and 1, got %f", efficiency)
		}
	})

	t.Run("HighWaste", func(t *testing.T) {
		data := &ResourceUtilizationData{
			MemoryAllocated: 10 * 1024 * 1024 * 1024,
			MemoryUsed:      2 * 1024 * 1024 * 1024,
			MemoryPeak:      2 * 1024 * 1024 * 1024,
		}

		efficiency := calc.calculateMemoryEfficiency(data)

		if efficiency >= 0.5 {
			t.Errorf("Expected low efficiency for high waste, got %f", efficiency)
		}
	})
}

func TestCalculateIOEfficiency(t *testing.T) {
	calc := newTestCalculator()

	t.Run("NoIOActivity", func(t *testing.T) {
		data := &ResourceUtilizationData{
			WallTime:     100.0,
			IOReadBytes:  0,
			IOWriteBytes: 0,
			IOReadOps:    0,
			IOWriteOps:   0,
		}

		efficiency := calc.calculateIOEfficiency(data)

		if efficiency != 1.0 {
			t.Errorf("Expected 1.0 for no I/O, got %f", efficiency)
		}
	})

	t.Run("ZeroWallTime", func(t *testing.T) {
		data := &ResourceUtilizationData{
			WallTime:     0.0,
			IOReadBytes:  1000,
			IOWriteBytes: 1000,
		}

		efficiency := calc.calculateIOEfficiency(data)

		if efficiency != 1.0 {
			t.Errorf("Expected 1.0 for zero wall time, got %f", efficiency)
		}
	})

	t.Run("OptimalIO", func(t *testing.T) {
		data := &ResourceUtilizationData{
			WallTime:     100.0,
			IOReadBytes:  10 * 1024 * 1024 * 1024, // 10GB
			IOWriteBytes: 5 * 1024 * 1024 * 1024,  // 5GB
			IOReadOps:    50000,
			IOWriteOps:   25000,
			IOWaitTime:   5.0,
		}

		efficiency := calc.calculateIOEfficiency(data)

		if efficiency <= 0.0 || efficiency > 1.0 {
			t.Errorf("Expected efficiency between 0 and 1, got %f", efficiency)
		}
	})
}

func TestCalculateNetworkEfficiency(t *testing.T) {
	calc := newTestCalculator()

	t.Run("NoNetworkActivity", func(t *testing.T) {
		data := &ResourceUtilizationData{
			WallTime:         100.0,
			NetworkRxBytes:   0,
			NetworkTxBytes:   0,
			NetworkRxPackets: 0,
			NetworkTxPackets: 0,
		}

		efficiency := calc.calculateNetworkEfficiency(data)

		if efficiency != 1.0 {
			t.Errorf("Expected 1.0 for no network, got %f", efficiency)
		}
	})

	t.Run("OptimalNetwork", func(t *testing.T) {
		data := &ResourceUtilizationData{
			WallTime:         100.0,
			NetworkRxBytes:   5 * 1024 * 1024 * 1024, // 5GB
			NetworkTxBytes:   2 * 1024 * 1024 * 1024, // 2GB
			NetworkRxPackets: 500000,
			NetworkTxPackets: 200000,
		}

		efficiency := calc.calculateNetworkEfficiency(data)

		if efficiency <= 0.0 || efficiency > 1.0 {
			t.Errorf("Expected efficiency between 0 and 1, got %f", efficiency)
		}
	})
}

func TestCalculateResourceEfficiency(t *testing.T) {
	calc := newTestCalculator()

	testCases := []struct {
		name   string
		cpuEff float64
		memEff float64
		min    float64
		max    float64
	}{
		{"BothHigh", 0.9, 0.85, 0.85, 0.9},
		{"BothLow", 0.3, 0.4, 0.3, 0.4},
		{"Mixed", 0.8, 0.5, 0.5, 0.8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			efficiency := calc.calculateResourceEfficiency(tc.cpuEff, tc.memEff)

			if efficiency < tc.min || efficiency > tc.max {
				t.Errorf("Expected efficiency between %.2f and %.2f, got %.2f",
					tc.min, tc.max, efficiency)
			}
		})
	}
}

func TestCalculateThroughputEfficiency(t *testing.T) {
	calc := newTestCalculator()

	testCases := []struct {
		name   string
		ioEff  float64
		netEff float64
		min    float64
		max    float64
	}{
		{"BothHigh", 0.9, 0.85, 0.85, 0.9},
		{"BothLow", 0.3, 0.4, 0.3, 0.4},
		{"Mixed", 0.8, 0.5, 0.5, 0.8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			efficiency := calc.calculateThroughputEfficiency(tc.ioEff, tc.netEff)

			if efficiency < tc.min || efficiency > tc.max {
				t.Errorf("Expected efficiency between %.2f and %.2f, got %.2f",
					tc.min, tc.max, efficiency)
			}
		})
	}
}

func TestCalculateWasteRatio(t *testing.T) {
	calc := newTestCalculator()

	t.Run("NoWaste", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated:    10.0,
			CPUUsed:         10.0,
			MemoryAllocated: 10 * 1024 * 1024 * 1024,
			MemoryUsed:      10 * 1024 * 1024 * 1024,
		}

		wasteRatio := calc.calculateWasteRatio(data)

		if wasteRatio > 0.05 {
			t.Errorf("Expected minimal waste ratio, got %f", wasteRatio)
		}
	})

	t.Run("HighWaste", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated:    10.0,
			CPUUsed:         2.0,
			MemoryAllocated: 10 * 1024 * 1024 * 1024,
			MemoryUsed:      2 * 1024 * 1024 * 1024,
		}

		wasteRatio := calc.calculateWasteRatio(data)

		if wasteRatio < 0.5 {
			t.Errorf("Expected high waste ratio, got %f", wasteRatio)
		}
	})

	t.Run("ZeroAllocated", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated:    0.0,
			MemoryAllocated: 0,
		}

		wasteRatio := calc.calculateWasteRatio(data)

		if wasteRatio != 0.0 {
			t.Errorf("Expected 0.0 waste ratio for zero allocation, got %f", wasteRatio)
		}
	})
}

func TestDetermineEfficiencyGrade(t *testing.T) {
	calc := newTestCalculator()

	testCases := []struct {
		score    float64
		expected string
	}{
		{0.95, "A"},
		{0.90, "A"},
		{0.85, "B"},
		{0.80, "B"},
		{0.75, "C"},
		{0.70, "C"},
		{0.65, "D"},
		{0.60, "D"},
		{0.55, "F"},
		{0.50, "F"},
		{0.0, "F"},
	}

	for _, tc := range testCases {
		t.Run("Score_"+formatFloat(tc.score), func(t *testing.T) {
			grade := calc.determineEfficiencyGrade(tc.score)

			if grade != tc.expected {
				t.Errorf("Score %.2f: expected grade '%s', got '%s'",
					tc.score, tc.expected, grade)
			}
		})
	}
}

func TestDetermineEfficiencyCategory(t *testing.T) {
	calc := newTestCalculator()

	testCases := []struct {
		score    float64
		expected string
	}{
		{0.95, "Excellent"},
		{0.90, "Excellent"},
		{0.85, "Good"},
		{0.80, "Good"},
		{0.75, "Fair"},
		{0.65, "Fair"},
		{0.55, "Poor"},
		{0.45, "Poor"},
		{0.35, "Critical"},
		{0.0, "Critical"},
	}

	for _, tc := range testCases {
		t.Run("Score_"+formatFloat(tc.score), func(t *testing.T) {
			category := calc.determineEfficiencyCategory(tc.score)

			if category != tc.expected {
				t.Errorf("Score %.2f: expected category '%s', got '%s'",
					tc.score, tc.expected, category)
			}
		})
	}
}

func TestGenerateRecommendations(t *testing.T) {
	calc := newTestCalculator()

	t.Run("GoodEfficiency", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated:    10.0,
			CPUUsed:         8.0,
			MemoryAllocated: 10 * 1024 * 1024 * 1024,
			MemoryUsed:      8 * 1024 * 1024 * 1024,
		}
		metrics := &EfficiencyMetrics{
			CPUEfficiency:     0.9,
			MemoryEfficiency:  0.85,
			OverallEfficiency: 0.875,
			WasteRatio:        0.1,
		}

		recs := calc.generateRecommendations(data, metrics)

		if len(recs) == 0 {
			t.Error("Expected at least one recommendation")
		}
	})

	t.Run("PoorCPUEfficiency", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated: 10.0,
			CPUUsed:      1.0,
		}
		metrics := &EfficiencyMetrics{
			CPUEfficiency: 0.2,
		}

		recs := calc.generateRecommendations(data, metrics)

		if len(recs) == 0 {
			t.Error("Expected CPU recommendations")
		}
	})

	t.Run("PoorMemoryEfficiency", func(t *testing.T) {
		data := &ResourceUtilizationData{
			MemoryAllocated: 10 * 1024 * 1024 * 1024,
			MemoryUsed:      1 * 1024 * 1024 * 1024,
		}
		metrics := &EfficiencyMetrics{
			MemoryEfficiency: 0.2,
		}

		recs := calc.generateRecommendations(data, metrics)

		if len(recs) == 0 {
			t.Error("Expected memory recommendations")
		}
	})

	t.Run("HighIOWait", func(t *testing.T) {
		data := &ResourceUtilizationData{
			WallTime:   100.0,
			IOWaitTime: 50.0,
		}
		metrics := &EfficiencyMetrics{
			IOEfficiency: 0.3,
		}

		recs := calc.generateRecommendations(data, metrics)

		if len(recs) == 0 {
			t.Error("Expected I/O recommendations")
		}
	})
}

func TestClampEfficiency(t *testing.T) {
	calc := NewEfficiencyCalculator(nil, &EfficiencyConfig{
		MinEfficiencyScore: 0.1,
		MaxEfficiencyScore: 0.9,
	})

	testCases := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"BelowMin", 0.05, 0.1},
		{"AtMin", 0.1, 0.1},
		{"InRange", 0.5, 0.5},
		{"AtMax", 0.9, 0.9},
		{"AboveMax", 1.0, 0.9},
		{"Zero", 0.0, 0.1},
		{"One", 1.5, 0.9},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calc.clampEfficiency(tc.input)

			if result != tc.expected {
				t.Errorf("Input %.2f: expected %.2f, got %.2f",
					tc.input, tc.expected, result)
			}
		})
	}
}

func TestCalculateUtilizationScore(t *testing.T) {
	calc := NewEfficiencyCalculator(nil, &EfficiencyConfig{
		CPUIdleThreshold:    0.1,
		CPUOptimalThreshold: 0.8,
		OptimalUtilization:  0.8,
	})

	testCases := []struct {
		name        string
		utilization float64
		minScore    float64
		maxScore    float64
	}{
		{"Idle", 0.05, 0.0, 0.15},
		{"Good", 0.5, 0.5, 1.0},
		{"Optimal", 0.8, 0.9, 1.0},
		{"High", 0.95, 0.5, 0.8},
		{"Over", 1.0, 0.5, 0.8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := calc.calculateUtilizationScore(tc.utilization)

			if score < tc.minScore || score > tc.maxScore {
				t.Errorf("Utilization %.2f: expected score between %.2f and %.2f, got %.2f",
					tc.utilization, tc.minScore, tc.maxScore, score)
			}
		})
	}
}

func TestCalculateOptimalityScore(t *testing.T) {
	calc := NewEfficiencyCalculator(nil, &EfficiencyConfig{
		OptimalUtilization: 0.8,
	})

	t.Run("OptimalCPU", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated: 10.0,
			CPUUsed:      8.0,
		}

		score := calc.calculateOptimalityScore(data)

		if score < 0.9 {
			t.Errorf("Expected high optimality score, got %f", score)
		}
	})

	t.Run("SuboptimalCPU", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated: 10.0,
			CPUUsed:      2.0,
		}

		score := calc.calculateOptimalityScore(data)

		if score > 0.5 {
			t.Errorf("Expected low optimality score, got %f", score)
		}
	})

	t.Run("NoAllocation", func(t *testing.T) {
		data := &ResourceUtilizationData{
			CPUAllocated:    0.0,
			MemoryAllocated: 0,
		}

		score := calc.calculateOptimalityScore(data)

		if score != 0.0 {
			t.Errorf("Expected 0.0 for no allocation, got %f", score)
		}
	})
}

func TestCalculateImprovementPotential(t *testing.T) {
	calc := NewEfficiencyCalculator(nil, &EfficiencyConfig{
		MaxEfficiencyScore: 1.0,
	})

	t.Run("HighPotential", func(t *testing.T) {
		metrics := &EfficiencyMetrics{
			OverallEfficiency: 0.3,
			WasteRatio:        0.7,
			OptimalityScore:   0.5,
		}

		potential := calc.calculateImprovementPotential(metrics)

		if potential < 0.5 {
			t.Errorf("Expected high improvement potential, got %f", potential)
		}
	})

	t.Run("LowPotential", func(t *testing.T) {
		metrics := &EfficiencyMetrics{
			OverallEfficiency: 0.9,
			WasteRatio:        0.05,
			OptimalityScore:   0.95,
		}

		potential := calc.calculateImprovementPotential(metrics)

		if potential > 0.2 {
			t.Errorf("Expected low improvement potential, got %f", potential)
		}
	})
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
