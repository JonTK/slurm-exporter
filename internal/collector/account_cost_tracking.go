package collector

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type AccountCostTrackingSLURMClient interface {
	GetAccountCostMetrics(ctx context.Context, account string) (*AccountCostMetrics, error)
	GetAccountBudgetInfo(ctx context.Context, account string) (*AccountBudgetInfo, error)
	GetAccountCostHistory(ctx context.Context, account string, startTime, endTime time.Time) ([]*AccountCostHistoryEntry, error)
	GetAccountCostBreakdown(ctx context.Context, account string) (*AccountCostBreakdown, error)
	GetAccountCostForecasts(ctx context.Context, account string) (*AccountCostForecasts, error)
	GetAccountCostAlerts(ctx context.Context, account string) ([]*AccountCostAlert, error)
	GetAccountCostOptimizations(ctx context.Context, account string) ([]*AccountCostOptimization, error)
	GetAccountCostComparisons(ctx context.Context, account string) (*AccountCostComparisons, error)
	GetAccountBudgetUtilization(ctx context.Context, account string) (*AccountBudgetUtilization, error)
	GetAccountCostTrends(ctx context.Context, account string) (*AccountCostTrends, error)
	GetAccountCostAnalytics(ctx context.Context, account string) (*AccountCostAnalytics, error)
	GetAccountCostReports(ctx context.Context, account string) ([]*AccountCostReport, error)
	GetAccountCostPolicies(ctx context.Context, account string) ([]*AccountCostPolicy, error)
}

type AccountCostMetrics struct {
	AccountName          string
	TotalCost            float64
	CostToDate           float64
	EstimatedMonthlyCost float64
	CostPerCPUHour       float64
	CostPerGPUHour       float64
	CostPerMemoryGBHour  float64
	CostPerStorageGBHour float64
	AverageDailyCost     float64
	PeakDailyCost        float64
	CostEfficiency       float64
	CostPerJob           float64
	CostPerUser          float64
	LastUpdated          time.Time
}

type AccountBudgetInfo struct {
	AccountName       string
	TotalBudget       float64
	RemainingBudget   float64
	BudgetPeriod      string
	BudgetStartDate   time.Time
	BudgetEndDate     time.Time
	BudgetUtilization float64
	DaysRemaining     int
	BurnRate          float64
	ProjectedOverrun  float64
	BudgetStatus      string
	LastBudgetReset   time.Time
	BudgetAlertLevel  string
}

type AccountCostHistoryEntry struct {
	Date            time.Time
	DailyCost       float64
	CumulativeCost  float64
	JobCount        int
	CPUHours        float64
	GPUHours        float64
	MemoryGBHours   float64
	StorageGBHours  float64
	CostBreakdown   map[string]float64
	EfficiencyScore float64
}

type AccountCostBreakdown struct {
	AccountName       string
	ComputeCost       float64
	StorageCost       float64
	NetworkCost       float64
	LicenseCost       float64
	SupportCost       float64
	OverheadCost      float64
	CPUCost           float64
	GPUCost           float64
	MemoryCost        float64
	LocalStorageCost  float64
	SharedStorageCost float64
	BackupCost        float64
	MonitoringCost    float64
	SecurityCost      float64
}

type AccountCostForecasts struct {
	AccountName            string
	WeeklyForecast         float64
	MonthlyForecast        float64
	QuarterlyForecast      float64
	AnnualForecast         float64
	TrendDirection         string
	ConfidenceLevel        float64
	SeasonalFactors        map[string]float64
	GrowthRate             float64
	ProjectedBudgetOverrun float64
	OptimisticForecast     float64
	PessimisticForecast    float64
	ForecastAccuracy       float64
	LastForecastUpdate     time.Time
}

type AccountCostAlert struct {
	AlertID          string
	AccountName      string
	AlertType        string
	Severity         string
	Message          string
	CurrentValue     float64
	ThresholdValue   float64
	AlertTime        time.Time
	Status           string
	AcknowledgedBy   string
	AcknowledgedTime time.Time
	ResolvedTime     time.Time
	EscalationLevel  int
	AutoResolution   bool
}

type AccountCostOptimization struct {
	OptimizationID       string
	AccountName          string
	OptimizationType     string
	Description          string
	EstimatedSavings     float64
	ImplementationEffort string
	Priority             string
	Status               string
	CreatedTime          time.Time
	ImplementedTime      time.Time
	ActualSavings        float64
	ROI                  float64
	Recommendation       string
}

type AccountCostComparisons struct {
	AccountName             string
	PeerAccountsAvgCost     float64
	SimilarWorkloadsAvgCost float64
	IndustryBenchmark       float64
	CostRanking             int
	TotalAccounts           int
	CostPercentile          float64
	EfficiencyRanking       int
	CostVariance            float64
	ComparisonPeriod        string
	LastComparison          time.Time
}

type AccountBudgetUtilization struct {
	AccountName           string
	CurrentUtilization    float64
	ProjectedUtilization  float64
	UtilizationTrend      string
	DailyBurnRate         float64
	OptimalBurnRate       float64
	BurnRateVariance      float64
	TimeToDepletion       int
	BudgetHealth          string
	RecommendedAdjustment float64
	UtilizationHistory    []float64
	LastUtilizationUpdate time.Time
}

