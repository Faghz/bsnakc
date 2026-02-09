package domain

import (
	"regexp"
	"strings"
)

// Username validation: 3-32 characters, alphanumeric + underscore + hyphen
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)

// Username represents a unique public username value object.
// Unlike email (PII that's encrypted), username is a public identifier
// used for profile lookups and display.
type Username struct {
	value string
}

// NewUsername creates a new Username value object with validation.
// Rules:
// - 3-32 characters
// - Alphanumeric, underscore, and hyphen only
// - Case-insensitive for uniqueness (stored as-is, compared lowercase)
func NewUsername(username string) (Username, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return Username{}, ErrInvalidUsername
	}

	if !usernameRegex.MatchString(username) {
		return Username{}, ErrInvalidUsername
	}

	return Username{value: username}, nil
}

// String returns the username as a string.
func (u Username) String() string {
	return u.value
}

// Value returns the raw username value.
func (u Username) Value() string {
	return u.value
}

// Equals checks if two usernames are equal (case-insensitive).
func (u Username) Equals(other Username) bool {
	return strings.EqualFold(u.value, other.value)
}

// IsEmpty checks if the username is empty.
func (u Username) IsEmpty() bool {
	return u.value == ""
}
