package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/config"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
)

// Cache represents the Redis cache client
type Cache struct {
	Client *redis.Client
	Config *config.RedisConfig
	Logger *logger.Logger
}

// SessionData represents session data stored in Redis
type SessionData struct {
	UserID           uuid.UUID `json:"user_id"`
	DispensaryID     uuid.UUID `json:"dispensary_id"`
	Role             string    `json:"role"`
	Email            string    `json:"email"`
	
	// Cannabis compliance context
	AgeVerified      bool      `json:"age_verified"`
	StateVerified    bool      `json:"state_verified"`
	ComplianceStatus string    `json:"compliance_status"`
	State            string    `json:"state"`
	
	// Session metadata
	IPAddress        string    `json:"ip_address"`
	UserAgent        string    `json:"user_agent"`
	DeviceType       string    `json:"device_type"`
	Location         string    `json:"location"`
	
	// Timestamps
	CreatedAt        time.Time `json:"created_at"`
	LastActivity     time.Time `json:"last_activity"`
	ExpiresAt        time.Time `json:"expires_at"`
	
	// Session limits
	MaxSessions      int       `json:"max_sessions"`
	CurrentSessions  int       `json:"current_sessions"`
	
	// Cannabis audit context
	ComplianceVerified bool     `json:"compliance_verified"`
	ComplianceCheckedAt *time.Time `json:"compliance_checked_at"`
}

// NewCache creates a new Redis cache client
func NewCache(cfg *config.RedisConfig, log *logger.Logger) (*Cache, error) {
	log.Info("Initializing Redis cache connection",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database,
		"pool_size", cfg.PoolSize,
	)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Database,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConn,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.ReadTimeout, // Use ReadTimeout for WriteTimeout
		IdleTimeout:  300 * time.Second,
		
		// Connection pool settings
		PoolTimeout:  30 * time.Second,
		IdleCheckFrequency: 60 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &Cache{
		Client: client,
		Config: cfg,
		Logger: log,
	}

	log.Info("Redis cache connection established successfully",
		"pool_size", cfg.PoolSize,
		"min_idle_conns", cfg.MinIdleConn,
		"max_retries", cfg.MaxRetries,
	)

	return cache, nil
}

// Close closes the Redis connection
func (c *Cache) Close() error {
	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}

	c.Logger.Info("Redis cache connection closed successfully")
	return nil
}

// Health checks the Redis connection health
func (c *Cache) Health(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// GetStats returns Redis connection statistics
func (c *Cache) GetStats() map[string]interface{} {
	stats := c.Client.PoolStats()
	return map[string]interface{}{
		"hits":         stats.Hits,
		"misses":       stats.Misses,
		"timeouts":     stats.Timeouts,
		"total_conns":  stats.TotalConns,
		"idle_conns":   stats.IdleConns,
		"stale_conns":  stats.StaleConns,
	}
}

// Session Management Methods

// CreateSession creates a new session in Redis
func (c *Cache) CreateSession(ctx context.Context, sessionID string, data *SessionData, expiry time.Duration) error {
	key := c.sessionKey(sessionID)
	
	// Set session data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := c.Client.Set(ctx, key, jsonData, expiry).Err(); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Add to user's session set for tracking concurrent sessions
	userSessionsKey := c.userSessionsKey(data.UserID.String())
	if err := c.Client.SAdd(ctx, userSessionsKey, sessionID).Err(); err != nil {
		c.Logger.Warn("Failed to add session to user sessions set", "error", err)
	}

	// Set expiry for user sessions set
	if err := c.Client.Expire(ctx, userSessionsKey, expiry).Err(); err != nil {
		c.Logger.Warn("Failed to set expiry for user sessions set", "error", err)
	}

	// Log session creation for cannabis compliance
	c.Logger.LogSessionActivity(sessionID, data.UserID.String(), "session_created", map[string]interface{}{
		"ip_address":        data.IPAddress,
		"user_agent":        data.UserAgent,
		"compliance_status": data.ComplianceStatus,
		"state":            data.State,
		"age_verified":     data.AgeVerified,
		"state_verified":   data.StateVerified,
	})

	return nil
}

// GetSession retrieves a session from Redis
func (c *Cache) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	key := c.sessionKey(sessionID)
	
	result, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

// UpdateSession updates an existing session in Redis
func (c *Cache) UpdateSession(ctx context.Context, sessionID string, data *SessionData) error {
	key := c.sessionKey(sessionID)
	
	// Get current TTL
	ttl, err := c.Client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get session TTL: %w", err)
	}

	// Update session data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := c.Client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Log session update for cannabis compliance
	c.Logger.LogSessionActivity(sessionID, data.UserID.String(), "session_updated", map[string]interface{}{
		"last_activity":     data.LastActivity,
		"compliance_status": data.ComplianceStatus,
		"compliance_verified": data.ComplianceVerified,
	})

	return nil
}

