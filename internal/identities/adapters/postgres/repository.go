package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/elzestia/go-boilerplate/internal/identities/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/crypto"
	"github.com/elzestia/go-boilerplate/internal/shared/database"
	errorsx "github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/elzestia/go-boilerplate/internal/shared/types"
	"github.com/jmoiron/sqlx"
)

// IdentityRepository implements the domain.Repository interface using PostgreSQL
// with transparent encryption/decryption of provider IDs
type IdentityRepository struct {
	db            *sqlx.DB
	logger        ports.Logger
	cryptoService *crypto.CryptoService
}

// NewIdentityRepository creates a new PostgreSQL identity repository with crypto support
func NewIdentityRepository(db *sqlx.DB, logger ports.Logger, cryptoService *crypto.CryptoService) *IdentityRepository {
	return &IdentityRepository{
		db:            db,
		logger:        logger.With(ports.String("repository", "identities")),
		cryptoService: cryptoService,
	}
}

// identityRow represents an identity row from the database with encrypted fields
type identityRow struct {
	ID                   int64            `db:"id"`
	UID                  string           `db:"uid"`
	UserID               int64            `db:"user_id"`
	UserUID              string           `db:"user_uid"`
	Provider             string           `db:"provider"`
	ProviderIDEncrypted  types.NullString `db:"provider_id_encrypted"`
	ProviderIDLookupHash types.NullString `db:"provider_id_lookup_hash"`
	PasswordHash         string           `db:"password_hash"`
	CreatedAt            types.NullTime   `db:"created_at"`
	CreatedBy            int64            `db:"created_by"`
	UpdatedAt            types.NullTime   `db:"updated_at"`
	UpdatedBy            int64            `db:"updated_by"`
	DeletedAt            types.NullTime   `db:"deleted_at"`
	DeletedBy            types.NullString `db:"deleted_by"`
}

// toDomain converts a database row to a domain Identity entity
// Decrypts provider_id transparently
func (r *IdentityRepository) toDomain(row *identityRow) (*domain.Identity, error) {
	// Decrypt provider ID (handle null case)
	var providerID types.NullString
	if row.ProviderIDEncrypted.Valid {
		providerIDPlaintext, err := r.cryptoService.Decrypt(row.ProviderIDEncrypted.String)
		if err != nil {
			return nil, errorsx.NewInternalError("failed to decrypt provider_id", err)
		}
		providerID = types.NullStringFrom(providerIDPlaintext)
	}
	// If not valid, providerID remains as zero value (null)

	return domain.ReconstructIdentity(
		row.ID,
		row.UID,
		row.UserID,
		row.UserUID,
		domain.Provider(row.Provider),
		providerID,
		domain.NewHashedPassword(row.PasswordHash),
		row.CreatedAt.Time,
		row.CreatedBy,
		row.UpdatedAt.Time,
		row.UpdatedBy,
		row.DeletedAt,
		row.DeletedBy,
	), nil
}

// Create creates a new identity in the database
// Encrypts provider_id and computes HMAC for SSO login lookup
func (r *IdentityRepository) Create(ctx context.Context, i *domain.Identity) error {
	r.logger.Debug(ctx, "Creating identity", ports.String("identity_id", i.ID()), ports.String("provider", string(i.Provider())))

	// Encrypt provider ID (handle null case)
	var providerIDEncrypted, providerIDLookupHash types.NullString

	if i.ProviderID().Valid {
		encrypted, err := r.cryptoService.Encrypt(i.ProviderID().String)
		if err != nil {
			r.logger.Error(ctx, "Failed to encrypt provider_id", ports.Error(err))
			return errorsx.NewInternalError("failed encrypting provider_id", err)
		}
		providerIDEncrypted = types.NullStringFrom(encrypted)

		// Compute HMAC for provider ID lookup
		hash, err := r.cryptoService.HMAC(i.ProviderID().String)
		if err != nil {
			r.logger.Error(ctx, "Failed to compute provider_id HMAC", ports.Error(err))
			return errorsx.NewInternalError("failed computing provider_id HMAC", err)
		}
		providerIDLookupHash = types.NullStringFrom(hash)
	}
	// If not valid, both remain as null

	query := `
		INSERT INTO identities (
			uid, user_id, provider, provider_id_encrypted, provider_id_lookup_hash,
			password_hash, created_at, updated_at, created_by, updated_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, 0)
		RETURNING id
	`

	q := database.GetQuerier(ctx, r.db)
	var internalID int64
	err := q.QueryRowContext(ctx, query,
		i.UID(),
		i.InternalUserID(),
		string(i.Provider()),
		providerIDEncrypted,
		providerIDLookupHash,
		i.HashedPassword().Hash(),
		i.CreatedAt(),
		i.UpdatedAt(),
	).Scan(&internalID)

	if err != nil {
		r.logger.Error(ctx, "Failed to insert identity", ports.Error(err), ports.String("identity_id", i.ID()))
		return errorsx.NewInternalError("failed to insert identity", err)
	}

	// Set the internal ID returned by the database
	i.SetInternalID(internalID)

	r.logger.Debug(ctx, "Identity created successfully", ports.String("identity_id", i.ID()), ports.String("provider", string(i.Provider())))
	return nil
}

