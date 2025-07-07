// Developer: Saif Hamdan

package oauth2

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
	"greenlync-api-gateway/config"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/cache"
	"greenlync-api-gateway/pkg/logger"

	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"

	"gorm.io/gorm"
)

type Config struct {
	Id string
	// Client Id or accountId
	ClientId int32
	// Client Secret Id
	ClientSecretId string
	// ip address
	IpAddress string
	// Email
	Email string
	// Access token
	AccessToken string
	// Refresh token
	RefreshToken string
	// Access token
	SessionId string
	// Client's role based on Casbin
	Scope string
	// sessions Expiration date
	ExpiresIn int
	// CreatedAt
	StartedAt time.Time
	// Last Activity
	LastActivity time.Time
	// User Agent
	UserAgent string
	// WS, if true this means he is connected on websocket and he is active
	Ws bool
	// Remember me
	RememberMe bool
}

type OAuth2 struct {
	// Store tokens in Cache
	Cache *cache.Cache
	// DB gorm
	DB *gorm.DB
	// Logs
	Log *logger.Logger
	// password credentials expireation data from .env
	TokenExpiresIn int
	// Long Token Expiration
	LongTokenExpiresIn int
	// this is a list of the active sessions
	SessionsList ActiveSessionsList
	// protect sessionMap
	sync.RWMutex
}

func NewOAuth2(cache *cache.Cache, db *gorm.DB, cfg *config.Config, log *logger.Logger) *OAuth2 {
	return &OAuth2{
		Cache:              cache,
		DB:                 db,
		Log:                log,
		SessionsList:       make(ActiveSessionsList),
		TokenExpiresIn:     cfg.HTTP.OAuthTokenExpiresIn,
		LongTokenExpiresIn: cfg.HTTP.OAuthLongTokenExpiresIn,
	}
}

// func (o *OAuth2) ClientCredentialsToken(ctx context.Context, expiresIn int64, config *Config) (*model.Token, error) {
// 	tokenstr := o.GenerateToken()
// 	token := &model.Token{
// 		UserId:   config.ClientId,
// 		Scope:       config.Scope,
// 		AccessToken: tokenstr,
// 		ExpiresIn:   o.TokenExpiresIn,
// 	}

// 	js, err := json.Marshal(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = o.Cache.Set(ctx, token.AccessToken, js, 0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = o.DB.Create(token).Error

// 	if err != nil {
// 		o.DeleteToken(ctx, token.AccessToken)
// 		return nil, err
// 	}

// 	return token, nil
// }

func (o *OAuth2) PasswordCredentialsToken(ctx context.Context, config *Config) (*Config, error) {
	config.AccessToken = o.GenerateToken()
	config.RefreshToken = o.GenerateToken()
	config.SessionId = o.GenerateToken()
	config.StartedAt = time.Now()
	config.LastActivity = time.Now()

	token := &model.Token{
		AccessToken:  config.AccessToken,
		RefreshToken: config.RefreshToken,
		SessionId:    config.SessionId,
		ExpiresIn:    o.TokenExpiresIn,
		Scope:        config.Scope,
		IpAddress:    config.IpAddress,
		UserId:       config.ClientId,
	}
	session := &model.Session{
		SessionId: config.SessionId,
		Scope:     config.Scope,
		IpAddress: config.IpAddress,
		UserAgent: config.UserAgent,
		UserId: config.ClientId,
	}

	tx := o.DB.Begin()
	err := tx.Create(token).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Create(session).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	config.Id = session.Id

	js, err := json.Marshal(config)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// ctx.Deadline()
	err = o.Cache.Set(ctx, config.AccessToken, js, o.TokenExpiresIn)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = o.Cache.Set(ctx, cache.SessionsKey(config.SessionId), js, o.TokenExpiresIn)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	o.NewActiveSession(config)

	tx.Commit()

	return config, nil
}

