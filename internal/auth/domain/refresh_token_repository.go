package domain

import (
	"context"
	"time"
)

// RefreshTokenRepository defines the interface for refresh token persistence
type RefreshTokenRepository interface {
	// Store stores a refresh token with TTL
	Store(ctx context.Context, tokenID, userID string, ttl time.Duration) error

	// FindByID retrieves a refresh token by ID
	FindByID(ctx context.Context, tokenID string) (userID string, err error)

	// Delete removes a refresh token (revocation)
	Delete(ctx context.Context, tokenID string) error

	// DeleteAllForUser revokes all refresh tokens for a user
	DeleteAllForUser(ctx context.Context, userID string) error
}
