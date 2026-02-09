package domain

import (
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

// Domain errors
var (
	ErrIdentityNotFound      = errors.NewNotFoundError("identity")
	ErrIdentityAlreadyExists = errors.NewConflictError("identity already exists")
	ErrInvalidCredentials    = errors.NewAuthenticationError("invalid credentials")
	ErrInvalidProvider       = errors.NewValidationError("invalid provider").WithField("provider", "must be a valid OAuth provider")
	ErrWeakPassword          = errors.NewValidationError("password is too weak").WithField("password", "must be at least 8 characters long")
	ErrInvalidToken          = errors.NewAuthenticationError("invalid token")
	ErrTokenExpired          = errors.NewAuthenticationError("token expired")
	ErrRefreshTokenRevoked   = errors.NewAuthenticationError("refresh token revoked")
	ErrRefreshTokenNotFound  = errors.NewAuthenticationError("refresh token not found")
)
