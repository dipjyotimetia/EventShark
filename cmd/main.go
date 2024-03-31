package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/dipjyotimetia/event-stream/pkg/events"
	"github.com/dipjyotimetia/event-stream/pkg/router"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

const idleTimeout = 5 * time.Second

func main() {
	app := fiber.New(fiber.Config{
		IdleTimeout: idleTimeout,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})
	app.Use(cors.New())
	app.Use(helmet.New())
	app.Use(logger.New(logger.Config{
		Format:     "${cyan}[${time}] ${white}${pid} ${red}${status} ${blue}[${method}] ${white}${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "UTC",
	}))
	app.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.Send([]byte("healthy"))
	})

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("error loading config")
	}

	ctx := context.Background()
	cc := events.NewKafkaClient(cfg)
	api := app.Group("/api")
	router.ExpenseRouter(api, ctx, cc, cfg)
	router.PaymentRouter(api, ctx, cc, cfg)
	// Listen from a different goroutine
	go func() {
		if err := app.Listen(":8083"); err != nil {
			log.Panic(err)
		}
	}()
	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel

	<-c // This blocks the main thread until an interrupt is received
	log.Println("Gracefully shutting down...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Running cleanup tasks...")
	cc.Close()
	log.Println("Fiber was successful shutdown.")
}
