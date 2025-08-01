// Package errors provides custom error types and error handling utilities for the Event Shark application.
package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error with additional context.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code: %d, message: %s, error: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// NewAppError creates a new AppError instance.
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Predefined error types.
var (
	ErrBadRequest     = func(msg string, err error) *AppError { return NewAppError(http.StatusBadRequest, msg, err) }
	ErrInternalServer = func(msg string, err error) *AppError { return NewAppError(http.StatusInternalServerError, msg, err) }
	ErrUnauthorized   = func(msg string, err error) *AppError { return NewAppError(http.StatusUnauthorized, msg, err) }
	ErrForbidden      = func(msg string, err error) *AppError { return NewAppError(http.StatusForbidden, msg, err) }
	ErrNotFound       = func(msg string, err error) *AppError { return NewAppError(http.StatusNotFound, msg, err) }
	ErrValidation     = func(msg string, err error) *AppError { return NewAppError(http.StatusUnprocessableEntity, msg, err) }
)

// Common error messages.
const (
	MsgInvalidJSON        = "invalid JSON format"
	MsgValidationFailed   = "validation failed"
	MsgKafkaProduceFailed = "failed to produce message to Kafka"
	MsgSchemaError        = "schema processing error"
	MsgConfigError        = "configuration error"
)
