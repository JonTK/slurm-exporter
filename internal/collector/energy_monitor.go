// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
)

// EnergyMonitor provides energy consumption tracking and carbon footprint analysis
type EnergyMonitor struct {
	slurmClient slurm.SlurmClient
	logger      *slog.Logger
	config      *EnergyConfig
	metrics     *EnergyMetrics

	// Energy tracking data
	energyData     map[string]*JobEnergyData
	carbonData     map[string]*CarbonFootprintData
	clusterEnergy  *ClusterEnergyData
	lastCollection time.Time
	mu             sync.RWMutex

	// Carbon footprint calculation
	carbonCalculator *CarbonFootprintCalculator

	// Energy efficiency tracking
	efficiencyTracker *EnergyEfficiencyTracker
}

// EnergyConfig configures the energy monitoring collector
type EnergyConfig struct {
	MonitoringInterval       time.Duration
	MaxJobsPerCollection     int
	EnableCarbonTracking     bool
	EnableEfficiencyTracking bool

	// Energy calculation parameters
	EnergyCalculationMethod   string  // "power_model", "hardware_counters", "estimated"
	DefaultCPUPowerWatts      float64 // Watts per CPU core
	DefaultMemoryPowerWatts   float64 // Watts per GB of memory
	DefaultGPUPowerWatts      float64 // Watts per GPU
	DefaultNodeBasePowerWatts float64 // Base power consumption per node

	// Carbon footprint parameters
	CarbonIntensityGCO2kWh  float64 // grams CO2 per kWh
	CarbonTrackingRegion    string  // Geographic region for carbon intensity
	EnableRealTimeCarbonAPI bool    // Use real-time carbon intensity APIs

	// Efficiency parameters
	EnablePowerCapping     bool
	PowerCapThresholdWatts float64
	EnergyBudgetPerJobWh   float64

	// Data retention
	EnergyDataRetention time.Duration
	CarbonDataRetention time.Duration
}

// JobEnergyData contains energy consumption data for a job
type JobEnergyData struct {
	JobID     string
	Timestamp time.Time

	// Energy consumption metrics
	TotalEnergyWh   float64 // Total energy consumed in watt-hours
	CPUEnergyWh     float64 // CPU energy consumption
	MemoryEnergyWh  float64 // Memory energy consumption
	GPUEnergyWh     float64 // GPU energy consumption
	StorageEnergyWh float64 // Storage I/O energy consumption
	NetworkEnergyWh float64 // Network energy consumption
	CoolingEnergyWh float64 // Estimated cooling energy

	// Power consumption metrics
	AveragePowerW   float64 // Average power consumption in watts
	PeakPowerW      float64 // Peak power consumption
	PowerEfficiency float64 // Power utilization efficiency

	// Energy efficiency metrics
	EnergyPerCPUHour  float64 // Energy per CPU-hour
	EnergyPerTaskUnit float64 // Energy per unit of work completed
	EnergyWasteWh     float64 // Wasted energy due to inefficiency

	// Temporal analysis
	EnergyProfile []EnergyDataPoint // Time series energy data
	PowerProfile  []PowerDataPoint  // Time series power data

	// Cost analysis
	EnergyCostUSD  float64 // Energy cost in USD
	CostPerCPUHour float64 // Cost per CPU-hour
	CostEfficiency float64 // Cost efficiency score
}

// CarbonFootprintData contains carbon footprint analysis for a job
type CarbonFootprintData struct {
	JobID     string
	Timestamp time.Time

	// Carbon emissions
	TotalCO2grams  float64 // Total CO2 emissions in grams
	CO2PerCPUHour  float64 // CO2 emissions per CPU-hour
	CO2PerTaskUnit float64 // CO2 emissions per unit of work

	// Carbon intensity tracking
	CarbonIntensity float64 // gCO2/kWh at time of execution
	RegionalFactor  float64 // Regional carbon intensity factor
	TimeOfDayFactor float64 // Time-of-day carbon intensity variation

	// Carbon efficiency
	CarbonEfficiency float64 // Carbon efficiency score
	CarbonWasteGrams float64 // Wasted carbon emissions

	// Offset and credits
	CarbonOffsetGrams float64 // Available carbon offsets
	NetCarbonGrams    float64 // Net carbon after offsets
	CarbonCreditsUsed float64 // Carbon credits consumed

	// Environmental impact
	EnvironmentalImpactScore float64 // Overall environmental impact score
	SustainabilityGrade      string  // A-F sustainability grade
	GreenEnergyPercentage    float64 // Percentage from renewable sources
}

// ClusterEnergyData contains cluster-wide energy metrics
type ClusterEnergyData struct {
	Timestamp time.Time

	// Cluster energy consumption
	TotalClusterEnergyWh float64
	TotalClusterPowerW   float64
	AverageNodePowerW    float64

	// Efficiency metrics
	ClusterPowerEfficiency float64
	EnergyUtilizationRate  float64

	// Carbon metrics
	TotalClusterCO2grams   float64
	ClusterCarbonIntensity float64

	// Costs
	TotalEnergyCostUSD float64
	CostPerWh          float64
}

// EnergyDataPoint represents a single energy measurement
type EnergyDataPoint struct {
	Timestamp  time.Time
	EnergyWh   float64
	PowerW     float64
	Efficiency float64
}

// PowerDataPoint represents a single power measurement
type PowerDataPoint struct {
	Timestamp   time.Time
	PowerW      float64
	Voltage     float64
	Current     float64
	PowerFactor float64
}

// CarbonFootprintCalculator calculates carbon footprint metrics
type CarbonFootprintCalculator struct {
	config          *EnergyConfig
	logger          *slog.Logger
	carbonIntensity float64
	regionalFactors map[string]float64
	timeFactors     map[int]float64 // Hour of day factors
}

// EnergyEfficiencyTracker tracks energy efficiency metrics
type EnergyEfficiencyTracker struct {
	config            *EnergyConfig
	logger            *slog.Logger
	efficiencyData    map[string]*EfficiencyTrend
	benchmarkData     map[string]float64
	optimizationRules []EnergyOptimizationRule
}

// EfficiencyTrend tracks energy efficiency trends over time
type EfficiencyTrend struct {
	JobID            string
	Timestamps       []time.Time
	EfficiencyScores []float64
	Trend            string // "improving", "declining", "stable"
	TrendStrength    float64
}

