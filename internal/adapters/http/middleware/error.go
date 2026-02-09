package middleware

import (
	"net/http"

	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/labstack/echo/v4"
)

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorHandler is a custom error handler for Echo
func ErrorHandler(logger ports.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		errorCode := "internal_error"
		message := "internal server error"
		var details map[string]interface{}

		ctx := c.Request().Context()

		// Check if it's an AppError
		if appErr, ok := errors.AsAppError(err); ok {
			code = appErr.StatusCode()
			errorCode = appErr.ErrorCode()
			message = appErr.Error()
			details = appErr.Details()

			// Log internal errors with full context, but only log message for client errors
			if code >= 500 {
				logger.Error(ctx, "Internal server error",
					ports.Error(err),
					ports.Int("status", code),
					ports.String("error_code", errorCode),
					ports.String("path", c.Request().URL.Path),
					ports.String("method", c.Request().Method),
				)
				// Mask internal error details for security
				message = "internal server error"
				details = nil
			} else {
				logger.Warn(ctx, "Client error",
					ports.String("message", message),
					ports.Int("status", code),
					ports.String("error_code", errorCode),
					ports.String("path", c.Request().URL.Path),
					ports.String("method", c.Request().Method),
				)
			}
		} else if he, ok := err.(*echo.HTTPError); ok {
			// Handle Echo HTTP errors
			code = he.Code
			errorCode = http.StatusText(code)
			if msg, ok := he.Message.(string); ok {
				message = msg
			}

			logger.Warn(ctx, "Echo HTTP error",
				ports.Error(err),
				ports.Int("status", code),
				ports.String("path", c.Request().URL.Path),
				ports.String("method", c.Request().Method),
			)
		} else {
			// Unknown error - log with full context
			logger.Error(ctx, "Unhandled error",
				ports.Error(err),
				ports.Int("status", code),
				ports.String("path", c.Request().URL.Path),
				ports.String("method", c.Request().Method),
			)
		}

		// Don't send response if already committed
		if !c.Response().Committed {
			_ = c.JSON(code, ErrorResponse{
				Error:   errorCode,
				Message: message,
				Details: details,
			})
		}
	}
}
