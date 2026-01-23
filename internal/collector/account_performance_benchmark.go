// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type AccountPerformanceBenchmarkSLURMClient interface {
	GetAccountPerformanceMetrics(ctx context.Context, accountName string) (*AccountPerformanceMetrics, error)
	GetAccountBenchmarkResults(ctx context.Context, accountName string) ([]*AccountBenchmarkResult, error)
	GetAccountPerformanceComparisons(ctx context.Context, accountName string) (*AccountPerformanceComparisons, error)
	GetAccountPerformanceTrends(ctx context.Context, accountName string) (*AccountPerformanceTrends, error)
	GetAccountEfficiencyAnalysis(ctx context.Context, accountName string) (*AccountEfficiencyAnalysis, error)
	GetAccountResourceUtilization(ctx context.Context, accountName string) (*AccountResourceUtilization, error)
	GetAccountWorkloadCharacteristics(ctx context.Context, accountName string) (*AccountWorkloadCharacteristics, error)
	GetAccountPerformanceBaselines(ctx context.Context, accountName string) ([]*AccountPerformanceBaseline, error)
	GetAccountSLACompliance(ctx context.Context, accountName string) (*AccountSLACompliance, error)
	GetAccountPerformanceOptimization(ctx context.Context, accountName string) ([]*AccountPerformanceOptimization, error)
	GetAccountCapacityPredictions(ctx context.Context, accountName string) (*AccountCapacityPredictions, error)
	GetAccountPerformanceAlerts(ctx context.Context, accountName string) ([]*AccountPerformanceAlert, error)
	GetAccountBenchmarkHistory(ctx context.Context, accountName string) ([]*AccountBenchmarkHistory, error)
	GetSystemPerformanceOverview(ctx context.Context) (*SystemPerformanceOverview, error)
}

type AccountPerformanceMetrics struct {
	AccountName                string
	TotalJobsCompleted         int64
	AverageJobDuration         time.Duration
	MedianJobDuration          time.Duration
	JobThroughput              float64
	JobSuccessRate             float64
	JobFailureRate             float64
	QueueWaitTime              time.Duration
	AverageQueueTime           time.Duration
	MedianQueueTime            time.Duration
	TurnaroundTime             time.Duration
	ResourceUtilizationCPU     float64
	ResourceUtilizationGPU     float64
	ResourceUtilizationMemory  float64
	ResourceUtilizationStorage float64
	NetworkUtilization         float64
	EnergyEfficiency           float64
	CostEfficiency             float64
	PerformanceScore           float64
	ReliabilityScore           float64
	AvailabilityScore          float64
	ScalabilityScore           float64
	LastUpdated                time.Time
	MeasurementPeriod          string
	PerformanceGrade           string
}

type AccountBenchmarkResult struct {
	BenchmarkID           string
	AccountName           string
	BenchmarkType         string
	BenchmarkName         string
	ExecutionTime         time.Time
	Duration              time.Duration
	Score                 float64
	ReferenceScore        float64
	PerformanceRatio      float64
	Percentile            float64
	Ranking               int
	TotalParticipants     int
	BenchmarkCategory     string
	WorkloadType          string
	ResourceProfile       string
	TestConditions        map[string]interface{}
	Metrics               map[string]float64
	ComparisonBaseline    string
	PerformanceClass      string
	AccuracyLevel         float64
	ConfidenceInterval    float64
	StandardDeviation     float64
	Variance              float64
	OptimizationPotential float64
	RecommendedActions    []string
}

type AccountPerformanceComparisons struct {
	AccountName              string
	PeerAccounts             []string
	PeerAveragePerformance   float64
	IndustryBenchmark        float64
	TopPerformerBenchmark    float64
	PerformanceRank          int
	TotalAccountsCompared    int
	PerformancePercentile    float64
	RelativePerformance      float64
	PerformanceGap           float64
	CompetitivePosition      string
	StrengthAreas            []string
	ImprovementAreas         []string
	BenchmarkMetrics         map[string]float64
	ComparisonPeriod         string
	ComparisonMethodology    string
	StatisticalSignificance  float64
	ConfidenceLevel          float64
	LastComparison           time.Time
	TrendDirection           string
	YearOverYearChange       float64
	QuarterOverQuarterChange float64
}

type AccountPerformanceTrends struct {
	AccountName           string
	TrendPeriod           string
	PerformanceTrend      string
	TrendStrength         float64
	TrendDirection        string
	LinearRegression      map[string]float64
	SeasonalPatterns      map[string]float64
	CyclicalPatterns      []string
	AnomalyCount          int
	TrendChangePoints     []time.Time
	ForecastAccuracy      float64
	PredictiveConfidence  float64
	HistoricalVariability float64
	PerformanceVolatility float64
	GrowthRate            float64
	DeclineRate           float64
	StabilityIndex        float64
	ConsistencyScore      float64
	ReliabilityTrend      string
	QualityTrend          string
	EfficiencyTrend       string
	ScalabilityTrend      string
	TrendFactors          []string
	ExternalInfluences    []string
}

type AccountEfficiencyAnalysis struct {
	AccountName               string
	OverallEfficiencyScore    float64
	CPUEfficiency             float64
	GPUEfficiency             float64
	MemoryEfficiency          float64
	StorageEfficiency         float64
	NetworkEfficiency         float64
	EnergyEfficiency          float64
	CostEfficiency            float64
	TimeEfficiency            float64
	ResourceWaste             float64
	IdleTime                  float64
	OverprovisioningRatio     float64
	UnderprovisioningRatio    float64
	OptimalResourceAllocation map[string]float64
	BottleneckAnalysis        map[string]float64
	PerformanceConstraints    []string
	EfficiencyFactors         map[string]float64
	WasteReductionPotential   float64
	OptimizationOpportunities []string
	EfficiencyTrends          map[string]float64
	BenchmarkComparison       map[string]float64
	TargetEfficiency          float64
	EfficiencyGap             float64
	ImprovementPotential      float64
}

type AccountResourceUtilization struct {
	AccountName            string
	CPUUtilization         map[string]float64
	GPUUtilization         map[string]float64
	MemoryUtilization      map[string]float64
	StorageUtilization     map[string]float64
	NetworkUtilization     map[string]float64
	PeakUtilization        map[string]float64
	AverageUtilization     map[string]float64
	MinimumUtilization     map[string]float64
	UtilizationVariability map[string]float64
	ResourceContention     map[string]float64
	QueueingDelays         map[string]time.Duration
	ThroughputMetrics      map[string]float64
	LatencyMetrics         map[string]time.Duration
	BandwidthUtilization   map[string]float64
	IOPSMetrics            map[string]float64
	ConcurrencyMetrics     map[string]float64
	ScalingMetrics         map[string]float64
	ResourceBalance        float64
	AllocationEfficiency   float64
	CapacityUtilization    float64
	ResourceDemandPatterns map[string]float64
	LoadDistribution       map[string]float64
	ResourceBottlenecks    []string
}

