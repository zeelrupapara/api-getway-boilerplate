package middleware

import (
	"strings"
	"greenlync-api-gateway/pkg/errors"
	"greenlync-api-gateway/utils"

	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) Authorization(resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r := strings.Split(resource, "_")

		client, ok := utils.GetClient(c)
		if !ok {
			return m.App.HttpResponseInternalServerErrorRequest(c, errors.ErrCouldNotParseClientCfg)
		}
		ok = m.Authz.Enforcer.HasNamedPolicy("p", client.Scope, r[0], r[1])
		if !ok {
			return m.App.HttpResponseForbidden(c, errors.ErrUnauthorizedToAccessResource)
		}
		return c.Next()
	}
}
