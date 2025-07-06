//go:build integration
// +build integration

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/dipjyotimetia/event-shark/gen"
)

const baseURL = "http://localhost:8083"

func TestExpenseAPI(t *testing.T) {
	// Wait for service to be ready
	if !waitForService(t, baseURL+"/health", 30*time.Second) {
		t.Fatal("Service did not become ready in time")
	}

	testCases := []struct {
		name         string
		expense      gen.Expense
		expectedCode int
		shouldFail   bool
	}{
		{
			name: "Valid expense creation",
			expense: gen.Expense{
				ExpenseID:   "exp-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				UserID:      "user-10010",
				Category:    "food",
				Amount:      25.99,
				Currency:    "USD",
				Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
				Description: stringPtr("Integration test expense"),
				Receipt:     stringPtr("receipt-url"),
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Expense with minimum required fields",
			expense: gen.Expense{
				ExpenseID: "exp-min-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				UserID:    "user-10010",
				Category:  "transport",
				Amount:    15.50,
				Currency:  "USD",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Invalid expense - missing required fields",
			expense: gen.Expense{
				Amount:   10.00,
				Currency: "USD",
			},
			expectedCode: http.StatusUnprocessableEntity,
			shouldFail:   true,
		},
		{
			name: "Invalid expense - negative amount",
			expense: gen.Expense{
				ExpenseID: "exp-negative",
				UserID:    "user-10010",
				Category:  "food",
				Amount:    -25.99,
				Currency:  "USD",
			},
			expectedCode: http.StatusUnprocessableEntity,
			shouldFail:   true,
		},
		{
			name: "Invalid expense - invalid currency",
			expense: gen.Expense{
				ExpenseID: "exp-invalid-currency",
				UserID:    "user-10010",
				Category:  "food",
				Amount:    25.99,
				Currency:  "INVALID",
			},
			expectedCode: http.StatusUnprocessableEntity,
			shouldFail:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.expense)
			if err != nil {
				t.Fatalf("Error marshalling json: %v", err)
			}

			resp, err := http.Post(baseURL+"/api/expense", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Error making request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedCode {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", tc.expectedCode, resp.StatusCode, string(body))
			}

			// Verify response structure for successful requests
			if !tc.shouldFail && resp.StatusCode == http.StatusCreated {
				var response map[string]interface{}
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %v", err)
				}

				if err := json.Unmarshal(body, &response); err != nil {
					t.Errorf("Error unmarshalling response: %v", err)
				}

				if message, ok := response["message"].(string); !ok || message == "" {
					t.Errorf("Expected message in response")
				}

				if expenseID, ok := response["expense_id"].(string); !ok || expenseID == "" {
					t.Errorf("Expected expense_id in response")
				}
			}
		})
	}
}

func TestExpenseAPIHealth(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Error making health check request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected health check status 200, got %d", resp.StatusCode)
	}

	var healthResponse map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading health response: %v", err)
	}

	if err := json.Unmarshal(body, &healthResponse); err != nil {
		t.Fatalf("Error unmarshalling health response: %v", err)
	}

	if status, ok := healthResponse["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected healthy status, got %v", status)
	}
}

func TestExpenseAPIReadiness(t *testing.T) {
	resp, err := http.Get(baseURL + "/health/ready")
	if err != nil {
		t.Fatalf("Error making readiness check request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected readiness check status 200, got %d", resp.StatusCode)
	}

	var readinessResponse map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading readiness response: %v", err)
	}

	if err := json.Unmarshal(body, &readinessResponse); err != nil {
		t.Fatalf("Error unmarshalling readiness response: %v", err)
	}

	if ready, ok := readinessResponse["ready"].(bool); !ok || !ready {
		t.Errorf("Expected ready status true, got %v", ready)
	}
}

// Helper functions
func waitForService(t *testing.T, url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func stringPtr(s string) *string {
	return &s
}
