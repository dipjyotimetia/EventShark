package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/dipjyotimetia/event-stream/pkg/events"
	"github.com/dipjyotimetia/event-stream/pkg/router"
	"github.com/go-chi/chi/v5"
)

func main() {
	// The HTTP Server
	server := &http.Server{Addr: ":9050", Handler: service()}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Create a shutdown context with a grace period of 5 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("Error during server shutdown: %v\n", err)
		}

		serverStopCtx()
	}()

	// Run the server in a separate goroutine
	go func() {
		// Run the server and handle the error
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting server: %v\n", err)
		}
	}()

	// Wait for server context to be stopped
	<-serverCtx.Done()

	// Server has stopped, perform any cleanup or other actions
	log.Println("Server stopped gracefully")
}

func service() http.Handler {
	r := chi.NewRouter()

	// r.Use(middleware.RequestID)
	// r.Use(middleware.RealIP)
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("error loading config")
	}

	cc := events.NewKafkaClient(cfg)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Post("/expense", router.ExpenseRouter(cc, cfg))

	return r
}
