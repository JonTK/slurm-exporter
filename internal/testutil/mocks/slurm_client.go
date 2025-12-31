package mocks

import (
	"context"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
)

// MockSlurmClient is a mock implementation of slurm.SlurmClient
type MockSlurmClient struct {
	mock.Mock
}

// Jobs returns a mock JobManager
func (m *MockSlurmClient) Jobs() slurm.JobManager {
	args := m.Called()
	return args.Get(0).(slurm.JobManager)
}

// Nodes returns a mock NodeManager
func (m *MockSlurmClient) Nodes() slurm.NodeManager {
	args := m.Called()
	return args.Get(0).(slurm.NodeManager)
}

// Partitions returns a mock PartitionManager
func (m *MockSlurmClient) Partitions() slurm.PartitionManager {
	args := m.Called()
	return args.Get(0).(slurm.PartitionManager)
}

// Info returns a mock InfoManager
func (m *MockSlurmClient) Info() slurm.InfoManager {
	args := m.Called()
	return args.Get(0).(slurm.InfoManager)
}

// Close closes the client
func (m *MockSlurmClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockJobManager is a mock implementation of slurm.JobManager
type MockJobManager struct {
	mock.Mock
}

// List returns a list of jobs
func (m *MockJobManager) List(ctx context.Context, opts *slurm.ListJobsOptions) (*slurm.JobList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.JobList), args.Error(1)
}

// Get returns a specific job
func (m *MockJobManager) Get(ctx context.Context, jobID string) (*slurm.Job, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.Job), args.Error(1)
}

// Submit submits a new job
func (m *MockJobManager) Submit(ctx context.Context, job *slurm.JobSubmission) (*slurm.JobSubmitResponse, error) {
	args := m.Called(ctx, job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.JobSubmitResponse), args.Error(1)
}

// Cancel cancels a job
func (m *MockJobManager) Cancel(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

// MockNodeManager is a mock implementation of slurm.NodeManager
type MockNodeManager struct {
	mock.Mock
}

// List returns a list of nodes
func (m *MockNodeManager) List(ctx context.Context, opts *slurm.ListNodesOptions) (*slurm.NodeList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.NodeList), args.Error(1)
}

// Get returns a specific node
func (m *MockNodeManager) Get(ctx context.Context, nodeName string) (*slurm.Node, error) {
	args := m.Called(ctx, nodeName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.Node), args.Error(1)
}

// MockPartitionManager is a mock implementation of slurm.PartitionManager
type MockPartitionManager struct {
	mock.Mock
}

// List returns a list of partitions
func (m *MockPartitionManager) List(ctx context.Context, opts *slurm.ListPartitionsOptions) (*slurm.PartitionList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.PartitionList), args.Error(1)
}

// Get returns a specific partition
func (m *MockPartitionManager) Get(ctx context.Context, partitionName string) (*slurm.Partition, error) {
	args := m.Called(ctx, partitionName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.Partition), args.Error(1)
}

// MockInfoManager is a mock implementation of slurm.InfoManager
type MockInfoManager struct {
	mock.Mock
}

// Get returns cluster information
func (m *MockInfoManager) Get(ctx context.Context) (*slurm.ClusterInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.ClusterInfo), args.Error(1)
}

// Ping checks if the SLURM API is responsive
func (m *MockInfoManager) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Stats returns cluster statistics
func (m *MockInfoManager) Stats(ctx context.Context) (*slurm.ClusterStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.ClusterStats), args.Error(1)
}

// Version returns SLURM version information
func (m *MockInfoManager) Version(ctx context.Context) (*slurm.APIVersion, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*slurm.APIVersion), args.Error(1)
}

// MockCollectorInterface for testing collector interfaces
type MockCollectorInterface struct {
	mock.Mock
}

func (m *MockCollectorInterface) Describe(ch chan<- *prometheus.Desc) {
	m.Called(ch)
}

func (m *MockCollectorInterface) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	args := m.Called(ctx, ch)
	return args.Error(0)
}

// MockRegistry for testing collector registry
type MockRegistry struct {
	mock.Mock
}

func (m *MockRegistry) Configure(cfg interface{}) error {
	args := m.Called(cfg)
	return args.Error(0)
}

func (m *MockRegistry) Describe(ch chan<- *prometheus.Desc) {
	m.Called(ch)
}

func (m *MockRegistry) Collect(ch chan<- prometheus.Metric) {
	m.Called(ch)
}