type AccountCostTrends struct {
	AccountName        string
	ShortTermTrend     string
	LongTermTrend      string
	SeasonalPattern    string
	CostVolatility     float64
	TrendStrength      float64
	CyclicalPatterns   []string
	AnomalyCount       int
	TrendChangePoints  []time.Time
	PredictiveAccuracy float64
	TrendConfidence    float64
	LastTrendAnalysis  time.Time
}

type AccountCostAnalytics struct {
	AccountName               string
	CostElasticity            float64
	UsageCorrelation          float64
	SeasonalityIndex          float64
	CostPredictability        float64
	ResourceUtilizationImpact float64
	UserBehaviorImpact        float64
	WorkloadPatternImpact     float64
	ExternalFactorImpact      float64
	CostDrivers               []string
	CostInhibitors            []string
	AnalyticsScore            float64
	LastAnalyticsUpdate       time.Time
}

type AccountCostReport struct {
	ReportID        string
	AccountName     string
	ReportType      string
	GeneratedTime   time.Time
	ReportPeriod    string
	TotalCost       float64
	CostBreakdown   map[string]float64
	BudgetStatus    string
	Recommendations []string
	KeyMetrics      map[string]float64
	Comparisons     map[string]float64
	Trends          map[string]string
	Alerts          []string
}

type AccountCostPolicy struct {
	PolicyID        string
	AccountName     string
	PolicyType      string
	PolicyName      string
	Description     string
	CostLimit       float64
	TimeWindow      string
	EnforcementMode string
	Violations      int
	LastViolation   time.Time
	CreatedTime     time.Time
	ModifiedTime    time.Time
	Active          bool
}

type AccountCostTrackingCollector struct {
	client AccountCostTrackingSLURMClient
	mutex  sync.RWMutex

	// Cost metrics
	totalCost            *prometheus.GaugeVec
	costToDate           *prometheus.GaugeVec
	estimatedMonthlyCost *prometheus.GaugeVec
	costPerCPUHour       *prometheus.GaugeVec
	costPerGPUHour       *prometheus.GaugeVec
	costPerMemoryGBHour  *prometheus.GaugeVec
	costPerStorageGBHour *prometheus.GaugeVec
	averageDailyCost     *prometheus.GaugeVec
	peakDailyCost        *prometheus.GaugeVec
	costEfficiency       *prometheus.GaugeVec
	costPerJob           *prometheus.GaugeVec
	costPerUser          *prometheus.GaugeVec

	// Budget metrics
	totalBudget       *prometheus.GaugeVec
	remainingBudget   *prometheus.GaugeVec
	budgetUtilization *prometheus.GaugeVec
	burnRate          *prometheus.GaugeVec
	projectedOverrun  *prometheus.GaugeVec
	daysRemaining     *prometheus.GaugeVec
	budgetStatus      *prometheus.GaugeVec
	budgetAlertLevel  *prometheus.GaugeVec

	// Cost breakdown metrics
	computeCost       *prometheus.GaugeVec
	storageCost       *prometheus.GaugeVec
	networkCost       *prometheus.GaugeVec
	licenseCost       *prometheus.GaugeVec
	supportCost       *prometheus.GaugeVec
	overheadCost      *prometheus.GaugeVec
	cpuCost           *prometheus.GaugeVec
	gpuCost           *prometheus.GaugeVec
	memoryCost        *prometheus.GaugeVec
	localStorageCost  *prometheus.GaugeVec
	sharedStorageCost *prometheus.GaugeVec
	backupCost        *prometheus.GaugeVec

	// Forecast metrics
	weeklyForecast     *prometheus.GaugeVec
	monthlyForecast    *prometheus.GaugeVec
	quarterlyForecast  *prometheus.GaugeVec
	annualForecast     *prometheus.GaugeVec
	forecastConfidence *prometheus.GaugeVec
	growthRate         *prometheus.GaugeVec
	forecastAccuracy   *prometheus.GaugeVec
	trendDirection     *prometheus.GaugeVec

	// Alert metrics
	costAlertsTotal     *prometheus.CounterVec
	costAlertsActive    *prometheus.GaugeVec
	costAlertsSeverity  *prometheus.GaugeVec
	costAlertsResolved  *prometheus.CounterVec
	costAlertsEscalated *prometheus.CounterVec

	// Optimization metrics
	optimizationsTotal    *prometheus.CounterVec
	optimizationsActive   *prometheus.GaugeVec
	estimatedSavings      *prometheus.GaugeVec
	actualSavings         *prometheus.GaugeVec
	optimizationROI       *prometheus.GaugeVec
	optimizationsPriority *prometheus.GaugeVec

	// Comparison metrics
	peerAccountsAvgCost *prometheus.GaugeVec
	industryBenchmark   *prometheus.GaugeVec
	costRanking         *prometheus.GaugeVec
	costPercentile      *prometheus.GaugeVec
	efficiencyRanking   *prometheus.GaugeVec

	// Utilization metrics
	currentUtilization   *prometheus.GaugeVec
	projectedUtilization *prometheus.GaugeVec
	dailyBurnRate        *prometheus.GaugeVec
	optimalBurnRate      *prometheus.GaugeVec
	burnRateVariance     *prometheus.GaugeVec
	timeToDepletion      *prometheus.GaugeVec

	// Trend metrics
	costVolatility     *prometheus.GaugeVec
	trendStrength      *prometheus.GaugeVec
	predictiveAccuracy *prometheus.GaugeVec
	trendConfidence    *prometheus.GaugeVec

	// Analytics metrics
	costElasticity            *prometheus.GaugeVec
	usageCorrelation          *prometheus.GaugeVec
	seasonalityIndex          *prometheus.GaugeVec
	costPredictability        *prometheus.GaugeVec
	resourceUtilizationImpact *prometheus.GaugeVec
	userBehaviorImpact        *prometheus.GaugeVec
	workloadPatternImpact     *prometheus.GaugeVec
	analyticsScore            *prometheus.GaugeVec

	// Policy metrics
	costPoliciesTotal    *prometheus.CounterVec
	costPoliciesActive   *prometheus.GaugeVec
	costPolicyViolations *prometheus.CounterVec
	costPolicyLimit      *prometheus.GaugeVec
}

