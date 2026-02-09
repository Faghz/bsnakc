package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Custom validator functions that match domain validation rules

// validateUsername validates username matching domain regex: ^[a-zA-Z0-9_-]{3,32}$
// Rules:
// - 3-32 characters
// - Alphanumeric, underscore, and hyphen only
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{3,32}$`, username)
	return matched
}

// validateEmail validates email matching domain regex
// Pattern: ^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, email)
	return matched
}

// validatePasswordComplexity enforces password complexity requirements:
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
// - At least one special character
func validatePasswordComplexity(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSpecial
}
