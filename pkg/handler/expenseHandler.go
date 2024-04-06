// Package router provides an HTTP handler function for handling expense-related routes.

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

// ExpenseHandler returns an HTTP handler function for creating expense records.
// It takes a KafkaClient instance and a Config instance as input.
func ExpenseHandler(ctx context.Context, client *events.KafkaClient, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var expense gen.Expense

		if err := c.BodyParser(&expense); err != nil {
			c.Status(http.StatusBadRequest)
			return err
		}

		// Set the Timestamp field to current time if it's not already set
		if expense.Timestamp == 0 {
			expense.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}

		record, err := client.SetRecord(cfg, expense, "expense-topic", gen.Expense{})
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}

		err = client.Producer(ctx, record)
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}

		c.SendStatus(http.StatusOK) //nolint:errcheck
		return c.Send([]byte("expense created successfully"))
	}
}
