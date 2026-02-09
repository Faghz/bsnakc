package validation

import (
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/go-playground/validator/v10"
)

// Validator is the interface for request validation
type Validator interface {
	Validate(i interface{}) error
}

// PlaygroundValidator wraps go-playground/validator
type PlaygroundValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance with custom validators registered
func NewValidator() Validator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("username_format", validateUsername)
	v.RegisterValidation("password_complexity", validatePasswordComplexity)
	v.RegisterValidation("email_format", validateEmail)

	return &PlaygroundValidator{
		validator: v,
	}
}

// Validate validates a struct and returns ValidationError with field details
func (pv *PlaygroundValidator) Validate(i interface{}) error {
	if err := pv.validator.Struct(i); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			return MapValidationErrors(validationErrs)
		}
		// Non-validation error (e.g., invalid type passed)
		return errors.NewBadRequestError("validation failed").WithCause(err)
	}
	return nil
}
