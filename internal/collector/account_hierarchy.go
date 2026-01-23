// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// AccountHierarchySLURMClient defines the interface for SLURM client operations related to account hierarchy
type AccountHierarchySLURMClient interface {
	GetAccountHierarchy(ctx context.Context) (*AccountHierarchy, error)
	GetAccountUsers(ctx context.Context, accountName string) ([]*UserAccount, error)
	GetAccountAssociations(ctx context.Context, accountName string) (*AccountAssociations, error)
	GetAccountSubAccounts(ctx context.Context, parentAccount string) ([]*SubAccount, error)
	GetAccountPermissions(ctx context.Context, accountName string) (*AccountPermissions, error)
	GetUserAccountMapping(ctx context.Context) (*UserAccountMapping, error)
	GetAccountRelationships(ctx context.Context, accountName string) (*AccountRelationships, error)
	GetAccountInheritance(ctx context.Context, accountName string) (*AccountInheritance, error)
	GetAccountDelegations(ctx context.Context, accountName string) (*AccountDelegations, error)
	GetAccountAccessMatrix(ctx context.Context) (*AccountAccessMatrix, error)
	GetHierarchyValidation(ctx context.Context) (*HierarchyValidation, error)
	GetAccountConflicts(ctx context.Context) (*AccountConflicts, error)
}

// Note: AccountHierarchy and AccountNode types are defined in common_types.go

// UserAccount represents a user's account association
type UserAccount struct {
	UserName          string
	AccountName       string
	AssociationType   string
	Permissions       []string
	DefaultAccount    bool
	EffectiveFrom     time.Time
	EffectiveTo       time.Time
	Status            string
	ShareContribution float64
	QuotaAllocation   map[string]float64
	AccessLevel       string
	LastActivity      time.Time
}

// AccountAssociations represents all associations for an account
type AccountAssociations struct {
	AccountName          string
	Users                []*UserAccount
	ParentAssociations   []*ParentAssociation
	ChildAssociations    []*ChildAssociation
	PeerAssociations     []*PeerAssociation
	ExternalAssociations []*ExternalAssociation
	TotalAssociations    int
	ActiveAssociations   int
	PendingAssociations  int
}

// SubAccount represents a sub-account relationship
type SubAccount struct {
	AccountName      string
	ParentAccount    string
	InheritedQuotas  bool
	InheritedQoS     bool
	InheritedUsers   bool
	OverrideSettings map[string]interface{}
	Status           string
}

// ParentAssociation represents a parent account association
type ParentAssociation struct {
	ParentAccount    string
	AssociationType  string
	InheritanceRules map[string]bool
}

// ChildAssociation represents a child account association
type ChildAssociation struct {
	ChildAccount    string
	AssociationType string
	DelegatedRights map[string]bool
}

// PeerAssociation represents a peer account association
type PeerAssociation struct {
	PeerAccount       string
	AssociationType   string
	SharedResources   []string
	MutualPermissions map[string]bool
}

// ExternalAssociation represents external account associations
type ExternalAssociation struct {
	ExternalAccount  string
	AssociationType  string
	IntegrationType  string
	ValidationStatus string
}

// AccountPermissions represents account permissions
type AccountPermissions struct {
	AccountName    string
	AdminUsers     []string
	OperatorUsers  []string
	ViewOnlyUsers  []string
	SubmitUsers    []string
	CustomRoles    map[string][]string
	InheritedPerms map[string]bool
	EffectivePerms map[string][]string
}

// UserAccountMapping represents system-wide user-account mappings
type UserAccountMapping struct {
	TotalMappings     int
	UsersWithMultiple int
	OrphanedUsers     []string
	MappingsByType    map[string]int
}

// AccountRelationships represents account relationships
type AccountRelationships struct {
	ParentRelations []*ParentRelation
	ChildRelations  []*ChildRelation
	PeerRelations   []*PeerRelation
	DependencyGraph map[string][]string
}

// ParentRelation represents a parent relationship
type ParentRelation struct {
	ParentAccount string
	RelationType  string
	Strength      float64
}

// ChildRelation represents a child relationship
type ChildRelation struct {
	ChildAccount string
	RelationType string
	Strength     float64
}

// PeerRelation represents a peer relationship
type PeerRelation struct {
	PeerAccount   string
	RelationType  string
	Bidirectional bool
	Strength      float64
}

