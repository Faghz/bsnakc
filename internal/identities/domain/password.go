package domain

import (
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

// Password represents a password value object
type Password struct {
	value string
}

// NewPassword creates a new Password value object with validation
func NewPassword(plain string) (Password, error) {
	if len(plain) < 8 {
		return Password{}, errors.NewValidationError("password is too weak").WithField("password", "must be at least 8 characters")
	}

	if len(plain) > 128 {
		return Password{}, errors.NewValidationError("password is too long").WithField("password", "must be at most 128 characters")
	}

	return Password{value: plain}, nil
}

// Value returns the raw password value (use carefully, preferably only for hashing)
func (p Password) Value() string {
	return p.value
}

// HashedPassword represents a hashed password value object
type HashedPassword struct {
	hash string
}

// NewHashedPassword creates a hashed password from a hash string (from DB)
func NewHashedPassword(hash string) HashedPassword {
	return HashedPassword{hash: hash}
}

// Hash returns the password hash
func (hp HashedPassword) Hash() string {
	return hp.hash
}
