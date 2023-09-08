// Package router provides an HTTP handler function for handling expense-related routes.

package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/dipjyotimetia/event-stream/gen"
	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/dipjyotimetia/event-stream/pkg/events"
	"github.com/gofiber/fiber/v2"
)

// ExpenseHandler returns an HTTP handler function for creating expense records.
// It takes a KafkaClient instance and a Config instance as input.
// The handler function parses the JSON request body into an Expense struct,
// sets the Timestamp field to the current time if it's not already set,
// creates a Kafka record with the expense data, and sends it to the Kafka topic.
// Finally, it returns a success response.
func ExpenseHandler(client *events.KafkaClient, cfg *config.Config) fiber.Handler {
	// Parse the JSON request body into an Expense struct
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

		// Create a Kafka record with the expense data
		record := client.SetExpenseRecord(cfg, expense)

		// Send the Kafka record to the Kafka topic
		err := client.Producer(context.Background(), record)
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		// Return a success response
		c.SendStatus(http.StatusOK) //nolint:errcheck
		return c.Send([]byte("expense created successfully"))
	}
}
