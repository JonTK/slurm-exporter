package testutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"go.uber.org/mock/gomock"
)

// MockSLURMClient provides a test-friendly SLURM client
// Note: This is a manual mock implementation. Auto-generated mocks cannot be used
// because the slurm-client interfaces are in an internal package.
type MockSLURMClient struct {
	ctrl       *gomock.Controller
	mu         sync.RWMutex
	jobs       []slurm.Job
	nodes      []slurm.Node
	partitions []slurm.Partition
	callCount  map[string]int
	responses  map[string]interface{}
	errors     map[string]error
	delays     map[string]time.Duration
}

// NewMockSLURMClient creates a new mock SLURM client
func NewMockSLURMClient(ctrl *gomock.Controller) *MockSLURMClient {
	return &MockSLURMClient{
		ctrl:      ctrl,
		callCount: make(map[string]int),
		responses: make(map[string]interface{}),
		errors:    make(map[string]error),
		delays:    make(map[string]time.Duration),
	}
}

// EXPECT returns the mock recorder
func (m *MockSLURMClient) EXPECT() *MockSLURMClientMockRecorder {
	return &MockSLURMClientMockRecorder{mock: m}
}

// ListJobs mocks listing jobs
func (m *MockSLURMClient) ListJobs(ctx context.Context, opts *slurm.ListJobsOptions) ([]slurm.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	method := "ListJobs"
	m.callCount[method]++

	// Apply any configured delay
	if delay, exists := m.delays[method]; exists {
		time.Sleep(delay)
	}

	// Check for configured error
	if err, exists := m.errors[method]; exists {
		return nil, err
	}

	// Check for configured response
	if resp, exists := m.responses[method]; exists {
		if jobs, ok := resp.([]slurm.Job); ok {
			return jobs, nil
		}
	}

	// Return default jobs
	return m.jobs, nil
}

// GetJob mocks getting a single job
func (m *MockSLURMClient) GetJob(ctx context.Context, id string) (*slurm.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	method := "GetJob"
	m.callCount[method]++

	// Apply any configured delay
	if delay, exists := m.delays[method]; exists {
		time.Sleep(delay)
	}

	// Check for configured error
	if err, exists := m.errors[method]; exists {
		return nil, err
	}

	// Find job by ID
	for _, job := range m.jobs {
		if job.ID == id {
			return &job, nil
		}
	}

	return nil, fmt.Errorf("job %s not found", id)
}

// ListNodes mocks listing nodes
func (m *MockSLURMClient) ListNodes(ctx context.Context, opts *slurm.ListNodesOptions) ([]slurm.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	method := "ListNodes"
	m.callCount[method]++

	// Apply any configured delay
	if delay, exists := m.delays[method]; exists {
		time.Sleep(delay)
	}

	// Check for configured error
	if err, exists := m.errors[method]; exists {
		return nil, err
	}

	// Check for configured response
	if resp, exists := m.responses[method]; exists {
		if nodes, ok := resp.([]slurm.Node); ok {
			return nodes, nil
		}
	}

	// Return default nodes
	return m.nodes, nil
}

// GetNode mocks getting a single node
func (m *MockSLURMClient) GetNode(ctx context.Context, name string) (*slurm.Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	method := "GetNode"
	m.callCount[method]++

	// Apply any configured delay
	if delay, exists := m.delays[method]; exists {
		time.Sleep(delay)
	}

	// Check for configured error
	if err, exists := m.errors[method]; exists {
		return nil, err
	}

	// Find node by name
	for _, node := range m.nodes {
		if node.Name == name {
			return &node, nil
		}
	}

	return nil, fmt.Errorf("node %s not found", name)
}

// ListPartitions mocks listing partitions
func (m *MockSLURMClient) ListPartitions(ctx context.Context, opts *slurm.ListPartitionsOptions) ([]slurm.Partition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	method := "ListPartitions"
	m.callCount[method]++

	// Apply any configured delay
	if delay, exists := m.delays[method]; exists {
		time.Sleep(delay)
	}

	// Check for configured error
	if err, exists := m.errors[method]; exists {
		return nil, err
	}

	// Check for configured response
	if resp, exists := m.responses[method]; exists {
		if partitions, ok := resp.([]slurm.Partition); ok {
			return partitions, nil
		}
	}

	// Return default partitions
	return m.partitions, nil
}

// GetPartition mocks getting a single partition
func (m *MockSLURMClient) GetPartition(ctx context.Context, name string) (*slurm.Partition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	method := "GetPartition"
	m.callCount[method]++

	// Apply any configured delay
	if delay, exists := m.delays[method]; exists {
		time.Sleep(delay)
	}

	// Check for configured error
	if err, exists := m.errors[method]; exists {
		return nil, err
	}

	// Find partition by name
	for _, partition := range m.partitions {
		if partition.Name == name {
			return &partition, nil
		}
	}

	return nil, fmt.Errorf("partition %s not found", name)
}

