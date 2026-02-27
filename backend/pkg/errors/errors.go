package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents different types of errors
type ErrorCode int

const (
	CodeNotFound ErrorCode = iota
	CodeUnauthorized
	CodeValidation
	CodeConflict
	CodeInternal
	CodeBadRequest
	CodeForbidden
)

// Error represents a structured error with code and message
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Cause   error     `json:"-"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *Error) Unwrap() error {
	return e.Cause
}

// New creates a new error with the given code and message
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with a new message and code
func Wrap(code ErrorCode, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewNotFound creates a not found error
func NewNotFound(message string) *Error {
	return New(CodeNotFound, message)
}

// NewUnauthorized creates an unauthorized error
func NewUnauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

// NewValidation creates a validation error
func NewValidation(message string) *Error {
	return New(CodeValidation, message)
}

// NewConflict creates a conflict error
func NewConflict(message string) *Error {
	return New(CodeConflict, message)
}

// NewInternal creates an internal error
func NewInternal(message string) *Error {
	return New(CodeInternal, message)
}

// NewBadRequest creates a bad request error
func NewBadRequest(message string) *Error {
	return New(CodeBadRequest, message)
}

// NewForbidden creates a forbidden error
func NewForbidden(message string) *Error {
	return New(CodeForbidden, message)
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *Error) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeValidation:
		return http.StatusBadRequest
	case CodeConflict:
		return http.StatusConflict
	case CodeInternal:
		return http.StatusInternalServerError
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeNotFound
	}
	return false
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeUnauthorized
	}
	return false
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeValidation
	}
	return false
}

// IsConflict checks if the error is a conflict error
func IsConflict(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeConflict
	}
	return false
}

// IsInternal checks if the error is an internal error
func IsInternal(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == CodeInternal
	}
	return false
}
