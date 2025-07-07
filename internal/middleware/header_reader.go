// Developer: Saif Hamdan

package middleware

import (
	"context"
	"fmt"
	"strings"
	"greenlync-api-gateway/pkg/errors"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) HeaderReader(c *fiber.Ctx) error {
	token := ""

	// try get token from query
	if accessToken := c.Query("access_token"); accessToken != "" {
		token = accessToken
	}

	// try Get the Authorization header from the request
	authHeader := c.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.Split(authHeader, " ")[1]
	}

	if token != "" {
		// if it exists then it's valid, otherwise it's not
		ctx := context.Background()
		cfg, err := m.OAuth2.Inspect(ctx, token)
		if err == redis.Nil {
			return m.App.HttpResponseUnauthorized(c, fmt.Errorf("expired or invalid token"))
		}

		// store client's session in the requst
		if cfg != nil {
			c.Locals("client", cfg)
			c.Locals("token", token)
		} else {
			return m.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
		}
	}

	return c.Next()
}
