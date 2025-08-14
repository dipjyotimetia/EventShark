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

func PaymentHandler(ctx context.Context, client *events.KafkaClient, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set request timeout
		reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var payment gen.Payment

		if err := c.BodyParser(&payment); err != nil {
			return c.Status(http.StatusBadRequest).JSON(ErrorResponse{
				Error:   "invalid_request",
				Message: "Failed to parse request body",
			})
		}

		// Set the Timestamp field to current time if it's not already set
		if payment.Timestamp == 0 {
			payment.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}

		record, err := client.SetRecord(cfg, payment, "payment-topic", gen.Payment{})
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
			Message: "Payment created successfully",
		})
	}
}
