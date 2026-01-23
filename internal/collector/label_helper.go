// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/metrics"
)

// LabelHelper provides label validation and management for collectors
type LabelHelper struct {
	labelManager *metrics.LabelManager
	logger       *logrus.Entry
}

// NewLabelHelper creates a new label helper instance
func NewLabelHelper(labelManager *metrics.LabelManager, logger *logrus.Entry) *LabelHelper {
	return &LabelHelper{
		labelManager: labelManager,
		logger:       logger,
	}
}

// ValidateAndBuildMetric validates labels and builds a Prometheus metric
func (lh *LabelHelper) ValidateAndBuildMetric(
	desc *prometheus.Desc,
	valueType prometheus.ValueType,
	value float64,
	labelNames []string,
	labelValues ...string,
) (prometheus.Metric, error) {
	// Build label map for validation
	if len(labelNames) != len(labelValues) {
		return nil, fmt.Errorf("label names and values count mismatch: %d names, %d values", len(labelNames), len(labelValues))
	}

	labelMap := make(map[string]string)
	for i, name := range labelNames {
		if i < len(labelValues) {
			labelMap[name] = labelValues[i]
		}
	}

	// Validate labels
	validatedLabels, err := lh.labelManager.BuildLabelSet(labelMap)
	if err != nil {
		lh.logger.WithError(err).WithFields(logrus.Fields{
			"labels": labelMap,
		}).Warn("Label validation failed, using original values")

		// Continue with original values if validation fails
		validatedLabels = labelMap
	}

	// Convert back to slice format for Prometheus
	validatedValues := make([]string, len(labelNames))
	for i, name := range labelNames {
		if value, exists := validatedLabels[name]; exists {
			validatedValues[i] = value
		} else {
			validatedValues[i] = labelValues[i]
		}
	}

	// Build the metric
	metric, err := prometheus.NewConstMetric(desc, valueType, value, validatedValues...)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric: %w", err)
	}

	return metric, nil
}

// BuildMetricWithValidation is a convenience method that extends BaseCollector.BuildMetric with validation
func (b *BaseCollector) BuildMetricWithValidation(
	desc *prometheus.Desc,
	valueType prometheus.ValueType,
	value float64,
	labelValues ...string,
) prometheus.Metric {
	// For now, use the existing BuildMetric method
	// In a full implementation, we would extract label names from desc and validate
	return b.BuildMetric(desc, valueType, value, labelValues...)
}

// GetLabelNames extracts label names from a Prometheus descriptor
func GetLabelNames(desc *prometheus.Desc) []string {
	// This is a helper to extract label names from prometheus.Desc
	// In practice, this would need to use reflection or store label names separately
	// For now, return common SLURM label names
	return []string{"cluster_name", "instance"}
}

// StandardLabelSets provides pre-built label sets for common scenarios
type StandardLabelSets struct {
	labelManager *metrics.LabelManager
}

// NewStandardLabelSets creates standard label sets helper
func NewStandardLabelSets(labelManager *metrics.LabelManager) *StandardLabelSets {
	return &StandardLabelSets{
		labelManager: labelManager,
	}
}

// ClusterLabels builds standard cluster-level labels
func (sls *StandardLabelSets) ClusterLabels(clusterName string) (map[string]string, error) {
	labels := map[string]string{
		"cluster_name": clusterName,
	}
	return sls.labelManager.BuildLabelSet(labels)
}

// NodeLabels builds standard node-level labels
func (sls *StandardLabelSets) NodeLabels(clusterName, nodeName string) (map[string]string, error) {
	labels := map[string]string{
		"cluster_name": clusterName,
		"node_name":    nodeName,
	}
	return sls.labelManager.BuildLabelSet(labels)
}

// JobLabels builds standard job-level labels
func (sls *StandardLabelSets) JobLabels(clusterName, jobID, user, account, partition string) (map[string]string, error) {
	labels := map[string]string{
		"cluster_name": clusterName,
		"job_id":       jobID,
		"user":         user,
		"account":      account,
		"partition":    partition,
	}
	return sls.labelManager.BuildLabelSet(labels)
}

// PartitionLabels builds standard partition-level labels
func (sls *StandardLabelSets) PartitionLabels(clusterName, partitionName string) (map[string]string, error) {
	labels := map[string]string{
		"cluster_name": clusterName,
		"partition":    partitionName,
	}
	return sls.labelManager.BuildLabelSet(labels)
}

// UserLabels builds standard user-level labels
func (sls *StandardLabelSets) UserLabels(clusterName, user, account string) (map[string]string, error) {
	labels := map[string]string{
		"cluster_name": clusterName,
		"user":         user,
		"account":      account,
	}
	return sls.labelManager.BuildLabelSet(labels)
}

// StateLabels builds standard state-related labels
func (sls *StandardLabelSets) StateLabels(clusterName, entityName, state string) (map[string]string, error) {
	labels := map[string]string{
		"cluster_name": clusterName,
		"entity_name":  entityName,
		"job_state":    state, // Use a specific state type that exists in StandardLabels
	}
	return sls.labelManager.BuildLabelSet(labels)
}

// LabelSetBuilder provides a fluent interface for building label sets
type LabelSetBuilder struct {
	labels       map[string]string
	labelManager *metrics.LabelManager
}

// NewLabelSetBuilder creates a new label set builder
func NewLabelSetBuilder(labelManager *metrics.LabelManager) *LabelSetBuilder {
	return &LabelSetBuilder{
		labels:       make(map[string]string),
		labelManager: labelManager,
	}
}

