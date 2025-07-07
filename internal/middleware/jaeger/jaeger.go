package jaeger

import (
	fibertracing "github.com/aschenmaker/fiber-opentracing"
	"github.com/aschenmaker/fiber-opentracing/fjaeger"
	"github.com/gofiber/fiber/v2"
	"github.com/opentracing/opentracing-go"
)

func NewJaeger(operation string) func(*fiber.Ctx) {
	var ctx *fiber.Ctx
	
	opr := operation + ":  HTTP " + ctx.Method() + " URL: " + ctx.Path()
	
	// defualt
	fjaeger.New(fjaeger.Config{})

	return func(c *fiber.Ctx) {
		fibertracing.New(fibertracing.Config{
			Tracer: opentracing.GlobalTracer(),
			OperationName: func(ctx *fiber.Ctx) string {
				return opr
			},
		})
	}
	
}
