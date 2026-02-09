package postgres

import (
	"fmt"

	"github.com/elzestia/go-boilerplate/internal/shared/types"
	"github.com/elzestia/go-boilerplate/internal/users/domain"
)

// userRow represents a user row from the database with encrypted fields
type userRow struct {
	ID              int64            `db:"id"`
	UID             string           `db:"uid"`
	EmailEncrypted  string           `db:"email_encrypted"`
	EmailLookupHash string           `db:"email_lookup_hash"`
	NameEncrypted   string           `db:"name_encrypted"`
	Username        string           `db:"username"`
	AvatarURL       types.NullString `db:"avatar_url"`
	CreatedAt       types.NullTime   `db:"created_at"`
	CreatedBy       int64            `db:"created_by"`
	UpdatedAt       types.NullTime   `db:"updated_at"`
	UpdatedBy       int64            `db:"updated_by"`
	DeletedAt       types.NullTime   `db:"deleted_at"`
	DeletedBy       types.NullString `db:"deleted_by"`
}

// toDomain converts a database row to a domain User entity
// Decrypts email and name fields transparently
func (r *UserRepository) toDomain(row *userRow) (*domain.User, error) {
	// Decrypt email
	emailPlaintext, err := r.cryptoService.Decrypt(row.EmailEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt email: %w", err)
	}

	email, err := domain.NewEmail(emailPlaintext)
	if err != nil {
		return nil, fmt.Errorf("invalid email in database: %w", err)
	}

	// Decrypt name
	namePlaintext, err := r.cryptoService.Decrypt(row.NameEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt name: %w", err)
	}

	// Parse username
	username, err := domain.NewUsername(row.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid username in database: %w", err)
	}

	return domain.ReconstructUser(
		row.ID,
		row.UID,
		email,
		namePlaintext,
		username,
		row.AvatarURL,
		row.CreatedAt.Time,
		row.CreatedBy,
		row.UpdatedAt.Time,
		row.UpdatedBy,
		row.DeletedAt,
		row.DeletedBy,
	), nil
}
