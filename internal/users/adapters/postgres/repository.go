package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/crypto"
	"github.com/elzestia/go-boilerplate/internal/shared/database"
	apperrors "github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/elzestia/go-boilerplate/internal/users/domain"
	"github.com/jmoiron/sqlx"
)

// UserRepository implements the domain.Repository interface using PostgreSQL
// with transparent encryption/decryption of PII fields (email, name)
type UserRepository struct {
	db            *sqlx.DB
	logger        ports.Logger
	cryptoService *crypto.CryptoService
}

// NewUserRepository creates a new PostgreSQL user repository with crypto support
func NewUserRepository(db *sqlx.DB, logger ports.Logger, cryptoService *crypto.CryptoService) *UserRepository {
	return &UserRepository{
		db:            db,
		logger:        logger.With(ports.String("repository", "users")),
		cryptoService: cryptoService,
	}
}

// Create creates a new user in the database
// Encrypts email and name, computes HMAC for email lookup
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	r.logger.Debug(ctx, "Creating user", ports.Int64("user_id", u.ID()))

	// Encrypt email and name
	emailEncrypted, err := r.cryptoService.Encrypt(u.Email().String())
	if err != nil {
		r.logger.Error(ctx, "Failed to encrypt email", ports.Error(err))
		return apperrors.NewInternalError("failed to encrypt email", err)
	}

	nameEncrypted, err := r.cryptoService.Encrypt(u.Name())
	if err != nil {
		r.logger.Error(ctx, "Failed to encrypt name", ports.Error(err))
		return apperrors.NewInternalError("failed to encrypt name", err)
	}

	// Compute HMAC for email lookup (deterministic, case-insensitive)
	emailLookupHash, err := r.cryptoService.HMAC(u.Email().String())
	if err != nil {
		r.logger.Error(ctx, "Failed to compute email HMAC", ports.Error(err))
		return apperrors.NewInternalError("failed to compute email HMAC", err)
	}

	query := `
		INSERT INTO users (
			uid, email_encrypted, email_lookup_hash, name_encrypted, username,
			avatar_url, created_at, updated_at, created_by, updated_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '', '')
		RETURNING id
	`

	q := database.GetQuerier(ctx, r.db)
	var internalID int64
	err = q.QueryRowContext(ctx, query,
		u.UID(),
		emailEncrypted,
		emailLookupHash,
		nameEncrypted,
		u.Username().String(),
		u.AvatarURL(),
		u.CreatedAt(),
		u.UpdatedAt(),
	).Scan(&internalID)

	if err != nil {
		r.logger.Error(ctx, "Failed to insert user", ports.Error(err), ports.Int64("user_id", u.ID()))
		return apperrors.NewInternalError("failed to create user", err)
	}

	// Update created_by and updated_by to the user's own ID
	updateQuery := `
		UPDATE users
		SET created_by = $1, updated_by = $1
		WHERE id = $1
	`
	_, err = q.ExecContext(ctx, updateQuery, internalID)
	if err != nil {
		r.logger.Error(ctx, "Failed to update audit fields", ports.Error(err), ports.Int64("user_id", u.ID()))
		return apperrors.NewInternalError("failed to update audit fields", err)
	}

	// Set the internal ID returned by the database
	u.SetInternalID(internalID)

	r.logger.Debug(ctx, "User created successfully", ports.Int64("user_id", u.ID()))
	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var row userRow

	query := `
		SELECT id, uid, email_encrypted, email_lookup_hash, name_encrypted, username,
		       avatar_url, created_at, created_by,
		       updated_at, updated_by, deleted_at, deleted_by
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, apperrors.NewInternalError("failed to find user", err)
	}

	return r.toDomain(&row)
}

// FindByUID finds a user by UID
func (r *UserRepository) FindByUID(ctx context.Context, uid string) (*domain.User, error) {
	var row userRow

	query := `
		SELECT id, uid, email_encrypted, email_lookup_hash, name_encrypted, username,
		       avatar_url, created_at, created_by,
		       updated_at, updated_by, deleted_at, deleted_by
		FROM users
		WHERE uid = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, apperrors.NewInternalError("failed to find user", err)
	}

	return r.toDomain(&row)
}

// FindByEmail finds a user by email using HMAC lookup (case-insensitive)
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	// Compute HMAC for lookup (normalization happens inside HMAC function)
	emailLookupHash, err := r.cryptoService.HMAC(email)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to compute email HMAC", err)
	}

	var row userRow

	query := `
		SELECT id, uid, email_encrypted, email_lookup_hash, name_encrypted, username,
		       avatar_url, created_at, created_by,
		       updated_at, updated_by, deleted_at, deleted_by
		FROM users
		WHERE email_lookup_hash = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err = q.GetContext(ctx, &row, query, emailLookupHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, apperrors.NewInternalError("failed to find user", err)
	}

	return r.toDomain(&row)
}

// FindByUsername finds a user by username (case-insensitive, plaintext lookup)
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var row userRow

	query := `
		SELECT id, uid, email_encrypted, email_lookup_hash, name_encrypted, username,
		       avatar_url, created_at, created_by,
		       updated_at, updated_by, deleted_at, deleted_by
		FROM users
		WHERE LOWER(username) = LOWER($1) AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, apperrors.NewInternalError("failed to find user", err)
	}

	return r.toDomain(&row)
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	// Encrypt with current key
	emailEncrypted, err := r.cryptoService.Encrypt(u.Email().String())
	if err != nil {
		return apperrors.NewInternalError("failed to encrypt email", err)
	}

	nameEncrypted, err := r.cryptoService.Encrypt(u.Name())
	if err != nil {
		return apperrors.NewInternalError("failed to encrypt name", err)
	}

	// Compute HMAC for email lookup
	emailLookupHash, err := r.cryptoService.HMAC(u.Email().String())
	if err != nil {
		return apperrors.NewInternalError("failed to compute email HMAC", err)
	}

	query := `
		UPDATE users
		SET email_encrypted = $2, email_lookup_hash = $3, name_encrypted = $4,
		    username = $5, avatar_url = $6,
		    updated_at = $7, updated_by = $8
		WHERE uid = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query,
		u.UID(),
		emailEncrypted,
		emailLookupHash,
		nameEncrypted,
		u.Username().String(),
		u.AvatarURL(),
		u.UpdatedAt(),
		u.UpdatedBy(),
	)

	if err != nil {
		return apperrors.NewInternalError("failed to update user", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewInternalError("failed to get rows affected", err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete performs a hard delete of a user by ID
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.NewInternalError("failed to delete user", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewInternalError("failed to get rows affected", err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Ensure UserRepository implements domain.Repository
var _ domain.Repository = (*UserRepository)(nil)
