// Package middleware provides custom middleware for the Event Shark application.
package middleware

import (
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/errors"
	"github.com/dipjyotimetia/event-shark/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID middleware adds a unique request ID to each request.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

// RateLimit creates a simple rate limiting middleware.
func RateLimit(maxRequests int, window time.Duration) fiber.Handler {
	// This is a simplified rate limiter - in production, use a proper rate limiter
	requestCounts := make(map[string]int)
	lastReset := time.Now()

	return func(c *fiber.Ctx) error {
		now := time.Now()

		// Reset counters if window has passed
		if now.Sub(lastReset) > window {
			requestCounts = make(map[string]int)
			lastReset = now
		}

		clientIP := c.IP()
		requestCounts[clientIP]++

		if requestCounts[clientIP] > maxRequests {
			appErr := errors.ErrValidation("rate limit exceeded", nil)
			return c.Status(fiber.StatusTooManyRequests).JSON(appErr)
		}

		return c.Next()
	}
}

// ErrorHandler provides centralized error handling.
func ErrorHandler(log *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default error response
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// Check if it's an AppError
		if appErr, ok := err.(*errors.AppError); ok {
			code = appErr.Code
			message = appErr.Message
		} else if fiberErr, ok := err.(*fiber.Error); ok {
			code = fiberErr.Code
			message = fiberErr.Message
		}

		// Log the error
		requestID := c.Locals("request_id")
		if requestID != nil {
			log.Logger.Error("request error",
				"request_id", requestID,
				"method", c.Method(),
				"path", c.Path(),
				"status", code,
				"error", err.Error(),
			)
		}

		// Return error response
		return c.Status(code).JSON(fiber.Map{
			"error":   true,
			"message": message,
		})
	}
}

// Recovery middleware recovers from panics.
func Recovery(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				requestID := c.Locals("request_id")
				log.Logger.Error("panic recovered",
					"request_id", requestID,
					"method", c.Method(),
					"path", c.Path(),
					"panic", r,
				)

				appErr := errors.ErrInternalServer("internal server error", nil)
				c.Status(appErr.Code).JSON(appErr)
			}
		}()

		return c.Next()
	}
}

// CORS middleware with custom configuration.
func CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Set("Access-Control-Expose-Headers", "X-Request-ID")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}
