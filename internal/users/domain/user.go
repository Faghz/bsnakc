package domain

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/types"
	"github.com/google/uuid"
)

// User represents a user in the system with rich domain behavior
type User struct {
	id        int64  // Internal BIGSERIAL ID for database operations
	uid       string // External UUID V7 for API exposure
	email     Email
	name      string
	username  Username
	avatarURL types.NullString
	createdAt time.Time
	createdBy string
	updatedAt time.Time
	updatedBy string
	deletedAt types.NullTime
	deletedBy types.NullString
}

// NewUser creates a new user entity with validation
func NewUser(emailStr, name, usernameStr string) (*User, error) {
	email, err := NewEmail(emailStr)
	if err != nil {
		return nil, err
	}

	if err := validateName(name); err != nil {
		return nil, err
	}

	username, err := NewUsername(usernameStr)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	uid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &User{
		id:        0, // Will be assigned by database BIGSERIAL
		uid:       uid.String(),
		email:     email,
		name:      name,
		username:  username,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstructUser reconstructs a user from persistence (bypasses validation)
func ReconstructUser(
	id int64,
	uid string,
	email Email,
	name string,
	username Username,
	avatarURL types.NullString,
	createdAt time.Time,
	createdBy string,
	updatedAt time.Time,
	updatedBy string,
	deletedAt types.NullTime,
	deletedBy types.NullString,
) *User {
	return &User{
		id:        id,
		uid:       uid,
		email:     email,
		name:      name,
		username:  username,
		avatarURL: avatarURL,
		createdAt: createdAt,
		createdBy: createdBy,
		updatedAt: updatedAt,
		updatedBy: updatedBy,
		deletedAt: deletedAt,
		deletedBy: deletedBy,
	}
}

// Getters

func (u *User) ID() int64 {
	return u.id
}

func (u *User) UID() string {
	return u.uid // Explicit external UUID getter
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Username() Username {
	return u.username
}

func (u *User) AvatarURL() string {
	return u.avatarURL.String
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) CreatedBy() string {
	return u.createdBy
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) UpdatedBy() string {
	return u.updatedBy
}

func (u *User) DeletedAt() types.NullTime {
	return u.deletedAt
}

func (u *User) DeletedBy() types.NullString {
	return u.deletedBy
}

// Behavior methods

// SetInternalID sets the internal database ID (called by repository after INSERT)
func (u *User) SetInternalID(id int64) {
	u.id = id
}

// SetEmail updates the user's email with validation
func (u *User) SetEmail(emailStr string) error {
	email, err := NewEmail(emailStr)
	if err != nil {
		return err
	}

	u.email = email
	u.updatedAt = time.Now()
	return nil
}

// SetName updates the user's name with validation
func (u *User) SetName(name string) error {
	if err := validateName(name); err != nil {
		return err
	}

	u.name = name
	u.updatedAt = time.Now()
	return nil
}

// SetUsername updates the user's username with validation
func (u *User) SetUsername(usernameStr string) error {
	username, err := NewUsername(usernameStr)
	if err != nil {
		return err
	}

	u.username = username
	u.updatedAt = time.Now()
	return nil
}

// UpdateAvatar updates the user's avatar URL
func (u *User) UpdateAvatar(url string) error {
	// Basic URL validation could be added here
	u.avatarURL = types.NewNullString(url)
	u.updatedAt = time.Now()
	return nil
}

// MarkDeleted marks the user as deleted (soft delete)
func (u *User) MarkDeleted(deletedBy string) {
	now := time.Now()
	u.deletedAt = types.NewNullTime(now)
	if deletedBy != "" {
		u.deletedBy = types.NewNullString(deletedBy)
	}
	u.updatedAt = now
}

// IsDeleted checks if the user is soft-deleted
func (u *User) IsDeleted() bool {
	return u.deletedAt.Valid
}

// validateName validates a user's name (domain rule)
func validateName(name string) error {
	if len(name) < 2 || len(name) > 100 {
		return ErrInvalidName
	}
	return nil
}
