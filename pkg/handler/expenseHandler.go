// Package handler provides HTTP handler functions for handling expense-related routes.

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

// ExpenseHandler returns an HTTP handler function for creating expense records.
// It takes a KafkaClient instance and a Config instance as input.
func ExpenseHandler(ctx context.Context, client events.Producer, cfg *config.Config, val validator.Validator, log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Add request ID to context for tracing
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx = context.WithValue(ctx, "request_id", requestID)

		var expense gen.Expense

		// Parse request body
		if err := c.BodyParser(&expense); err != nil {
			log.LogError(ctx, err, "failed to parse expense request body")
			appErr := errors.ErrBadRequest(errors.MsgInvalidJSON, err)
			return c.Status(appErr.Code).JSON(appErr)
		}

		// Set the Timestamp field to current time if it's not already set
		if expense.Timestamp == 0 {
			expense.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}

		// Validate expense data
		if err := val.ValidateExpense(expense); err != nil {
			log.LogError(ctx, err, "expense validation failed", "expense_id", expense.ExpenseID)
			if appErr, ok := err.(*errors.AppError); ok {
				return c.Status(appErr.Code).JSON(appErr)
			}
			appErr := errors.ErrValidation(errors.MsgValidationFailed, err)
			return c.Status(appErr.Code).JSON(appErr)
		}

		// Create Kafka record
		record, err := client.SetRecord(cfg, expense, "expense-topic", gen.Expense{})
		if err != nil {
			log.LogError(ctx, err, "failed to create Kafka record for expense", "expense_id", expense.ExpenseID)
			appErr := errors.ErrInternalServer(errors.MsgSchemaError, err)
			return c.Status(appErr.Code).JSON(appErr)
		}

		// Produce message to Kafka
		if err := client.ProduceSync(ctx, record); err != nil {
			log.LogError(ctx, err, "failed to produce expense message to Kafka", "expense_id", expense.ExpenseID)
			appErr := errors.ErrInternalServer(errors.MsgKafkaProduceFailed, err)
			return c.Status(appErr.Code).JSON(appErr)
		}

		log.LogInfo(ctx, "expense created successfully", "expense_id", expense.ExpenseID, "user_id", expense.UserID)

		response := map[string]interface{}{
			"message":    "expense created successfully",
			"expense_id": expense.ExpenseID,
			"timestamp":  expense.Timestamp,
		}

		return c.Status(fiber.StatusCreated).JSON(response)
	}
}
