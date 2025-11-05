// Package errors provides standardized error handling for EventShark
package errors

import (
	"fmt"
	"time"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeInvalidRequest       ErrorCode = "INVALID_REQUEST"
	ErrCodeValidationFailed     ErrorCode = "VALIDATION_FAILED"
	ErrCodeSchemaValidation     ErrorCode = "SCHEMA_VALIDATION_FAILED"
	ErrCodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden            ErrorCode = "FORBIDDEN"
	ErrCodeNotFound             ErrorCode = "NOT_FOUND"
	ErrCodeRateLimitExceeded    ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeDuplicateRequest     ErrorCode = "DUPLICATE_REQUEST"
	ErrCodeUnsupportedFormat    ErrorCode = "UNSUPPORTED_FORMAT"

	// Server errors (5xx)
	ErrCodeInternalError        ErrorCode = "INTERNAL_ERROR"
	ErrCodeKafkaError           ErrorCode = "KAFKA_ERROR"
	ErrCodeSchemaRegistryError  ErrorCode = "SCHEMA_REGISTRY_ERROR"
	ErrCodeSerializationError   ErrorCode = "SERIALIZATION_ERROR"
	ErrCodeCircuitBreakerOpen   ErrorCode = "CIRCUIT_BREAKER_OPEN"
	ErrCodeServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout              ErrorCode = "TIMEOUT"
)

// AppError represents a standardized application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Retryable  bool                   `json:"retryable"`
	RetryAfter *int                   `json:"retry_after_seconds,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ErrorResponse represents the HTTP error response structure
type ErrorResponse struct {
	Error *AppError `json:"error"`
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC(),
		Details:   make(map[string]interface{}),
		Retryable: false,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithRetryable marks the error as retryable
func (e *AppError) WithRetryable(retryable bool) *AppError {
	e.Retryable = retryable
	return e
}

// WithRetryAfter sets the retry-after duration in seconds
func (e *AppError) WithRetryAfter(seconds int) *AppError {
	e.RetryAfter = &seconds
	return e
}

// GetHTTPStatus returns the appropriate HTTP status code for the error
func (e *AppError) GetHTTPStatus() int {
	switch e.Code {
	case ErrCodeInvalidRequest, ErrCodeValidationFailed, ErrCodeSchemaValidation, ErrCodeUnsupportedFormat:
		return 400
	case ErrCodeUnauthorized:
		return 401
	case ErrCodeForbidden:
		return 403
	case ErrCodeNotFound:
		return 404
	case ErrCodeDuplicateRequest:
		return 409
	case ErrCodeRateLimitExceeded:
		return 429
	case ErrCodeCircuitBreakerOpen, ErrCodeServiceUnavailable:
		return 503
	case ErrCodeTimeout:
		return 504
	default:
		return 500
	}
}

// Common error constructors

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidationFailed, message)
}

// NewSchemaValidationError creates a schema validation error
func NewSchemaValidationError(message string) *AppError {
	return NewAppError(ErrCodeSchemaValidation, message)
}

// NewKafkaError creates a Kafka error
func NewKafkaError(message string) *AppError {
	return NewAppError(ErrCodeKafkaError, message).WithRetryable(true)
}

// NewCircuitBreakerError creates a circuit breaker error
func NewCircuitBreakerError(message string) *AppError {
	return NewAppError(ErrCodeCircuitBreakerOpen, message).
		WithRetryable(true).
		WithRetryAfter(30)
}

// NewDuplicateError creates a duplicate request error
func NewDuplicateError(message string) *AppError {
	return NewAppError(ErrCodeDuplicateRequest, message)
}

// NewInternalError creates an internal error
func NewInternalError(message string) *AppError {
	return NewAppError(ErrCodeInternalError, message)
}

// NewSerializationError creates a serialization error
func NewSerializationError(format string, err error) *AppError {
	return NewAppError(ErrCodeSerializationError, fmt.Sprintf("Failed to serialize to %s", format)).
		WithDetails("original_error", err.Error())
}
