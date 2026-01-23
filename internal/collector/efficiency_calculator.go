// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jontk/slurm-client"
)

// EfficiencyCalculator provides algorithms for calculating various efficiency metrics
type EfficiencyCalculator struct {
	logger *slog.Logger
	config *EfficiencyConfig
}

// EfficiencyConfig holds configuration for efficiency calculations
type EfficiencyConfig struct {
	// CPU efficiency thresholds
	CPUIdleThreshold    float64 `yaml:"cpu_idle_threshold"`    // Below this is considered idle
	CPUOptimalThreshold float64 `yaml:"cpu_optimal_threshold"` // Above this is considered optimal

	// Memory efficiency thresholds
	MemoryWasteThreshold    float64 `yaml:"memory_waste_threshold"`    // Above this allocation is wasteful
	MemoryPressureThreshold float64 `yaml:"memory_pressure_threshold"` // Above this is memory pressure

	// I/O efficiency thresholds
	IOOptimalBandwidth float64 `yaml:"io_optimal_bandwidth"` // Optimal I/O bandwidth in MB/s
	IOLatencyThreshold float64 `yaml:"io_latency_threshold"` // Maximum acceptable I/O latency in ms

	// Network efficiency thresholds
	NetworkOptimalBandwidth float64 `yaml:"network_optimal_bandwidth"` // Optimal network bandwidth in MB/s

	// Overall efficiency calculation weights
	CPUWeight     float64 `yaml:"cpu_weight"`     // Weight for CPU efficiency in overall score
	MemoryWeight  float64 `yaml:"memory_weight"`  // Weight for memory efficiency in overall score
	IOWeight      float64 `yaml:"io_weight"`      // Weight for I/O efficiency in overall score
	NetworkWeight float64 `yaml:"network_weight"` // Weight for network efficiency in overall score

	// Efficiency scoring parameters
	MinEfficiencyScore float64 `yaml:"min_efficiency_score"` // Minimum efficiency score (0.0-1.0)
	MaxEfficiencyScore float64 `yaml:"max_efficiency_score"` // Maximum efficiency score (0.0-1.0)
	OptimalUtilization float64 `yaml:"optimal_utilization"`  // Target utilization percentage
}

// Note: EfficiencyMetrics type is defined in common_types.go

// ResourceUtilizationData represents actual resource usage data
type ResourceUtilizationData struct {
	// CPU metrics
	CPURequested  float64 `json:"cpu_requested"`   // CPUs requested
	CPUAllocated  float64 `json:"cpu_allocated"`   // CPUs allocated
	CPUUsed       float64 `json:"cpu_used"`        // CPUs actually used
	CPUTimeTotal  float64 `json:"cpu_time_total"`  // Total CPU time in seconds
	CPUTimeUser   float64 `json:"cpu_time_user"`   // User CPU time in seconds
	CPUTimeSystem float64 `json:"cpu_time_system"` // System CPU time in seconds
	WallTime      float64 `json:"wall_time"`       // Wall clock time in seconds

	// Memory metrics
	MemoryRequested int64 `json:"memory_requested"` // Memory requested in bytes
	MemoryAllocated int64 `json:"memory_allocated"` // Memory allocated in bytes
	MemoryUsed      int64 `json:"memory_used"`      // Memory used in bytes
	MemoryPeak      int64 `json:"memory_peak"`      // Peak memory usage in bytes

	// I/O metrics
	IOReadBytes  int64   `json:"io_read_bytes"`  // Bytes read
	IOWriteBytes int64   `json:"io_write_bytes"` // Bytes written
	IOReadOps    int64   `json:"io_read_ops"`    // Read operations
	IOWriteOps   int64   `json:"io_write_ops"`   // Write operations
	IOWaitTime   float64 `json:"io_wait_time"`   // I/O wait time in seconds

	// Network metrics
	NetworkRxBytes   int64 `json:"network_rx_bytes"`   // Bytes received
	NetworkTxBytes   int64 `json:"network_tx_bytes"`   // Bytes transmitted
	NetworkRxPackets int64 `json:"network_rx_packets"` // Packets received
	NetworkTxPackets int64 `json:"network_tx_packets"` // Packets transmitted

	// Job timing
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	JobState  string    `json:"job_state"`
}

