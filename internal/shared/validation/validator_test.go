package validation

import (
	"testing"

	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

type testRequest struct {
	Email    string `validate:"required,email_format"`
	Password string `validate:"required,min=8,max=128,password_complexity"`
	Username string `validate:"required,username_format"`
	Name     string `validate:"required,min=2,max=100"`
}

func TestValidator_ValidInput(t *testing.T) {
	validator := NewValidator()

	req := testRequest{
		Email:    "test@example.com",
		Password: "Test123!@#",
		Username: "test_user",
		Name:     "Test User",
	}

	err := validator.Validate(&req)
	if err != nil {
		t.Errorf("Expected no error for valid input, got: %v", err)
	}
}

func TestValidator_MissingRequiredFields(t *testing.T) {
	validator := NewValidator()

	req := testRequest{}

	err := validator.Validate(&req)
	if err == nil {
		t.Error("Expected validation error for missing required fields")
		return
	}

	valErr, ok := err.(*errors.ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError, got %T", err)
		return
	}

	details := valErr.Details()
	fields, ok := details["fields"].(map[string]string)
	if !ok {
		t.Error("Expected fields in error details")
		return
	}

	if len(fields) != 4 {
		t.Errorf("Expected 4 field errors, got %d", len(fields))
	}
}

func TestValidator_InvalidEmail(t *testing.T) {
	validator := NewValidator()

	req := testRequest{
		Email:    "invalid-email",
		Password: "Test123!@#",
		Username: "test_user",
		Name:     "Test User",
	}

	err := validator.Validate(&req)
	if err == nil {
		t.Error("Expected validation error for invalid email")
		return
	}

	valErr, ok := err.(*errors.ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError, got %T", err)
		return
	}

	details := valErr.Details()
	fields, ok := details["fields"].(map[string]string)
	if !ok {
		t.Error("Expected fields in error details")
		return
	}

	if _, exists := fields["email"]; !exists {
		t.Error("Expected email field error")
	}
}

func TestValidator_WeakPassword(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name     string
		password string
	}{
		{"too short", "Pass1!"},
		{"no uppercase", "pass123!@#"},
		{"no lowercase", "PASS123!@#"},
		{"no digit", "Password!@#"},
		{"no special", "Password123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := testRequest{
				Email:    "test@example.com",
				Password: tc.password,
				Username: "test_user",
				Name:     "Test User",
			}

			err := validator.Validate(&req)
			if err == nil {
				t.Errorf("Expected validation error for weak password: %s", tc.name)
				return
			}

			valErr, ok := err.(*errors.ValidationError)
			if !ok {
				t.Errorf("Expected ValidationError, got %T", err)
				return
			}

			details := valErr.Details()
			fields, ok := details["fields"].(map[string]string)
			if !ok {
				t.Error("Expected fields in error details")
				return
			}

			if _, exists := fields["password"]; !exists {
				t.Error("Expected password field error")
			}
		})
	}
}

func TestValidator_InvalidUsername(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name     string
		username string
	}{
		{"too short", "ab"},
		{"too long", "this_is_a_very_long_username_that_exceeds_the_limit"},
		{"invalid chars", "test@user"},
		{"spaces", "test user"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := testRequest{
				Email:    "test@example.com",
				Password: "Test123!@#",
				Username: tc.username,
				Name:     "Test User",
			}

			err := validator.Validate(&req)
			if err == nil {
				t.Errorf("Expected validation error for invalid username: %s", tc.name)
				return
			}

			valErr, ok := err.(*errors.ValidationError)
			if !ok {
				t.Errorf("Expected ValidationError, got %T", err)
				return
			}

			details := valErr.Details()
			fields, ok := details["fields"].(map[string]string)
			if !ok {
				t.Error("Expected fields in error details")
				return
			}

			if _, exists := fields["username"]; !exists {
				t.Error("Expected username field error")
			}
		})
	}
}

func TestValidator_NameValidation(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name      string
		inputName string
		shouldErr bool
	}{
		{"valid name", "John Doe", false},
		{"too short", "J", true},
		{"minimum length", "Jo", false},
		{"maximum length", string(make([]byte, 100)), false},
		{"too long", string(make([]byte, 101)), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := testRequest{
				Email:    "test@example.com",
				Password: "Test123!@#",
				Username: "test_user",
				Name:     tc.inputName,
			}

			err := validator.Validate(&req)
			if tc.shouldErr && err == nil {
				t.Error("Expected validation error")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
