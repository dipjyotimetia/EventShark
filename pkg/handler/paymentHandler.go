// Package handler provides HTTP handler functions for handling payment-related routes.

package handler

import (
	"context"
	"time"

	"github.com/dipjyotimetia/event-shark/gen"
	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/errors"
	"github.com/dipjyotimetia/event-shark/pkg/events"
	"github.com/dipjyotimetia/event-shark/pkg/logger"
	"github.com/dipjyotimetia/event-shark/pkg/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// PaymentHandler returns an HTTP handler function for creating payment records.
// It takes a KafkaClient instance and a Config instance as input.
func PaymentHandler(ctx context.Context, client events.Producer, cfg *config.Config, val validator.Validator, log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add request ID to context for tracing
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx = context.WithValue(ctx, "request_id", requestID)

		var payment gen.Payment

		// Parse request body
		if err := c.BodyParser(&payment); err != nil {
			log.LogError(ctx, err, "failed to parse payment request body")
			appErr := errors.ErrBadRequest(errors.MsgInvalidJSON, err)
			return c.Status(appErr.Code).JSON(appErr)
		}

		// Set the Timestamp field to current time if it's not already set
		if payment.Timestamp == 0 {
			payment.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}

		// Validate payment data
		if err := val.ValidatePayment(payment); err != nil {
			log.LogError(ctx, err, "payment validation failed", "transaction_id", payment.TransactionID)
			if appErr, ok := err.(*errors.AppError); ok {
				return c.Status(appErr.Code).JSON(appErr)
			}
			appErr := errors.ErrValidation(errors.MsgValidationFailed, err)
			return c.Status(appErr.Code).JSON(appErr)
		}

		// Create Kafka record
		record, err := client.SetRecord(cfg, payment, "payment-topic", gen.Payment{})
		if err != nil {
			log.LogError(ctx, err, "failed to create Kafka record for payment", "transaction_id", payment.TransactionID)
			appErr := errors.ErrInternalServer(errors.MsgSchemaError, err)
			return c.Status(appErr.Code).JSON(appErr)
		}

		// Produce message to Kafka
		if err := client.ProduceSync(ctx, record); err != nil {
			log.LogError(ctx, err, "failed to produce payment message to Kafka", "transaction_id", payment.TransactionID)
			appErr := errors.ErrInternalServer(errors.MsgKafkaProduceFailed, err)
			return c.Status(appErr.Code).JSON(appErr)
		}

		log.LogInfo(ctx, "payment created successfully", "transaction_id", payment.TransactionID, "user_id", payment.UserID)

		response := map[string]interface{}{
			"message":        "payment created successfully",
			"transaction_id": payment.TransactionID,
			"timestamp":      payment.Timestamp,
			"status":         payment.Status,
		}

		return c.Status(fiber.StatusCreated).JSON(response)
	}
}
