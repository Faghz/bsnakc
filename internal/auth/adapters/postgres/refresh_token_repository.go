package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	authdomain "github.com/elzestia/go-boilerplate/internal/auth/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/database"
	apperrors "github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/elzestia/go-boilerplate/internal/shared/types"
	"github.com/jmoiron/sqlx"
)

// RefreshTokenRepository implements authdomain.RefreshTokenRepository using PostgreSQL
type RefreshTokenRepository struct {
	db     *sqlx.DB
	logger ports.Logger
}

// NewRefreshTokenRepository creates a new PostgreSQL refresh token repository
func NewRefreshTokenRepository(db *sqlx.DB, logger ports.Logger) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db:     db,
		logger: logger.With(ports.String("repository", "refresh_tokens")),
	}
}

// refreshTokenRow represents a refresh token row from the database
type refreshTokenRow struct {
	UID        string           `db:"uid"`
	UserID     int64            `db:"user_id"`
	UserUID    string           `db:"user_uid"`
	ExpiresAt  time.Time        `db:"expires_at"`
	CreatedAt  time.Time        `db:"created_at"`
	RevokedAt  types.NullTime   `db:"revoked_at"`
	LastUsedAt types.NullTime   `db:"last_used_at"`
	IPAddress  types.NullString `db:"ip_address"`
	UserAgent  types.NullString `db:"user_agent"`
}

// Store stores a refresh token in the database with TTL
func (r *RefreshTokenRepository) Store(ctx context.Context, tokenID, userID string, ttl time.Duration) error {
	r.logger.Debug(ctx, "Storing refresh token", ports.String("token_id", tokenID), ports.String("user_id", userID))

	// Lookup internal user ID by user UID
	var userInternalID int64
	userQuery := `SELECT id FROM users WHERE uid = $1 AND deleted_at IS NULL`
	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &userInternalID, userQuery, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NewNotFoundError("user")
		}
		return apperrors.NewInternalError("failed to lookup user", err)
	}

	expiresAt := time.Now().Add(ttl)

	query := `
		INSERT INTO refresh_tokens (uid, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err = q.ExecContext(ctx, query, tokenID, userInternalID, expiresAt, time.Now())
	if err != nil {
		return apperrors.NewInternalError("failed to store refresh token", err)
	}

	r.logger.Debug(ctx, "Refresh token stored successfully")
	return nil
}

// FindByID retrieves a refresh token by ID and validates it
func (r *RefreshTokenRepository) FindByID(ctx context.Context, tokenID string) (string, error) {
	r.logger.Debug(ctx, "Finding refresh token", ports.String("token_id", tokenID))

	query := `
		SELECT rt.uid, rt.user_id, u.uid as user_uid, rt.expires_at,
		       rt.created_at, rt.revoked_at, rt.last_used_at, rt.ip_address, rt.user_agent
		FROM refresh_tokens rt
		JOIN users u ON rt.user_id = u.id
		WHERE rt.uid = $1
	`

	var row refreshTokenRow
	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, tokenID)
	if err == sql.ErrNoRows {
		return "", authdomain.ErrRefreshTokenNotFound
	}
	if err != nil {
		return "", apperrors.NewInternalError("failed to find refresh token", err)
	}

	// Check if revoked
	if row.RevokedAt.Valid {
		r.logger.Warn(ctx, "Refresh token is revoked", ports.String("token_id", tokenID))
		return "", authdomain.ErrRefreshTokenRevoked
	}

	// Check if expired
	if time.Now().After(row.ExpiresAt) {
		r.logger.Warn(ctx, "Refresh token is expired", ports.String("token_id", tokenID))
		return "", authdomain.ErrTokenExpired
	}

	// Update last used timestamp
	go func() {
		updateCtx := context.Background()
		updateQuery := `UPDATE refresh_tokens SET last_used_at = $1 WHERE uid = $2`
		_, _ = r.db.ExecContext(updateCtx, updateQuery, time.Now(), tokenID)
	}()

	r.logger.Debug(ctx, "Refresh token found and valid")
	// Return user UID (external UUID) for token claims
	return row.UserUID, nil
}

// Delete removes a refresh token (revocation)
func (r *RefreshTokenRepository) Delete(ctx context.Context, tokenID string) error {
	r.logger.Debug(ctx, "Revoking refresh token", ports.String("token_id", tokenID))

	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE uid = $2 AND revoked_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query, time.Now(), tokenID)
	if err != nil {
		return apperrors.NewInternalError("failed to revoke refresh token", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewInternalError("failed to check revocation result", err)
	}

	if rowsAffected == 0 {
		r.logger.Debug(ctx, "Refresh token not found or already revoked")
		return nil // Idempotent - not an error if already revoked
	}

	r.logger.Debug(ctx, "Refresh token revoked successfully")
	return nil
}

// DeleteAllForUser revokes all refresh tokens for a user
func (r *RefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	r.logger.Debug(ctx, "Revoking all refresh tokens for user", ports.String("user_id", userID))

	// Lookup internal user ID by user UID
	var userInternalID int64
	userQuery := `SELECT id FROM users WHERE uid = $1 AND deleted_at IS NULL`
	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &userInternalID, userQuery, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NewNotFoundError("user")
		}
		return apperrors.NewInternalError("failed to lookup user", err)
	}

	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL
	`

	result, err := q.ExecContext(ctx, query, time.Now(), userInternalID)
	if err != nil {
		return apperrors.NewInternalError("failed to revoke user tokens", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewInternalError("failed to check revocation result", err)
	}

	r.logger.Debug(ctx, "All refresh tokens revoked", ports.Int64("count", rowsAffected))
	return nil
}

// CleanupExpired removes expired tokens (for background cleanup job)
func (r *RefreshTokenRepository) CleanupExpired(ctx context.Context) (int64, error) {
	r.logger.Debug(ctx, "Cleaning up expired refresh tokens")

	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1 OR revoked_at < $2
	`

	// Delete tokens expired more than 7 days ago or revoked more than 7 days ago
	cutoff := time.Now().Add(-7 * 24 * time.Hour)

	result, err := r.db.ExecContext(ctx, query, time.Now(), cutoff)
	if err != nil {
		return 0, apperrors.NewInternalError("failed to cleanup expired tokens", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, apperrors.NewInternalError("failed to check cleanup result", err)
	}

	r.logger.Debug(ctx, "Expired refresh tokens cleaned up", ports.Int64("count", rowsAffected))
	return rowsAffected, nil
}

// Ensure RefreshTokenRepository implements authdomain.RefreshTokenRepository
var _ authdomain.RefreshTokenRepository = (*RefreshTokenRepository)(nil)
