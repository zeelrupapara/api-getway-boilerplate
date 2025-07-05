package authz

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/db"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// Service represents the Casbin RBAC authorization service
type Service struct {
	Enforcer *casbin.Enforcer
	Logger   *logger.Logger
	DB       *db.Database
}

// CannabisPermission represents a cannabis-specific permission
type CannabisPermission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Object   string `json:"object,omitempty"`
}

// Cannabis platform resources
const (
	// User management resources
	ResourceUsers     = "users"
	ResourceProfiles  = "profiles"
	ResourceSessions  = "sessions"
	
	// Cannabis business resources
	ResourceProducts     = "products"
	ResourceInventory    = "inventory"
	ResourceOrders       = "orders"
	ResourceTransactions = "transactions"
	ResourceDispensaries = "dispensaries"
	
	// Cannabis compliance resources
	ResourceCompliance   = "compliance"
	ResourceAgeVerification = "age_verification"
	ResourceStateVerification = "state_verification"
	ResourceAuditLogs    = "audit_logs"
	
	// Cannabis business operations
	ResourcePOS          = "pos"
	ResourceReports      = "reports"
	ResourceAnalytics    = "analytics"
	ResourceMarketing    = "marketing"
	
	// System administration
	ResourceSystem       = "system"
	ResourceConfiguration = "configuration"
	ResourceRoles        = "roles"
	ResourcePermissions  = "permissions"
)

// Cannabis platform actions
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"
	ActionView   = "view"
	ActionManage = "manage"
	ActionProcess = "process"
	ActionVerify = "verify"
	ActionApprove = "approve"
	ActionReject = "reject"
	ActionExport = "export"
	ActionImport = "import"
	ActionExecute = "execute"
)

