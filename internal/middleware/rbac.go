package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/oauth2"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/authz"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// RBACConfig represents RBAC middleware configuration
type RBACConfig struct {
	OAuth2Service *oauth2.Service
	AuthzService  *authz.Service
	Logger        *logger.Logger
	Resource      string
	Action        string
	SkipPaths     []string
}

// RequireCasbinPermission creates middleware that enforces Casbin RBAC permissions
func RequireCasbinPermission(oauth2Service *oauth2.Service, authzService *authz.Service, logger *logger.Logger, resource, action string) fiber.Handler {
	return RBACMiddleware(RBACConfig{
		OAuth2Service: oauth2Service,
		AuthzService:  authzService,
		Logger:        logger,
		Resource:      resource,
		Action:        action,
	})
}

// RBACMiddleware creates comprehensive RBAC middleware with Casbin integration
func RBACMiddleware(config RBACConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip RBAC for certain paths
		path := c.Path()
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// First ensure user is authenticated
		authMiddleware := AuthMiddleware(AuthConfig{
			OAuth2Service: config.OAuth2Service,
			Logger:        config.Logger,
		})
		
		if err := authMiddleware(c); err != nil {
			return err
		}

		// Get auth context from locals
		authCtx, ok := GetAuthContext(c)
		if !ok {
			config.Logger.Error("Auth context not found in RBAC middleware")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Enforce Casbin RBAC permission
		allowed, err := config.AuthzService.Enforce(authCtx.Role, config.Resource, config.Action)
		if err != nil {
			config.Logger.Error("RBAC enforcement error",
				"error", err,
				"user_id", authCtx.UserID.String(),
				"role", authCtx.Role,
				"resource", config.Resource,
				"action", config.Action,
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authorization check failed",
			})
		}

		if !allowed {
			config.Logger.Warn("RBAC permission denied",
				"user_id", authCtx.UserID.String(),
				"role", authCtx.Role,
				"resource", config.Resource,
				"action", config.Action,
				"path", path,
				"method", c.Method(),
				"ip", c.IP(),
			)

			// Log authorization denial for cannabis compliance
			config.Logger.LogCannabisAudit(authCtx.UserID.String(), "authorization_denied", "rbac", map[string]interface{}{
				"role":         authCtx.Role,
				"resource":     config.Resource,
				"action":       config.Action,
				"path":         path,
				"method":       c.Method(),
				"ip_address":   c.IP(),
				"dispensary_id": authCtx.DispensaryID.String(),
			})

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
				"required_permission": map[string]interface{}{
					"resource": config.Resource,
					"action":   config.Action,
				},
				"user_role": authCtx.Role,
				"cannabis_notice": "This platform requires appropriate role permissions for cannabis operations",
			})
		}

		// Log successful authorization for cannabis compliance
		config.Logger.LogCannabisAudit(authCtx.UserID.String(), "authorization_granted", "rbac", map[string]interface{}{
			"role":         authCtx.Role,
			"resource":     config.Resource,
			"action":       config.Action,
			"path":         path,
			"method":       c.Method(),
			"dispensary_id": authCtx.DispensaryID.String(),
		})

		return c.Next()
	}
}

// Cannabis-specific RBAC middleware

// RequireCannabisPermission creates middleware for cannabis-specific resources
func RequireCannabisPermission(oauth2Service *oauth2.Service, authzService *authz.Service, logger *logger.Logger, resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First check authentication
		authMiddleware := AuthMiddleware(AuthConfig{
			OAuth2Service: oauth2Service,
			Logger:        logger,
		})
		
		if err := authMiddleware(c); err != nil {
			return err
		}

		// Get auth context
		authCtx, ok := GetAuthContext(c)
		if !ok {
			logger.Error("Auth context not found in cannabis RBAC middleware")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Check cannabis-specific permissions
		allowed, err := authzService.CanAccessCannabisResource(authCtx.Role, resource, action)
		if err != nil {
			logger.Error("Cannabis RBAC enforcement error",
				"error", err,
				"user_id", authCtx.UserID.String(),
				"role", authCtx.Role,
				"resource", resource,
				"action", action,
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Cannabis authorization check failed",
			})
		}

		if !allowed {
			logger.Warn("Cannabis RBAC permission denied",
				"user_id", authCtx.UserID.String(),
				"role", authCtx.Role,
				"resource", resource,
				"action", action,
				"path", c.Path(),
				"method", c.Method(),
				"ip", c.IP(),
			)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions for cannabis operations",
				"required_permission": map[string]interface{}{
					"resource": resource,
					"action":   action,
				},
				"user_role": authCtx.Role,
				"cannabis_notice": "This cannabis platform operation requires elevated permissions",
			})
		}

		return c.Next()
	}
}