func NewAccountCostTrackingCollector(client AccountCostTrackingSLURMClient) *AccountCostTrackingCollector {
	return &AccountCostTrackingCollector{
		client: client,

		// Cost metrics
		totalCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_total_cost",
				Help: "Total cost for account",
			},
			[]string{"account"},
		),
		costToDate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_to_date",
				Help: "Cost to date for account",
			},
			[]string{"account"},
		),
		estimatedMonthlyCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_estimated_monthly_cost",
				Help: "Estimated monthly cost for account",
			},
			[]string{"account"},
		),
		costPerCPUHour: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_per_cpu_hour",
				Help: "Cost per CPU hour for account",
			},
			[]string{"account"},
		),
		costPerGPUHour: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_per_gpu_hour",
				Help: "Cost per GPU hour for account",
			},
			[]string{"account"},
		),
		costPerMemoryGBHour: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_per_memory_gb_hour",
				Help: "Cost per memory GB hour for account",
			},
			[]string{"account"},
		),
		costPerStorageGBHour: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_per_storage_gb_hour",
				Help: "Cost per storage GB hour for account",
			},
			[]string{"account"},
		),
		averageDailyCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_average_daily_cost",
				Help: "Average daily cost for account",
			},
			[]string{"account"},
		),
		peakDailyCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_peak_daily_cost",
				Help: "Peak daily cost for account",
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
		costPerJob: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_per_job",
				Help: "Cost per job for account",
			},
			[]string{"account"},
		),
		costPerUser: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_per_user",
				Help: "Cost per user for account",
			},
			[]string{"account"},
		),

		// Budget metrics
		totalBudget: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_total_budget",
				Help: "Total budget for account",
			},
			[]string{"account", "period"},
		),
		remainingBudget: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_remaining_budget",
				Help: "Remaining budget for account",
			},
			[]string{"account", "period"},
		),
		budgetUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_budget_utilization",
				Help: "Budget utilization percentage for account",
			},
			[]string{"account", "period"},
		),
		burnRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_burn_rate",
				Help: "Budget burn rate for account",
			},
			[]string{"account"},
		),
		projectedOverrun: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_projected_overrun",
				Help: "Projected budget overrun for account",
			},
			[]string{"account"},
		),
		daysRemaining: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_budget_days_remaining",
				Help: "Days remaining in budget period for account",
			},
			[]string{"account"},
		),
		budgetStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_budget_status",
				Help: "Budget status for account (0=healthy, 1=warning, 2=critical)",
			},
			[]string{"account", "status"},
		),
		budgetAlertLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_budget_alert_level",
				Help: "Budget alert level for account (0=none, 1=low, 2=medium, 3=high)",
			},
			[]string{"account", "level"},
		),

		// Cost breakdown metrics
		computeCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_compute_cost",
				Help: "Compute cost for account",
			},
			[]string{"account"},
		),
		storageCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_storage_cost",
				Help: "Storage cost for account",
			},
			[]string{"account"},
		),
		networkCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_network_cost",
				Help: "Network cost for account",
			},
			[]string{"account"},
		),
		licenseCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_license_cost",
				Help: "License cost for account",
			},
			[]string{"account"},
		),
		supportCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_support_cost",
				Help: "Support cost for account",
			},
			[]string{"account"},
		),
		overheadCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_overhead_cost",
				Help: "Overhead cost for account",
			},
			[]string{"account"},
		),
		cpuCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cpu_cost",
				Help: "CPU cost for account",
			},
			[]string{"account"},
		),
		gpuCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_gpu_cost",
				Help: "GPU cost for account",
			},
			[]string{"account"},
		),
		memoryCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_memory_cost",
				Help: "Memory cost for account",
			},
			[]string{"account"},
		),
		localStorageCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_local_storage_cost",
				Help: "Local storage cost for account",
			},
			[]string{"account"},
		),
		sharedStorageCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_shared_storage_cost",
				Help: "Shared storage cost for account",
			},
			[]string{"account"},
		),
		backupCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_backup_cost",
				Help: "Backup cost for account",
			},
			[]string{"account"},
		),

		// Forecast metrics
		weeklyForecast: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_weekly_forecast",
				Help: "Weekly cost forecast for account",
			},
			[]string{"account"},
		),
		monthlyForecast: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_monthly_forecast",
				Help: "Monthly cost forecast for account",
			},
			[]string{"account"},
		),
		quarterlyForecast: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_quarterly_forecast",
				Help: "Quarterly cost forecast for account",
			},
			[]string{"account"},
		),
		annualForecast: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_annual_forecast",
				Help: "Annual cost forecast for account",
			},
			[]string{"account"},
		),
		forecastConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_forecast_confidence",
				Help: "Cost forecast confidence level for account",
			},
			[]string{"account"},
		),
		growthRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_growth_rate",
				Help: "Cost growth rate for account",
			},
			[]string{"account"},
		),
		forecastAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_forecast_accuracy",
				Help: "Cost forecast accuracy for account",
			},
			[]string{"account"},
		),
		trendDirection: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_trend_direction",
				Help: "Cost trend direction for account (1=up, 0=stable, -1=down)",
			},
			[]string{"account", "direction"},
		),

		// Alert metrics
		costAlertsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_cost_alerts_total",
				Help: "Total number of cost alerts for account",
			},
			[]string{"account", "type", "severity"},
		),
		costAlertsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_alerts_active",
				Help: "Number of active cost alerts for account",
			},
			[]string{"account", "type", "severity"},
		),
		costAlertsSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_alerts_by_severity",
				Help: "Number of cost alerts by severity for account",
			},
			[]string{"account", "severity"},
		),
		costAlertsResolved: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_cost_alerts_resolved_total",
				Help: "Total number of resolved cost alerts for account",
			},
			[]string{"account", "type"},
		),
		costAlertsEscalated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_cost_alerts_escalated_total",
				Help: "Total number of escalated cost alerts for account",
			},
			[]string{"account", "type"},
		),

		// Optimization metrics
		optimizationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_cost_optimizations_total",
				Help: "Total number of cost optimizations for account",
			},
			[]string{"account", "type", "status"},
		),
		optimizationsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_optimizations_active",
				Help: "Number of active cost optimizations for account",
			},
			[]string{"account", "type"},
		),
		estimatedSavings: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_estimated_savings",
				Help: "Estimated cost savings for account",
			},
			[]string{"account", "optimization_type"},
		),
		actualSavings: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_actual_savings",
				Help: "Actual cost savings for account",
			},
			[]string{"account", "optimization_type"},
		),
		optimizationROI: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_optimization_roi",
				Help: "Return on investment for cost optimizations",
			},
			[]string{"account", "optimization_type"},
		),
		optimizationsPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_optimizations_by_priority",
				Help: "Number of cost optimizations by priority for account",
			},
			[]string{"account", "priority"},
		),

		// Comparison metrics
		peerAccountsAvgCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_peer_accounts_avg_cost",
				Help: "Average cost of peer accounts",
			},
			[]string{"account"},
		),
		industryBenchmark: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_industry_benchmark_cost",
				Help: "Industry benchmark cost for account",
			},
			[]string{"account"},
		),
		costRanking: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_ranking",
				Help: "Cost ranking of account among peers",
			},
			[]string{"account"},
		),
		costPercentile: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_percentile",
				Help: "Cost percentile of account among peers",
			},
			[]string{"account"},
		),
		efficiencyRanking: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_efficiency_ranking",
				Help: "Efficiency ranking of account among peers",
			},
			[]string{"account"},
		),

		// Utilization metrics
		currentUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_current_budget_utilization",
				Help: "Current budget utilization for account",
			},
			[]string{"account"},
		),
		projectedUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_projected_budget_utilization",
				Help: "Projected budget utilization for account",
			},
			[]string{"account"},
		),
		dailyBurnRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_daily_burn_rate",
				Help: "Daily budget burn rate for account",
			},
			[]string{"account"},
		),
		optimalBurnRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_optimal_burn_rate",
				Help: "Optimal daily burn rate for account",
			},
			[]string{"account"},
		),
		burnRateVariance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_burn_rate_variance",
				Help: "Burn rate variance for account",
			},
			[]string{"account"},
		),
		timeToDepletion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_time_to_budget_depletion_days",
				Help: "Estimated days until budget depletion for account",
			},
			[]string{"account"},
		),

		// Trend metrics
		costVolatility: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_volatility",
				Help: "Cost volatility measure for account",
			},
			[]string{"account"},
		),
		trendStrength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_trend_strength",
				Help: "Cost trend strength for account",
			},
			[]string{"account"},
		),
		predictiveAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_predictive_accuracy",
				Help: "Cost prediction accuracy for account",
			},
			[]string{"account"},
		),
		trendConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_trend_confidence",
				Help: "Cost trend confidence for account",
			},
			[]string{"account"},
		),

		// Analytics metrics
		costElasticity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_elasticity",
				Help: "Cost elasticity measure for account",
			},
			[]string{"account"},
		),
		usageCorrelation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_usage_cost_correlation",
				Help: "Usage-cost correlation for account",
			},
			[]string{"account"},
		),
		seasonalityIndex: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_seasonality_index",
				Help: "Cost seasonality index for account",
			},
			[]string{"account"},
		),
		costPredictability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_predictability",
				Help: "Cost predictability score for account",
			},
			[]string{"account"},
		),
		resourceUtilizationImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_utilization_cost_impact",
				Help: "Resource utilization impact on cost for account",
			},
			[]string{"account"},
		),
		userBehaviorImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_user_behavior_cost_impact",
				Help: "User behavior impact on cost for account",
			},
			[]string{"account"},
		),
		workloadPatternImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_workload_pattern_cost_impact",
				Help: "Workload pattern impact on cost for account",
			},
			[]string{"account"},
		),
		analyticsScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_analytics_score",
				Help: "Overall cost analytics score for account",
			},
			[]string{"account"},
		),

		// Policy metrics
		costPoliciesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_cost_policies_total",
				Help: "Total number of cost policies for account",
			},
			[]string{"account", "type"},
		),
		costPoliciesActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_policies_active",
				Help: "Number of active cost policies for account",
			},
			[]string{"account", "type"},
		),
		costPolicyViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_cost_policy_violations_total",
				Help: "Total number of cost policy violations for account",
			},
			[]string{"account", "policy_type"},
		),
		costPolicyLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_policy_limit",
				Help: "Cost policy limit for account",
			},
			[]string{"account", "policy_type"},
		),
	}
}

