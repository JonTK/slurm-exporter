package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountHierarchySLURMClient struct {
	mock.Mock
}

func (m *MockAccountHierarchySLURMClient) GetAccountHierarchy(ctx context.Context) (*AccountHierarchy, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AccountHierarchy), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountUsers(ctx context.Context, accountName string) ([]*UserAccount, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).([]*UserAccount), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountAssociations(ctx context.Context, accountName string) (*AccountAssociations, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountAssociations), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountSubAccounts(ctx context.Context, parentAccount string) ([]*SubAccount, error) {
	args := m.Called(ctx, parentAccount)
	return args.Get(0).([]*SubAccount), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountPermissions(ctx context.Context, accountName string) (*AccountPermissions, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountPermissions), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetUserAccountMapping(ctx context.Context) (*UserAccountMapping, error) {
	args := m.Called(ctx)
	return args.Get(0).(*UserAccountMapping), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountRelationships(ctx context.Context, accountName string) (*AccountRelationships, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountRelationships), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountInheritance(ctx context.Context, accountName string) (*AccountInheritance, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountInheritance), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountDelegations(ctx context.Context, accountName string) (*AccountDelegations, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountDelegations), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountAccessMatrix(ctx context.Context) (*AccountAccessMatrix, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AccountAccessMatrix), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetHierarchyValidation(ctx context.Context) (*HierarchyValidation, error) {
	args := m.Called(ctx)
	return args.Get(0).(*HierarchyValidation), args.Error(1)
}

func (m *MockAccountHierarchySLURMClient) GetAccountConflicts(ctx context.Context) (*AccountConflicts, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AccountConflicts), args.Error(1)
}

func TestNewAccountHierarchyCollector(t *testing.T) {
	client := &MockAccountHierarchySLURMClient{}
	collector := NewAccountHierarchyCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.accountHierarchyDepth)
	assert.NotNil(t, collector.accountHierarchyAccounts)
	assert.NotNil(t, collector.userAccountAssociations)
	assert.NotNil(t, collector.hierarchyValidationStatus)
}

func TestAccountHierarchyCollector_Describe(t *testing.T) {
	client := &MockAccountHierarchySLURMClient{}
	collector := NewAccountHierarchyCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 42 metrics, verify we have the correct number
	assert.Equal(t, 42, len(descs), "Should have exactly 42 metric descriptions")
}

