package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/config"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/cache"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// Service represents the OAuth2 JWT service
type Service struct {
	Config *config.JWTConfig
	Logger *logger.Logger
	Cache  *cache.Cache
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"`
	SessionID             string    `json:"session_id"`
}

// Claims represents JWT claims for cannabis platform
type Claims struct {
	UserID       string `json:"user_id"`
	DispensaryID string `json:"dispensary_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	SessionID    string `json:"session_id"`
	
	// Cannabis compliance claims
	AgeVerified      bool   `json:"age_verified"`
	StateVerified    bool   `json:"state_verified"`
	ComplianceStatus string `json:"compliance_status"`
	State            string `json:"state"`
	
	// Token metadata
	TokenType string `json:"token_type"` // "access" or "refresh"
	
	jwt.RegisteredClaims
}

// AuthContext represents authenticated user context
type AuthContext struct {
	UserID       uuid.UUID `json:"user_id"`
	DispensaryID uuid.UUID `json:"dispensary_id"`
	Email        string    `json:"email"`
	Role         v1.UserRole `json:"role"`
	SessionID    string    `json:"session_id"`
	
	// Cannabis compliance context
	AgeVerified      bool                 `json:"age_verified"`
	StateVerified    bool                 `json:"state_verified"`
	ComplianceStatus v1.ComplianceStatus  `json:"compliance_status"`
	State            string               `json:"state"`
	
	// Session metadata
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DeviceType   string    `json:"device_type"`
	Location     string    `json:"location"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewService creates a new OAuth2 JWT service
func NewService(cfg *config.JWTConfig, log *logger.Logger, cache *cache.Cache) *Service {
	return &Service{
		Config: cfg,
		Logger: log,
		Cache:  cache,
	}
}

// Authentication methods

// GenerateTokens generates access and refresh tokens for a user
func (s *Service) GenerateTokens(ctx context.Context, user *v1.User, sessionMetadata map[string]string) (*TokenPair, error) {
	sessionID := s.generateSessionID()
	
	// Check session limit
	exceeded, err := s.Cache.CheckSessionLimit(ctx, user.ID.String(), s.Config.MaxSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to check session limit: %w", err)
	}
	
	if exceeded {
		// Remove oldest session if limit exceeded
		if err := s.cleanupOldestSession(ctx, user.ID.String()); err != nil {
			s.Logger.Warn("Failed to cleanup oldest session", "error", err)
		}
	}

	// Create access token claims
	accessClaims := &Claims{
		UserID:           user.ID.String(),
		DispensaryID:     user.DispensaryID.String(),
		Email:            user.Email,
		Role:             string(user.Role),
		SessionID:        sessionID,
		AgeVerified:      user.AgeVerified,
		StateVerified:    user.StateVerified,
		ComplianceStatus: string(user.ComplianceStatus),
		State:            user.State,
		TokenType:        "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Config.Issuer,
			Audience:  []string{s.Config.Audience},
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Config.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	// Create refresh token claims
	refreshClaims := &Claims{
		UserID:           user.ID.String(),
		DispensaryID:     user.DispensaryID.String(),
		Email:            user.Email,
		Role:             string(user.Role),
		SessionID:        sessionID,
		AgeVerified:      user.AgeVerified,
		StateVerified:    user.StateVerified,
		ComplianceStatus: string(user.ComplianceStatus),
		State:            user.State,
		TokenType:        "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Config.Issuer,
			Audience:  []string{s.Config.Audience},
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	// Generate tokens
	accessToken, err := s.generateToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session data
	sessionData := &cache.SessionData{
		UserID:               user.ID,
		DispensaryID:         user.DispensaryID,
		Role:                 string(user.Role),
		Email:                user.Email,
		AgeVerified:          user.AgeVerified,
		StateVerified:        user.StateVerified,
		ComplianceStatus:     string(user.ComplianceStatus),
		State:                user.State,
		IPAddress:            sessionMetadata["ip_address"],
		UserAgent:            sessionMetadata["user_agent"],
		DeviceType:           sessionMetadata["device_type"],
		Location:             sessionMetadata["location"],
		CreatedAt:            time.Now(),
		LastActivity:         time.Now(),
		ExpiresAt:            accessClaims.ExpiresAt.Time,
		MaxSessions:          s.Config.MaxSessions,
		CurrentSessions:      1,
		ComplianceVerified:   user.IsCompliant(),
		ComplianceCheckedAt:  &time.Time{},
	}

	// Store session in Redis
	if err := s.Cache.CreateSession(ctx, sessionID, sessionData, s.Config.RefreshExpiry); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	tokenPair := &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time,
		TokenType:             "Bearer",
		SessionID:             sessionID,
	}

	// Log authentication success for cannabis compliance
	s.Logger.LogCannabisAudit(user.ID.String(), "tokens_generated", "authentication", map[string]interface{}{
		"session_id":        sessionID,
		"dispensary_id":     user.DispensaryID.String(),
		"compliance_status": user.ComplianceStatus,
		"age_verified":      user.AgeVerified,
		"state_verified":    user.StateVerified,
		"ip_address":        sessionMetadata["ip_address"],
		"user_agent":        sessionMetadata["user_agent"],
	})

	return tokenPair, nil
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.Config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional validation for cannabis compliance
	if claims.TokenType != "access" && claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: %s", claims.TokenType)
	}

	return claims, nil
}

// RefreshTokens refreshes access token using refresh token
func (s *Service) RefreshTokens(ctx context.Context, refreshTokenString string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Get session data
	sessionData, err := s.Cache.GetSession(ctx, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if session is still valid
	if time.Now().After(sessionData.ExpiresAt) {
		if err := s.Cache.DeleteSession(ctx, claims.SessionID); err != nil {
			s.Logger.Warn("Failed to delete expired session", "error", err)
		}
		return nil, fmt.Errorf("session expired")
	}

	// Create new access token claims
	accessClaims := &Claims{
		UserID:           claims.UserID,
		DispensaryID:     claims.DispensaryID,
		Email:            claims.Email,
		Role:             claims.Role,
		SessionID:        claims.SessionID,
		AgeVerified:      claims.AgeVerified,
		StateVerified:    claims.StateVerified,
		ComplianceStatus: claims.ComplianceStatus,
		State:            claims.State,
		TokenType:        "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.Config.Issuer,
			Audience:  []string{s.Config.Audience},
			Subject:   claims.Subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.Config.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	// Generate new access token
	accessToken, err := s.generateToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	// Update session last activity
	sessionData.LastActivity = time.Now()
	if err := s.Cache.UpdateSession(ctx, claims.SessionID, sessionData); err != nil {
		s.Logger.Warn("Failed to update session last activity", "error", err)
	}

	tokenPair := &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshTokenString, // Keep the same refresh token
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: claims.ExpiresAt.Time,
		TokenType:             "Bearer",
		SessionID:             claims.SessionID,
	}

	// Log token refresh for cannabis compliance
	s.Logger.LogSessionActivity(claims.SessionID, claims.UserID, "tokens_refreshed", map[string]interface{}{
		"dispensary_id": claims.DispensaryID,
		"new_expires_at": accessClaims.ExpiresAt.Time,
	})

	return tokenPair, nil
}

// GetAuthContext retrieves authentication context from session
func (s *Service) GetAuthContext(ctx context.Context, sessionID string) (*AuthContext, error) {
	sessionData, err := s.Cache.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check session expiry
	if time.Now().After(sessionData.ExpiresAt) {
		if err := s.Cache.DeleteSession(ctx, sessionID); err != nil {
			s.Logger.Warn("Failed to delete expired session", "error", err)
		}
		return nil, fmt.Errorf("session expired")
	}

	// Update last activity
	sessionData.LastActivity = time.Now()
	if err := s.Cache.UpdateSession(ctx, sessionID, sessionData); err != nil {
		s.Logger.Warn("Failed to update session last activity", "error", err)
	}

	authContext := &AuthContext{
		UserID:           sessionData.UserID,
		DispensaryID:     sessionData.DispensaryID,
		Email:            sessionData.Email,
		Role:             v1.UserRole(sessionData.Role),
		SessionID:        sessionID,
		AgeVerified:      sessionData.AgeVerified,
		StateVerified:    sessionData.StateVerified,
		ComplianceStatus: v1.ComplianceStatus(sessionData.ComplianceStatus),
		State:            sessionData.State,
		IPAddress:        sessionData.IPAddress,
		UserAgent:        sessionData.UserAgent,
		DeviceType:       sessionData.DeviceType,
		Location:         sessionData.Location,
		LastActivity:     sessionData.LastActivity,
		ExpiresAt:        sessionData.ExpiresAt,
	}

	return authContext, nil
}

// RevokeSession revokes a specific session
func (s *Service) RevokeSession(ctx context.Context, sessionID string) error {
	// Get session data for logging
	sessionData, err := s.Cache.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Delete session
	if err := s.Cache.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Log session revocation for cannabis compliance
	s.Logger.LogSessionActivity(sessionID, sessionData.UserID.String(), "session_revoked", map[string]interface{}{
		"dispensary_id": sessionData.DispensaryID.String(),
		"reason":       "explicit_revocation",
	})

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *Service) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if err := s.Cache.DeleteUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke all user sessions: %w", err)
	}

	// Log all sessions revocation for cannabis compliance
	s.Logger.LogSessionActivity("", userID, "all_sessions_revoked", map[string]interface{}{
		"reason": "user_logout_all",
	})

	return nil
}