type AccountWorkloadCharacteristics struct {
	AccountName                   string
	WorkloadTypes                 []string
	JobSizeDistribution           map[string]float64
	JobDurationDistribution       map[string]float64
	ResourceDemandPatterns        map[string]float64
	TemporalPatterns              map[string]float64
	ConcurrencyPatterns           map[string]float64
	BatchJobCharacteristics       map[string]float64
	InteractiveJobCharacteristics map[string]float64
	LongRunningJobCharacteristics map[string]float64
	CPUIntensiveRatio             float64
	GPUIntensiveRatio             float64
	MemoryIntensiveRatio          float64
	IOIntensiveRatio              float64
	NetworkIntensiveRatio         float64
	MixedWorkloadRatio            float64
	WorkloadComplexity            float64
	WorkloadVariability           float64
	WorkloadPredictability        float64
	SeasonalityIndex              float64
	WorkloadEfficiency            float64
	ResourceDiversity             float64
	ApplicationProfiles           map[string]float64
	UserBehaviorPatterns          map[string]float64
	WorkloadTrends                map[string]float64
}

type AccountPerformanceBaseline struct {
	BaselineID           string
	AccountName          string
	BaselineType         string
	BaselineName         string
	EstablishedDate      time.Time
	BaselineMetrics      map[string]float64
	PerformanceTargets   map[string]float64
	AcceptableRanges     map[string]float64
	AlertThresholds      map[string]float64
	BaselinePeriod       string
	MeasurementFrequency string
	BaselineValidity     bool
	LastUpdated          time.Time
	UpdateFrequency      string
	BaselineSource       string
	StatisticalBasis     string
	ConfidenceLevel      float64
	SampleSize           int
	Methodology          string
	BaselineStatus       string
	DeviationTolerance   float64
	ReviewSchedule       string
	ApprovalStatus       string
	BaselineOwner        string
}

type AccountSLACompliance struct {
	AccountName            string
	SLATargets             map[string]float64
	CurrentPerformance     map[string]float64
	ComplianceStatus       map[string]string
	CompliancePercentage   map[string]float64
	ViolationCount         map[string]int
	ViolationSeverity      map[string]string
	TimeToResolution       map[string]time.Duration
	SLACredits             map[string]float64
	PerformanceMargin      map[string]float64
	RiskLevel              string
	ComplianceScore        float64
	AvailabilityCompliance float64
	PerformanceCompliance  float64
	ReliabilityCompliance  float64
	SLAPeriod              string
	MonitoringFrequency    string
	ReportingPeriod        string
	EscalationTriggers     map[string]interface{}
	RemediationActions     []string
	SLAReviewDate          time.Time
	ContractualObligations map[string]interface{}
	PenaltyExposure        float64
	ComplianceHistory      []string
	TrendAnalysis          map[string]string
}

type AccountPerformanceOptimization struct {
	OptimizationID           string
	AccountName              string
	OptimizationType         string
	OptimizationArea         string
	CurrentPerformance       float64
	TargetPerformance        float64
	PotentialImprovement     float64
	ImplementationComplexity string
	ImplementationCost       float64
	ExpectedROI              float64
	TimeToImplement          time.Duration
	RiskLevel                string
	Priority                 string
	Status                   string
	Recommendation           string
	DetailedAnalysis         string
	Prerequisites            []string
	Dependencies             []string
	ResourceRequirements     map[string]interface{}
	ExpectedBenefits         []string
	PotentialRisks           []string
	MitigationStrategies     []string
	SuccessMetrics           []string
	MonitoringPlan           string
	ReviewSchedule           string
	ApprovalRequired         bool
	BusinessImpact           string
}

type AccountCapacityPredictions struct {
	AccountName                 string
	PredictionPeriod            string
	PredictionConfidence        float64
	CPUCapacityForecast         map[string]float64
	GPUCapacityForecast         map[string]float64
	MemoryCapacityForecast      map[string]float64
	StorageCapacityForecast     map[string]float64
	NetworkCapacityForecast     map[string]float64
	WorkloadGrowthPrediction    float64
	UserGrowthPrediction        float64
	ResourceDemandGrowth        map[string]float64
	CapacityUtilizationForecast map[string]float64
	BottleneckPredictions       []string
	ScalingRequirements         map[string]interface{}
	CapacityShortfalls          map[string]float64
	CapacityExcess              map[string]float64
	OptimalCapacityPlan         map[string]interface{}
	InvestmentRequirements      map[string]float64
	CostProjections             map[string]float64
	RiskFactors                 []string
	ScenarioAnalysis            map[string]interface{}
	ContingencyPlanning         map[string]interface{}
	CapacityMilestones          []string
	ReviewPoints                []time.Time
	AccuracyTracking            map[string]float64
	ModelValidation             map[string]float64
}

type AccountPerformanceAlert struct {
	AlertID                    string
	AccountName                string
	AlertType                  string
	Severity                   string
	Metric                     string
	CurrentValue               float64
	ThresholdValue             float64
	DeviationPercentage        float64
	AlertTime                  time.Time
	Status                     string
	Description                string
	Impact                     string
	RootCause                  string
	RecommendedActions         []string
	EscalationLevel            int
	NotificationsSent          int
	AcknowledgedBy             string
	AcknowledgedTime           time.Time
	ResolvedTime               time.Time
	ResolutionNotes            string
	RelatedAlerts              []string
	TrendData                  map[string]float64
	HistoricalContext          string
	BusinessImpact             string
	TechnicalImpact            string
	SLAImpact                  string
	AutoRemediationApplied     bool
	ManualInterventionRequired bool
	AlertFrequency             string
	SuppressUntil              time.Time
}

type AccountBenchmarkHistory struct {
	HistoryID                string
	AccountName              string
	BenchmarkDate            time.Time
	BenchmarkType            string
	PerformanceScore         float64
	PerformanceGrade         string
	Ranking                  int
	PercentileRank           float64
	YearOverYearChange       float64
	QuarterOverQuarterChange float64
	MonthOverMonthChange     float64
	TrendDirection           string
	PerformanceFactors       map[string]float64
	ExternalFactors          []string
	SeasonalAdjustment       float64
	NormalizedScore          float64
	WeightedScore            float64
	ComponentScores          map[string]float64
	BenchmarkNotes           string
	DataQuality              string
	MeasurementAccuracy      float64
	ComparisonBaseline       string
	MethodologyVersion       string
	CertificationLevel       string
	AuditTrail               []string
	ValidationStatus         string
}

type SystemPerformanceOverview struct {
	OverviewDate              time.Time
	TotalAccounts             int
	AveragePerformanceScore   float64
	MedianPerformanceScore    float64
	PerformanceDistribution   map[string]int
	TopPerformingAccounts     []string
	UnderperformingAccounts   []string
	PerformanceTrends         map[string]float64
	SystemUtilization         map[string]float64
	SystemEfficiency          float64
	SystemReliability         float64
	SystemAvailability        float64
	SystemScalability         float64
	ResourceContention        map[string]float64
	CapacityUtilization       map[string]float64
	PerformanceBottlenecks    []string
	SystemAlerts              int
	SystemViolations          int
	ComplianceScore           float64
	OptimizationOpportunities int
	SystemHealthScore         float64
	OverallSystemGrade        string
	BenchmarkComparisons      map[string]float64
	IndustryRanking           int
	SystemMetrics             map[string]float64
	QualityMetrics            map[string]float64
	OperationalMetrics        map[string]float64
}

