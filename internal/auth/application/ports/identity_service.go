package ports

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/identities/domain"
)

// IdentityService is the Auth context's view of identity operations.
// This is an Anti-Corruption Layer (ACL) port interface.
// It shields Auth from depending on Identities domain entities.
type IdentityService interface {
	// Local identity operations
	CreateLocalIdentity(ctx context.Context, req *IdentityCreateLocalRequest) (*IdentityInfo, error)
	VerifyLocalCredentials(ctx context.Context, userID int64, password string) error

	// SSO identity operations
	CreateSSOIdentity(ctx context.Context, req *IdentityCreateSSORequest) (*IdentityInfo, error)
	FindByProviderAndUserID(ctx context.Context, provider domain.Provider, userID int64) (*IdentityInfo, error)
	FindByProviderAndProviderID(ctx context.Context, provider domain.Provider, providerID string) (*IdentityInfo, error)
}
