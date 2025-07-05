package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/oauth2"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// AuthConfig represents authentication middleware configuration
type AuthConfig struct {
	OAuth2Service *oauth2.Service
	Logger        *logger.Logger
	SkipPaths     []string
	RequiredRole  *v1.UserRole
}

// AuthMiddleware creates authentication middleware
func AuthMiddleware(config AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for certain paths
		path := c.Path()
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			config.Logger.Warn("Missing Authorization header",
				"path", path,
				"method", c.Method(),
				"ip", c.IP(),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
				"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
			})
		}

		token := oauth2.ExtractTokenFromHeader(authHeader)
		if token == "" {
			config.Logger.Warn("Invalid Authorization header format",
				"path", path,
				"method", c.Method(),
				"ip", c.IP(),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format. Expected: Bearer <token>",
				"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
			})
		}

		// Validate token
		claims, err := config.OAuth2Service.ValidateToken(token)
		if err != nil {
			config.Logger.Warn("Invalid token",
				"error", err,
				"path", path,
				"method", c.Method(),
				"ip", c.IP(),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
				"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
			})
		}

		// Check if token is access token
		if claims.TokenType != "access" {
			config.Logger.Warn("Wrong token type",
				"token_type", claims.TokenType,
				"path", path,
				"method", c.Method(),
				"ip", c.IP(),
				"user_id", claims.UserID,
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Access token required",
				"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
			})
		}

		// Get authentication context
		authCtx, err := config.OAuth2Service.GetAuthContext(c.Context(), claims.SessionID)
		if err != nil {
			config.Logger.Warn("Failed to get auth context",
				"error", err,
				"session_id", claims.SessionID,
				"path", path,
				"method", c.Method(),
				"ip", c.IP(),
				"user_id", claims.UserID,
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Session not found or expired",
				"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
			})
		}

		// Check role requirement if specified
		if config.RequiredRole != nil && authCtx.Role != *config.RequiredRole {
			// Allow system admin to access everything
			if authCtx.Role != v1.RoleSystemAdmin {
				config.Logger.Warn("Insufficient role permissions",
					"required_role", *config.RequiredRole,
					"user_role", authCtx.Role,
					"path", path,
					"method", c.Method(),
					"ip", c.IP(),
					"user_id", claims.UserID,
				)
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Insufficient permissions",
					"required_role": *config.RequiredRole,
					"user_role": authCtx.Role,
				})
			}
		}

		// Store authentication context in fiber locals
		c.Locals("auth", authCtx)
		c.Locals("user_id", authCtx.UserID.String())
		c.Locals("dispensary_id", authCtx.DispensaryID.String())
		c.Locals("role", authCtx.Role)
		c.Locals("session_id", authCtx.SessionID)

		// Log authenticated request for cannabis compliance
		config.Logger.LogHTTPRequest(
			c.Method(),
			path,
			authCtx.UserID.String(),
			authCtx.SessionID,
			200, // Will be updated in response middleware
			0,   // Will be updated in response middleware
			map[string]interface{}{
				"dispensary_id":     authCtx.DispensaryID.String(),
				"role":             authCtx.Role,
				"compliance_status": authCtx.ComplianceStatus,
				"age_verified":     authCtx.AgeVerified,
				"state_verified":   authCtx.StateVerified,
				"ip_address":       c.IP(),
				"user_agent":       c.Get("User-Agent"),
			},
		)

		return c.Next()
	}
}

// RequireRole creates middleware that requires specific role
func RequireRole(oauth2Service *oauth2.Service, logger *logger.Logger, role v1.UserRole) fiber.Handler {
	return AuthMiddleware(AuthConfig{
		OAuth2Service: oauth2Service,
		Logger:        logger,
		RequiredRole:  &role,
	})
}

// RequirePermission creates middleware that requires specific permission
func RequirePermission(oauth2Service *oauth2.Service, logger *logger.Logger, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First check authentication
		authMiddleware := AuthMiddleware(AuthConfig{
			OAuth2Service: oauth2Service,
			Logger:        logger,
		})
		
		if err := authMiddleware(c); err != nil {
			return err
		}

		// Get auth context from locals
		authCtx, ok := c.Locals("auth").(*oauth2.AuthContext)
		if !ok {
			logger.Error("Auth context not found in locals")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Check permission
		if !oauth2Service.CheckRolePermission(authCtx.Role, permission) {
			logger.Warn("Insufficient permissions",
				"required_permission", permission,
				"user_role", authCtx.Role,
				"path", c.Path(),
				"method", c.Method(),
				"ip", c.IP(),
				"user_id", authCtx.UserID.String(),
			)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Insufficient permissions",
				"required_permission": permission,
				"user_role": authCtx.Role,
			})
		}

		return c.Next()
	}
}

