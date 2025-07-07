package utils

import (
	"greenlync-api-gateway/pkg/manager"
	"greenlync-api-gateway/pkg/oauth2"

	"github.com/gofiber/fiber/v2"
)

func GetClient(c *fiber.Ctx) (*oauth2.Config, bool) {
	v := c.Locals("client")
	if v == nil {
		return nil, false
	}
	return v.(*oauth2.Config), true
}

func GetToken(c *fiber.Ctx) (string, bool) {
	v := c.Locals("token")
	if v == nil {
		return "", false
	}
	return v.(string), true
}

func GetUserAgent(c *fiber.Ctx) string {
	usg := c.Get("User-Agent")
	// ua := useragent.Parse(usg)
	// ua.Version
	return usg
}

func GetWSUserAgent(c *manager.Ctx) string {
	return c.Client.Conn.Headers("User-Agent")
}

// Get real IP address of the client X-Forwarded-For
func GetRealIP(c *fiber.Ctx) string {
	ip := "unknown"
	if c.Get("X-Forwarded-For") != "" {
		ip = c.Get("X-Forwarded-For")
	} else if c.IP() != "" {
		ip = c.IP()
	}
	return ip
}