func (c *AccountCostTrackingCollector) Describe(ch chan<- *prometheus.Desc) {
	c.totalCost.Describe(ch)
	c.costToDate.Describe(ch)
	c.estimatedMonthlyCost.Describe(ch)
	c.costPerCPUHour.Describe(ch)
	c.costPerGPUHour.Describe(ch)
	c.costPerMemoryGBHour.Describe(ch)
	c.costPerStorageGBHour.Describe(ch)
	c.averageDailyCost.Describe(ch)
	c.peakDailyCost.Describe(ch)
	c.costEfficiency.Describe(ch)
	c.costPerJob.Describe(ch)
	c.costPerUser.Describe(ch)
	c.totalBudget.Describe(ch)
	c.remainingBudget.Describe(ch)
	c.budgetUtilization.Describe(ch)
	c.burnRate.Describe(ch)
	c.projectedOverrun.Describe(ch)
	c.daysRemaining.Describe(ch)
	c.budgetStatus.Describe(ch)
	c.budgetAlertLevel.Describe(ch)
	c.computeCost.Describe(ch)
	c.storageCost.Describe(ch)
	c.networkCost.Describe(ch)
	c.licenseCost.Describe(ch)
	c.supportCost.Describe(ch)
	c.overheadCost.Describe(ch)
	c.cpuCost.Describe(ch)
	c.gpuCost.Describe(ch)
	c.memoryCost.Describe(ch)
	c.localStorageCost.Describe(ch)
	c.sharedStorageCost.Describe(ch)
	c.backupCost.Describe(ch)
	c.weeklyForecast.Describe(ch)
	c.monthlyForecast.Describe(ch)
	c.quarterlyForecast.Describe(ch)
	c.annualForecast.Describe(ch)
	c.forecastConfidence.Describe(ch)
	c.growthRate.Describe(ch)
	c.forecastAccuracy.Describe(ch)
	c.trendDirection.Describe(ch)
	c.costAlertsTotal.Describe(ch)
	c.costAlertsActive.Describe(ch)
	c.costAlertsSeverity.Describe(ch)
	c.costAlertsResolved.Describe(ch)
	c.costAlertsEscalated.Describe(ch)
	c.optimizationsTotal.Describe(ch)
	c.optimizationsActive.Describe(ch)
	c.estimatedSavings.Describe(ch)
	c.actualSavings.Describe(ch)
	c.optimizationROI.Describe(ch)
	c.optimizationsPriority.Describe(ch)
	c.peerAccountsAvgCost.Describe(ch)
	c.industryBenchmark.Describe(ch)
	c.costRanking.Describe(ch)
	c.costPercentile.Describe(ch)
	c.efficiencyRanking.Describe(ch)
	c.currentUtilization.Describe(ch)
	c.projectedUtilization.Describe(ch)
	c.dailyBurnRate.Describe(ch)
	c.optimalBurnRate.Describe(ch)
	c.burnRateVariance.Describe(ch)
	c.timeToDepletion.Describe(ch)
	c.costVolatility.Describe(ch)
	c.trendStrength.Describe(ch)
	c.predictiveAccuracy.Describe(ch)
	c.trendConfidence.Describe(ch)
	c.costElasticity.Describe(ch)
	c.usageCorrelation.Describe(ch)
	c.seasonalityIndex.Describe(ch)
	c.costPredictability.Describe(ch)
	c.resourceUtilizationImpact.Describe(ch)
	c.userBehaviorImpact.Describe(ch)
	c.workloadPatternImpact.Describe(ch)
	c.analyticsScore.Describe(ch)
	c.costPoliciesTotal.Describe(ch)
	c.costPoliciesActive.Describe(ch)
	c.costPolicyViolations.Describe(ch)
	c.costPolicyLimit.Describe(ch)
}

