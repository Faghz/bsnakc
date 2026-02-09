package domain

import (
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/types"
	"github.com/google/uuid"
)

// RefreshToken represents a refresh token entity
type RefreshToken struct {
	uid        string // UUID for API exposure and database primary key
	userID     int64  // Internal user ID (BIGSERIAL FK) for database operations
	userUID    string // External user UUID for API and token claims
	expiresAt  time.Time
	createdAt  time.Time
	revokedAt  types.NullTime
	lastUsedAt types.NullTime // Track when token was last used
	ipAddress  string         // IP address from where token was created
	userAgent  string         // User agent string
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(userInternalID int64, userUID string, duration time.Duration, ipAddress, userAgent string) *RefreshToken {
	now := time.Now()
	uid, err := uuid.NewV7()
	if err != nil {
		// UUID V7 generation should rarely fail, but if it does, panic as we can't proceed without a unique ID
		panic("failed to generate UUID v7: " + err.Error())
	}
	return &RefreshToken{
		uid:       uid.String(), // Generated UUID for primary key
		userID:    userInternalID,
		userUID:   userUID,
		expiresAt: now.Add(duration),
		createdAt: now,
		ipAddress: ipAddress,
		userAgent: userAgent,
	}
}

// ReconstructRefreshToken reconstructs a refresh token from persistence
func ReconstructRefreshToken(
	uid string,
	userInternalID int64,
	userUID string,
	expiresAt, createdAt time.Time,
	revokedAt, lastUsedAt types.NullTime,
	ipAddress, userAgent string,
) *RefreshToken {
	return &RefreshToken{
		uid:        uid,
		userID:     userInternalID,
		userUID:    userUID,
		expiresAt:  expiresAt,
		createdAt:  createdAt,
		revokedAt:  revokedAt,
		lastUsedAt: lastUsedAt,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
	}
}

// ID returns the token ID (external UUID)
func (t *RefreshToken) ID() string { return t.uid }

// UID returns the external UUID
func (t *RefreshToken) UID() string { return t.uid }

// UserID returns the user ID (external UUID for API)
func (t *RefreshToken) UserID() string { return t.userUID }

// InternalUserID returns the internal user ID
func (t *RefreshToken) InternalUserID() int64 { return t.userID }

// ExpiresAt returns expiration time
func (t *RefreshToken) ExpiresAt() time.Time { return t.expiresAt }

// CreatedAt returns creation time
func (t *RefreshToken) CreatedAt() time.Time { return t.createdAt }

// RevokedAt returns revocation time
func (t *RefreshToken) RevokedAt() types.NullTime { return t.revokedAt }

// LastUsedAt returns last used time
func (t *RefreshToken) LastUsedAt() types.NullTime { return t.lastUsedAt }

// IPAddress returns the IP address
func (t *RefreshToken) IPAddress() string { return t.ipAddress }

// UserAgent returns the user agent
func (t *RefreshToken) UserAgent() string { return t.userAgent }

// IsExpired checks if the token is expired
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.expiresAt)
}

// IsRevoked checks if the token is revoked
func (t *RefreshToken) IsRevoked() bool {
	return t.revokedAt.Valid
}

// Revoke marks the token as revoked
func (t *RefreshToken) Revoke() {
	t.revokedAt = types.NewNullTime(time.Now())
}

// UpdateLastUsed updates the last used timestamp
func (t *RefreshToken) UpdateLastUsed() {
	t.lastUsedAt = types.NewNullTime(time.Now())
}
