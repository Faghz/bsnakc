package ports

import (
	"context"
	"time"
)

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken        string
	AccessTokenExpiry  time.Time
	RefreshToken       string
	RefreshTokenExpiry time.Time
}

// TokenService is the port for token operations (implemented by PASETO adapter)
type TokenService interface {
	// GenerateTokenPair creates a new access + refresh token pair for a user
	// Requires both internal ID (for DB FK) and external UID (for token claims)
	GenerateTokenPair(ctx context.Context, userInternalID int64, userUID string) (*TokenPair, error)

	// ValidateAccessToken validates an access token and returns the user ID
	ValidateAccessToken(token string) (userID string, err error)

	// ValidateRefreshToken validates a refresh token and returns user ID + token ID
	ValidateRefreshToken(token string) (userID string, tokenID string, err error)
}