// EnergyOptimizationRule defines energy optimization recommendations
type EnergyOptimizationRule struct {
	RuleID          string
	Condition       string
	Recommendation  string
	ExpectedSavings float64
	Priority        string
}

// EnergyMetrics holds Prometheus metrics for energy monitoring
type EnergyMetrics struct {
	// Energy consumption metrics
	JobEnergyConsumption *prometheus.GaugeVec
	JobPowerConsumption  *prometheus.GaugeVec
	JobEnergyEfficiency  *prometheus.GaugeVec
	JobEnergyWaste       *prometheus.GaugeVec

	// Carbon footprint metrics
	JobCarbonEmissions     *prometheus.GaugeVec
	JobCarbonEfficiency    *prometheus.GaugeVec
	JobCarbonWaste         *prometheus.GaugeVec
	JobSustainabilityScore *prometheus.GaugeVec

	// Energy cost metrics
	JobEnergyCost     *prometheus.GaugeVec
	JobCostEfficiency *prometheus.GaugeVec
	JobEnergyROI      *prometheus.GaugeVec

	// Cluster metrics
	ClusterTotalEnergy      *prometheus.GaugeVec
	ClusterTotalPower       *prometheus.GaugeVec
	ClusterEnergyEfficiency *prometheus.GaugeVec
	ClusterCarbonEmissions  *prometheus.GaugeVec

	// Per-resource energy metrics
	CPUEnergyConsumption     *prometheus.GaugeVec
	MemoryEnergyConsumption  *prometheus.GaugeVec
	GPUEnergyConsumption     *prometheus.GaugeVec
	NetworkEnergyConsumption *prometheus.GaugeVec
	StorageEnergyConsumption *prometheus.GaugeVec

	// Carbon intensity metrics
	CarbonIntensityGCO2kWh *prometheus.GaugeVec
	GreenEnergyPercentage  *prometheus.GaugeVec
	CarbonOffsetCredits    *prometheus.GaugeVec

	// Energy optimization metrics
	EnergyOptimizationOpportunities *prometheus.GaugeVec
	EnergySavingsPotential          *prometheus.GaugeVec
	PowerCappingEvents              *prometheus.CounterVec

	// Collection metrics
	EnergyDataCollectionDuration *prometheus.HistogramVec
	EnergyDataCollectionErrors   *prometheus.CounterVec
	EnergyMonitoredJobsCount     *prometheus.GaugeVec
}

// NewEnergyMonitor creates a new energy monitoring collector
func NewEnergyMonitor(client slurm.SlurmClient, logger *slog.Logger, config *EnergyConfig) (*EnergyMonitor, error) {
	if config == nil {
		config = &EnergyConfig{
			MonitoringInterval:        60 * time.Second,
			MaxJobsPerCollection:      100,
			EnableCarbonTracking:      true,
			EnableEfficiencyTracking:  true,
			EnergyCalculationMethod:   "power_model",
			DefaultCPUPowerWatts:      15.0,  // 15W per CPU core
			DefaultMemoryPowerWatts:   1.5,   // 1.5W per GB
			DefaultGPUPowerWatts:      250.0, // 250W per GPU
			DefaultNodeBasePowerWatts: 50.0,  // 50W base power per node
			CarbonIntensityGCO2kWh:    400.0, // 400g CO2/kWh (average grid)
			CarbonTrackingRegion:      "US",
			EnableRealTimeCarbonAPI:   false,
			EnablePowerCapping:        false,
			PowerCapThresholdWatts:    1000.0,
			EnergyBudgetPerJobWh:      1000.0,
			EnergyDataRetention:       7 * 24 * time.Hour,
			CarbonDataRetention:       30 * 24 * time.Hour,
		}
	}

	carbonCalculator := &CarbonFootprintCalculator{
		config:          config,
		logger:          logger,
		carbonIntensity: config.CarbonIntensityGCO2kWh,
		regionalFactors: getRegionalCarbonFactors(),
		timeFactors:     getTimeOfDayFactors(),
	}

	efficiencyTracker := &EnergyEfficiencyTracker{
		config:            config,
		logger:            logger,
		efficiencyData:    make(map[string]*EfficiencyTrend),
		benchmarkData:     getEnergyBenchmarks(),
		optimizationRules: getEnergyOptimizationRules(),
	}

	return &EnergyMonitor{
		slurmClient:       client,
		logger:            logger,
		config:            config,
		metrics:           newEnergyMetrics(),
		energyData:        make(map[string]*JobEnergyData),
		carbonData:        make(map[string]*CarbonFootprintData),
		clusterEnergy:     &ClusterEnergyData{},
		carbonCalculator:  carbonCalculator,
		efficiencyTracker: efficiencyTracker,
	}, nil
}