// RequireDispensaryAccess creates middleware that enforces dispensary-level access control
func RequireDispensaryAccess(oauth2Service *oauth2.Service, authzService *authz.Service, logger *logger.Logger, resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First check authentication
		authMiddleware := AuthMiddleware(AuthConfig{
			OAuth2Service: oauth2Service,
			Logger:        logger,
		})
		
		if err := authMiddleware(c); err != nil {
			return err
		}

		// Get auth context
		authCtx, ok := GetAuthContext(c)
		if !ok {
			logger.Error("Auth context not found in dispensary access middleware")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Get dispensary ID from request (could be from path params, query, or body)
		resourceDispensaryID := extractDispensaryIDFromRequest(c)
		
		// Check dispensary-specific access
		allowed, err := authzService.CanAccessDispensaryResource(
			authCtx.Role,
			authCtx.DispensaryID.String(),
			resourceDispensaryID,
			resource,
			action,
		)
		if err != nil {
			logger.Error("Dispensary RBAC enforcement error",
				"error", err,
				"user_id", authCtx.UserID.String(),
				"user_dispensary_id", authCtx.DispensaryID.String(),
				"resource_dispensary_id", resourceDispensaryID,
				"resource", resource,
				"action", action,
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Dispensary authorization check failed",
			})
		}

		if !allowed {
			logger.Warn("Dispensary access denied",
				"user_id", authCtx.UserID.String(),
				"user_dispensary_id", authCtx.DispensaryID.String(),
				"resource_dispensary_id", resourceDispensaryID,
				"resource", resource,
				"action", action,
				"path", c.Path(),
				"method", c.Method(),
				"ip", c.IP(),
			)

			// Log dispensary access violation for cannabis compliance
			logger.LogCannabisAudit(authCtx.UserID.String(), "dispensary_access_denied", "rbac", map[string]interface{}{
				"user_dispensary_id":     authCtx.DispensaryID.String(),
				"resource_dispensary_id": resourceDispensaryID,
				"resource":              resource,
				"action":                action,
				"path":                  c.Path(),
				"method":                c.Method(),
				"ip_address":            c.IP(),
			})

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied to dispensary resource",
				"user_dispensary_id": authCtx.DispensaryID.String(),
				"resource_dispensary_id": resourceDispensaryID,
				"cannabis_notice": "Users can only access resources from their assigned dispensary",
			})
		}

		// Store dispensary context for use in handlers
		c.Locals("resource_dispensary_id", resourceDispensaryID)

		return c.Next()
	}
}

// Role-specific middleware shortcuts

// RequireCustomerRole creates middleware that requires customer role
func RequireCustomerRole(oauth2Service *oauth2.Service, logger *logger.Logger) fiber.Handler {
	return RequireRole(oauth2Service, logger, v1.RoleCustomer)
}

// RequireBudtenderRole creates middleware that requires budtender role or higher
func RequireBudtenderRole(oauth2Service *oauth2.Service, authzService *authz.Service, logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, ok := c.Locals("auth").(*oauth2.AuthContext)
		if !ok {
			// Run auth middleware first
			authMiddleware := AuthMiddleware(AuthConfig{
				OAuth2Service: oauth2Service,
				Logger:        logger,
			})
			if err := authMiddleware(c); err != nil {
				return err
			}
			authCtx, _ = GetAuthContext(c)
		}

		allowedRoles := []v1.UserRole{v1.RoleBudtender, v1.RoleDispensaryManager, v1.RoleSystemAdmin}
		for _, role := range allowedRoles {
			if authCtx.Role == role {
				return c.Next()
			}
		}

		logger.Warn("Budtender role required",
			"user_id", authCtx.UserID.String(),
			"user_role", authCtx.Role,
			"required_roles", allowedRoles,
		)

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Budtender role or higher required",
			"user_role": authCtx.Role,
			"required_roles": allowedRoles,
		})
	}
}