// AccountInheritance represents inheritance settings
type AccountInheritance struct {
	AccountName          string
	InheritedFrom        []string
	InheritedProperties  map[string]interface{}
	OverriddenProperties map[string]interface{}
	InheritanceChain     []string
	EffectiveSettings    map[string]interface{}
}

// AccountDelegations represents delegation settings
type AccountDelegations struct {
	AccountName        string
	DelegatedTo        map[string][]string
	DelegatedFrom      map[string][]string
	DelegationRules    map[string]interface{}
	ActiveDelegations  int
	PendingDelegations int
}

// AccountAccessMatrix represents the access control matrix
type AccountAccessMatrix struct {
	Matrix           map[string]map[string][]string
	TotalPermissions int
	ConflictingPerms int
	EffectivePerms   map[string]map[string]bool
}

// HierarchyValidation represents hierarchy validation results
type HierarchyValidation struct {
	Valid              bool
	Errors             []string
	Warnings           []string
	CircularReferences []string
	OrphanedAccounts   []string
	ConflictingRules   []string
	ValidationTime     time.Time
}

// AccountConflicts represents detected conflicts
type AccountConflicts struct {
	PermissionConflicts []PermissionConflict
	QuotaConflicts      []QuotaConflict
	HierarchyConflicts  []HierarchyConflict
	TotalConflicts      int
	CriticalConflicts   int
}

// PermissionConflict represents a permission conflict
type PermissionConflict struct {
	AccountName      string
	ConflictType     string
	ConflictingPerms []string
	Resolution       string
}

// QuotaConflict represents a quota conflict
type QuotaConflict struct {
	AccountName       string
	ConflictType      string
	ConflictingQuotas map[string]float64
	Resolution        string
}

// HierarchyConflict represents a hierarchy conflict
type HierarchyConflict struct {
	AccountName      string
	ConflictType     string
	ConflictingPaths []string
	Resolution       string
}

// OrganizationalUnit represents an organizational unit
type OrganizationalUnit struct {
	Name          string
	Accounts      []string
	TotalUsers    int
	ResourceShare float64
}

// AccountHierarchyCollector collects account hierarchy metrics from SLURM
type AccountHierarchyCollector struct {
	client AccountHierarchySLURMClient
	mutex  sync.RWMutex

	// Hierarchy structure metrics
	accountHierarchyDepth    *prometheus.GaugeVec
	accountHierarchyAccounts *prometheus.GaugeVec
	accountHierarchyUsers    *prometheus.GaugeVec
	accountChildCount        *prometheus.GaugeVec
	accountUserCount         *prometheus.GaugeVec
	accountActiveUsers       *prometheus.GaugeVec

	// Association metrics
	userAccountAssociations *prometheus.GaugeVec
	accountAssociationCount *prometheus.GaugeVec
	userDefaultAccount      *prometheus.GaugeVec
	accountAssociationType  *prometheus.GaugeVec
	orphanedUsers           *prometheus.GaugeVec
	multiAccountUsers       *prometheus.GaugeVec

	// Permission metrics
	accountPermissionUsers *prometheus.GaugeVec
	accountEffectivePerms  *prometheus.GaugeVec
	accountInheritedPerms  *prometheus.GaugeVec
	accountCustomRoles     *prometheus.GaugeVec
	permissionConflicts    *prometheus.GaugeVec

	// Inheritance metrics
	accountInheritanceDepth *prometheus.GaugeVec
	accountInheritedProps   *prometheus.GaugeVec
	accountOverriddenProps  *prometheus.GaugeVec
	inheritanceChainLength  *prometheus.GaugeVec

	// Delegation metrics
	accountDelegationsActive  *prometheus.GaugeVec
	accountDelegationsPending *prometheus.GaugeVec
	accountDelegatedRights    *prometheus.GaugeVec

	// Validation metrics
	hierarchyValidationStatus *prometheus.GaugeVec
	hierarchyErrors           *prometheus.GaugeVec
	hierarchyWarnings         *prometheus.GaugeVec
	circularReferences        *prometheus.GaugeVec
	hierarchyConflicts        *prometheus.GaugeVec

	// Access matrix metrics
	accessMatrixPermissions *prometheus.GaugeVec
	accessMatrixConflicts   *prometheus.GaugeVec
	effectivePermissions    *prometheus.GaugeVec

	// Organizational metrics
	organizationalUnitCount *prometheus.GaugeVec
	organizationalUnitUsers *prometheus.GaugeVec
	organizationalUnitShare *prometheus.GaugeVec

	// Relationship metrics
	accountRelationships   *prometheus.GaugeVec
	relationshipStrength   *prometheus.GaugeVec
	bidirectionalRelations *prometheus.GaugeVec

	// Collection metrics
	collectionDuration *prometheus.HistogramVec
	collectionErrors   *prometheus.CounterVec
	lastCollectionTime *prometheus.GaugeVec
}

