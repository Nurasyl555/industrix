package errors

import (
	"fmt"
	"net/http"
)

type Code string

const (
	CodeNotFound     Code = "NOT_FOUND"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeValidation   Code = "VALIDATION"
	CodeConflict     Code = "CONFLICT"
	CodeInternal     Code = "INTERNAL"
)

type Error struct {
	Code    Code
	Message string
	Details map[string]interface{}
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func HTTPStatus(code Code) int {
	switch code {
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
	default:
		return http.StatusInternalServerError
	}
}
