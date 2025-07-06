// Package validator provides input validation utilities for the Event Shark application.
package validator

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dipjyotimetia/event-shark/gen"
	"github.com/dipjyotimetia/event-shark/pkg/errors"
)

// Validator interface defines validation methods.
type Validator interface {
	ValidateExpense(expense gen.Expense) error
	ValidatePayment(payment gen.Payment) error
}

// AppValidator implements the Validator interface.
type AppValidator struct{}

// NewValidator creates a new validator instance.
func NewValidator() Validator {
	return &AppValidator{}
}

// ValidateExpense validates expense data.
func (v *AppValidator) ValidateExpense(expense gen.Expense) error {
	var errs []string

	// Validate ExpenseID
	if expense.ExpenseID == "" {
		errs = append(errs, "expense_id is required")
	} else if len(expense.ExpenseID) > 50 {
		errs = append(errs, "expense_id must be less than 50 characters")
	}

	// Validate UserID
	if expense.UserID == "" {
		errs = append(errs, "user_id is required")
	} else if len(expense.UserID) > 50 {
		errs = append(errs, "user_id must be less than 50 characters")
	}

	// Validate Category
	if expense.Category == "" {
		errs = append(errs, "category is required")
	} else if !isValidCategory(expense.Category) {
		errs = append(errs, "category must be one of: food, transport, entertainment, healthcare, education, other")
	}

	// Validate Amount
	if expense.Amount <= 0 {
		errs = append(errs, "amount must be greater than 0")
	} else if expense.Amount > 999999.99 {
		errs = append(errs, "amount must be less than 999999.99")
	}

	// Validate Currency
	if expense.Currency == "" {
		errs = append(errs, "currency is required")
	} else if !isValidCurrency(expense.Currency) {
		errs = append(errs, "invalid currency code")
	}

	// Validate Timestamp
	if expense.Timestamp < 0 {
		errs = append(errs, "timestamp must be non-negative")
	} else if expense.Timestamp > time.Now().UnixNano()/int64(time.Millisecond)+86400000 { // 24 hours in future
		errs = append(errs, "timestamp cannot be more than 24 hours in the future")
	}

	// Validate Description (optional)
	if expense.Description != nil && len(*expense.Description) > 500 {
		errs = append(errs, "description must be less than 500 characters")
	}

	if len(errs) > 0 {
		return errors.ErrValidation(strings.Join(errs, "; "), nil)
	}

	return nil
}

// ValidatePayment validates payment data.
func (v *AppValidator) ValidatePayment(payment gen.Payment) error {
	var errs []string

	// Validate TransactionID
	if payment.TransactionID == "" {
		errs = append(errs, "transaction_id is required")
	} else if len(payment.TransactionID) > 50 {
		errs = append(errs, "transaction_id must be less than 50 characters")
	}

	// Validate UserID
	if payment.UserID == "" {
		errs = append(errs, "user_id is required")
	} else if len(payment.UserID) > 50 {
		errs = append(errs, "user_id must be less than 50 characters")
	}

	// Validate Amount
	if payment.Amount <= 0 {
		errs = append(errs, "amount must be greater than 0")
	} else if payment.Amount > 999999.99 {
		errs = append(errs, "amount must be less than 999999.99")
	}

	// Validate Currency
	if payment.Currency == "" {
		errs = append(errs, "currency is required")
	} else if !isValidCurrency(payment.Currency) {
		errs = append(errs, "invalid currency code")
	}

	// Validate PaymentMethod
	if payment.PaymentMethod == "" {
		errs = append(errs, "payment_method is required")
	} else if !isValidPaymentMethod(payment.PaymentMethod) {
		errs = append(errs, "payment_method must be one of: CREDIT_CARD, DEBIT_CARD, BANK_TRANSFER, PAYPAL, CASH")
	}

	// Validate Status
	if payment.Status == "" {
		errs = append(errs, "status is required")
	} else if !isValidPaymentStatus(payment.Status) {
		errs = append(errs, "status must be one of: PENDING, COMPLETED, FAILED, CANCELLED")
	}

	// Validate Timestamp
	if payment.Timestamp < 0 {
		errs = append(errs, "timestamp must be non-negative")
	} else if payment.Timestamp > time.Now().UnixNano()/int64(time.Millisecond)+86400000 { // 24 hours in future
		errs = append(errs, "timestamp cannot be more than 24 hours in the future")
	}

	if len(errs) > 0 {
		return errors.ErrValidation(strings.Join(errs, "; "), nil)
	}

	return nil
}

// isValidCategory checks if the category is valid.
func isValidCategory(category string) bool {
	validCategories := []string{"food", "transport", "entertainment", "healthcare", "education", "other"}
	for _, vc := range validCategories {
		if strings.EqualFold(category, vc) {
			return true
		}
	}
	return false
}

// isValidCurrency checks if the currency code is valid (ISO 4217).
func isValidCurrency(currency string) bool {
	// Simplified validation - in production, use a comprehensive list
	validCurrencies := []string{"USD", "EUR", "GBP", "JPY", "AUD", "CAD", "CHF", "CNY", "SEK", "NZD"}
	for _, vc := range validCurrencies {
		if strings.EqualFold(currency, vc) {
			return true
		}
	}
	return false
}

// isValidPaymentMethod checks if the payment method is valid.
func isValidPaymentMethod(method string) bool {
	validMethods := []string{"CREDIT_CARD", "DEBIT_CARD", "BANK_TRANSFER", "PAYPAL", "CASH"}
	for _, vm := range validMethods {
		if strings.EqualFold(method, vm) {
			return true
		}
	}
	return false
}

// isValidPaymentStatus checks if the payment status is valid.
func isValidPaymentStatus(status string) bool {
	validStatuses := []string{"PENDING", "COMPLETED", "FAILED", "CANCELLED"}
	for _, vs := range validStatuses {
		if strings.EqualFold(status, vs) {
			return true
		}
	}
	return false
}

// ValidateID checks if an ID matches a specific format.
func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}

	// Allow alphanumeric, hyphens, and underscores
	matched, err := regexp.MatchString("^[a-zA-Z0-9_-]+$", id)
	if err != nil {
		return fmt.Errorf("error validating ID format: %w", err)
	}

	if !matched {
		return fmt.Errorf("ID can only contain alphanumeric characters, hyphens, and underscores")
	}

	return nil
}
