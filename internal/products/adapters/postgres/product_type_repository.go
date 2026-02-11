package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/elzestia/go-boilerplate/internal/products/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/database"
	apperrors "github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/jmoiron/sqlx"
)

// ProductTypeRepository implements domain.ProductTypeRepository using PostgreSQL
type ProductTypeRepository struct {
	db     *sqlx.DB
	logger ports.Logger
}

// NewProductTypeRepository creates a new PostgreSQL product type repository
func NewProductTypeRepository(db *sqlx.DB, logger ports.Logger) *ProductTypeRepository {
	return &ProductTypeRepository{
		db:     db,
		logger: logger.With(ports.String("repository", "product_types")),
	}
}

func (r *ProductTypeRepository) Create(ctx context.Context, pt *domain.ProductType) error {
	r.logger.Debug(ctx, "Creating product type", ports.String("name", pt.Name()))

	query := `
		INSERT INTO product_types (uid, name, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	q := database.GetQuerier(ctx, r.db)
	var internalID int64
	err := q.QueryRowContext(ctx, query,
		pt.UID(),
		pt.Name(),
		pt.CreatedAt(),
		pt.CreatedBy(),
		pt.UpdatedAt(),
		pt.UpdatedBy(),
	).Scan(&internalID)

	if err != nil {
		r.logger.Error(ctx, "Failed to insert product type", ports.Error(err))
		return apperrors.NewInternalError("failed to create product type", err)
	}

	pt.SetInternalID(internalID)
	return nil
}

func (r *ProductTypeRepository) FindByID(ctx context.Context, id int64) (*domain.ProductType, error) {
	var row productTypeRow

	query := `
		SELECT id, uid, name, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM product_types
		WHERE id = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductTypeNotFound
		}
		return nil, apperrors.NewInternalError("failed to find product type", err)
	}

	return productTypeToDomain(&row), nil
}

func (r *ProductTypeRepository) FindByUID(ctx context.Context, uid string) (*domain.ProductType, error) {
	var row productTypeRow

	query := `
		SELECT id, uid, name, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM product_types
		WHERE uid = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductTypeNotFound
		}
		return nil, apperrors.NewInternalError("failed to find product type", err)
	}

	return productTypeToDomain(&row), nil
}

func (r *ProductTypeRepository) FindByName(ctx context.Context, name string) (*domain.ProductType, error) {
	var row productTypeRow

	query := `
		SELECT id, uid, name, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM product_types
		WHERE name = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductTypeNotFound
		}
		return nil, apperrors.NewInternalError("failed to find product type", err)
	}

	return productTypeToDomain(&row), nil
}

func (r *ProductTypeRepository) FindAll(ctx context.Context) ([]*domain.ProductType, error) {
	var rows []productTypeRow

	query := `
		SELECT id, uid, name, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM product_types
		WHERE deleted_at IS NULL
		ORDER BY id ASC
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.SelectContext(ctx, &rows, query)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to find product types", err)
	}

	result := make([]*domain.ProductType, len(rows))
	for i, row := range rows {
		result[i] = productTypeToDomain(&row)
	}

	return result, nil
}

func (r *ProductTypeRepository) FindAllPaginated(ctx context.Context, offset, limit int) ([]*domain.ProductType, int64, error) {
	var rows []productTypeRow

	query := `
		SELECT id, uid, name, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by
		FROM product_types
		WHERE deleted_at IS NULL
		ORDER BY id ASC
		LIMIT $1 OFFSET $2
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.SelectContext(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, 0, apperrors.NewInternalError("failed to find product types", err)
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM product_types WHERE deleted_at IS NULL`
	err = q.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, apperrors.NewInternalError("failed to count product types", err)
	}

	result := make([]*domain.ProductType, len(rows))
	for i, row := range rows {
		result[i] = productTypeToDomain(&row)
	}

	return result, total, nil
}

func (r *ProductTypeRepository) Update(ctx context.Context, pt *domain.ProductType) error {
	query := `
		UPDATE product_types
		SET name = $2, updated_at = $3, updated_by = $4, deleted_at = $5, deleted_by = $6
		WHERE id = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query,
		pt.ID(),
		pt.Name(),
		pt.UpdatedAt(),
		pt.UpdatedBy(),
		pt.DeletedAt(),
		pt.DeletedBy(),
	)

	if err != nil {
		return apperrors.NewInternalError("failed to update product type", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewInternalError("failed to get rows affected", err)
	}

	if rows == 0 {
		return domain.ErrProductTypeNotFound
	}

	return nil
}

func (r *ProductTypeRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM product_types WHERE id = $1`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.NewInternalError("failed to delete product type", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewInternalError("failed to get rows affected", err)
	}

	if rows == 0 {
		return domain.ErrProductTypeNotFound
	}

	return nil
}

// Ensure ProductTypeRepository implements domain.ProductTypeRepository
var _ domain.ProductTypeRepository = (*ProductTypeRepository)(nil)