// GetClusterInfo mocks getting cluster information
func (m *MockSLURMClient) GetClusterInfo(ctx context.Context) (*slurm.ClusterInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	method := "GetClusterInfo"
	m.callCount[method]++

	// Apply any configured delay
	if delay, exists := m.delays[method]; exists {
		time.Sleep(delay)
	}

	// Check for configured error
	if err, exists := m.errors[method]; exists {
		return nil, err
	}

	// Check for configured response
	if resp, exists := m.responses[method]; exists {
		if info, ok := resp.(*slurm.ClusterInfo); ok {
			return info, nil
		}
	}

	// Return default cluster info
	return &slurm.ClusterInfo{
		ClusterName: "test-cluster",
		Version:     "23.02.0",
	}, nil
}

// Ping mocks ping
func (m *MockSLURMClient) Ping(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	method := "Ping"
	m.callCount[method]++

	// Apply any configured delay
	if delay, exists := m.delays[method]; exists {
		time.Sleep(delay)
	}

	// Check for configured error
	if err, exists := m.errors[method]; exists {
		return err
	}

	return nil
}

// SetJobs sets the jobs that will be returned
func (m *MockSLURMClient) SetJobs(jobs []slurm.Job) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs = jobs
}

// SetNodes sets the nodes that will be returned
func (m *MockSLURMClient) SetNodes(nodes []slurm.Node) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes = nodes
}

// SetPartitions sets the partitions that will be returned
func (m *MockSLURMClient) SetPartitions(partitions []slurm.Partition) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.partitions = partitions
}

// SetResponse sets a custom response for a method
func (m *MockSLURMClient) SetResponse(method string, response interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[method] = response
}

// SetError sets an error for a method
func (m *MockSLURMClient) SetError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method] = err
}

// SetDelay sets a delay for a method
func (m *MockSLURMClient) SetDelay(method string, delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delays[method] = delay
}

// GetCallCount returns the number of times a method was called
func (m *MockSLURMClient) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount[method]
}

// Reset resets all mock state
func (m *MockSLURMClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.jobs = nil
	m.nodes = nil
	m.partitions = nil
	m.callCount = make(map[string]int)
	m.responses = make(map[string]interface{})
	m.errors = make(map[string]error)
	m.delays = make(map[string]time.Duration)
}

// MockSLURMClientMockRecorder provides expectation recording
type MockSLURMClientMockRecorder struct {
	mock *MockSLURMClient
}

// ListJobs records an expectation for ListJobs
func (mr *MockSLURMClientMockRecorder) ListJobs(ctx, opts interface{}) *MockSLURMClientListJobsCall {
	return &MockSLURMClientListJobsCall{
		mock: mr.mock,
		ctx:  ctx,
		opts: opts,
	}
}

// MockSLURMClientListJobsCall represents a call expectation
type MockSLURMClientListJobsCall struct {
	mock     *MockSLURMClient
	ctx      interface{}
	opts     interface{}
	response []slurm.Job
	err      error
	times    int
}

// Return sets the return values
func (c *MockSLURMClientListJobsCall) Return(jobs []slurm.Job, err error) *MockSLURMClientListJobsCall {
	c.response = jobs
	c.err = err
	if err != nil {
		c.mock.SetError("ListJobs", err)
	} else {
		c.mock.SetResponse("ListJobs", jobs)
	}
	return c
}

// Times sets the expected number of calls
func (c *MockSLURMClientListJobsCall) Times(times int) *MockSLURMClientListJobsCall {
	c.times = times
	return c
}

// DoAndReturn allows custom behavior
func (c *MockSLURMClientListJobsCall) DoAndReturn(fn func(context.Context, interface{}) ([]slurm.Job, error)) *MockSLURMClientListJobsCall {
	// This would be implemented to support DoAndReturn behavior
	return c
}

// SetupMockJobList is a helper function to setup job list behavior
func SetupMockJobList(mock *MockSLURMClient, jobs []slurm.Job) {
	mock.SetJobs(jobs)
}

// SetupMockNodeList is a helper function to setup node list behavior
func SetupMockNodeList(mock *MockSLURMClient, nodes []slurm.Node) {
	mock.SetNodes(nodes)
}

// SetupMockPartitionList is a helper function to setup partition list behavior
func SetupMockPartitionList(mock *MockSLURMClient, partitions []slurm.Partition) {
	mock.SetPartitions(partitions)
}