// newEnergyMetrics creates Prometheus metrics for energy monitoring
func newEnergyMetrics() *EnergyMetrics {
	return &EnergyMetrics{
		JobEnergyConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_energy_consumption_wh",
				Help: "Total energy consumption for jobs in watt-hours",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobPowerConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_power_consumption_watts",
				Help: "Average power consumption for jobs in watts",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobEnergyEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_energy_efficiency",
				Help: "Energy efficiency score for jobs (0-1)",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobEnergyWaste: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_energy_waste_wh",
				Help: "Wasted energy for jobs in watt-hours",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobCarbonEmissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_carbon_emissions_grams",
				Help: "Carbon emissions for jobs in grams CO2",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobCarbonEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_carbon_efficiency",
				Help: "Carbon efficiency score for jobs (0-1)",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobCarbonWaste: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_carbon_waste_grams",
				Help: "Wasted carbon emissions for jobs in grams CO2",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobSustainabilityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_sustainability_score",
				Help: "Overall sustainability score for jobs (0-1)",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobEnergyCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_energy_cost_usd",
				Help: "Energy cost for jobs in USD",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobCostEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_energy_cost_efficiency",
				Help: "Energy cost efficiency for jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobEnergyROI: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_energy_roi",
				Help: "Energy return on investment for jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		ClusterTotalEnergy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_cluster_total_energy_wh",
				Help: "Total cluster energy consumption in watt-hours",
			},
			[]string{"cluster"},
		),
		ClusterTotalPower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_cluster_total_power_watts",
				Help: "Total cluster power consumption in watts",
			},
			[]string{"cluster"},
		),
		ClusterEnergyEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_cluster_energy_efficiency",
				Help: "Cluster-wide energy efficiency score",
			},
			[]string{"cluster"},
		),
		ClusterCarbonEmissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_cluster_carbon_emissions_grams",
				Help: "Total cluster carbon emissions in grams CO2",
			},
			[]string{"cluster"},
		),
		CPUEnergyConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_cpu_energy_consumption_wh",
				Help: "CPU energy consumption for jobs in watt-hours",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		MemoryEnergyConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_memory_energy_consumption_wh",
				Help: "Memory energy consumption for jobs in watt-hours",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		GPUEnergyConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_gpu_energy_consumption_wh",
				Help: "GPU energy consumption for jobs in watt-hours",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		NetworkEnergyConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_network_energy_consumption_wh",
				Help: "Network energy consumption for jobs in watt-hours",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		StorageEnergyConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_storage_energy_consumption_wh",
				Help: "Storage I/O energy consumption for jobs in watt-hours",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		CarbonIntensityGCO2kWh: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_carbon_intensity_gco2_kwh",
				Help: "Current carbon intensity in grams CO2 per kWh",
			},
			[]string{"region"},
		),
		GreenEnergyPercentage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_green_energy_percentage",
				Help: "Percentage of energy from renewable sources",
			},
			[]string{"region"},
		),
		CarbonOffsetCredits: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_carbon_offset_credits",
				Help: "Available carbon offset credits",
			},
			[]string{"account"},
		),
		EnergyOptimizationOpportunities: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_energy_optimization_opportunities",
				Help: "Number of energy optimization opportunities identified",
			},
			[]string{"job_id", "user", "account", "partition", "opportunity_type"},
		),
		EnergySavingsPotential: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_energy_savings_potential_wh",
				Help: "Potential energy savings in watt-hours",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		PowerCappingEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_power_capping_events_total",
				Help: "Total number of power capping events",
			},
			[]string{"job_id", "user", "account", "partition", "reason"},
		),
		EnergyDataCollectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_energy_data_collection_duration_seconds",
				Help:    "Duration of energy data collection operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		EnergyDataCollectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_energy_data_collection_errors_total",
				Help: "Total number of energy data collection errors",
			},
			[]string{"operation", "error_type"},
		),
		EnergyMonitoredJobsCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_energy_monitored_jobs_count",
				Help: "Number of jobs currently being monitored for energy",
			},
			[]string{"status"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (e *EnergyMonitor) Describe(ch chan<- *prometheus.Desc) {
	e.metrics.JobEnergyConsumption.Describe(ch)
	e.metrics.JobPowerConsumption.Describe(ch)
	e.metrics.JobEnergyEfficiency.Describe(ch)
	e.metrics.JobEnergyWaste.Describe(ch)
	e.metrics.JobCarbonEmissions.Describe(ch)
	e.metrics.JobCarbonEfficiency.Describe(ch)
	e.metrics.JobCarbonWaste.Describe(ch)
	e.metrics.JobSustainabilityScore.Describe(ch)
	e.metrics.JobEnergyCost.Describe(ch)
	e.metrics.JobCostEfficiency.Describe(ch)
	e.metrics.JobEnergyROI.Describe(ch)
	e.metrics.ClusterTotalEnergy.Describe(ch)
	e.metrics.ClusterTotalPower.Describe(ch)
	e.metrics.ClusterEnergyEfficiency.Describe(ch)
	e.metrics.ClusterCarbonEmissions.Describe(ch)
	e.metrics.CPUEnergyConsumption.Describe(ch)
	e.metrics.MemoryEnergyConsumption.Describe(ch)
	e.metrics.GPUEnergyConsumption.Describe(ch)
	e.metrics.NetworkEnergyConsumption.Describe(ch)
	e.metrics.StorageEnergyConsumption.Describe(ch)
	e.metrics.CarbonIntensityGCO2kWh.Describe(ch)
	e.metrics.GreenEnergyPercentage.Describe(ch)
	e.metrics.CarbonOffsetCredits.Describe(ch)
	e.metrics.EnergyOptimizationOpportunities.Describe(ch)
	e.metrics.EnergySavingsPotential.Describe(ch)
	e.metrics.PowerCappingEvents.Describe(ch)
	e.metrics.EnergyDataCollectionDuration.Describe(ch)
	e.metrics.EnergyDataCollectionErrors.Describe(ch)
	e.metrics.EnergyMonitoredJobsCount.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (e *EnergyMonitor) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if err := e.collectEnergyMetrics(ctx); err != nil {
		e.logger.Error("Failed to collect energy metrics", "error", err)
		e.metrics.EnergyDataCollectionErrors.WithLabelValues("collect", "collection_error").Inc()
	}

	e.metrics.JobEnergyConsumption.Collect(ch)
	e.metrics.JobPowerConsumption.Collect(ch)
	e.metrics.JobEnergyEfficiency.Collect(ch)
	e.metrics.JobEnergyWaste.Collect(ch)
	e.metrics.JobCarbonEmissions.Collect(ch)
	e.metrics.JobCarbonEfficiency.Collect(ch)
	e.metrics.JobCarbonWaste.Collect(ch)
	e.metrics.JobSustainabilityScore.Collect(ch)
	e.metrics.JobEnergyCost.Collect(ch)
	e.metrics.JobCostEfficiency.Collect(ch)
	e.metrics.JobEnergyROI.Collect(ch)
	e.metrics.ClusterTotalEnergy.Collect(ch)
	e.metrics.ClusterTotalPower.Collect(ch)
	e.metrics.ClusterEnergyEfficiency.Collect(ch)
	e.metrics.ClusterCarbonEmissions.Collect(ch)
	e.metrics.CPUEnergyConsumption.Collect(ch)
	e.metrics.MemoryEnergyConsumption.Collect(ch)
	e.metrics.GPUEnergyConsumption.Collect(ch)
	e.metrics.NetworkEnergyConsumption.Collect(ch)
	e.metrics.StorageEnergyConsumption.Collect(ch)
	e.metrics.CarbonIntensityGCO2kWh.Collect(ch)
	e.metrics.GreenEnergyPercentage.Collect(ch)
	e.metrics.CarbonOffsetCredits.Collect(ch)
	e.metrics.EnergyOptimizationOpportunities.Collect(ch)
	e.metrics.EnergySavingsPotential.Collect(ch)
	e.metrics.PowerCappingEvents.Collect(ch)
	e.metrics.EnergyDataCollectionDuration.Collect(ch)
	e.metrics.EnergyDataCollectionErrors.Collect(ch)
	e.metrics.EnergyMonitoredJobsCount.Collect(ch)
}

// collectEnergyMetrics collects energy consumption and carbon footprint data
func (e *EnergyMonitor) collectEnergyMetrics(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		e.metrics.EnergyDataCollectionDuration.WithLabelValues("collect_energy_metrics").Observe(time.Since(startTime).Seconds())
	}()

	// Get running and recently completed jobs
	jobManager := e.slurmClient.Jobs()
	// Using nil for options as the exact structure is not clear
	jobs, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs for energy monitoring: %w", err)
	}

	e.metrics.EnergyMonitoredJobsCount.WithLabelValues("total").Set(float64(len(jobs.Jobs)))

	var clusterTotalEnergy, clusterTotalPower, clusterTotalCarbon float64

	// Process each job for energy metrics
	for _, job := range jobs.Jobs {
		if err := e.processJobEnergyMetrics(ctx, &job); err != nil {
			e.logger.Error("Failed to process job energy metrics", "job_id", job.ID, "error", err)
			e.metrics.EnergyDataCollectionErrors.WithLabelValues("process_job", "job_error").Inc()
			continue
		}

		// Accumulate cluster totals
		jobID := job.ID
		if energyData, exists := e.energyData[jobID]; exists {
			clusterTotalEnergy += energyData.TotalEnergyWh
			clusterTotalPower += energyData.AveragePowerW
		}

		if carbonData, exists := e.carbonData[jobID]; exists {
			clusterTotalCarbon += carbonData.TotalCO2grams
		}
	}

	// Update cluster-wide metrics
	e.updateClusterEnergyMetrics(clusterTotalEnergy, clusterTotalPower, clusterTotalCarbon)

	// Clean old data
	e.cleanOldEnergyData()

	e.lastCollection = time.Now()
	return nil
}

