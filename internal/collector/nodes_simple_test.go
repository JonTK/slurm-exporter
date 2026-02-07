// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/fixtures"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestNodesSimpleCollector_Describe(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewNodesSimpleCollector(mockClient, logger)

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

func TestNodesSimpleCollector_Collect_Success(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
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
	mockNodeManager.AssertExpectations(t)
}

func TestNodesSimpleCollector_Collect_Disabled(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(false)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Should not collect any metrics when disabled
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not collect metrics when disabled")
}

func TestNodesSimpleCollector_Collect_Error(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock to return error
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.Error(t, err)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestNodesSimpleCollector_AllocatedCPUs_UsesRealData(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Verify allocated CPU metrics use real data from API
	// Test fixtures define:
	// - node01: IDLE, CPUs=32, AllocCPUs=0
	// - node02: ALLOCATED, CPUs=64, AllocCPUs=64
	// - node03: MIXED, CPUs=48, AllocCPUs=24
	// - node04: DOWN, CPUs=32, AllocCPUs=0

	allocCPUMetrics := make(map[string]float64)
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "cpus_allocated") {
			// Extract metric value
			pb := &dto.Metric{}
			if err := metric.Write(pb); err == nil {
				nodeName := ""
				for _, label := range pb.GetLabel() {
					if label.GetName() == "node" {
						nodeName = label.GetValue()
						break
					}
				}
				if nodeName != "" {
					allocCPUMetrics[nodeName] = pb.GetGauge().GetValue()
				}
			}
		}
	}

	// Verify real data is used (not estimates based on state)
	assert.Equal(t, float64(0), allocCPUMetrics["node01"], "node01 should have 0 allocated CPUs")
	assert.Equal(t, float64(64), allocCPUMetrics["node02"], "node02 should have 64 allocated CPUs")
	assert.Equal(t, float64(24), allocCPUMetrics["node03"], "node03 should have 24 allocated CPUs")
	assert.Equal(t, float64(0), allocCPUMetrics["node04"], "node04 should have 0 allocated CPUs")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestNodesSimpleCollector_AllocatedMemory_UsesRealData(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Verify allocated memory metrics use real data from API
	// Test fixtures define:
	// - node01: IDLE, AllocMemory=0MB
	// - node02: ALLOCATED, AllocMemory=128GB (131072MB)
	// - node03: MIXED, AllocMemory=48GB (49152MB)
	// - node04: DOWN, AllocMemory=0MB
	allocMemoryMetrics := make(map[string]float64)
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "memory_allocated_bytes") {
			// Extract metric value
			pb := &dto.Metric{}
			if err := metric.Write(pb); err == nil {
				nodeName := ""
				for _, label := range pb.GetLabel() {
					if label.GetName() == "node" {
						nodeName = label.GetValue()
						break
					}
				}
				if nodeName != "" {
					allocMemoryMetrics[nodeName] = pb.Gauge.GetValue()
				}
			}
		}
	}

	// Verify real data is used (converted from MB to bytes)
	assert.Equal(t, float64(0), allocMemoryMetrics["node01"], "node01 should have 0 allocated memory")
	assert.Equal(t, float64(128*1024*1024*1024), allocMemoryMetrics["node02"], "node02 should have 128GB allocated memory")
	assert.Equal(t, float64(48*1024*1024*1024), allocMemoryMetrics["node03"], "node03 should have 48GB allocated memory")
	assert.Equal(t, float64(0), allocMemoryMetrics["node04"], "node04 should have 0 allocated memory")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestNodesSimpleCollector_Architecture_UsesRealData(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Verify node info metrics have correct architecture and OS
	type nodeInfo struct {
		arch string
		os   string
	}
	nodeInfoLabels := make(map[string]nodeInfo)
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "node_info") {
			// Extract metric labels
			pb := &dto.Metric{}
			if err := metric.Write(pb); err == nil {
				nodeName := ""
				arch := ""
				os := ""
				for _, label := range pb.GetLabel() {
					if label.GetName() == "node" {
						nodeName = label.GetValue()
					}
					if label.GetName() == "arch" {
						arch = label.GetValue()
					}
					if label.GetName() == "os" {
						os = label.GetValue()
					}
				}
				if nodeName != "" {
					nodeInfoLabels[nodeName] = nodeInfo{arch: arch, os: os}
				}
			}
		}
	}

	// Verify real data is used from API
	assert.Equal(t, "x86_64", nodeInfoLabels["node01"].arch, "node01 should have x86_64 architecture from API")
	assert.Equal(t, "Linux", nodeInfoLabels["node01"].os, "node01 should have Linux OS from API")

	assert.Equal(t, "x86_64", nodeInfoLabels["node02"].arch, "node02 should have x86_64 architecture from API")
	assert.Equal(t, "Linux", nodeInfoLabels["node02"].os, "node02 should have Linux OS from API")

	assert.Equal(t, "aarch64", nodeInfoLabels["node03"].arch, "node03 should have aarch64 architecture from API")
	assert.Equal(t, "Linux 5.15", nodeInfoLabels["node03"].os, "node03 should have Linux 5.15 OS from API")

	assert.Equal(t, "x86_64", nodeInfoLabels["node04"].arch, "node04 should have x86_64 architecture from API")
	assert.Equal(t, "Linux", nodeInfoLabels["node04"].os, "node04 should have Linux OS from API")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestNodesSimpleCollector_StateMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check that we have node state metrics
	hasStateMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "node_state") {
			hasStateMetrics = true
			break
		}
	}

	assert.True(t, hasStateMetrics, "should have node state metrics")
}

func TestNodesSimpleCollector_ResourceMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check for resource metrics
	metricTypes := make(map[string]bool)
	for metric := range ch {
		desc := metric.Desc()
		descStr := desc.String()
		if contains(descStr, "cpus_total") {
			metricTypes["cpus_total"] = true
		}
		if contains(descStr, "cpus_allocated") {
			metricTypes["cpus_allocated"] = true
		}
		if contains(descStr, "memory_total") {
			metricTypes["memory_total"] = true
		}
		if contains(descStr, "memory_allocated") {
			metricTypes["memory_allocated"] = true
		}
	}

	assert.True(t, metricTypes["cpus_total"], "should have cpus_total metrics")
	assert.True(t, metricTypes["cpus_allocated"], "should have cpus_allocated metrics")
	assert.True(t, metricTypes["memory_total"], "should have memory_total metrics")
	assert.True(t, metricTypes["memory_allocated"], "should have memory_allocated metrics")
}