// NewEfficiencyCalculator creates a new efficiency calculator
func NewEfficiencyCalculator(logger *slog.Logger, config *EfficiencyConfig) *EfficiencyCalculator {
	if config == nil {
		config = &EfficiencyConfig{
			CPUIdleThreshold:        0.1,    // 10% CPU usage considered idle
			CPUOptimalThreshold:     0.8,    // 80% CPU usage considered optimal
			MemoryWasteThreshold:    0.5,    // 50% memory waste is excessive
			MemoryPressureThreshold: 0.9,    // 90% memory usage is pressure
			IOOptimalBandwidth:      100.0,  // 100 MB/s optimal I/O
			IOLatencyThreshold:      10.0,   // 10ms maximum I/O latency
			NetworkOptimalBandwidth: 1000.0, // 1 GB/s optimal network
			CPUWeight:               0.4,    // 40% weight for CPU
			MemoryWeight:            0.3,    // 30% weight for memory
			IOWeight:                0.2,    // 20% weight for I/O
			NetworkWeight:           0.1,    // 10% weight for network
			MinEfficiencyScore:      0.0,
			MaxEfficiencyScore:      1.0,
			OptimalUtilization:      0.8, // 80% optimal utilization
		}
	}

	return &EfficiencyCalculator{
		logger: logger,
		config: config,
	}
}

// CalculateEfficiency calculates comprehensive efficiency metrics
func (e *EfficiencyCalculator) CalculateEfficiency(data *ResourceUtilizationData) (*EfficiencyMetrics, error) {
	if data == nil {
		return nil, fmt.Errorf("resource utilization data is nil")
	}

	metrics := &EfficiencyMetrics{}

	// Calculate individual component efficiencies
	metrics.CPUEfficiency = e.calculateCPUEfficiency(data)
	metrics.MemoryEfficiency = e.calculateMemoryEfficiency(data)
	metrics.IOEfficiency = e.calculateIOEfficiency(data)
	metrics.NetworkEfficiency = e.calculateNetworkEfficiency(data)

	// Calculate composite efficiencies
	metrics.ResourceEfficiency = e.calculateResourceEfficiency(metrics.CPUEfficiency, metrics.MemoryEfficiency)
	metrics.ThroughputEfficiency = e.calculateThroughputEfficiency(metrics.IOEfficiency, metrics.NetworkEfficiency)
	metrics.OverallEfficiency = e.calculateOverallEfficiency(metrics)

	// Calculate additional metrics
	metrics.WasteRatio = e.calculateWasteRatio(data)
	metrics.OptimalityScore = e.calculateOptimalityScore(data)
	metrics.ImprovementPotential = e.calculateImprovementPotential(metrics)

	// Determine efficiency grade and category
	metrics.EfficiencyGrade = e.determineEfficiencyGrade(metrics.OverallEfficiency)
	metrics.EfficiencyCategory = e.determineEfficiencyCategory(metrics.OverallEfficiency)

	// Generate recommendations
	metrics.Recommendations = e.generateRecommendations(data, metrics)

	e.logger.Debug("Calculated efficiency metrics",
		"cpu_efficiency", metrics.CPUEfficiency,
		"memory_efficiency", metrics.MemoryEfficiency,
		"overall_efficiency", metrics.OverallEfficiency,
		"grade", metrics.EfficiencyGrade)

	return metrics, nil
}

