// Developer: zeelrupapara@gmail.com
// Description: Health check endpoints for GreenLync boilerplate

package v1

import (
	"greenlync-api-gateway/pkg/monitor"

	"github.com/gofiber/fiber/v2"
)

func (s *HttpServer) CheckSystemHealth(c *fiber.Ctx) error {
	return s.App.HttpResponseOK(c, monitor.GetHealthStatus())
}