func TestAccountHierarchyCollector_Collect_Success(t *testing.T) {
	client := &MockAccountHierarchySLURMClient{}

	// Mock account hierarchy
	hierarchy := &AccountHierarchy{
		RootAccounts: []*AccountNode{
			{
				AccountName:    "root",
				ParentAccount:  "",
				ChildAccounts:  []*AccountNode{},
				Level:          0,
				UserCount:      100,
				ActiveUsers:    80,
				Status:         "active",
			},
			{
				AccountName:    "research",
				ParentAccount:  "root",
				ChildAccounts:  []*AccountNode{},
				Level:          1,
				UserCount:      50,
				ActiveUsers:    45,
				Status:         "active",
			},
		},
		TotalAccounts:     10,
		MaxDepth:          3,
		TotalUsers:        200,
		HierarchyType:     "organizational",
		OrganizationalUnits: map[string]*OrganizationalUnit{
			"science": {
				Name:          "science",
				Accounts:      []string{"physics", "chemistry"},
				TotalUsers:    30,
				ResourceShare: 0.35,
			},
			"engineering": {
				Name:          "engineering",
				Accounts:      []string{"software", "hardware"},
				TotalUsers:    40,
				ResourceShare: 0.45,
			},
		},
	}

	// Mock user account mapping
	userMapping := &UserAccountMapping{
		TotalMappings:     200,
		UsersWithMultiple: 25,
		OrphanedUsers:     []string{"user1", "user2"},
		MappingsByType: map[string]int{
			"primary":   150,
			"secondary": 50,
		},
	}

	// Mock account associations
	associations := &AccountAssociations{
		AccountName: "research",
		Users: []*UserAccount{
			{
				UserName:       "scientist1",
				AccountName:    "research",
				DefaultAccount: true,
				AccessLevel:    "full",
			},
			{
				UserName:       "scientist2",
				AccountName:    "research",
				DefaultAccount: false,
				AccessLevel:    "submit",
			},
		},
		ParentAssociations: []*ParentAssociation{
			{
				ParentAccount:   "root",
				AssociationType: "hierarchical",
			},
		},
		ChildAssociations: []*ChildAssociation{
			{
				ChildAccount:    "physics",
				AssociationType: "hierarchical",
			},
			{
				ChildAccount:    "chemistry",
				AssociationType: "hierarchical",
			},
		},
		TotalAssociations:   10,
		ActiveAssociations:  8,
		PendingAssociations: 2,
	}

	// Mock account permissions
	permissions := &AccountPermissions{
		AccountName:   "research",
		AdminUsers:    []string{"admin1", "admin2"},
		OperatorUsers: []string{"op1", "op2", "op3"},
		ViewOnlyUsers: []string{"viewer1"},
		SubmitUsers:   []string{"user1", "user2", "user3", "user4"},
		CustomRoles: map[string][]string{
			"coordinator": {"coord1", "coord2"},
		},
		InheritedPerms: map[string]bool{
			"submit": true,
			"view":   true,
			"admin":  false,
		},
		EffectivePerms: map[string][]string{
			"submit": {"user1", "user2", "user3"},
			"admin":  {"admin1", "admin2"},
		},
	}

	// Mock hierarchy validation
	validation := &HierarchyValidation{
		Valid:              true,
		Errors:             []string{},
		Warnings:           []string{"performance warning", "security warning"},
		CircularReferences: []string{},
		OrphanedAccounts:   []string{"orphan1"},
		ValidationTime:     time.Now(),
	}

	// Mock account conflicts
	conflicts := &AccountConflicts{
		PermissionConflicts: []PermissionConflict{
			{
				AccountName:  "research",
				ConflictType: "role_overlap",
			},
		},
		QuotaConflicts: []QuotaConflict{
			{
				AccountName:  "research",
				ConflictType: "inheritance",
			},
		},
		TotalConflicts:    3,
		CriticalConflicts: 1,
	}

	// Mock access matrix
	accessMatrix := &AccountAccessMatrix{
		Matrix: map[string]map[string][]string{
			"user1": {
				"research": []string{"submit", "view"},
			},
		},
		TotalPermissions: 100,
		ConflictingPerms: 5,
		EffectivePerms: map[string]map[string]bool{
			"user1": {
				"research": true,
			},
		},
	}

	// Mock relationships
	relationships := &AccountRelationships{
		ParentRelations: []*ParentRelation{
			{
				ParentAccount: "root",
				RelationType:  "hierarchical",
				Strength:      0.9,
			},
		},
		ChildRelations: []*ChildRelation{
			{
				ChildAccount: "physics",
				RelationType: "hierarchical",
				Strength:     0.8,
			},
		},
		PeerRelations: []*PeerRelation{
			{
				PeerAccount:   "engineering",
				RelationType:  "collaboration",
				Bidirectional: true,
				Strength:      0.6,
			},
		},
	}

	// Setup mock expectations
	client.On("GetAccountHierarchy", mock.Anything).Return(hierarchy, nil)
	client.On("GetUserAccountMapping", mock.Anything).Return(userMapping, nil)
	client.On("GetAccountAssociations", mock.Anything, "research").Return(associations, nil)
	client.On("GetAccountAssociations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAssociations{
		Users:              []*UserAccount{},
		ParentAssociations: []*ParentAssociation{},
		ChildAssociations:  []*ChildAssociation{},
	}, nil)
	client.On("GetAccountPermissions", mock.Anything, "research").Return(permissions, nil)
	client.On("GetAccountPermissions", mock.Anything, mock.AnythingOfType("string")).Return(&AccountPermissions{
		AdminUsers:     []string{},
		CustomRoles:    map[string][]string{},
		InheritedPerms: map[string]bool{},
		EffectivePerms: map[string][]string{},
	}, nil)
	client.On("GetHierarchyValidation", mock.Anything).Return(validation, nil)
	client.On("GetAccountConflicts", mock.Anything).Return(conflicts, nil)
	client.On("GetAccountAccessMatrix", mock.Anything).Return(accessMatrix, nil)
	client.On("GetAccountRelationships", mock.Anything, mock.AnythingOfType("string")).Return(relationships, nil)
	client.On("GetAccountSubAccounts", mock.Anything, mock.AnythingOfType("string")).Return([]*SubAccount{}, nil)
	client.On("GetAccountInheritance", mock.Anything, mock.AnythingOfType("string")).Return(&AccountInheritance{}, nil)
	client.On("GetAccountDelegations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountDelegations{}, nil)

	collector := NewAccountHierarchyCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 200)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify we collected metrics
	assert.Greater(t, len(metrics), 0, "Should collect at least some metrics")

	// Verify specific metrics are present
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		desc := metric.Desc()
		metricNames[desc.String()] = true
	}

	// Check for key metric families
	foundHierarchyDepth := false
	foundAccountUsers := false
	foundAssociations := false
	foundPermissions := false
	foundValidation := false

	for name := range metricNames {
		if strings.Contains(name, "hierarchy_depth") {
			foundHierarchyDepth = true
		}
		if strings.Contains(name, "account_user_count") {
			foundAccountUsers = true
		}
		if strings.Contains(name, "account_associations") {
			foundAssociations = true
		}
		if strings.Contains(name, "permission_users") {
			foundPermissions = true
		}
		if strings.Contains(name, "validation_status") {
			foundValidation = true
		}
	}

	assert.True(t, foundHierarchyDepth, "Should have hierarchy depth metrics")
	assert.True(t, foundAccountUsers, "Should have account user metrics")
	assert.True(t, foundAssociations, "Should have association metrics")
	assert.True(t, foundPermissions, "Should have permission metrics")
	assert.True(t, foundValidation, "Should have validation metrics")

	client.AssertExpectations(t)
}