func (c *AccountCostTrackingCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx := context.Background()
	accounts := []string{"account1", "account2", "account3"}

	for _, account := range accounts {
		c.collectCostMetrics(ctx, account, ch)
		c.collectBudgetMetrics(ctx, account, ch)
		c.collectCostBreakdown(ctx, account, ch)
		c.collectForecastMetrics(ctx, account, ch)
		c.collectAlertMetrics(ctx, account, ch)
		c.collectOptimizationMetrics(ctx, account, ch)
		c.collectComparisonMetrics(ctx, account, ch)
		c.collectUtilizationMetrics(ctx, account, ch)
		c.collectTrendMetrics(ctx, account, ch)
		c.collectAnalyticsMetrics(ctx, account, ch)
		c.collectPolicyMetrics(ctx, account, ch)
	}

	c.totalCost.Collect(ch)
	c.costToDate.Collect(ch)
	c.estimatedMonthlyCost.Collect(ch)
	c.costPerCPUHour.Collect(ch)
	c.costPerGPUHour.Collect(ch)
	c.costPerMemoryGBHour.Collect(ch)
	c.costPerStorageGBHour.Collect(ch)
	c.averageDailyCost.Collect(ch)
	c.peakDailyCost.Collect(ch)
	c.costEfficiency.Collect(ch)
	c.costPerJob.Collect(ch)
	c.costPerUser.Collect(ch)
	c.totalBudget.Collect(ch)
	c.remainingBudget.Collect(ch)
	c.budgetUtilization.Collect(ch)
	c.burnRate.Collect(ch)
	c.projectedOverrun.Collect(ch)
	c.daysRemaining.Collect(ch)
	c.budgetStatus.Collect(ch)
	c.budgetAlertLevel.Collect(ch)
	c.computeCost.Collect(ch)
	c.storageCost.Collect(ch)
	c.networkCost.Collect(ch)
	c.licenseCost.Collect(ch)
	c.supportCost.Collect(ch)
	c.overheadCost.Collect(ch)
	c.cpuCost.Collect(ch)
	c.gpuCost.Collect(ch)
	c.memoryCost.Collect(ch)
	c.localStorageCost.Collect(ch)
	c.sharedStorageCost.Collect(ch)
	c.backupCost.Collect(ch)
	c.weeklyForecast.Collect(ch)
	c.monthlyForecast.Collect(ch)
	c.quarterlyForecast.Collect(ch)
	c.annualForecast.Collect(ch)
	c.forecastConfidence.Collect(ch)
	c.growthRate.Collect(ch)
	c.forecastAccuracy.Collect(ch)
	c.trendDirection.Collect(ch)
	c.costAlertsTotal.Collect(ch)
	c.costAlertsActive.Collect(ch)
	c.costAlertsSeverity.Collect(ch)
	c.costAlertsResolved.Collect(ch)
	c.costAlertsEscalated.Collect(ch)
	c.optimizationsTotal.Collect(ch)
	c.optimizationsActive.Collect(ch)
	c.estimatedSavings.Collect(ch)
	c.actualSavings.Collect(ch)
	c.optimizationROI.Collect(ch)
	c.optimizationsPriority.Collect(ch)
	c.peerAccountsAvgCost.Collect(ch)
	c.industryBenchmark.Collect(ch)
	c.costRanking.Collect(ch)
	c.costPercentile.Collect(ch)
	c.efficiencyRanking.Collect(ch)
	c.currentUtilization.Collect(ch)
	c.projectedUtilization.Collect(ch)
	c.dailyBurnRate.Collect(ch)
	c.optimalBurnRate.Collect(ch)
	c.burnRateVariance.Collect(ch)
	c.timeToDepletion.Collect(ch)
	c.costVolatility.Collect(ch)
	c.trendStrength.Collect(ch)
	c.predictiveAccuracy.Collect(ch)
	c.trendConfidence.Collect(ch)
	c.costElasticity.Collect(ch)
	c.usageCorrelation.Collect(ch)
	c.seasonalityIndex.Collect(ch)
	c.costPredictability.Collect(ch)
	c.resourceUtilizationImpact.Collect(ch)
	c.userBehaviorImpact.Collect(ch)
	c.workloadPatternImpact.Collect(ch)
	c.analyticsScore.Collect(ch)
	c.costPoliciesTotal.Collect(ch)
	c.costPoliciesActive.Collect(ch)
	c.costPolicyViolations.Collect(ch)
	c.costPolicyLimit.Collect(ch)
}

