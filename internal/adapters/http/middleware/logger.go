package middleware

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/labstack/echo/v4"
)

// Logger returns a logging middleware that uses the Logger port interface
func Logger(logger ports.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			fields := []ports.Field{
				ports.String("method", req.Method),
				ports.String("uri", req.RequestURI),
				ports.Int("status", res.Status),
				ports.Duration("latency", latency),
				ports.String("remote_ip", c.RealIP()),
				ports.String("user_agent", req.UserAgent()),
			}

			// Add user_id if available in context
			if userID := getUserIDFromContext(c.Request().Context()); userID != "" {
				fields = append(fields, ports.String("user_id", userID))
			}

			if err != nil {
				fields = append(fields, ports.Error(err))
			}

			if res.Status >= 500 {
				logger.Error(req.Context(), "Server error", fields...)
			} else if res.Status >= 400 {
				logger.Warn(req.Context(), "Client error", fields...)
			} else {
				logger.Info(req.Context(), "Request completed", fields...)
			}

			return nil
		}
	}
}

// getUserIDFromContext extracts user_id from Go context (set by auth middleware)
func getUserIDFromContext(ctx interface{}) string {
	type userIDKey struct{}
	if v := ctx.(interface{ Value(interface{}) interface{} }).Value(userIDKey{}); v != nil {
		if userID, ok := v.(string); ok {
			return userID
		}
	}
	return ""
}
