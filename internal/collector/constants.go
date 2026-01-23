// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

// Status and state constants used across collectors
const (
	// Result statuses
	StatusSuccess = "success"
	StatusFailed  = "failed"

	// Alert and validation statuses
	StatusActive   = "active"
	StatusResolved = "resolved"

	// Health statuses
	StatusHealthy = "healthy"
	StatusWarning = "warning"

	// Trend and state indicators
	StateNone      = "none"
	StateStable    = "stable"
	StateImproving = "improving"
	StateUnknown   = "unknown"

	// Job states
	JobStateRunning   = "RUNNING"
	JobStateCompleted = "COMPLETED"

	// Default values
	DefaultClusterName = "default"

	// Authentication methods
	AuthMethodJWT = "jwt"

	// Format types
	FormatJSON = "json"
)
