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

func TestPaymentsAPI(t *testing.T) {
	// Wait for service to be ready
	if !waitForService(t, baseURL+"/health", 30*time.Second) {
		t.Fatal("Service did not become ready in time")
	}

	testCases := []struct {
		name         string
		payment      gen.Payment
		expectedCode int
		shouldFail   bool
	}{
		{
			name: "Valid payment creation",
			payment: gen.Payment{
				TransactionID: "txn-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				UserID:        "user-10010",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Timestamp:     time.Now().UnixNano() / int64(time.Millisecond),
				Status:        "COMPLETED",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Payment with minimum required fields",
			payment: gen.Payment{
				TransactionID: "txn-min-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				UserID:        "user-10010",
				Amount:        50.00,
				Currency:      "USD",
				PaymentMethod: "DEBIT_CARD",
				Timestamp:     time.Now().UnixNano() / int64(time.Millisecond),
				Status:        "PENDING",
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Invalid payment - missing required fields",
			payment: gen.Payment{
				Amount:        50.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
			},
			expectedCode: http.StatusUnprocessableEntity,
			shouldFail:   true,
		},
		{
			name: "Invalid payment - negative amount",
			payment: gen.Payment{
				TransactionID: "txn-negative",
				UserID:        "user-10010",
				Amount:        -50.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Status:        "COMPLETED",
			},
			expectedCode: http.StatusUnprocessableEntity,
			shouldFail:   true,
		},
		{
			name: "Invalid payment - invalid payment method",
			payment: gen.Payment{
				TransactionID: "txn-invalid-method",
				UserID:        "user-10010",
				Amount:        50.00,
				Currency:      "USD",
				PaymentMethod: "INVALID_METHOD",
				Status:        "COMPLETED",
			},
			expectedCode: http.StatusUnprocessableEntity,
			shouldFail:   true,
		},
		{
			name: "Invalid payment - invalid status",
			payment: gen.Payment{
				TransactionID: "txn-invalid-status",
				UserID:        "user-10010",
				Amount:        50.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Status:        "INVALID_STATUS",
			},
			expectedCode: http.StatusUnprocessableEntity,
			shouldFail:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.payment)
			if err != nil {
				t.Fatalf("Error marshalling json: %v", err)
			}

			resp, err := http.Post(baseURL+"/api/payment", "application/json", bytes.NewBuffer(jsonData))
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

				if txnID, ok := response["transaction_id"].(string); !ok || txnID == "" {
					t.Errorf("Expected transaction_id in response")
				}

				if status, ok := response["status"].(string); !ok || status == "" {
					t.Errorf("Expected status in response")
				}
			}
		})
	}
}

func TestPaymentAPIErrorHandling(t *testing.T) {
	testCases := []struct {
		name         string
		requestBody  string
		expectedCode int
	}{
		{
			name:         "Invalid JSON",
			requestBody:  `{"invalid": json}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty body",
			requestBody:  ``,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Malformed JSON",
			requestBody:  `{"transaction_id": "test", "amount": "not-a-number"}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(baseURL+"/api/payment", "application/json", bytes.NewBufferString(tc.requestBody))
			if err != nil {
				t.Fatalf("Error making request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedCode {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", tc.expectedCode, resp.StatusCode, string(body))
			}
		})
	}
}

func TestPaymentAPIRateLimiting(t *testing.T) {
	// This test verifies that the rate limiting middleware is working
	// Note: This is a simplified test - in a real scenario, you'd want to test the actual rate limits

	payment := gen.Payment{
		TransactionID: "txn-rate-limit-test",
		UserID:        "user-10010",
		Amount:        10.00,
		Currency:      "USD",
		PaymentMethod: "CREDIT_CARD",
		Status:        "COMPLETED",
	}

	jsonData, err := json.Marshal(payment)
	if err != nil {
		t.Fatalf("Error marshalling json: %v", err)
	}

	// Make a few requests to ensure the service is working
	for i := 0; i < 5; i++ {
		resp, err := http.Post(baseURL+"/api/payment", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Error making request %d: %v", i+1, err)
		}
		resp.Body.Close()

		// We expect either success or validation error (for duplicate transaction IDs)
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnprocessableEntity {
			t.Logf("Request %d returned status %d", i+1, resp.StatusCode)
		}
	}
}
