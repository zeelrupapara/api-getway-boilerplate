package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/oauth2"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/db"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/messaging"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/internal/middleware"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	OAuth2Service *oauth2.Service
	Logger        *logger.Logger
	DB            *db.Database
	Messaging     *messaging.NATS
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	
	// Optional metadata
	DeviceType string `json:"device_type,omitempty"`
	Location   string `json:"location,omitempty"`
}

// LoginResponse represents login response
type LoginResponse struct {
	*oauth2.TokenPair
	User         UserInfo `json:"user"`
	Compliance   ComplianceInfo `json:"compliance"`
	Message      string `json:"message"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	ID           string `json:"id"`
	DispensaryID string `json:"dispensary_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Role         string `json:"role"`
	IsActive     bool   `json:"is_active"`
}

// ComplianceInfo represents cannabis compliance information
type ComplianceInfo struct {
	AgeVerified          bool   `json:"age_verified"`
	StateVerified        bool   `json:"state_verified"`
	ComplianceStatus     string `json:"compliance_status"`
	State                string `json:"state"`
	ComplianceRequired   bool   `json:"compliance_required"`
	MinimumAge          int    `json:"minimum_age"`
	LegalStatesRequired bool   `json:"legal_states_required"`
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	LogoutAll bool `json:"logout_all,omitempty"`
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(oauth2Service *oauth2.Service, logger *logger.Logger, db *db.Database, messaging *messaging.NATS) *AuthHandler {
	return &AuthHandler{
		OAuth2Service: oauth2Service,
		Logger:        logger,
		DB:            db,
		Messaging:     messaging,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.Logger.Warn("Invalid login request body",
			"error", err,
			"ip", c.IP(),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Find user by email
	var user v1.User
	if err := h.DB.DB.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			h.Logger.Warn("Login attempt with non-existent email",
				"email", req.Email,
				"ip", c.IP(),
			)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid email or password",
				"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
			})
		}
		
		h.Logger.Error("Database error during login",
			"error", err,
			"email", req.Email,
			"ip", c.IP(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.Logger.Warn("Login attempt with invalid password",
			"email", req.Email,
			"user_id", user.ID.String(),
			"ip", c.IP(),
		)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
			"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
		})
	}

	// Create session metadata
	sessionMetadata := map[string]string{
		"ip_address": c.IP(),
		"user_agent": c.Get("User-Agent"),
		"device_type": req.DeviceType,
		"location":   req.Location,
	}

	// Generate tokens
	tokenPair, err := h.OAuth2Service.GenerateTokens(c.Context(), &user, sessionMetadata)
	if err != nil {
		h.Logger.Error("Failed to generate tokens",
			"error", err,
			"user_id", user.ID.String(),
			"ip", c.IP(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate authentication tokens",
		})
	}

	// Update user last login
	now := time.Now()
	user.LastLoginAt = &now
	if err := h.DB.DB.Save(&user).Error; err != nil {
		h.Logger.Warn("Failed to update last login time",
			"error", err,
			"user_id", user.ID.String(),
		)
	}

	// Publish login event
	h.Messaging.PublishUserEvent("login", user.ID.String(), user.DispensaryID.String(), map[string]interface{}{
		"ip_address":        c.IP(),
		"user_agent":        c.Get("User-Agent"),
		"device_type":       req.DeviceType,
		"location":          req.Location,
		"session_id":        tokenPair.SessionID,
		"compliance_status": user.ComplianceStatus,
	})

	// Log successful login for cannabis compliance
	h.Logger.LogCannabisAudit(user.ID.String(), "user_login", "authentication", map[string]interface{}{
		"dispensary_id":     user.DispensaryID.String(),
		"session_id":        tokenPair.SessionID,
		"ip_address":        c.IP(),
		"user_agent":        c.Get("User-Agent"),
		"compliance_status": user.ComplianceStatus,
		"age_verified":      user.AgeVerified,
		"state_verified":    user.StateVerified,
	})

	response := LoginResponse{
		TokenPair: tokenPair,
		User: UserInfo{
			ID:           user.ID.String(),
			DispensaryID: user.DispensaryID.String(),
			Email:        user.Email,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Role:         string(user.Role),
			IsActive:     user.IsActive,
		},
		Compliance: ComplianceInfo{
			AgeVerified:          user.AgeVerified,
			StateVerified:        user.StateVerified,
			ComplianceStatus:     string(user.ComplianceStatus),
			State:                user.State,
			ComplianceRequired:   true,
			MinimumAge:          21,
			LegalStatesRequired: true,
		},
		Message: "Login successful",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		h.Logger.Warn("Invalid refresh request body",
			"error", err,
			"ip", c.IP(),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Refresh tokens
	tokenPair, err := h.OAuth2Service.RefreshTokens(c.Context(), req.RefreshToken)
	if err != nil {
		h.Logger.Warn("Token refresh failed",
			"error", err,
			"ip", c.IP(),
		)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
			"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":            tokenPair.AccessToken,
		"refresh_token":           tokenPair.RefreshToken,
		"access_token_expires_at": tokenPair.AccessTokenExpiresAt,
		"token_type":              tokenPair.TokenType,
		"session_id":              tokenPair.SessionID,
		"message":                 "Token refreshed successfully",
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		// If body parsing fails, just logout current session
		req.LogoutAll = false
	}

	if req.LogoutAll {
		// Logout all sessions
		if err := h.OAuth2Service.RevokeAllUserSessions(c.Context(), authCtx.UserID.String()); err != nil {
			h.Logger.Error("Failed to revoke all user sessions",
				"error", err,
				"user_id", authCtx.UserID.String(),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to logout from all sessions",
			})
		}

		// Publish logout all event
		h.Messaging.PublishUserEvent("logout_all", authCtx.UserID.String(), authCtx.DispensaryID.String(), map[string]interface{}{
			"ip_address": c.IP(),
			"user_agent": c.Get("User-Agent"),
		})

		h.Logger.LogCannabisAudit(authCtx.UserID.String(), "user_logout_all", "authentication", map[string]interface{}{
			"dispensary_id": authCtx.DispensaryID.String(),
			"ip_address":    c.IP(),
			"user_agent":    c.Get("User-Agent"),
		})

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Logged out from all sessions successfully",
		})
	} else {
		// Logout current session only
		if err := h.OAuth2Service.RevokeSession(c.Context(), authCtx.SessionID); err != nil {
			h.Logger.Error("Failed to revoke session",
				"error", err,
				"session_id", authCtx.SessionID,
				"user_id", authCtx.UserID.String(),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to logout",
			})
		}

		// Publish logout event
		h.Messaging.PublishUserEvent("logout", authCtx.UserID.String(), authCtx.DispensaryID.String(), map[string]interface{}{
			"session_id": authCtx.SessionID,
			"ip_address": c.IP(),
			"user_agent": c.Get("User-Agent"),
		})

		h.Logger.LogCannabisAudit(authCtx.UserID.String(), "user_logout", "authentication", map[string]interface{}{
			"session_id":    authCtx.SessionID,
			"dispensary_id": authCtx.DispensaryID.String(),
			"ip_address":    c.IP(),
			"user_agent":    c.Get("User-Agent"),
		})

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Logged out successfully",
		})
	}
}