// NewService creates a new Casbin authorization service
func NewService(database *db.Database, log *logger.Logger) (*Service, error) {
	log.Info("Initializing Casbin RBAC authorization service...")

	// Create Gorm adapter for Casbin
	adapter, err := gormadapter.NewAdapterByDB(database.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Create Casbin model
	casbinModel := model.NewModel()
	casbinModel.AddDef("r", "r", "sub, obj, act")
	casbinModel.AddDef("p", "p", "sub, obj, act")
	casbinModel.AddDef("g", "g", "_, _")
	casbinModel.AddDef("e", "e", "some(where (p.eft == allow))")
	casbinModel.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	// Create Casbin enforcer
	enforcer, err := casbin.NewEnforcer(casbinModel, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Enable auto-save
	enforcer.EnableAutoSave(true)

	service := &Service{
		Enforcer: enforcer,
		Logger:   log,
		DB:       database,
	}

	// Initialize cannabis roles and policies
	if err := service.initializeCannabisRBAC(); err != nil {
		log.Warn("Failed to initialize cannabis RBAC", "error", err)
	}

	log.Info("Casbin authorization service initialized successfully")
	return service, nil
}

// Cannabis RBAC initialization

// initializeCannabisRBAC sets up cannabis-specific roles and policies
func (s *Service) initializeCannabisRBAC() error {
	s.Logger.Info("Initializing cannabis RBAC roles and policies...")

	// Define cannabis roles and their permissions
	roles := s.getCannabisRoleDefinitions()

	// Clear existing policies (for fresh setup)
	if _, err := s.Enforcer.RemoveFilteredPolicy(0); err != nil {
		s.Logger.Warn("Failed to clear existing policies", "error", err)
	}

	// Add role-based policies
	for role, permissions := range roles {
		for _, permission := range permissions {
			if _, err := s.Enforcer.AddPolicy(string(role), permission.Resource, permission.Action); err != nil {
				s.Logger.Warn("Failed to add policy",
					"role", role,
					"resource", permission.Resource,
					"action", permission.Action,
					"error", err,
				)
			}
		}
	}

	// Setup role hierarchy (inheritance)
	if err := s.setupRoleHierarchy(); err != nil {
		return fmt.Errorf("failed to setup role hierarchy: %w", err)
	}

	// Save policies
	if err := s.Enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save RBAC policies: %w", err)
	}

	s.Logger.Info("Cannabis RBAC initialization completed")
	return nil
}

// getCannabisRoleDefinitions returns cannabis-specific role definitions
func (s *Service) getCannabisRoleDefinitions() map[v1.UserRole][]CannabisPermission {
	return map[v1.UserRole][]CannabisPermission{
		v1.RoleCustomer: {
			// Profile management
			{ResourceProfiles, ActionRead},
			{ResourceProfiles, ActionUpdate},
			{ResourceSessions, ActionRead},
			
			// Product viewing
			{ResourceProducts, ActionRead},
			{ResourceProducts, ActionList},
			{ResourceProducts, ActionView},
			
			// Order management
			{ResourceOrders, ActionCreate},
			{ResourceOrders, ActionRead},
			{ResourceOrders, ActionList},
			{ResourceOrders, ActionUpdate}, // For cancellation
			
			// Transaction viewing
			{ResourceTransactions, ActionRead},
			{ResourceTransactions, ActionList},
			
			// Age verification (self)
			{ResourceAgeVerification, ActionCreate},
			{ResourceAgeVerification, ActionRead},
			{ResourceStateVerification, ActionCreate},
			{ResourceStateVerification, ActionRead},
		},
		
		v1.RoleBudtender: {
			// Customer interaction
			{ResourceUsers, ActionRead},
			{ResourceUsers, ActionList},
			{ResourceProfiles, ActionRead},
			
			// Product management (limited)
			{ResourceProducts, ActionRead},
			{ResourceProducts, ActionList},
			{ResourceProducts, ActionView},
			{ResourceInventory, ActionRead},
			{ResourceInventory, ActionUpdate}, // For sales
			
			// Order processing
			{ResourceOrders, ActionRead},
			{ResourceOrders, ActionList},
			{ResourceOrders, ActionUpdate},
			{ResourceOrders, ActionProcess},
			
			// POS operations
			{ResourcePOS, ActionRead},
			{ResourcePOS, ActionProcess},
			
			// Transaction processing
			{ResourceTransactions, ActionCreate},
			{ResourceTransactions, ActionRead},
			{ResourceTransactions, ActionProcess},
			
			// Basic compliance verification
			{ResourceAgeVerification, ActionVerify},
			{ResourceStateVerification, ActionVerify},
			{ResourceCompliance, ActionRead},
			
			// Basic reporting
			{ResourceReports, ActionRead},
		},
		
		v1.RoleDispensaryManager: {
			// User management (dispensary level)
			{ResourceUsers, ActionCreate},
			{ResourceUsers, ActionRead},
			{ResourceUsers, ActionUpdate},
			{ResourceUsers, ActionList},
			{ResourceUsers, ActionManage},
			{ResourceProfiles, ActionRead},
			{ResourceProfiles, ActionUpdate},
			
			// Product management
			{ResourceProducts, ActionCreate},
			{ResourceProducts, ActionRead},
			{ResourceProducts, ActionUpdate},
			{ResourceProducts, ActionDelete},
			{ResourceProducts, ActionList},
			{ResourceProducts, ActionManage},
			
			// Inventory management
			{ResourceInventory, ActionCreate},
			{ResourceInventory, ActionRead},
			{ResourceInventory, ActionUpdate},
			{ResourceInventory, ActionDelete},
			{ResourceInventory, ActionList},
			{ResourceInventory, ActionManage},
			{ResourceInventory, ActionImport},
			{ResourceInventory, ActionExport},
			
			// Order management
			{ResourceOrders, ActionCreate},
			{ResourceOrders, ActionRead},
			{ResourceOrders, ActionUpdate},
			{ResourceOrders, ActionDelete},
			{ResourceOrders, ActionList},
			{ResourceOrders, ActionManage},
			{ResourceOrders, ActionProcess},
			{ResourceOrders, ActionApprove},
			{ResourceOrders, ActionReject},
			
			// Transaction management
			{ResourceTransactions, ActionRead},
			{ResourceTransactions, ActionList},
			{ResourceTransactions, ActionManage},
			{ResourceTransactions, ActionExport},
			
			// Dispensary management
			{ResourceDispensaries, ActionRead},
			{ResourceDispensaries, ActionUpdate},
			{ResourceDispensaries, ActionManage},
			
			// POS management
			{ResourcePOS, ActionRead},
			{ResourcePOS, ActionManage},
			{ResourcePOS, ActionProcess},
			
			// Compliance management
			{ResourceCompliance, ActionRead},
			{ResourceCompliance, ActionManage},
			{ResourceAgeVerification, ActionRead},
			{ResourceAgeVerification, ActionVerify},
			{ResourceAgeVerification, ActionManage},
			{ResourceStateVerification, ActionRead},
			{ResourceStateVerification, ActionVerify},
			{ResourceStateVerification, ActionManage},
			{ResourceAuditLogs, ActionRead},
			{ResourceAuditLogs, ActionList},
			{ResourceAuditLogs, ActionExport},
			
			// Reporting and analytics
			{ResourceReports, ActionRead},
			{ResourceReports, ActionCreate},
			{ResourceReports, ActionExport},
			{ResourceAnalytics, ActionRead},
			{ResourceAnalytics, ActionView},
			
			// Marketing
			{ResourceMarketing, ActionRead},
			{ResourceMarketing, ActionCreate},
			{ResourceMarketing, ActionUpdate},
			{ResourceMarketing, ActionManage},
			
			// Role management (dispensary level)
			{ResourceRoles, ActionRead},
			{ResourceRoles, ActionList},
		},
		
		v1.RoleBrandPartner: {
			// Product management (brand products only)
			{ResourceProducts, ActionCreate},
			{ResourceProducts, ActionRead},
			{ResourceProducts, ActionUpdate},
			{ResourceProducts, ActionList},
			{ResourceProducts, ActionView},
			
			// Inventory (brand products only)
			{ResourceInventory, ActionRead},
			{ResourceInventory, ActionUpdate},
			{ResourceInventory, ActionList},
			{ResourceInventory, ActionView},
			
			// Order viewing (brand products only)
			{ResourceOrders, ActionRead},
			{ResourceOrders, ActionList},
			{ResourceOrders, ActionView},
			
			// Transaction viewing (brand products only)
			{ResourceTransactions, ActionRead},
			{ResourceTransactions, ActionList},
			{ResourceTransactions, ActionView},
			
			// Brand reporting
			{ResourceReports, ActionRead},
			{ResourceReports, ActionView},
			{ResourceAnalytics, ActionRead},
			{ResourceAnalytics, ActionView},
			
			// Marketing for brand
			{ResourceMarketing, ActionRead},
			{ResourceMarketing, ActionCreate},
			{ResourceMarketing, ActionUpdate},
			{ResourceMarketing, ActionManage},
		},
		
		v1.RoleSystemAdmin: {
			// Full system access - all resources and actions
			{ResourceUsers, ActionCreate},
			{ResourceUsers, ActionRead},
			{ResourceUsers, ActionUpdate},
			{ResourceUsers, ActionDelete},
			{ResourceUsers, ActionList},
			{ResourceUsers, ActionManage},
			
			{ResourceProducts, ActionCreate},
			{ResourceProducts, ActionRead},
			{ResourceProducts, ActionUpdate},
			{ResourceProducts, ActionDelete},
			{ResourceProducts, ActionList},
			{ResourceProducts, ActionManage},
			
			{ResourceInventory, ActionCreate},
			{ResourceInventory, ActionRead},
			{ResourceInventory, ActionUpdate},
			{ResourceInventory, ActionDelete},
			{ResourceInventory, ActionList},
			{ResourceInventory, ActionManage},
			{ResourceInventory, ActionImport},
			{ResourceInventory, ActionExport},
			
			{ResourceOrders, ActionCreate},
			{ResourceOrders, ActionRead},
			{ResourceOrders, ActionUpdate},
			{ResourceOrders, ActionDelete},
			{ResourceOrders, ActionList},
			{ResourceOrders, ActionManage},
			{ResourceOrders, ActionProcess},
			{ResourceOrders, ActionApprove},
			{ResourceOrders, ActionReject},
			
			{ResourceTransactions, ActionCreate},
			{ResourceTransactions, ActionRead},
			{ResourceTransactions, ActionUpdate},
			{ResourceTransactions, ActionDelete},
			{ResourceTransactions, ActionList},
			{ResourceTransactions, ActionManage},
			{ResourceTransactions, ActionExport},
			
			{ResourceDispensaries, ActionCreate},
			{ResourceDispensaries, ActionRead},
			{ResourceDispensaries, ActionUpdate},
			{ResourceDispensaries, ActionDelete},
			{ResourceDispensaries, ActionList},
			{ResourceDispensaries, ActionManage},
			
			{ResourceCompliance, ActionCreate},
			{ResourceCompliance, ActionRead},
			{ResourceCompliance, ActionUpdate},
			{ResourceCompliance, ActionDelete},
			{ResourceCompliance, ActionList},
			{ResourceCompliance, ActionManage},
			
			{ResourceAuditLogs, ActionRead},
			{ResourceAuditLogs, ActionList},
			{ResourceAuditLogs, ActionExport},
			{ResourceAuditLogs, ActionManage},
			
			{ResourceReports, ActionCreate},
			{ResourceReports, ActionRead},
			{ResourceReports, ActionUpdate},
			{ResourceReports, ActionDelete},
			{ResourceReports, ActionList},
			{ResourceReports, ActionManage},
			{ResourceReports, ActionExport},
			
			{ResourceAnalytics, ActionRead},
			{ResourceAnalytics, ActionView},
			{ResourceAnalytics, ActionManage},
			
			{ResourceSystem, ActionRead},
			{ResourceSystem, ActionUpdate},
			{ResourceSystem, ActionManage},
			{ResourceSystem, ActionExecute},
			
			{ResourceConfiguration, ActionRead},
			{ResourceConfiguration, ActionUpdate},
			{ResourceConfiguration, ActionManage},
			
			{ResourceRoles, ActionCreate},
			{ResourceRoles, ActionRead},
			{ResourceRoles, ActionUpdate},
			{ResourceRoles, ActionDelete},
			{ResourceRoles, ActionList},
			{ResourceRoles, ActionManage},
			
			{ResourcePermissions, ActionCreate},
			{ResourcePermissions, ActionRead},
			{ResourcePermissions, ActionUpdate},
			{ResourcePermissions, ActionDelete},
			{ResourcePermissions, ActionList},
			{ResourcePermissions, ActionManage},
		},
	}
}

// setupRoleHierarchy sets up role inheritance
func (s *Service) setupRoleHierarchy() error {
	// Define role inheritance (child inherits from parent)
	hierarchy := map[v1.UserRole][]v1.UserRole{
		v1.RoleDispensaryManager: {v1.RoleBudtender}, // Manager inherits budtender permissions
		v1.RoleSystemAdmin:       {v1.RoleDispensaryManager, v1.RoleBrandPartner}, // Admin inherits all
	}

	for role, inherits := range hierarchy {
		for _, inheritRole := range inherits {
			if _, err := s.Enforcer.AddGroupingPolicy(string(role), string(inheritRole)); err != nil {
				return fmt.Errorf("failed to add role inheritance %s -> %s: %w", role, inheritRole, err)
			}
		}
	}

	return nil
}

// Authorization methods

// Enforce checks if a user has permission to perform an action on a resource
func (s *Service) Enforce(userRole v1.UserRole, resource, action string) (bool, error) {
	allowed, err := s.Enforcer.Enforce(string(userRole), resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to enforce authorization: %w", err)
	}

	// Log authorization check for cannabis compliance
	s.Logger.LogCannabisAudit("", "authorization_check", "rbac", map[string]interface{}{
		"role":     userRole,
		"resource": resource,
		"action":   action,
		"allowed":  allowed,
	})

	return allowed, nil
}

// EnforceWithObject checks permission with specific object context
func (s *Service) EnforceWithObject(userRole v1.UserRole, resource, action, object string) (bool, error) {
	// For object-level permissions, combine resource and object
	resourceWithObject := fmt.Sprintf("%s:%s", resource, object)
	
	allowed, err := s.Enforcer.Enforce(string(userRole), resourceWithObject, action)
	if err != nil {
		return false, fmt.Errorf("failed to enforce object authorization: %w", err)
	}

	// If object-specific permission not found, check general permission
	if !allowed {
		allowed, err = s.Enforce(userRole, resource, action)
		if err != nil {
			return false, err
		}
	}

	// Log authorization check with object context
	s.Logger.LogCannabisAudit("", "authorization_check_object", "rbac", map[string]interface{}{
		"role":     userRole,
		"resource": resource,
		"action":   action,
		"object":   object,
		"allowed":  allowed,
	})

	return allowed, nil
}

// Cannabis-specific authorization methods

// CanAccessCannabisResource checks if user can access cannabis-specific resources
func (s *Service) CanAccessCannabisResource(userRole v1.UserRole, resource, action string) (bool, error) {
	// All cannabis resources require basic compliance
	if !s.isCannabisResource(resource) {
		return s.Enforce(userRole, resource, action)
	}

	// Check basic permission first
	allowed, err := s.Enforce(userRole, resource, action)
	if err != nil || !allowed {
		return allowed, err
	}

	// Additional cannabis-specific checks
	if resource == ResourceCompliance && action == ActionManage {
		// Only managers and above can manage compliance
		return userRole == v1.RoleDispensaryManager || userRole == v1.RoleSystemAdmin, nil
	}

	if resource == ResourceAuditLogs {
		// Only managers and above can access audit logs
		return userRole == v1.RoleDispensaryManager || userRole == v1.RoleSystemAdmin, nil
	}

	return allowed, nil
}

// CanAccessDispensaryResource checks dispensary-specific access
func (s *Service) CanAccessDispensaryResource(userRole v1.UserRole, userDispensaryID, resourceDispensaryID, resource, action string) (bool, error) {
	// System admin can access all dispensaries
	if userRole == v1.RoleSystemAdmin {
		return s.Enforce(userRole, resource, action)
	}

	// Users can only access resources from their dispensary
	if userDispensaryID != resourceDispensaryID {
		s.Logger.LogCannabisAudit("", "dispensary_access_denied", "rbac", map[string]interface{}{
			"role":                   userRole,
			"user_dispensary":        userDispensaryID,
			"resource_dispensary":    resourceDispensaryID,
			"resource":              resource,
			"action":                action,
		})
		return false, nil
	}

	return s.Enforce(userRole, resource, action)
}

// Policy management methods

// AddCannabisPolicy adds a new cannabis-specific policy
func (s *Service) AddCannabisPolicy(role v1.UserRole, resource, action string) error {
	if _, err := s.Enforcer.AddPolicy(string(role), resource, action); err != nil {
		return fmt.Errorf("failed to add cannabis policy: %w", err)
	}

	// Log policy addition for audit
	s.Logger.LogCannabisAudit("", "policy_added", "rbac", map[string]interface{}{
		"role":     role,
		"resource": resource,
		"action":   action,
	})

	return nil
}

// RemoveCannabisPolicy removes a cannabis-specific policy
func (s *Service) RemoveCannabisPolicy(role v1.UserRole, resource, action string) error {
	if _, err := s.Enforcer.RemovePolicy(string(role), resource, action); err != nil {
		return fmt.Errorf("failed to remove cannabis policy: %w", err)
	}

	// Log policy removal for audit
	s.Logger.LogCannabisAudit("", "policy_removed", "rbac", map[string]interface{}{
		"role":     role,
		"resource": resource,
		"action":   action,
	})

	return nil
}

// GetRolePermissions returns all permissions for a role
func (s *Service) GetRolePermissions(role v1.UserRole) ([]CannabisPermission, error) {
	policies := s.Enforcer.GetFilteredPolicy(0, string(role))
	
	permissions := make([]CannabisPermission, 0, len(policies))
	for _, policy := range policies {
		if len(policy) >= 3 {
			permissions = append(permissions, CannabisPermission{
				Resource: policy[1],
				Action:   policy[2],
			})
		}
	}

	return permissions, nil
}

// Utility methods

// isCannabisResource checks if resource is cannabis-specific
func (s *Service) isCannabisResource(resource string) bool {
	cannabisResources := map[string]bool{
		ResourceProducts:          true,
		ResourceInventory:         true,
		ResourceOrders:           true,
		ResourceTransactions:     true,
		ResourceDispensaries:     true,
		ResourceCompliance:       true,
		ResourceAgeVerification:  true,
		ResourceStateVerification: true,
		ResourceAuditLogs:        true,
		ResourcePOS:              true,
		ResourceReports:          true,
		ResourceAnalytics:        true,
		ResourceMarketing:        true,
	}
	
	return cannabisResources[resource]
}

// GetCannabisPermissionString creates a permission string for cannabis operations
func GetCannabisPermissionString(resource, action string) string {
	return fmt.Sprintf("%s:%s", resource, action)
}

// ValidateRoleForCannabisOperation validates if role can perform cannabis operation
func (s *Service) ValidateRoleForCannabisOperation(role v1.UserRole, operation string) bool {
	operationPermissions := map[string][]v1.UserRole{
		"purchase": {v1.RoleCustomer},
		"sell": {v1.RoleBudtender, v1.RoleDispensaryManager, v1.RoleSystemAdmin},
		"manage_inventory": {v1.RoleDispensaryManager, v1.RoleSystemAdmin},
		"verify_age": {v1.RoleBudtender, v1.RoleDispensaryManager, v1.RoleSystemAdmin},
		"verify_state": {v1.RoleBudtender, v1.RoleDispensaryManager, v1.RoleSystemAdmin},
		"view_audit_logs": {v1.RoleDispensaryManager, v1.RoleSystemAdmin},
		"manage_compliance": {v1.RoleDispensaryManager, v1.RoleSystemAdmin},
		"system_admin": {v1.RoleSystemAdmin},
	}

	allowedRoles, exists := operationPermissions[operation]
	if !exists {
		return false
	}

	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}

	return false
}