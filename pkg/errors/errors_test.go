package errors

import (
	"testing"
	"time"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(ErrCodeValidationFailed, "test error")

	if err.Code != ErrCodeValidationFailed {
		t.Errorf("Expected code %s, got %s", ErrCodeValidationFailed, err.Code)
	}

	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got %s", err.Message)
	}

	if err.Details == nil {
		t.Error("Expected details map to be initialized")
	}

	if err.Retryable {
		t.Error("Expected retryable to be false by default")
	}
}

func TestAppError_Error(t *testing.T) {
	err := NewAppError(ErrCodeKafkaError, "connection failed")
	expected := "[KAFKA_ERROR] connection failed"

	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestAppError_WithDetails(t *testing.T) {
	err := NewAppError(ErrCodeValidationFailed, "validation failed")
	err.WithDetails("field", "amount").WithDetails("reason", "must be positive")

	if err.Details["field"] != "amount" {
		t.Error("Expected field detail to be set")
	}

	if err.Details["reason"] != "must be positive" {
		t.Error("Expected reason detail to be set")
	}
}

func TestAppError_WithRetryable(t *testing.T) {
	err := NewAppError(ErrCodeKafkaError, "timeout")
	err.WithRetryable(true)

	if !err.Retryable {
		t.Error("Expected retryable to be true")
	}
}

func TestAppError_WithRetryAfter(t *testing.T) {
	err := NewAppError(ErrCodeRateLimitExceeded, "rate limit exceeded")
	err.WithRetryAfter(60)

	if err.RetryAfter == nil {
		t.Fatal("Expected retry_after to be set")
	}

	if *err.RetryAfter != 60 {
		t.Errorf("Expected retry_after to be 60, got %d", *err.RetryAfter)
	}
}

func TestAppError_GetHTTPStatus(t *testing.T) {
	tests := []struct {
		code           ErrorCode
		expectedStatus int
	}{
		{ErrCodeInvalidRequest, 400},
		{ErrCodeValidationFailed, 400},
		{ErrCodeUnauthorized, 401},
		{ErrCodeForbidden, 403},
		{ErrCodeNotFound, 404},
		{ErrCodeDuplicateRequest, 409},
		{ErrCodeRateLimitExceeded, 429},
		{ErrCodeCircuitBreakerOpen, 503},
		{ErrCodeServiceUnavailable, 503},
		{ErrCodeTimeout, 504},
		{ErrCodeInternalError, 500},
		{ErrCodeKafkaError, 500},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := NewAppError(tt.code, "test")
			status := err.GetHTTPStatus()

			if status != tt.expectedStatus {
				t.Errorf("Expected status %d for code %s, got %d", tt.expectedStatus, tt.code, status)
			}
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("invalid input")
		if err.Code != ErrCodeValidationFailed {
			t.Error("Expected VALIDATION_FAILED code")
		}
	})

	t.Run("NewSchemaValidationError", func(t *testing.T) {
		err := NewSchemaValidationError("schema mismatch")
		if err.Code != ErrCodeSchemaValidation {
			t.Error("Expected SCHEMA_VALIDATION_FAILED code")
		}
	})

	t.Run("NewKafkaError", func(t *testing.T) {
		err := NewKafkaError("connection timeout")
		if err.Code != ErrCodeKafkaError {
			t.Error("Expected KAFKA_ERROR code")
		}
		if !err.Retryable {
			t.Error("Expected Kafka errors to be retryable")
		}
	})

	t.Run("NewCircuitBreakerError", func(t *testing.T) {
		err := NewCircuitBreakerError("circuit open")
		if err.Code != ErrCodeCircuitBreakerOpen {
			t.Error("Expected CIRCUIT_BREAKER_OPEN code")
		}
		if !err.Retryable {
			t.Error("Expected circuit breaker errors to be retryable")
		}
		if err.RetryAfter == nil || *err.RetryAfter != 30 {
			t.Error("Expected retry_after to be 30 seconds")
		}
	})

	t.Run("NewDuplicateError", func(t *testing.T) {
		err := NewDuplicateError("duplicate key")
		if err.Code != ErrCodeDuplicateRequest {
			t.Error("Expected DUPLICATE_REQUEST code")
		}
	})

	t.Run("NewInternalError", func(t *testing.T) {
		err := NewInternalError("internal failure")
		if err.Code != ErrCodeInternalError {
			t.Error("Expected INTERNAL_ERROR code")
		}
	})
}

func TestAppError_Timestamp(t *testing.T) {
	before := time.Now()
	err := NewAppError(ErrCodeInternalError, "test")
	after := time.Now()

	if err.Timestamp.Before(before) || err.Timestamp.After(after) {
		t.Error("Timestamp should be set to current time")
	}
}
