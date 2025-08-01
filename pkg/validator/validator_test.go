package validator

import (
	"testing"
	"time"

	"github.com/dipjyotimetia/event-shark/gen"
	"github.com/dipjyotimetia/event-shark/pkg/errors"
)

func TestAppValidator_ValidateExpense(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		expense     gen.Expense
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid expense",
			expense: gen.Expense{
				ExpenseID:   "exp-001",
				UserID:      "user-001",
				Category:    "food",
				Amount:      25.99,
				Currency:    "USD",
				Timestamp:   time.Now().UnixNano() / int64(time.Millisecond),
				Description: Ptr("Lunch at restaurant"),
			},
			expectError: false,
		},
		{
			name: "missing expense ID",
			expense: gen.Expense{
				UserID:    "user-001",
				Category:  "food",
				Amount:    25.99,
				Currency:  "USD",
				Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			},
			expectError: true,
			errorMsg:    "expense_id is required",
		},
		{
			name: "invalid category",
			expense: gen.Expense{
				ExpenseID: "exp-001",
				UserID:    "user-001",
				Category:  "invalid",
				Amount:    25.99,
				Currency:  "USD",
				Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			},
			expectError: true,
			errorMsg:    "category must be one of",
		},
		{
			name: "negative amount",
			expense: gen.Expense{
				ExpenseID: "exp-001",
				UserID:    "user-001",
				Category:  "food",
				Amount:    -25.99,
				Currency:  "USD",
				Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			},
			expectError: true,
			errorMsg:    "amount must be greater than 0",
		},
		{
			name: "invalid currency",
			expense: gen.Expense{
				ExpenseID: "exp-001",
				UserID:    "user-001",
				Category:  "food",
				Amount:    25.99,
				Currency:  "XYZ",
				Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			},
			expectError: true,
			errorMsg:    "invalid currency code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateExpense(tt.expense)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				if appErr, ok := err.(*errors.AppError); ok {
					if !containsString(appErr.Message, tt.errorMsg) {
						t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, appErr.Message)
					}
				} else {
					t.Errorf("expected AppError but got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestAppValidator_ValidatePayment(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		payment     gen.Payment
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid payment",
			payment: gen.Payment{
				TransactionID: "txn-001",
				UserID:        "user-001",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Timestamp:     time.Now().UnixNano() / int64(time.Millisecond),
				Status:        "COMPLETED",
			},
			expectError: false,
		},
		{
			name: "missing transaction ID",
			payment: gen.Payment{
				UserID:        "user-001",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Timestamp:     time.Now().UnixNano() / int64(time.Millisecond),
				Status:        "COMPLETED",
			},
			expectError: true,
			errorMsg:    "transaction_id is required",
		},
		{
			name: "invalid payment method",
			payment: gen.Payment{
				TransactionID: "txn-001",
				UserID:        "user-001",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "INVALID_METHOD",
				Timestamp:     time.Now().UnixNano() / int64(time.Millisecond),
				Status:        "COMPLETED",
			},
			expectError: true,
			errorMsg:    "payment_method must be one of",
		},
		{
			name: "invalid status",
			payment: gen.Payment{
				TransactionID: "txn-001",
				UserID:        "user-001",
				Amount:        100.00,
				Currency:      "USD",
				PaymentMethod: "CREDIT_CARD",
				Timestamp:     time.Now().UnixNano() / int64(time.Millisecond),
				Status:        "INVALID_STATUS",
			},
			expectError: true,
			errorMsg:    "status must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePayment(tt.payment)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				if appErr, ok := err.(*errors.AppError); ok {
					if !containsString(appErr.Message, tt.errorMsg) {
						t.Errorf("expected error message to contain '%s', got '%s'", tt.errorMsg, appErr.Message)
					}
				} else {
					t.Errorf("expected AppError but got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{"valid alphanumeric ID", "abc123", false},
		{"valid ID with hyphens", "abc-123", false},
		{"valid ID with underscores", "abc_123", false},
		{"empty ID", "", true},
		{"ID with spaces", "abc 123", true},
		{"ID with special chars", "abc@123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

// Helper functions
func Ptr[T any](val T) *T {
	return &val
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(substr) > 0 && len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
