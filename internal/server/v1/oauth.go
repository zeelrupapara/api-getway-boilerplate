// Developer: zeelrupapara@gmail.com
// Description: Cannabis user authentication for GreenLync
package v1

import (
	"context"
	"fmt"
	"time"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/errors"
	"greenlync-api-gateway/pkg/http"
	"greenlync-api-gateway/pkg/oauth2"
	"greenlync-api-gateway/utils"


	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	GRANT_TYPE_CLIENT_CREDENTIALS = "client_credentials"
	GRANT_TYPE_PASSWORD           = "password"
)
const MAX_SESSIONS = 5

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserId       int32  `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionId    string `json:"session_id"`
	ExpiresIn    int32  `json:"expires_in"`
	IpAddress    string `json:"ip_address"`
	Scope        string `json:"scope"`
}

type SessionEvent struct {
	UserId    int32  `json:"user_id"`
	SessionId string `json:"session_id"`
	Scope     string `json:"scope"`
	Ts        int64  `json:"ts" format:"int64"`
}

//	@Id				Login
//	@Description	Login using account credentials passed using basic auth method
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	LoginResponse
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BasicAuth
//	@Authorization:	Basic username:password
//	@Param			remember_me	query	boolean	false	"remember me"
//	@Router			/auth/v1/oauth2/login [post]
func (s *HttpServer) Login(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	password := c.Locals("password").(string)

	rememberMe := c.QueryBool("remember_me", false)

	grantType := c.Query("grant_type", GRANT_TYPE_PASSWORD)
	if grantType == "" {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("grant_type %s", errors.RequiredField))
	}

	user := &model.User{}
	query := s.DB
	if grantType == GRANT_TYPE_CLIENT_CREDENTIALS {
		query = query.Where("username = ?", username)
	} else {
		query = query.Where("id = ? OR username = ?", username, username)
	}
	err := query.First(user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseUnauthorized(c, fmt.Errorf("incorrect password or username"))
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var ok bool
	if grantType == GRANT_TYPE_CLIENT_CREDENTIALS || grantType == GRANT_TYPE_PASSWORD {
		// Simplified password check for boilerplate - just compare password hash
		ok = oauth2.ComparePassword(user.PasswordHash, password)
	}
	if !ok {
		return s.App.HttpResponseUnauthorized(c, fmt.Errorf("incorrect username or password"))
	}

	// check account is active
	if !user.IsActive {
		return s.App.HttpResponseUnauthorized(c, fmt.Errorf("the account has been deactivated"))
	}

	// Simplified session management for boilerplate
	// TODO: Implement proper multi-session management
	oldestSession, count := s.OAuth2.GetActiveSessionsCountByClientId(user.Id)
	
	// Simple single session limit for boilerplate
	if count >= 1 {
		// kill the oldest session
		s.killClientSession(oldestSession, SessionDescionnectionReason_SessionsLimit)

		// log the operation
		s.queueSystemOperationLog(&model.OperationsLog{
			Action:    "logout",
			Resource:  "session",
			UserId:    oldestSession.ClientId,
			Method:    "DELETE",
			URL:       c.OriginalURL(),
			IpAddress: oldestSession.IpAddress,
			UserAgent: c.Get("User-Agent"),
		})
	}

	// Simplified role handling for boilerplate
	// Use the role field from User model directly
	role := &model.Role{
		Id:   1, // Default role ID
		Desc: user.Role,
	}

	cfg := &oauth2.Config{
		ClientId:       user.Id,
		ClientSecretId: "",
		Scope:          role.Desc,
		IpAddress:      utils.GetRealIP(c),
		ExpiresIn:      s.OAuth2.TokenExpiresIn,
		UserAgent:      utils.GetUserAgent(c),
		RememberMe:     rememberMe,
	}
	if rememberMe {
		cfg.ExpiresIn = s.OAuth2.LongTokenExpiresIn
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = s.OAuth2.PasswordCredentialsToken(ctxTimeout, cfg)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, fiber.ErrInternalServerError)
	}

	res := &LoginResponse{
		UserId:       cfg.ClientId,
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		SessionId:    cfg.SessionId,
		ExpiresIn:    int32(cfg.ExpiresIn),
		Scope:        role.Desc,
		IpAddress:    cfg.IpAddress,
	}

	// TODO: Implement event logging for login events
	// sessionData, _ := json.Marshal(map[string]interface{}{
	// 	"id":         cfg.Id,
	// 	"user_id":    user.Id,
	// 	"started_at": time.Now(),
	// 	"client_id":  cfg.ClientId,
	// 	"session_id": cfg.SessionId,
	// 	"full_name":  user.FirstName + " " + user.LastName,
	// 	"ip_address": cfg.IpAddress,
	// })
	// TODO: Implement WebSocket publishing for online sessions
	// err = s.PublishWS(subject, event)
	// if err != nil {
	// 	s.Log.Logger.Errorf("error publishing to %s", subject)
	// }

	s.queueSystemOperationLog(&model.OperationsLog{
		Action:    "login",
		Resource:  "session",
		UserId:    cfg.ClientId,
		Method:    "POST",
		URL:       c.OriginalURL(),
		IpAddress: cfg.IpAddress,
		UserAgent: c.Get("User-Agent"),
	})

	return s.App.HttpResponseOK(c, res)
}

//	@Id				Token
//	@Description	Login using account client_id and client secret passed using basic auth method
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	LoginResponse
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BasicAuth
//	@Param			grant_type	query	string	true	"either client_credentials or password"
//	@Param			remember_me	query	boolean	false	"remember me"
//	@Router			/auth/v1/oauth2/token [post]
func (s *HttpServer) Token(c *fiber.Ctx) error {
	username := c.Locals("username").(string)
	password := c.Locals("password").(string)

	rememberMe := c.QueryBool("remember_me", false)

	grantType := c.Query("grant_type", GRANT_TYPE_PASSWORD)
	if grantType == "" {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("grant_type %s", errors.RequiredField))
	}

	user := &model.User{}
	query := s.DB
	if grantType == GRANT_TYPE_CLIENT_CREDENTIALS {
		query = query.Where("username = ?", username)
	} else {
		query = query.Where("id = ? OR username = ?", username, username)
	}
	err := query.First(user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.App.HttpResponseUnauthorized(c, fmt.Errorf("incorrect password or username"))
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	var ok bool
	if grantType == GRANT_TYPE_CLIENT_CREDENTIALS || grantType == GRANT_TYPE_PASSWORD {
		// Simplified password check for boilerplate - just compare password hash
		ok = oauth2.ComparePassword(user.PasswordHash, password)
	}
	if !ok {
		return s.App.HttpResponseUnauthorized(c, fmt.Errorf("incorrect username or password"))
	}

	// check account is active
	if !user.IsActive {
		return s.App.HttpResponseUnauthorized(c, fmt.Errorf("the account has been deactivated"))
	}

	// Simplified session management for boilerplate
	// TODO: Implement proper multi-session management
	oldestSession, count := s.OAuth2.GetActiveSessionsCountByClientId(user.Id)
	
	// Simple single session limit for boilerplate
	if count >= 1 {
		// kill the oldest session
		s.killClientSession(oldestSession, SessionDescionnectionReason_SessionsLimit)

		// log the operation
		s.queueSystemOperationLog(&model.OperationsLog{
			Action:    "logout",
			Resource:  "session",
			UserId:    oldestSession.ClientId,
			Method:    "DELETE",
			URL:       c.OriginalURL(),
			IpAddress: oldestSession.IpAddress,
			UserAgent: c.Get("User-Agent"),
		})
	}

	// Simplified role handling for boilerplate
	// Use the role field from User model directly
	role := &model.Role{
		Id:   1, // Default role ID
		Desc: user.Role,
	}

	cfg := &oauth2.Config{
		ClientId:       user.Id,
		ClientSecretId: "",
		Scope:          role.Desc,
		IpAddress:      utils.GetRealIP(c),
		ExpiresIn:      s.OAuth2.TokenExpiresIn,
		UserAgent:      utils.GetUserAgent(c),
		RememberMe:     rememberMe,
	}
	if rememberMe {
		cfg.ExpiresIn = s.OAuth2.LongTokenExpiresIn
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = s.OAuth2.PasswordCredentialsToken(ctxTimeout, cfg)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, fiber.ErrInternalServerError)
	}

	res := &LoginResponse{
		UserId:       cfg.ClientId,
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		SessionId:    cfg.SessionId,
		ExpiresIn:    int32(cfg.ExpiresIn),
		Scope:        role.Desc,
		IpAddress:    cfg.IpAddress,
	}

	// TODO: Implement event logging for login events
	// sessionData, _ := json.Marshal(map[string]interface{}{
	// 	"id":         cfg.Id,
	// 	"user_id":    user.Id,
	// 	"started_at": time.Now(),
	// 	"client_id":  cfg.ClientId,
	// 	"session_id": cfg.SessionId,
	// 	"full_name":  user.FirstName + " " + user.LastName,
	// 	"ip_address": cfg.IpAddress,
	// })
	// TODO: Implement WebSocket publishing for online sessions
	// err = s.PublishWS(subject, event)
	// if err != nil {
	// 	s.Log.Logger.Errorf("error publishing to %s", subject)
	// }

	s.queueSystemOperationLog(&model.OperationsLog{
		Action:    "login",
		Resource:  "session",
		UserId:    cfg.ClientId,
		Method:    "POST",
		URL:       c.OriginalURL(),
		IpAddress: cfg.IpAddress,
		UserAgent: c.Get("User-Agent"),
	})

	return s.App.HttpResponseOK(c, res)
}

type RefreshTokenBody struct {
	RefreshToken string `json:"refresh_token"`
}

//	@Id				RefreshToken
//	@Description	Refresh account's Token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	model.Token
//	@Failure		500		{object}	http.HttpResponse
//	@Param			body	body		RefreshTokenBody	true	"Refresh Token Request body"
//	@Router			/auth/v1/oauth2/refresh/token [post]
func (s *HttpServer) RefreshToken(c *fiber.Ctx) error {
	body := &RefreshTokenBody{}
	err := c.BodyParser(body)
	if err != nil {
		return s.App.HttpResponseBadRequest(c, err)
	}
	if body.RefreshToken == "" {
		return s.App.HttpResponseBadRequest(c, fmt.Errorf("refresh_token %s", errors.RequiredField))
	}

	// Get user ip and user agent
	ipAddr := utils.GetRealIP(c)
	useragent := utils.GetUserAgent(c)

	// if it exists then it's valid, otherwise it's not
	ctx := context.Background()
	cfg, err := s.OAuth2.RefreshToken(ctx, ipAddr, useragent, body.RefreshToken)
	if err != nil {
		if err == redis.Nil {
			return s.App.HttpResponseUnauthorized(c, errors.ErrInvalidToken)
		}
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	res := &LoginResponse{
		UserId:       cfg.ClientId,
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		ExpiresIn:    int32(cfg.ExpiresIn),
		SessionId:    cfg.SessionId,
		Scope:        cfg.Scope,
		IpAddress:    cfg.IpAddress,
	}

	return s.App.HttpResponseOK(c, res)
}

//	@Id				Logout
//	@Description	Logout
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		500	{object}	http.HttpResponse
//	@Security		BearerAuth
//	@Router			/auth/v1/oauth2/logout [DELETE]
func (s *HttpServer) Logout(c *fiber.Ctx) error {
	accessToken := c.Locals(http.LocalsToken).(string)

	ctx := context.Background()
	cfg, err := s.OAuth2.Inspect(ctx, accessToken)
	if err != nil {
		return s.App.HttpResponseNotFound(c, err)
	}

	err = s.OAuth2.Logout(ctx, cfg.AccessToken, cfg.SessionId)
	if err != nil {
		return s.App.HttpResponseInternalServerErrorRequest(c, err)
	}

	// logout from the websocket if there is a connection
	err = s.Hub.Delete(cfg.SessionId)
	if err != nil {
		s.Log.Logger.Error(err)
	}

	// TODO: Implement client disconnection publishing
	// s.publishClientDisconnected(cfg)

	s.queueSystemOperationLog(&model.OperationsLog{
		Action:    "logout",
		Resource:  "session",
		UserId:    cfg.ClientId,
		Method:    "DELETE",
		URL:       c.OriginalURL(),
		IpAddress: cfg.IpAddress,
		UserAgent: c.Get("User-Agent"),
	})

	return s.App.HttpResponseNoContent(c)
}

func (s *HttpServer) checkIfMutlipleSessionsAvailable(groupId int32) (bool, error) {
	gc := &model.ConfigGroup{}
	err := s.DB.Where("group_id = ?", groupId).First(gc).Error
	if err != nil {
		return false, err
	}

	// Simplified for boilerplate - always return false (single session only)
	return false, nil
}

type SessionDescionnectionReason int32

const (
	SessionDescionnectionReason_SessionsLimit SessionDescionnectionReason = 0
	SessionDescionnectionReason_SessionKilled SessionDescionnectionReason = 1
)

// Kill Client Session
func (s *HttpServer) killClientSession(cfg *oauth2.Config, reason SessionDescionnectionReason) error {
	// remove from

	// publish to the client that your session has been killed or limit reached and you will be logged out
	// TODO: Implement client disconnection with reason
	// s.publishClientDisconnectedWithReason(cfg, reason)

	// TODO: Implement broker disconnection publishing
	// s.publishClientDisconnected(cfg)

	// delete the oldest session
	err := s.OAuth2.Logout(context.Background(), cfg.AccessToken, cfg.SessionId)
	if err != nil {
		s.Log.Logger.Error(err)
		err = nil
	}

	// wait for the client to receive the message then logout anyway
	time.AfterFunc(time.Second*5, func() {
		s.Log.Logger.Infof("killClientSession: %s", cfg.SessionId)

		// logout from the websocket if there is a connection
		err = s.Hub.Delete(cfg.SessionId)
		if err != nil {
			s.Log.Logger.Error(err)
		}
	})

	return nil
}
