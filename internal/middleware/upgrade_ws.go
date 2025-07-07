package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// a basic middileware to check if really the client has requested his request to be upgraded
// to WebSocket protocol
func (m *Middleware) UpgradeWS(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)

		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}