// processJobEnergyMetrics processes energy metrics for a single job
func (e *EnergyMonitor) processJobEnergyMetrics(ctx context.Context, job *slurm.Job) error {
	// Calculate energy consumption for the job
	energyData := e.calculateJobEnergyConsumption(job)

	// Calculate carbon footprint if enabled
	var carbonData *CarbonFootprintData
	if e.config.EnableCarbonTracking {
		carbonData = e.carbonCalculator.calculateCarbonFootprint(job, energyData)
	}

	// Track efficiency trends if enabled
	if e.config.EnableEfficiencyTracking {
		e.efficiencyTracker.trackEnergyEfficiency(job.ID, energyData.PowerEfficiency)
	}

	// Store data
	e.mu.Lock()
	e.energyData[job.ID] = energyData
	if carbonData != nil {
		e.carbonData[job.ID] = carbonData
	}
	e.mu.Unlock()

	// Update Prometheus metrics
	e.updateJobEnergyMetrics(job, energyData, carbonData)

	return nil
}

// calculateJobEnergyConsumption calculates energy consumption for a job
func (e *EnergyMonitor) calculateJobEnergyConsumption(job *slurm.Job) *JobEnergyData {
	now := time.Now()

	// Calculate runtime
	var runtimeHours float64
	if job.StartTime != nil {
		if job.State == "COMPLETED" && job.EndTime != nil {
			runtimeHours = job.EndTime.Sub(*job.StartTime).Hours()
		} else {
			runtimeHours = time.Since(*job.StartTime).Hours()
		}
	} else {
		runtimeHours = 1.0 // Default to 1 hour if no start time
	}

	// Calculate resource-specific energy consumption based on power model
	cpuEnergyWh := e.calculateCPUEnergy(job, runtimeHours)
	memoryEnergyWh := e.calculateMemoryEnergy(job, runtimeHours)
	gpuEnergyWh := e.calculateGPUEnergy(job, runtimeHours)
	storageEnergyWh := e.calculateStorageEnergy(job, runtimeHours)
	networkEnergyWh := e.calculateNetworkEnergy(job, runtimeHours)

	// Base node power consumption
	nodeBasePowerWh := e.config.DefaultNodeBasePowerWatts * float64(len(job.Nodes)) * runtimeHours

	// Cooling energy (estimated as 30% of compute energy for data centers)
	computeEnergyWh := cpuEnergyWh + memoryEnergyWh + gpuEnergyWh
	coolingEnergyWh := computeEnergyWh * 0.3

	// Total energy consumption
	totalEnergyWh := nodeBasePowerWh + cpuEnergyWh + memoryEnergyWh + gpuEnergyWh +
		storageEnergyWh + networkEnergyWh + coolingEnergyWh

	// Calculate power metrics
	averagePowerW := totalEnergyWh / math.Max(runtimeHours, 0.001) // Avoid division by zero
	peakPowerW := averagePowerW * 1.5                              // Estimate peak as 50% higher than average

	// Calculate efficiency metrics
	powerEfficiency := e.calculatePowerEfficiency(job, totalEnergyWh, runtimeHours)
	energyPerCPUHour := totalEnergyWh / (float64(job.CPUs) * math.Max(runtimeHours, 0.001))
	energyPerTaskUnit := e.calculateEnergyPerTaskUnit(job, totalEnergyWh)
	energyWasteWh := e.calculateEnergyWaste(job, totalEnergyWh, powerEfficiency)

	// Calculate cost metrics
	energyCostUSD := e.calculateEnergyCost(totalEnergyWh)
	costPerCPUHour := energyCostUSD / (float64(job.CPUs) * math.Max(runtimeHours, 0.001))
	costEfficiency := e.calculateCostEfficiency(energyCostUSD, totalEnergyWh, powerEfficiency)

	return &JobEnergyData{
		JobID:             job.ID,
		Timestamp:         now,
		TotalEnergyWh:     totalEnergyWh,
		CPUEnergyWh:       cpuEnergyWh,
		MemoryEnergyWh:    memoryEnergyWh,
		GPUEnergyWh:       gpuEnergyWh,
		StorageEnergyWh:   storageEnergyWh,
		NetworkEnergyWh:   networkEnergyWh,
		CoolingEnergyWh:   coolingEnergyWh,
		AveragePowerW:     averagePowerW,
		PeakPowerW:        peakPowerW,
		PowerEfficiency:   powerEfficiency,
		EnergyPerCPUHour:  energyPerCPUHour,
		EnergyPerTaskUnit: energyPerTaskUnit,
		EnergyWasteWh:     energyWasteWh,
		EnergyCostUSD:     energyCostUSD,
		CostPerCPUHour:    costPerCPUHour,
		CostEfficiency:    costEfficiency,
	}
}