// NewAccountHierarchyCollector creates a new account hierarchy collector
func NewAccountHierarchyCollector(client AccountHierarchySLURMClient) *AccountHierarchyCollector {
	return &AccountHierarchyCollector{
		client: client,

		// Hierarchy structure metrics
		accountHierarchyDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_hierarchy_depth",
				Help: "Maximum depth of the account hierarchy",
			},
			[]string{"root_account"},
		),
		accountHierarchyAccounts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_hierarchy_accounts_total",
				Help: "Total number of accounts in the hierarchy",
			},
			[]string{"hierarchy_type"},
		),
		accountHierarchyUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_hierarchy_users_total",
				Help: "Total number of users in the hierarchy",
			},
			[]string{"hierarchy_type"},
		),
		accountChildCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_child_count",
				Help: "Number of child accounts for each account",
			},
			[]string{"account", "parent_account"},
		),
		accountUserCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_user_count",
				Help: "Number of users associated with each account",
			},
			[]string{"account", "status"},
		),
		accountActiveUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_active_users",
				Help: "Number of active users in each account",
			},
			[]string{"account"},
		),

		// Association metrics
		userAccountAssociations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_account_associations_total",
				Help: "Total number of user-account associations",
			},
			[]string{"association_type"},
		),
		accountAssociationCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_association_count",
				Help: "Number of associations per account",
			},
			[]string{"account", "association_status"},
		),
		userDefaultAccount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_default_account",
				Help: "Whether this is the user's default account (1 = yes, 0 = no)",
			},
			[]string{"user", "account"},
		),
		accountAssociationType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_association_type_count",
				Help: "Count of associations by type",
			},
			[]string{"account", "type"},
		),
		orphanedUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_orphaned_users_total",
				Help: "Number of users without account associations",
			},
			[]string{},
		),
		multiAccountUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_multi_account_users_total",
				Help: "Number of users with multiple account associations",
			},
			[]string{},
		),

		// Permission metrics
		accountPermissionUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_permission_users",
				Help: "Number of users with specific permissions on account",
			},
			[]string{"account", "permission_level"},
		),
		accountEffectivePerms: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_effective_permissions",
				Help: "Number of effective permissions per account",
			},
			[]string{"account", "permission_type"},
		),
		accountInheritedPerms: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_inherited_permissions",
				Help: "Number of inherited permissions (1 = inherited, 0 = not)",
			},
			[]string{"account", "permission"},
		),
		accountCustomRoles: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_custom_roles_count",
				Help: "Number of custom roles defined per account",
			},
			[]string{"account"},
		),
		permissionConflicts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_permission_conflicts_total",
				Help: "Number of permission conflicts detected",
			},
			[]string{"conflict_type"},
		),

		// Inheritance metrics
		accountInheritanceDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_inheritance_depth",
				Help: "Inheritance chain depth for account",
			},
			[]string{"account"},
		),
		accountInheritedProps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_inherited_properties_count",
				Help: "Number of inherited properties",
			},
			[]string{"account"},
		),
		accountOverriddenProps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_overridden_properties_count",
				Help: "Number of overridden properties",
			},
			[]string{"account"},
		),
		inheritanceChainLength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_inheritance_chain_length",
				Help: "Length of inheritance chain",
			},
			[]string{"account"},
		),

		// Delegation metrics
		accountDelegationsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_delegations_active",
				Help: "Number of active delegations",
			},
			[]string{"account"},
		),
		accountDelegationsPending: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_delegations_pending",
				Help: "Number of pending delegations",
			},
			[]string{"account"},
		),
		accountDelegatedRights: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_delegated_rights_count",
				Help: "Number of delegated rights",
			},
			[]string{"account", "direction"},
		),

		// Validation metrics
		hierarchyValidationStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_hierarchy_validation_status",
				Help: "Hierarchy validation status (1 = valid, 0 = invalid)",
			},
			[]string{},
		),
		hierarchyErrors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_hierarchy_errors_total",
				Help: "Number of hierarchy validation errors",
			},
			[]string{"error_type"},
		),
		hierarchyWarnings: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_hierarchy_warnings_total",
				Help: "Number of hierarchy validation warnings",
			},
			[]string{"warning_type"},
		),
		circularReferences: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_circular_references_total",
				Help: "Number of circular references detected",
			},
			[]string{},
		),
		hierarchyConflicts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_hierarchy_conflicts_total",
				Help: "Number of hierarchy conflicts",
			},
			[]string{"conflict_type", "severity"},
		),

		// Access matrix metrics
		accessMatrixPermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_access_matrix_permissions_total",
				Help: "Total permissions in access matrix",
			},
			[]string{},
		),
		accessMatrixConflicts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_access_matrix_conflicts_total",
				Help: "Number of conflicting permissions in matrix",
			},
			[]string{},
		),
		effectivePermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_effective_permissions_count",
				Help: "Number of effective permissions",
			},
			[]string{"user", "account"},
		),

		// Organizational metrics
		organizationalUnitCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_organizational_unit_count",
				Help: "Number of organizational units",
			},
			[]string{},
		),
		organizationalUnitUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_organizational_unit_users",
				Help: "Number of users per organizational unit",
			},
			[]string{"unit"},
		),
		organizationalUnitShare: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_organizational_unit_resource_share",
				Help: "Resource share percentage for organizational unit",
			},
			[]string{"unit"},
		),

		// Relationship metrics
		accountRelationships: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_relationships_count",
				Help: "Number of relationships per account",
			},
			[]string{"account", "relationship_type"},
		),
		relationshipStrength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_relationship_strength",
				Help: "Strength of account relationships",
			},
			[]string{"account", "related_account", "type"},
		),
		bidirectionalRelations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_bidirectional_relations_count",
				Help: "Number of bidirectional relationships",
			},
			[]string{"account"},
		),

		// Collection metrics
		collectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_account_hierarchy_collection_duration_seconds",
				Help:    "Time spent collecting account hierarchy metrics",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		collectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_account_hierarchy_collection_errors_total",
				Help: "Total number of errors during account hierarchy collection",
			},
			[]string{"operation", "error_type"},
		),
		lastCollectionTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_hierarchy_last_collection_timestamp",
				Help: "Timestamp of last successful collection",
			},
			[]string{"metric_type"},
		),
	}
}