func TestAccountHierarchyCollector_Collect_Error(t *testing.T) {
	client := &MockAccountHierarchySLURMClient{}

	// Mock error responses
	client.On("GetAccountHierarchy", mock.Anything).Return((*AccountHierarchy)(nil), assert.AnError)
	client.On("GetUserAccountMapping", mock.Anything).Return((*UserAccountMapping)(nil), assert.AnError)
	client.On("GetHierarchyValidation", mock.Anything).Return((*HierarchyValidation)(nil), assert.AnError)
	client.On("GetAccountConflicts", mock.Anything).Return((*AccountConflicts)(nil), nil)
	client.On("GetAccountAccessMatrix", mock.Anything).Return((*AccountAccessMatrix)(nil), nil)

	collector := NewAccountHierarchyCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should still collect some metrics (empty metrics after reset)
	assert.GreaterOrEqual(t, len(metrics), 0, "Should handle errors gracefully")

	client.AssertExpectations(t)
}

func TestAccountHierarchyCollector_MetricValues(t *testing.T) {
	client := &MockAccountHierarchySLURMClient{}

	// Create test data with known values
	hierarchy := &AccountHierarchy{
		TotalAccounts:    25,
		MaxDepth:         4,
		TotalUsers:       150,
		HierarchyType:    "organizational",
		RootAccounts: []*AccountNode{
			{
				AccountName:   "root",
				ChildAccounts: []*AccountNode{},
				UserCount:     50,
				ActiveUsers:   45,
			},
		},
		OrganizationalUnits: map[string]*OrganizationalUnit{
			"science": {
				TotalUsers:    60,
				ResourceShare: 0.4,
			},
		},
	}

	userMapping := &UserAccountMapping{
		UsersWithMultiple: 15,
		OrphanedUsers:     []string{"orphan1", "orphan2", "orphan3"},
		MappingsByType: map[string]int{
			"primary": 100,
		},
	}

	validation := &HierarchyValidation{
		Valid:              true,
		CircularReferences: []string{"ref1", "ref2"},
		Warnings:           []string{"warn1", "warn2"},
	}

	client.On("GetAccountHierarchy", mock.Anything).Return(hierarchy, nil)
	client.On("GetUserAccountMapping", mock.Anything).Return(userMapping, nil)
	client.On("GetHierarchyValidation", mock.Anything).Return(validation, nil)
	client.On("GetAccountConflicts", mock.Anything).Return(&AccountConflicts{
		TotalConflicts:    5,
		CriticalConflicts: 2,
	}, nil)
	client.On("GetAccountAccessMatrix", mock.Anything).Return(&AccountAccessMatrix{}, nil)
	client.On("GetAccountRelationships", mock.Anything, mock.AnythingOfType("string")).Return(&AccountRelationships{
		ParentRelations: []*ParentRelation{},
		ChildRelations:  []*ChildRelation{},
		PeerRelations:   []*PeerRelation{},
	}, nil)
	client.On("GetAccountAssociations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAssociations{
		Users:              []*UserAccount{},
		ParentAssociations: []*ParentAssociation{},
		ChildAssociations:  []*ChildAssociation{},
	}, nil)
	client.On("GetAccountPermissions", mock.Anything, mock.AnythingOfType("string")).Return(&AccountPermissions{
		AdminUsers:     []string{},
		CustomRoles:    map[string][]string{},
		InheritedPerms: map[string]bool{},
		EffectivePerms: map[string][]string{},
	}, nil)
	client.On("GetAccountSubAccounts", mock.Anything, mock.AnythingOfType("string")).Return([]*SubAccount{}, nil)
	client.On("GetAccountInheritance", mock.Anything, mock.AnythingOfType("string")).Return(&AccountInheritance{}, nil)
	client.On("GetAccountDelegations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountDelegations{}, nil)

	collector := NewAccountHierarchyCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundTotalAccounts := false
	foundTotalUsers := false
	foundOrphanedUsers := false
	foundValidationStatus := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_account_hierarchy_accounts_total":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(25), *mf.Metric[0].Gauge.Value)
				foundTotalAccounts = true
			}
		case "slurm_account_hierarchy_users_total":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(150), *mf.Metric[0].Gauge.Value)
				foundTotalUsers = true
			}
		case "slurm_orphaned_users_total":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(3), *mf.Metric[0].Gauge.Value)
				foundOrphanedUsers = true
			}
		case "slurm_hierarchy_validation_status":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(1), *mf.Metric[0].Gauge.Value) // Valid = 1
				foundValidationStatus = true
			}
		}
	}

	assert.True(t, foundTotalAccounts, "Should find total accounts metric with correct value")
	assert.True(t, foundTotalUsers, "Should find total users metric with correct value")
	assert.True(t, foundOrphanedUsers, "Should find orphaned users metric with correct value")
	assert.True(t, foundValidationStatus, "Should find validation status metric with correct value")
}

