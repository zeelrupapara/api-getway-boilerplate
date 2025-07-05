package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/oauth2"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/cache"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// ComplianceConfig represents cannabis compliance middleware configuration
type ComplianceConfig struct {
	OAuth2Service    *oauth2.Service
	SessionManager   *cache.SessionManager
	Logger           *logger.Logger
	RequireAge       bool
	RequireState     bool
	MinimumAge       int
	LegalStates      []string
	SkipPaths        []string
	StrictMode       bool
}

// CannabisComplianceMiddleware creates comprehensive cannabis compliance middleware
func CannabisComplianceMiddleware(config ComplianceConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip compliance for certain paths
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

		// Get auth context
		authCtx, ok := GetAuthContext(c)
		if !ok {
			config.Logger.Error("Auth context not found in compliance middleware")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Check cannabis compliance requirements
		if err := validateCannabisCompliance(authCtx, config, c); err != nil {
			return err
		}

		// Track compliance check activity
		if config.SessionManager != nil {
			config.SessionManager.TrackCannabisActivity(c.Context(), authCtx.SessionID, "compliance_check", "middleware", map[string]interface{}{
				"path":         path,
				"method":       c.Method(),
				"age_verified": authCtx.AgeVerified,
				"state_verified": authCtx.StateVerified,
				"compliance_status": authCtx.ComplianceStatus,
				"state":        authCtx.State,
			})
		}

		// Log successful compliance check
		config.Logger.LogCannabisAudit(authCtx.UserID.String(), "compliance_middleware_passed", "compliance", map[string]interface{}{
			"path":             path,
			"method":           c.Method(),
			"age_verified":     authCtx.AgeVerified,
			"state_verified":   authCtx.StateVerified,
			"compliance_status": authCtx.ComplianceStatus,
			"state":            authCtx.State,
			"dispensary_id":    authCtx.DispensaryID.String(),
			"ip_address":       c.IP(),
		})

		return c.Next()
	}
}

// AgeVerificationMiddleware specifically checks age verification
func AgeVerificationMiddleware(oauth2Service *oauth2.Service, logger *logger.Logger, minimumAge int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, ok := GetAuthContext(c)
		if !ok {
			// Run auth middleware first if not authenticated
			authMiddleware := AuthMiddleware(AuthConfig{
				OAuth2Service: oauth2Service,
				Logger:        logger,
			})
			if err := authMiddleware(c); err != nil {
				return err
			}
			authCtx, _ = GetAuthContext(c)
		}

		if !authCtx.AgeVerified {
			logger.LogCannabisAudit(authCtx.UserID.String(), "age_verification_required", "compliance", map[string]interface{}{
				"path":         c.Path(),
				"method":       c.Method(),
				"age_verified": false,
				"minimum_age":  minimumAge,
				"ip_address":   c.IP(),
			})

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Age verification required",
				"compliance_requirement": map[string]interface{}{
					"type":        "age_verification",
					"minimum_age": minimumAge,
					"verified":    false,
					"message":     "You must verify you are 21 years or older to access cannabis products",
				},
				"verification_endpoints": map[string]string{
					"verify": "/api/v1/auth/verify-compliance",
					"profile": "/api/v1/auth/profile",
				},
			})
		}

		return c.Next()
	}
}

// StateVerificationMiddleware specifically checks state verification
func StateVerificationMiddleware(oauth2Service *oauth2.Service, logger *logger.Logger, legalStates []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, ok := GetAuthContext(c)
		if !ok {
			// Run auth middleware first if not authenticated
			authMiddleware := AuthMiddleware(AuthConfig{
				OAuth2Service: oauth2Service,
				Logger:        logger,
			})
			if err := authMiddleware(c); err != nil {
				return err
			}
			authCtx, _ = GetAuthContext(c)
		}

		if !authCtx.StateVerified || !isValidCannabisState(authCtx.State, legalStates) {
			logger.LogCannabisAudit(authCtx.UserID.String(), "state_verification_required", "compliance", map[string]interface{}{
				"path":           c.Path(),
				"method":         c.Method(),
				"state_verified": authCtx.StateVerified,
				"current_state":  authCtx.State,
				"legal_states":   legalStates,
				"ip_address":     c.IP(),
			})

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "State verification required",
				"compliance_requirement": map[string]interface{}{
					"type":         "state_verification",
					"legal_states": legalStates,
					"current_state": authCtx.State,
					"verified":     authCtx.StateVerified,
					"message":      "You must be located in a state where cannabis is legal",
				},
				"verification_endpoints": map[string]string{
					"verify": "/api/v1/auth/verify-compliance",
					"profile": "/api/v1/auth/profile",
				},
			})
		}

		return c.Next()
	}
}

