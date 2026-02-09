package domain

import (
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

var (
	ErrUserNotFound    = errors.NewNotFoundError("user")
	ErrEmailRegistered = errors.NewConflictError("email already registered")
	ErrUsernameTaken   = errors.NewConflictError("username already taken")
	ErrInvalidEmail    = errors.NewValidationError("invalid email address").WithField("email", "must be a valid email address")
	ErrInvalidName     = errors.NewValidationError("invalid name").WithField("name", "must be between 2 and 100 characters")
	ErrInvalidUsername = errors.NewValidationError("invalid username").WithField("username", "must be 3-32 characters (alphanumeric, underscore, hyphen only)")
)
