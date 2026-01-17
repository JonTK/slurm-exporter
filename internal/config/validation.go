package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationError represents a configuration validation error with context
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (v ValidationError) Error() string {
	if v.Value != "" {
		return fmt.Sprintf("[%s] %s: '%s' - %s", v.Code, v.Field, v.Value, v.Message)
	}
	return fmt.Sprintf("[%s] %s: %s", v.Code, v.Field, v.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no validation errors"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("configuration validation failed with %d error(s):\n", len(v)))
	for i, err := range v {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
	}
	return sb.String()
}

// ValidateEnhanced performs comprehensive validation of the configuration
func (c *Config) ValidateEnhanced() error {
	var errors ValidationErrors

	// Basic validation first
	if err := c.Validate(); err != nil {
		errors = append(errors, ValidationError{
			Field:   "config",
			Message: err.Error(),
			Code:    "BASIC_VALIDATION_FAILED",
		})
	}

	// Enhanced validations
	errors = append(errors, c.validateSecurity()...)
	errors = append(errors, c.validatePerformance()...)
	errors = append(errors, c.validateBusinessLogic()...)
	errors = append(errors, c.validateCrossField()...)

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateSecurity performs security-related validations
func (c *Config) validateSecurity() ValidationErrors {
	var errors ValidationErrors

	// TLS validation
	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" {
			errors = append(errors, ValidationError{
				Field:   "server.tls.cert_file",
				Message: "certificate file path is required when TLS is enabled",
				Code:    "TLS_CERT_REQUIRED",
			})
		}
		if c.Server.TLS.KeyFile == "" {
			errors = append(errors, ValidationError{
				Field:   "server.tls.key_file",
				Message: "key file path is required when TLS is enabled",
				Code:    "TLS_KEY_REQUIRED",
			})
		}

		// Validate TLS version
		if c.Server.TLS.MinVersion != "" {
			validVersions := []string{"1.0", "1.1", "1.2", "1.3"}
			valid := false
			for _, v := range validVersions {
				if c.Server.TLS.MinVersion == v {
					valid = true
					break
				}
			}
			if !valid {
				errors = append(errors, ValidationError{
					Field:   "server.tls.min_version",
					Value:   c.Server.TLS.MinVersion,
					Message: "invalid TLS version (supported: 1.0, 1.1, 1.2, 1.3)",
					Code:    "INVALID_TLS_VERSION",
				})
			}
		}
	}

	// Basic Auth validation
	if c.Server.BasicAuth.Enabled {
		if c.Server.BasicAuth.Username == "" {
			errors = append(errors, ValidationError{
				Field:   "server.basic_auth.username",
				Message: "username is required when basic auth is enabled",
				Code:    "BASIC_AUTH_USERNAME_REQUIRED",
			})
		}
		if c.Server.BasicAuth.Password == "" {
			errors = append(errors, ValidationError{
				Field:   "server.basic_auth.password",
				Message: "password is required when basic auth is enabled",
				Code:    "BASIC_AUTH_PASSWORD_REQUIRED",
			})
		}

		// Password strength validation
		if len(c.Server.BasicAuth.Password) < 8 {
			errors = append(errors, ValidationError{
				Field:   "server.basic_auth.password",
				Message: "password must be at least 8 characters long for security",
				Code:    "WEAK_PASSWORD",
			})
		}
	}

	// SLURM authentication validation
	if c.SLURM.Auth.Type == "jwt" && c.SLURM.Auth.Token == "" {
		errors = append(errors, ValidationError{
			Field:   "slurm.auth.token",
			Message: "JWT token is required when auth type is 'jwt'",
			Code:    "JWT_TOKEN_REQUIRED",
		})
	}

	// Validate SLURM URL is HTTPS in production-like configurations
	if c.SLURM.BaseURL != "" && !c.Validation.AllowInsecureConnections {
		if u, err := url.Parse(c.SLURM.BaseURL); err == nil {
			if u.Scheme == "http" && !isLocalhost(u.Host) {
				errors = append(errors, ValidationError{
					Field:   "slurm.base_url",
					Value:   c.SLURM.BaseURL,
					Message: "consider using HTTPS for non-localhost SLURM connections for security",
					Code:    "INSECURE_CONNECTION",
				})
			}
		}
	}

	return errors
}

