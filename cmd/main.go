package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/dipjyotimetia/event-stream/pkg/events"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// The HTTP Server
	server := &http.Server{Addr: "0.0.0.0:8080", Handler: service()}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

func service() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	cfg, err := config.NewConfig()
	if err != nil {
		_ = fmt.Errorf("error loading config")
	}

	cc := events.NewKafkaClient(cfg)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// tt := gen.Expense{
	// 	Expense_id:  "ioweiofh",
	// 	User_id:     "fwef",
	// 	Category:    "New",
	// 	Amount:      23.34,
	// 	Currency:    "AUD",
	// 	Timestamp:   time.Now().UTC().UnixMilli(),
	// 	Description: nil,
	// 	Receipt:     nil,
	// }

	r.Get("/expense", func(w http.ResponseWriter, r *http.Request) {
		// record := cc.SetExpenseRecord(cfg, buf)
		record := cc.SetExpenseRecordAvro(cfg)
		cc.Producer(context.Background(), record)
		w.Write([]byte(fmt.Sprintf("all done.\n")))
	})

	return r
}