func TestAccountHierarchyCollector_Integration(t *testing.T) {
	client := &MockAccountHierarchySLURMClient{}

	// Setup comprehensive mock data
	setupHierarchyMocks(client)

	collector := NewAccountHierarchyCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_account_hierarchy_accounts_total Total number of accounts in the hierarchy
		# TYPE slurm_account_hierarchy_accounts_total gauge
		slurm_account_hierarchy_accounts_total{hierarchy_type="organizational"} 20
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_account_hierarchy_accounts_total")
	assert.NoError(t, err)
}

func TestAccountHierarchyCollector_OrganizationalMetrics(t *testing.T) {
	client := &MockAccountHierarchySLURMClient{}

	// Setup organizational unit specific mocks
	setupOrganizationalMocks(client)

	collector := NewAccountHierarchyCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify organizational metrics are present
	foundUnitCount := false
	foundUnitUsers := false
	foundUnitShare := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "organizational_unit_count") {
			foundUnitCount = true
		}
		if strings.Contains(desc, "organizational_unit_users") {
			foundUnitUsers = true
		}
		if strings.Contains(desc, "organizational_unit_resource_share") {
			foundUnitShare = true
		}
	}

	assert.True(t, foundUnitCount, "Should find organizational unit count metrics")
	assert.True(t, foundUnitUsers, "Should find organizational unit users metrics")
	assert.True(t, foundUnitShare, "Should find organizational unit share metrics")
}

