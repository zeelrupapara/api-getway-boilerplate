package handlers

import (
	"github.com/gofiber/fiber/v2"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/authz"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/messaging"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/internal/middleware"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// RBACHandler handles RBAC management requests
type RBACHandler struct {
	AuthzService *authz.Service
	Logger       *logger.Logger
	Messaging    *messaging.NATS
}

// PermissionCheckRequest represents permission check request
type PermissionCheckRequest struct {
	Role     string `json:"role" validate:"required"`
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
	Object   string `json:"object,omitempty"`
}

// PolicyRequest represents policy management request
type PolicyRequest struct {
	Role     string `json:"role" validate:"required"`
	Resource string `json:"resource" validate:"required"`
	Action   string `json:"action" validate:"required"`
}

// RolePermissionsResponse represents role permissions response
type RolePermissionsResponse struct {
	Role        string                      `json:"role"`
	Permissions []authz.CannabisPermission  `json:"permissions"`
	Total       int                         `json:"total"`
}

// NewRBACHandler creates a new RBAC handler
func NewRBACHandler(authzService *authz.Service, logger *logger.Logger, messaging *messaging.NATS) *RBACHandler {
	return &RBACHandler{
		AuthzService: authzService,
		Logger:       logger,
		Messaging:    messaging,
	}
}

// CheckPermission checks if a role has permission for a resource/action
func (h *RBACHandler) CheckPermission(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req PermissionCheckRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate role
	role := v1.UserRole(req.Role)
	if !h.isValidRole(role) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role",
			"valid_roles": h.getValidRoles(),
		})
	}

	// Check permission
	var allowed bool
	var err error
	
	if req.Object != "" {
		allowed, err = h.AuthzService.EnforceWithObject(role, req.Resource, req.Action, req.Object)
	} else {
		allowed, err = h.AuthzService.Enforce(role, req.Resource, req.Action)
	}

	if err != nil {
		h.Logger.Error("Permission check failed",
			"error", err,
			"role", req.Role,
			"resource", req.Resource,
			"action", req.Action,
			"object", req.Object,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Permission check failed",
		})
	}

	// Log permission check for audit
	h.Logger.LogCannabisAudit(authCtx.UserID.String(), "permission_checked", "rbac", map[string]interface{}{
		"checked_role": req.Role,
		"resource":     req.Resource,
		"action":       req.Action,
		"object":       req.Object,
		"allowed":      allowed,
		"checker_role": authCtx.Role,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"permission_check": map[string]interface{}{
			"role":     req.Role,
			"resource": req.Resource,
			"action":   req.Action,
			"object":   req.Object,
			"allowed":  allowed,
		},
		"checked_at": c.Context().Value("start_time"),
	})
}

