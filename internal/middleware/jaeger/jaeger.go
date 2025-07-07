package jaeger

import (
	"fmt"
	"io"
	"log"
	"os"

	fibertracing "github.com/aschenmaker/fiber-opentracing"
	"github.com/aschenmaker/fiber-opentracing/fjaeger"
	"github.com/gofiber/fiber/v2"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

type JaegerTracer struct {
	tracer opentracing.Tracer
	closer io.Closer
}

func NewJaegerTracer(serviceName string) (*JaegerTracer, error) {
	cfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
			LocalAgentHostPort: fmt.Sprintf("%s:%s", 
				getEnvOr("JAEGER_AGENT_HOST", "localhost"),
				getEnvOr("JAEGER_AGENT_PORT", "6831")),
		},
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger tracer: %w", err)
	}

	opentracing.SetGlobalTracer(tracer)
	log.Printf("Jaeger tracer initialized for service: %s", serviceName)

	return &JaegerTracer{
		tracer: tracer,
		closer: closer,
	}, nil
}

func (j *JaegerTracer) Close() error {
	if j.closer != nil {
		return j.closer.Close()
	}
	return nil
}

func NewJaegerMiddleware(serviceName string) fiber.Handler {
	jaegerTracer, err := NewJaegerTracer(serviceName)
	if err != nil {
		log.Printf("Failed to initialize Jaeger tracer: %v", err)
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	fjaegerConfig := fjaeger.Config{
		ServiceName: serviceName,
	}

	fjaeger.New(fjaegerConfig)

	return fibertracing.New(fibertracing.Config{
		Tracer: jaegerTracer.tracer,
		OperationName: func(ctx *fiber.Ctx) string {
			return fmt.Sprintf("%s %s", ctx.Method(), ctx.Path())
		},
		Filter: func(ctx *fiber.Ctx) bool {
			return ctx.Path() != "/metrics" && ctx.Path() != "/health"
		},
		Modify: func(ctx *fiber.Ctx, span opentracing.Span) {
			span.SetTag("http.method", ctx.Method())
			span.SetTag("http.url", ctx.OriginalURL())
			span.SetTag("http.status_code", ctx.Response().StatusCode())
			span.SetTag("user_agent", ctx.Get("User-Agent"))
			span.SetTag("remote_addr", ctx.IP())
			
			if userID := ctx.Locals("user_id"); userID != nil {
				span.SetTag("user.id", userID)
			}
		},
	})
}

func getEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