// RequireManagerRole creates middleware that requires dispensary manager role or higher
func RequireManagerRole(oauth2Service *oauth2.Service, logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, ok := c.Locals("auth").(*oauth2.AuthContext)
		if !ok {
			// Run auth middleware first
			authMiddleware := AuthMiddleware(AuthConfig{
				OAuth2Service: oauth2Service,
				Logger:        logger,
			})
			if err := authMiddleware(c); err != nil {
				return err
			}
			authCtx, _ = GetAuthContext(c)
		}

		allowedRoles := []v1.UserRole{v1.RoleDispensaryManager, v1.RoleSystemAdmin}
		for _, role := range allowedRoles {
			if authCtx.Role == role {
				return c.Next()
			}
		}

		logger.Warn("Manager role required",
			"user_id", authCtx.UserID.String(),
			"user_role", authCtx.Role,
			"required_roles", allowedRoles,
		)

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Dispensary manager role or higher required",
			"user_role": authCtx.Role,
			"required_roles": allowedRoles,
		})
	}
}

// RequireAdminRole creates middleware that requires system admin role
func RequireAdminRole(oauth2Service *oauth2.Service, logger *logger.Logger) fiber.Handler {
	return RequireRole(oauth2Service, logger, v1.RoleSystemAdmin)
}

// Cannabis operation-specific middleware

// RequireCannabisOperationPermission creates middleware for specific cannabis operations
func RequireCannabisOperationPermission(oauth2Service *oauth2.Service, authzService *authz.Service, logger *logger.Logger, operation string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First check authentication
		authMiddleware := AuthMiddleware(AuthConfig{
			OAuth2Service: oauth2Service,
			Logger:        logger,
		})
		
		if err := authMiddleware(c); err != nil {
			return err
		}

		// Get auth context
		authCtx, ok := GetAuthContext(c)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Validate role for cannabis operation
		if !authzService.ValidateRoleForCannabisOperation(authCtx.Role, operation) {
			logger.Warn("Cannabis operation permission denied",
				"user_id", authCtx.UserID.String(),
				"role", authCtx.Role,
				"operation", operation,
				"path", c.Path(),
				"method", c.Method(),
			)

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions for cannabis operation",
				"operation": operation,
				"user_role": authCtx.Role,
				"cannabis_notice": "This cannabis operation requires specific role permissions",
			})
		}

		return c.Next()
	}
}

// Helper functions

// extractDispensaryIDFromRequest extracts dispensary ID from various request sources
func extractDispensaryIDFromRequest(c *fiber.Ctx) string {
	// Try path parameter first
	if dispensaryID := c.Params("dispensary_id"); dispensaryID != "" {
		return dispensaryID
	}

	// Try query parameter
	if dispensaryID := c.Query("dispensary_id"); dispensaryID != "" {
		return dispensaryID
	}

	// Try to parse from request body
	var body map[string]interface{}
	if err := c.BodyParser(&body); err == nil {
		if dispensaryID, ok := body["dispensary_id"].(string); ok && dispensaryID != "" {
			return dispensaryID
		}
	}

	// Default to user's dispensary ID from auth context
	if authCtx, ok := GetAuthContext(c); ok {
		return authCtx.DispensaryID.String()
	}

	return ""
}

// GetResourceDispensaryID is a helper function to get resource dispensary ID from locals
func GetResourceDispensaryID(c *fiber.Ctx) (string, bool) {
	dispensaryID, ok := c.Locals("resource_dispensary_id").(string)
	return dispensaryID, ok
}