type AccountPerformanceBenchmarkCollector struct {
	client AccountPerformanceBenchmarkSLURMClient
	mutex  sync.RWMutex

	// Performance metrics
	totalJobsCompleted *prometheus.GaugeVec
	averageJobDuration *prometheus.GaugeVec
	medianJobDuration  *prometheus.GaugeVec
	jobThroughput      *prometheus.GaugeVec
	jobSuccessRate     *prometheus.GaugeVec
	jobFailureRate     *prometheus.GaugeVec
	queueWaitTime      *prometheus.GaugeVec
	averageQueueTime   *prometheus.GaugeVec
	medianQueueTime    *prometheus.GaugeVec
	turnaroundTime     *prometheus.GaugeVec
	performanceScore   *prometheus.GaugeVec
	reliabilityScore   *prometheus.GaugeVec
	availabilityScore  *prometheus.GaugeVec
	scalabilityScore   *prometheus.GaugeVec
	performanceGrade   *prometheus.GaugeVec

	// Resource utilization metrics
	resourceUtilizationCPU     *prometheus.GaugeVec
	resourceUtilizationGPU     *prometheus.GaugeVec
	resourceUtilizationMemory  *prometheus.GaugeVec
	resourceUtilizationStorage *prometheus.GaugeVec
	networkUtilization         *prometheus.GaugeVec
	energyEfficiency           *prometheus.GaugeVec
	costEfficiency             *prometheus.GaugeVec

	// Benchmark metrics
	benchmarkScore            *prometheus.GaugeVec
	benchmarkReferenceScore   *prometheus.GaugeVec
	benchmarkPerformanceRatio *prometheus.GaugeVec
	benchmarkPercentile       *prometheus.GaugeVec
	benchmarkRanking          *prometheus.GaugeVec
	benchmarkAccuracy         *prometheus.GaugeVec
	benchmarkConfidence       *prometheus.GaugeVec

	// Comparison metrics
	peerAveragePerformance   *prometheus.GaugeVec
	industryBenchmark        *prometheus.GaugeVec
	topPerformerBenchmark    *prometheus.GaugeVec
	performanceRank          *prometheus.GaugeVec
	performancePercentile    *prometheus.GaugeVec
	relativePerformance      *prometheus.GaugeVec
	performanceGap           *prometheus.GaugeVec
	yearOverYearChange       *prometheus.GaugeVec
	quarterOverQuarterChange *prometheus.GaugeVec

	// Trend metrics
	trendStrength         *prometheus.GaugeVec
	forecastAccuracy      *prometheus.GaugeVec
	predictiveConfidence  *prometheus.GaugeVec
	performanceVolatility *prometheus.GaugeVec
	stabilityIndex        *prometheus.GaugeVec
	consistencyScore      *prometheus.GaugeVec
	growthRate            *prometheus.GaugeVec

	// Efficiency metrics
	overallEfficiencyScore  *prometheus.GaugeVec
	cpuEfficiency           *prometheus.GaugeVec
	gpuEfficiency           *prometheus.GaugeVec
	memoryEfficiency        *prometheus.GaugeVec
	storageEfficiency       *prometheus.GaugeVec
	timeEfficiency          *prometheus.GaugeVec
	resourceWaste           *prometheus.GaugeVec
	idleTime                *prometheus.GaugeVec
	wasteReductionPotential *prometheus.GaugeVec
	improvementPotential    *prometheus.GaugeVec

	// Workload characteristics metrics
	workloadComplexity     *prometheus.GaugeVec
	workloadVariability    *prometheus.GaugeVec
	workloadPredictability *prometheus.GaugeVec
	cpuIntensiveRatio      *prometheus.GaugeVec
	gpuIntensiveRatio      *prometheus.GaugeVec
	memoryIntensiveRatio   *prometheus.GaugeVec
	ioIntensiveRatio       *prometheus.GaugeVec
	workloadEfficiency     *prometheus.GaugeVec
	resourceDiversity      *prometheus.GaugeVec

	// SLA compliance metrics
	slaComplianceScore     *prometheus.GaugeVec
	availabilityCompliance *prometheus.GaugeVec
	performanceCompliance  *prometheus.GaugeVec
	reliabilityCompliance  *prometheus.GaugeVec
	slaViolationsTotal     *prometheus.CounterVec
	slaCredits             *prometheus.GaugeVec
	penaltyExposure        *prometheus.GaugeVec

	// Optimization metrics
	optimizationsTotal       *prometheus.CounterVec
	potentialImprovement     *prometheus.GaugeVec
	expectedROI              *prometheus.GaugeVec
	implementationComplexity *prometheus.GaugeVec
	optimizationPriority     *prometheus.GaugeVec

	// Capacity prediction metrics
	capacityForecastCPU      *prometheus.GaugeVec
	capacityForecastGPU      *prometheus.GaugeVec
	capacityForecastMemory   *prometheus.GaugeVec
	capacityForecastStorage  *prometheus.GaugeVec
	workloadGrowthPrediction *prometheus.GaugeVec
	capacityShortfalls       *prometheus.GaugeVec
	capacityExcess           *prometheus.GaugeVec

	// Alert metrics
	performanceAlertsTotal *prometheus.CounterVec
	alertSeverity          *prometheus.GaugeVec
	alertResolutionTime    *prometheus.HistogramVec
	alertEscalationLevel   *prometheus.GaugeVec
	autoRemediationApplied *prometheus.CounterVec

	// System overview metrics
	systemUtilization     *prometheus.GaugeVec
	systemEfficiency      *prometheus.GaugeVec
	systemReliability     *prometheus.GaugeVec
	systemAvailability    *prometheus.GaugeVec
	systemHealthScore     *prometheus.GaugeVec
	systemAlertsTotal     *prometheus.CounterVec
	systemViolationsTotal *prometheus.CounterVec
}