// Cannabis compliance methods

// VerifyComplianceStatus checks and updates user compliance status
func (s *Service) VerifyComplianceStatus(ctx context.Context, sessionID string, ageVerified, stateVerified bool) error {
	sessionData, err := s.Cache.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Update compliance status
	sessionData.AgeVerified = ageVerified
	sessionData.StateVerified = stateVerified
	sessionData.ComplianceVerified = ageVerified && stateVerified
	now := time.Now()
	sessionData.ComplianceCheckedAt = &now

	// Update compliance status
	if sessionData.ComplianceVerified {
		sessionData.ComplianceStatus = string(v1.ComplianceStatusVerified)
	} else {
		sessionData.ComplianceStatus = string(v1.ComplianceStatusPending)
	}

	// Update session
	if err := s.Cache.UpdateSession(ctx, sessionID, sessionData); err != nil {
		return fmt.Errorf("failed to update session compliance: %w", err)
	}

	// Log compliance verification for cannabis audit
	s.Logger.LogCannabisAudit(sessionData.UserID.String(), "compliance_verified", "session", map[string]interface{}{
		"session_id":     sessionID,
		"age_verified":   ageVerified,
		"state_verified": stateVerified,
		"compliance_verified": sessionData.ComplianceVerified,
		"dispensary_id":  sessionData.DispensaryID.String(),
	})

	return nil
}

