package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents an error code
type ErrorCode string

const (
	// CodeOK represents no error
	CodeOK ErrorCode = "OK"
	// CodeNotFound represents resource not found
	CodeNotFound ErrorCode = "NOT_FOUND"
	// CodeUnauthorized represents unauthorized access
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// CodeForbidden represents forbidden access
	CodeForbidden ErrorCode = "FORBIDDEN"
	// CodeValidation represents validation error
	CodeValidation ErrorCode = "VALIDATION"
	// CodeConflict represents conflict error
	CodeConflict ErrorCode = "CONFLICT"
	// CodeInternal represents internal server error
	CodeInternal ErrorCode = "INTERNAL"
	// CodeBadRequest represents bad request
	CodeBadRequest ErrorCode = "BAD_REQUEST"
	// CodeTooManyRequests represents rate limit error
	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
	// CodeServiceUnavailable represents service unavailable
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// Error represents an application error
type Error struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Internal   error       `json:"-"`
	StackTrace string      `json:"-"`
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the internal error
func (e *Error) Unwrap() error {
	return e.Internal
}

// WithDetails adds details to the error
func (e *Error) WithDetails(details interface{}) *Error {
	e.Details = details
	return e
}

// WithInternal adds an internal error
func (e *Error) WithInternal(internal error) *Error {
	e.Internal = internal
	return e
}

// WithStack adds a stack trace
func (e *Error) WithStack(stack string) *Error {
	e.StackTrace = stack
	return e
}

// New creates a new error
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an error with code and message
func Wrap(code ErrorCode, message string, err error) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Internal: err,
	}
}

// Is checks if the error is of the given type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As casts the error to the given type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// HTTPStatus returns the HTTP status code for an error code
func HTTPStatus(code ErrorCode) int {
	switch code {
	case CodeOK:
		return http.StatusOK
	case CodeNotFound:
		return http.StatusNotFound
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeValidation:
		return http.StatusUnprocessableEntity
	case CodeConflict:
		return http.StatusConflict
	case CodeInternal:
		return http.StatusInternalServerError
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors
var (
	ErrNotFound     = New(CodeNotFound, "resource not found")
	ErrUnauthorized = New(CodeUnauthorized, "unauthorized")
	ErrForbidden    = New(CodeForbidden, "forbidden")
	ErrValidation   = New(CodeValidation, "validation failed")
	ErrConflict     = New(CodeConflict, "conflict")
	ErrInternal     = New(CodeInternal, "internal server error")
	ErrBadRequest   = New(CodeBadRequest, "bad request")
)

// NotFound creates a NOT_FOUND error
func NotFound(message string) *Error {
	return New(CodeNotFound, message)
}

// Unauthorized creates an UNAUTHORIZED error
func Unauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

// Forbidden creates a FORBIDDEN error
func Forbidden(message string) *Error {
	return New(CodeForbidden, message)
}

// Validation creates a VALIDATION error
func Validation(message string) *Error {
	return New(CodeValidation, message)
}

// Conflict creates a CONFLICT error
func Conflict(message string) *Error {
	return New(CodeConflict, message)
}

// Internal creates an INTERNAL error
func Internal(message string) *Error {
	return New(CodeInternal, message)
}

// BadRequest creates a BAD_REQUEST error
func BadRequest(message string) *Error {
	return New(CodeBadRequest, message)
}

// ToHTTPStatus converts error to HTTP status
func ToHTTPStatus(err error) int {
	var appErr *Error
	if As(err, &appErr) {
		return HTTPStatus(appErr.Code)
	}
	return http.StatusInternalServerError
}