// FindByID finds an identity by ID
func (r *IdentityRepository) FindByID(ctx context.Context, id string) (*domain.Identity, error) {
	var row identityRow

	query := `
		SELECT i.id, i.uid, i.user_id, u.uid as user_uid, i.provider,
		       i.provider_id_encrypted, i.provider_id_lookup_hash,
		       i.password_hash, i.created_at, i.created_by,
		       i.updated_at, i.updated_by, i.deleted_at, i.deleted_by
		FROM identities i
		JOIN users u ON i.user_id = u.id
		WHERE i.uid = $1 AND i.deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIdentityNotFound
		}
		return nil, errors.ErrUnsupported
	}

	return r.toDomain(&row)
}

// FindByUserID finds all identities for a user
func (r *IdentityRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Identity, error) {
	var rows []identityRow

	query := `
		SELECT i.id, i.uid, i.user_id, u.uid as user_uid, i.provider,
		       i.provider_id_encrypted, i.provider_id_lookup_hash,
		       i.password_hash, i.created_at, i.created_by,
		       i.updated_at, i.updated_by, i.deleted_at, i.deleted_by
		FROM identities i
		JOIN users u ON i.user_id = u.id
		WHERE u.uid = $1 AND i.deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.SelectContext(ctx, &rows, query, userID)
	if err != nil {
		return nil, errorsx.NewInternalError("failed to find identities", err)
	}

	identities := make([]*domain.Identity, len(rows))
	for i, row := range rows {
		identity, err := r.toDomain(&row)
		if err != nil {
			return nil, err
		}
		identities[i] = identity
	}

	return identities, nil
}

// Update updates an existing identity
func (r *IdentityRepository) Update(ctx context.Context, i *domain.Identity) error {
	// Encrypt provider ID (handle null case)
	var providerIDEncrypted, providerIDLookupHash types.NullString

	if i.ProviderID().Valid {
		encrypted, err := r.cryptoService.Encrypt(i.ProviderID().String)
		if err != nil {
			return errorsx.NewInternalError("failed encrypting provider_id", err)
		}
		providerIDEncrypted = types.NullStringFrom(encrypted)

		// Compute HMAC for provider ID lookup
		hash, err := r.cryptoService.HMAC(i.ProviderID().String)
		if err != nil {
			return errorsx.NewInternalError("failed computing provider_id HMAC", err)
		}
		providerIDLookupHash = types.NullStringFrom(hash)
	}
	// If not valid, both remain as null

	query := `
		UPDATE identities
		SET user_id = $2, provider = $3, provider_id_encrypted = $4,
		    provider_id_lookup_hash = $5, password_hash = $6,
		    updated_at = $7, updated_by = $8
		WHERE uid = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query,
		i.UID(),
		i.InternalUserID(),
		string(i.Provider()),
		providerIDEncrypted,
		providerIDLookupHash,
		i.HashedPassword().Hash(),
		i.UpdatedAt(),
		i.UpdatedBy(),
	)

	if err != nil {
		return errorsx.NewInternalError("failed to update identity", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errorsx.NewInternalError("failed to get rows affected`", err)
	}

	if rows == 0 {
		return domain.ErrIdentityNotFound
	}

	return nil
}

// Delete deletes an identity by ID
func (r *IdentityRepository) Delete(ctx context.Context, id string, deletedBy string) error {
	query := `UPDATE identities SET deleted_at = NOW(), deleted_by = $2 WHERE id = $1`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query, id, deletedBy)
	if err != nil {
		return errorsx.NewInternalError("failed to delete identity", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errorsx.NewInternalError("failed to get rows affected", err)
	}

	if rows == 0 {
		return domain.ErrIdentityNotFound
	}

	return nil
}

func (r *IdentityRepository) FindByProviderAndUserID(ctx context.Context, provider domain.Provider, userID int64) (*domain.Identity, error) {
	var row identityRow

	query := `
		SELECT i.id, i.uid, i.user_id,  i.provider,
		       i.provider_id_encrypted, i.provider_id_lookup_hash,
		       i.password_hash, i.created_at, i.created_by,
		       i.updated_at, i.updated_by, i.deleted_at, i.deleted_by
		FROM identities i
		WHERE i.provider = $1 AND i.user_id = $2 AND i.deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, string(provider), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIdentityNotFound
		}

		return nil, errorsx.NewInternalError("failed to find identities", err)
	}

	identity, err := r.toDomain(&row)
	if err != nil {
		return nil, err
	}

	return identity, nil
}

func (r *IdentityRepository) FindByProviderAndProviderID(ctx context.Context, provider domain.Provider, providerID string) (*domain.Identity, error) {
	// Compute HMAC for provider ID lookup
	hash, err := r.cryptoService.HMAC(providerID)
	if err != nil {
		return nil, errorsx.NewInternalError("failed computing provider_id HMAC", err)
	}

	var row identityRow

	query := `
		SELECT i.id, i.uid, i.user_id, u.uid as user_uid, i.provider,
		       i.provider_id_encrypted, i.provider_id_lookup_hash,
		       i.password_hash, i.created_at, i.created_by,
		       i.updated_at, i.updated_by, i.deleted_at, i.deleted_by
		FROM identities i
		WHERE i.provider = $1 AND i.provider_id_lookup_hash = $2 AND i.deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err = q.GetContext(ctx, &row, query, string(provider), hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIdentityNotFound
		}

		return nil, errorsx.NewInternalError("failed to get user identity", err)
	}

	identity, err := r.toDomain(&row)
	if err != nil {
		return nil, err
	}

	return identity, nil
}

// Ensure IdentityRepository implements domain.Repository
var _ domain.Repository = (*IdentityRepository)(nil)