// PurchaseLimitMiddleware enforces cannabis purchase limits
func PurchaseLimitMiddleware(oauth2Service *oauth2.Service, logger *logger.Logger, sessionManager *cache.SessionManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, ok := GetAuthContext(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Only apply to purchase-related endpoints
		if !isPurchaseEndpoint(c.Path(), c.Method()) {
			return c.Next()
		}

		// Check daily purchase limits (example: $1000 per day)
		dailyLimit := 1000.0
		if exceeded, currentAmount, err := checkDailyPurchaseLimit(authCtx.UserID.String(), dailyLimit); err != nil {
			logger.Error("Failed to check purchase limit",
				"error", err,
				"user_id", authCtx.UserID.String(),
			)
		} else if exceeded {
			logger.LogCannabisAudit(authCtx.UserID.String(), "purchase_limit_exceeded", "compliance", map[string]interface{}{
				"daily_limit":    dailyLimit,
				"current_amount": currentAmount,
				"path":          c.Path(),
				"method":        c.Method(),
				"ip_address":    c.IP(),
			})

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Daily purchase limit exceeded",
				"compliance_limit": map[string]interface{}{
					"type":           "daily_purchase_limit",
					"limit":          dailyLimit,
					"current_amount": currentAmount,
					"message":        "You have reached your daily cannabis purchase limit",
				},
			})
		}

		// Track purchase attempt
		if sessionManager != nil {
			sessionManager.TrackCannabisActivity(c.Context(), authCtx.SessionID, "purchase_attempt", "compliance", map[string]interface{}{
				"daily_limit":  dailyLimit,
				"path":        c.Path(),
				"method":      c.Method(),
			})
		}

		return c.Next()
	}
}

// TimeBasedAccessMiddleware enforces time-based access controls
func TimeBasedAccessMiddleware(oauth2Service *oauth2.Service, logger *logger.Logger, startHour, endHour int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, ok := GetAuthContext(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		currentHour := time.Now().Hour()
		if currentHour < startHour || currentHour >= endHour {
			logger.LogCannabisAudit(authCtx.UserID.String(), "time_based_access_denied", "compliance", map[string]interface{}{
				"current_hour": currentHour,
				"allowed_start": startHour,
				"allowed_end":   endHour,
				"path":         c.Path(),
				"method":       c.Method(),
				"ip_address":   c.IP(),
			})

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Cannabis access not allowed at this time",
				"compliance_requirement": map[string]interface{}{
					"type":          "time_based_access",
					"current_time":  time.Now().Format("15:04"),
					"allowed_hours": map[string]interface{}{
						"start": startHour,
						"end":   endHour,
					},
					"message": "Cannabis purchases are only allowed during business hours",
				},
			})
		}

		return c.Next()
	}
}

// SessionTimeoutMiddleware enforces session timeout for cannabis operations
func SessionTimeoutMiddleware(oauth2Service *oauth2.Service, logger *logger.Logger, sessionTimeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, ok := GetAuthContext(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Check if session has been inactive too long
		if time.Since(authCtx.LastActivity) > sessionTimeout {
			logger.LogCannabisAudit(authCtx.UserID.String(), "session_timeout", "compliance", map[string]interface{}{
				"last_activity":   authCtx.LastActivity,
				"timeout_duration": sessionTimeout.String(),
				"session_id":      authCtx.SessionID,
				"path":           c.Path(),
				"method":         c.Method(),
			})

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Session timeout",
				"compliance_requirement": map[string]interface{}{
					"type":            "session_timeout",
					"timeout_duration": sessionTimeout.String(),
					"last_activity":   authCtx.LastActivity,
					"message":         "Your session has expired due to inactivity",
				},
				"action_required": "Please login again to continue",
			})
		}

		return c.Next()
	}
}

// GeoLocationMiddleware enforces geographic restrictions
func GeoLocationMiddleware(oauth2Service *oauth2.Service, logger *logger.Logger, allowedStates []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authCtx, ok := GetAuthContext(c)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Get client IP and attempt to determine location
		clientIP := c.IP()
		detectedState := detectStateFromIP(clientIP) // Would integrate with IP geolocation service

		if detectedState != "" && !contains(allowedStates, detectedState) {
			logger.LogCannabisAudit(authCtx.UserID.String(), "geographic_restriction_violation", "compliance", map[string]interface{}{
				"client_ip":       clientIP,
				"detected_state":  detectedState,
				"allowed_states":  allowedStates,
				"user_state":      authCtx.State,
				"path":           c.Path(),
				"method":         c.Method(),
			})

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Geographic restriction",
				"compliance_requirement": map[string]interface{}{
					"type":            "geographic_restriction",
					"detected_location": detectedState,
					"allowed_states":   allowedStates,
					"message":          "Cannabis access is restricted to legal states only",
				},
			})
		}

		return c.Next()
	}
}

// Helper functions