// calculateCPUEfficiency calculates CPU efficiency based on utilization patterns
func (e *EfficiencyCalculator) calculateCPUEfficiency(data *ResourceUtilizationData) float64 {
	if data.CPUAllocated <= 0 || data.WallTime <= 0 {
		return 0.0
	}

	// Calculate actual CPU utilization
	actualUtilization := data.CPUUsed / data.CPUAllocated

	// Calculate CPU time efficiency
	maxPossibleCPUTime := data.CPUAllocated * data.WallTime
	actualCPUTime := data.CPUTimeTotal
	timeEfficiency := 0.0
	if maxPossibleCPUTime > 0 {
		timeEfficiency = actualCPUTime / maxPossibleCPUTime
	}

	// Combine utilization and time efficiency
	utilizationScore := e.calculateUtilizationScore(actualUtilization)
	timeScore := e.calculateUtilizationScore(timeEfficiency)

	// Weight the scores (favor actual utilization)
	cpuEfficiency := (0.7 * utilizationScore) + (0.3 * timeScore)

	return e.clampEfficiency(cpuEfficiency)
}

// calculateMemoryEfficiency calculates memory efficiency based on allocation vs usage
func (e *EfficiencyCalculator) calculateMemoryEfficiency(data *ResourceUtilizationData) float64 {
	if data.MemoryAllocated <= 0 {
		return 0.0
	}

	// Calculate memory utilization efficiency
	memoryUtilization := float64(data.MemoryUsed) / float64(data.MemoryAllocated)
	utilizationScore := e.calculateUtilizationScore(memoryUtilization)

	// Factor in peak usage vs allocated
	peakUtilization := float64(data.MemoryPeak) / float64(data.MemoryAllocated)
	peakScore := e.calculateUtilizationScore(peakUtilization)

	// Penalize significant over-allocation
	wasteRatio := 1.0 - memoryUtilization
	wastePenalty := 1.0
	if wasteRatio > e.config.MemoryWasteThreshold {
		wastePenalty = 1.0 - ((wasteRatio - e.config.MemoryWasteThreshold) * 0.5)
	}

	// Combine scores with waste penalty
	memoryEfficiency := ((0.6 * utilizationScore) + (0.4 * peakScore)) * wastePenalty

	return e.clampEfficiency(memoryEfficiency)
}

// calculateIOEfficiency calculates I/O efficiency based on throughput and patterns
func (e *EfficiencyCalculator) calculateIOEfficiency(data *ResourceUtilizationData) float64 {
	if data.WallTime <= 0 {
		return 1.0 // No I/O means perfect I/O efficiency
	}

	// Calculate total I/O throughput
	totalIOBytes := data.IOReadBytes + data.IOWriteBytes
	totalIOOps := data.IOReadOps + data.IOWriteOps

	if totalIOBytes == 0 && totalIOOps == 0 {
		return 1.0 // No I/O activity
	}

	// Calculate I/O bandwidth efficiency
	ioThroughputMBps := float64(totalIOBytes) / (data.WallTime * 1024 * 1024)
	bandwidthEfficiency := math.Min(ioThroughputMBps/e.config.IOOptimalBandwidth, 1.0)

	// Calculate I/O wait efficiency (lower wait time is better)
	ioWaitRatio := data.IOWaitTime / data.WallTime
	waitEfficiency := math.Max(0.0, 1.0-(ioWaitRatio*2.0)) // Penalize high I/O wait

	// Calculate operation efficiency (operations per second)
	opsPerSecond := float64(totalIOOps) / data.WallTime
	opsEfficiency := math.Min(opsPerSecond/1000.0, 1.0) // Assume 1000 ops/sec is optimal

	// Combine efficiency scores
	ioEfficiency := (0.5 * bandwidthEfficiency) + (0.3 * waitEfficiency) + (0.2 * opsEfficiency)

	return e.clampEfficiency(ioEfficiency)
}

