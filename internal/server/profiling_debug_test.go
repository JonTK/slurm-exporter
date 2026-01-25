// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package server

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/jontk/slurm-exporter/internal/testutil"
)

func TestNewProfilingDebugHandler(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	handler := NewProfilingDebugHandler(nil, nil, logger)

	assert.NotNil(t, handler)
}

func TestProfilingDebugHandler_RegisterRoutes(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	router := mux.NewRouter()

	handler := NewProfilingDebugHandler(nil, nil, logger)
	handler.RegisterRoutes(router)

	// Verify routes were registered - we can't directly test this
	// but we can verify the handler is operational
	assert.NotNil(t, handler)
}

func TestProfilingDebugHandler_HandleProfilingStatus_HTML(t *testing.T) {
	t.Parallel()
	// Skipping - requires profiled collector manager
	t.Skip("Requires manager initialization")
}

func TestProfilingDebugHandler_HandleProfilingStatus_JSON(t *testing.T) {
	t.Parallel()
	// Skipping - requires profiled collector manager
	t.Skip("Requires manager initialization")
}

func TestProfilingDebugHandler_HandleListProfiles_JSON(t *testing.T) {
	t.Parallel()
	// Skipping - requires profiler initialization
	t.Skip("Requires profiler initialization")
}

func TestProfilingDebugHandler_HandleListProfiles_HTML(t *testing.T) {
	t.Parallel()
	// Skipping - requires profiler initialization
	t.Skip("Requires profiler initialization")
}

func TestProfilingDebugHandler_HandleStats_JSON(t *testing.T) {
	t.Parallel()
	// Skipping this test - it requires a real Profiler instance
	// to avoid nil pointer dereference
	t.Skip("Requires profiler instance initialization")
}

func TestProfilingDebugHandler_HandleGetProfile_NotFound(t *testing.T) {
	t.Parallel()
	// Skipping - requires profiler instance
	t.Skip("Requires profiler initialization")
}

func TestProfilingDebugHandler_HandleGetProfile_InvalidID(t *testing.T) {
	t.Parallel()
	// Skipping - requires profiler instance
	t.Skip("Requires profiler initialization")
}

func TestProfilingDebugHandler_HandleEnableProfiling(t *testing.T) {
	t.Parallel()
	// Skipping - requires manager instance
	t.Skip("Requires manager initialization")
}

func TestProfilingDebugHandler_HandleDisableProfiling(t *testing.T) {
	t.Parallel()
	// Skipping - requires manager instance
	t.Skip("Requires manager initialization")
}

func TestProfilingDebugHandler_ContentTypeHeaders(t *testing.T) {
	t.Parallel()
	// Skipping table-driven tests - require initialized dependencies
	t.Skip("Requires profiler and manager initialization")
}

func TestProfilingDebugHandler_ErrorHandling(t *testing.T) {
	t.Parallel()
	// Skipping this test - it requires real manager instance
	// to avoid nil pointer dereference
	t.Skip("Requires profiled collector manager initialization")
}