func NewAccountPerformanceBenchmarkCollector(client AccountPerformanceBenchmarkSLURMClient) *AccountPerformanceBenchmarkCollector {
	return &AccountPerformanceBenchmarkCollector{
		client: client,

		// Performance metrics
		totalJobsCompleted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_total_jobs_completed",
				Help: "Total number of jobs completed by account",
			},
			[]string{"account"},
		),
		averageJobDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_average_job_duration_seconds",
				Help: "Average job duration for account",
			},
			[]string{"account"},
		),
		medianJobDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_median_job_duration_seconds",
				Help: "Median job duration for account",
			},
			[]string{"account"},
		),
		jobThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_job_throughput",
				Help: "Job throughput (jobs per hour) for account",
			},
			[]string{"account"},
		),
		jobSuccessRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_job_success_rate",
				Help: "Job success rate for account",
			},
			[]string{"account"},
		),
		jobFailureRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_job_failure_rate",
				Help: "Job failure rate for account",
			},
			[]string{"account"},
		),
		queueWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_queue_wait_time_seconds",
				Help: "Current queue wait time for account",
			},
			[]string{"account"},
		),
		averageQueueTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_average_queue_time_seconds",
				Help: "Average queue time for account",
			},
			[]string{"account"},
		),
		medianQueueTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_median_queue_time_seconds",
				Help: "Median queue time for account",
			},
			[]string{"account"},
		),
		turnaroundTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_turnaround_time_seconds",
				Help: "Turnaround time for account",
			},
			[]string{"account"},
		),
		performanceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_score",
				Help: "Overall performance score for account",
			},
			[]string{"account"},
		),
		reliabilityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_reliability_score",
				Help: "Reliability score for account",
			},
			[]string{"account"},
		),
		availabilityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_availability_score",
				Help: "Availability score for account",
			},
			[]string{"account"},
		),
		scalabilityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_scalability_score",
				Help: "Scalability score for account",
			},
			[]string{"account"},
		),
		performanceGrade: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_grade",
				Help: "Performance grade for account (A=4, B=3, C=2, D=1, F=0)",
			},
			[]string{"account", "grade"},
		),

		// Resource utilization metrics
		resourceUtilizationCPU: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cpu_utilization",
				Help: "CPU utilization for account",
			},
			[]string{"account"},
		),
		resourceUtilizationGPU: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_gpu_utilization",
				Help: "GPU utilization for account",
			},
			[]string{"account"},
		),
		resourceUtilizationMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_memory_utilization",
				Help: "Memory utilization for account",
			},
			[]string{"account"},
		),
		resourceUtilizationStorage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_storage_utilization",
				Help: "Storage utilization for account",
			},
			[]string{"account"},
		),
		networkUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_network_utilization",
				Help: "Network utilization for account",
			},
			[]string{"account"},
		),
		energyEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_energy_efficiency",
				Help: "Energy efficiency score for account",
			},
			[]string{"account"},
		),
		costEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_efficiency",
				Help: "Cost efficiency score for account",
			},
			[]string{"account"},
		),

		// Benchmark metrics
		benchmarkScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_score",
				Help: "Benchmark score for account",
			},
			[]string{"account", "benchmark_type", "benchmark_name"},
		),
		benchmarkReferenceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_reference_score",
				Help: "Benchmark reference score for account",
			},
			[]string{"account", "benchmark_type", "benchmark_name"},
		),
		benchmarkPerformanceRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_performance_ratio",
				Help: "Benchmark performance ratio for account",
			},
			[]string{"account", "benchmark_type"},
		),
		benchmarkPercentile: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_percentile",
				Help: "Benchmark percentile for account",
			},
			[]string{"account", "benchmark_type"},
		),
		benchmarkRanking: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_ranking",
				Help: "Benchmark ranking for account",
			},
			[]string{"account", "benchmark_type"},
		),
		benchmarkAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_accuracy",
				Help: "Benchmark accuracy level for account",
			},
			[]string{"account", "benchmark_type"},
		),
		benchmarkConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_confidence",
				Help: "Benchmark confidence interval for account",
			},
			[]string{"account", "benchmark_type"},
		),

		// Comparison metrics
		peerAveragePerformance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_peer_average_performance",
				Help: "Peer average performance for account",
			},
			[]string{"account"},
		),
		industryBenchmark: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_industry_benchmark",
				Help: "Industry benchmark performance for account",
			},
			[]string{"account"},
		),
		topPerformerBenchmark: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_top_performer_benchmark",
				Help: "Top performer benchmark for account",
			},
			[]string{"account"},
		),
		performanceRank: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_rank",
				Help: "Performance ranking for account",
			},
			[]string{"account"},
		),
		performancePercentile: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_percentile",
				Help: "Performance percentile for account",
			},
			[]string{"account"},
		),
		relativePerformance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_relative_performance",
				Help: "Relative performance compared to peers",
			},
			[]string{"account"},
		),
		performanceGap: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_gap",
				Help: "Performance gap compared to top performers",
			},
			[]string{"account"},
		),
		yearOverYearChange: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_year_over_year_change",
				Help: "Year-over-year performance change",
			},
			[]string{"account"},
		),
		quarterOverQuarterChange: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quarter_over_quarter_change",
				Help: "Quarter-over-quarter performance change",
			},
			[]string{"account"},
		),

		// Trend metrics
		trendStrength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_trend_strength",
				Help: "Performance trend strength for account",
			},
			[]string{"account"},
		),
		forecastAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_forecast_accuracy",
				Help: "Performance forecast accuracy for account",
			},
			[]string{"account"},
		),
		predictiveConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_predictive_confidence",
				Help: "Predictive confidence for account",
			},
			[]string{"account"},
		),
		performanceVolatility: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_volatility",
				Help: "Performance volatility for account",
			},
			[]string{"account"},
		),
		stabilityIndex: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_stability_index",
				Help: "Performance stability index for account",
			},
			[]string{"account"},
		),
		consistencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_consistency_score",
				Help: "Performance consistency score for account",
			},
			[]string{"account"},
		),
		growthRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_growth_rate",
				Help: "Performance growth rate for account",
			},
			[]string{"account"},
		),

		// Efficiency metrics
		overallEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_overall_efficiency_score",
				Help: "Overall efficiency score for account",
			},
			[]string{"account"},
		),
		cpuEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cpu_efficiency",
				Help: "CPU efficiency for account",
			},
			[]string{"account"},
		),
		gpuEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_gpu_efficiency",
				Help: "GPU efficiency for account",
			},
			[]string{"account"},
		),
		memoryEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_memory_efficiency",
				Help: "Memory efficiency for account",
			},
			[]string{"account"},
		),
		storageEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_storage_efficiency",
				Help: "Storage efficiency for account",
			},
			[]string{"account"},
		),
		timeEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_time_efficiency",
				Help: "Time efficiency for account",
			},
			[]string{"account"},
		),
		resourceWaste: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_waste",
				Help: "Resource waste percentage for account",
			},
			[]string{"account"},
		),
		idleTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_idle_time",
				Help: "Idle time percentage for account",
			},
			[]string{"account"},
		),
		wasteReductionPotential: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_waste_reduction_potential",
				Help: "Waste reduction potential for account",
			},
			[]string{"account"},
		),
		improvementPotential: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_improvement_potential",
				Help: "Overall improvement potential for account",
			},
			[]string{"account"},
		),

		// Workload characteristics metrics
		workloadComplexity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_workload_complexity",
				Help: "Workload complexity score for account",
			},
			[]string{"account"},
		),
		workloadVariability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_workload_variability",
				Help: "Workload variability for account",
			},
			[]string{"account"},
		),
		workloadPredictability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_workload_predictability",
				Help: "Workload predictability for account",
			},
			[]string{"account"},
		),
		cpuIntensiveRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cpu_intensive_ratio",
				Help: "CPU-intensive workload ratio for account",
			},
			[]string{"account"},
		),
		gpuIntensiveRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_gpu_intensive_ratio",
				Help: "GPU-intensive workload ratio for account",
			},
			[]string{"account"},
		),
		memoryIntensiveRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_memory_intensive_ratio",
				Help: "Memory-intensive workload ratio for account",
			},
			[]string{"account"},
		),
		ioIntensiveRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_io_intensive_ratio",
				Help: "I/O-intensive workload ratio for account",
			},
			[]string{"account"},
		),
		workloadEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_workload_efficiency",
				Help: "Workload efficiency score for account",
			},
			[]string{"account"},
		),
		resourceDiversity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_diversity",
				Help: "Resource diversity index for account",
			},
			[]string{"account"},
		),

		// SLA compliance metrics
		slaComplianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_sla_compliance_score",
				Help: "SLA compliance score for account",
			},
			[]string{"account"},
		),
		availabilityCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_availability_compliance",
				Help: "Availability compliance percentage for account",
			},
			[]string{"account"},
		),
		performanceCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_compliance",
				Help: "Performance compliance percentage for account",
			},
			[]string{"account"},
		),
		reliabilityCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_reliability_compliance",
				Help: "Reliability compliance percentage for account",
			},
			[]string{"account"},
		),
		slaViolationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_sla_violations_total",
				Help: "Total SLA violations for account",
			},
			[]string{"account", "violation_type"},
		),
		slaCredits: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_sla_credits",
				Help: "SLA credits earned for account",
			},
			[]string{"account"},
		),
		penaltyExposure: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_penalty_exposure",
				Help: "Penalty exposure amount for account",
			},
			[]string{"account"},
		),

		// Optimization metrics
		optimizationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_optimizations_total",
				Help: "Total optimization opportunities for account",
			},
			[]string{"account", "optimization_type", "status"},
		),
		potentialImprovement: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_potential_improvement",
				Help: "Potential improvement percentage for account",
			},
			[]string{"account", "optimization_area"},
		),
		expectedROI: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_expected_roi",
				Help: "Expected ROI for optimization",
			},
			[]string{"account", "optimization_type"},
		),
		implementationComplexity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_implementation_complexity",
				Help: "Implementation complexity score (1=low, 2=medium, 3=high)",
			},
			[]string{"account", "optimization_type", "complexity"},
		),
		optimizationPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_optimization_priority",
				Help: "Optimization priority score (1=low, 2=medium, 3=high)",
			},
			[]string{"account", "optimization_type", "priority"},
		),

		// Capacity prediction metrics
		capacityForecastCPU: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_capacity_forecast_cpu",
				Help: "CPU capacity forecast for account",
			},
			[]string{"account", "time_horizon"},
		),
		capacityForecastGPU: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_capacity_forecast_gpu",
				Help: "GPU capacity forecast for account",
			},
			[]string{"account", "time_horizon"},
		),
		capacityForecastMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_capacity_forecast_memory",
				Help: "Memory capacity forecast for account",
			},
			[]string{"account", "time_horizon"},
		),
		capacityForecastStorage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_capacity_forecast_storage",
				Help: "Storage capacity forecast for account",
			},
			[]string{"account", "time_horizon"},
		),
		workloadGrowthPrediction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_workload_growth_prediction",
				Help: "Workload growth prediction for account",
			},
			[]string{"account"},
		),
		capacityShortfalls: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_capacity_shortfalls",
				Help: "Predicted capacity shortfalls for account",
			},
			[]string{"account", "resource_type"},
		),
		capacityExcess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_capacity_excess",
				Help: "Predicted capacity excess for account",
			},
			[]string{"account", "resource_type"},
		),

		// Alert metrics
		performanceAlertsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_performance_alerts_total",
				Help: "Total performance alerts for account",
			},
			[]string{"account", "alert_type", "severity"},
		),
		alertSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_alert_severity_score",
				Help: "Alert severity score for account",
			},
			[]string{"account", "severity"},
		),
		alertResolutionTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_account_alert_resolution_time_seconds",
				Help:    "Alert resolution time for account",
				Buckets: []float64{300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"account", "alert_type"},
		),
		alertEscalationLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_alert_escalation_level",
				Help: "Alert escalation level for account",
			},
			[]string{"account"},
		),
		autoRemediationApplied: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_auto_remediation_applied_total",
				Help: "Total auto-remediation actions applied for account",
			},
			[]string{"account", "remediation_type"},
		),

		// System overview metrics
		systemUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_utilization",
				Help: "System utilization by resource type",
			},
			[]string{"resource_type"},
		),
		systemEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_efficiency",
				Help: "Overall system efficiency score",
			},
			[]string{},
		),
		systemReliability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_reliability",
				Help: "Overall system reliability score",
			},
			[]string{},
		),
		systemAvailability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_availability",
				Help: "Overall system availability score",
			},
			[]string{},
		),
		systemHealthScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_health_score",
				Help: "Overall system health score",
			},
			[]string{},
		),
		systemAlertsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_system_alerts_total",
				Help: "Total system alerts",
			},
			[]string{"alert_type"},
		),
		systemViolationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_system_violations_total",
				Help: "Total system violations",
			},
			[]string{"violation_type"},
		),
	}
}