// Describe sends the super-set of all possible descriptors
func (c *AccountHierarchyCollector) Describe(ch chan<- *prometheus.Desc) {
	c.accountHierarchyDepth.Describe(ch)
	c.accountHierarchyAccounts.Describe(ch)
	c.accountHierarchyUsers.Describe(ch)
	c.accountChildCount.Describe(ch)
	c.accountUserCount.Describe(ch)
	c.accountActiveUsers.Describe(ch)
	c.userAccountAssociations.Describe(ch)
	c.accountAssociationCount.Describe(ch)
	c.userDefaultAccount.Describe(ch)
	c.accountAssociationType.Describe(ch)
	c.orphanedUsers.Describe(ch)
	c.multiAccountUsers.Describe(ch)
	c.accountPermissionUsers.Describe(ch)
	c.accountEffectivePerms.Describe(ch)
	c.accountInheritedPerms.Describe(ch)
	c.accountCustomRoles.Describe(ch)
	c.permissionConflicts.Describe(ch)
	c.accountInheritanceDepth.Describe(ch)
	c.accountInheritedProps.Describe(ch)
	c.accountOverriddenProps.Describe(ch)
	c.inheritanceChainLength.Describe(ch)
	c.accountDelegationsActive.Describe(ch)
	c.accountDelegationsPending.Describe(ch)
	c.accountDelegatedRights.Describe(ch)
	c.hierarchyValidationStatus.Describe(ch)
	c.hierarchyErrors.Describe(ch)
	c.hierarchyWarnings.Describe(ch)
	c.circularReferences.Describe(ch)
	c.hierarchyConflicts.Describe(ch)
	c.accessMatrixPermissions.Describe(ch)
	c.accessMatrixConflicts.Describe(ch)
	c.effectivePermissions.Describe(ch)
	c.organizationalUnitCount.Describe(ch)
	c.organizationalUnitUsers.Describe(ch)
	c.organizationalUnitShare.Describe(ch)
	c.accountRelationships.Describe(ch)
	c.relationshipStrength.Describe(ch)
	c.bidirectionalRelations.Describe(ch)
	c.collectionDuration.Describe(ch)
	c.collectionErrors.Describe(ch)
	c.lastCollectionTime.Describe(ch)
}

