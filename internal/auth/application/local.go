package application

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	"github.com/elzestia/go-boilerplate/internal/auth/domain"
	authdomain "github.com/elzestia/go-boilerplate/internal/auth/domain"
	sharedports "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

// RegisterLocal registers a new user with email and password using transactional operations
func (s *AuthService) RegisterLocal(ctx context.Context, email, password, name, username string) (*domain.AuthResponse, error) {
	s.logger.Debug(ctx, "Starting local registration", sharedports.String("username", username))

	var userInfo *ports.UserInfo
	var identityInfo *ports.IdentityInfo
	var err error

	// A user without an identity cannot authenticate - these must be atomic
	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Create user via user service (using transaction context)
		userInfo, err = s.userService.CreateUser(txCtx, email, name, username)
		if err != nil {
			return err
		}

		// Create identity via identity service (using transaction context)
		identityInfo, err = s.identityService.CreateLocalIdentity(txCtx, &ports.IdentityCreateLocalRequest{
			UserID:   userInfo.ID,
			UserUID:  userInfo.UID,
			Password: password,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error(ctx, "Registration transaction failed", sharedports.Error(err))
		return nil, err
	}

	// Generate token pair (outside transaction)
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, userInfo.ID, userInfo.UID)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate token pair", sharedports.Error(err))
		return nil, errors.NewInternalError("failed to generate token pair", err)
	}

	s.logger.Info(ctx, "User registered successfully",
		sharedports.Int64("user_id", userInfo.ID),
		sharedports.String("identity_id", identityInfo.UID),
		sharedports.String("provider", "local"))

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

// LoginLocal authenticates a user with email and password
func (s *AuthService) LoginLocal(ctx context.Context, email, password string) (*domain.AuthResponse, error) {
	s.logger.Debug(ctx, "Starting local login")

	// Find user via user service (ACL)
	userInfo, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Debug(ctx, "Login failed: user not found", sharedports.String("email", email))
		return nil, authdomain.ErrInvalidCredentials
	}

	// Verify credentials via identity service (ACL)
	err = s.identityService.VerifyLocalCredentials(ctx, userInfo.ID, password)
	if err != nil {
		s.logger.Debug(ctx, "Login failed: invalid credentials", sharedports.Int64("user_id", userInfo.ID))
		return nil, err // Error already translated by ACL adapter
	}

	// Generate token pair
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, userInfo.ID, userInfo.UID)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate token pair", sharedports.Error(err))
		return nil, errors.NewInternalError("failed to generate token pair", err)
	}

	s.logger.Debug(ctx, "User logged in successfully",
		sharedports.Int64("user_id", userInfo.ID),
		sharedports.String("provider", "local"))

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
