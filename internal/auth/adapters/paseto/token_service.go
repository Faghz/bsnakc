package paseto

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/elzestia/go-boilerplate/internal/auth/application/ports"
	authdomain "github.com/elzestia/go-boilerplate/internal/auth/domain"
)

const (
	claimUserID  = "user_id"
	claimTokenID = "token_id"

	footerTypeAccess  = "access"
	footerTypeRefresh = "refresh"
)

// TokenService implements ports.TokenService using PASETO v4.local
type TokenService struct {
	symmetricKey         paseto.V4SymmetricKey
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	refreshTokenRepo     authdomain.RefreshTokenRepository
}

// NewTokenService creates a new PASETO token service
// keyHex should be a 64-character hex string (32 bytes)
func NewTokenService(
	keyHex string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	refreshTokenRepo authdomain.RefreshTokenRepository,
) (*TokenService, error) {
	// Decode hex key to bytes
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid PASETO key format (must be hex): %w", err)
	}

	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("invalid PASETO key length: expected 32 bytes, got %d", len(keyBytes))
	}

	// Create PASETO v4 symmetric key
	symmetricKey, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create PASETO key: %w", err)
	}

	return &TokenService{
		symmetricKey:         symmetricKey,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		refreshTokenRepo:     refreshTokenRepo,
	}, nil
}

// GenerateTokenPair creates a new access + refresh token pair
func (s *TokenService) GenerateTokenPair(ctx context.Context, userInternalID int64, userUID string) (*ports.TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessToken := paseto.NewToken()
	accessToken.SetIssuedAt(now)
	accessToken.SetExpiration(now.Add(s.accessTokenDuration))
	accessToken.SetString(claimUserID, userUID)

	accessTokenString := accessToken.V4Encrypt(s.symmetricKey, []byte(footerTypeAccess))

	// Generate refresh token with unique token ID
	// TODO: Extract ipAddress and userAgent from request context
	refreshTokenEntity := authdomain.NewRefreshToken(userInternalID, userUID, s.refreshTokenDuration, "", "")

	refreshToken := paseto.NewToken()
	refreshToken.SetIssuedAt(now)
	refreshToken.SetExpiration(refreshTokenEntity.ExpiresAt())
	refreshToken.SetString(claimUserID, userUID)
	refreshToken.SetString(claimTokenID, refreshTokenEntity.ID())

	refreshTokenString := refreshToken.V4Encrypt(s.symmetricKey, []byte(footerTypeRefresh))

	// Store refresh token in database with TTL
	if err := s.refreshTokenRepo.Store(ctx, refreshTokenEntity.ID(), userUID, s.refreshTokenDuration); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &ports.TokenPair{
		AccessToken:        accessTokenString,
		AccessTokenExpiry:  now.Add(s.accessTokenDuration),
		RefreshToken:       refreshTokenString,
		RefreshTokenExpiry: refreshTokenEntity.ExpiresAt(),
	}, nil
}

// ValidateAccessToken validates an access token and returns the user ID
func (s *TokenService) ValidateAccessToken(tokenString string) (string, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())

	token, err := parser.ParseV4Local(s.symmetricKey, tokenString, []byte(footerTypeAccess))
	if err != nil {
		return "", authdomain.ErrInvalidToken
	}

	userID, err := token.GetString(claimUserID)
	if err != nil {
		return "", authdomain.ErrInvalidToken
	}

	return userID, nil
}

// ValidateRefreshToken validates a refresh token and returns user ID + token ID
func (s *TokenService) ValidateRefreshToken(tokenString string) (string, string, error) {
	parser := paseto.NewParser()
	parser.AddRule(paseto.NotExpired())

	token, err := parser.ParseV4Local(s.symmetricKey, tokenString, []byte(footerTypeRefresh))
	if err != nil {
		return "", "", authdomain.ErrInvalidToken
	}

	userID, err := token.GetString(claimUserID)
	if err != nil {
		return "", "", authdomain.ErrInvalidToken
	}

	tokenID, err := token.GetString(claimTokenID)
	if err != nil {
		return "", "", authdomain.ErrInvalidToken
	}

	// Check if token exists in Redis (not revoked)
	storedUserID, err := s.refreshTokenRepo.FindByID(context.Background(), tokenID)
	if err != nil {
		return "", "", authdomain.ErrRefreshTokenRevoked
	}

	// Verify user ID matches
	if storedUserID != userID {
		return "", "", authdomain.ErrInvalidToken
	}

	return userID, tokenID, nil
}

// Ensure TokenService implements ports.TokenService
var _ ports.TokenService = (*TokenService)(nil)
