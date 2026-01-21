package collector

import (
	"context"
	"testing"

	"github.com/jontk/slurm-client/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestAccountsSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewAccountsSimpleCollector(mockClient, logger)

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

func TestAccountsSimpleCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockAccountManager := new(mocks.MockAccountManager)
	mockAssociationManager := new(mocks.MockAssociationManager)

	// Setup mock expectations with test data
	maxTRES := map[string]int{
		"mem": 102400, // 100GB in MB
	}

	accountList := &interfaces.AccountList{
		Accounts: []interfaces.Account{
			{
				Name:             "account1",
				Organization:     "org1",
				Description:      "Test account 1",
				ParentAccount:    "root",
				DefaultPartition: "compute",
				CPULimit:         1000,
				MaxJobs:          100,
				MaxNodes:         50,
				MaxTRES:          maxTRES,
			},
			{
				Name:          "account2",
				Organization:  "org2",
				Description:   "Test account 2",
				ParentAccount: "root",
				CPULimit:      500,
			},
		},
		Total: 2,
	}

	// Setup associations for user counting
	associationList := &interfaces.AssociationList{
		Associations: []*interfaces.Association{
			{
				ID:      1,
				User:    "user1",
				Account: "account1",
			},
			{
				ID:      2,
				User:    "user2",
				Account: "account1",
			},
			{
				ID:      3,
				User:    "user3",
				Account: "account2",
			},
		},
		Total: 3,
	}

	mockClient.On("Accounts").Return(mockAccountManager)
	mockAccountManager.On("List", mock.Anything, mock.Anything).Return(accountList, nil)
	mockClient.On("Associations").Return(mockAssociationManager)
	mockAssociationManager.On("List", mock.Anything, mock.Anything).Return(associationList, nil)

	collector := NewAccountsSimpleCollector(mockClient, logger)
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
	mockAccountManager.AssertExpectations(t)
	mockAssociationManager.AssertExpectations(t)
}

func TestAccountsSimpleCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewAccountsSimpleCollector(mockClient, logger)
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

func TestAccountsSimpleCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockAccountManager := new(mocks.MockAccountManager)

	// Setup mock to return error
	mockClient.On("Accounts").Return(mockAccountManager)
	mockAccountManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewAccountsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockAccountManager.AssertExpectations(t)
}

func TestAccountsSimpleCollector_EmptyAccountList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockAccountManager := new(mocks.MockAccountManager)
	mockAssociationManager := new(mocks.MockAssociationManager)

	// Setup mock to return empty lists
	emptyAccountList := &interfaces.AccountList{
		Accounts: []interfaces.Account{},
		Total:    0,
	}

	emptyAssociationList := &interfaces.AssociationList{
		Associations: []*interfaces.Association{},
		Total:        0,
	}

	mockClient.On("Accounts").Return(mockAccountManager)
	mockAccountManager.On("List", mock.Anything, mock.Anything).Return(emptyAccountList, nil)
	mockClient.On("Associations").Return(mockAssociationManager)
	mockAssociationManager.On("List", mock.Anything, mock.Anything).Return(emptyAssociationList, nil)

	collector := NewAccountsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty account list, no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not emit metrics when account list is empty")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockAccountManager.AssertExpectations(t)
	mockAssociationManager.AssertExpectations(t)
}