// calculateCPUEnergy calculates CPU energy consumption
func (e *EnergyMonitor) calculateCPUEnergy(job *slurm.Job, runtimeHours float64) float64 {
	// CPU energy = CPUs × power per CPU × runtime × utilization factor
	utilizationFactor := e.estimateCPUUtilization(job)
	return float64(job.CPUs) * e.config.DefaultCPUPowerWatts * runtimeHours * utilizationFactor
}

// calculateMemoryEnergy calculates memory energy consumption
func (e *EnergyMonitor) calculateMemoryEnergy(job *slurm.Job, runtimeHours float64) float64 {
	// Memory energy = Memory (GB) × power per GB × runtime
	memoryGB := float64(job.Memory) / 1024 // Convert MB to GB
	return memoryGB * e.config.DefaultMemoryPowerWatts * runtimeHours
}

// calculateGPUEnergy calculates GPU energy consumption
func (e *EnergyMonitor) calculateGPUEnergy(job *slurm.Job, runtimeHours float64) float64 {
	// Estimate GPU count based on job characteristics
	gpuCount := e.estimateGPUCount(job)
	if gpuCount == 0 {
		return 0
	}

	utilizationFactor := e.estimateGPUUtilization(job)
	return float64(gpuCount) * e.config.DefaultGPUPowerWatts * runtimeHours * utilizationFactor
}

// calculateStorageEnergy calculates storage I/O energy consumption
func (e *EnergyMonitor) calculateStorageEnergy(job *slurm.Job, runtimeHours float64) float64 {
	// Estimate storage energy based on job characteristics and I/O patterns
	baseStoragePower := 5.0 // 5W base storage power per job
	ioIntensityFactor := e.estimateIOIntensity(job)
	return baseStoragePower * runtimeHours * ioIntensityFactor
}

// calculateNetworkEnergy calculates network energy consumption
func (e *EnergyMonitor) calculateNetworkEnergy(job *slurm.Job, runtimeHours float64) float64 {
	// Estimate network energy based on job characteristics and communication patterns
	baseNetworkPower := 2.0 // 2W base network power per job
	networkIntensityFactor := e.estimateNetworkIntensity(job)
	return baseNetworkPower * runtimeHours * networkIntensityFactor * float64(len(job.Nodes))
}

// estimateCPUUtilization estimates CPU utilization for energy calculation
func (e *EnergyMonitor) estimateCPUUtilization(job *slurm.Job) float64 {
	// Simple heuristic based on job characteristics
	// In reality, this would use actual utilization data
	baseUtilization := 0.7 // 70% base utilization

	// Adjust based on job type heuristics
	if job.CPUs > 16 {
		baseUtilization += 0.1 // High-CPU jobs tend to have higher utilization
	}
	if float64(job.Memory)/float64(job.CPUs) > 4000 { // Memory-intensive jobs
		baseUtilization -= 0.1 // May have lower CPU utilization
	}

	// Add some job-specific variation
	// Use job ID hash for variation
	idHash := 0
	for _, c := range job.ID {
		idHash += int(c)
	}
	variation := float64(idHash%20) / 100.0 // 0-19% variation
	baseUtilization += variation - 0.1

	// Clamp between 0.1 and 1.0
	return math.Max(0.1, math.Min(1.0, baseUtilization))
}

// estimateGPUCount estimates GPU count for a job
func (e *EnergyMonitor) estimateGPUCount(job *slurm.Job) int {
	// Simple heuristic - assume GPU jobs have high CPU and memory allocation
	if job.CPUs >= 8 && job.Memory >= 32768 { // 8+ CPUs and 32+ GB
		return job.CPUs / 8 // Assume 1 GPU per 8 CPUs
	}
	return 0
}

// estimateGPUUtilization estimates GPU utilization
func (e *EnergyMonitor) estimateGPUUtilization(job *slurm.Job) float64 {
	// GPU utilization is typically lower than CPU utilization
	return e.estimateCPUUtilization(job) * 0.8
}

// estimateIOIntensity estimates I/O intensity factor
func (e *EnergyMonitor) estimateIOIntensity(job *slurm.Job) float64 {
	baseIntensity := 1.0

	// High-memory jobs may be I/O intensive
	if float64(job.Memory)/float64(job.CPUs) > 8000 {
		baseIntensity += 0.5
	}

	// Multi-node jobs may have more I/O
	if len(job.Nodes) > 1 {
		baseIntensity += 0.3
	}

	return math.Min(2.0, baseIntensity)
}

// estimateNetworkIntensity estimates network intensity factor
func (e *EnergyMonitor) estimateNetworkIntensity(job *slurm.Job) float64 {
	baseIntensity := 1.0

	// Multi-node jobs have higher network intensity
	if len(job.Nodes) > 1 {
		baseIntensity += float64(len(job.Nodes)-1) * 0.2
	}

	// High-CPU jobs may be communication-intensive
	if job.CPUs > 32 {
		baseIntensity += 0.3
	}

	return math.Min(3.0, baseIntensity)
}

// calculatePowerEfficiency calculates power utilization efficiency
func (e *EnergyMonitor) calculatePowerEfficiency(job *slurm.Job, totalEnergyWh, runtimeHours float64) float64 {
	if runtimeHours <= 0 {
		return 0
	}

	// Calculate theoretical minimum power based on allocated resources
	minCPUPower := float64(job.CPUs) * e.config.DefaultCPUPowerWatts * 0.1                // 10% minimum CPU power
	minMemoryPower := float64(job.Memory) / 1024 * e.config.DefaultMemoryPowerWatts * 0.5 // 50% memory power
	minBasePower := e.config.DefaultNodeBasePowerWatts * float64(len(job.Nodes))

	theoreticalMinEnergy := (minCPUPower + minMemoryPower + minBasePower) * runtimeHours

	// Efficiency = work done / energy consumed
	// Here we approximate work done as inversely related to energy waste
	if totalEnergyWh <= theoreticalMinEnergy {
		return 1.0 // Perfect efficiency if at minimum
	}

	efficiency := theoreticalMinEnergy / totalEnergyWh
	return math.Max(0.0, math.Min(1.0, efficiency))
}