func (c *AccountCostTrackingCollector) collectCostMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetAccountCostMetrics(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost metrics for account %s: %v", account, err)
		return
	}

	c.totalCost.WithLabelValues(account).Set(metrics.TotalCost)
	c.costToDate.WithLabelValues(account).Set(metrics.CostToDate)
	c.estimatedMonthlyCost.WithLabelValues(account).Set(metrics.EstimatedMonthlyCost)
	c.costPerCPUHour.WithLabelValues(account).Set(metrics.CostPerCPUHour)
	c.costPerGPUHour.WithLabelValues(account).Set(metrics.CostPerGPUHour)
	c.costPerMemoryGBHour.WithLabelValues(account).Set(metrics.CostPerMemoryGBHour)
	c.costPerStorageGBHour.WithLabelValues(account).Set(metrics.CostPerStorageGBHour)
	c.averageDailyCost.WithLabelValues(account).Set(metrics.AverageDailyCost)
	c.peakDailyCost.WithLabelValues(account).Set(metrics.PeakDailyCost)
	c.costEfficiency.WithLabelValues(account).Set(metrics.CostEfficiency)
	c.costPerJob.WithLabelValues(account).Set(metrics.CostPerJob)
	c.costPerUser.WithLabelValues(account).Set(metrics.CostPerUser)
}

