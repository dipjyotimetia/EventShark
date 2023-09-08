package router

import (
	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/dipjyotimetia/event-stream/pkg/events"
	"github.com/dipjyotimetia/event-stream/pkg/handler"
	"github.com/gofiber/fiber/v2"
)

// ExpenseRouter is the Router for GoFiber App
func ExpenseRouter(app fiber.Router, client *events.KafkaClient, cfg *config.Config) {
	app.Post("/expense", handler.ExpenseHandler(client, cfg))
}
