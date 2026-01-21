package metrics

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// LabelManager provides advanced label management and dimensional analysis
type LabelManager struct {
	// Known label dimensions and their allowed values
	labelDimensions map[string]*LabelDimension
	// Cardinality tracking
	cardinalityLimits map[string]int
	cardinalityStats  map[string]int
	// Label validation rules
	validationRules map[string]*regexp.Regexp
}

// LabelDimension represents a metric label dimension with metadata
type LabelDimension struct {
	Name           string
	Description    string
	Type           LabelType
	AllowedValues  []string
	MaxCardinality int
	Required       bool
	Sensitive      bool // For PII or sensitive data
}

// LabelType represents the type of a label dimension
type LabelType string

const (
	LabelTypeCluster   LabelType = "cluster"
	LabelTypeNode      LabelType = "node"
	LabelTypeJob       LabelType = "job"
	LabelTypeUser      LabelType = "user"
	LabelTypeAccount   LabelType = "account"
	LabelTypePartition LabelType = "partition"
	LabelTypeMetric    LabelType = "metric"
	LabelTypeState     LabelType = "state"
	LabelTypeResource  LabelType = "resource"
	LabelTypeError     LabelType = "error"
	LabelTypeTemporal  LabelType = "temporal"
	LabelTypeSystem    LabelType = "system"
)

// StandardLabels defines the standard label dimensions used across SLURM metrics
var StandardLabels = map[string]*LabelDimension{
	"cluster_name": {
		Name:           "cluster_name",
		Description:    "Name of the SLURM cluster",
		Type:           LabelTypeCluster,
		MaxCardinality: 10,
		Required:       true,
		Sensitive:      false,
	},
	"node_name": {
		Name:           "node_name",
		Description:    "Name of the compute node",
		Type:           LabelTypeNode,
		MaxCardinality: 10000,
		Required:       false,
		Sensitive:      false,
	},
	"job_id": {
		Name:           "job_id",
		Description:    "Unique job identifier",
		Type:           LabelTypeJob,
		MaxCardinality: 1000000,
		Required:       false,
		Sensitive:      false,
	},
	"user": {
		Name:           "user",
		Description:    "Username of job owner",
		Type:           LabelTypeUser,
		MaxCardinality: 10000,
		Required:       false,
		Sensitive:      true,
	},
	"account": {
		Name:           "account",
		Description:    "Account name for billing",
		Type:           LabelTypeAccount,
		MaxCardinality: 1000,
		Required:       false,
		Sensitive:      false,
	},
	"partition": {
		Name:           "partition",
		Description:    "SLURM partition name",
		Type:           LabelTypePartition,
		MaxCardinality: 100,
		Required:       false,
		Sensitive:      false,
	},
	"job_state": {
		Name:           "job_state",
		Description:    "Current state of the job",
		Type:           LabelTypeState,
		AllowedValues:  []string{"pending", "running", "suspended", "completed", "cancelled", "failed", "timeout", "node_fail"},
		MaxCardinality: 10,
		Required:       false,
		Sensitive:      false,
	},
	"node_state": {
		Name:           "node_state",
		Description:    "Current state of the node",
		Type:           LabelTypeState,
		AllowedValues:  []string{"idle", "allocated", "mixed", "down", "drain", "draining", "drained", "fail", "failing", "future", "maint", "perfctrs", "reserved", "unknown"},
		MaxCardinality: 15,
		Required:       false,
		Sensitive:      false,
	},
	"partition_state": {
		Name:           "partition_state",
		Description:    "Current state of the partition",
		Type:           LabelTypeState,
		AllowedValues:  []string{"UP", "DOWN", "DRAIN", "INACTIVE"},
		MaxCardinality: 5,
		Required:       false,
		Sensitive:      false,
	},
	"resource_type": {
		Name:           "resource_type",
		Description:    "Type of compute resource",
		Type:           LabelTypeResource,
		AllowedValues:  []string{"cpu", "memory", "storage", "network", "gpu", "nodes"},
		MaxCardinality: 10,
		Required:       false,
		Sensitive:      false,
	},
	"error_type": {
		Name:           "error_type",
		Description:    "Type of error encountered",
		Type:           LabelTypeError,
		AllowedValues:  []string{"timeout", "connection_error", "auth_error", "api_error", "parse_error", "rate_limit", "internal_error"},
		MaxCardinality: 20,
		Required:       false,
		Sensitive:      false,
	},
	"collector": {
		Name:           "collector",
		Description:    "Name of the metrics collector",
		Type:           LabelTypeSystem,
		AllowedValues:  []string{"cluster", "node", "job", "user", "partition", "performance", "selfmon"},
		MaxCardinality: 10,
		Required:       false,
		Sensitive:      false,
	},
	"metric_type": {
		Name:           "metric_type",
		Description:    "Type of metric being collected",
		Type:           LabelTypeMetric,
		AllowedValues:  []string{"jobs_completed", "jobs_submitted", "cpu_hours_delivered", "throughput", "efficiency", "utilization"},
		MaxCardinality: 50,
		Required:       false,
		Sensitive:      false,
	},
	"time_window": {
		Name:           "time_window",
		Description:    "Time window for aggregated metrics",
		Type:           LabelTypeTemporal,
		AllowedValues:  []string{"1m", "5m", "15m", "1h", "6h", "24h"},
		MaxCardinality: 10,
		Required:       false,
		Sensitive:      false,
	},
	"priority": {
		Name:           "priority",
		Description:    "Priority level for scheduling",
		Type:           LabelTypeSystem,
		AllowedValues:  []string{"urgent", "high", "normal", "low"},
		MaxCardinality: 5,
		Required:       false,
		Sensitive:      false,
	},
	"statistic": {
		Name:           "statistic",
		Description:    "Statistical measure type",
		Type:           LabelTypeMetric,
		AllowedValues:  []string{"average", "median", "p95", "p99", "max", "min"},
		MaxCardinality: 10,
		Required:       false,
		Sensitive:      false,
	},
	"entity_name": {
		Name:           "entity_name",
		Description:    "Name of the entity being monitored",
		Type:           LabelTypeSystem,
		MaxCardinality: 100000,
		Required:       false,
		Sensitive:      false,
	},
}

