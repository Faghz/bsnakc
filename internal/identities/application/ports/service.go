package ports

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/identities/domain"
)

// Service is the input port for identity operations.
// This interface defines what the identities context exposes to external consumers.
type Service interface {
	// Local identity operations
	CreateLocalIdentity(ctx context.Context, userID int64, userUID, password string) (*domain.Identity, error)
	VerifyLocalCredentials(ctx context.Context, userID int64, password string) error
	UpdatePassword(ctx context.Context, identityID string, oldPassword, newPassword string) error

	// SSO identity operations
	CreateSSOIdentity(ctx context.Context, userID int64, userUID string, provider domain.Provider, providerID string) (*domain.Identity, error)
	FindByProviderAndUserID(ctx context.Context, provider domain.Provider, userID int64) (*domain.Identity, error)
	FindByProviderAndProviderID(ctx context.Context, provider domain.Provider, providerID string) (*domain.Identity, error)

	// Management operations
	GetIdentitiesByUserID(ctx context.Context, userID int64) ([]*domain.Identity, error)
	DeleteIdentity(ctx context.Context, id, actor string) error
}
