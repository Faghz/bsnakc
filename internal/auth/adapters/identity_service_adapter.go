package adapters

import (
	"context"
	"errors"
	"strings"

	authports "github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	authdomain "github.com/elzestia/go-boilerplate/internal/auth/domain"
	identitiesapp "github.com/elzestia/go-boilerplate/internal/identities/application"
	identitydomain "github.com/elzestia/go-boilerplate/internal/identities/domain"
	sharederrors "github.com/elzestia/go-boilerplate/internal/shared/errors"
)

// IdentityServiceAdapter is an Anti-Corruption Layer that adapts the Identities context
// for use by the Auth context. It prevents Auth from depending on Identity domain entities.
//
// This adapter:
// - Converts Identity domain entities → IdentityInfo DTOs
// - Translates Identities domain errors → Auth domain errors
// - Provides Auth-specific view of identity operations
type IdentityServiceAdapter struct {
	identityService identitiesapp.Service
}

// NewIdentityServiceAdapter creates a new identity service adapter.
func NewIdentityServiceAdapter(identityService identitiesapp.Service) *IdentityServiceAdapter {
	return &IdentityServiceAdapter{
		identityService: identityService,
	}
}

// CreateLocalIdentity creates a local authentication identity.
func (a *IdentityServiceAdapter) CreateLocalIdentity(ctx context.Context, req *authports.IdentityCreateLocalRequest) (*authports.IdentityInfo, error) {
	identity, err := a.identityService.CreateLocalIdentity(ctx, req.UserID, req.UserUID, req.Password)
	if err != nil {
		return nil, a.translateError(err)
	}
	return toIdentityInfo(identity), nil
}

// VerifyLocalCredentials verifies a user's local password.
func (a *IdentityServiceAdapter) VerifyLocalCredentials(ctx context.Context, userID int64, password string) error {
	err := a.identityService.VerifyLocalCredentials(ctx, userID, password)
	if err != nil {
		return a.translateError(err)
	}
	return nil
}

// CreateSSOIdentity creates an SSO authentication identity.
func (a *IdentityServiceAdapter) CreateSSOIdentity(ctx context.Context, req *authports.IdentityCreateSSORequest) (*authports.IdentityInfo, error) {
	identity, err := a.identityService.CreateSSOIdentity(ctx, req.UserID, req.UserUID, req.Provider, req.ProviderID)
	if err != nil {
		return nil, a.translateError(err)
	}
	return toIdentityInfo(identity), nil
}

// FindByProviderAndUserID finds an identity by provider and user ID.
func (a *IdentityServiceAdapter) FindByProviderAndUserID(ctx context.Context, provider identitydomain.Provider, userID int64) (*authports.IdentityInfo, error) {
	identity, err := a.identityService.FindByProviderAndUserID(ctx, provider, userID)
	if err != nil {
		return nil, a.translateError(err)
	}
	return toIdentityInfo(identity), nil
}

// FindByProviderAndProviderID finds an identity by provider and provider ID.
func (a *IdentityServiceAdapter) FindByProviderAndProviderID(ctx context.Context, provider identitydomain.Provider, providerID string) (*authports.IdentityInfo, error) {
	identity, err := a.identityService.FindByProviderAndProviderID(ctx, provider, providerID)
	if err != nil {
		return nil, a.translateError(err)
	}
	return toIdentityInfo(identity), nil
}

// toIdentityInfo converts an Identity domain entity to IdentityInfo DTO.
// This prevents Auth from directly depending on Identity domain types.
func toIdentityInfo(identity *identitydomain.Identity) *authports.IdentityInfo {
	var providerID string
	if identity.ProviderID().Valid {
		providerID = identity.ProviderID().String
	}

	return &authports.IdentityInfo{
		ID:         identity.UID(),
		UID:        identity.UID(),
		UserID:     identity.InternalUserID(),
		Provider:   identity.Provider(),
		ProviderID: providerID,
		IsLocal:    identity.IsLocal(),
		IsSSO:      identity.IsSSO(),
		CreatedAt:  identity.CreatedAt(),
	}
}

// translateError maps Identities domain errors to Auth domain errors.
// This prevents error leakage across domain boundaries.
//
// Error translation rules:
// - identitydomain.ErrInvalidCredentials → authdomain.ErrInvalidCredentials
// - identitydomain.ErrWeakPassword → authdomain.ErrWeakPassword
// - identitydomain.ErrInvalidProvider → authdomain.ErrInvalidProvider
// - Validation errors → Pass through (shared errors)
// - Not found errors → authdomain.ErrIdentityNotFound
// - Others → Pass through
func (a *IdentityServiceAdapter) translateError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific Identity domain errors
	switch {
	case errors.Is(err, identitydomain.ErrInvalidCredentials):
		return authdomain.ErrInvalidCredentials

	case errors.Is(err, identitydomain.ErrWeakPassword):
		return authdomain.ErrWeakPassword

	case errors.Is(err, identitydomain.ErrInvalidProvider):
		return authdomain.ErrInvalidProvider

	case errors.Is(err, identitydomain.ErrIdentityAlreadyExists):
		return authdomain.ErrIdentityAlreadyExists

	// Check shared error types by category
	case sharederrors.IsNotFoundError(err):
		// Check if it's an identity-related not found error
		if strings.Contains(strings.ToLower(err.Error()), "identity") {
			return authdomain.ErrIdentityNotFound
		}
		return err // Pass through other not found errors

	default:
		// Pass through all other errors (validation, infrastructure, internal, etc.)
		return err
	}
}

// Ensure IdentityServiceAdapter implements ports.IdentityService
var _ authports.IdentityService = (*IdentityServiceAdapter)(nil)