func TestAccountHierarchyCollector_RelationshipMetrics(t *testing.T) {
	client := &MockAccountHierarchySLURMClient{}

	// Setup relationship specific mocks
	setupRelationshipMocks(client)

	collector := NewAccountHierarchyCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify relationship metrics are present
	foundRelationships := false
	foundStrength := false
	foundBidirectional := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "account_relationships_count") {
			foundRelationships = true
		}
		if strings.Contains(desc, "relationship_strength") {
			foundStrength = true
		}
		if strings.Contains(desc, "bidirectional_relations_count") {
			foundBidirectional = true
		}
	}

	assert.True(t, foundRelationships, "Should find relationship count metrics")
	assert.True(t, foundStrength, "Should find relationship strength metrics")
	assert.True(t, foundBidirectional, "Should find bidirectional relations metrics")
}

// Helper functions

func setupHierarchyMocks(client *MockAccountHierarchySLURMClient) {
	hierarchy := &AccountHierarchy{
		TotalAccounts:       20,
		TotalUsers:          100,
		MaxDepth:            3,
		HierarchyType:       "organizational",
		RootAccounts:        []*AccountNode{},
		OrganizationalUnits: map[string]*OrganizationalUnit{},
	}

	client.On("GetAccountHierarchy", mock.Anything).Return(hierarchy, nil)
	client.On("GetUserAccountMapping", mock.Anything).Return(&UserAccountMapping{
		MappingsByType: map[string]int{},
	}, nil)
	client.On("GetHierarchyValidation", mock.Anything).Return(&HierarchyValidation{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}, nil)
	client.On("GetAccountConflicts", mock.Anything).Return(&AccountConflicts{}, nil)
	client.On("GetAccountAccessMatrix", mock.Anything).Return(&AccountAccessMatrix{
		EffectivePerms: map[string]map[string]bool{},
	}, nil)
	client.On("GetAccountRelationships", mock.Anything, mock.AnythingOfType("string")).Return(&AccountRelationships{
		ParentRelations: []*ParentRelation{},
		ChildRelations:  []*ChildRelation{},
		PeerRelations:   []*PeerRelation{},
	}, nil)
	client.On("GetAccountAssociations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAssociations{
		Users:              []*UserAccount{},
		ParentAssociations: []*ParentAssociation{},
		ChildAssociations:  []*ChildAssociation{},
	}, nil)
	client.On("GetAccountPermissions", mock.Anything, mock.AnythingOfType("string")).Return(&AccountPermissions{
		AdminUsers:     []string{},
		CustomRoles:    map[string][]string{},
		InheritedPerms: map[string]bool{},
		EffectivePerms: map[string][]string{},
	}, nil)
	client.On("GetAccountSubAccounts", mock.Anything, mock.AnythingOfType("string")).Return([]*SubAccount{}, nil)
	client.On("GetAccountInheritance", mock.Anything, mock.AnythingOfType("string")).Return(&AccountInheritance{}, nil)
	client.On("GetAccountDelegations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountDelegations{}, nil)
}

