package errors

import (
	"fmt"
	"net/http"
)

// AppError is the interface that all application errors must implement.
// It provides methods to get HTTP status codes, error codes, and additional details.
type AppError interface {
	error
	// StatusCode returns the HTTP status code for this error
	StatusCode() int
	// ErrorCode returns a machine-readable error code
	ErrorCode() string
	// Details returns additional context about the error (e.g., validation field errors)
	Details() map[string]interface{}
	// Unwrap returns the underlying error if any
	Unwrap() error
}

// baseError provides common functionality for all error types
type baseError struct {
	message    string
	statusCode int
	errorCode  string
	details    map[string]interface{}
	cause      error
}

func (e *baseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *baseError) StatusCode() int {
	return e.statusCode
}

func (e *baseError) ErrorCode() string {
	return e.errorCode
}

func (e *baseError) Details() map[string]interface{} {
	return e.details
}

func (e *baseError) Unwrap() error {
	return e.cause
}

// ValidationError represents a validation error (HTTP 400)
type ValidationError struct {
	*baseError
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		baseError: &baseError{
			message:    message,
			statusCode: http.StatusBadRequest,
			errorCode:  "validation_error",
			details:    make(map[string]interface{}),
		},
	}
}

// WithField adds a field-level validation error
func (e *ValidationError) WithField(field, message string) *ValidationError {
	if e.details == nil {
		e.details = make(map[string]interface{})
	}
	if e.details["fields"] == nil {
		e.details["fields"] = make(map[string]string)
	}
	e.details["fields"].(map[string]string)[field] = message
	return e
}

// WithDetails adds custom details to the error
func (e *ValidationError) WithDetails(key string, value interface{}) *ValidationError {
	if e.details == nil {
		e.details = make(map[string]interface{})
	}
	e.details[key] = value
	return e
}

// NotFoundError represents a resource not found error (HTTP 404)
type NotFoundError struct {
	*baseError
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *NotFoundError {
	return &NotFoundError{
		baseError: &baseError{
			message:    fmt.Sprintf("%s not found", resource),
			statusCode: http.StatusNotFound,
			errorCode:  "not_found",
			details:    map[string]interface{}{"resource": resource},
		},
	}
}

// WithCause wraps an underlying error
func (e *NotFoundError) WithCause(cause error) *NotFoundError {
	e.cause = cause
	return e
}

// ConflictError represents a resource conflict error (HTTP 409)
type ConflictError struct {
	*baseError
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		baseError: &baseError{
			message:    message,
			statusCode: http.StatusConflict,
			errorCode:  "conflict",
			details:    make(map[string]interface{}),
		},
	}
}

// WithDetails adds custom details to the error
func (e *ConflictError) WithDetails(key string, value interface{}) *ConflictError {
	if e.details == nil {
		e.details = make(map[string]interface{})
	}
	e.details[key] = value
	return e
}

// AuthenticationError represents an authentication error (HTTP 401)
type AuthenticationError struct {
	*baseError
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		baseError: &baseError{
			message:    message,
			statusCode: http.StatusUnauthorized,
			errorCode:  "unauthorized",
			details:    make(map[string]interface{}),
		},
	}
}

// AuthorizationError represents an authorization error (HTTP 403)
type AuthorizationError struct {
	*baseError
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string) *AuthorizationError {
	return &AuthorizationError{
		baseError: &baseError{
			message:    message,
			statusCode: http.StatusForbidden,
			errorCode:  "forbidden",
			details:    make(map[string]interface{}),
		},
	}
}

// BadRequestError represents a bad request error (HTTP 400)
type BadRequestError struct {
	*baseError
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *BadRequestError {
	return &BadRequestError{
		baseError: &baseError{
			message:    message,
			statusCode: http.StatusBadRequest,
			errorCode:  "bad_request",
			details:    make(map[string]interface{}),
		},
	}
}

// WithCause wraps an underlying error
func (e *BadRequestError) WithCause(cause error) *BadRequestError {
	e.cause = cause
	return e
}

// InternalError represents an internal server error (HTTP 500)
type InternalError struct {
	*baseError
}

// NewInternalError creates a new internal error
// The actual error message is masked for security; the cause is logged but not exposed
func NewInternalError(publicMessage string, cause error) *InternalError {
	return &InternalError{
		baseError: &baseError{
			message:    publicMessage,
			statusCode: http.StatusInternalServerError,
			errorCode:  "internal_error",
			details:    make(map[string]interface{}),
			cause:      cause,
		},
	}
}

// Wrap wraps a standard error into an InternalError
func Wrap(err error, message string) *InternalError {
	return NewInternalError(message, err)
}

// IsAppError checks if an error implements the AppError interface
func IsAppError(err error) bool {
	_, ok := err.(AppError)
	return ok
}

// AsAppError attempts to cast an error to AppError
func AsAppError(err error) (AppError, bool) {
	appErr, ok := err.(AppError)
	return appErr, ok
}

func IsNotFoundError(err error) bool {
	appErr, ok := AsAppError(err)
	return ok && appErr.StatusCode() == http.StatusNotFound
}