// validateCannabisCompliance performs comprehensive cannabis compliance validation
func validateCannabisCompliance(authCtx *oauth2.AuthContext, config ComplianceConfig, c *fiber.Ctx) error {
	violations := make([]map[string]interface{}, 0)

	// Check age verification
	if config.RequireAge && !authCtx.AgeVerified {
		violations = append(violations, map[string]interface{}{
			"type":        "age_verification",
			"required":    true,
			"verified":    false,
			"minimum_age": config.MinimumAge,
			"message":     "Age verification required (21+)",
		})
	}

	// Check state verification
	if config.RequireState && (!authCtx.StateVerified || !isValidCannabisState(authCtx.State, config.LegalStates)) {
		violations = append(violations, map[string]interface{}{
			"type":         "state_verification",
			"required":     true,
			"verified":     authCtx.StateVerified,
			"current_state": authCtx.State,
			"legal_states": config.LegalStates,
			"message":      "Valid state verification required",
		})
	}

	// Check overall compliance status
	if authCtx.ComplianceStatus != v1.ComplianceStatusVerified && config.StrictMode {
		violations = append(violations, map[string]interface{}{
			"type":             "compliance_status",
			"required_status":  v1.ComplianceStatusVerified,
			"current_status":   authCtx.ComplianceStatus,
			"message":          "Full compliance verification required",
		})
	}

	if len(violations) > 0 {
		config.Logger.LogCannabisAudit(authCtx.UserID.String(), "compliance_violations_detected", "compliance", map[string]interface{}{
			"violations":    violations,
			"path":         c.Path(),
			"method":       c.Method(),
			"ip_address":   c.IP(),
			"dispensary_id": authCtx.DispensaryID.String(),
		})

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Cannabis compliance requirements not met",
			"compliance_violations": violations,
			"user_status": map[string]interface{}{
				"age_verified":      authCtx.AgeVerified,
				"state_verified":    authCtx.StateVerified,
				"compliance_status": authCtx.ComplianceStatus,
				"state":            authCtx.State,
			},
			"verification_endpoints": map[string]string{
				"verify_compliance": "/api/v1/auth/verify-compliance",
				"update_profile":   "/api/v1/auth/profile",
			},
			"cannabis_notice": "This platform requires full compliance verification for cannabis access",
		})
	}

	return nil
}

// isValidCannabisState checks if state is in the legal cannabis states list
func isValidCannabisState(state string, legalStates []string) bool {
	for _, legalState := range legalStates {
		if state == legalState {
			return true
		}
	}
	return false
}

// isPurchaseEndpoint checks if the endpoint is related to purchases
func isPurchaseEndpoint(path, method string) bool {
	purchasePatterns := []string{
		"/api/v1/orders",
		"/api/v1/cart",
		"/api/v1/checkout",
		"/api/v1/payments",
		"/api/v1/transactions",
	}

	if method != "POST" && method != "PUT" {
		return false
	}

	for _, pattern := range purchasePatterns {
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}
	return false
}

// checkDailyPurchaseLimit checks if user has exceeded daily purchase limit
func checkDailyPurchaseLimit(userID string, limit float64) (bool, float64, error) {
	// This would integrate with order/transaction database
	// For now, return mock data
	currentAmount := 750.0 // Example current daily spending
	return currentAmount >= limit, currentAmount, nil
}

// detectStateFromIP attempts to detect state from IP address
func detectStateFromIP(ip string) string {
	// This would integrate with IP geolocation service
	// For now, return empty string
	return ""
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Convenience middleware creators

// RequireFullCannabisCompliance creates middleware requiring full cannabis compliance
func RequireFullCannabisCompliance(oauth2Service *oauth2.Service, sessionManager *cache.SessionManager, logger *logger.Logger) fiber.Handler {
	legalStates := []string{
		"CA", "CO", "WA", "OR", "NV", "AZ", "NY", "IL", "NJ", "VA",
		"CT", "MT", "VT", "AK", "MA", "ME", "MI", "MD", "MO", "OH",
	}

	return CannabisComplianceMiddleware(ComplianceConfig{
		OAuth2Service:  oauth2Service,
		SessionManager: sessionManager,
		Logger:         logger,
		RequireAge:     true,
		RequireState:   true,
		MinimumAge:     21,
		LegalStates:    legalStates,
		StrictMode:     true,
	})
}

// RequireMinimalCannabisCompliance creates middleware for basic compliance
func RequireMinimalCannabisCompliance(oauth2Service *oauth2.Service, logger *logger.Logger) fiber.Handler {
	return CannabisComplianceMiddleware(ComplianceConfig{
		OAuth2Service: oauth2Service,
		Logger:        logger,
		RequireAge:    true,
		RequireState:  false,
		MinimumAge:    21,
		StrictMode:    false,
	})
}

// RequireAgeOnly creates middleware that only requires age verification
func RequireAgeOnly(oauth2Service *oauth2.Service, logger *logger.Logger) fiber.Handler {
	return AgeVerificationMiddleware(oauth2Service, logger, 21)
}

// RequireStateOnly creates middleware that only requires state verification
func RequireStateOnly(oauth2Service *oauth2.Service, logger *logger.Logger) fiber.Handler {
	legalStates := []string{
		"CA", "CO", "WA", "OR", "NV", "AZ", "NY", "IL", "NJ", "VA",
		"CT", "MT", "VT", "AK", "MA", "ME", "MI", "MD", "MO", "OH",
	}
	return StateVerificationMiddleware(oauth2Service, logger, legalStates)
}