// validatePerformance performs performance-related validations
func (c *Config) validatePerformance() ValidationErrors {
	var errors ValidationErrors

	// Check for potential performance issues
	if c.Collectors.Global.DefaultInterval < 10*time.Second {
		errors = append(errors, ValidationError{
			Field:   "collectors.global.default_interval",
			Value:   c.Collectors.Global.DefaultInterval.String(),
			Message: "very short collection intervals may impact SLURM performance and generate high load",
			Code:    "HIGH_FREQUENCY_COLLECTION",
		})
	}

	// Validate cardinality limits are reasonable
	if c.Metrics.Cardinality.MaxSeries > 1000000 {
		errors = append(errors, ValidationError{
			Field:   "metrics.cardinality.max_series",
			Value:   strconv.Itoa(c.Metrics.Cardinality.MaxSeries),
			Message: "very high cardinality limits may impact memory usage and Prometheus performance",
			Code:    "HIGH_CARDINALITY_LIMIT",
		})
	}

	// Check timeout configurations
	if c.SLURM.Timeout > 5*time.Minute {
		errors = append(errors, ValidationError{
			Field:   "slurm.timeout",
			Value:   c.SLURM.Timeout.String(),
			Message: "very long SLURM timeouts may cause collection delays and timeouts",
			Code:    "LONG_TIMEOUT",
		})
	}

	// Validate server timeouts are reasonable
	if c.Server.ReadTimeout > 2*time.Minute {
		errors = append(errors, ValidationError{
			Field:   "server.read_timeout",
			Value:   c.Server.ReadTimeout.String(),
			Message: "very long read timeouts may impact server performance under load",
			Code:    "LONG_READ_TIMEOUT",
		})
	}

	return errors
}

// validateBusinessLogic performs business logic validations
func (c *Config) validateBusinessLogic() ValidationErrors {
	var errors ValidationErrors

	// Ensure at least one collector is enabled
	enabled := false
	collectors := map[string]bool{
		"jobs":         c.Collectors.Jobs.Enabled,
		"nodes":        c.Collectors.Nodes.Enabled,
		"partitions":   c.Collectors.Partitions.Enabled,
		"cluster":      c.Collectors.Cluster.Enabled,
		"users":        c.Collectors.Users.Enabled,
		"qos":          c.Collectors.QoS.Enabled,
		"reservations": c.Collectors.Reservations.Enabled,
		"accounts":     c.Collectors.Accounts.Enabled,
		"associations": c.Collectors.Associations.Enabled,
		"performance":  c.Collectors.Performance.Enabled,
		"system":       c.Collectors.System.Enabled,
	}

	for _, isEnabled := range collectors {
		if isEnabled {
			enabled = true
			break
		}
	}

	if !enabled {
		errors = append(errors, ValidationError{
			Field:   "collectors",
			Message: "at least one collector must be enabled for the exporter to be useful",
			Code:    "NO_COLLECTORS_ENABLED",
		})
	}

	// Validate custom labels are reasonable
	for collectorName, collector := range map[string]CollectorConfig{
		"jobs":         c.Collectors.Jobs,
		"nodes":        c.Collectors.Nodes,
		"partitions":   c.Collectors.Partitions,
		"cluster":      c.Collectors.Cluster,
		"users":        c.Collectors.Users,
		"qos":          c.Collectors.QoS,
		"reservations": c.Collectors.Reservations,
		"accounts":     c.Collectors.Accounts,
		"associations": c.Collectors.Associations,
		"performance":  c.Collectors.Performance,
		"system":       c.Collectors.System,
	} {
		if collector.Enabled {
			errors = append(errors, validateCollectorLabels(collectorName, collector.Labels)...)
		}
	}

	// Validate metric filter configurations
	errors = append(errors, c.validateMetricFilters()...)

	return errors
}

// validateCrossField performs cross-field validations
func (c *Config) validateCrossField() ValidationErrors {
	var errors ValidationErrors

	// Server timeout should be longer than collection timeout
	maxCollectionTimeout := c.Collectors.Global.DefaultTimeout
	for _, collector := range []CollectorConfig{
		c.Collectors.Jobs, c.Collectors.Nodes, c.Collectors.Partitions,
		c.Collectors.Cluster, c.Collectors.Users, c.Collectors.QoS,
		c.Collectors.Reservations, c.Collectors.Accounts, c.Collectors.Associations,
		c.Collectors.Performance, c.Collectors.System,
	} {
		if collector.Enabled && collector.Timeout > maxCollectionTimeout {
			maxCollectionTimeout = collector.Timeout
		}
	}

	if c.Server.ReadTimeout > 0 && c.Server.ReadTimeout <= maxCollectionTimeout {
		errors = append(errors, ValidationError{
			Field:   "server.read_timeout",
			Value:   c.Server.ReadTimeout.String(),
			Message: fmt.Sprintf("server read timeout should be longer than max collection timeout (%s)", maxCollectionTimeout),
			Code:    "TIMEOUT_MISMATCH",
		})
	}

	// SLURM timeout should be longer than collection timeout
	if c.SLURM.Timeout <= maxCollectionTimeout {
		errors = append(errors, ValidationError{
			Field:   "slurm.timeout",
			Value:   c.SLURM.Timeout.String(),
			Message: fmt.Sprintf("SLURM timeout should be longer than max collection timeout (%s)", maxCollectionTimeout),
			Code:    "SLURM_TIMEOUT_TOO_SHORT",
		})
	}

	// Cardinality warn limit should be less than max limit
	if c.Metrics.Cardinality.WarnLimit >= c.Metrics.Cardinality.MaxSeries {
		errors = append(errors, ValidationError{
			Field:   "metrics.cardinality.warn_limit",
			Value:   strconv.Itoa(c.Metrics.Cardinality.WarnLimit),
			Message: "warn limit should be less than max series limit",
			Code:    "INVALID_CARDINALITY_LIMITS",
		})
	}

	return errors
}

