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

		record := client.SetExpenseRecord(cfg, expense)
		client.Producer(context.Background(), record)

		// Return a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("expense created successfully"))
	}
}