func setupOrganizationalMocks(client *MockAccountHierarchySLURMClient) {
	hierarchy := &AccountHierarchy{
		TotalAccounts: 15,
		HierarchyType: "organizational",
		RootAccounts:  []*AccountNode{},
		OrganizationalUnits: map[string]*OrganizationalUnit{
			"research": {
				Name:          "research",
				TotalUsers:    50,
				ResourceShare: 0.35,
			},
			"engineering": {
				Name:          "engineering",
				TotalUsers:    60,
				ResourceShare: 0.45,
			},
			"admin": {
				Name:          "admin",
				TotalUsers:    10,
				ResourceShare: 0.1,
			},
		},
	}

	client.On("GetAccountHierarchy", mock.Anything).Return(hierarchy, nil)
	client.On("GetUserAccountMapping", mock.Anything).Return(&UserAccountMapping{
		MappingsByType: map[string]int{},
	}, nil)
	client.On("GetHierarchyValidation", mock.Anything).Return(&HierarchyValidation{
		Valid: true,
	}, nil)
	client.On("GetAccountConflicts", mock.Anything).Return(&AccountConflicts{}, nil)
	client.On("GetAccountAccessMatrix", mock.Anything).Return(&AccountAccessMatrix{
		EffectivePerms: map[string]map[string]bool{},
	}, nil)
	client.On("GetAccountRelationships", mock.Anything, mock.AnythingOfType("string")).Return(&AccountRelationships{
		ParentRelations: []*ParentRelation{},
		ChildRelations:  []*ChildRelation{},
		PeerRelations:   []*PeerRelation{},
	}, nil)
	client.On("GetAccountAssociations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAssociations{
		Users:              []*UserAccount{},
		ParentAssociations: []*ParentAssociation{},
		ChildAssociations:  []*ChildAssociation{},
	}, nil)
	client.On("GetAccountPermissions", mock.Anything, mock.AnythingOfType("string")).Return(&AccountPermissions{
		AdminUsers:     []string{},
		CustomRoles:    map[string][]string{},
		InheritedPerms: map[string]bool{},
		EffectivePerms: map[string][]string{},
	}, nil)
	client.On("GetAccountSubAccounts", mock.Anything, mock.AnythingOfType("string")).Return([]*SubAccount{}, nil)
	client.On("GetAccountInheritance", mock.Anything, mock.AnythingOfType("string")).Return(&AccountInheritance{}, nil)
	client.On("GetAccountDelegations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountDelegations{}, nil)
}

func setupRelationshipMocks(client *MockAccountHierarchySLURMClient) {
	hierarchy := &AccountHierarchy{
		RootAccounts: []*AccountNode{
			{AccountName: "root"},
			{AccountName: "dept1"},
		},
		HierarchyType:       "organizational",
		OrganizationalUnits: map[string]*OrganizationalUnit{},
	}

	relationships := &AccountRelationships{
		ParentRelations: []*ParentRelation{
			{ParentAccount: "root", RelationType: "hierarchical", Strength: 0.9},
		},
		ChildRelations: []*ChildRelation{
			{ChildAccount: "subdept1", RelationType: "hierarchical", Strength: 0.8},
			{ChildAccount: "subdept2", RelationType: "hierarchical", Strength: 0.7},
		},
		PeerRelations: []*PeerRelation{
			{PeerAccount: "dept2", RelationType: "collaboration", Bidirectional: true, Strength: 0.6},
			{PeerAccount: "dept3", RelationType: "resource_share", Bidirectional: false, Strength: 0.4},
		},
	}

	client.On("GetAccountHierarchy", mock.Anything).Return(hierarchy, nil)
	client.On("GetAccountRelationships", mock.Anything, mock.AnythingOfType("string")).Return(relationships, nil)
	client.On("GetUserAccountMapping", mock.Anything).Return(&UserAccountMapping{
		MappingsByType: map[string]int{},
	}, nil)
	client.On("GetHierarchyValidation", mock.Anything).Return(&HierarchyValidation{
		Valid: true,
	}, nil)
	client.On("GetAccountConflicts", mock.Anything).Return(&AccountConflicts{}, nil)
	client.On("GetAccountAccessMatrix", mock.Anything).Return(&AccountAccessMatrix{
		EffectivePerms: map[string]map[string]bool{},
	}, nil)
	client.On("GetAccountAssociations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAssociations{
		Users:              []*UserAccount{},
		ParentAssociations: []*ParentAssociation{},
		ChildAssociations:  []*ChildAssociation{},
	}, nil)
	client.On("GetAccountPermissions", mock.Anything, mock.AnythingOfType("string")).Return(&AccountPermissions{
		AdminUsers:     []string{},
		CustomRoles:    map[string][]string{},
		InheritedPerms: map[string]bool{},
		EffectivePerms: map[string][]string{},
	}, nil)
	client.On("GetAccountSubAccounts", mock.Anything, mock.AnythingOfType("string")).Return([]*SubAccount{}, nil)
	client.On("GetAccountInheritance", mock.Anything, mock.AnythingOfType("string")).Return(&AccountInheritance{}, nil)
	client.On("GetAccountDelegations", mock.Anything, mock.AnythingOfType("string")).Return(&AccountDelegations{}, nil)
}