// calculateNetworkEfficiency calculates network efficiency based on throughput
func (e *EfficiencyCalculator) calculateNetworkEfficiency(data *ResourceUtilizationData) float64 {
	if data.WallTime <= 0 {
		return 1.0 // No network activity
	}

	// Calculate total network throughput
	totalNetworkBytes := data.NetworkRxBytes + data.NetworkTxBytes
	totalNetworkPackets := data.NetworkRxPackets + data.NetworkTxPackets

	if totalNetworkBytes == 0 && totalNetworkPackets == 0 {
		return 1.0 // No network activity
	}

	// Calculate network bandwidth efficiency
	networkThroughputMBps := float64(totalNetworkBytes) / (data.WallTime * 1024 * 1024)
	bandwidthEfficiency := math.Min(networkThroughputMBps/e.config.NetworkOptimalBandwidth, 1.0)

	// Calculate packet efficiency
	packetsPerSecond := float64(totalNetworkPackets) / data.WallTime
	packetEfficiency := math.Min(packetsPerSecond/10000.0, 1.0) // Assume 10k packets/sec is optimal

	// Calculate average packet size efficiency (larger packets are generally more efficient)
	avgPacketSize := 0.0
	if totalNetworkPackets > 0 {
		avgPacketSize = float64(totalNetworkBytes) / float64(totalNetworkPackets)
	}
	sizeEfficiency := math.Min(avgPacketSize/1500.0, 1.0) // 1500 bytes (MTU) is optimal

	// Combine efficiency scores
	networkEfficiency := (0.6 * bandwidthEfficiency) + (0.2 * packetEfficiency) + (0.2 * sizeEfficiency)

	return e.clampEfficiency(networkEfficiency)
}

// calculateResourceEfficiency calculates composite resource efficiency (CPU + Memory)
func (e *EfficiencyCalculator) calculateResourceEfficiency(cpuEff, memoryEff float64) float64 {
	// Weight CPU more heavily as it's often the primary resource
	resourceEfficiency := (0.6 * cpuEff) + (0.4 * memoryEff)
	return e.clampEfficiency(resourceEfficiency)
}

// calculateThroughputEfficiency calculates composite throughput efficiency (I/O + Network)
func (e *EfficiencyCalculator) calculateThroughputEfficiency(ioEff, networkEff float64) float64 {
	// Weight I/O more heavily as it's often more critical than network
	throughputEfficiency := (0.7 * ioEff) + (0.3 * networkEff)
	return e.clampEfficiency(throughputEfficiency)
}

// calculateOverallEfficiency calculates the overall weighted efficiency score
func (e *EfficiencyCalculator) calculateOverallEfficiency(metrics *EfficiencyMetrics) float64 {
	overallEfficiency := (e.config.CPUWeight * metrics.CPUEfficiency) +
		(e.config.MemoryWeight * metrics.MemoryEfficiency) +
		(e.config.IOWeight * metrics.IOEfficiency) +
		(e.config.NetworkWeight * metrics.NetworkEfficiency)

	return e.clampEfficiency(overallEfficiency)
}

// calculateWasteRatio calculates the ratio of wasted resources
func (e *EfficiencyCalculator) calculateWasteRatio(data *ResourceUtilizationData) float64 {
	cpuWaste := 0.0
	if data.CPUAllocated > 0 {
		cpuWaste = math.Max(0.0, (data.CPUAllocated-data.CPUUsed)/data.CPUAllocated)
	}

	memoryWaste := 0.0
	if data.MemoryAllocated > 0 {
		memoryWaste = math.Max(0.0, float64(data.MemoryAllocated-data.MemoryUsed)/float64(data.MemoryAllocated))
	}

	// Average waste across CPU and memory (weight CPU more heavily)
	wasteRatio := (0.6 * cpuWaste) + (0.4 * memoryWaste)

	return math.Min(wasteRatio, 1.0)
}

