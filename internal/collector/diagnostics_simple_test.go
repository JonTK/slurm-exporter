// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-client/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestDiagnosticsCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	collector := NewDiagnosticsCollector(mockClient, logger, timeout)

	ch := make(chan *prometheus.Desc, 100)
	collector.Describe(ch)
	close(ch)

	// Should have at least the basic metrics
	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	assert.True(t, len(descs) > 0, "should have metric descriptors")
}

func TestDiagnosticsCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	timeout := 30 * time.Second

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	// Setup mock expectations with test data
	now := time.Now()
	diagnostics := &interfaces.Diagnostics{
		DataCollected:        now,
		ReqTime:              12345,
		ReqTimeStart:         now.Unix(),
		ServerThreadCount:    10,
		AgentCount:           5,
		AgentThreadCount:     20,
		DBDAgentCount:        2,
		JobsSubmitted:        1000,
		JobsStarted:          950,
		JobsCompleted:        900,
		JobsCanceled:         30,
		JobsFailed:           20,
		ScheduleCycleMax:     5000,
		ScheduleCycleLast:    3000,
		ScheduleCycleTotal:   150000,
		ScheduleCycleCounter: 50,
		ScheduleCycleMean:    3000.0,
		BackfillCycleMax:     10000,
		BackfillCycleLast:    8000,
		BackfillCycleTotal:   400000,
		BackfillCycleCounter: 50,
		BackfillCycleMean:    8000.0,
		BfBackfilledJobs:     100,
		BfLastBackfilledJobs: 10,
		BfCycleSum:           50000,
		BfQueueLen:           5,
		BfQueueLenSum:        250,
		BfWhenLastCycle:      now.Unix(),
		BfActive:             true,
		RPCsByMessageType: map[string]int{
			"REQUEST_JOB_INFO":  100,
			"REQUEST_NODE_INFO": 50,
		},
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("GetDiagnostics", mock.Anything).Return(diagnostics, nil)

	collector := NewDiagnosticsCollector(mockClient, logger, timeout)

	// Collect metrics
	ch := make(chan prometheus.Metric, 200)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Count metrics
	count := 0
	for range ch {
		count++
	}

	assert.True(t, count > 0, "should have collected metrics")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
}

func TestDiagnosticsCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	timeout := 30 * time.Second

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	// Setup mock to return error
	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("GetDiagnostics", mock.Anything).Return(nil, assert.AnError)

	collector := NewDiagnosticsCollector(mockClient, logger, timeout)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
}