func (c *AccountPerformanceBenchmarkCollector) Describe(ch chan<- *prometheus.Desc) {
	c.totalJobsCompleted.Describe(ch)
	c.averageJobDuration.Describe(ch)
	c.medianJobDuration.Describe(ch)
	c.jobThroughput.Describe(ch)
	c.jobSuccessRate.Describe(ch)
	c.jobFailureRate.Describe(ch)
	c.queueWaitTime.Describe(ch)
	c.averageQueueTime.Describe(ch)
	c.medianQueueTime.Describe(ch)
	c.turnaroundTime.Describe(ch)
	c.performanceScore.Describe(ch)
	c.reliabilityScore.Describe(ch)
	c.availabilityScore.Describe(ch)
	c.scalabilityScore.Describe(ch)
	c.performanceGrade.Describe(ch)
	c.resourceUtilizationCPU.Describe(ch)
	c.resourceUtilizationGPU.Describe(ch)
	c.resourceUtilizationMemory.Describe(ch)
	c.resourceUtilizationStorage.Describe(ch)
	c.networkUtilization.Describe(ch)
	c.energyEfficiency.Describe(ch)
	c.costEfficiency.Describe(ch)
	c.benchmarkScore.Describe(ch)
	c.benchmarkReferenceScore.Describe(ch)
	c.benchmarkPerformanceRatio.Describe(ch)
	c.benchmarkPercentile.Describe(ch)
	c.benchmarkRanking.Describe(ch)
	c.benchmarkAccuracy.Describe(ch)
	c.benchmarkConfidence.Describe(ch)
	c.peerAveragePerformance.Describe(ch)
	c.industryBenchmark.Describe(ch)
	c.topPerformerBenchmark.Describe(ch)
	c.performanceRank.Describe(ch)
	c.performancePercentile.Describe(ch)
	c.relativePerformance.Describe(ch)
	c.performanceGap.Describe(ch)
	c.yearOverYearChange.Describe(ch)
	c.quarterOverQuarterChange.Describe(ch)
	c.trendStrength.Describe(ch)
	c.forecastAccuracy.Describe(ch)
	c.predictiveConfidence.Describe(ch)
	c.performanceVolatility.Describe(ch)
	c.stabilityIndex.Describe(ch)
	c.consistencyScore.Describe(ch)
	c.growthRate.Describe(ch)
	c.overallEfficiencyScore.Describe(ch)
	c.cpuEfficiency.Describe(ch)
	c.gpuEfficiency.Describe(ch)
	c.memoryEfficiency.Describe(ch)
	c.storageEfficiency.Describe(ch)
	c.timeEfficiency.Describe(ch)
	c.resourceWaste.Describe(ch)
	c.idleTime.Describe(ch)
	c.wasteReductionPotential.Describe(ch)
	c.improvementPotential.Describe(ch)
	c.workloadComplexity.Describe(ch)
	c.workloadVariability.Describe(ch)
	c.workloadPredictability.Describe(ch)
	c.cpuIntensiveRatio.Describe(ch)
	c.gpuIntensiveRatio.Describe(ch)
	c.memoryIntensiveRatio.Describe(ch)
	c.ioIntensiveRatio.Describe(ch)
	c.workloadEfficiency.Describe(ch)
	c.resourceDiversity.Describe(ch)
	c.slaComplianceScore.Describe(ch)
	c.availabilityCompliance.Describe(ch)
	c.performanceCompliance.Describe(ch)
	c.reliabilityCompliance.Describe(ch)
	c.slaViolationsTotal.Describe(ch)
	c.slaCredits.Describe(ch)
	c.penaltyExposure.Describe(ch)
	c.optimizationsTotal.Describe(ch)
	c.potentialImprovement.Describe(ch)
	c.expectedROI.Describe(ch)
	c.implementationComplexity.Describe(ch)
	c.optimizationPriority.Describe(ch)
	c.capacityForecastCPU.Describe(ch)
	c.capacityForecastGPU.Describe(ch)
	c.capacityForecastMemory.Describe(ch)
	c.capacityForecastStorage.Describe(ch)
	c.workloadGrowthPrediction.Describe(ch)
	c.capacityShortfalls.Describe(ch)
	c.capacityExcess.Describe(ch)
	c.performanceAlertsTotal.Describe(ch)
	c.alertSeverity.Describe(ch)
	c.alertResolutionTime.Describe(ch)
	c.alertEscalationLevel.Describe(ch)
	c.autoRemediationApplied.Describe(ch)
	c.systemUtilization.Describe(ch)
	c.systemEfficiency.Describe(ch)
	c.systemReliability.Describe(ch)
	c.systemAvailability.Describe(ch)
	c.systemHealthScore.Describe(ch)
	c.systemAlertsTotal.Describe(ch)
	c.systemViolationsTotal.Describe(ch)
}

func (c *AccountPerformanceBenchmarkCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx := context.Background()
	accounts := []string{"account1", "account2", "account3"}

	for _, account := range accounts {
		c.collectPerformanceMetrics(ctx, account, ch)
		c.collectBenchmarkResults(ctx, account, ch)
		c.collectPerformanceComparisons(ctx, account, ch)
		c.collectPerformanceTrends(ctx, account, ch)
		c.collectEfficiencyAnalysis(ctx, account, ch)
		c.collectResourceUtilization(ctx, account, ch)
		c.collectWorkloadCharacteristics(ctx, account, ch)
		c.collectSLACompliance(ctx, account, ch)
		c.collectPerformanceOptimization(ctx, account, ch)
		c.collectCapacityPredictions(ctx, account, ch)
		c.collectPerformanceAlerts(ctx, account, ch)
	}

	c.collectSystemOverview(ctx, ch)

	c.totalJobsCompleted.Collect(ch)
	c.averageJobDuration.Collect(ch)
	c.medianJobDuration.Collect(ch)
	c.jobThroughput.Collect(ch)
	c.jobSuccessRate.Collect(ch)
	c.jobFailureRate.Collect(ch)
	c.queueWaitTime.Collect(ch)
	c.averageQueueTime.Collect(ch)
	c.medianQueueTime.Collect(ch)
	c.turnaroundTime.Collect(ch)
	c.performanceScore.Collect(ch)
	c.reliabilityScore.Collect(ch)
	c.availabilityScore.Collect(ch)
	c.scalabilityScore.Collect(ch)
	c.performanceGrade.Collect(ch)
	c.resourceUtilizationCPU.Collect(ch)
	c.resourceUtilizationGPU.Collect(ch)
	c.resourceUtilizationMemory.Collect(ch)
	c.resourceUtilizationStorage.Collect(ch)
	c.networkUtilization.Collect(ch)
	c.energyEfficiency.Collect(ch)
	c.costEfficiency.Collect(ch)
	c.benchmarkScore.Collect(ch)
	c.benchmarkReferenceScore.Collect(ch)
	c.benchmarkPerformanceRatio.Collect(ch)
	c.benchmarkPercentile.Collect(ch)
	c.benchmarkRanking.Collect(ch)
	c.benchmarkAccuracy.Collect(ch)
	c.benchmarkConfidence.Collect(ch)
	c.peerAveragePerformance.Collect(ch)
	c.industryBenchmark.Collect(ch)
	c.topPerformerBenchmark.Collect(ch)
	c.performanceRank.Collect(ch)
	c.performancePercentile.Collect(ch)
	c.relativePerformance.Collect(ch)
	c.performanceGap.Collect(ch)
	c.yearOverYearChange.Collect(ch)
	c.quarterOverQuarterChange.Collect(ch)
	c.trendStrength.Collect(ch)
	c.forecastAccuracy.Collect(ch)
	c.predictiveConfidence.Collect(ch)
	c.performanceVolatility.Collect(ch)
	c.stabilityIndex.Collect(ch)
	c.consistencyScore.Collect(ch)
	c.growthRate.Collect(ch)
	c.overallEfficiencyScore.Collect(ch)
	c.cpuEfficiency.Collect(ch)
	c.gpuEfficiency.Collect(ch)
	c.memoryEfficiency.Collect(ch)
	c.storageEfficiency.Collect(ch)
	c.timeEfficiency.Collect(ch)
	c.resourceWaste.Collect(ch)
	c.idleTime.Collect(ch)
	c.wasteReductionPotential.Collect(ch)
	c.improvementPotential.Collect(ch)
	c.workloadComplexity.Collect(ch)
	c.workloadVariability.Collect(ch)
	c.workloadPredictability.Collect(ch)
	c.cpuIntensiveRatio.Collect(ch)
	c.gpuIntensiveRatio.Collect(ch)
	c.memoryIntensiveRatio.Collect(ch)
	c.ioIntensiveRatio.Collect(ch)
	c.workloadEfficiency.Collect(ch)
	c.resourceDiversity.Collect(ch)
	c.slaComplianceScore.Collect(ch)
	c.availabilityCompliance.Collect(ch)
	c.performanceCompliance.Collect(ch)
	c.reliabilityCompliance.Collect(ch)
	c.slaViolationsTotal.Collect(ch)
	c.slaCredits.Collect(ch)
	c.penaltyExposure.Collect(ch)
	c.optimizationsTotal.Collect(ch)
	c.potentialImprovement.Collect(ch)
	c.expectedROI.Collect(ch)
	c.implementationComplexity.Collect(ch)
	c.optimizationPriority.Collect(ch)
	c.capacityForecastCPU.Collect(ch)
	c.capacityForecastGPU.Collect(ch)
	c.capacityForecastMemory.Collect(ch)
	c.capacityForecastStorage.Collect(ch)
	c.workloadGrowthPrediction.Collect(ch)
	c.capacityShortfalls.Collect(ch)
	c.capacityExcess.Collect(ch)
	c.performanceAlertsTotal.Collect(ch)
	c.alertSeverity.Collect(ch)
	c.alertResolutionTime.Collect(ch)
	c.alertEscalationLevel.Collect(ch)
	c.autoRemediationApplied.Collect(ch)
	c.systemUtilization.Collect(ch)
	c.systemEfficiency.Collect(ch)
	c.systemReliability.Collect(ch)
	c.systemAvailability.Collect(ch)
	c.systemHealthScore.Collect(ch)
	c.systemAlertsTotal.Collect(ch)
	c.systemViolationsTotal.Collect(ch)
}

