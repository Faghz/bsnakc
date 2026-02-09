package http

import (
	"github.com/elzestia/go-boilerplate/internal/auth/domain"
)

// ToAuthResponse converts application.AuthResponse to DTO AuthResponse
func ToAuthResponse(authResp *domain.AuthResponse) *AuthResponse {
	return &AuthResponse{
		User: UserResponse{
			ID:        authResp.UserID,
			Email:     authResp.Email,
			Name:      authResp.Name,
			Username:  authResp.Username,
			AvatarURL: authResp.AvatarURL,
		},
		AccessToken:      authResp.AccessToken,
		ExpiresAt:        authResp.AccessTokenExpiry.Format("2006-01-02T15:04:05Z07:00"),
		RefreshToken:     authResp.RefreshToken,
		RefreshExpiresAt: authResp.RefreshTokenExpiry.Format("2006-01-02T15:04:05Z07:00"),
	}
}