// Collect fetches the stats and delivers them as Prometheus metrics
func (c *AccountHierarchyCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Reset metrics
	c.accountHierarchyDepth.Reset()
	c.accountHierarchyAccounts.Reset()
	c.accountHierarchyUsers.Reset()
	c.accountChildCount.Reset()
	c.accountUserCount.Reset()
	c.accountActiveUsers.Reset()
	c.userAccountAssociations.Reset()
	c.accountAssociationCount.Reset()
	c.userDefaultAccount.Reset()
	c.accountAssociationType.Reset()
	c.orphanedUsers.Reset()
	c.multiAccountUsers.Reset()
	c.accountPermissionUsers.Reset()
	c.accountEffectivePerms.Reset()
	c.accountInheritedPerms.Reset()
	c.accountCustomRoles.Reset()
	c.permissionConflicts.Reset()
	c.accountInheritanceDepth.Reset()
	c.accountInheritedProps.Reset()
	c.accountOverriddenProps.Reset()
	c.inheritanceChainLength.Reset()
	c.accountDelegationsActive.Reset()
	c.accountDelegationsPending.Reset()
	c.accountDelegatedRights.Reset()
	c.hierarchyValidationStatus.Reset()
	c.hierarchyErrors.Reset()
	c.hierarchyWarnings.Reset()
	c.circularReferences.Reset()
	c.hierarchyConflicts.Reset()
	c.accessMatrixPermissions.Reset()
	c.accessMatrixConflicts.Reset()
	c.effectivePermissions.Reset()
	c.organizationalUnitCount.Reset()
	c.organizationalUnitUsers.Reset()
	c.organizationalUnitShare.Reset()
	c.accountRelationships.Reset()
	c.relationshipStrength.Reset()
	c.bidirectionalRelations.Reset()

	// Collect hierarchy metrics
	c.collectHierarchyMetrics()
	c.collectAssociationMetrics()
	c.collectPermissionMetrics()
	c.collectValidationMetrics()
	c.collectOrganizationalMetrics()

	// Collect all metrics
	c.accountHierarchyDepth.Collect(ch)
	c.accountHierarchyAccounts.Collect(ch)
	c.accountHierarchyUsers.Collect(ch)
	c.accountChildCount.Collect(ch)
	c.accountUserCount.Collect(ch)
	c.accountActiveUsers.Collect(ch)
	c.userAccountAssociations.Collect(ch)
	c.accountAssociationCount.Collect(ch)
	c.userDefaultAccount.Collect(ch)
	c.accountAssociationType.Collect(ch)
	c.orphanedUsers.Collect(ch)
	c.multiAccountUsers.Collect(ch)
	c.accountPermissionUsers.Collect(ch)
	c.accountEffectivePerms.Collect(ch)
	c.accountInheritedPerms.Collect(ch)
	c.accountCustomRoles.Collect(ch)
	c.permissionConflicts.Collect(ch)
	c.accountInheritanceDepth.Collect(ch)
	c.accountInheritedProps.Collect(ch)
	c.accountOverriddenProps.Collect(ch)
	c.inheritanceChainLength.Collect(ch)
	c.accountDelegationsActive.Collect(ch)
	c.accountDelegationsPending.Collect(ch)
	c.accountDelegatedRights.Collect(ch)
	c.hierarchyValidationStatus.Collect(ch)
	c.hierarchyErrors.Collect(ch)
	c.hierarchyWarnings.Collect(ch)
	c.circularReferences.Collect(ch)
	c.hierarchyConflicts.Collect(ch)
	c.accessMatrixPermissions.Collect(ch)
	c.accessMatrixConflicts.Collect(ch)
	c.effectivePermissions.Collect(ch)
	c.organizationalUnitCount.Collect(ch)
	c.organizationalUnitUsers.Collect(ch)
	c.organizationalUnitShare.Collect(ch)
	c.accountRelationships.Collect(ch)
	c.relationshipStrength.Collect(ch)
	c.bidirectionalRelations.Collect(ch)
	c.collectionDuration.Collect(ch)
	c.collectionErrors.Collect(ch)
	c.lastCollectionTime.Collect(ch)
}

