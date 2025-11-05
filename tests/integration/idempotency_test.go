//go:build integration
// +build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/dipjyotimetia/event-shark/gen"
)

func TestIdempotency(t *testing.T) {
	expense := gen.Expense{
		ExpenseID:   "idem-test-001",
		UserID:      "user-001",
		Category:    "testing",
		Amount:      50.00,
		Currency:    "USD",
		Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
		Description: nil,
		Receipt:     nil,
	}

	jsonData, err := json.Marshal(expense)
	if err != nil {
		t.Fatalf("Error marshalling json: %v", err)
	}

	idempotencyKey := "test-key-" + time.Now().Format("20060102150405")

	// First request should succeed
	req1, _ := http.NewRequest("POST", "http://localhost:8083/api/expense", bytes.NewBuffer(jsonData))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Idempotency-Key", idempotencyKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatalf("Error making first request: %v", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 on first request, got %v", resp1.StatusCode)
	}

	// Second request with same idempotency key should fail
	req2, _ := http.NewRequest("POST", "http://localhost:8083/api/expense", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Idempotency-Key", idempotencyKey)

	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("Error making second request: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusConflict {
		t.Fatalf("Expected status 409 (conflict) on duplicate request, got %v", resp2.StatusCode)
	}

	// Verify error response
	var errorResp map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&errorResp); err != nil {
		t.Fatalf("Error decoding error response: %v", err)
	}

	if _, ok := errorResp["error"]; !ok {
		t.Error("Expected error field in response")
	}
}

func TestIdempotencyStats(t *testing.T) {
	resp, err := http.Get("http://localhost:8083/api/stats")
	if err != nil {
		t.Fatalf("Error getting stats: %v", err)
	}
	defer resp.Body.Close()

	var stats map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&stats)

	idempotencyStats, ok := stats["idempotency"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected idempotency stats in response")
	}

	// Verify key metrics exist
	requiredFields := []string{"cache_size", "max_entries", "ttl_seconds"}
	for _, field := range requiredFields {
		if _, exists := idempotencyStats[field]; !exists {
			t.Errorf("Expected field %s in idempotency stats", field)
		}
	}
}

func TestWithoutIdempotencyKey(t *testing.T) {
	expense := gen.Expense{
		ExpenseID:   "no-idem-test-001",
		UserID:      "user-001",
		Category:    "testing",
		Amount:      25.00,
		Currency:    "USD",
		Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
		Description: nil,
		Receipt:     nil,
	}

	jsonData, _ := json.Marshal(expense)

	// Request without idempotency key should succeed
	req, _ := http.NewRequest("POST", "http://localhost:8083/api/expense", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %v", resp.StatusCode)
	}

	// Second identical request without idempotency key should also succeed
	req2, _ := http.NewRequest("POST", "http://localhost:8083/api/expense", bytes.NewBuffer(jsonData))
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("Error making second request: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 on second request without idempotency, got %v", resp2.StatusCode)
	}
}
