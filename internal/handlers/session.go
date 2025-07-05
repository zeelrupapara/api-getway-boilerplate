package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/cache"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/messaging"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/internal/middleware"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// SessionHandler handles session management requests
type SessionHandler struct {
	SessionManager *cache.SessionManager
	Logger         *logger.Logger
	Messaging      *messaging.NATS
}

// ActivityTrackRequest represents activity tracking request
type ActivityTrackRequest struct {
	Action   string                 `json:"action" validate:"required"`
	Resource string                 `json:"resource" validate:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceUpdateRequest represents compliance update request
type ComplianceUpdateRequest struct {
	AgeVerified   bool   `json:"age_verified"`
	StateVerified bool   `json:"state_verified"`
	State         string `json:"state,omitempty"`
}

// MetricsRequest represents metrics request parameters
type MetricsRequest struct {
	DispensaryID string `json:"dispensary_id,omitempty"`
	TimeRange    string `json:"time_range,omitempty"` // "1h", "24h", "7d", "30d"
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionManager *cache.SessionManager, logger *logger.Logger, messaging *messaging.NATS) *SessionHandler {
	return &SessionHandler{
		SessionManager: sessionManager,
		Logger:         logger,
		Messaging:      messaging,
	}
}

// TrackActivity tracks user activity in current session
func (h *SessionHandler) TrackActivity(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req ActivityTrackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Add request context to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["ip_address"] = c.IP()
	req.Metadata["user_agent"] = c.Get("User-Agent")
	req.Metadata["path"] = c.Path()
	req.Metadata["method"] = c.Method()

	// Track activity
	if err := h.SessionManager.TrackCannabisActivity(c.Context(), authCtx.SessionID, req.Action, req.Resource, req.Metadata); err != nil {
		h.Logger.Error("Failed to track activity",
			"error", err,
			"session_id", authCtx.SessionID,
			"user_id", authCtx.UserID.String(),
			"action", req.Action,
			"resource", req.Resource,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to track activity",
		})
	}

	// Publish activity event for real-time processing
	h.Messaging.PublishUserEvent("activity_tracked", authCtx.UserID.String(), authCtx.DispensaryID.String(), map[string]interface{}{
		"session_id": authCtx.SessionID,
		"action":     req.Action,
		"resource":   req.Resource,
		"metadata":   req.Metadata,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Activity tracked successfully",
		"tracked_at": time.Now().UTC(),
	})
}

// UpdateCompliance updates cannabis compliance for current session
func (h *SessionHandler) UpdateCompliance(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req ComplianceUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate state if provided
	if req.State != "" && !h.isValidState(req.State) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state code",
			"valid_states": h.getLegalStates(),
		})
	}

	// Use current state if not provided
	if req.State == "" {
		req.State = authCtx.State
	}

	// Update compliance in session manager
	if err := h.SessionManager.UpdateCannabisCompliance(c.Context(), authCtx.SessionID, req.AgeVerified, req.StateVerified, req.State); err != nil {
		h.Logger.Error("Failed to update compliance",
			"error", err,
			"session_id", authCtx.SessionID,
			"user_id", authCtx.UserID.String(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update compliance status",
		})
	}

	// Publish compliance update event
	h.Messaging.PublishComplianceEvent("updated", authCtx.UserID.String(), authCtx.DispensaryID.String(), map[string]interface{}{
		"session_id":      authCtx.SessionID,
		"age_verified":    req.AgeVerified,
		"state_verified":  req.StateVerified,
		"state":          req.State,
		"updated_at":     time.Now().UTC(),
		"ip_address":     c.IP(),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Compliance updated successfully",
		"compliance": map[string]interface{}{
			"age_verified":   req.AgeVerified,
			"state_verified": req.StateVerified,
			"state":         req.State,
			"updated_at":    time.Now().UTC(),
		},
	})
}

// GetActivity retrieves activity history for current session
func (h *SessionHandler) GetActivity(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Parse limit parameter
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	// Get activity history
	activities, err := h.SessionManager.GetSessionActivity(c.Context(), authCtx.SessionID, limit)
	if err != nil {
		h.Logger.Error("Failed to get session activity",
			"error", err,
			"session_id", authCtx.SessionID,
			"user_id", authCtx.UserID.String(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve activity history",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"activities": activities,
		"total":      len(activities),
		"session_id": authCtx.SessionID,
	})
}

// GetMetrics retrieves cannabis session metrics (admin/manager only)
func (h *SessionHandler) GetMetrics(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Check if user has permission to view metrics
	if !h.canViewMetrics(authCtx.Role) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions to view metrics",
			"required_roles": []string{"dispensary_manager", "system_admin"},
		})
	}

	// Parse request parameters
	dispensaryID := c.Query("dispensary_id")
	timeRangeStr := c.Query("time_range", "24h")

	// For non-admin users, restrict to their dispensary
	if authCtx.Role != v1.RoleSystemAdmin {
		dispensaryID = authCtx.DispensaryID.String()
	}

	// Parse time range
	timeRange, err := h.parseTimeRange(timeRangeStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid time range format",
			"valid_ranges": []string{"1h", "24h", "7d", "30d"},
		})
	}

	// Get cannabis session metrics
	metrics, err := h.SessionManager.GetCannabisSessionMetrics(c.Context(), dispensaryID, timeRange)
	if err != nil {
		h.Logger.Error("Failed to get cannabis session metrics",
			"error", err,
			"user_id", authCtx.UserID.String(),
			"dispensary_id", dispensaryID,
			"time_range", timeRangeStr,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve session metrics",
		})
	}

	// Log metrics access for audit
	h.Logger.LogCannabisAudit(authCtx.UserID.String(), "metrics_accessed", "session_metrics", map[string]interface{}{
		"dispensary_id": dispensaryID,
		"time_range":   timeRangeStr,
		"user_role":    authCtx.Role,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"metrics": metrics,
		"filters": map[string]interface{}{
			"dispensary_id": dispensaryID,
			"time_range":   timeRangeStr,
			"generated_at": time.Now().UTC(),
		},
	})
}

// CleanupExpiredSessions triggers cleanup of expired sessions (admin only)
func (h *SessionHandler) CleanupExpiredSessions(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Check if user is admin
	if authCtx.Role != v1.RoleSystemAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only system administrators can trigger session cleanup",
		})
	}

	// Trigger cleanup
	if err := h.SessionManager.CleanupExpiredSessions(c.Context()); err != nil {
		h.Logger.Error("Failed to cleanup expired sessions",
			"error", err,
			"user_id", authCtx.UserID.String(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cleanup expired sessions",
		})
	}

	// Log cleanup action for audit
	h.Logger.LogCannabisAudit(authCtx.UserID.String(), "session_cleanup_triggered", "admin_action", map[string]interface{}{
		"triggered_at": time.Now().UTC(),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Session cleanup completed successfully",
		"triggered_at": time.Now().UTC(),
	})
}

// Helper methods

// isValidState checks if state code is valid
func (h *SessionHandler) isValidState(state string) bool {
	validStates := h.getLegalStates()
	for _, validState := range validStates {
		if validState == state {
			return true
		}
	}
	return false
}

// getLegalStates returns list of legal cannabis states
func (h *SessionHandler) getLegalStates() []string {
	return []string{
		"CA", "CO", "WA", "OR", "NV", "AZ", "NY", "IL", "NJ", "VA",
		"CT", "MT", "VT", "AK", "MA", "ME", "MI", "MD", "MO", "OH",
	}
}

// canViewMetrics checks if user role can view metrics
func (h *SessionHandler) canViewMetrics(role v1.UserRole) bool {
	allowedRoles := []v1.UserRole{
		v1.RoleDispensaryManager,
		v1.RoleSystemAdmin,
	}

	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return true
		}
	}
	return false
}

// parseTimeRange parses time range string to duration
func (h *SessionHandler) parseTimeRange(timeRangeStr string) (time.Duration, error) {
	switch timeRangeStr {
	case "1h":
		return time.Hour, nil
	case "24h":
		return 24 * time.Hour, nil
	case "7d":
		return 7 * 24 * time.Hour, nil
	case "30d":
		return 30 * 24 * time.Hour, nil
	default:
		// Try to parse as duration
		return time.ParseDuration(timeRangeStr)
	}
}