// IsSessionCompliant checks if session meets cannabis compliance requirements
func (s *Service) IsSessionCompliant(ctx context.Context, sessionID string) (bool, error) {
	sessionData, err := s.Cache.GetSession(ctx, sessionID)
	if err != nil {
		return false, fmt.Errorf("session not found: %w", err)
	}

	return sessionData.ComplianceVerified, nil
}

// Helper methods

// generateToken generates a JWT token with claims
func (s *Service) generateToken(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Config.Secret))
}

// generateSessionID generates a unique session ID
func (s *Service) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// cleanupOldestSession removes the oldest session for a user
func (s *Service) cleanupOldestSession(ctx context.Context, userID string) error {
	sessionIDs, err := s.Cache.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// Find oldest session (simplified - remove first one)
	oldestSessionID := sessionIDs[0]
	return s.Cache.DeleteSession(ctx, oldestSessionID)
}

// Utility methods for middleware

// ExtractTokenFromHeader extracts token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// Cannabis compliance validation

// ValidateCannabisAccess validates if user can access cannabis-related resources
func (s *Service) ValidateCannabisAccess(authCtx *AuthContext) error {
	if !authCtx.AgeVerified {
		return fmt.Errorf("age verification required for cannabis access")
	}

	if !authCtx.StateVerified {
		return fmt.Errorf("state verification required for cannabis access")
	}

	if authCtx.ComplianceStatus != v1.ComplianceStatusVerified {
		return fmt.Errorf("compliance verification required for cannabis access")
	}

	return nil
}

// CheckRolePermission checks if user role has permission for action
func (s *Service) CheckRolePermission(role v1.UserRole, action string) bool {
	permissions := map[v1.UserRole][]string{
		v1.RoleCustomer: {
			"view_products", "create_order", "view_order", "cancel_order",
			"view_profile", "update_profile",
		},
		v1.RoleBudtender: {
			"view_products", "manage_inventory", "process_order", "view_customers",
			"view_reports", "manage_pos",
		},
		v1.RoleDispensaryManager: {
			"manage_products", "manage_inventory", "manage_orders", "manage_users",
			"view_reports", "manage_dispensary", "manage_compliance",
		},
		v1.RoleBrandPartner: {
			"manage_brand_products", "view_brand_reports", "manage_brand_inventory",
		},
		v1.RoleSystemAdmin: {
			"manage_all", "view_all", "delete_all", "configure_system",
		},
	}

	rolePermissions, exists := permissions[role]
	if !exists {
		return false
	}

	// System admin has all permissions
	if role == v1.RoleSystemAdmin {
		return true
	}

	// Check specific permission
	for _, permission := range rolePermissions {
		if permission == action {
			return true
		}
	}

	return false
}