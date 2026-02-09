package application

import (
	"context"
	"fmt"

	"github.com/elzestia/go-boilerplate/internal/auth/domain"
	sharedports "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
)

// ValidateToken validates an access token and returns the user ID
func (s *AuthService) ValidateToken(ctx context.Context, token string) (string, error) {
	return s.tokenService.ValidateAccessToken(token)
}

// RefreshTokens validates a refresh token and issues a new token pair
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	s.logger.Debug(ctx, "Refreshing tokens")

	// Validate refresh token
	userID, tokenID, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.Warn(ctx, "Refresh token validation failed", sharedports.Error(err))
		return nil, err
	}

	// Revoke old refresh token (rotation)
	if err := s.refreshTokenRepo.Delete(ctx, tokenID); err != nil {
		s.logger.Error(ctx, "Failed to revoke old refresh token", sharedports.Error(err))
		// Continue anyway - better to issue new tokens than fail
	}

	// Get user details via user service to get both internal and external IDs
	userInfo, err := s.userService.GetByUID(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "Failed to find user", sharedports.Error(err), sharedports.String("user_id", userID))
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Generate new token pair
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, userInfo.ID, userInfo.UID)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate token pair", sharedports.Error(err))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.logger.Debug(ctx, "Tokens refreshed successfully", sharedports.String("user_id", userID))

	return &domain.AuthResponse{
		UserID:             userInfo.UID,
		Email:              userInfo.Email,
		Name:               userInfo.Name,
		Username:           userInfo.Username,
		AvatarURL:          userInfo.AvatarURL,
		AccessToken:        tokenPair.AccessToken,
		AccessTokenExpiry:  tokenPair.AccessTokenExpiry,
		RefreshToken:       tokenPair.RefreshToken,
		RefreshTokenExpiry: tokenPair.RefreshTokenExpiry,
	}, nil
}
