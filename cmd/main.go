package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/app"
)

// main is the entry point for the GreenLync API Gateway
// Following VFX Server pattern with graceful shutdown
func main() {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize application
	application, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start application
	if err := application.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	// Log startup completion
	application.Logger.Info("GreenLync API Gateway started successfully",
		"port", application.Config.Server.Port,
		"cannabis_compliance", application.Config.Cannabis.ComplianceMode,
		"age_verification_required", application.Config.Cannabis.AgeVerificationRequired,
	)

	// Wait for shutdown signal
	sig := <-quit
	application.Logger.Info("Received shutdown signal",
		"signal", sig.String(),
	)

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Graceful shutdown
	application.Logger.Info("Starting graceful shutdown...")
	
	if err := application.Shutdown(shutdownCtx); err != nil {
		application.Logger.Error("Error during shutdown",
			"error", err,
		)
		os.Exit(1)
	}

	application.Logger.Info("GreenLync API Gateway stopped gracefully")
}