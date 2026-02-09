package domain

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email represents an email address value object
type Email struct {
	value string
}

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	if email == "" {
		return Email{}, ErrInvalidEmail
	}

	if !emailRegex.MatchString(email) {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: email}, nil
}

// String returns the email as a string
func (e Email) String() string {
	return e.value
}

// Value returns the raw email value
func (e Email) Value() string {
	return e.value
}

// Equals checks if two emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// IsEmpty checks if the email is empty
func (e Email) IsEmpty() bool {
	return e.value == ""
}