// GetRolePermissions returns all permissions for a specific role
func (h *RBACHandler) GetRolePermissions(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	roleStr := c.Params("role")
	if roleStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Role parameter required",
		})
	}

	role := v1.UserRole(roleStr)
	if !h.isValidRole(role) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role",
			"valid_roles": h.getValidRoles(),
		})
	}

	// Get permissions for role
	permissions, err := h.AuthzService.GetRolePermissions(role)
	if err != nil {
		h.Logger.Error("Failed to get role permissions",
			"error", err,
			"role", roleStr,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve role permissions",
		})
	}

	// Log role permissions access for audit
	h.Logger.LogCannabisAudit(authCtx.UserID.String(), "role_permissions_accessed", "rbac", map[string]interface{}{
		"target_role":      roleStr,
		"permissions_count": len(permissions),
		"accessor_role":    authCtx.Role,
	})

	response := RolePermissionsResponse{
		Role:        roleStr,
		Permissions: permissions,
		Total:       len(permissions),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetAllRoles returns all available cannabis roles and their descriptions
func (h *RBACHandler) GetAllRoles(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	roles := map[string]interface{}{
		string(v1.RoleCustomer): map[string]interface{}{
			"name":        "Customer",
			"description": "Cannabis customers who can browse products and place orders",
			"level":       1,
			"cannabis_permissions": []string{
				"Browse products", "Place orders", "View order history",
				"Manage profile", "Age verification", "State verification",
			},
		},
		string(v1.RoleBudtender): map[string]interface{}{
			"name":        "Budtender",
			"description": "Cannabis retail staff who assist customers and process sales",
			"level":       2,
			"cannabis_permissions": []string{
				"Process sales", "Verify customer compliance", "View inventory",
				"Assist customers", "Use POS system", "Process transactions",
			},
		},
		string(v1.RoleDispensaryManager): map[string]interface{}{
			"name":        "Dispensary Manager",
			"description": "Cannabis dispensary managers with full operational control",
			"level":       3,
			"cannabis_permissions": []string{
				"Manage inventory", "Manage staff", "View reports",
				"Manage compliance", "Configure dispensary", "Access audit logs",
				"Manage products", "Approve transactions", "Manage marketing",
			},
		},
		string(v1.RoleBrandPartner): map[string]interface{}{
			"name":        "Brand Partner",
			"description": "Cannabis brand partners who manage their product lines",
			"level":       2,
			"cannabis_permissions": []string{
				"Manage brand products", "View brand analytics", "Manage brand inventory",
				"View brand sales", "Brand marketing", "Product compliance",
			},
		},
		string(v1.RoleSystemAdmin): map[string]interface{}{
			"name":        "System Administrator",
			"description": "Full system access for platform administration",
			"level":       4,
			"cannabis_permissions": []string{
				"Full system access", "Manage all dispensaries", "System configuration",
				"Platform analytics", "Compliance oversight", "Security management",
				"Role management", "Platform monitoring",
			},
		},
	}

	// Log roles access for audit
	h.Logger.LogCannabisAudit(authCtx.UserID.String(), "roles_list_accessed", "rbac", map[string]interface{}{
		"accessor_role": authCtx.Role,
		"roles_count":   len(roles),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"roles": roles,
		"total": len(roles),
		"cannabis_platform": true,
	})
}

// AddPolicy adds a new policy (admin only)
func (h *RBACHandler) AddPolicy(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req PolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate role
	role := v1.UserRole(req.Role)
	if !h.isValidRole(role) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role",
			"valid_roles": h.getValidRoles(),
		})
	}

	// Add policy
	if err := h.AuthzService.AddCannabisPolicy(role, req.Resource, req.Action); err != nil {
		h.Logger.Error("Failed to add policy",
			"error", err,
			"role", req.Role,
			"resource", req.Resource,
			"action", req.Action,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add policy",
		})
	}

	// Publish policy change event
	h.Messaging.PublishSystemEvent("policy_added", map[string]interface{}{
		"role":       req.Role,
		"resource":   req.Resource,
		"action":     req.Action,
		"admin_id":   authCtx.UserID.String(),
		"admin_role": authCtx.Role,
	})

	// Log policy addition for audit
	h.Logger.LogCannabisAudit(authCtx.UserID.String(), "policy_added", "rbac", map[string]interface{}{
		"role":       req.Role,
		"resource":   req.Resource,
		"action":     req.Action,
		"admin_role": authCtx.Role,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Policy added successfully",
		"policy": map[string]interface{}{
			"role":     req.Role,
			"resource": req.Resource,
			"action":   req.Action,
		},
		"added_by": authCtx.UserID.String(),
	})
}

// RemovePolicy removes a policy (admin only)
func (h *RBACHandler) RemovePolicy(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req PolicyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate role
	role := v1.UserRole(req.Role)
	if !h.isValidRole(role) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role",
			"valid_roles": h.getValidRoles(),
		})
	}

	// Remove policy
	if err := h.AuthzService.RemoveCannabisPolicy(role, req.Resource, req.Action); err != nil {
		h.Logger.Error("Failed to remove policy",
			"error", err,
			"role", req.Role,
			"resource", req.Resource,
			"action", req.Action,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove policy",
		})
	}

	// Publish policy change event
	h.Messaging.PublishSystemEvent("policy_removed", map[string]interface{}{
		"role":       req.Role,
		"resource":   req.Resource,
		"action":     req.Action,
		"admin_id":   authCtx.UserID.String(),
		"admin_role": authCtx.Role,
	})

	// Log policy removal for audit
	h.Logger.LogCannabisAudit(authCtx.UserID.String(), "policy_removed", "rbac", map[string]interface{}{
		"role":       req.Role,
		"resource":   req.Resource,
		"action":     req.Action,
		"admin_role": authCtx.Role,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Policy removed successfully",
		"policy": map[string]interface{}{
			"role":     req.Role,
			"resource": req.Resource,
			"action":   req.Action,
		},
		"removed_by": authCtx.UserID.String(),
	})
}

