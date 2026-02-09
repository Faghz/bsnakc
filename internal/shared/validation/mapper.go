package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/go-playground/validator/v10"
)

// MapValidationErrors converts validator.ValidationErrors to errors.ValidationError
func MapValidationErrors(validationErrs validator.ValidationErrors) *errors.ValidationError {
	valErr := errors.NewValidationError("validation failed")

	for _, err := range validationErrs {
		fieldName := toSnakeCase(err.Field())
		message := formatErrorMessage(err)
		valErr.WithField(fieldName, message)
	}

	return valErr
}

// formatErrorMessage creates user-friendly validation error messages
func formatErrorMessage(err validator.FieldError) string {
	field := err.Field()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "must be a valid email address"
	case "email_format":
		return "must be a valid email address"
	case "username_format":
		return "must contain only letters, numbers, underscore, and hyphen"
	case "password_complexity":
		return "must contain uppercase, lowercase, number, and special character"
	case "min":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("must be at least %s characters long", err.Param())
		}
		return fmt.Sprintf("must be at least %s", err.Param())
	case "max":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("must be at most %s characters long", err.Param())
		}
		return fmt.Sprintf("must be at most %s", err.Param())
	case "url":
		return "must be a valid URL"
	case "uri":
		return "must be a valid URI"
	case "alpha":
		return "must contain only alphabetic characters"
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "numeric":
		return "must be a valid number"
	case "len":
		return fmt.Sprintf("must be exactly %s characters long", err.Param())
	case "eq":
		return fmt.Sprintf("must be equal to %s", err.Param())
	case "ne":
		return fmt.Sprintf("must not be equal to %s", err.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", err.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", err.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", err.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", err.Param())
	case "eqfield":
		return fmt.Sprintf("must be equal to %s", err.Param())
	case "nefield":
		return fmt.Sprintf("must not be equal to %s", err.Param())
	default:
		return fmt.Sprintf("failed validation for '%s'", err.Tag())
	}
}

// toSnakeCase converts PascalCase/camelCase to snake_case for JSON field names
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