func (c *AccountHierarchyCollector) collectHierarchyMetrics() {
	ctx := context.Background()
	start := time.Now()

	hierarchy, err := c.client.GetAccountHierarchy(ctx)
	if err != nil {
		c.collectionErrors.WithLabelValues("hierarchy", "fetch_error").Inc()
		return
	}

	// Process hierarchy structure
	if hierarchy != nil {
		c.accountHierarchyAccounts.WithLabelValues(hierarchy.HierarchyType).Set(float64(hierarchy.TotalAccounts))
		c.accountHierarchyUsers.WithLabelValues(hierarchy.HierarchyType).Set(float64(hierarchy.TotalUsers))

		// Process each root account and its subtree
		for _, root := range hierarchy.RootAccounts {
			c.processAccountNode(root, hierarchy.MaxDepth)
		}

		// Process organizational units
		c.organizationalUnitCount.WithLabelValues().Set(float64(len(hierarchy.OrganizationalUnits)))
		for name, unit := range hierarchy.OrganizationalUnits {
			c.organizationalUnitUsers.WithLabelValues(name).Set(float64(unit.TotalUsers))
			c.organizationalUnitShare.WithLabelValues(name).Set(unit.ResourceShare)
		}
	}

	// Collect sub-accounts for various parent accounts
	c.collectSubAccountMetrics()

	// Collect inheritance metrics
	c.collectInheritanceMetrics()

	// Collect delegation metrics
	c.collectDelegationMetrics()

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("hierarchy").Observe(duration)
	c.lastCollectionTime.WithLabelValues("hierarchy").Set(float64(time.Now().Unix()))
}

func (c *AccountHierarchyCollector) processAccountNode(node *AccountNode, maxDepth int) {
	if node == nil {
		return
	}

	// Set metrics for this account
	c.accountChildCount.WithLabelValues(node.AccountName, node.ParentAccount).Set(float64(len(node.ChildAccounts)))
	c.accountUserCount.WithLabelValues(node.AccountName, node.Status).Set(float64(node.UserCount))
	c.accountActiveUsers.WithLabelValues(node.AccountName).Set(float64(node.ActiveUsers))

	// Set hierarchy depth for root accounts
	if node.ParentAccount == "" {
		c.accountHierarchyDepth.WithLabelValues(node.AccountName).Set(float64(maxDepth))
	}

	// Collect relationships
	ctx := context.Background()
	relationships, err := c.client.GetAccountRelationships(ctx, node.AccountName)
	if err == nil && relationships != nil {
		c.accountRelationships.WithLabelValues(node.AccountName, "parent").Set(float64(len(relationships.ParentRelations)))
		c.accountRelationships.WithLabelValues(node.AccountName, "child").Set(float64(len(relationships.ChildRelations)))
		c.accountRelationships.WithLabelValues(node.AccountName, "peer").Set(float64(len(relationships.PeerRelations)))

		// Process relationship strengths
		for _, rel := range relationships.ParentRelations {
			c.relationshipStrength.WithLabelValues(node.AccountName, rel.ParentAccount, "parent").Set(rel.Strength)
		}
		for _, rel := range relationships.ChildRelations {
			c.relationshipStrength.WithLabelValues(node.AccountName, rel.ChildAccount, "child").Set(rel.Strength)
		}

		bidirectionalCount := 0
		for _, rel := range relationships.PeerRelations {
			c.relationshipStrength.WithLabelValues(node.AccountName, rel.PeerAccount, "peer").Set(rel.Strength)
			if rel.Bidirectional {
				bidirectionalCount++
			}
		}
		c.bidirectionalRelations.WithLabelValues(node.AccountName).Set(float64(bidirectionalCount))
	}
}