// GetCannabisResources returns all available cannabis resources
func (h *RBACHandler) GetCannabisResources(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	resources := map[string]interface{}{
		authz.ResourceUsers: map[string]interface{}{
			"name":        "Users",
			"description": "User account management",
			"category":    "User Management",
		},
		authz.ResourceProducts: map[string]interface{}{
			"name":        "Products",
			"description": "Cannabis product catalog",
			"category":    "Cannabis Business",
		},
		authz.ResourceInventory: map[string]interface{}{
			"name":        "Inventory",
			"description": "Cannabis inventory management",
			"category":    "Cannabis Business",
		},
		authz.ResourceOrders: map[string]interface{}{
			"name":        "Orders",
			"description": "Cannabis order processing",
			"category":    "Cannabis Business",
		},
		authz.ResourceTransactions: map[string]interface{}{
			"name":        "Transactions",
			"description": "Payment transactions",
			"category":    "Cannabis Business",
		},
		authz.ResourceDispensaries: map[string]interface{}{
			"name":        "Dispensaries",
			"description": "Dispensary management",
			"category":    "Cannabis Business",
		},
		authz.ResourceCompliance: map[string]interface{}{
			"name":        "Compliance",
			"description": "Cannabis compliance management",
			"category":    "Cannabis Compliance",
		},
		authz.ResourceAgeVerification: map[string]interface{}{
			"name":        "Age Verification",
			"description": "Customer age verification (21+)",
			"category":    "Cannabis Compliance",
		},
		authz.ResourceStateVerification: map[string]interface{}{
			"name":        "State Verification",
			"description": "Legal state verification",
			"category":    "Cannabis Compliance",
		},
		authz.ResourceAuditLogs: map[string]interface{}{
			"name":        "Audit Logs",
			"description": "Cannabis audit trails",
			"category":    "Cannabis Compliance",
		},
		authz.ResourcePOS: map[string]interface{}{
			"name":        "POS System",
			"description": "Point of sale operations",
			"category":    "Cannabis Operations",
		},
		authz.ResourceReports: map[string]interface{}{
			"name":        "Reports",
			"description": "Cannabis business reports",
			"category":    "Analytics",
		},
		authz.ResourceAnalytics: map[string]interface{}{
			"name":        "Analytics",
			"description": "Cannabis business analytics",
			"category":    "Analytics",
		},
	}

	// Log resources access for audit
	h.Logger.LogCannabisAudit(authCtx.UserID.String(), "resources_list_accessed", "rbac", map[string]interface{}{
		"accessor_role":   authCtx.Role,
		"resources_count": len(resources),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"resources": resources,
		"total":     len(resources),
		"cannabis_platform": true,
	})
}

// Helper methods

// isValidRole checks if role is valid
func (h *RBACHandler) isValidRole(role v1.UserRole) bool {
	validRoles := []v1.UserRole{
		v1.RoleCustomer,
		v1.RoleBudtender,
		v1.RoleDispensaryManager,
		v1.RoleBrandPartner,
		v1.RoleSystemAdmin,
	}

	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// getValidRoles returns list of valid roles
func (h *RBACHandler) getValidRoles() []string {
	return []string{
		string(v1.RoleCustomer),
		string(v1.RoleBudtender),
		string(v1.RoleDispensaryManager),
		string(v1.RoleBrandPartner),
		string(v1.RoleSystemAdmin),
	}
}