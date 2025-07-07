package middleware

import (
	"strings"
	"greenlync-api-gateway/pkg/errors"
	"greenlync-api-gateway/utils"

	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) Protect(c *fiber.Ctx) error {
	// in case the access token is not in the header, check if it's in the query
	if accessToken := c.Query("access_token"); accessToken != "" {
		if _, ok := utils.GetToken(c); !ok {
			return m.App.HttpResponseUnauthorized(c, errors.ErrInvalidToken)
		}
		return c.Next()
	}

	// try Get the Authorization header from the request
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return m.App.HttpResponseUnauthorized(c, errors.ErrMissingAuthoirzationHeader)
	}

	// Check if the header is not empty and starts with "Bearer "
	if strings.HasPrefix(authHeader, "Bearer ") {
		if _, ok := utils.GetToken(c); !ok {
			return m.App.HttpResponseUnauthorized(c, errors.ErrInvalidToken)
		}

		return c.Next()
	}

	return m.App.HttpResponseUnauthorized(c, errors.ErrInvalidBearerToken)
}
