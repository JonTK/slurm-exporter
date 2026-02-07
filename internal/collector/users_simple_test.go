// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"

	slurm "github.com/jontk/slurm-client"
	"github.com/jontk/slurm-client/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestUsersSimpleCollector_Describe(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewUsersSimpleCollector(mockClient, logger)

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

func TestUsersSimpleCollector_Collect_Success(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockUserManager := new(mocks.MockUserManager)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock expectations with test data
	// Helper function for pointers
	strPtr := func(s string) *string { return &s }

	userList := &slurm.UserList{
		Users: []slurm.User{
			{
				Name: "user1",
				Default: &slurm.UserDefault{
					Account: strPtr("account1"),
				},
				AdministratorLevel: []slurm.AdministratorLevelValue{api.AdministratorLevelNone},
			},
			{
				Name: "user2",
				Default: &slurm.UserDefault{
					Account: strPtr("account2"),
				},
				AdministratorLevel: []slurm.AdministratorLevelValue{api.AdministratorLevelOperator},
			},
		},
	}

	jobList := &slurm.JobList{
		Jobs: []slurm.Job{},
	}

	mockClient.On("Users").Return(mockUserManager)
	mockUserManager.On("List", mock.Anything, mock.Anything).Return(userList, nil)
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(jobList, nil)

	collector := NewUsersSimpleCollector(mockClient, logger)
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
	mockUserManager.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
}

func TestUsersSimpleCollector_Collect_Disabled(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewUsersSimpleCollector(mockClient, logger)
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

func TestUsersSimpleCollector_Collect_Error(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockUserManager := new(mocks.MockUserManager)

	// Setup mock to return error
	mockClient.On("Users").Return(mockUserManager)
	mockUserManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewUsersSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockUserManager.AssertExpectations(t)
}

func TestUsersSimpleCollector_EmptyUserList(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockUserManager := new(mocks.MockUserManager)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock to return empty lists
	emptyUserList := &slurm.UserList{
		Users: []slurm.User{},
	}

	emptyJobList := &slurm.JobList{
		Jobs: []slurm.Job{},
	}

	mockClient.On("Users").Return(mockUserManager)
	mockUserManager.On("List", mock.Anything, mock.Anything).Return(emptyUserList, nil)
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(emptyJobList, nil)

	collector := NewUsersSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty user list, no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not emit metrics when user list is empty")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockUserManager.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
}
