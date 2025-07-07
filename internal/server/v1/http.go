// Developer: zeelrupapara@gmail.com
// Description: HTTP server for GreenLync boilerplate

package v1

import (
	"greenlync-api-gateway/config"
	"greenlync-api-gateway/internal/middleware"
	model "greenlync-api-gateway/model/common/v1"
	"greenlync-api-gateway/pkg/authz"
	"greenlync-api-gateway/pkg/cache"
	"greenlync-api-gateway/pkg/http"
	"greenlync-api-gateway/pkg/logger"
	"greenlync-api-gateway/pkg/manager"
	"greenlync-api-gateway/pkg/nats"
	"greenlync-api-gateway/pkg/oauth2"
	"greenlync-api-gateway/pkg/smtp"

	"github.com/go-co-op/gocron"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type HttpServer struct {
	// Config
	Cfg *config.Config
	// Middleware
	Middleware *middleware.Middleware
	// Fiber app
	App *http.App
	// MySQl DB
	DB *gorm.DB
	// Cache
	Cache *cache.Cache
	// Authorization
	Authz *authz.Authz
	// OAuth2.0
	OAuth2 *oauth2.OAuth2
	// Nats
	Nats *nats.Nats
	// zab logger for log to files and stdout
	Log *logger.Logger
	// Websocket manager
	Hub *manager.Hub
	// SMTP Client
	Smtp *smtp.SMTP
	// Validetor move to http server !!! This is not needed as server
	Validate *validator.Validate
	// Go-Cron
	Cron *gocron.Scheduler
	// operations channel
	operationCh chan *model.OperationsLog
}

func NewHTTP(app *http.App, db *gorm.DB, log *logger.Logger, cache *cache.Cache, nats *nats.Nats, authz *authz.Authz, oauth *oauth2.OAuth2, hub *manager.Hub, middleware *middleware.Middleware, smtp *smtp.SMTP, cfg *config.Config, validate *validator.Validate, cron *gocron.Scheduler) *HttpServer {

	h := &HttpServer{
		Middleware:      middleware,
		App:             app,
		DB:              db,
		Log:             log,
		Cache:           cache,
		Nats:            nats,
		Authz:           authz,
		OAuth2:          oauth,
		Hub:             hub,
		Smtp:            smtp,
		Validate:        validate,
		Cron:            cron,
		Cfg:             cfg,
		operationCh:     make(chan *model.OperationsLog),
	}

	hub.SetErrorHandler(h.WSErrorHandler)
	// TODO: Implement session deletion callback for event-driven architecture
	// oauth.SetOnSessionDelete(h.publishClientDisconnected)

	// Removed NATS system router (trading-specific)
	go h.writeSystemOperationsLogs()

	return h
}