// GetProfile returns current user profile
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Get user from database
	var user v1.User
	if err := h.DB.DB.Where("id = ?", authCtx.UserID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			h.Logger.Warn("User not found for authenticated session",
				"user_id", authCtx.UserID.String(),
				"session_id", authCtx.SessionID,
			)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		
		h.Logger.Error("Database error getting user profile",
			"error", err,
			"user_id", authCtx.UserID.String(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	profile := map[string]interface{}{
		"id":            user.ID.String(),
		"dispensary_id": user.DispensaryID.String(),
		"email":         user.Email,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"phone":         user.Phone,
		"role":          string(user.Role),
		"is_active":     user.IsActive,
		"date_of_birth": user.DateOfBirth,
		"state":         user.State,
		"address":       user.Address,
		"city":          user.City,
		"zip_code":      user.ZipCode,
		"avatar":        user.Avatar,
		"bio":           user.Bio,
		"compliance": map[string]interface{}{
			"age_verified":          user.AgeVerified,
			"state_verified":        user.StateVerified,
			"compliance_status":     string(user.ComplianceStatus),
			"age_verified_at":       user.AgeVerifiedAt,
			"state_verified_at":     user.StateVerifiedAt,
			"last_compliance_check": user.LastComplianceCheckAt,
		},
		"cannabis_profile": map[string]interface{}{
			"preferred_products":      user.PreferredProducts,
			"consumption_goals":       user.ConsumptionGoals,
			"medical_recommendation": user.MedicalRecommendation,
		},
		"session": map[string]interface{}{
			"session_id":     authCtx.SessionID,
			"last_activity":  authCtx.LastActivity,
			"expires_at":     authCtx.ExpiresAt,
			"ip_address":     authCtx.IPAddress,
			"device_type":    authCtx.DeviceType,
		},
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
		"last_login_at": user.LastLoginAt,
	}

	return c.Status(fiber.StatusOK).JSON(profile)
}

// VerifyCompliance handles cannabis compliance verification
func (h *AuthHandler) VerifyCompliance(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	type ComplianceVerifyRequest struct {
		AgeVerified   bool   `json:"age_verified"`
		StateVerified bool   `json:"state_verified"`
		State         string `json:"state,omitempty"`
		DateOfBirth   *time.Time `json:"date_of_birth,omitempty"`
	}

	var req ComplianceVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update compliance in OAuth2 session
	if err := h.OAuth2Service.VerifyComplianceStatus(c.Context(), authCtx.SessionID, req.AgeVerified, req.StateVerified); err != nil {
		h.Logger.Error("Failed to update session compliance",
			"error", err,
			"session_id", authCtx.SessionID,
			"user_id", authCtx.UserID.String(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update compliance status",
		})
	}

	// Update user in database
	var user v1.User
	if err := h.DB.DB.Where("id = ?", authCtx.UserID).First(&user).Error; err != nil {
		h.Logger.Error("Failed to get user for compliance update",
			"error", err,
			"user_id", authCtx.UserID.String(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user compliance",
		})
	}

	// Update compliance fields
	user.AgeVerified = req.AgeVerified
	user.StateVerified = req.StateVerified
	now := time.Now()
	user.LastComplianceCheckAt = &now

	if req.AgeVerified {
		user.AgeVerifiedAt = &now
	}
	if req.StateVerified {
		user.StateVerifiedAt = &now
	}
	if req.State != "" {
		user.State = req.State
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = req.DateOfBirth
	}

	// Update compliance status
	if req.AgeVerified && req.StateVerified {
		user.ComplianceStatus = v1.ComplianceStatusVerified
	} else {
		user.ComplianceStatus = v1.ComplianceStatusPending
	}

	// Save user
	if err := h.DB.DB.Save(&user).Error; err != nil {
		h.Logger.Error("Failed to save user compliance update",
			"error", err,
			"user_id", authCtx.UserID.String(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save compliance status",
		})
	}

	// Publish compliance event
	h.Messaging.PublishComplianceEvent("verified", authCtx.UserID.String(), authCtx.DispensaryID.String(), map[string]interface{}{
		"age_verified":      req.AgeVerified,
		"state_verified":    req.StateVerified,
		"compliance_status": user.ComplianceStatus,
		"state":            user.State,
		"ip_address":       c.IP(),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Compliance status updated successfully",
		"compliance": map[string]interface{}{
			"age_verified":      user.AgeVerified,
			"state_verified":    user.StateVerified,
			"compliance_status": string(user.ComplianceStatus),
			"state":            user.State,
			"verified_at":      now,
		},
	})
}

// GetSessions returns all active sessions for the current user
func (h *AuthHandler) GetSessions(c *fiber.Ctx) error {
	authCtx, ok := middleware.GetAuthContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	sessionIDs, err := h.OAuth2Service.Cache.GetUserSessions(c.Context(), authCtx.UserID.String())
	if err != nil {
		h.Logger.Error("Failed to get user sessions",
			"error", err,
			"user_id", authCtx.UserID.String(),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sessions",
		})
	}

	sessions := make([]map[string]interface{}, 0)
	for _, sessionID := range sessionIDs {
		sessionData, err := h.OAuth2Service.Cache.GetSession(c.Context(), sessionID)
		if err != nil {
			h.Logger.Warn("Failed to get session data",
				"error", err,
				"session_id", sessionID,
			)
			continue
		}

		sessions = append(sessions, map[string]interface{}{
			"session_id":         sessionID,
			"is_current":         sessionID == authCtx.SessionID,
			"created_at":         sessionData.CreatedAt,
			"last_activity":      sessionData.LastActivity,
			"expires_at":         sessionData.ExpiresAt,
			"ip_address":         sessionData.IPAddress,
			"user_agent":         sessionData.UserAgent,
			"device_type":        sessionData.DeviceType,
			"location":           sessionData.Location,
			"compliance_verified": sessionData.ComplianceVerified,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"sessions": sessions,
		"total":    len(sessions),
	})
}