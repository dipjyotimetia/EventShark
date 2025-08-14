// Package handler provides an HTTP handler function for handling expense-related routes.

package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/dipjyotimetia/event-shark/gen"
	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/events"
	"github.com/gofiber/fiber/v2"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

// ExpenseHandler returns an HTTP handler function for creating expense records.
// It takes a KafkaClient instance and a Config instance as input.
func ExpenseHandler(ctx context.Context, client *events.KafkaClient, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set request timeout
		reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var expense gen.Expense

		if err := c.BodyParser(&expense); err != nil {
			return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
				Error:   "invalid_request",
				Message: "Failed to parse request body",
			})
		}

		// Set the Timestamp field to current time if it's not already set
		if expense.Timestamp == 0 {
			expense.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}

		record, err := client.SetRecord(cfg, expense, "expense-topic", gen.Expense{})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
				Error:   "schema_error",
				Message: "Failed to encode message",
			})
		}

		err = client.Producer(reqCtx, record)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
				Error:   "kafka_error",
				Message: "Failed to produce message",
			})
		}

		return c.Status(http.StatusOK).JSON(SuccessResponse{
			Message: "Expense created successfully",
		})
	}
}
