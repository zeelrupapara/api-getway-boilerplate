package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/config"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/db"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/model/common/v1"
)

// Application represents the main application instance
// Following VFX Server pattern with dependency injection
type Application struct {
	Config *config.Config
	Logger *logger.Logger
	Fiber  *fiber.App
	
	// Services will be added as we implement them
	DB     *db.Database
	// Cache  *cache.Cache
	// NATS   *nats.Client
	// Hub    *manager.Hub
	// OAuth2 *oauth2.Service
	// Authz  *authz.Service
}

// NewApplication creates a new application instance with dependency injection
func NewApplication() (*Application, error) {
	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logConfig := &logger.LogConfig{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		Output:      cfg.Logging.Output,
		Structured:  cfg.Logging.Structured,
		FileOutput:  cfg.Logging.FileOutput,
		MaxFileSize: cfg.Logging.MaxFileSize,
		MaxBackups:  cfg.Logging.MaxBackups,
		MaxAge:      cfg.Logging.MaxAge,
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Create Fiber app with cannabis-specific configuration
	fiberApp := fiber.New(fiber.Config{
		ServerHeader:      "GreenLync-Gateway/1.0",
		AppName:          "GreenLync Cannabis Social Commerce API Gateway",
		ReadTimeout:      cfg.Server.ReadTimeout,
		WriteTimeout:     cfg.Server.WriteTimeout,
		IdleTimeout:      cfg.Server.IdleTimeout,
		EnablePrintRoutes: cfg.Server.Debug,
		ErrorHandler:     createErrorHandler(log),
		BodyLimit:        50 * 1024 * 1024, // 50MB for file uploads (ID verification documents)
	})

	app := &Application{
		Config: cfg,
		Logger: log,
		Fiber:  fiberApp,
	}

	// Log application initialization
	log.Info("Application initialized successfully",
		"service", "greenlync-api-gateway",
		"cannabis_compliance_mode", cfg.Cannabis.ComplianceMode,
		"age_verification_required", cfg.Cannabis.AgeVerificationRequired,
		"legal_states_count", len(cfg.Cannabis.LegalStates),
	)

	return app, nil
}

// Start starts the application services
func (a *Application) Start() error {
	a.Logger.Info("Starting GreenLync API Gateway services...")

	// Initialize dependencies (will be implemented in next phases)
	if err := a.initializeDependencies(); err != nil {
		return fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Setup middleware
	if err := a.setupMiddleware(); err != nil {
		return fmt.Errorf("failed to setup middleware: %w", err)
	}

	// Setup routes
	if err := a.setupRoutes(); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}

	// Start HTTP server
	go func() {
		addr := a.Config.Server.GetServerAddress()
		a.Logger.Info("Starting HTTP server",
			"address", addr,
			"cannabis_platform", true,
		)
		
		if err := a.Fiber.Listen(addr); err != nil {
			a.Logger.Fatal("Failed to start HTTP server",
				"error", err,
			)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown(ctx context.Context) error {
	a.Logger.Info("Shutting down application...")

	// Shutdown Fiber server
	if err := a.Fiber.ShutdownWithContext(ctx); err != nil {
		a.Logger.Error("Error shutting down HTTP server",
			"error", err,
		)
		return err
	}

	// Shutdown other services (will be implemented in next phases)
	if err := a.shutdownDependencies(ctx); err != nil {
		a.Logger.Error("Error shutting down dependencies",
			"error", err,
		)
		return err
	}

	// Sync logger
	if err := a.Logger.Sync(); err != nil {
		return err
	}

	return nil
}

// initializeDependencies initializes all application dependencies
func (a *Application) initializeDependencies() error {
	a.Logger.Info("Initializing application dependencies...")

	// Initialize database connection
	database, err := db.NewDatabase(&a.Config.Database, a.Logger)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.DB = database

	// Run database migrations
	if err := a.DB.AutoMigrate(v1.GetAllModels()...); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Setup multi-tenancy (dispensary isolation)
	if err := a.DB.SetupMultiTenancy(); err != nil {
		a.Logger.Warn("Failed to setup multi-tenancy", "error", err)
	}

	// TODO: Initialize Redis cache
	// TODO: Initialize NATS messaging
	// TODO: Initialize WebSocket hub
	// TODO: Initialize OAuth2 service
	// TODO: Initialize authorization service

	a.Logger.Info("Dependencies initialization completed")
	return nil
}

// setupMiddleware sets up the middleware chain
func (a *Application) setupMiddleware() error {
	a.Logger.Info("Setting up middleware chain...")

	// Add global middleware (basic setup for now)
	// TODO: Add comprehensive middleware in Phase 5

	// Health check endpoint
	a.Fiber.Get("/health", a.healthCheckHandler)

	a.Logger.Info("Middleware setup completed")
	return nil
}

// setupRoutes sets up all application routes
func (a *Application) setupRoutes() error {
	a.Logger.Info("Setting up application routes...")

	// API version prefix
	api := a.Fiber.Group("/api/v1")

	// Cannabis compliance notice endpoint
	api.Get("/compliance", a.complianceInfoHandler)

	// TODO: Add authentication routes in Phase 7
	// TODO: Add user management routes in Phase 7
	// TODO: Add RBAC routes in Phase 7
	// TODO: Add WebSocket routes in Phase 4

	a.Logger.Info("Routes setup completed")
	return nil
}

// shutdownDependencies gracefully shuts down all dependencies
func (a *Application) shutdownDependencies(ctx context.Context) error {
	a.Logger.Info("Shutting down dependencies...")

	// Close database connections
	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			a.Logger.Error("Failed to close database connection", "error", err)
			return err
		}
	}

	// TODO: Close Redis connections
	// TODO: Close NATS connections
	// TODO: Close WebSocket hub

	a.Logger.Info("Dependencies shutdown completed")
	return nil
}

// healthCheckHandler provides health check endpoint
func (a *Application) healthCheckHandler(c *fiber.Ctx) error {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "greenlync-api-gateway",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"cannabis_compliance": map[string]interface{}{
			"age_verification_required": a.Config.Cannabis.AgeVerificationRequired,
			"minimum_age":              a.Config.Cannabis.MinimumAge,
			"compliance_mode":          a.Config.Cannabis.ComplianceMode,
			"legal_states_count":       len(a.Config.Cannabis.LegalStates),
			"audit_logging_enabled":    a.Config.Cannabis.AuditLogging,
		},
		"dependencies": a.getDependencyHealth(),
	}

	return c.JSON(health)
}

// complianceInfoHandler provides cannabis compliance information
func (a *Application) complianceInfoHandler(c *fiber.Ctx) error {
	compliance := map[string]interface{}{
		"platform_type":            "cannabis_social_commerce",
		"age_requirement":          a.Config.Cannabis.MinimumAge,
		"age_verification_required": a.Config.Cannabis.AgeVerificationRequired,
		"compliance_mode":          a.Config.Cannabis.ComplianceMode,
		"legal_states":             a.Config.Cannabis.LegalStates,
		"state_check_enabled":      a.Config.Cannabis.StateCheckEnabled,
		"purchase_limit_tracking":  a.Config.Cannabis.PurchaseLimitTracking,
		"audit_logging":            a.Config.Cannabis.AuditLogging,
		"notice": "This platform is restricted to users 21+ in states where cannabis is legal. Age verification and state compliance checks are required.",
	}

	return c.JSON(compliance)
}

// getDependencyHealth returns the health status of all dependencies
func (a *Application) getDependencyHealth() map[string]interface{} {
	health := map[string]interface{}{
		"redis": "not_initialized",
		"nats":  "not_initialized",
	}

	// Check database health
	if a.DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := a.DB.Health(ctx); err != nil {
			health["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			health["database"] = map[string]interface{}{
				"status": "healthy",
				"stats":  a.DB.GetStats(),
			}
		}
	} else {
		health["database"] = "not_initialized"
	}

	return health
}

// createErrorHandler creates a custom error handler for cannabis compliance
func createErrorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError

		// Type assertion for Fiber errors
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		// Log error with cannabis compliance context
		log.Error("API Gateway error",
			"error", err.Error(),
			"status_code", code,
			"path", c.Path(),
			"method", c.Method(),
			"ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
			"cannabis_platform", true,
		)

		// Return error response with cannabis compliance notice
		return c.Status(code).JSON(fiber.Map{
			"error": fiber.Map{
				"message":   err.Error(),
				"code":      code,
				"timestamp": time.Now().UTC(),
				"path":      c.Path(),
			},
			"cannabis_notice": "This platform requires age verification (21+) and compliance with local cannabis laws",
		})
	}
}