package application

import (
	"context"
	"fmt"

	"github.com/elzestia/go-boilerplate/internal/identities/application/ports"
	"github.com/elzestia/go-boilerplate/internal/identities/domain"
	sharedports "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

// CreateLocalIdentity creates a new local authentication identity.
// It validates the password, hashes it, and persists the identity.
func (s *IdentityService) CreateLocalIdentity(ctx context.Context, userID int64, userUID, password string) (*domain.Identity, error) {
	s.logger.Debug(ctx, "Creating local identity", sharedports.Int64("user_id", userID))

	// Validate password strength (domain validation)
	passwordVO, err := domain.NewPassword(password)
	if err != nil {
		s.logger.Warn(ctx, "Password validation failed", sharedports.Error(err))
		return nil, err
	}

	// Hash password using the port
	hash, err := s.passwordHasher.Hash(passwordVO.Value())
	if err != nil {
		s.logger.Error(ctx, "Failed to hash password", sharedports.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create domain entity with hashed password
	hashedPassword := domain.NewHashedPassword(hash)
	identity, err := domain.NewLocalIdentity(userID, userUID, hashedPassword)
	if err != nil {
		s.logger.Error(ctx, "Failed to create local identity", sharedports.Error(err))
		return nil, err
	}

	// Persist identity
	if err := s.repo.Create(ctx, identity); err != nil {
		s.logger.Error(ctx, "Failed to persist local identity", sharedports.Error(err))
		return nil, fmt.Errorf("failed to create identity: %w", err)
	}

	s.logger.Debug(ctx, "Local identity created successfully",
		sharedports.Int64("user_id", userID),
		sharedports.String("identity_id", identity.UID()))
	return identity, nil
}

// VerifyLocalCredentials verifies a user's password for local authentication.
// Returns nil if credentials are valid, error otherwise.
func (s *IdentityService) VerifyLocalCredentials(ctx context.Context, userID int64, password string) error {
	s.logger.Debug(ctx, "Verifying local credentials", sharedports.Int64("user_id", userID))

	// Find local identity for user
	identity, err := s.repo.FindByProviderAndUserID(ctx, domain.ProviderLocal, userID)
	if err != nil {
		s.logger.Warn(ctx, "Identity not found for verification", sharedports.Error(err))
		return domain.ErrInvalidCredentials
	}

	// Verify password using domain method
	if err := identity.VerifyPassword(s.passwordHasher, password); err != nil {
		s.logger.Warn(ctx, "Password verification failed", sharedports.Error(err))
		return domain.ErrInvalidCredentials
	}

	s.logger.Debug(ctx, "Credentials verified successfully", sharedports.Int64("user_id", userID))
	return nil
}

// UpdatePassword updates the password for a local identity.
// Validates old password before setting new one.
func (s *IdentityService) UpdatePassword(ctx context.Context, identityID string, oldPassword, newPassword string) error {
	s.logger.Debug(ctx, "Updating password", sharedports.String("identity_id", identityID))

	// Find identity by UID (external ID)
	// Note: Need to add FindByUID method to repository
	// For now, this will return an error until repository is updated
	return errors.NewNotFoundError("update password not yet fully implemented - requires repository enhancement")
}

// CreateSSOIdentity creates a new SSO authentication identity.
// Validates provider and provider ID, then persists the identity.
func (s *IdentityService) CreateSSOIdentity(ctx context.Context, userID int64, userUID string, provider domain.Provider, providerID string) (*domain.Identity, error) {
	s.logger.Debug(ctx, "Creating SSO identity",
		sharedports.Int64("user_id", userID),
		sharedports.String("provider", string(provider)))

	// Create domain entity (includes provider validation)
	identity, err := domain.NewSSOIdentity(userID, userUID, provider, providerID)
	if err != nil {
		s.logger.Error(ctx, "Failed to create SSO identity", sharedports.Error(err))
		return nil, err
	}

	// Persist identity
	if err := s.repo.Create(ctx, identity); err != nil {
		s.logger.Error(ctx, "Failed to persist SSO identity", sharedports.Error(err))
		return nil, fmt.Errorf("failed to create SSO identity: %w", err)
	}

	s.logger.Debug(ctx, "SSO identity created successfully",
		sharedports.Int64("user_id", userID),
		sharedports.String("identity_id", identity.UID()),
		sharedports.String("provider", string(provider)))
	return identity, nil
}

// FindByProviderAndUserID finds an identity by provider and user ID.
func (s *IdentityService) FindByProviderAndUserID(ctx context.Context, provider domain.Provider, userID int64) (*domain.Identity, error) {
	s.logger.Debug(ctx, "Finding identity by provider and user ID",
		sharedports.String("provider", string(provider)),
		sharedports.Int64("user_id", userID))

	identity, err := s.repo.FindByProviderAndUserID(ctx, provider, userID)
	if err != nil {
		s.logger.Warn(ctx, "Identity not found", sharedports.Error(err))
		return nil, err
	}

	return identity, nil
}

// FindByProviderAndProviderID finds an identity by provider and provider ID.
// Used for SSO login to find existing identities by external provider ID.
func (s *IdentityService) FindByProviderAndProviderID(ctx context.Context, provider domain.Provider, providerID string) (*domain.Identity, error) {
	s.logger.Debug(ctx, "Finding identity by provider and provider ID",
		sharedports.String("provider", string(provider)),
		sharedports.String("provider_id", providerID))

	identity, err := s.repo.FindByProviderAndProviderID(ctx, provider, providerID)
	if err != nil {
		s.logger.Warn(ctx, "Identity not found", sharedports.Error(err))
		return nil, err
	}

	return identity, nil
}

// GetIdentitiesByUserID retrieves all identities for a user.
// Returns empty slice if user has no identities.
func (s *IdentityService) GetIdentitiesByUserID(ctx context.Context, userID int64) ([]*domain.Identity, error) {
	s.logger.Debug(ctx, "Getting identities for user", sharedports.Int64("user_id", userID))

	// Note: This requires a new repository method that doesn't exist yet
	// For now, return an error indicating incomplete implementation
	return nil, errors.NewNotFoundError("get identities by user not yet implemented - requires repository enhancement")
}

// DeleteIdentity soft-deletes an identity.
func (s *IdentityService) DeleteIdentity(ctx context.Context, id string, actor string) error {
	s.logger.Debug(ctx, "Deleting identity", sharedports.String("identity_id", id))

	if err := s.repo.Delete(ctx, id, actor); err != nil {
		s.logger.Error(ctx, "Failed to delete identity", sharedports.Error(err))
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	s.logger.Debug(ctx, "Identity deleted successfully", sharedports.String("identity_id", id))
	return nil
}

// Ensure IdentityService implements Service interface
var _ ports.Service = (*IdentityService)(nil)
