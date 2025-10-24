package errors

import (
	"fmt"
)

// ErrorCode represents specific error types
type ErrorCode string

const (
	ErrConfigMissing     ErrorCode = "CONFIG_MISSING"
	ErrConfigInvalid     ErrorCode = "CONFIG_INVALID"
	ErrDBConnection      ErrorCode = "DB_CONNECTION"
	ErrQueryExecution    ErrorCode = "QUERY_EXECUTION"
	ErrFileRead          ErrorCode = "FILE_READ"
	ErrFileWrite         ErrorCode = "FILE_WRITE"
	ErrDataValidation    ErrorCode = "DATA_VALIDATION"
	ErrDataSerialization ErrorCode = "DATA_SERIALIZATION"
	ErrAPICall           ErrorCode = "API_CALL"
	ErrRateLimit         ErrorCode = "RATE_LIMIT"
	ErrTimeout           ErrorCode = "TIMEOUT"
)

// Error represents a structured error with context
type Error struct {
	Code      ErrorCode
	Component string
	Operation string
	Message   string
	Cause     error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s/%s] %s: %s (cause: %v)", e.Component, e.Code, e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s/%s] %s: %s", e.Component, e.Code, e.Operation, e.Message)
}

// New creates a new structured error
func New(code ErrorCode, component, operation, message string) *Error {
	return &Error{
		Code:      code,
		Component: component,
		Operation: operation,
		Message:   message,
	}
}

// Wrap wraps an existing error with context
func Wrap(cause error, code ErrorCode, component, operation, message string) *Error {
	return &Error{
		Code:      code,
		Component: component,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// NewInitError creates an initialization error
func NewInitError(component, message string, cause error) *Error {
	return Wrap(cause, ErrConfigInvalid, component, "init", message)
}
