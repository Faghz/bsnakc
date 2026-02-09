package domain

import "github.com/elzestia/go-boilerplate/internal/shared/errors"

// Authentication errors (owned by auth domain)
// These errors are used throughout the auth context for authentication-related failures.

var (
	// Credential errors
	ErrInvalidCredentials = errors.NewAuthenticationError("invalid credentials")
	ErrWeakPassword       = errors.NewValidationError("password does not meet requirements").
				WithField("password", "must be at least 8 characters")
	ErrInvalidProvider       = errors.NewValidationError("invalid authentication provider")
	ErrIdentityNotFound      = errors.NewNotFoundError("identity not found")
	ErrIdentityAlreadyExists = errors.NewConflictError("identity already exists for this user")

	// Token errors
	ErrInvalidToken         = errors.NewAuthenticationError("invalid token")
	ErrTokenExpired         = errors.NewAuthenticationError("token has expired")
	ErrRefreshTokenRevoked  = errors.NewAuthenticationError("refresh token has been revoked")
	ErrRefreshTokenNotFound = errors.NewNotFoundError("refresh token not found")

	// User errors (auth's view of user operations)
	ErrUserNotFound        = errors.NewNotFoundError("user not found")
	ErrEmailAlreadyUsed    = errors.NewConflictError("email already used by another user")
	ErrUsernameAlreadyUsed = errors.NewConflictError("username already used by another user")
	ErrUserAlreadyExists   = errors.NewConflictError("user already exists")
)
