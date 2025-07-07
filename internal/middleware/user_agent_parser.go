package middleware

import (
	"greenlync-api-gateway/pkg/http"
	"greenlync-api-gateway/utils"

	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) UserAgentParser(c *fiber.Ctx) error {
	s := c.Get("User-Agent")

	ua := utils.UserAgentParser(s)

	c.Locals(http.LocalsOs, ua.OS)
	c.Locals(http.LocalsDevice, ua.Device)
	c.Locals(http.LocalsChannel, ua.Channel)

	return c.Next()
}
