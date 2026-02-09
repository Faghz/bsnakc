package middleware

import (
	"strings"

	"github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware creates a middleware for PASETO authentication
func AuthMiddleware(tokenService ports.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return errors.NewAuthenticationError("missing authorization header")
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return errors.NewAuthenticationError("invalid authorization header format")
			}

			token := parts[1]

			// Validate access token
			userID, err := tokenService.ValidateAccessToken(token)
			if err != nil {
				return errors.NewAuthenticationError("invalid or expired token")
			}
			c.Set("user_id", userID)

			return next(c)
		}
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c echo.Context) string {
	userID, _ := c.Get("user_id").(string)
	return userID
}
