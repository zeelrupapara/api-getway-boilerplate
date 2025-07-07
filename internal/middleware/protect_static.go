package middleware

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) ProtectStatic(c *fiber.Ctx) error {
	accessToken := c.Cookies("accessToken")

	if accessToken != "" {
		// if it exists then it's valid, otherwise it's not
		ctx := context.Background()
		cfg, err := m.OAuth2.Inspect(ctx, accessToken)
		if err == redis.Nil {
			return c.Redirect("/login", 302)
		}

		// store client's session in the requst
		if cfg != nil {
			c.Locals("client", cfg)
			c.Locals("token", accessToken)
		} else {
			return c.Redirect("/login", 302)
		}
	} else {
		return c.Redirect("/login", 302)
	}

	return c.Next()
}
