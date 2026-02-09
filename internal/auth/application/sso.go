package application

import (
	"context"

	"github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	"github.com/elzestia/go-boilerplate/internal/auth/domain"
	identitydomain "github.com/elzestia/go-boilerplate/internal/identities/domain"
	sharedports "github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/errors"
)

// LoginSSO authenticates or registers a user via SSO provider using transactional operations
func (s *AuthService) loginSSO(
	ctx context.Context,
	oauthData *domain.UserOauthData,
) (*domain.AuthResponse, error) {
	s.logger.Debug(ctx, "Starting SSO login", sharedports.String("provider", oauthData.Provider))

	// Validate and convert provider
	providerEnum := identitydomain.Provider(oauthData.Provider)
	if err := identitydomain.ValidateProvider(providerEnum); err != nil {
		s.logger.Error(ctx, "Invalid SSO provider", sharedports.String("provider", oauthData.Provider))
		return nil, err
	}

	// Check if identity exists via identity service
	identityInfo, err := s.identityService.FindByProviderAndProviderID(ctx, providerEnum, oauthData.ProviderID)
	if err != nil && !errors.IsNotFoundError(err) {
		s.logger.Error(ctx, "Failed to find SSO identity", sharedports.Error(err))
		return nil, err
	}

	var userInfo *ports.UserInfo

	// If identity exists, get the user
	if identityInfo != nil {
		// User already registered with this SSO provider - just login
		userInfo, err = s.userService.GetUserByID(ctx, identityInfo.UserID)
		if err != nil {
			s.logger.Error(ctx, "Failed to find user for existing identity", sharedports.Error(err))
			return nil, err
		}
	} else {
		// Identity doesn't exist - check if user exists by email
		s.logger.Debug(ctx, "SSO identity not found, checking for existing user", sharedports.String("provider", oauthData.Provider))

		existingUser, err := s.userService.GetUserByEmail(ctx, oauthData.Email)
		if err != nil && !errors.IsNotFoundError(err) {
			s.logger.Error(ctx, "Failed to find user by email during SSO login", sharedports.Error(err))
			return nil, err
		}

		if existingUser != nil {
			// User exists - link SSO identity to existing user
			s.logger.Debug(ctx, "Linking SSO identity to existing user",
				sharedports.Int64("user_id", existingUser.ID),
				sharedports.String("provider", oauthData.Provider))

			identityInfo, err = s.identityService.CreateSSOIdentity(ctx, &ports.IdentityCreateSSORequest{
				UserID:     existingUser.ID,
				UserUID:    existingUser.UID,
				Provider:   providerEnum,
				ProviderID: oauthData.ProviderID,
			})
			if err != nil {
				s.logger.Error(ctx, "Failed to link SSO identity", sharedports.Error(err))
				return nil, err
			}

			userInfo = existingUser
		} else {
			// User doesn't exist - create new user and identity atomically
			userName := oauthData.Username
			if userName == "" {
				userName = generateUsernameFromName(oauthData.Name)
			}

			// EXCEPTION: Create both user AND identity in same transaction for atomicity
			// This is critical - a user without an identity cannot authenticate
			err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
				// Create user via user service (using transaction context)
				userInfo, err = s.userService.CreateUser(txCtx, oauthData.Email, oauthData.Name, userName)
				if err != nil {
					return err
				}

				// Create SSO identity via service (which uses the transaction context)
				identityInfo, err = s.identityService.CreateSSOIdentity(txCtx, &ports.IdentityCreateSSORequest{
					UserID:     userInfo.ID,
					UserUID:    userInfo.UID,
					Provider:   providerEnum,
					ProviderID: oauthData.ProviderID,
				})
				if err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				s.logger.Error(ctx, "Failed to create user and identity during SSO registration", sharedports.Error(err))
				return nil, err
			}

			s.logger.Debug(ctx, "User auto-registered via SSO",
				sharedports.Int64("user_id", userInfo.ID),
				sharedports.String("identity_id", identityInfo.UID),
				sharedports.String("provider", oauthData.Provider))
		}
	}

	// Generate token pair
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, userInfo.ID, userInfo.UID)
	if err != nil {
		s.logger.Error(ctx, "Failed to generate token pair", sharedports.Error(err))
		return nil, err
	}

	s.logger.Debug(ctx, "User logged in successfully via SSO",
		sharedports.Int64("user_id", userInfo.ID),
		sharedports.String("provider", oauthData.Provider))

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

func (s *AuthService) DiscordSSO(ctx context.Context, code string) (*domain.AuthResponse, error) {
	// Exchange code for token
	token, err := s.discordProvider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, errors.NewAuthenticationError("failed to exchange code for token")
	}

	// Get user info
	userInfo, err := s.discordProvider.GetUser(ctx, token)
	if err != nil {
		return nil, errors.NewAuthenticationError("failed to get user info from Discord")
	}

	oauthData := &domain.UserOauthData{
		Provider:   string(identitydomain.ProviderDiscord),
		ProviderID: userInfo.ID,
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		AvatarURL:  userInfo.AvatarURL,
	}

	authData, err := s.loginSSO(ctx, oauthData)
	if err != nil {
		return nil, err
	}

	return authData, nil
}
