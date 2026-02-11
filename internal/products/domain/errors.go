package domain

import (
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

var (
	ErrProductTypeNotFound    = errors.NewNotFoundError("product type")
	ErrProductTypeTaken       = errors.NewConflictError("product type name already exists")
	ErrInvalidProductTypeName = errors.NewValidationError("invalid product type name").WithField("name", "must be between 2 and 100 characters")
	ErrProductNotFound        = errors.NewNotFoundError("product")
	ErrInvalidProductName     = errors.NewValidationError("invalid product name").WithField("name", "must be between 2 and 100 characters")
)
