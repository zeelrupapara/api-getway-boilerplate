// Developer: zeelrupapara@gmail.com
// Description: Health check endpoints for GreenLync boilerplate

package v1

import (
	"context"
	"greenlync-api-gateway/pkg/monitor"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *HttpServer) CheckSystemHealth(c *fiber.Ctx) error {
	return s.App.HttpResponseOK(c, monitor.GetHealthStatus())
}

func (s *HttpServer) CheckReadiness(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthStatus := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().UTC(),
		"checks":    make(map[string]interface{}),
	}

	checks := healthStatus["checks"].(map[string]interface{})
	allHealthy := true

	// Check database connection
	if sqlDB, err := s.DB.DB(); err == nil {
		if err := sqlDB.PingContext(ctx); err == nil {
			checks["database"] = map[string]interface{}{
				"status": "healthy",
				"type":   "mysql",
			}
		} else {
			checks["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
				"type":   "mysql",
			}
			allHealthy = false
		}
	} else {
		checks["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
			"type":   "mysql",
		}
		allHealthy = false
	}

	// Check Redis connection
	if s.Cache != nil {
		if err := s.Cache.Ping(ctx); err == nil {
			checks["redis"] = map[string]interface{}{
				"status": "healthy",
				"type":   "cache",
			}
		} else {
			checks["redis"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
				"type":   "cache",
			}
			allHealthy = false
		}
	}

	// Check NATS connection
	if s.Nats != nil && s.Nats.NC != nil {
		if s.Nats.NC.IsConnected() {
			checks["nats"] = map[string]interface{}{
				"status": "healthy",
				"type":   "messaging",
			}
		} else {
			checks["nats"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  "not connected",
				"type":   "messaging",
			}
			allHealthy = false
		}
	}

	if !allHealthy {
		healthStatus["status"] = "not_ready"
		return c.Status(503).JSON(healthStatus)
	}

	return s.App.HttpResponseOK(c, healthStatus)
}

func (s *HttpServer) CheckLiveness(c *fiber.Ctx) error {
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(s.StartTime).String(),
		"version":   "1.0.0", // This should come from build info
	}

	return s.App.HttpResponseOK(c, response)
}