// calculateEnergyPerTaskUnit calculates energy consumption per unit of work
func (e *EnergyMonitor) calculateEnergyPerTaskUnit(job *slurm.Job, totalEnergyWh float64) float64 {
	// Simple heuristic - assume task units are CPU-hours
	var runtimeHours float64
	if job.StartTime != nil {
		if job.State == "COMPLETED" && job.EndTime != nil {
			runtimeHours = job.EndTime.Sub(*job.StartTime).Hours()
		} else {
			runtimeHours = time.Since(*job.StartTime).Hours()
		}
	} else {
		runtimeHours = 1.0
	}

	taskUnits := float64(job.CPUs) * runtimeHours
	if taskUnits <= 0 {
		return totalEnergyWh
	}

	return totalEnergyWh / taskUnits
}

// calculateEnergyWaste calculates wasted energy due to inefficiency
func (e *EnergyMonitor) calculateEnergyWaste(job *slurm.Job, totalEnergyWh, powerEfficiency float64) float64 {
	// Energy waste = total energy × (1 - efficiency)
	wasteRatio := 1.0 - powerEfficiency
	return totalEnergyWh * wasteRatio
}

// calculateEnergyCost calculates energy cost in USD
func (e *EnergyMonitor) calculateEnergyCost(energyWh float64) float64 {
	// Convert Wh to kWh and multiply by cost per kWh
	energyKWh := energyWh / 1000.0
	costPerKWh := 0.12 // $0.12 per kWh (average US rate)
	return energyKWh * costPerKWh
}

// calculateCostEfficiency calculates cost efficiency score
func (e *EnergyMonitor) calculateCostEfficiency(costUSD, energyWh, powerEfficiency float64) float64 {
	// Cost efficiency combines power efficiency with cost-effectiveness
	if costUSD <= 0 {
		return 1.0
	}

	// Normalize cost per Wh
	costPerWh := costUSD / energyWh
	maxReasonableCostPerWh := 0.0002 // $0.0002 per Wh

	costEfficiencyFactor := math.Max(0.1, math.Min(1.0, maxReasonableCostPerWh/costPerWh))

	// Combine with power efficiency
	return (powerEfficiency + costEfficiencyFactor) / 2.0
}

// calculateCarbonFootprint calculates carbon footprint for a job
func (c *CarbonFootprintCalculator) calculateCarbonFootprint(job *slurm.Job, energyData *JobEnergyData) *CarbonFootprintData {
	now := time.Now()

	// Get current carbon intensity
	carbonIntensity := c.getCurrentCarbonIntensity(now)

	// Calculate total CO2 emissions
	energyKWh := energyData.TotalEnergyWh / 1000.0
	totalCO2grams := energyKWh * carbonIntensity

	// Calculate per-unit emissions
	var runtimeHours float64
	if job.StartTime != nil {
		if job.State == "COMPLETED" && job.EndTime != nil {
			runtimeHours = job.EndTime.Sub(*job.StartTime).Hours()
		} else {
			runtimeHours = time.Since(*job.StartTime).Hours()
		}
	} else {
		runtimeHours = 1.0
	}

	cpuHours := float64(job.CPUs) * runtimeHours
	co2PerCPUHour := totalCO2grams / math.Max(cpuHours, 0.001)
	co2PerTaskUnit := totalCO2grams / math.Max(cpuHours, 0.001) // Same as CPU-hour for now

	// Calculate carbon efficiency
	carbonEfficiency := c.calculateCarbonEfficiency(energyData.PowerEfficiency, carbonIntensity)

	// Calculate carbon waste
	carbonWasteGrams := totalCO2grams * (1.0 - carbonEfficiency)

	// Calculate sustainability metrics
	greenEnergyPercentage := c.getGreenEnergyPercentage(now)
	sustainabilityScore := c.calculateSustainabilityScore(carbonEfficiency, greenEnergyPercentage)
	sustainabilityGrade := c.calculateSustainabilityGrade(sustainabilityScore)

	// Environmental impact score (0-1, lower is better)
	environmentalImpact := c.calculateEnvironmentalImpact(totalCO2grams, energyData.TotalEnergyWh)

	return &CarbonFootprintData{
		JobID:                    job.ID,
		Timestamp:                now,
		TotalCO2grams:            totalCO2grams,
		CO2PerCPUHour:            co2PerCPUHour,
		CO2PerTaskUnit:           co2PerTaskUnit,
		CarbonIntensity:          carbonIntensity,
		RegionalFactor:           c.regionalFactors[c.config.CarbonTrackingRegion],
		TimeOfDayFactor:          c.timeFactors[now.Hour()],
		CarbonEfficiency:         carbonEfficiency,
		CarbonWasteGrams:         carbonWasteGrams,
		CarbonOffsetGrams:        0, // Would be populated from offset programs
		NetCarbonGrams:           totalCO2grams,
		CarbonCreditsUsed:        0, // Would be calculated based on credit usage
		EnvironmentalImpactScore: environmentalImpact,
		SustainabilityGrade:      sustainabilityGrade,
		GreenEnergyPercentage:    greenEnergyPercentage,
	}
}

// getCurrentCarbonIntensity gets current carbon intensity
func (c *CarbonFootprintCalculator) getCurrentCarbonIntensity(timestamp time.Time) float64 {
	// Base carbon intensity
	intensity := c.config.CarbonIntensityGCO2kWh

	// Apply regional factor
	if regionalFactor, exists := c.regionalFactors[c.config.CarbonTrackingRegion]; exists {
		intensity *= regionalFactor
	}

	// Apply time-of-day factor
	if timeFactor, exists := c.timeFactors[timestamp.Hour()]; exists {
		intensity *= timeFactor
	}

	return intensity
}

// calculateCarbonEfficiency calculates carbon efficiency score
func (c *CarbonFootprintCalculator) calculateCarbonEfficiency(powerEfficiency, carbonIntensity float64) float64 {
	// Carbon efficiency is inversely related to carbon intensity and directly related to power efficiency
	baseCarbonIntensity := 500.0 // 500g CO2/kWh baseline
	carbonIntensityFactor := baseCarbonIntensity / carbonIntensity

	// Combine power efficiency with carbon intensity factor
	carbonEfficiency := (powerEfficiency + carbonIntensityFactor) / 2.0
	return math.Max(0.0, math.Min(1.0, carbonEfficiency))
}

