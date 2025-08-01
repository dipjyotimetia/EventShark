// Package health provides health check functionality for the Event Shark application.
package health

import (
	"context"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/events"
	"github.com/gofiber/fiber/v2"
)

// HealthStatus represents the health status of the application.
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]Health `json:"checks"`
}

// Health represents the health of a component.
type Health struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
}

// HealthChecker provides health check functionality.
type HealthChecker struct {
	kafkaClient events.Producer
	config      *config.Config
	version     string
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(kafkaClient events.Producer, cfg *config.Config, version string) *HealthChecker {
	return &HealthChecker{
		kafkaClient: kafkaClient,
		config:      cfg,
		version:     version,
	}
}

// Handler returns the health check handler.
func (h *HealthChecker) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status := h.CheckHealth(ctx)

		statusCode := fiber.StatusOK
		if status.Status != "healthy" {
			statusCode = fiber.StatusServiceUnavailable
		}

		c.Set("Content-Type", "application/json")
		return c.Status(statusCode).JSON(status)
	}
}

// CheckHealth performs health checks on all components.
func (h *HealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	checks := make(map[string]Health)

	// Check Kafka connectivity
	checks["kafka"] = h.checkKafka(ctx)

	// Check configuration
	checks["config"] = h.checkConfig()

	// Determine overall status
	overallStatus := "healthy"
	for _, check := range checks {
		if check.Status != "healthy" {
			overallStatus = "unhealthy"
			break
		}
	}

	return HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   h.version,
		Checks:    checks,
	}
}

// checkKafka checks Kafka connectivity.
func (h *HealthChecker) checkKafka(ctx context.Context) Health {
	start := time.Now()

	// For health checks, we don't actually need to create records
	// Just verify that the Kafka client is available
	if h.kafkaClient == nil {
		return Health{
			Status:  "unhealthy",
			Message: "kafka client is not available",
			Latency: time.Since(start),
		}
	}

	// Simple connectivity check - if we have topics configured, that's a good sign
	if len(h.config.Topics) == 0 {
		return Health{
			Status:  "unhealthy",
			Message: "no topics configured",
			Latency: time.Since(start),
		}
	}

	return Health{
		Status:  "healthy",
		Latency: time.Since(start),
	}
}

// checkConfig validates configuration.
func (h *HealthChecker) checkConfig() Health {
	if h.config == nil {
		return Health{
			Status:  "unhealthy",
			Message: "configuration is nil",
		}
	}

	if err := h.config.Validate(); err != nil {
		return Health{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return Health{
		Status: "healthy",
	}
}

// ReadinessHandler returns the readiness check handler.
func (h *HealthChecker) ReadinessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Quick readiness check - just verify basic components are available
		ready := true
		checks := make(map[string]string)

		// Check if Kafka client is available
		if h.kafkaClient == nil {
			ready = false
			checks["kafka"] = "not available"
		} else {
			checks["kafka"] = "available"
		}

		// Check if config is available
		if h.config == nil {
			ready = false
			checks["config"] = "not available"
		} else {
			checks["config"] = "available"
		}

		response := map[string]interface{}{
			"ready":  ready,
			"checks": checks,
		}

		statusCode := fiber.StatusOK
		if !ready {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(response)
	}
}

// LivenessHandler returns the liveness check handler.
func (h *HealthChecker) LivenessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Simple liveness check - just return OK if the service is running
		return c.JSON(map[string]interface{}{
			"alive":     true,
			"timestamp": time.Now(),
		})
	}
}
