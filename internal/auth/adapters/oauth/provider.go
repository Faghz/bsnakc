package oauth

import "context"

// OAuthUser represents OAuth user data
type OAuthUser struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
}

// Provider defines the interface for OAuth providers
type Provider interface {
	// GetAuthURL returns the OAuth authorization URL
	GetAuthURL(state string) string

	// ExchangeCode exchanges an authorization code for an access token
	ExchangeCode(ctx context.Context, code string) (string, error)

	// GetUser retrieves user information using an access token
	GetUser(ctx context.Context, accessToken string) (*OAuthUser, error)
}