// validateCollectorLabels validates custom labels for a collector
func validateCollectorLabels(collectorName string, labels map[string]string) ValidationErrors {
	var errors ValidationErrors

	if len(labels) > 20 {
		errors = append(errors, ValidationError{
			Field:   fmt.Sprintf("collectors.%s.labels", collectorName),
			Value:   strconv.Itoa(len(labels)),
			Message: "too many custom labels may impact performance (recommended: < 20)",
			Code:    "TOO_MANY_LABELS",
		})
	}

	// Validate label names and values
	labelNameRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	for name, value := range labels {
		if !labelNameRegex.MatchString(name) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("collectors.%s.labels.%s", collectorName, name),
				Value:   name,
				Message: "label name must match [a-zA-Z_][a-zA-Z0-9_]* (Prometheus requirements)",
				Code:    "INVALID_LABEL_NAME",
			})
		}

		if len(value) > 256 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("collectors.%s.labels.%s", collectorName, name),
				Value:   value,
				Message: "label value is very long, may impact performance",
				Code:    "LONG_LABEL_VALUE",
			})
		}

		// Check for reserved Prometheus label names
		reservedLabels := []string{"__name__", "__address__", "__scheme__", "__metrics_path__", "__param_", "__tmp_"}
		for _, reserved := range reservedLabels {
			if strings.HasPrefix(name, reserved) {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("collectors.%s.labels.%s", collectorName, name),
					Value:   name,
					Message: "label name conflicts with Prometheus reserved prefix",
					Code:    "RESERVED_LABEL_NAME",
				})
			}
		}
	}

	return errors
}

// validateMetricFilters validates metric filter configurations
func (c *Config) validateMetricFilters() ValidationErrors {
	var errors ValidationErrors

	collectors := map[string]CollectorConfig{
		"jobs":         c.Collectors.Jobs,
		"nodes":        c.Collectors.Nodes,
		"partitions":   c.Collectors.Partitions,
		"cluster":      c.Collectors.Cluster,
		"users":        c.Collectors.Users,
		"qos":          c.Collectors.QoS,
		"reservations": c.Collectors.Reservations,
		"accounts":     c.Collectors.Accounts,
		"associations": c.Collectors.Associations,
		"performance":  c.Collectors.Performance,
		"system":       c.Collectors.System,
	}

	for collectorName, collector := range collectors {
		if !collector.Enabled {
			continue
		}

		// Check for conflicting filter settings
		filters := collector.Filters.Metrics
		if !filters.EnableAll && len(filters.IncludeMetrics) == 0 && len(filters.ExcludeMetrics) == 0 {
			// Check if any type-specific filters are enabled
			hasTypeFilters := filters.OnlyInfo || filters.OnlyCounters || filters.OnlyGauges ||
				filters.OnlyHistograms || filters.SkipHistograms || filters.SkipTimingMetrics ||
				filters.SkipResourceMetrics

			if !hasTypeFilters {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("collectors.%s.filters.metrics", collectorName),
					Message: "no metrics will be collected with current filter configuration",
					Code:    "NO_METRICS_FILTERED",
				})
			}
		}

		// Check for mutually exclusive filters
		exclusiveFilters := []bool{filters.OnlyInfo, filters.OnlyCounters, filters.OnlyGauges, filters.OnlyHistograms}
		enabledCount := 0
		for _, enabled := range exclusiveFilters {
			if enabled {
				enabledCount++
			}
		}
		if enabledCount > 1 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("collectors.%s.filters.metrics", collectorName),
				Message: "only one 'only_*' filter can be enabled at a time",
				Code:    "CONFLICTING_FILTERS",
			})
		}

		// Validate metric name patterns
		for _, pattern := range filters.IncludeMetrics {
			if err := validateMetricPattern(pattern); err != nil {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("collectors.%s.filters.metrics.include_metrics", collectorName),
					Value:   pattern,
					Message: err.Error(),
					Code:    "INVALID_METRIC_PATTERN",
				})
			}
		}

		for _, pattern := range filters.ExcludeMetrics {
			if err := validateMetricPattern(pattern); err != nil {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("collectors.%s.filters.metrics.exclude_metrics", collectorName),
					Value:   pattern,
					Message: err.Error(),
					Code:    "INVALID_METRIC_PATTERN",
				})
			}
		}
	}

	return errors
}

// validateMetricPattern validates a metric name pattern
func validateMetricPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("metric pattern cannot be empty")
	}

	// Allow wildcards but check for invalid characters
	validChars := regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:*]*$`)
	if !validChars.MatchString(pattern) {
		return fmt.Errorf("metric pattern contains invalid characters (allowed: a-z, A-Z, 0-9, _, :, *)")
	}

	return nil
}

// isLocalhost checks if the host is localhost or a local IP
func isLocalhost(host string) bool {
	localHosts := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0"}
	// Remove port if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex > 0 {
		host = host[:colonIndex]
	}

	for _, local := range localHosts {
		if host == local {
			return true
		}
	}
	return false
}
