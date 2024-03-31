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

func PaymentHandler(ctx context.Context, client *events.KafkaClient, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payment gen.Payment

		if err := c.BodyParser(&payment); err != nil {
			c.Status(http.StatusBadRequest)
			return err
		}

		// Set the Timestamp field to current time if it's not already set
		if payment.Timestamp == 0 {
			payment.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}

		record, err := client.SetRecord(cfg, payment, "payment-topic", gen.Payment{})
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