func (c *AccountHierarchyCollector) collectAssociationMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Get user-account mapping
	mapping, err := c.client.GetUserAccountMapping(ctx)
	if err != nil {
		c.collectionErrors.WithLabelValues("associations", "mapping_error").Inc()
		return
	}

	if mapping != nil {
		c.orphanedUsers.WithLabelValues().Set(float64(len(mapping.OrphanedUsers)))
		c.multiAccountUsers.WithLabelValues().Set(float64(mapping.UsersWithMultiple))

		for associationType, count := range mapping.MappingsByType {
			c.userAccountAssociations.WithLabelValues(associationType).Set(float64(count))
		}
	}

	// Sample accounts for detailed association metrics
	sampleAccounts := []string{"research", "engineering", "finance", "admin"}
	for _, accountName := range sampleAccounts {
		associations, err := c.client.GetAccountAssociations(ctx, accountName)
		if err != nil {
			continue
		}

		if associations != nil {
			c.accountAssociationCount.WithLabelValues(accountName, "active").Set(float64(associations.ActiveAssociations))
			c.accountAssociationCount.WithLabelValues(accountName, "pending").Set(float64(associations.PendingAssociations))
			c.accountAssociationCount.WithLabelValues(accountName, "total").Set(float64(associations.TotalAssociations))

			// Count association types
			typeCount := map[string]int{
				"parent":   len(associations.ParentAssociations),
				"child":    len(associations.ChildAssociations),
				"peer":     len(associations.PeerAssociations),
				"external": len(associations.ExternalAssociations),
			}

			for assocType, count := range typeCount {
				c.accountAssociationType.WithLabelValues(accountName, assocType).Set(float64(count))
			}

			// Process user associations
			for _, user := range associations.Users {
				defaultValue := 0.0
				if user.DefaultAccount {
					defaultValue = 1.0
				}
				c.userDefaultAccount.WithLabelValues(user.UserName, accountName).Set(defaultValue)
			}
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("associations").Observe(duration)
	c.lastCollectionTime.WithLabelValues("associations").Set(float64(time.Now().Unix()))
}

func (c *AccountHierarchyCollector) collectPermissionMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample accounts for permission metrics
	sampleAccounts := []string{"research", "engineering", "finance", "admin"}
	for _, accountName := range sampleAccounts {
		permissions, err := c.client.GetAccountPermissions(ctx, accountName)
		if err != nil {
			continue
		}

		if permissions != nil {
			c.accountPermissionUsers.WithLabelValues(accountName, "admin").Set(float64(len(permissions.AdminUsers)))
			c.accountPermissionUsers.WithLabelValues(accountName, "operator").Set(float64(len(permissions.OperatorUsers)))
			c.accountPermissionUsers.WithLabelValues(accountName, "view_only").Set(float64(len(permissions.ViewOnlyUsers)))
			c.accountPermissionUsers.WithLabelValues(accountName, "submit").Set(float64(len(permissions.SubmitUsers)))

			c.accountCustomRoles.WithLabelValues(accountName).Set(float64(len(permissions.CustomRoles)))

			// Count inherited permissions
			for perm, inherited := range permissions.InheritedPerms {
				value := 0.0
				if inherited {
					value = 1.0
				}
				c.accountInheritedPerms.WithLabelValues(accountName, perm).Set(value)
			}

			// Count effective permissions by type
			for permType, users := range permissions.EffectivePerms {
				c.accountEffectivePerms.WithLabelValues(accountName, permType).Set(float64(len(users)))
			}
		}
	}

	// Get access matrix for conflict detection
	matrix, err := c.client.GetAccountAccessMatrix(ctx)
	if err == nil && matrix != nil {
		c.accessMatrixPermissions.WithLabelValues().Set(float64(matrix.TotalPermissions))
		c.accessMatrixConflicts.WithLabelValues().Set(float64(matrix.ConflictingPerms))

		// Sample effective permissions
		for user, accounts := range matrix.EffectivePerms {
			for account, hasPerm := range accounts {
				permCount := 0
				if hasPerm {
					permCount = 1
				}
				c.effectivePermissions.WithLabelValues(user, account).Set(float64(permCount))
			}
		}
	}

	// Get conflicts
	conflicts, err := c.client.GetAccountConflicts(ctx)
	if err == nil && conflicts != nil {
		c.permissionConflicts.WithLabelValues("permission").Set(float64(len(conflicts.PermissionConflicts)))
		c.permissionConflicts.WithLabelValues("quota").Set(float64(len(conflicts.QuotaConflicts)))
		c.permissionConflicts.WithLabelValues("hierarchy").Set(float64(len(conflicts.HierarchyConflicts)))
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("permissions").Observe(duration)
	c.lastCollectionTime.WithLabelValues("permissions").Set(float64(time.Now().Unix()))
}

func (c *AccountHierarchyCollector) collectValidationMetrics() {
	ctx := context.Background()
	start := time.Now()

	validation, err := c.client.GetHierarchyValidation(ctx)
	if err != nil {
		c.collectionErrors.WithLabelValues("validation", "validation_error").Inc()
		return
	}

	if validation != nil {
		validValue := 0.0
		if validation.Valid {
			validValue = 1.0
		}
		c.hierarchyValidationStatus.WithLabelValues().Set(validValue)

		// Count errors by type
		errorTypes := make(map[string]int)
		for _, err := range validation.Errors {
			// Simple categorization based on error message
			if hierarchyContains(err, "circular") {
				errorTypes["circular"]++
			} else if hierarchyContains(err, "orphan") {
				errorTypes["orphan"]++
			} else if hierarchyContains(err, "conflict") {
				errorTypes["conflict"]++
			} else {
				errorTypes["other"]++
			}
		}

		for errorType, count := range errorTypes {
			c.hierarchyErrors.WithLabelValues(errorType).Set(float64(count))
		}

		// Count warnings similarly
		warningTypes := make(map[string]int)
		for _, warning := range validation.Warnings {
			if hierarchyContains(warning, "performance") {
				warningTypes["performance"]++
			} else if hierarchyContains(warning, "security") {
				warningTypes["security"]++
			} else {
				warningTypes["other"]++
			}
		}

		for warningType, count := range warningTypes {
			c.hierarchyWarnings.WithLabelValues(warningType).Set(float64(count))
		}

		c.circularReferences.WithLabelValues().Set(float64(len(validation.CircularReferences)))
	}

	// Get conflicts
	conflicts, err := c.client.GetAccountConflicts(ctx)
	if err == nil && conflicts != nil {
		c.hierarchyConflicts.WithLabelValues("total", "all").Set(float64(conflicts.TotalConflicts))
		c.hierarchyConflicts.WithLabelValues("critical", "high").Set(float64(conflicts.CriticalConflicts))
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("validation").Observe(duration)
	c.lastCollectionTime.WithLabelValues("validation").Set(float64(time.Now().Unix()))
}

func (c *AccountHierarchyCollector) collectSubAccountMetrics() {
	ctx := context.Background()

	// Sample parent accounts
	parentAccounts := []string{"root", "research", "engineering"}
	for _, parent := range parentAccounts {
		subAccounts, err := c.client.GetAccountSubAccounts(ctx, parent)
		if err != nil {
			continue
		}

		for _, sub := range subAccounts {
			// Process sub-account metrics
			if sub.InheritedQuotas {
				c.accountInheritedProps.WithLabelValues(sub.AccountName).Inc()
			}
			if sub.InheritedQoS {
				c.accountInheritedProps.WithLabelValues(sub.AccountName).Inc()
			}
			if sub.InheritedUsers {
				c.accountInheritedProps.WithLabelValues(sub.AccountName).Inc()
			}

			c.accountOverriddenProps.WithLabelValues(sub.AccountName).Set(float64(len(sub.OverrideSettings)))
		}
	}
}

func (c *AccountHierarchyCollector) collectInheritanceMetrics() {
	ctx := context.Background()

	// Sample accounts for inheritance metrics
	sampleAccounts := []string{"research", "engineering", "finance"}
	for _, accountName := range sampleAccounts {
		inheritance, err := c.client.GetAccountInheritance(ctx, accountName)
		if err != nil {
			continue
		}

		if inheritance != nil {
			c.accountInheritanceDepth.WithLabelValues(accountName).Set(float64(len(inheritance.InheritedFrom)))
			c.inheritanceChainLength.WithLabelValues(accountName).Set(float64(len(inheritance.InheritanceChain)))
			c.accountInheritedProps.WithLabelValues(accountName).Set(float64(len(inheritance.InheritedProperties)))
			c.accountOverriddenProps.WithLabelValues(accountName).Set(float64(len(inheritance.OverriddenProperties)))
		}
	}
}

func (c *AccountHierarchyCollector) collectDelegationMetrics() {
	ctx := context.Background()

	// Sample accounts for delegation metrics
	sampleAccounts := []string{"research", "engineering", "admin"}
	for _, accountName := range sampleAccounts {
		delegations, err := c.client.GetAccountDelegations(ctx, accountName)
		if err != nil {
			continue
		}

		if delegations != nil {
			c.accountDelegationsActive.WithLabelValues(accountName).Set(float64(delegations.ActiveDelegations))
			c.accountDelegationsPending.WithLabelValues(accountName).Set(float64(delegations.PendingDelegations))

			delegatedToCount := 0
			for _, rights := range delegations.DelegatedTo {
				delegatedToCount += len(rights)
			}
			c.accountDelegatedRights.WithLabelValues(accountName, "to").Set(float64(delegatedToCount))

			delegatedFromCount := 0
			for _, rights := range delegations.DelegatedFrom {
				delegatedFromCount += len(rights)
			}
			c.accountDelegatedRights.WithLabelValues(accountName, "from").Set(float64(delegatedFromCount))
		}
	}
}

func (c *AccountHierarchyCollector) collectOrganizationalMetrics() {
	// Organizational metrics are collected as part of hierarchy collection
	// This method is a placeholder for any additional organizational-specific metrics
}

// Helper function to check if a string contains a substring
func hierarchyContains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && hierarchyContainsMiddle(s, substr)
}

func hierarchyContainsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
