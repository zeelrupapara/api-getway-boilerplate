// Developer: Saif Hamdan

package middleware

import (
	"greenlync-api-gateway/pkg/authz"
	"greenlync-api-gateway/pkg/http"
	"greenlync-api-gateway/pkg/logger"
	"greenlync-api-gateway/pkg/nats"
	"greenlync-api-gateway/pkg/oauth2"

	"gorm.io/gorm"
)

type Middleware struct {
	// Fiber App
	App *http.App
	// MySql DB
	DB *gorm.DB
	// Authorization
	Authz *authz.Authz
	// Authentucation
	OAuth2 *oauth2.OAuth2
	// Nats
	Nats *nats.Nats
	// zab logger for log to files and stdout
	Log *logger.Logger
}

func NewMiddleware(app *http.App, db *gorm.DB, authz *authz.Authz, oauth2 *oauth2.OAuth2, log *logger.Logger, nats *nats.Nats) *Middleware {

	m := &Middleware{
		App:    app,
		DB:     db,
		Authz:  authz,
		OAuth2: oauth2,
		Log:    log,
		Nats:   nats,
	}

	return m
}