// WithCluster adds cluster label
func (lsb *LabelSetBuilder) WithCluster(clusterName string) *LabelSetBuilder {
	lsb.labels["cluster_name"] = clusterName
	return lsb
}

// WithNode adds node label
func (lsb *LabelSetBuilder) WithNode(nodeName string) *LabelSetBuilder {
	lsb.labels["node_name"] = nodeName
	return lsb
}

// WithJob adds job-related labels
func (lsb *LabelSetBuilder) WithJob(jobID, user, account, partition string) *LabelSetBuilder {
	lsb.labels["job_id"] = jobID
	lsb.labels["user"] = user
	lsb.labels["account"] = account
	lsb.labels["partition"] = partition
	return lsb
}

// WithPartition adds partition label
func (lsb *LabelSetBuilder) WithPartition(partitionName string) *LabelSetBuilder {
	lsb.labels["partition"] = partitionName
	return lsb
}

// WithState adds state label
func (lsb *LabelSetBuilder) WithState(stateType, state string) *LabelSetBuilder {
	lsb.labels[stateType+"_state"] = state
	return lsb
}

// WithResource adds resource type label
func (lsb *LabelSetBuilder) WithResource(resourceType string) *LabelSetBuilder {
	lsb.labels["resource_type"] = resourceType
	return lsb
}

// WithError adds error type label
func (lsb *LabelSetBuilder) WithError(errorType string) *LabelSetBuilder {
	lsb.labels["error_type"] = errorType
	return lsb
}

// WithMetricType adds metric type label
func (lsb *LabelSetBuilder) WithMetricType(metricType string) *LabelSetBuilder {
	lsb.labels["metric_type"] = metricType
	return lsb
}

// WithTimeWindow adds time window label
func (lsb *LabelSetBuilder) WithTimeWindow(timeWindow string) *LabelSetBuilder {
	lsb.labels["time_window"] = timeWindow
	return lsb
}

// WithCustom adds a custom label
func (lsb *LabelSetBuilder) WithCustom(key, value string) *LabelSetBuilder {
	lsb.labels[key] = value
	return lsb
}

// Build validates and builds the final label set
func (lsb *LabelSetBuilder) Build() (map[string]string, error) {
	return lsb.labelManager.BuildLabelSet(lsb.labels)
}

// BuildSlice builds the label set and returns it as ordered slices
func (lsb *LabelSetBuilder) BuildSlice() ([]string, []string, error) {
	labelSet, err := lsb.Build()
	if err != nil {
		return nil, nil, err
	}

	var names, values []string
	for name, value := range labelSet {
		names = append(names, name)
		values = append(values, value)
	}

	return names, values, nil
}

// DimensionalAnalysisReporter provides reporting for dimensional analysis
type DimensionalAnalysisReporter struct {
	labelManager *metrics.LabelManager
	logger       *logrus.Entry
}

// NewDimensionalAnalysisReporter creates a new dimensional analysis reporter
func NewDimensionalAnalysisReporter(labelManager *metrics.LabelManager, logger *logrus.Entry) *DimensionalAnalysisReporter {
	return &DimensionalAnalysisReporter{
		labelManager: labelManager,
		logger:       logger,
	}
}

// LogCardinalityAnalysis logs the current cardinality analysis
func (dar *DimensionalAnalysisReporter) LogCardinalityAnalysis() {
	analysis := dar.labelManager.GetDimensionalAnalysis()

	dar.logger.Info("Dimensional Analysis Summary:")
	for labelName, info := range analysis {
		dar.logger.WithFields(logrus.Fields{
			"label":       labelName,
			"cardinality": info.CurrentCardinality,
			"max":         info.MaxCardinality,
			"utilization": fmt.Sprintf("%.1f%%", info.UtilizationPercent),
			"type":        info.LabelType,
			"required":    info.Required,
			"sensitive":   info.Sensitive,
		}).Info("Label dimension analysis")
	}
}

// LogHighCardinalityWarnings logs warnings for high cardinality labels
func (dar *DimensionalAnalysisReporter) LogHighCardinalityWarnings(thresholdPercent float64) {
	highCardinality := dar.labelManager.GetHighCardinalityLabels(thresholdPercent)

	if len(highCardinality) == 0 {
		dar.logger.Info("No high cardinality labels detected")
		return
	}

	dar.logger.WithField("threshold", fmt.Sprintf("%.1f%%", thresholdPercent)).
		Warn("High cardinality labels detected")

	for _, info := range highCardinality {
		dar.logger.WithFields(logrus.Fields{
			"label":       info.LabelName,
			"utilization": fmt.Sprintf("%.1f%%", info.UtilizationPercent),
			"cardinality": info.CurrentCardinality,
			"max":         info.MaxCardinality,
		}).Warn("Label exceeds cardinality threshold")
	}
}

// LogRecommendations logs label optimization recommendations
func (dar *DimensionalAnalysisReporter) LogRecommendations() {
	recommendations := dar.labelManager.GetLabelRecommendations()

	if len(recommendations) == 0 {
		dar.logger.Info("No label optimization recommendations at this time")
		return
	}

	dar.logger.Info("Label optimization recommendations:")
	for _, rec := range recommendations {
		dar.logger.WithFields(logrus.Fields{
			"label":      rec.LabelName,
			"type":       rec.Type,
			"severity":   rec.Severity,
			"message":    rec.Message,
			"suggestion": rec.Suggestion,
		}).Info("Label recommendation")
	}
}
