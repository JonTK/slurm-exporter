package collector

import "time"

// Common types used across multiple collectors to avoid duplication

// AccountHierarchy represents the hierarchical structure of accounts in SLURM
type AccountHierarchy struct {
	Root             *AccountNode
	RootAccounts     []*AccountNode          // Multiple root accounts
	AccountMap       map[string]*AccountNode // Map for quick lookups
	TotalAccounts    int
	MaxDepth         int
	BalanceScore     float64 // 0-1, how well balanced the tree is
	UtilizationScore float64 // 0-1, how well utilized the accounts are

	// Resource distribution
	TotalShares      int64
	TotalRawUsage    int64
	EffectiveUsage   int64
	NormalizedShares float64

	// Structural analysis
	LeafAccounts   int
	BranchAccounts int
	OrphanAccounts int
	CircularRefs   int

	// Share distribution metrics
	ShareGini     float64 // Gini coefficient for share distribution
	ShareEntropy  float64 // Shannon entropy of share distribution
	ShareSkewness float64 // Statistical skewness

	// Quality metrics
	ValidationErrors []string
	Warnings         []string
	LastUpdated      time.Time
	EfficiencyScore  float64
	CoverageScore    float64
	DataCompleteness float64
	DataConsistency  float64

	// Additional fields
	HierarchyType       string // Type of hierarchy (e.g., "account", "user", etc.)
	TotalUsers          int    // Total number of users in the hierarchy
	OrganizationalUnits map[string]*OrganizationalUnit
}

// AccountNode represents a single node in the account hierarchy tree
type AccountNode struct {
	// Basic information
	Name          string
	AccountName   string // Alternative field name
	Description   string
	Parent        string
	ParentAccount string // Alternative field name for Parent
	Children      []*AccountNode
	ChildAccounts []*AccountNode // Alternative field name for Children
	Level         int
	Path          string // Full path from root
	Status        string // Account status (active, inactive, etc.)
	ActiveUsers   int    // Number of active users

	// Resource limits and quotas
	CPULimit      int64
	MemoryLimit   int64
	GPULimit      int64
	WallTimeLimit int64

	// Usage and quotas
	RawUsage        int64
	EffectiveUsage  int64
	NormalizedUsage float64
	QuotaAllocated  float64
	QuotaUsed       float64
	QuotaRemaining  float64

	// Fair-share related
	RawShares        int64
	NormalizedShares float64
	EffectiveShares  float64
	FairShareValue   float64
	Priority         int

	// Metrics
	JobCount        int
	UserCount       int
	DirectUsers     int
	SubAccountCount int
	TotalCPUHours   float64
	TotalGPUHours   float64
	AvgJobSize      float64
	AvgWaitTime     time.Duration

	// Efficiency metrics
	CPUEfficiency     float64
	MemoryEfficiency  float64
	OverallEfficiency float64

	// Status and validation
	IsActive         bool
	IsDefault        bool
	IsRoot           bool
	IsLeaf           bool
	HasQuota         bool
	HasDeficit       bool
	ValidationStatus string
	ValidationErrors []string

	// Additional fields for account collector
	ActiveUserCount int
	CurrentUsage    *AccountUsageInfo
	DataQuality     float64
	AllUsers        []string // All users in this account (for iteration)
}

// AccountUsageInfo represents current usage information for an account
type AccountUsageInfo struct {
	CPUHours           float64
	MemoryMBHours      float64
	ResourceEfficiency float64
	WasteRatio         float64
	EstimatedCost      float64
}

// QuotaViolation represents a quota violation event with comprehensive tracking
type QuotaViolation struct {
	Account       string    `json:"account"`
	UserName      string    `json:"username,omitempty"`
	JobID         string    `json:"job_id,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	Type          string    `json:"type"`           // CPU, Memory, GPU, Jobs, etc.
	ViolationType string    `json:"violation_type"` // Alternative field name for Type
	Severity      string    `json:"severity"`       // Warning, Error, Critical
	Limit         float64   `json:"limit"`
	Requested     float64   `json:"requested"`
	Current       float64   `json:"current"`
	Message       string    `json:"message"`
	Impact        string    `json:"impact,omitempty"`

	// Additional tracking
	ResourceType string        `json:"resource_type"`
	QuotaType    string        `json:"quota_type"` // user, account, qos
	Duration     time.Duration `json:"duration,omitempty"`

	// Resolution tracking
	ResolutionRequired bool       `json:"resolution_required"`
	ResolutionActions  []string   `json:"resolution_actions,omitempty"`
	AutoResolved       bool       `json:"auto_resolved"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy         string     `json:"resolved_by,omitempty"`

	// Recurrence tracking
	RecurrenceCount   int       `json:"recurrence_count"`
	FirstOccurrence   time.Time `json:"first_occurrence"`
	LastOccurrence    time.Time `json:"last_occurrence"`
	RecurrencePattern string    `json:"recurrence_pattern,omitempty"`
}