// NewLabelManager creates a new label manager with dimensional analysis
func NewLabelManager() *LabelManager {
	lm := &LabelManager{
		labelDimensions:   make(map[string]*LabelDimension),
		cardinalityLimits: make(map[string]int),
		cardinalityStats:  make(map[string]int),
		validationRules:   make(map[string]*regexp.Regexp),
	}

	// Initialize with standard labels
	for name, dimension := range StandardLabels {
		lm.labelDimensions[name] = dimension
		lm.cardinalityLimits[name] = dimension.MaxCardinality
	}

	// Set up validation rules
	lm.setupValidationRules()

	return lm
}

// setupValidationRules initializes label validation patterns
func (lm *LabelManager) setupValidationRules() {
	lm.validationRules["cluster_name"] = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_\s]{0,62}$`)
	lm.validationRules["node_name"] = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_\.]{0,253}$`)
	lm.validationRules["job_id"] = regexp.MustCompile(`^[0-9]+(_[0-9]+)?(\.[0-9]+)?$`)
	lm.validationRules["user"] = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_\.]{0,31}$`)
	lm.validationRules["account"] = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_]{0,63}$`)
	lm.validationRules["partition"] = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_]{0,63}$`)
}

// ValidateLabels validates a set of label values against their dimensions
func (lm *LabelManager) ValidateLabels(labels map[string]string) error {
	for labelName, labelValue := range labels {
		if err := lm.ValidateLabel(labelName, labelValue); err != nil {
			return fmt.Errorf("label %s: %w", labelName, err)
		}
	}
	return nil
}

// ValidateLabel validates a single label value
func (lm *LabelManager) ValidateLabel(labelName, labelValue string) error {
	dimension, exists := lm.labelDimensions[labelName]
	if !exists {
		// Allow unknown labels for flexibility
		return nil
	}

	// Check if value is in allowed values list
	if len(dimension.AllowedValues) > 0 {
		if !lm.isValueAllowed(labelValue, dimension.AllowedValues) {
			return fmt.Errorf("value '%s' not in allowed values %v", labelValue, dimension.AllowedValues)
		}
	}

	// Apply validation regex if available
	if rule, exists := lm.validationRules[labelName]; exists {
		if !rule.MatchString(labelValue) {
			return fmt.Errorf("value '%s' does not match validation pattern", labelValue)
		}
	}

	// Check cardinality limits
	if err := lm.checkCardinality(labelName, labelValue); err != nil {
		return err
	}

	return nil
}

// isValueAllowed checks if a value is in the allowed values list
func (lm *LabelManager) isValueAllowed(value string, allowedValues []string) bool {
	for _, allowed := range allowedValues {
		if value == allowed {
			return true
		}
	}
	return false
}

// checkCardinality verifies that adding this label value won't exceed cardinality limits
func (lm *LabelManager) checkCardinality(labelName, labelValue string) error {
	limit, exists := lm.cardinalityLimits[labelName]
	if !exists {
		return nil
	}

	// For simulation, we'll accept any value but warn on high cardinality
	if lm.cardinalityStats[labelName] >= limit {
		return fmt.Errorf("cardinality limit exceeded for label %s (limit: %d)", labelName, limit)
	}

	return nil
}

// TrackLabelUsage tracks the usage of a label value for cardinality analysis
func (lm *LabelManager) TrackLabelUsage(labelName, labelValue string) {
	// In a real implementation, this would track unique values
	// For now, we'll just increment the counter
	lm.cardinalityStats[labelName]++
}

// GetDimensionalAnalysis returns cardinality analysis for all labels
func (lm *LabelManager) GetDimensionalAnalysis() map[string]CardinalityInfo {
	analysis := make(map[string]CardinalityInfo)

	for labelName, dimension := range lm.labelDimensions {
		currentCardinality := lm.cardinalityStats[labelName]
		utilizationPct := float64(currentCardinality) / float64(dimension.MaxCardinality) * 100

		analysis[labelName] = CardinalityInfo{
			LabelName:          labelName,
			CurrentCardinality: currentCardinality,
			MaxCardinality:     dimension.MaxCardinality,
			UtilizationPercent: utilizationPct,
			LabelType:          dimension.Type,
			Required:           dimension.Required,
			Sensitive:          dimension.Sensitive,
		}
	}

	return analysis
}

// CardinalityInfo represents cardinality statistics for a label dimension
type CardinalityInfo struct {
	LabelName          string
	CurrentCardinality int
	MaxCardinality     int
	UtilizationPercent float64
	LabelType          LabelType
	Required           bool
	Sensitive          bool
}

// GetHighCardinalityLabels returns labels that exceed a utilization threshold
func (lm *LabelManager) GetHighCardinalityLabels(thresholdPercent float64) []CardinalityInfo {
	var highCardinality []CardinalityInfo

	analysis := lm.GetDimensionalAnalysis()
	for _, info := range analysis {
		if info.UtilizationPercent > thresholdPercent {
			highCardinality = append(highCardinality, info)
		}
	}

	// Sort by utilization percentage descending
	sort.Slice(highCardinality, func(i, j int) bool {
		return highCardinality[i].UtilizationPercent > highCardinality[j].UtilizationPercent
	})

	return highCardinality
}

// SanitizeLabels sanitizes label values to ensure they're safe for Prometheus
func (lm *LabelManager) SanitizeLabels(labels map[string]string) map[string]string {
	sanitized := make(map[string]string)

	for labelName, labelValue := range labels {
		// Handle sensitive labels
		if dimension, exists := lm.labelDimensions[labelName]; exists && dimension.Sensitive {
			sanitized[labelName] = lm.sanitizeSensitiveValue(labelValue)
		} else {
			sanitized[labelName] = lm.sanitizeValue(labelValue)
		}
	}

	return sanitized
}

// sanitizeValue sanitizes a label value for Prometheus compatibility
func (lm *LabelManager) sanitizeValue(value string) string {
	// Remove or replace invalid characters
	sanitized := strings.ReplaceAll(value, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ToLower(sanitized)

	// For job_id, numbers are allowed to start the value
	return sanitized
}

// sanitizeSensitiveValue sanitizes sensitive label values (e.g., usernames)
func (lm *LabelManager) sanitizeSensitiveValue(value string) string {
	// For sensitive values, we might hash or anonymize them
	// For now, just sanitize normally but could be enhanced
	return lm.sanitizeValue(value)
}

// TODO: Following helper function is unused - preserved for future character validation
/*
// isLetter checks if a character is a letter
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}
*/

// GetLabelRecommendations provides recommendations for label optimization
func (lm *LabelManager) GetLabelRecommendations() []LabelRecommendation {
	var recommendations []LabelRecommendation

	analysis := lm.GetDimensionalAnalysis()

	for labelName, info := range analysis {
		if info.UtilizationPercent > 80 {
			recommendations = append(recommendations, LabelRecommendation{
				LabelName:  labelName,
				Type:       RecommendationTypeCardinality,
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("Label '%s' is at %.1f%% cardinality utilization", labelName, info.UtilizationPercent),
				Suggestion: "Consider reducing label cardinality or increasing limits",
			})
		}

		if info.UtilizationPercent > 95 {
			recommendations = append(recommendations, LabelRecommendation{
				LabelName:  labelName,
				Type:       RecommendationTypeCardinality,
				Severity:   SeverityCritical,
				Message:    fmt.Sprintf("Label '%s' is at critical cardinality utilization (%.1f%%)", labelName, info.UtilizationPercent),
				Suggestion: "Immediate action required to prevent metric explosion",
			})
		}

		if info.Sensitive && info.CurrentCardinality > 100 {
			recommendations = append(recommendations, LabelRecommendation{
				LabelName:  labelName,
				Type:       RecommendationTypeSecurity,
				Severity:   SeverityInfo,
				Message:    fmt.Sprintf("Sensitive label '%s' has high cardinality", labelName),
				Suggestion: "Consider anonymizing or aggregating sensitive label values",
			})
		}
	}

	return recommendations
}

// LabelRecommendation represents a recommendation for label optimization
type LabelRecommendation struct {
	LabelName  string
	Type       RecommendationType
	Severity   Severity
	Message    string
	Suggestion string
}

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeCardinality RecommendationType = "cardinality"
	RecommendationTypeSecurity    RecommendationType = "security"
	RecommendationTypePerformance RecommendationType = "performance"
)

// Severity represents the severity level of a recommendation
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// BuildLabelSet creates a validated and sanitized label set
func (lm *LabelManager) BuildLabelSet(labels map[string]string) (map[string]string, error) {
	// Validate labels
	if err := lm.ValidateLabels(labels); err != nil {
		return nil, fmt.Errorf("label validation failed: %w", err)
	}

	// Sanitize labels
	sanitized := lm.SanitizeLabels(labels)

	// Track usage for cardinality analysis
	for labelName, labelValue := range sanitized {
		lm.TrackLabelUsage(labelName, labelValue)
	}

	return sanitized, nil
}

// ExportDimensionalAnalysis exports the current dimensional analysis as a formatted report
func (lm *LabelManager) ExportDimensionalAnalysis() string {
	var report strings.Builder

	report.WriteString("SLURM Exporter Dimensional Analysis Report\n")
	report.WriteString("========================================\n\n")

	analysis := lm.GetDimensionalAnalysis()

	// Sort by utilization percentage
	var sorted []CardinalityInfo
	for _, info := range analysis {
		sorted = append(sorted, info)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UtilizationPercent > sorted[j].UtilizationPercent
	})

	report.WriteString("Label Cardinality Summary:\n")
	report.WriteString("--------------------------\n")
	for _, info := range sorted {
		status := "OK"
		if info.UtilizationPercent > 80 {
			status = "WARNING"
		}
		if info.UtilizationPercent > 95 {
			status = "CRITICAL"
		}

		report.WriteString(fmt.Sprintf("%-20s: %6d/%6d (%5.1f%%) [%s] %s\n",
			info.LabelName,
			info.CurrentCardinality,
			info.MaxCardinality,
			info.UtilizationPercent,
			info.LabelType,
			status,
		))
	}

	report.WriteString("\nRecommendations:\n")
	report.WriteString("----------------\n")
	recommendations := lm.GetLabelRecommendations()
	if len(recommendations) == 0 {
		report.WriteString("No recommendations at this time.\n")
	} else {
		for _, rec := range recommendations {
			report.WriteString(fmt.Sprintf("• [%s] %s: %s\n",
				strings.ToUpper(string(rec.Severity)),
				rec.LabelName,
				rec.Message,
			))
			report.WriteString(fmt.Sprintf("  Suggestion: %s\n\n", rec.Suggestion))
		}
	}

	return report.String()
}