// getGreenEnergyPercentage gets percentage of energy from renewable sources
func (c *CarbonFootprintCalculator) getGreenEnergyPercentage(timestamp time.Time) float64 {
	// Simulate green energy percentage based on time and region
	baseGreenPercentage := 0.30 // 30% base renewable energy

	// Higher renewable percentage during day hours (solar)
	hour := timestamp.Hour()
	if hour >= 10 && hour <= 16 {
		baseGreenPercentage += 0.20 // +20% during peak solar hours
	}

	// Regional variations
	if c.config.CarbonTrackingRegion == "EU" {
		baseGreenPercentage += 0.25 // EU has higher renewable percentage
	}

	return math.Min(1.0, baseGreenPercentage)
}

// calculateSustainabilityScore calculates overall sustainability score
func (c *CarbonFootprintCalculator) calculateSustainabilityScore(carbonEfficiency, greenEnergyPercentage float64) float64 {
	// Weighted combination of carbon efficiency and green energy percentage
	return (carbonEfficiency*0.6 + greenEnergyPercentage*0.4)
}

// calculateSustainabilityGrade calculates sustainability grade
func (c *CarbonFootprintCalculator) calculateSustainabilityGrade(score float64) string {
	switch {
	case score >= 0.9:
		return "A"
	case score >= 0.8:
		return "B"
	case score >= 0.7:
		return "C"
	case score >= 0.6:
		return "D"
	default:
		return "F"
	}
}

// calculateEnvironmentalImpact calculates environmental impact score
func (c *CarbonFootprintCalculator) calculateEnvironmentalImpact(co2Grams, energyWh float64) float64 {
	// Normalize impact based on energy consumption and emissions
	// This is a simplified model - real implementations would be more sophisticated

	energyKWh := energyWh / 1000.0
	impactPerKWh := co2Grams / energyKWh

	// Normalize against average grid emissions (400g CO2/kWh)
	normalizedImpact := impactPerKWh / 400.0

	return math.Max(0.0, math.Min(1.0, normalizedImpact))
}

// trackEnergyEfficiency tracks energy efficiency trends
func (t *EnergyEfficiencyTracker) trackEnergyEfficiency(jobID string, efficiency float64) {
	now := time.Now()

	if trend, exists := t.efficiencyData[jobID]; exists {
		trend.Timestamps = append(trend.Timestamps, now)
		trend.EfficiencyScores = append(trend.EfficiencyScores, efficiency)

		// Keep only recent data (last 24 hours)
		cutoff := now.Add(-24 * time.Hour)
		var newTimestamps []time.Time
		var newScores []float64

		for i, timestamp := range trend.Timestamps {
			if timestamp.After(cutoff) {
				newTimestamps = append(newTimestamps, timestamp)
				newScores = append(newScores, trend.EfficiencyScores[i])
			}
		}

		trend.Timestamps = newTimestamps
		trend.EfficiencyScores = newScores

		// Calculate trend
		trend.Trend, trend.TrendStrength = t.calculateTrend(newScores)
	} else {
		t.efficiencyData[jobID] = &EfficiencyTrend{
			JobID:            jobID,
			Timestamps:       []time.Time{now},
			EfficiencyScores: []float64{efficiency},
			Trend:            "stable",
			TrendStrength:    0.0,
		}
	}
}