// calculateOptimalityScore calculates how close the configuration is to optimal
func (e *EfficiencyCalculator) calculateOptimalityScore(data *ResourceUtilizationData) float64 {
	scores := []float64{}

	// CPU optimality
	if data.CPUAllocated > 0 {
		cpuUtil := data.CPUUsed / data.CPUAllocated
		cpuOptimality := 1.0 - math.Abs(cpuUtil-e.config.OptimalUtilization)
		scores = append(scores, math.Max(0.0, cpuOptimality))
	}

	// Memory optimality
	if data.MemoryAllocated > 0 {
		memoryUtil := float64(data.MemoryUsed) / float64(data.MemoryAllocated)
		memoryOptimality := 1.0 - math.Abs(memoryUtil-e.config.OptimalUtilization)
		scores = append(scores, math.Max(0.0, memoryOptimality))
	}

	if len(scores) == 0 {
		return 0.0
	}

	// Average optimality scores
	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

// calculateImprovementPotential calculates potential for efficiency improvement
func (e *EfficiencyCalculator) calculateImprovementPotential(metrics *EfficiencyMetrics) float64 {
	// Higher potential when current efficiency is low
	currentEfficiency := metrics.OverallEfficiency
	maxPossibleImprovement := e.config.MaxEfficiencyScore - currentEfficiency

	// Factor in waste ratio (more waste = more improvement potential)
	wasteFactor := metrics.WasteRatio

	// Factor in distance from optimality
	optimalityFactor := 1.0 - metrics.OptimalityScore

	// Combine factors
	improvementPotential := maxPossibleImprovement * (0.5 + (0.3 * wasteFactor) + (0.2 * optimalityFactor))

	return e.clampEfficiency(improvementPotential)
}

// calculateUtilizationScore converts raw utilization to efficiency score
func (e *EfficiencyCalculator) calculateUtilizationScore(utilization float64) float64 {
	if utilization < e.config.CPUIdleThreshold {
		// Very low utilization - poor efficiency
		return utilization / e.config.CPUIdleThreshold * 0.3
	} else if utilization <= e.config.OptimalUtilization {
		// Good utilization range
		progress := (utilization - e.config.CPUIdleThreshold) / (e.config.OptimalUtilization - e.config.CPUIdleThreshold)
		return 0.3 + (progress * 0.7) // Scale from 0.3 to 1.0
	} else if utilization <= e.config.CPUOptimalThreshold {
		// Still good, but slightly over optimal
		excess := utilization - e.config.OptimalUtilization
		maxExcess := e.config.CPUOptimalThreshold - e.config.OptimalUtilization
		penalty := (excess / maxExcess) * 0.2
		return 1.0 - penalty
	} else {
		// Over-utilization - declining efficiency
		overuse := utilization - e.config.CPUOptimalThreshold
		return math.Max(0.5, 0.8-(overuse*2.0))
	}
}

// clampEfficiency ensures efficiency score is within valid range
func (e *EfficiencyCalculator) clampEfficiency(efficiency float64) float64 {
	return math.Max(e.config.MinEfficiencyScore, math.Min(e.config.MaxEfficiencyScore, efficiency))
}

// determineEfficiencyGrade assigns a letter grade based on efficiency score
func (e *EfficiencyCalculator) determineEfficiencyGrade(efficiency float64) string {
	if efficiency >= 0.9 {
		return "A"
	} else if efficiency >= 0.8 {
		return "B"
	} else if efficiency >= 0.7 {
		return "C"
	} else if efficiency >= 0.6 {
		return "D"
	} else {
		return "F"
	}
}

// determineEfficiencyCategory assigns a category based on efficiency score
func (e *EfficiencyCalculator) determineEfficiencyCategory(efficiency float64) string {
	if efficiency >= 0.9 {
		return "Excellent"
	} else if efficiency >= 0.8 {
		return "Good"
	} else if efficiency >= 0.6 {
		return "Fair"
	} else if efficiency >= 0.4 {
		return "Poor"
	} else {
		return "Critical"
	}
}

// generateRecommendations provides actionable recommendations for improvement
func (e *EfficiencyCalculator) generateRecommendations(data *ResourceUtilizationData, metrics *EfficiencyMetrics) []string {
	recommendations := []string{}

	// CPU recommendations
	if metrics.CPUEfficiency < 0.6 {
		if data.CPUAllocated > 0 {
			cpuUtil := data.CPUUsed / data.CPUAllocated
			if cpuUtil < e.config.CPUIdleThreshold {
				recommendations = append(recommendations, "Consider reducing CPU allocation - current usage is very low")
			} else if cpuUtil > e.config.CPUOptimalThreshold {
				recommendations = append(recommendations, "Consider increasing CPU allocation or optimizing CPU-intensive operations")
			}
		}
	}

	// Memory recommendations
	if metrics.MemoryEfficiency < 0.6 {
		if data.MemoryAllocated > 0 {
			memoryUtil := float64(data.MemoryUsed) / float64(data.MemoryAllocated)
			if memoryUtil < 0.5 {
				recommendations = append(recommendations, "Consider reducing memory allocation - significant memory waste detected")
			} else if memoryUtil > e.config.MemoryPressureThreshold {
				recommendations = append(recommendations, "Consider increasing memory allocation - memory pressure detected")
			}
		}
	}

	// I/O recommendations
	if metrics.IOEfficiency < 0.6 {
		if data.IOWaitTime > 0 && data.WallTime > 0 {
			ioWaitRatio := data.IOWaitTime / data.WallTime
			if ioWaitRatio > 0.1 {
				recommendations = append(recommendations, "High I/O wait time detected - consider optimizing I/O patterns or using faster storage")
			}
		}
	}

	// Network recommendations
	if metrics.NetworkEfficiency < 0.6 {
		if data.NetworkRxBytes > 0 || data.NetworkTxBytes > 0 {
			recommendations = append(recommendations, "Network efficiency is low - consider optimizing data transfer patterns")
		}
	}

	// Overall recommendations
	if metrics.OverallEfficiency < 0.5 {
		recommendations = append(recommendations, "Overall efficiency is critical - review resource allocation and job configuration")
	}

	if metrics.WasteRatio > 0.4 {
		recommendations = append(recommendations, "High resource waste detected - consider right-sizing resource requests")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Job efficiency is good - continue current configuration")
	}

	return recommendations
}

// CreateResourceUtilizationDataFromJob creates utilization data from basic job information
func CreateResourceUtilizationDataFromJob(job *slurm.Job) *ResourceUtilizationData {
	data := &ResourceUtilizationData{
		CPURequested:    float64(job.CPUs),
		CPUAllocated:    float64(job.CPUs),
		CPUUsed:         float64(job.CPUs) * 0.75,        // Estimate 75% usage
		MemoryRequested: int64(job.Memory) * 1024 * 1024, // Convert MB to bytes
		MemoryAllocated: int64(job.Memory) * 1024 * 1024,
		MemoryUsed:      int64(job.Memory) * 1024 * 1024 * 65 / 100, // Estimate 65% usage
		JobState:        job.State,
	}

	if job.StartTime != nil {
		data.StartTime = *job.StartTime
		if job.EndTime != nil {
			data.EndTime = *job.EndTime
			data.WallTime = job.EndTime.Sub(*job.StartTime).Seconds()
		} else if job.State == JobStateRunning {
			data.EndTime = time.Now()
			data.WallTime = time.Since(*job.StartTime).Seconds()
		}
	}

	// Estimate CPU time based on wall time and usage
	if data.WallTime > 0 {
		data.CPUTimeTotal = data.CPUUsed * data.WallTime
		data.CPUTimeUser = data.CPUTimeTotal * 0.9   // 90% user time
		data.CPUTimeSystem = data.CPUTimeTotal * 0.1 // 10% system time
	}

	return data
}
