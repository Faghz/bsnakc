package domain

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/elzestia/go-boilerplate/internal/shared/types"
	"github.com/google/uuid"
)

// Provider represents the authentication provider type
type Provider string

const (
	ProviderLocal   Provider = "local"
	ProviderDiscord Provider = "discord"
)

// Identity represents an authentication identity with rich domain behavior
type Identity struct {
	id             int64  // Internal BIGSERIAL ID for database operations
	uid            string // External UUID for API exposure
	userID         int64  // Internal user ID (BIGSERIAL FK)
	provider       Provider
	providerID     types.NullString
	hashedPassword HashedPassword // Only for local provider
	createdAt      time.Time
	createdBy      int64
	updatedAt      time.Time
	updatedBy      int64
	deletedAt      types.NullTime
	deletedBy      types.NullString
}

// NewLocalIdentity creates a new local authentication identity
func NewLocalIdentity(userID int64, userUID string, hashedPassword HashedPassword) (*Identity, error) {
	if userID == 0 {
		return nil, errors.NewValidationError("userID is required")
	}

	if userUID == "" {
		return nil, errors.NewValidationError("userUID is required")
	}

	now := time.Now()
	uid, err := uuid.NewV7() // Generate UUID V7 for external use
	if err != nil {
		return nil, err
	}

	return &Identity{
		id:             0,            // Will be assigned by database BIGSERIAL
		uid:            uid.String(), // Generated here for external API use
		userID:         userID,
		provider:       ProviderLocal,
		providerID:     types.NullString{}, // For local auth, no external provider ID (null)
		hashedPassword: hashedPassword,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// NewSSOIdentity creates a new SSO authentication identity
func NewSSOIdentity(userID int64, userUID string, provider Provider, providerID string) (*Identity, error) {
	if userID == 0 {
		return nil, errors.NewValidationError("userID is required")
	}

	if userUID == "" {
		return nil, errors.NewValidationError("userUID is required")
	}

	if err := ValidateProvider(provider); err != nil {
		return nil, err
	}

	if providerID == "" {
		return nil, errors.NewValidationError("providerID is required for SSO providers")
	}

	now := time.Now()
	uid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Identity{
		id:         0,            // Will be assigned by database BIGSERIAL
		uid:        uid.String(), // Generated here for external API use
		userID:     userID,
		provider:   provider,
		providerID: types.NullStringFrom(providerID),
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// ReconstructIdentity reconstructs an identity from persistence (bypasses validation)
func ReconstructIdentity(
	id int64,
	uid string,
	userID int64,
	userUID string,
	provider Provider,
	providerID types.NullString,
	hashedPassword HashedPassword,
	createdAt time.Time,
	createdBy int64,
	updatedAt time.Time,
	updatedBy int64,
	deletedAt types.NullTime,
	deletedBy types.NullString,
) *Identity {
	return &Identity{
		id:             id,
		uid:            uid,
		userID:         userID,
		provider:       provider,
		providerID:     providerID,
		hashedPassword: hashedPassword,
		createdAt:      createdAt,
		createdBy:      createdBy,
		updatedAt:      updatedAt,
		updatedBy:      updatedBy,
		deletedAt:      deletedAt,
		deletedBy:      deletedBy,
	}
}

// Getters

func (i *Identity) ID() string {
	return i.uid // Returns external UUID for backward compatibility
}

func (i *Identity) UID() string {
	return i.uid // Explicit external UUID getter
}

func (i *Identity) InternalID() int64 {
	return i.id // Internal BIGSERIAL ID for database operations
}

func (i *Identity) InternalUserID() int64 {
	return i.userID // Internal user ID for FK operations
}

func (i *Identity) Provider() Provider {
	return i.provider
}

func (i *Identity) ProviderID() types.NullString {
	return i.providerID
}

func (i *Identity) HashedPassword() HashedPassword {
	return i.hashedPassword
}

func (i *Identity) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Identity) CreatedBy() int64 {
	return i.createdBy
}

func (i *Identity) UpdatedAt() time.Time {
	return i.updatedAt
}

func (i *Identity) UpdatedBy() int64 {
	return i.updatedBy
}

func (i *Identity) DeletedAt() types.NullTime {
	return i.deletedAt
}

func (i *Identity) DeletedBy() types.NullString {
	return i.deletedBy
}

// Behavior methods

// SetInternalID sets the internal database ID (called by repository after INSERT)
func (i *Identity) SetInternalID(id int64) {
	i.id = id
}

// VerifyPassword verifies a plain password against the stored hash
func (i *Identity) VerifyPassword(hasher PasswordHasher, plainPassword string) error {
	if !i.IsLocal() {
		return errors.NewValidationError("cannot verify password for non-local identity")
	}

	password, err := NewPassword(plainPassword)
	if err != nil {
		return ErrInvalidCredentials
	}

	match, err := hasher.Verify(password.Value(), i.hashedPassword.Hash())
	if err != nil {
		return err
	}

	if !match {
		return ErrInvalidCredentials
	}

	return nil
}

// ChangePassword changes the identity's password (only for local identities)
func (i *Identity) ChangePassword(hasher PasswordHasher, newPlainPassword string) error {
	if !i.IsLocal() {
		return errors.NewValidationError("cannot change password for non-local identity")
	}

	password, err := NewPassword(newPlainPassword)
	if err != nil {
		return err
	}

	hash, err := hasher.Hash(password.Value())
	if err != nil {
		return err
	}

	i.hashedPassword = NewHashedPassword(hash)
	i.updatedAt = time.Now()
	return nil
}

// IsLocal returns true if this is a local authentication identity
func (i *Identity) IsLocal() bool {
	return i.provider == ProviderLocal
}

// IsSSO returns true if this is an SSO authentication identity
func (i *Identity) IsSSO() bool {
	return i.provider != ProviderLocal
}

// MarkDeleted marks the identity as deleted (soft delete)
func (i *Identity) MarkDeleted(deletedBy string) {
	now := time.Now()
	i.deletedAt = types.NewNullTime(now)
	if deletedBy != "" {
		i.deletedBy = types.NewNullString(deletedBy)
	}
	i.updatedAt = now
}

// IsDeleted checks if the identity is soft-deleted
func (i *Identity) IsDeleted() bool {
	return i.deletedAt.Valid
}

// ValidateProvider validates a provider value
func ValidateProvider(provider Provider) error {
	switch provider {
	case ProviderLocal, ProviderDiscord:
		return nil
	default:
		return ErrInvalidProvider
	}
}
