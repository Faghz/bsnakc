package application

import (
	"context"
	"fmt"

	sharedports "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
)

// Logout revokes a user's refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	s.logger.Debug(ctx, "Logging out user")

	// Validate and extract token ID
	_, tokenID, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.Warn(ctx, "Invalid refresh token during logout", sharedports.Error(err))
		// Don't fail - token might already be expired/revoked
		return nil
	}

	// Revoke refresh token
	if err := s.refreshTokenRepo.Delete(ctx, tokenID); err != nil {
		s.logger.Error(ctx, "Failed to revoke refresh token", sharedports.Error(err))
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	s.logger.Debug(ctx, "User logged out successfully")
	return nil
}

// LogoutAllDevices revokes all refresh tokens for a user
func (s *AuthService) LogoutAllDevices(ctx context.Context, userID string) error {
	s.logger.Debug(ctx, "Logging out all devices", sharedports.String("user_id", userID))

	if err := s.refreshTokenRepo.DeleteAllForUser(ctx, userID); err != nil {
		s.logger.Error(ctx, "Failed to revoke all tokens", sharedports.Error(err))
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}

	s.logger.Debug(ctx, "All devices logged out successfully", sharedports.String("user_id", userID))
	return nil
}

// generateUsernameFromName creates a username from a Discord display name
// Converts to lowercase, replaces spaces with underscores, removes invalid chars
func generateUsernameFromName(name string) string {
	// TODO: Add proper sanitization logic
	// For now, just use a simple conversion
	username := name
	if len(username) > 32 {
		username = username[:32]
	}
	if len(username) < 3 {
		username = username + "123"
	}
	return username
}
