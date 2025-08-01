package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/events"
	"github.com/dipjyotimetia/event-shark/pkg/health"
	"github.com/dipjyotimetia/event-shark/pkg/logger"
	"github.com/dipjyotimetia/event-shark/pkg/middleware"
	"github.com/dipjyotimetia/event-shark/pkg/router"
	"github.com/dipjyotimetia/event-shark/pkg/validator"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	version = "1.0.0"
)

func main() {
	// Initialize logger
	appLogger := logger.New()

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Kafka client
	kafkaClient, err := events.NewKafkaClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	// Initialize validator
	val := validator.NewValidator()

	// Initialize health checker
	healthChecker := health.NewHealthChecker(kafkaClient, cfg, version)

	// Create Fiber app with custom configuration
	app := fiber.New(fiber.Config{
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ErrorHandler: middleware.ErrorHandler(appLogger),
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(middleware.Recovery(appLogger))
	app.Use(middleware.RequestID())
	app.Use(middleware.CORS())
	app.Use(middleware.RateLimit(1000, time.Minute)) // 1000 requests per minute

	// Health check endpoints
	app.Get("/health", healthChecker.Handler())
	app.Get("/health/ready", healthChecker.ReadinessHandler())
	app.Get("/health/live", healthChecker.LivenessHandler())

	// API routes
	ctx := context.Background()
	api := app.Group("/api")
	router.ExpenseRouter(api, ctx, kafkaClient, cfg, val, appLogger)
	router.PaymentRouter(api, ctx, kafkaClient, cfg, val, appLogger)

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + cfg.ServerPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	appLogger.LogInfo(context.Background(), "Event Shark server started",
		"version", version,
		"port", cfg.ServerPort,
		"environment", cfg.Environment,
	)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	appLogger.LogInfo(context.Background(), "Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		appLogger.LogError(context.Background(), err, "Server forced to shutdown")
	}

	appLogger.LogInfo(context.Background(), "Server gracefully stopped")
}
