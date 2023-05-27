// Package router provides an HTTP handler function for handling expense-related routes.

package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	exp "github.com/dipjyotimetia/event-stream/gen/expense"
	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/dipjyotimetia/event-stream/pkg/events"
)

// ExpenseRouter returns an HTTP handler function for creating expense records.
// It takes a KafkaClient instance and a Config instance as input.
// The handler function parses the JSON request body into an Expense struct,
// sets the Timestamp field to the current time if it's not already set,
// creates a Kafka record with the expense data, and sends it to the Kafka topic.
// Finally, it returns a success response.
func ExpenseRouter(client events.KafkaClient, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON request body into an Expense struct
		var expense exp.Expense
		err := json.NewDecoder(r.Body).Decode(&expense)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Set the Timestamp field to current time if it's not already set
		if expense.Timestamp == 0 {
			expense.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}

		// Create a Kafka record with the expense data
		record := client.SetExpenseRecord(cfg, expense)

		// Send the Kafka record to the Kafka topic
		client.Producer(context.Background(), record)

		// Return a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("expense created successfully"))
	}
}
