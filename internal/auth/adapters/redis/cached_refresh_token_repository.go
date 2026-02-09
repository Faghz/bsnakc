package redis

import (
	"context"
	"time"

	authdomain "github.com/elzestia/go-boilerplate/internal/auth/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenKeyPrefix = "refresh_token:"
	userTokensKeyPrefix   = "user_tokens:"
)

// CachedRefreshTokenRepository wraps a primary repository with Redis caching
// Database is the source of truth, Redis is used for fast lookups
type CachedRefreshTokenRepository struct {
	primary authdomain.RefreshTokenRepository // Database (source of truth)
	redis   *redis.Client                     // Cache layer
	logger  ports.Logger
}

// NewCachedRefreshTokenRepository creates a cached repository
func NewCachedRefreshTokenRepository(
	primary authdomain.RefreshTokenRepository,
	redis *redis.Client,
	logger ports.Logger,
) *CachedRefreshTokenRepository {
	return &CachedRefreshTokenRepository{
		primary: primary,
		redis:   redis,
		logger:  logger.With(ports.String("repository", "cached_refresh_tokens")),
	}
}

// Store stores a refresh token in both database and Redis cache
func (r *CachedRefreshTokenRepository) Store(ctx context.Context, tokenID, userID string, ttl time.Duration) error {
	// Store in database (source of truth) first
	if err := r.primary.Store(ctx, tokenID, userID, ttl); err != nil {
		return err
	}

	// Cache in Redis for fast lookups (best effort)
	if err := r.cacheToken(ctx, tokenID, userID, ttl); err != nil {
		r.logger.Warn(ctx, "Failed to cache refresh token in Redis",
			ports.String("token_id", tokenID),
			ports.Error(err))
		// Don't fail - database write succeeded
	}

	return nil
}

// FindByID checks Redis cache first, then falls back to database
func (r *CachedRefreshTokenRepository) FindByID(ctx context.Context, tokenID string) (string, error) {
	// Try cache first (fast path)
	userID, err := r.getCachedToken(ctx, tokenID)
	if err == nil {
		r.logger.Debug(ctx, "Refresh token found in cache", ports.String("token_id", tokenID))
		return userID, nil
	}

	// Cache miss or error - fall back to database
	r.logger.Debug(ctx, "Cache miss, querying database", ports.String("token_id", tokenID))
	userID, err = r.primary.FindByID(ctx, tokenID)
	if err != nil {
		return "", err
	}

	// Warm the cache for next time (best effort)
	// Use shorter TTL since we don't know the original expiry
	go func() {
		cacheCtx := context.Background()
		if err := r.cacheToken(cacheCtx, tokenID, userID, 15*time.Minute); err != nil {
			r.logger.Debug(cacheCtx, "Failed to warm cache after database lookup", ports.Error(err))
		}
	}()

	return userID, nil
}

// Delete removes from both cache and database
func (r *CachedRefreshTokenRepository) Delete(ctx context.Context, tokenID string) error {
	// Delete from database first
	if err := r.primary.Delete(ctx, tokenID); err != nil {
		return err
	}

	// Invalidate cache (best effort)
	if err := r.invalidateToken(ctx, tokenID); err != nil {
		r.logger.Warn(ctx, "Failed to invalidate cache", ports.String("token_id", tokenID), ports.Error(err))
		// Don't fail - database delete succeeded
	}

	return nil
}

// DeleteAllForUser revokes all tokens in database and invalidates cache
func (r *CachedRefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	// Delete from database first
	if err := r.primary.DeleteAllForUser(ctx, userID); err != nil {
		return err
	}

	// Invalidate user's cached tokens (best effort)
	if err := r.invalidateUserTokens(ctx, userID); err != nil {
		r.logger.Warn(ctx, "Failed to invalidate user token cache",
			ports.String("user_id", userID),
			ports.Error(err))
		// Don't fail - database delete succeeded
	}

	return nil
}

// cacheToken stores a token in Redis
func (r *CachedRefreshTokenRepository) cacheToken(ctx context.Context, tokenID, userID string, ttl time.Duration) error {
	key := refreshTokenKeyPrefix + tokenID
	userKey := userTokensKeyPrefix + userID

	pipe := r.redis.Pipeline()

	// Store token -> userID mapping
	pipe.Set(ctx, key, userID, ttl)

	// Add token to user's token set
	pipe.SAdd(ctx, userKey, tokenID)
	pipe.Expire(ctx, userKey, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

// getCachedToken retrieves a token from Redis cache
func (r *CachedRefreshTokenRepository) getCachedToken(ctx context.Context, tokenID string) (string, error) {
	key := refreshTokenKeyPrefix + tokenID

	userID, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", authdomain.ErrRefreshTokenNotFound
	}
	if err != nil {
		return "", err
	}

	return userID, nil
}

// invalidateToken removes a token from Redis cache
func (r *CachedRefreshTokenRepository) invalidateToken(ctx context.Context, tokenID string) error {
	key := refreshTokenKeyPrefix + tokenID

	// We don't know the userID from cache, so we can't remove from user's set
	// This is acceptable - the set will expire naturally with TTL
	return r.redis.Del(ctx, key).Err()
}

// invalidateUserTokens removes all of a user's tokens from cache
func (r *CachedRefreshTokenRepository) invalidateUserTokens(ctx context.Context, userID string) error {
	userKey := userTokensKeyPrefix + userID

	// Get all token IDs for this user
	tokenIDs, err := r.redis.SMembers(ctx, userKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	if len(tokenIDs) == 0 {
		return nil
	}

	// Build all keys to delete
	keys := make([]string, 0, len(tokenIDs)+1)
	for _, tokenID := range tokenIDs {
		keys = append(keys, refreshTokenKeyPrefix+tokenID)
	}
	keys = append(keys, userKey)

	// Delete all keys
	return r.redis.Del(ctx, keys...).Err()
}

// Ensure CachedRefreshTokenRepository implements authdomain.RefreshTokenRepository
var _ authdomain.RefreshTokenRepository = (*CachedRefreshTokenRepository)(nil)
