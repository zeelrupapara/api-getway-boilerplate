package server

import (
	// "encoding/json"

	"github.com/go-co-op/gocron"
	"github.com/go-playground/validator/v10"

	//"github.com/gofiber/fiber/v2"
	//"github.com/gofiber/swagger"
	"gorm.io/gorm"

	// import local pkg
	"greenlync-api-gateway/config"
	"greenlync-api-gateway/internal/middleware"
	v1 "greenlync-api-gateway/internal/server/v1"
	"greenlync-api-gateway/pkg/authz"
	"greenlync-api-gateway/pkg/cache"
	"greenlync-api-gateway/pkg/http"
	"greenlync-api-gateway/pkg/i18n"
	"greenlync-api-gateway/pkg/logger"
	"greenlync-api-gateway/pkg/manager"
	"greenlync-api-gateway/pkg/nats"
	"greenlync-api-gateway/pkg/oauth2"
	"greenlync-api-gateway/pkg/smtp"
	// Removed influxdb, news, and support imports
	// fiber local middleware
	//"greenlync-api-gateway/pkg/fiber/middleware/jaeger"
)

// Emran  17.3.2023
// Now server is a generic builder with http server as v1 every http, ws or any public interface should
// move to v1 including routes,controlers
// also model/v1 the same
// all resources
// remove any resources here that was in v1 and duplicated
type Server struct {
	App *http.App
	// Middleware
	Middleware *middleware.Middleware
	// Our OAuth API
	// OAuth *oauth.HttpOAuth
	// Our v1 Http API
	Web *v1.HttpServer

	// we can have v2 Fiber app , each also can run in diffrent server if we want in the future
	// so we serve v1 & v2 customers in the same time
	//Web *v2.HttpServer

	// Transalte
	Local *i18n.Lang
	// zab logger for log to files and stdout
	Log *logger.Logger
	// Cache for Redis Caching
	Cache *cache.Cache
	// DB
	DB *gorm.DB
	// Authorization
	Authz *authz.Authz
	// OAuth2.0
	OAuth2 *oauth2.OAuth2
	// Nats
	Nats *nats.Nats
	// Websocket manager
	Hub *manager.Hub
	// SMTP Client
	Smtp *smtp.SMTP
	// Go-Cron
	Cron *gocron.Scheduler
}

func NewServer(local *i18n.Lang, log *logger.Logger, cache *cache.Cache, db *gorm.DB, authz *authz.Authz, nats *nats.Nats, validate *validator.Validate, cfg *config.Config, smtp *smtp.SMTP, cron *gocron.Scheduler) *Server {
	// fiber instence
	app := http.NewApp(log)

	// Websocket Hub
	newHub := manager.NewHub(log)
	manager.SetMaxWebsocketConnections()

	// OAuth2
	oauth2 := oauth2.NewOAuth2(cache, db, cfg, log)

	// Removed InfluxDB, News, and Support for minimal boilerplate

	// middleware
	middleware := middleware.NewMiddleware(app, db, authz, oauth2, log, nats)
	// v1 HTTP
	newHttp := v1.NewHTTP(app, db, log, cache, nats, authz, oauth2, newHub, middleware, smtp, cfg, validate, cron)

	// start Monitoring Sessions Activity
	go oauth2.MonitoryActivity()

	server := &Server{
		App:           app,
		Middleware:    middleware,
		Web:           newHttp,
		DB:            db,
		Local:         local,
		Log:           log,
		Cache:         cache,
		Nats:          nats,
		Hub:           newHub,
		Authz:         authz,
		OAuth2:        oauth2,
		Smtp:          smtp,
		Cron:          cron,
	}

	return server
}
