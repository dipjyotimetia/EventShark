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

func TestAsyncPublishing(t *testing.T) {
	expense := gen.Expense{
		ExpenseID:   "async-test-001",
		UserID:      "user-001",
		Category:    "testing",
		Amount:      99.99,
		Currency:    "USD",
		Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
		Description: nil,
		Receipt:     nil,
	}

	jsonData, err := json.Marshal(expense)
	if err != nil {
		t.Fatalf("Error marshalling json: %v", err)
	}

	// Create request with async header
	req, err := http.NewRequest("POST", "http://localhost:8083/api/expense", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Async", "true")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Expected status %v, got %v", http.StatusAccepted, resp.StatusCode)
	}

	// Parse response to get job ID
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	jobID, ok := result["job_id"].(string)
	if !ok || jobID == "" {
		t.Fatal("Expected job_id in response")
	}

	// Wait a bit for processing
	time.Sleep(2 * time.Second)

	// Check job status
	statusResp, err := http.Get("http://localhost:8083/api/jobs/" + jobID)
	if err != nil {
		t.Fatalf("Error checking job status: %v", err)
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 for job status, got %v", statusResp.StatusCode)
	}

	var jobStatus map[string]interface{}
	if err := json.NewDecoder(statusResp.Body).Decode(&jobStatus); err != nil {
		t.Fatalf("Error decoding job status: %v", err)
	}

	status, ok := jobStatus["status"].(string)
	if !ok {
		t.Fatal("Expected status in job response")
	}

	// Should be COMPLETED or PROCESSING
	if status != "COMPLETED" && status != "PROCESSING" {
		t.Errorf("Expected job status COMPLETED or PROCESSING, got %v", status)
	}
}

func TestAsyncStats(t *testing.T) {
	resp, err := http.Get("http://localhost:8083/api/stats")
	if err != nil {
		t.Fatalf("Error getting stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %v", resp.StatusCode)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Error decoding stats: %v", err)
	}

	// Check async stats exist
	asyncStats, ok := stats["async"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected async stats in response")
	}

	// Verify key metrics exist
	requiredFields := []string{"total_jobs", "queue_length", "queue_capacity", "workers"}
	for _, field := range requiredFields {
		if _, exists := asyncStats[field]; !exists {
			t.Errorf("Expected field %s in async stats", field)
		}
	}
}
