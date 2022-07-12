package model

import (
	"errors"
	"fmt"
)

type (
	// BackendError represents a common beckend error
	// from this object we can extract certain information about the error
	BackendError struct {
		Cause    error  `json:"-"`
		Message  string `json:"message,omitempty"`
		TechInfo string `json:"tech_info,omitempty"`
	}
)

// Error implements error interface
func (e *BackendError) Error() string {
	return e.Message
}

// Unwrap returns cause for this error
// it implements interface for errors.Unwrap() function
func (e *BackendError) Unwrap() error {
	return e.Cause
}

// Errors
var (
	ErrEntityNotFound           = errors.New("entity not found")
	ErrUnableToLogin            = errors.New("unable to login")
	ErrApplicationAlreadyExists = errors.New("application already exists")
	ErrUnauthorized             = errors.New("user not authorized")
	ErrInsufficientRights       = errors.New("insufficient rights")
	ErrDuplication              = errors.New("duplication")
	ErrInappropriateAction      = errors.New("inappropriate action")
	ErrValidationFailed         = errors.New("validation failed")
	ErrInvalidFormat            = errors.New("invalid format")
)

// Validation errors
var (
	ValidationErrorRequired       = func(field string) string { return fmt.Sprintf("%s: is required", field) }
	ValidationErrorMustBePositive = func(field string) string { return fmt.Sprintf("%s: must be positive", field) }
)