func (c *AccountPerformanceBenchmarkCollector) collectPerformanceMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetAccountPerformanceMetrics(ctx, account)
	if err != nil {
		log.Printf("Error collecting performance metrics for account %s: %v", account, err)
		return
	}

	c.totalJobsCompleted.WithLabelValues(account).Set(float64(metrics.TotalJobsCompleted))
	c.averageJobDuration.WithLabelValues(account).Set(metrics.AverageJobDuration.Seconds())
	c.medianJobDuration.WithLabelValues(account).Set(metrics.MedianJobDuration.Seconds())
	c.jobThroughput.WithLabelValues(account).Set(metrics.JobThroughput)
	c.jobSuccessRate.WithLabelValues(account).Set(metrics.JobSuccessRate)
	c.jobFailureRate.WithLabelValues(account).Set(metrics.JobFailureRate)
	c.queueWaitTime.WithLabelValues(account).Set(metrics.QueueWaitTime.Seconds())
	c.averageQueueTime.WithLabelValues(account).Set(metrics.AverageQueueTime.Seconds())
	c.medianQueueTime.WithLabelValues(account).Set(metrics.MedianQueueTime.Seconds())
	c.turnaroundTime.WithLabelValues(account).Set(metrics.TurnaroundTime.Seconds())
	c.performanceScore.WithLabelValues(account).Set(metrics.PerformanceScore)
	c.reliabilityScore.WithLabelValues(account).Set(metrics.ReliabilityScore)
	c.availabilityScore.WithLabelValues(account).Set(metrics.AvailabilityScore)
	c.scalabilityScore.WithLabelValues(account).Set(metrics.ScalabilityScore)

	var gradeValue float64
	switch metrics.PerformanceGrade {
	case "A":
		gradeValue = 4
	case "B":
		gradeValue = 3
	case "C":
		gradeValue = 2
	case "D":
		gradeValue = 1
	case "F":
		gradeValue = 0
	}
	c.performanceGrade.WithLabelValues(account, metrics.PerformanceGrade).Set(gradeValue)

	c.resourceUtilizationCPU.WithLabelValues(account).Set(metrics.ResourceUtilizationCPU)
	c.resourceUtilizationGPU.WithLabelValues(account).Set(metrics.ResourceUtilizationGPU)
	c.resourceUtilizationMemory.WithLabelValues(account).Set(metrics.ResourceUtilizationMemory)
	c.resourceUtilizationStorage.WithLabelValues(account).Set(metrics.ResourceUtilizationStorage)
	c.networkUtilization.WithLabelValues(account).Set(metrics.NetworkUtilization)
	c.energyEfficiency.WithLabelValues(account).Set(metrics.EnergyEfficiency)
	c.costEfficiency.WithLabelValues(account).Set(metrics.CostEfficiency)
}

func (c *AccountPerformanceBenchmarkCollector) collectBenchmarkResults(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	results, err := c.client.GetAccountBenchmarkResults(ctx, account)
	if err != nil {
		log.Printf("Error collecting benchmark results for account %s: %v", account, err)
		return
	}

	for _, result := range results {
		c.benchmarkScore.WithLabelValues(account, result.BenchmarkType, result.BenchmarkName).Set(result.Score)
		c.benchmarkReferenceScore.WithLabelValues(account, result.BenchmarkType, result.BenchmarkName).Set(result.ReferenceScore)
		c.benchmarkPerformanceRatio.WithLabelValues(account, result.BenchmarkType).Set(result.PerformanceRatio)
		c.benchmarkPercentile.WithLabelValues(account, result.BenchmarkType).Set(result.Percentile)
		c.benchmarkRanking.WithLabelValues(account, result.BenchmarkType).Set(float64(result.Ranking))
		c.benchmarkAccuracy.WithLabelValues(account, result.BenchmarkType).Set(result.AccuracyLevel)
		c.benchmarkConfidence.WithLabelValues(account, result.BenchmarkType).Set(result.ConfidenceInterval)
	}
}

func (c *AccountPerformanceBenchmarkCollector) collectPerformanceComparisons(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	comparisons, err := c.client.GetAccountPerformanceComparisons(ctx, account)
	if err != nil {
		log.Printf("Error collecting performance comparisons for account %s: %v", account, err)
		return
	}

	c.peerAveragePerformance.WithLabelValues(account).Set(comparisons.PeerAveragePerformance)
	c.industryBenchmark.WithLabelValues(account).Set(comparisons.IndustryBenchmark)
	c.topPerformerBenchmark.WithLabelValues(account).Set(comparisons.TopPerformerBenchmark)
	c.performanceRank.WithLabelValues(account).Set(float64(comparisons.PerformanceRank))
	c.performancePercentile.WithLabelValues(account).Set(comparisons.PerformancePercentile)
	c.relativePerformance.WithLabelValues(account).Set(comparisons.RelativePerformance)
	c.performanceGap.WithLabelValues(account).Set(comparisons.PerformanceGap)
	c.yearOverYearChange.WithLabelValues(account).Set(comparisons.YearOverYearChange)
	c.quarterOverQuarterChange.WithLabelValues(account).Set(comparisons.QuarterOverQuarterChange)
}

func (c *AccountPerformanceBenchmarkCollector) collectPerformanceTrends(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	trends, err := c.client.GetAccountPerformanceTrends(ctx, account)
	if err != nil {
		log.Printf("Error collecting performance trends for account %s: %v", account, err)
		return
	}

	c.trendStrength.WithLabelValues(account).Set(trends.TrendStrength)
	c.forecastAccuracy.WithLabelValues(account).Set(trends.ForecastAccuracy)
	c.predictiveConfidence.WithLabelValues(account).Set(trends.PredictiveConfidence)
	c.performanceVolatility.WithLabelValues(account).Set(trends.PerformanceVolatility)
	c.stabilityIndex.WithLabelValues(account).Set(trends.StabilityIndex)
	c.consistencyScore.WithLabelValues(account).Set(trends.ConsistencyScore)
	c.growthRate.WithLabelValues(account).Set(trends.GrowthRate)
}

func (c *AccountPerformanceBenchmarkCollector) collectEfficiencyAnalysis(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	efficiency, err := c.client.GetAccountEfficiencyAnalysis(ctx, account)
	if err != nil {
		log.Printf("Error collecting efficiency analysis for account %s: %v", account, err)
		return
	}

	c.overallEfficiencyScore.WithLabelValues(account).Set(efficiency.OverallEfficiencyScore)
	c.cpuEfficiency.WithLabelValues(account).Set(efficiency.CPUEfficiency)
	c.gpuEfficiency.WithLabelValues(account).Set(efficiency.GPUEfficiency)
	c.memoryEfficiency.WithLabelValues(account).Set(efficiency.MemoryEfficiency)
	c.storageEfficiency.WithLabelValues(account).Set(efficiency.StorageEfficiency)
	c.timeEfficiency.WithLabelValues(account).Set(efficiency.TimeEfficiency)
	c.resourceWaste.WithLabelValues(account).Set(efficiency.ResourceWaste)
	c.idleTime.WithLabelValues(account).Set(efficiency.IdleTime)
	c.wasteReductionPotential.WithLabelValues(account).Set(efficiency.WasteReductionPotential)
	c.improvementPotential.WithLabelValues(account).Set(efficiency.ImprovementPotential)
}

func (c *AccountPerformanceBenchmarkCollector) collectResourceUtilization(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	_, err := c.client.GetAccountResourceUtilization(ctx, account)
	if err != nil {
		log.Printf("Error collecting resource utilization for account %s: %v", account, err)
		return
	}
}

