package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/elzestia/go-boilerplate/internal/shared/config"
)

// Ensure DiscordProvider implements Provider interface
var _ Provider = (*DiscordProvider)(nil)

// DiscordProvider handles Discord OAuth authentication
type DiscordProvider struct {
	config config.DiscordConfig
	client *http.Client
}

// NewDiscordProvider creates a new Discord OAuth provider
func NewDiscordProvider(cfg config.DiscordConfig) *DiscordProvider {
	return &DiscordProvider{
		config: cfg,
		client: &http.Client{},
	}
}

// DiscordUser represents the Discord user data
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
}

// GetAuthURL returns the Discord OAuth authorization URL
func (p *DiscordProvider) GetAuthURL(state string) string {
	params := url.Values{}

	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "identify email")
	params.Set("state", state)

	return fmt.Sprintf("https://discord.com/api/oauth2/authorize?%s", params.Encode())
}

// ExchangeCode exchanges authorization code for access token
func (p *DiscordProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("discord returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.AccessToken, nil
}

// GetUser retrieves the Discord user information using access token
func (p *DiscordProvider) GetUser(ctx context.Context, accessToken string) (*OAuthUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord returned status %d: %s", resp.StatusCode, string(body))
	}

	var discordUser DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	// Convert Discord user to generic OAuth user
	avatarURL := ""
	if discordUser.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)
	}

	return &OAuthUser{
		ID:        discordUser.ID,
		Email:     discordUser.Email,
		Name:      discordUser.Username,
		AvatarURL: avatarURL,
	}, nil
}
