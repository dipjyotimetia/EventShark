package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/dipjyotimetia/event-shark/gen"
	"github.com/dipjyotimetia/event-shark/pkg/config"
	appErrors "github.com/dipjyotimetia/event-shark/pkg/errors"
	"github.com/dipjyotimetia/event-shark/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Mock implementations for testing
type MockProducer struct {
	shouldFailSetRecord bool
	shouldFailProduce   bool
	records             []*kgo.Record
}

func (m *MockProducer) ProduceSync(ctx context.Context, record *kgo.Record) error {
	if m.shouldFailProduce {
		return errors.New("producer error")
	}
	m.records = append(m.records, record)
	return nil
}

func (m *MockProducer) ProduceAsync(ctx context.Context, record *kgo.Record) error {
	return m.ProduceSync(ctx, record)
}

func (m *MockProducer) SetRecord(cfg *config.Config, data interface{}, topic string, schemaType interface{}) (*kgo.Record, error) {
	if m.shouldFailSetRecord {
		return nil, errors.New("set record error")
	}
	return &kgo.Record{
		Value:     []byte("test-value"),
		Topic:     topic,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockProducer) Close() error {
	return nil
}

type MockValidator struct {
	shouldFailValidation bool
}

func (m *MockValidator) ValidateExpense(expense gen.Expense) error {
	if m.shouldFailValidation {
		return appErrors.ErrValidation("validation failed", nil)
	}
	return nil
}

func (m *MockValidator) ValidatePayment(payment gen.Payment) error {
	if m.shouldFailValidation {
		return appErrors.ErrValidation("validation failed", nil)
	}
	return nil
}

func TestExpenseHandler(t *testing.T) {
	tests := []struct {
		name                 string
		requestBody          interface{}
		shouldFailSetRecord  bool
		shouldFailProduce    bool
		shouldFailValidation bool
		expectedStatus       int
	}{
		{
			name: "successful expense creation",
			requestBody: gen.Expense{
				ExpenseID: "exp-001",
				UserID:    "user-001",
				Category:  "food",
				Amount:    25.99,
				Currency:  "USD",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation failure",
			requestBody: gen.Expense{
				ExpenseID: "exp-001",
				UserID:    "user-001",
				Category:  "food",
				Amount:    25.99,
				Currency:  "USD",
			},
			shouldFailValidation: true,
			expectedStatus:       http.StatusUnprocessableEntity,
		},
		{
			name: "set record failure",
			requestBody: gen.Expense{
				ExpenseID: "exp-001",
				UserID:    "user-001",
				Category:  "food",
				Amount:    25.99,
				Currency:  "USD",
			},
			shouldFailSetRecord: true,
			expectedStatus:      http.StatusInternalServerError,
		},
		{
			name: "producer failure",
			requestBody: gen.Expense{
				ExpenseID: "exp-001",
				UserID:    "user-001",
				Category:  "food",
				Amount:    25.99,
				Currency:  "USD",
			},
			shouldFailProduce: true,
			expectedStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := fiber.New()

			mockProducer := &MockProducer{
				shouldFailSetRecord: tt.shouldFailSetRecord,
				shouldFailProduce:   tt.shouldFailProduce,
			}

			mockValidator := &MockValidator{
				shouldFailValidation: tt.shouldFailValidation,
			}

			cfg := &config.Config{
				Brokers: "localhost:9092",
				Topics:  []string{"expense-topic"},
			}

			log := logger.New()
			ctx := context.Background()

			handler := ExpenseHandler(ctx, mockProducer, cfg, mockValidator, log)
			app.Post("/expense", handler)

			// Prepare request
			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			// Make request
			req, _ := http.NewRequest("POST", "/expense", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Check status
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestPaymentHandler(t *testing.T) {
	tests := []struct {
		name                 string
		requestBody          interface{}
		shouldFailSetRecord  bool
		shouldFailProduce    bool
		shouldFailValidation bool
		expectedStatus       int
	}{
		{
			name: "successful payment creation",
			requestBody: gen.Payment{
				TransactionID: "txn-001",
				UserID:        "user-001",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Status:        "COMPLETED",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation failure",
			requestBody: gen.Payment{
				TransactionID: "txn-001",
				UserID:        "user-001",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Status:        "COMPLETED",
			},
			shouldFailValidation: true,
			expectedStatus:       http.StatusUnprocessableEntity,
		},
		{
			name: "set record failure",
			requestBody: gen.Payment{
				TransactionID: "txn-001",
				UserID:        "user-001",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Status:        "COMPLETED",
			},
			shouldFailSetRecord: true,
			expectedStatus:      http.StatusInternalServerError,
		},
		{
			name: "producer failure",
			requestBody: gen.Payment{
				TransactionID: "txn-001",
				UserID:        "user-001",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Status:        "COMPLETED",
			},
			shouldFailProduce: true,
			expectedStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := fiber.New()

			mockProducer := &MockProducer{
				shouldFailSetRecord: tt.shouldFailSetRecord,
				shouldFailProduce:   tt.shouldFailProduce,
			}

			mockValidator := &MockValidator{
				shouldFailValidation: tt.shouldFailValidation,
			}

			cfg := &config.Config{
				Brokers: "localhost:9092",
				Topics:  []string{"payment-topic"},
			}

			log := logger.New()
			ctx := context.Background()

			handler := PaymentHandler(ctx, mockProducer, cfg, mockValidator, log)
			app.Post("/payment", handler)

			// Prepare request
			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			// Make request
			req, _ := http.NewRequest("POST", "/payment", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Check status
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d but got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