func (o *OAuth2) Inspect(ctx context.Context, accessToken string) (*Config, error) {
	js, err := o.Cache.Get(ctx, accessToken)
	if err != nil {

		return nil, err
	}

	config := &Config{}
	err = json.Unmarshal([]byte(js), config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (o *OAuth2) RefreshToken(ctx context.Context, ipAddr, useragent, refreshToken string) (*Config, error) {
	// find the session

	token := &model.Token{}
	config := &Config{}

	err := o.DB.Where("refresh_token = ?", refreshToken).First(token).Error
	if err != nil {
		return nil, err
	}

	if token.RefreshToken != refreshToken {
		return nil, fmt.Errorf("invalid refresh token")
	}

	config = &Config{
		AccessToken:  o.GenerateToken(),
		RefreshToken: o.GenerateToken(),
		SessionId:    token.SessionId,
		ExpiresIn:    o.TokenExpiresIn,
		Scope:        token.Scope,
		IpAddress:    ipAddr,
		ClientId:     token.UserId,
		LastActivity: time.Now(),
		StartedAt:    time.Now(),
		UserAgent:    useragent,
	}

	// get the old access token of exist in cache
	js, err := o.Cache.Get(ctx, token.AccessToken)
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
	}

	if js != "" {
		// if access token exist in the cache delete from the cache
		err := o.Cache.Delete(ctx, token.AccessToken)
		if err != nil {
			return nil, err
		}
	}

	ss, err := o.Cache.Get(ctx, cache.SessionsKey(token.SessionId))
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
	}

	if ss != "" {
		err := o.Cache.Delete(ctx, cache.SessionsKey(token.SessionId))
		if err != nil {
			return nil, err
		}
	}

	tx := o.DB.Begin()
	err = tx.Save(token).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	newToken := &model.Token{
		AccessToken:  config.AccessToken,
		RefreshToken: config.RefreshToken,
		SessionId:    token.SessionId,
		ExpiresIn:    o.TokenExpiresIn,
		Scope:        config.Scope,
		IpAddress:    token.IpAddress,
		UserId:    token.UserId,
	}

	err = tx.Save(newToken).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	err = o.Cache.Set(ctx, config.AccessToken, b, o.TokenExpiresIn)
	if err != nil {
		return nil, err
	}

	err = o.Cache.Set(ctx, cache.SessionsKey(config.SessionId), b, o.TokenExpiresIn)
	if err != nil {
		return nil, err
	}

	o.NewActiveSession(config)

	tx.Commit()

	return config, nil
}

// returns true(valid) if the token exists otherwise returns false(unvalid)
func (o *OAuth2) VerifyToken(ctx context.Context, accessToken string) (bool, error) {
	_, err := o.Cache.Get(ctx, accessToken)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// deletes token if exists
func (o *OAuth2) DeleteToken(ctx context.Context, accessToken string) error {
	if ok, err := o.VerifyToken(ctx, accessToken); !ok {
		return err
	} else if err != nil {
		return err
	}
	err := o.Cache.Delete(ctx, accessToken)
	if err != nil {
		return err
	}

	return nil
}

// deletes token if exists
func (o *OAuth2) DeleteSessionId(ctx context.Context, sessionId string) error {
	o.Lock()
	defer o.Unlock()

	if ok, err := o.VerifySessionId(ctx, sessionId); !ok {
		return err
	} else if err != nil {
		return err
	}

	if err := o.Cache.Delete(ctx, cache.SessionsKey(sessionId)); err != nil {
		return err
	}

	return nil
}

// Verify SessionId
func (o *OAuth2) VerifySessionId(ctx context.Context, sessionId string) (bool, error) {
	_, err := o.Cache.Get(ctx, cache.SessionsKey(sessionId))
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Logout
func (o *OAuth2) Logout(ctx context.Context, accessToken string, sessionId string) error {
	session := &model.Session{}
	tx := o.DB.Begin()
	err := tx.Where("session_id = ?", sessionId).First(session).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	time := time.Now()
	session.FinishedAt = &time

	err = tx.Save(session).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = o.DeleteToken(ctx, accessToken)
	if err != nil {
		if err != redis.Nil {
			tx.Rollback()
			return err
		}
	}

	err = o.DeleteSessionId(ctx, sessionId)
	if err != nil {
		if err != redis.Nil {
			tx.Rollback()
			return err
		}
	}

	o.DeleteActiveSession(sessionId)

	tx.Commit()

	return nil
}

func (o *OAuth2) LogoutAll() {
	for _, v := range o.SessionsList {
		o.Logout(context.Background(), v.AccessToken, v.SessionId)
	}
}

// Generate SHA256 hash, then encode the hash using hexadecimal encoding
// accord to OAuth2.0 specification
func (o *OAuth2) GenerateToken() string {
	// Generate a random byte array
	randomBytes := randStringBytes()

	// Hash the message using SHA256
	hash := sha256.Sum256(randomBytes)

	// Encode the hash using hexadecimal encoding
	encodedToken := hex.EncodeToString(hash[:])

	return encodedToken
}

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes() []byte {
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}