func (c *AccountCostTrackingCollector) collectBudgetMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	budget, err := c.client.GetAccountBudgetInfo(ctx, account)
	if err != nil {
		log.Printf("Error collecting budget info for account %s: %v", account, err)
		return
	}

	c.totalBudget.WithLabelValues(account, budget.BudgetPeriod).Set(budget.TotalBudget)
	c.remainingBudget.WithLabelValues(account, budget.BudgetPeriod).Set(budget.RemainingBudget)
	c.budgetUtilization.WithLabelValues(account, budget.BudgetPeriod).Set(budget.BudgetUtilization)
	c.burnRate.WithLabelValues(account).Set(budget.BurnRate)
	c.projectedOverrun.WithLabelValues(account).Set(budget.ProjectedOverrun)
	c.daysRemaining.WithLabelValues(account).Set(float64(budget.DaysRemaining))

	var statusValue float64
	switch budget.BudgetStatus {
	case "healthy":
		statusValue = 0
	case "warning":
		statusValue = 1
	case "critical":
		statusValue = 2
	}
	c.budgetStatus.WithLabelValues(account, budget.BudgetStatus).Set(statusValue)

	var alertValue float64
	switch budget.BudgetAlertLevel {
	case "none":
		alertValue = 0
	case "low":
		alertValue = 1
	case "medium":
		alertValue = 2
	case "high":
		alertValue = 3
	}
	c.budgetAlertLevel.WithLabelValues(account, budget.BudgetAlertLevel).Set(alertValue)
}

func (c *AccountCostTrackingCollector) collectCostBreakdown(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	breakdown, err := c.client.GetAccountCostBreakdown(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost breakdown for account %s: %v", account, err)
		return
	}

	c.computeCost.WithLabelValues(account).Set(breakdown.ComputeCost)
	c.storageCost.WithLabelValues(account).Set(breakdown.StorageCost)
	c.networkCost.WithLabelValues(account).Set(breakdown.NetworkCost)
	c.licenseCost.WithLabelValues(account).Set(breakdown.LicenseCost)
	c.supportCost.WithLabelValues(account).Set(breakdown.SupportCost)
	c.overheadCost.WithLabelValues(account).Set(breakdown.OverheadCost)
	c.cpuCost.WithLabelValues(account).Set(breakdown.CPUCost)
	c.gpuCost.WithLabelValues(account).Set(breakdown.GPUCost)
	c.memoryCost.WithLabelValues(account).Set(breakdown.MemoryCost)
	c.localStorageCost.WithLabelValues(account).Set(breakdown.LocalStorageCost)
	c.sharedStorageCost.WithLabelValues(account).Set(breakdown.SharedStorageCost)
	c.backupCost.WithLabelValues(account).Set(breakdown.BackupCost)
}

func (c *AccountCostTrackingCollector) collectForecastMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	forecasts, err := c.client.GetAccountCostForecasts(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost forecasts for account %s: %v", account, err)
		return
	}

	c.weeklyForecast.WithLabelValues(account).Set(forecasts.WeeklyForecast)
	c.monthlyForecast.WithLabelValues(account).Set(forecasts.MonthlyForecast)
	c.quarterlyForecast.WithLabelValues(account).Set(forecasts.QuarterlyForecast)
	c.annualForecast.WithLabelValues(account).Set(forecasts.AnnualForecast)
	c.forecastConfidence.WithLabelValues(account).Set(forecasts.ConfidenceLevel)
	c.growthRate.WithLabelValues(account).Set(forecasts.GrowthRate)
	c.forecastAccuracy.WithLabelValues(account).Set(forecasts.ForecastAccuracy)

	var trendValue float64
	switch forecasts.TrendDirection {
	case "up":
		trendValue = 1
	case "stable":
		trendValue = 0
	case "down":
		trendValue = -1
	}
	c.trendDirection.WithLabelValues(account, forecasts.TrendDirection).Set(trendValue)
}

func (c *AccountCostTrackingCollector) collectAlertMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	alerts, err := c.client.GetAccountCostAlerts(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost alerts for account %s: %v", account, err)
		return
	}

	alertCounts := make(map[string]map[string]int)
	activeCounts := make(map[string]map[string]int)
	severityCounts := make(map[string]int)
	resolvedCounts := make(map[string]int)
	escalatedCounts := make(map[string]int)

	for _, alert := range alerts {
		if alertCounts[alert.AlertType] == nil {
			alertCounts[alert.AlertType] = make(map[string]int)
		}
		alertCounts[alert.AlertType][alert.Severity]++

		if alert.Status == "active" {
			if activeCounts[alert.AlertType] == nil {
				activeCounts[alert.AlertType] = make(map[string]int)
			}
			activeCounts[alert.AlertType][alert.Severity]++
		}

		severityCounts[alert.Severity]++

		if alert.Status == "resolved" {
			resolvedCounts[alert.AlertType]++
		}

		if alert.EscalationLevel > 0 {
			escalatedCounts[alert.AlertType]++
		}
	}

	for alertType, severityMap := range alertCounts {
		for severity, count := range severityMap {
			c.costAlertsTotal.WithLabelValues(account, alertType, severity).Add(float64(count))
		}
	}

	for alertType, severityMap := range activeCounts {
		for severity, count := range severityMap {
			c.costAlertsActive.WithLabelValues(account, alertType, severity).Set(float64(count))
		}
	}

	for severity, count := range severityCounts {
		c.costAlertsSeverity.WithLabelValues(account, severity).Set(float64(count))
	}

	for alertType, count := range resolvedCounts {
		c.costAlertsResolved.WithLabelValues(account, alertType).Add(float64(count))
	}

	for alertType, count := range escalatedCounts {
		c.costAlertsEscalated.WithLabelValues(account, alertType).Add(float64(count))
	}
}

