//go:build integration
// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/dipjyotimetia/event-stream/gen"
)

func TestExpenseAPI(t *testing.T) {
	testCases := []struct {
		name     string
		expense  gen.Expense
		expected int
	}{
		{
			name: "Test Expense API",
			expense: gen.Expense{
				ExpenseID:   "test",
				UserID:      "10010",
				Category:    "kafka",
				Amount:      20.5,
				Currency:    "AUD",
				Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
				Description: nil,
				Receipt:     nil,
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.expense)
			if err != nil {
				t.Fatalf("Error marshalling json: %v", err)
			}

			resp, err := http.Post("http://localhost:8083/api/expense", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Error making request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expected {
				t.Fatalf("Expected status %v, got %v", tc.expected, resp.StatusCode)
			}
		})
	}
}
