package ports

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/auth/domain"
)

// Service is the input port for authentication operations (consumed by HTTP adapters)
type Service interface {
	// RegisterLocal registers a new user with email and password
	RegisterLocal(ctx context.Context, email, password, name, username string) (*domain.AuthResponse, error)

	// LoginLocal authenticates a user with email and password
	LoginLocal(ctx context.Context, email, password string) (*domain.AuthResponse, error)

	// ValidateToken validates an access token and returns the user ID
	ValidateToken(ctx context.Context, token string) (userID string, err error)

	// RefreshTokens validates a refresh token and issues a new token pair
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)

	// Logout revokes a user's refresh token
	Logout(ctx context.Context, refreshToken string) error

	// LogoutAllDevices revokes all refresh tokens for a user
	LogoutAllDevices(ctx context.Context, userID string) error

	// DiscordSSO handles Discord SSO login/registration
	DiscordSSO(ctx context.Context, code string) (*domain.AuthResponse, error)
}
