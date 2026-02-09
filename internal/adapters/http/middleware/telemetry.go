package middleware

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// OTelMiddleware returns OpenTelemetry tracing middleware for Echo
// This automatically instruments all HTTP routes with distributed tracing.
// Spans include HTTP method, route, status code, and duration.
// Must be added BEFORE logger middleware to ensure trace context is available for logs.
func OTelMiddleware(serviceName string) echo.MiddlewareFunc {
	return otelecho.Middleware(serviceName)
}
