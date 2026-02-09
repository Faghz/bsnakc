package ports

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/identities/domain"
)

// IdentityInfo is the ACL DTO for identity data (read-only view).
// This prevents the Auth context from directly depending on Identities domain entities.
// Auth never sees identities.Identity entity directly - only this DTO.
type IdentityInfo struct {
	ID         string          // External UUID (UID)
	UID        string          // External UUID (explicit)
	UserID     int64           // Internal user ID (for operations)
	Provider   domain.Provider // Provider type (local, discord)
	ProviderID string          // External provider ID (for SSO)
	IsLocal    bool            // True if local authentication
	IsSSO      bool            // True if SSO authentication
	CreatedAt  time.Time       // Creation timestamp
}

// IdentityCreateLocalRequest is the ACL DTO for creating local identities.
type IdentityCreateLocalRequest struct {
	UserID   int64  // Internal user ID
	UserUID  string // External user UUID
	Password string // Plain password (will be hashed by identity service)
}

// IdentityCreateSSORequest is the ACL DTO for creating SSO identities.
type IdentityCreateSSORequest struct {
	UserID     int64           // Internal user ID
	UserUID    string          // External user UUID
	Provider   domain.Provider // SSO provider (discord, etc.)
	ProviderID string          // External provider user ID
}
