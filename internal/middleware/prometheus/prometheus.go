package prometheus

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
)

/*

these are by default metrics in this middleware 

http_requests_total
http_request_duration_seconds
http_requests_in_progress_total

*/


func NewPrometheus(app *fiber.App) *fiberprometheus.FiberPrometheus {

	prometheus := fiberprometheus.New("greenlync-api-gateway")	

	prometheus.RegisterAt(app, "/metrics")

	return prometheus
}