// calculateTrend calculates trend direction and strength
func (t *EnergyEfficiencyTracker) calculateTrend(scores []float64) (string, float64) {
	if len(scores) < 3 {
		return "stable", 0.0
	}

	// Simple linear regression to determine trend
	n := float64(len(scores))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, score := range scores {
		x := float64(i)
		sumX += x
		sumY += score
		sumXY += x * score
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Determine trend direction
	direction := "stable"
	strength := math.Abs(slope)

	if slope > 0.01 {
		direction = "improving"
	} else if slope < -0.01 {
		direction = "declining"
	}

	return direction, strength
}

// updateJobEnergyMetrics updates Prometheus metrics for a job
func (e *EnergyMonitor) updateJobEnergyMetrics(job *slurm.Job, energyData *JobEnergyData, carbonData *CarbonFootprintData) {
	labels := []string{job.ID, "", "", job.Partition} // TODO: job.UserName and job.Account fields not available

	// Update energy metrics
	e.metrics.JobEnergyConsumption.WithLabelValues(labels...).Set(energyData.TotalEnergyWh)
	e.metrics.JobPowerConsumption.WithLabelValues(labels...).Set(energyData.AveragePowerW)
	e.metrics.JobEnergyEfficiency.WithLabelValues(labels...).Set(energyData.PowerEfficiency)
	e.metrics.JobEnergyWaste.WithLabelValues(labels...).Set(energyData.EnergyWasteWh)
	e.metrics.JobEnergyCost.WithLabelValues(labels...).Set(energyData.EnergyCostUSD)
	e.metrics.JobCostEfficiency.WithLabelValues(labels...).Set(energyData.CostEfficiency)

	// Update per-resource energy metrics
	e.metrics.CPUEnergyConsumption.WithLabelValues(labels...).Set(energyData.CPUEnergyWh)
	e.metrics.MemoryEnergyConsumption.WithLabelValues(labels...).Set(energyData.MemoryEnergyWh)
	e.metrics.GPUEnergyConsumption.WithLabelValues(labels...).Set(energyData.GPUEnergyWh)
	e.metrics.NetworkEnergyConsumption.WithLabelValues(labels...).Set(energyData.NetworkEnergyWh)
	e.metrics.StorageEnergyConsumption.WithLabelValues(labels...).Set(energyData.StorageEnergyWh)

	// Update carbon metrics if available
	if carbonData != nil {
		e.metrics.JobCarbonEmissions.WithLabelValues(labels...).Set(carbonData.TotalCO2grams)
		e.metrics.JobCarbonEfficiency.WithLabelValues(labels...).Set(carbonData.CarbonEfficiency)
		e.metrics.JobCarbonWaste.WithLabelValues(labels...).Set(carbonData.CarbonWasteGrams)
		e.metrics.JobSustainabilityScore.WithLabelValues(labels...).Set(carbonData.EnvironmentalImpactScore)

		// Update global carbon metrics
		e.metrics.CarbonIntensityGCO2kWh.WithLabelValues(e.config.CarbonTrackingRegion).Set(carbonData.CarbonIntensity)
		e.metrics.GreenEnergyPercentage.WithLabelValues(e.config.CarbonTrackingRegion).Set(carbonData.GreenEnergyPercentage)
	}

	// Calculate and update ROI
	if energyData.EnergyCostUSD > 0 {
		// Simple ROI calculation based on efficiency
		roi := energyData.PowerEfficiency / energyData.EnergyCostUSD * 100
		e.metrics.JobEnergyROI.WithLabelValues(labels...).Set(roi)
	}

	// Update optimization opportunities
	savingsPotential := energyData.EnergyWasteWh
	e.metrics.EnergySavingsPotential.WithLabelValues(labels...).Set(savingsPotential)

	if energyData.PowerEfficiency < 0.7 {
		e.metrics.EnergyOptimizationOpportunities.WithLabelValues(append(labels, "efficiency_improvement")...).Set(1)
	}
	if energyData.EnergyWasteWh > energyData.TotalEnergyWh*0.3 {
		e.metrics.EnergyOptimizationOpportunities.WithLabelValues(append(labels, "waste_reduction")...).Set(1)
	}
}

// updateClusterEnergyMetrics updates cluster-wide energy metrics
func (e *EnergyMonitor) updateClusterEnergyMetrics(totalEnergy, totalPower, totalCarbon float64) {
	clusterName := "default"

	e.clusterEnergy.Timestamp = time.Now()
	e.clusterEnergy.TotalClusterEnergyWh = totalEnergy
	e.clusterEnergy.TotalClusterPowerW = totalPower
	e.clusterEnergy.TotalClusterCO2grams = totalCarbon

	// Calculate cluster efficiency (simplified)
	jobCount := float64(len(e.energyData))
	if jobCount > 0 {
		e.clusterEnergy.AverageNodePowerW = totalPower / jobCount

		// Calculate cluster-wide efficiency
		totalEfficiency := 0.0
		for _, energyData := range e.energyData {
			totalEfficiency += energyData.PowerEfficiency
		}
		e.clusterEnergy.ClusterPowerEfficiency = totalEfficiency / jobCount
	}

	// Update Prometheus metrics
	e.metrics.ClusterTotalEnergy.WithLabelValues(clusterName).Set(totalEnergy)
	e.metrics.ClusterTotalPower.WithLabelValues(clusterName).Set(totalPower)
	e.metrics.ClusterEnergyEfficiency.WithLabelValues(clusterName).Set(e.clusterEnergy.ClusterPowerEfficiency)
	e.metrics.ClusterCarbonEmissions.WithLabelValues(clusterName).Set(totalCarbon)
}

// cleanOldEnergyData removes old energy and carbon data
func (e *EnergyMonitor) cleanOldEnergyData() {
	e.mu.Lock()
	defer e.mu.Unlock()

	energyCutoff := time.Now().Add(-e.config.EnergyDataRetention)
	carbonCutoff := time.Now().Add(-e.config.CarbonDataRetention)

	for jobID, data := range e.energyData {
		if data.Timestamp.Before(energyCutoff) {
			delete(e.energyData, jobID)
		}
	}

	for jobID, data := range e.carbonData {
		if data.Timestamp.Before(carbonCutoff) {
			delete(e.carbonData, jobID)
		}
	}
}

// Helper functions for configuration data

// getRegionalCarbonFactors returns carbon intensity factors by region
func getRegionalCarbonFactors() map[string]float64 {
	return map[string]float64{
		"US":   1.0, // Baseline
		"EU":   0.8, // Lower carbon intensity
		"ASIA": 1.2, // Higher carbon intensity
		"AU":   0.9, // Moderate carbon intensity
	}
}

// getTimeOfDayFactors returns carbon intensity factors by hour
func getTimeOfDayFactors() map[int]float64 {
	factors := make(map[int]float64)
	for hour := 0; hour < 24; hour++ {
		if hour >= 10 && hour <= 16 {
			factors[hour] = 0.8 // Lower during peak solar hours
		} else if hour >= 18 && hour <= 22 {
			factors[hour] = 1.2 // Higher during peak demand
		} else {
			factors[hour] = 1.0 // Baseline
		}
	}
	return factors
}

// getEnergyBenchmarks returns energy efficiency benchmarks
func getEnergyBenchmarks() map[string]float64 {
	return map[string]float64{
		"cpu_efficiency":     0.75,
		"memory_efficiency":  0.80,
		"gpu_efficiency":     0.70,
		"overall_efficiency": 0.75,
	}
}

// getEnergyOptimizationRules returns energy optimization rules
func getEnergyOptimizationRules() []EnergyOptimizationRule {
	return []EnergyOptimizationRule{
		{
			RuleID:          "cpu_underutilization",
			Condition:       "cpu_efficiency < 0.5",
			Recommendation:  "Reduce CPU allocation or optimize CPU-bound operations",
			ExpectedSavings: 0.3,
			Priority:        "high",
		},
		{
			RuleID:          "memory_waste",
			Condition:       "memory_efficiency < 0.6",
			Recommendation:  "Optimize memory allocation and usage patterns",
			ExpectedSavings: 0.25,
			Priority:        "medium",
		},
		{
			RuleID:          "power_inefficiency",
			Condition:       "power_efficiency < 0.6",
			Recommendation:  "Review job scheduling and resource allocation",
			ExpectedSavings: 0.2,
			Priority:        "medium",
		},
	}
}

// GetEnergyData returns energy data for a specific job
func (e *EnergyMonitor) GetEnergyData(jobID string) (*JobEnergyData, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	data, exists := e.energyData[jobID]
	return data, exists
}

// GetCarbonData returns carbon footprint data for a specific job
func (e *EnergyMonitor) GetCarbonData(jobID string) (*CarbonFootprintData, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	data, exists := e.carbonData[jobID]
	return data, exists
}

// GetClusterEnergyData returns cluster-wide energy data
func (e *EnergyMonitor) GetClusterEnergyData() *ClusterEnergyData {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.clusterEnergy
}

// GetEnergyStats returns energy monitoring statistics
func (e *EnergyMonitor) GetEnergyStats() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"monitored_jobs":  len(e.energyData),
		"carbon_tracked":  len(e.carbonData),
		"last_collection": e.lastCollection,
		"cluster_energy":  e.clusterEnergy,
		"config":          e.config,
	}
}