// RequireCannabisCompliance creates middleware that requires cannabis compliance
func RequireCannabisCompliance(oauth2Service *oauth2.Service, logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First check authentication
		authMiddleware := AuthMiddleware(AuthConfig{
			OAuth2Service: oauth2Service,
			Logger:        logger,
		})
		
		if err := authMiddleware(c); err != nil {
			return err
		}

		// Get auth context from locals
		authCtx, ok := c.Locals("auth").(*oauth2.AuthContext)
		if !ok {
			logger.Error("Auth context not found in locals")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Check cannabis compliance
		if err := oauth2Service.ValidateCannabisAccess(authCtx); err != nil {
			logger.LogCannabisAudit(authCtx.UserID.String(), "compliance_check_failed", c.Path(), map[string]interface{}{
				"error":            err.Error(),
				"age_verified":     authCtx.AgeVerified,
				"state_verified":   authCtx.StateVerified,
				"compliance_status": authCtx.ComplianceStatus,
				"state":            authCtx.State,
				"ip_address":       c.IP(),
				"path":             c.Path(),
				"method":           c.Method(),
			})

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": err.Error(),
				"compliance_requirements": map[string]interface{}{
					"age_verification_required":  true,
					"minimum_age":               21,
					"state_verification_required": true,
					"current_status": map[string]interface{}{
						"age_verified":     authCtx.AgeVerified,
						"state_verified":   authCtx.StateVerified,
						"compliance_status": authCtx.ComplianceStatus,
						"state":            authCtx.State,
					},
				},
				"cannabis_notice": "This platform is restricted to users 21+ in states where cannabis is legal",
			})
		}

		// Log successful compliance check
		logger.LogCannabisAudit(authCtx.UserID.String(), "compliance_check_passed", c.Path(), map[string]interface{}{
			"age_verified":     authCtx.AgeVerified,
			"state_verified":   authCtx.StateVerified,
			"compliance_status": authCtx.ComplianceStatus,
			"state":            authCtx.State,
			"dispensary_id":    authCtx.DispensaryID.String(),
		})

		return c.Next()
	}
}

// OptionalAuth creates middleware for optional authentication
func OptionalAuth(oauth2Service *oauth2.Service, logger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if Authorization header exists
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			return c.Next()
		}

		token := oauth2.ExtractTokenFromHeader(authHeader)
		if token == "" {
			// Invalid header format, continue without authentication
			return c.Next()
		}

		// Try to validate token
		claims, err := oauth2Service.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			logger.Debug("Optional auth failed",
				"error", err,
				"path", c.Path(),
				"ip", c.IP(),
			)
			return c.Next()
		}

		// Get authentication context
		authCtx, err := oauth2Service.GetAuthContext(c.Context(), claims.SessionID)
		if err != nil {
			// Session not found, continue without authentication
			logger.Debug("Optional auth context failed",
				"error", err,
				"session_id", claims.SessionID,
				"path", c.Path(),
				"ip", c.IP(),
			)
			return c.Next()
		}

		// Store authentication context in fiber locals
		c.Locals("auth", authCtx)
		c.Locals("user_id", authCtx.UserID.String())
		c.Locals("dispensary_id", authCtx.DispensaryID.String())
		c.Locals("role", authCtx.Role)
		c.Locals("session_id", authCtx.SessionID)
		c.Locals("authenticated", true)

		return c.Next()
	}
}

// GetAuthContext is a helper function to get auth context from fiber locals
func GetAuthContext(c *fiber.Ctx) (*oauth2.AuthContext, bool) {
	authCtx, ok := c.Locals("auth").(*oauth2.AuthContext)
	return authCtx, ok
}

// IsAuthenticated checks if request is authenticated
func IsAuthenticated(c *fiber.Ctx) bool {
	_, ok := GetAuthContext(c)
	return ok
}

// GetUserID is a helper function to get user ID from fiber locals
func GetUserID(c *fiber.Ctx) (string, bool) {
	userID, ok := c.Locals("user_id").(string)
	return userID, ok
}

// GetDispensaryID is a helper function to get dispensary ID from fiber locals
func GetDispensaryID(c *fiber.Ctx) (string, bool) {
	dispensaryID, ok := c.Locals("dispensary_id").(string)
	return dispensaryID, ok
}

// GetUserRole is a helper function to get user role from fiber locals
func GetUserRole(c *fiber.Ctx) (v1.UserRole, bool) {
	role, ok := c.Locals("role").(v1.UserRole)
	return role, ok
}

// GetSessionID is a helper function to get session ID from fiber locals
func GetSessionID(c *fiber.Ctx) (string, bool) {
	sessionID, ok := c.Locals("session_id").(string)
	return sessionID, ok
}