func (c *AccountPerformanceBenchmarkCollector) collectWorkloadCharacteristics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	workload, err := c.client.GetAccountWorkloadCharacteristics(ctx, account)
	if err != nil {
		log.Printf("Error collecting workload characteristics for account %s: %v", account, err)
		return
	}

	c.workloadComplexity.WithLabelValues(account).Set(workload.WorkloadComplexity)
	c.workloadVariability.WithLabelValues(account).Set(workload.WorkloadVariability)
	c.workloadPredictability.WithLabelValues(account).Set(workload.WorkloadPredictability)
	c.cpuIntensiveRatio.WithLabelValues(account).Set(workload.CPUIntensiveRatio)
	c.gpuIntensiveRatio.WithLabelValues(account).Set(workload.GPUIntensiveRatio)
	c.memoryIntensiveRatio.WithLabelValues(account).Set(workload.MemoryIntensiveRatio)
	c.ioIntensiveRatio.WithLabelValues(account).Set(workload.IOIntensiveRatio)
	c.workloadEfficiency.WithLabelValues(account).Set(workload.WorkloadEfficiency)
	c.resourceDiversity.WithLabelValues(account).Set(workload.ResourceDiversity)
}

func (c *AccountPerformanceBenchmarkCollector) collectSLACompliance(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	sla, err := c.client.GetAccountSLACompliance(ctx, account)
	if err != nil {
		log.Printf("Error collecting SLA compliance for account %s: %v", account, err)
		return
	}

	c.slaComplianceScore.WithLabelValues(account).Set(sla.ComplianceScore)
	c.availabilityCompliance.WithLabelValues(account).Set(sla.AvailabilityCompliance)
	c.performanceCompliance.WithLabelValues(account).Set(sla.PerformanceCompliance)
	c.reliabilityCompliance.WithLabelValues(account).Set(sla.ReliabilityCompliance)
	c.penaltyExposure.WithLabelValues(account).Set(sla.PenaltyExposure)

	for violationType, count := range sla.ViolationCount {
		c.slaViolationsTotal.WithLabelValues(account, violationType).Add(float64(count))
	}

	for _, credits := range sla.SLACredits {
		c.slaCredits.WithLabelValues(account).Set(credits)
		break // Use first credits value as overall metric
	}
}

func (c *AccountPerformanceBenchmarkCollector) collectPerformanceOptimization(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	optimizations, err := c.client.GetAccountPerformanceOptimization(ctx, account)
	if err != nil {
		log.Printf("Error collecting performance optimization for account %s: %v", account, err)
		return
	}

	for _, opt := range optimizations {
		c.optimizationsTotal.WithLabelValues(account, opt.OptimizationType, opt.Status).Inc()
		c.potentialImprovement.WithLabelValues(account, opt.OptimizationArea).Set(opt.PotentialImprovement)
		c.expectedROI.WithLabelValues(account, opt.OptimizationType).Set(opt.ExpectedROI)

		var complexityScore float64
		switch opt.ImplementationComplexity {
		case "low":
			complexityScore = 1
		case "medium":
			complexityScore = 2
		case "high":
			complexityScore = 3
		}
		c.implementationComplexity.WithLabelValues(account, opt.OptimizationType, opt.ImplementationComplexity).Set(complexityScore)

		var priorityScore float64
		switch opt.Priority {
		case "low":
			priorityScore = 1
		case "medium":
			priorityScore = 2
		case "high":
			priorityScore = 3
		}
		c.optimizationPriority.WithLabelValues(account, opt.OptimizationType, opt.Priority).Set(priorityScore)
	}
}

func (c *AccountPerformanceBenchmarkCollector) collectCapacityPredictions(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	predictions, err := c.client.GetAccountCapacityPredictions(ctx, account)
	if err != nil {
		log.Printf("Error collecting capacity predictions for account %s: %v", account, err)
		return
	}

	c.workloadGrowthPrediction.WithLabelValues(account).Set(predictions.WorkloadGrowthPrediction)

	for timeHorizon, cpuForecast := range predictions.CPUCapacityForecast {
		c.capacityForecastCPU.WithLabelValues(account, timeHorizon).Set(cpuForecast)
	}

	for timeHorizon, gpuForecast := range predictions.GPUCapacityForecast {
		c.capacityForecastGPU.WithLabelValues(account, timeHorizon).Set(gpuForecast)
	}

	for timeHorizon, memoryForecast := range predictions.MemoryCapacityForecast {
		c.capacityForecastMemory.WithLabelValues(account, timeHorizon).Set(memoryForecast)
	}

	for timeHorizon, storageForecast := range predictions.StorageCapacityForecast {
		c.capacityForecastStorage.WithLabelValues(account, timeHorizon).Set(storageForecast)
	}

	for resourceType, shortfall := range predictions.CapacityShortfalls {
		c.capacityShortfalls.WithLabelValues(account, resourceType).Set(shortfall)
	}

	for resourceType, excess := range predictions.CapacityExcess {
		c.capacityExcess.WithLabelValues(account, resourceType).Set(excess)
	}
}

func (c *AccountPerformanceBenchmarkCollector) collectPerformanceAlerts(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	alerts, err := c.client.GetAccountPerformanceAlerts(ctx, account)
	if err != nil {
		log.Printf("Error collecting performance alerts for account %s: %v", account, err)
		return
	}

	alertCounts := make(map[string]map[string]int)
	severityCounts := make(map[string]float64)
	var totalEscalationLevel float64
	var alertCount int

	for _, alert := range alerts {
		if alertCounts[alert.AlertType] == nil {
			alertCounts[alert.AlertType] = make(map[string]int)
		}
		alertCounts[alert.AlertType][alert.Severity]++

		var severityScore float64
		switch alert.Severity {
		case "low":
			severityScore = 0.25
		case "medium":
			severityScore = 0.5
		case "high":
			severityScore = 0.75
		case "critical":
			severityScore = 1.0
		}
		severityCounts[alert.Severity] += severityScore

		totalEscalationLevel += float64(alert.EscalationLevel)

		if !alert.ResolvedTime.IsZero() {
			resolutionDuration := alert.ResolvedTime.Sub(alert.AlertTime)
			c.alertResolutionTime.WithLabelValues(account, alert.AlertType).Observe(resolutionDuration.Seconds())
		}

		if alert.AutoRemediationApplied {
			c.autoRemediationApplied.WithLabelValues(account, "performance").Inc()
		}

		alertCount++
	}

	for alertType, severityMap := range alertCounts {
		for severity, count := range severityMap {
			c.performanceAlertsTotal.WithLabelValues(account, alertType, severity).Add(float64(count))
		}
	}

	for severity, score := range severityCounts {
		c.alertSeverity.WithLabelValues(account, severity).Set(score)
	}

	if alertCount > 0 {
		c.alertEscalationLevel.WithLabelValues(account).Set(totalEscalationLevel / float64(alertCount))
	}
}

func (c *AccountPerformanceBenchmarkCollector) collectSystemOverview(ctx context.Context, ch chan<- prometheus.Metric) {
	overview, err := c.client.GetSystemPerformanceOverview(ctx)
	if err != nil {
		log.Printf("Error collecting system performance overview: %v", err)
		return
	}

	c.systemEfficiency.WithLabelValues().Set(overview.SystemEfficiency)
	c.systemReliability.WithLabelValues().Set(overview.SystemReliability)
	c.systemAvailability.WithLabelValues().Set(overview.SystemAvailability)
	c.systemHealthScore.WithLabelValues().Set(overview.SystemHealthScore)

	for resourceType, utilization := range overview.SystemUtilization {
		c.systemUtilization.WithLabelValues(resourceType).Set(utilization)
	}

	c.systemAlertsTotal.WithLabelValues("performance").Add(float64(overview.SystemAlerts))
	c.systemViolationsTotal.WithLabelValues("performance").Add(float64(overview.SystemViolations))
}