// QuotaRecommendation provides suggestions for quota optimization
type QuotaRecommendation struct {
	Account          string  `json:"account"`
	Type             string  `json:"type"`
	Priority         string  `json:"priority"`
	CurrentValue     float64 `json:"current_value"`
	RecommendedValue float64 `json:"recommended_value"`
	Reason           string  `json:"reason"`
	PotentialSavings float64 `json:"potential_savings,omitempty"`
	EstimatedImpact  string  `json:"estimated_impact,omitempty"`

	// Implementation details
	ImplementationSteps []string `json:"implementation_steps,omitempty"`
	Prerequisites       []string `json:"prerequisites,omitempty"`
	RiskAssessment      string   `json:"risk_assessment,omitempty"`

	// Metrics
	Confidence        float64 `json:"confidence"`                   // 0-1
	HistoricalSuccess float64 `json:"historical_success,omitempty"` // Success rate of similar recommendations
	ExpectedROI       float64 `json:"expected_roi,omitempty"`
}

// OptimizationOpportunity represents a potential optimization
type OptimizationOpportunity struct {
	Type             string  `json:"type"`
	Description      string  `json:"description"`
	Account          string  `json:"account,omitempty"`
	PotentialSavings float64 `json:"potential_savings"`
	Implementation   string  `json:"implementation"`
	Priority         int     `json:"priority"`         // 1-5, 5 being highest
	Effort           string  `json:"effort,omitempty"` // Low, Medium, High

	// Detailed actions and impact
	Actions        []string           `json:"actions,omitempty"`
	ResourceImpact map[string]float64 `json:"resource_impact,omitempty"`

	// Confidence and validation
	Confidence    float64   `json:"confidence"` // 0-1
	DataPoints    int       `json:"data_points,omitempty"`
	LastValidated time.Time `json:"last_validated,omitempty"`
}

// ResourceUsagePattern represents patterns in resource usage
type ResourceUsagePattern struct {
	PeakHours      []int
	OffPeakHours   []int
	AverageUsage   float64
	PeakUsage      float64
	Variability    float64
	Predictability float64
	SeasonalTrends []string

	// Efficiency metrics
	EfficiencyScore float64
	WastePercentage float64

	// Optimization potential
	OptimizationPotential float64
	RecommendedActions    []string
}

// EfficiencyMetrics holds comprehensive efficiency measurements
type EfficiencyMetrics struct {
	// Core efficiency metrics
	OverallEfficiency float64 `json:"overall_efficiency"`
	CPUEfficiency     float64 `json:"cpu_efficiency"`
	MemoryEfficiency  float64 `json:"memory_efficiency"`
	GPUEfficiency     float64 `json:"gpu_efficiency,omitempty"`
	IOEfficiency      float64 `json:"io_efficiency,omitempty"`
	NetworkEfficiency float64 `json:"network_efficiency,omitempty"`
	QueueEfficiency   float64 `json:"queue_efficiency,omitempty"`

	// Composite scores
	ResourceEfficiency   float64 `json:"resource_efficiency"`   // CPU + Memory
	ThroughputEfficiency float64 `json:"throughput_efficiency"` // I/O + Network

	// Waste and optimization
	ResourceWaste        *ResourceWaste `json:"resource_waste,omitempty"`
	WasteRatio           float64        `json:"waste_ratio"`
	OptimizationScore    float64        `json:"optimization_score"`
	OptimalityScore      float64        `json:"optimality_score"`
	ImprovementPotential float64        `json:"improvement_potential"`

	// Performance metrics
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics,omitempty"`

	// Grading and categorization
	EfficiencyGrade    string `json:"efficiency_grade"`    // A, B, C, D, F
	EfficiencyCategory string `json:"efficiency_category"` // Excellent, Good, Fair, Poor, Critical

	// Recommendations
	Recommendations []string `json:"recommendations,omitempty"`
}

// ResourceWaste represents wasted resources
type ResourceWaste struct {
	WastedCPUHours      float64 `json:"wasted_cpu_hours"`
	WastedMemoryGBHours float64 `json:"wasted_memory_gb_hours"`
	WastedMemoryGB      float64 `json:"wasted_memory_gb"`
	WastedGPUHours      float64 `json:"wasted_gpu_hours,omitempty"`
	TotalWasteCost      float64 `json:"total_waste_cost,omitempty"`
	WastePercentage     float64 `json:"waste_percentage"`
	WasteCost           float64 `json:"waste_cost"`
	IdleTime            float64 `json:"idle_time"`
}

// PerformanceMetrics represents performance-related metrics
type PerformanceMetrics struct {
	Throughput     float64 `json:"throughput"`
	Latency        float64 `json:"latency"`
	QueueTime      float64 `json:"queue_time"`
	ExecutionTime  float64 `json:"execution_time"`
	TurnaroundTime float64 `json:"turnaround_time"`
	SuccessRate    float64 `json:"success_rate"`
	ErrorRate      float64 `json:"error_rate"`
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp       time.Time `json:"timestamp"`
	CPUHours        float64   `json:"cpu_hours"`
	MemoryGB        float64   `json:"memory_gb"`
	GPUHours        float64   `json:"gpu_hours"`
	JobCount        int64     `json:"job_count"`
	UserCount       int64     `json:"user_count"`
	SubmissionCount int64     `json:"submission_count"`
	WaitTime        float64   `json:"wait_time"`
	RunTime         float64   `json:"run_time"`
	CostUSD         float64   `json:"cost_usd"`
}
