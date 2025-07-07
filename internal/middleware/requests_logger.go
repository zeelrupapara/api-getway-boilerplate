// Developer: Saif Hamdan

package middleware

import (
	"fmt"
	"time"
	"greenlync-api-gateway/pkg/http"
	"greenlync-api-gateway/utils"

	"github.com/gofiber/fiber/v2"
)

// this Middleware is responsible for logging incoming HTTP Requests
func (m *Middleware) RequestsLogger(c *fiber.Ctx) error {
	startTime := time.Now()
	//startTime := time.Now()

	var sessionId string = "-"
	var clientId int32 = 0
	cfg, ok := utils.GetClient(c)
	if ok {
		sessionId = cfg.SessionId
		clientId = cfg.ClientId
		// new activity
		m.OAuth2.NewActivity(sessionId)
	}

	// Next() proceeds to the next middleware or route handler
	err := c.Next()

	// Calculate the time taken
	elapsed := time.Since(startTime)

	// IP accountId sessionId [time] "method path code duration" "userAgent"
	msg := fmt.Sprintf("%s %d %s [%s] \"%s %s %d %dms\" %s",
		utils.GetRealIP(c), clientId, sessionId, time.Now().UTC().Format(time.Layout), c.Method(), c.Path(), c.Response().StatusCode(), elapsed.Milliseconds(), c.Locals(http.LocalsDevice))
	m.Log.Logger.Info(msg)

	// TODO: later on we can also log this to loki

	return err
}