func (c *AccountCostTrackingCollector) collectOptimizationMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	optimizations, err := c.client.GetAccountCostOptimizations(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost optimizations for account %s: %v", account, err)
		return
	}

	optimizationCounts := make(map[string]map[string]int)
	activeCounts := make(map[string]int)
	priorityCounts := make(map[string]int)

	for _, opt := range optimizations {
		if optimizationCounts[opt.OptimizationType] == nil {
			optimizationCounts[opt.OptimizationType] = make(map[string]int)
		}
		optimizationCounts[opt.OptimizationType][opt.Status]++

		if opt.Status == "active" {
			activeCounts[opt.OptimizationType]++
		}

		priorityCounts[opt.Priority]++

		c.estimatedSavings.WithLabelValues(account, opt.OptimizationType).Set(opt.EstimatedSavings)
		c.actualSavings.WithLabelValues(account, opt.OptimizationType).Set(opt.ActualSavings)
		c.optimizationROI.WithLabelValues(account, opt.OptimizationType).Set(opt.ROI)
	}

	for optType, statusMap := range optimizationCounts {
		for status, count := range statusMap {
			c.optimizationsTotal.WithLabelValues(account, optType, status).Add(float64(count))
		}
	}

	for optType, count := range activeCounts {
		c.optimizationsActive.WithLabelValues(account, optType).Set(float64(count))
	}

	for priority, count := range priorityCounts {
		c.optimizationsPriority.WithLabelValues(account, priority).Set(float64(count))
	}
}

func (c *AccountCostTrackingCollector) collectComparisonMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	comparisons, err := c.client.GetAccountCostComparisons(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost comparisons for account %s: %v", account, err)
		return
	}

	c.peerAccountsAvgCost.WithLabelValues(account).Set(comparisons.PeerAccountsAvgCost)
	c.industryBenchmark.WithLabelValues(account).Set(comparisons.IndustryBenchmark)
	c.costRanking.WithLabelValues(account).Set(float64(comparisons.CostRanking))
	c.costPercentile.WithLabelValues(account).Set(comparisons.CostPercentile)
	c.efficiencyRanking.WithLabelValues(account).Set(float64(comparisons.EfficiencyRanking))
}

func (c *AccountCostTrackingCollector) collectUtilizationMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	utilization, err := c.client.GetAccountBudgetUtilization(ctx, account)
	if err != nil {
		log.Printf("Error collecting budget utilization for account %s: %v", account, err)
		return
	}

	c.currentUtilization.WithLabelValues(account).Set(utilization.CurrentUtilization)
	c.projectedUtilization.WithLabelValues(account).Set(utilization.ProjectedUtilization)
	c.dailyBurnRate.WithLabelValues(account).Set(utilization.DailyBurnRate)
	c.optimalBurnRate.WithLabelValues(account).Set(utilization.OptimalBurnRate)
	c.burnRateVariance.WithLabelValues(account).Set(utilization.BurnRateVariance)
	c.timeToDepletion.WithLabelValues(account).Set(float64(utilization.TimeToDepletion))
}

func (c *AccountCostTrackingCollector) collectTrendMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	trends, err := c.client.GetAccountCostTrends(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost trends for account %s: %v", account, err)
		return
	}

	c.costVolatility.WithLabelValues(account).Set(trends.CostVolatility)
	c.trendStrength.WithLabelValues(account).Set(trends.TrendStrength)
	c.predictiveAccuracy.WithLabelValues(account).Set(trends.PredictiveAccuracy)
	c.trendConfidence.WithLabelValues(account).Set(trends.TrendConfidence)
}

func (c *AccountCostTrackingCollector) collectAnalyticsMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	analytics, err := c.client.GetAccountCostAnalytics(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost analytics for account %s: %v", account, err)
		return
	}

	c.costElasticity.WithLabelValues(account).Set(analytics.CostElasticity)
	c.usageCorrelation.WithLabelValues(account).Set(analytics.UsageCorrelation)
	c.seasonalityIndex.WithLabelValues(account).Set(analytics.SeasonalityIndex)
	c.costPredictability.WithLabelValues(account).Set(analytics.CostPredictability)
	c.resourceUtilizationImpact.WithLabelValues(account).Set(analytics.ResourceUtilizationImpact)
	c.userBehaviorImpact.WithLabelValues(account).Set(analytics.UserBehaviorImpact)
	c.workloadPatternImpact.WithLabelValues(account).Set(analytics.WorkloadPatternImpact)
	c.analyticsScore.WithLabelValues(account).Set(analytics.AnalyticsScore)
}

func (c *AccountCostTrackingCollector) collectPolicyMetrics(ctx context.Context, account string, ch chan<- prometheus.Metric) {
	policies, err := c.client.GetAccountCostPolicies(ctx, account)
	if err != nil {
		log.Printf("Error collecting cost policies for account %s: %v", account, err)
		return
	}

	policyCounts := make(map[string]int)
	activeCounts := make(map[string]int)
	violationCounts := make(map[string]int)

	for _, policy := range policies {
		policyCounts[policy.PolicyType]++

		if policy.Active {
			activeCounts[policy.PolicyType]++
		}

		violationCounts[policy.PolicyType] += policy.Violations

		c.costPolicyLimit.WithLabelValues(account, policy.PolicyType).Set(policy.CostLimit)
	}

	for policyType, count := range policyCounts {
		c.costPoliciesTotal.WithLabelValues(account, policyType).Add(float64(count))
	}

	for policyType, count := range activeCounts {
		c.costPoliciesActive.WithLabelValues(account, policyType).Set(float64(count))
	}

	for policyType, count := range violationCounts {
		c.costPolicyViolations.WithLabelValues(account, policyType).Add(float64(count))
	}
}
