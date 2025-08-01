// Package router provides routing functionality for the Event Shark application.

package router

import (
	"context"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/events"
	"github.com/dipjyotimetia/event-shark/pkg/handler"
	"github.com/dipjyotimetia/event-shark/pkg/logger"
	"github.com/dipjyotimetia/event-shark/pkg/validator"
	"github.com/gofiber/fiber/v2"
)

// ExpenseRouter sets up the expense-related routes.
func ExpenseRouter(app fiber.Router, ctx context.Context, client events.Producer, cfg *config.Config, val validator.Validator, log *logger.Logger) {
	app.Post("/expense", handler.ExpenseHandler(ctx, client, cfg, val, log))
}

// PaymentRouter sets up the payment-related routes.
func PaymentRouter(app fiber.Router, ctx context.Context, client events.Producer, cfg *config.Config, val validator.Validator, log *logger.Logger) {
	app.Post("/payment", handler.PaymentHandler(ctx, client, cfg, val, log))
}
