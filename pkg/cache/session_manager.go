package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// SessionManager provides advanced session management for cannabis platform
type SessionManager struct {
	Cache  *Cache
	Logger *logger.Logger
}

// SessionActivity represents user activity in a session
type SessionActivity struct {
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Timestamp   time.Time              `json:"timestamp"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Metadata    map[string]interface{} `json:"metadata"`
	Compliance  bool                   `json:"compliance"`
}

// SessionStats represents session statistics
type SessionStats struct {
	TotalSessions       int           `json:"total_sessions"`
	ActiveSessions      int           `json:"active_sessions"`
	ExpiredSessions     int           `json:"expired_sessions"`
	ComplianceRate      float64       `json:"compliance_rate"`
	AverageSessionTime  time.Duration `json:"average_session_time"`
	TopUserAgents       []string      `json:"top_user_agents"`
	TopDeviceTypes      []string      `json:"top_device_types"`
	GeographicDistribution map[string]int `json:"geographic_distribution"`
}

// CannabisSessionMetrics represents cannabis-specific session metrics
type CannabisSessionMetrics struct {
	TotalCompliantSessions    int     `json:"total_compliant_sessions"`
	AgeVerificationRate       float64 `json:"age_verification_rate"`
	StateVerificationRate     float64 `json:"state_verification_rate"`
	ComplianceViolations      int     `json:"compliance_violations"`
	StateDistribution        map[string]int `json:"state_distribution"`
	RoleDistribution         map[string]int `json:"role_distribution"`
	DispensaryDistribution   map[string]int `json:"dispensary_distribution"`
}

// NewSessionManager creates a new session manager
func NewSessionManager(cache *Cache, logger *logger.Logger) *SessionManager {
	return &SessionManager{
		Cache:  cache,
		Logger: logger,
	}
}

// Cannabis-specific session operations

// CreateCannabisSession creates a session with enhanced cannabis tracking
func (sm *SessionManager) CreateCannabisSession(ctx context.Context, sessionID string, data *SessionData, expiry time.Duration) error {
	// Validate cannabis compliance data
	if err := sm.validateCannabisCompliance(data); err != nil {
		sm.Logger.Warn("Cannabis compliance validation failed during session creation",
			"error", err,
			"user_id", data.UserID.String(),
			"dispensary_id", data.DispensaryID.String(),
		)
	}

	// Create base session
	if err := sm.Cache.CreateSession(ctx, sessionID, data, expiry); err != nil {
		return fmt.Errorf("failed to create cannabis session: %w", err)
	}

	// Track cannabis-specific metrics
	if err := sm.trackCannabisSessionMetrics(ctx, data); err != nil {
		sm.Logger.Warn("Failed to track cannabis session metrics", "error", err)
	}

	// Initialize session activity tracking
	activity := &SessionActivity{
		Action:     "session_created",
		Resource:   "authentication",
		Timestamp:  time.Now().UTC(),
		IPAddress:  data.IPAddress,
		UserAgent:  data.UserAgent,
		Metadata: map[string]interface{}{
			"compliance_verified": data.ComplianceVerified,
			"age_verified":       data.AgeVerified,
			"state_verified":     data.StateVerified,
			"state":             data.State,
			"role":              data.Role,
		},
		Compliance: true,
	}

	if err := sm.trackSessionActivity(ctx, sessionID, activity); err != nil {
		sm.Logger.Warn("Failed to track session activity", "error", err)
	}

	// Log cannabis session creation for audit
	sm.Logger.LogCannabisAudit(data.UserID.String(), "cannabis_session_created", "session_manager", map[string]interface{}{
		"session_id":         sessionID,
		"dispensary_id":      data.DispensaryID.String(),
		"compliance_verified": data.ComplianceVerified,
		"expiry":            expiry.String(),
		"state":             data.State,
		"device_type":       data.DeviceType,
	})

	return nil
}

// UpdateCannabisCompliance updates cannabis compliance for a session
func (sm *SessionManager) UpdateCannabisCompliance(ctx context.Context, sessionID string, ageVerified, stateVerified bool, state string) error {
	// Get current session
	sessionData, err := sm.Cache.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Update compliance fields
	oldComplianceStatus := sessionData.ComplianceVerified
	sessionData.AgeVerified = ageVerified
	sessionData.StateVerified = stateVerified
	sessionData.State = state
	sessionData.ComplianceVerified = ageVerified && stateVerified
	now := time.Now()
	sessionData.ComplianceCheckedAt = &now

	// Update compliance status
	if sessionData.ComplianceVerified {
		sessionData.ComplianceStatus = string(v1.ComplianceStatusVerified)
	} else {
		sessionData.ComplianceStatus = string(v1.ComplianceStatusPending)
	}

	// Save updated session
	if err := sm.Cache.UpdateSession(ctx, sessionID, sessionData); err != nil {
		return fmt.Errorf("failed to update session compliance: %w", err)
	}

	// Track compliance change activity
	activity := &SessionActivity{
		Action:     "compliance_updated",
		Resource:   "cannabis_verification",
		Timestamp:  time.Now().UTC(),
		IPAddress:  sessionData.IPAddress,
		UserAgent:  sessionData.UserAgent,
		Metadata: map[string]interface{}{
			"previous_compliance": oldComplianceStatus,
			"new_compliance":     sessionData.ComplianceVerified,
			"age_verified":       ageVerified,
			"state_verified":     stateVerified,
			"state":             state,
		},
		Compliance: true,
	}

	if err := sm.trackSessionActivity(ctx, sessionID, activity); err != nil {
		sm.Logger.Warn("Failed to track compliance activity", "error", err)
	}

	// Store compliance verification in separate cache for quick access
	complianceData := map[string]interface{}{
		"age_verified":       ageVerified,
		"state_verified":     stateVerified,
		"state":             state,
		"compliance_verified": sessionData.ComplianceVerified,
		"verified_at":       now,
		"session_id":        sessionID,
		"user_id":          sessionData.UserID.String(),
		"dispensary_id":     sessionData.DispensaryID.String(),
	}

	if err := sm.Cache.SetCannabisCompliance(ctx, sessionData.UserID.String(), complianceData, 24*time.Hour); err != nil {
		sm.Logger.Warn("Failed to cache compliance data", "error", err)
	}

	// Log compliance update for cannabis audit
	sm.Logger.LogCannabisAudit(sessionData.UserID.String(), "compliance_status_updated", "session_manager", map[string]interface{}{
		"session_id":         sessionID,
		"dispensary_id":      sessionData.DispensaryID.String(),
		"age_verified":       ageVerified,
		"state_verified":     stateVerified,
		"state":             state,
		"compliance_verified": sessionData.ComplianceVerified,
		"previous_compliance": oldComplianceStatus,
	})

	return nil
}

// TrackCannabisActivity tracks cannabis-specific user activity
func (sm *SessionManager) TrackCannabisActivity(ctx context.Context, sessionID, action, resource string, metadata map[string]interface{}) error {
	// Get session for user context
	sessionData, err := sm.Cache.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Create activity record
	activity := &SessionActivity{
		Action:     action,
		Resource:   resource,
		Timestamp:  time.Now().UTC(),
		IPAddress:  sessionData.IPAddress,
		UserAgent:  sessionData.UserAgent,
		Metadata:   metadata,
		Compliance: sm.isCannabisComplianceAction(action),
	}

	// Track activity
	if err := sm.trackSessionActivity(ctx, sessionID, activity); err != nil {
		return fmt.Errorf("failed to track cannabis activity: %w", err)
	}

	// Update session last activity
	sessionData.LastActivity = time.Now()
	if err := sm.Cache.UpdateSession(ctx, sessionID, sessionData); err != nil {
		sm.Logger.Warn("Failed to update session last activity", "error", err)
	}

	// Log cannabis-specific activities for audit
	if activity.Compliance {
		sm.Logger.LogCannabisAudit(sessionData.UserID.String(), action, resource, map[string]interface{}{
			"session_id":    sessionID,
			"dispensary_id": sessionData.DispensaryID.String(),
			"metadata":      metadata,
			"ip_address":    sessionData.IPAddress,
			"timestamp":     activity.Timestamp,
		})
	}

	return nil
}

// Session management and analytics

// GetSessionActivity retrieves activity history for a session
func (sm *SessionManager) GetSessionActivity(ctx context.Context, sessionID string, limit int) ([]*SessionActivity, error) {
	key := sm.sessionActivityKey(sessionID)
	
	// Get activity list from Redis
	activities, err := sm.Cache.Client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session activity: %w", err)
	}

	result := make([]*SessionActivity, 0, len(activities))
	for _, activityStr := range activities {
		var activity SessionActivity
		if err := json.Unmarshal([]byte(activityStr), &activity); err != nil {
			sm.Logger.Warn("Failed to unmarshal session activity", "error", err)
			continue
		}
		result = append(result, &activity)
	}

	return result, nil
}

// GetCannabisSessionMetrics returns cannabis-specific session metrics
func (sm *SessionManager) GetCannabisSessionMetrics(ctx context.Context, dispensaryID string, timeRange time.Duration) (*CannabisSessionMetrics, error) {
	// Get all sessions for the dispensary in the time range
	pattern := fmt.Sprintf("greenlync:session:*")
	keys, err := sm.Cache.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}

	metrics := &CannabisSessionMetrics{
		StateDistribution:      make(map[string]int),
		RoleDistribution:       make(map[string]int),
		DispensaryDistribution: make(map[string]int),
	}

	totalSessions := 0
	compliantSessions := 0
	ageVerifiedSessions := 0
	stateVerifiedSessions := 0
	complianceViolations := 0

	cutoffTime := time.Now().Add(-timeRange)

	for _, key := range keys {
		sessionDataStr, err := sm.Cache.Client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var sessionData SessionData
		if err := json.Unmarshal([]byte(sessionDataStr), &sessionData); err != nil {
			continue
		}

		// Filter by dispensary if specified
		if dispensaryID != "" && sessionData.DispensaryID.String() != dispensaryID {
			continue
		}

		// Filter by time range
		if sessionData.CreatedAt.Before(cutoffTime) {
			continue
		}

		totalSessions++

		// Track cannabis compliance metrics
		if sessionData.AgeVerified {
			ageVerifiedSessions++
		}
		if sessionData.StateVerified {
			stateVerifiedSessions++
		}
		if sessionData.ComplianceVerified {
			compliantSessions++
		} else {
			complianceViolations++
		}

		// Track distributions
		metrics.StateDistribution[sessionData.State]++
		metrics.RoleDistribution[sessionData.Role]++
		metrics.DispensaryDistribution[sessionData.DispensaryID.String()]++
	}

	// Calculate rates
	if totalSessions > 0 {
		metrics.AgeVerificationRate = float64(ageVerifiedSessions) / float64(totalSessions) * 100
		metrics.StateVerificationRate = float64(stateVerifiedSessions) / float64(totalSessions) * 100
	}

	metrics.TotalCompliantSessions = compliantSessions
	metrics.ComplianceViolations = complianceViolations

	return metrics, nil
}

// CleanupExpiredSessions removes expired sessions and their associated data
func (sm *SessionManager) CleanupExpiredSessions(ctx context.Context) error {
	sm.Logger.Info("Starting cleanup of expired sessions...")

	pattern := "greenlync:session:*"
	keys, err := sm.Cache.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	cleanedCount := 0
	for _, key := range keys {
		// Check if session exists and is expired
		ttl, err := sm.Cache.Client.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		if ttl <= 0 {
			// Session is expired, clean it up
			sessionID := key[len("greenlync:session:"):]
			
			// Get session data for logging before deletion
			sessionDataStr, err := sm.Cache.Client.Get(ctx, key).Result()
			if err == nil {
				var sessionData SessionData
				if err := json.Unmarshal([]byte(sessionDataStr), &sessionData); err == nil {
					// Log cleanup for audit
					sm.Logger.LogSessionActivity(sessionID, sessionData.UserID.String(), "session_expired_cleanup", map[string]interface{}{
						"dispensary_id": sessionData.DispensaryID.String(),
						"cleanup_time":  time.Now().UTC(),
						"expired_at":    sessionData.ExpiresAt,
					})
				}
			}

			// Delete session and associated data
			if err := sm.cleanupSessionData(ctx, sessionID); err != nil {
				sm.Logger.Warn("Failed to cleanup session data", "session_id", sessionID, "error", err)
			} else {
				cleanedCount++
			}
		}
	}

	sm.Logger.Info("Expired sessions cleanup completed",
		"cleaned_sessions", cleanedCount,
		"total_checked", len(keys),
	)

	return nil
}

// Helper methods

// validateCannabisCompliance validates cannabis compliance data
func (sm *SessionManager) validateCannabisCompliance(data *SessionData) error {
	// Check age verification requirement
	if !data.AgeVerified && data.ComplianceVerified {
		return fmt.Errorf("age verification required for compliance")
	}

	// Check state verification requirement
	if !data.StateVerified && data.ComplianceVerified {
		return fmt.Errorf("state verification required for compliance")
	}

	// Validate state is in legal cannabis states
	if data.StateVerified && !sm.isLegalCannabisState(data.State) {
		return fmt.Errorf("state %s is not a legal cannabis state", data.State)
	}

	return nil
}

// isLegalCannabisState checks if state allows cannabis
func (sm *SessionManager) isLegalCannabisState(state string) bool {
	legalStates := map[string]bool{
		"CA": true, "CO": true, "WA": true, "OR": true, "NV": true,
		"AZ": true, "NY": true, "IL": true, "NJ": true, "VA": true,
		"CT": true, "MT": true, "VT": true, "AK": true, "MA": true,
		"ME": true, "MI": true, "MD": true, "MO": true, "OH": true,
	}
	
	return legalStates[state]
}

// isCannabisComplianceAction checks if action requires cannabis compliance audit
func (sm *SessionManager) isCannabisComplianceAction(action string) bool {
	complianceActions := map[string]bool{
		"age_verification":      true,
		"state_verification":    true,
		"compliance_updated":    true,
		"product_viewed":        true,
		"product_purchased":     true,
		"order_created":         true,
		"payment_processed":     true,
		"compliance_violation":  true,
		"session_created":       true,
		"session_expired":       true,
	}
	
	return complianceActions[action]
}

// trackCannabisSessionMetrics tracks session metrics for cannabis compliance
func (sm *SessionManager) trackCannabisSessionMetrics(ctx context.Context, data *SessionData) error {
	date := time.Now().Format("2006-01-02")
	metricsKey := fmt.Sprintf("greenlync:metrics:sessions:%s", date)
	
	// Increment session counters
	pipe := sm.Cache.Client.Pipeline()
	pipe.HIncrBy(ctx, metricsKey, "total_sessions", 1)
	
	if data.ComplianceVerified {
		pipe.HIncrBy(ctx, metricsKey, "compliant_sessions", 1)
	}
	if data.AgeVerified {
		pipe.HIncrBy(ctx, metricsKey, "age_verified_sessions", 1)
	}
	if data.StateVerified {
		pipe.HIncrBy(ctx, metricsKey, "state_verified_sessions", 1)
	}
	
	// Track by state
	if data.State != "" {
		pipe.HIncrBy(ctx, metricsKey, fmt.Sprintf("state_%s", data.State), 1)
	}
	
	// Track by role
	pipe.HIncrBy(ctx, metricsKey, fmt.Sprintf("role_%s", data.Role), 1)
	
	// Track by dispensary
	pipe.HIncrBy(ctx, metricsKey, fmt.Sprintf("dispensary_%s", data.DispensaryID.String()), 1)
	
	// Set expiry for metrics (keep for 90 days)
	pipe.Expire(ctx, metricsKey, 90*24*time.Hour)
	
	_, err := pipe.Exec(ctx)
	return err
}

// trackSessionActivity stores session activity for audit and analytics
func (sm *SessionManager) trackSessionActivity(ctx context.Context, sessionID string, activity *SessionActivity) error {
	key := sm.sessionActivityKey(sessionID)
	
	activityJSON, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("failed to marshal activity: %w", err)
	}
	
	// Add to activity list (keep last 100 activities)
	pipe := sm.Cache.Client.Pipeline()
	pipe.LPush(ctx, key, activityJSON)
	pipe.LTrim(ctx, key, 0, 99) // Keep only last 100 activities
	pipe.Expire(ctx, key, 30*24*time.Hour) // Keep for 30 days
	
	_, err = pipe.Exec(ctx)
	return err
}

// cleanupSessionData removes all data associated with a session
func (sm *SessionManager) cleanupSessionData(ctx context.Context, sessionID string) error {
	sessionKey := sm.Cache.sessionKey(sessionID)
	activityKey := sm.sessionActivityKey(sessionID)
	
	pipe := sm.Cache.Client.Pipeline()
	pipe.Del(ctx, sessionKey)
	pipe.Del(ctx, activityKey)
	
	_, err := pipe.Exec(ctx)
	return err
}

// Key generation helpers

// sessionActivityKey generates a Redis key for session activity
func (sm *SessionManager) sessionActivityKey(sessionID string) string {
	return fmt.Sprintf("greenlync:session_activity:%s", sessionID)
}