// DeleteSession removes a session from Redis
func (c *Cache) DeleteSession(ctx context.Context, sessionID string) error {
	key := c.sessionKey(sessionID)
	
	// Get session data before deletion for logging
	sessionData, err := c.GetSession(ctx, sessionID)
	if err != nil {
		c.Logger.Warn("Failed to get session data before deletion", "error", err)
	}

	// Delete session
	if err := c.Client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from user's session set
	if sessionData != nil {
		userSessionsKey := c.userSessionsKey(sessionData.UserID.String())
		if err := c.Client.SRem(ctx, userSessionsKey, sessionID).Err(); err != nil {
			c.Logger.Warn("Failed to remove session from user sessions set", "error", err)
		}

		// Log session deletion for cannabis compliance
		c.Logger.LogSessionActivity(sessionID, sessionData.UserID.String(), "session_deleted", map[string]interface{}{
			"reason": "explicit_deletion",
		})
	}

	return nil
}

// RefreshSession extends the session expiry time
func (c *Cache) RefreshSession(ctx context.Context, sessionID string, expiry time.Duration) error {
	key := c.sessionKey(sessionID)
	
	if err := c.Client.Expire(ctx, key, expiry).Err(); err != nil {
		return fmt.Errorf("failed to refresh session: %w", err)
	}

	return nil
}

// GetUserSessions returns all active sessions for a user
func (c *Cache) GetUserSessions(ctx context.Context, userID string) ([]string, error) {
	key := c.userSessionsKey(userID)
	
	members, err := c.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return members, nil
}

// DeleteUserSessions removes all sessions for a user
func (c *Cache) DeleteUserSessions(ctx context.Context, userID string) error {
	sessionIDs, err := c.GetUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Delete individual sessions
	for _, sessionID := range sessionIDs {
		if err := c.DeleteSession(ctx, sessionID); err != nil {
			c.Logger.Warn("Failed to delete user session", "session_id", sessionID, "error", err)
		}
	}

	// Delete user sessions set
	userSessionsKey := c.userSessionsKey(userID)
	if err := c.Client.Del(ctx, userSessionsKey).Err(); err != nil {
		return fmt.Errorf("failed to delete user sessions set: %w", err)
	}

	// Log user sessions deletion for cannabis compliance
	c.Logger.LogSessionActivity("", userID, "user_sessions_deleted", map[string]interface{}{
		"sessions_count": len(sessionIDs),
		"reason":        "user_logout_all",
	})

	return nil
}

// CheckSessionLimit checks if user has reached session limit
func (c *Cache) CheckSessionLimit(ctx context.Context, userID string, maxSessions int) (bool, error) {
	sessionIDs, err := c.GetUserSessions(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user sessions: %w", err)
	}

	// Clean up expired sessions
	validSessions := 0
	for _, sessionID := range sessionIDs {
		if exists, err := c.Client.Exists(ctx, c.sessionKey(sessionID)).Result(); err == nil && exists > 0 {
			validSessions++
		} else {
			// Remove expired session from user sessions set
			userSessionsKey := c.userSessionsKey(userID)
			c.Client.SRem(ctx, userSessionsKey, sessionID)
		}
	}

	return validSessions >= maxSessions, nil
}

// Cannabis-specific cache operations

// SetCannabisCompliance stores cannabis compliance data
func (c *Cache) SetCannabisCompliance(ctx context.Context, userID string, compliance map[string]interface{}, expiry time.Duration) error {
	key := c.complianceKey(userID)
	
	jsonData, err := json.Marshal(compliance)
	if err != nil {
		return fmt.Errorf("failed to marshal compliance data: %w", err)
	}

	if err := c.Client.Set(ctx, key, jsonData, expiry).Err(); err != nil {
		return fmt.Errorf("failed to set compliance data: %w", err)
	}

	// Log compliance data caching
	c.Logger.LogCannabisAudit(userID, "compliance_cached", "redis", map[string]interface{}{
		"compliance_data": compliance,
		"expiry":         expiry.String(),
	})

	return nil
}

// GetCannabisCompliance retrieves cannabis compliance data
func (c *Cache) GetCannabisCompliance(ctx context.Context, userID string) (map[string]interface{}, error) {
	key := c.complianceKey(userID)
	
	result, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("compliance data not found")
		}
		return nil, fmt.Errorf("failed to get compliance data: %w", err)
	}

	var compliance map[string]interface{}
	if err := json.Unmarshal([]byte(result), &compliance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal compliance data: %w", err)
	}

	return compliance, nil
}

// General cache operations

// Set stores a value in Redis
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.Client.Set(ctx, key, jsonData, expiry).Err()
}

// Get retrieves a value from Redis
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	result, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	return json.Unmarshal([]byte(result), dest)
}

// Delete removes a key from Redis
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.Client.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Helper methods for key generation

// sessionKey generates a Redis key for session data
func (c *Cache) sessionKey(sessionID string) string {
	return fmt.Sprintf("greenlync:session:%s", sessionID)
}

// userSessionsKey generates a Redis key for user sessions set
func (c *Cache) userSessionsKey(userID string) string {
	return fmt.Sprintf("greenlync:user_sessions:%s", userID)
}

// complianceKey generates a Redis key for compliance data
func (c *Cache) complianceKey(userID string) string {
	return fmt.Sprintf("greenlync:compliance:%s", userID)
}