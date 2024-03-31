package router

import (
	"context"

	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/dipjyotimetia/event-stream/pkg/events"
	"github.com/dipjyotimetia/event-stream/pkg/handler"
	"github.com/gofiber/fiber/v2"
)

// ExpenseRouter is the Router for GoFiber App
func ExpenseRouter(app fiber.Router, ctx context.Context, client *events.KafkaClient, cfg *config.Config) {
	app.Post("/expense", handler.ExpenseHandler(ctx, client, cfg))
}

func PaymentRouter(app fiber.Router, ctx context.Context, client *events.KafkaClient, cfg *config.Config) {
	app.Post("/payment", handler.PaymentHandler(ctx, client, cfg))
}
