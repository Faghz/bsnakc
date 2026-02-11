package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/elzestia/go-boilerplate/internal/products/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/database"
	apperrors "github.com/elzestia/go-boilerplate/internal/shared/errors"
	"github.com/jmoiron/sqlx"
)

// ProductRepository implements domain.ProductRepository using PostgreSQL
type ProductRepository struct {
	db     *sqlx.DB
	logger ports.Logger
}

// NewProductRepository creates a new PostgreSQL product repository
func NewProductRepository(db *sqlx.DB, logger ports.Logger) *ProductRepository {
	return &ProductRepository{
		db:     db,
		logger: logger.With(ports.String("repository", "products")),
	}
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	r.logger.Debug(ctx, "Creating product", ports.String("name", p.Name()))

	query := `
		INSERT INTO products (uid, name, product_type_id, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	q := database.GetQuerier(ctx, r.db)
	var internalID int64
	err := q.QueryRowContext(ctx, query,
		p.UID(),
		p.Name(),
		p.ProductTypeID(),
		p.CreatedAt(),
		p.CreatedBy(),
		p.UpdatedAt(),
		p.UpdatedBy(),
	).Scan(&internalID)

	if err != nil {
		r.logger.Error(ctx, "Failed to insert product", ports.Error(err))
		return apperrors.NewInternalError("failed to create product", err)
	}

	p.SetInternalID(internalID)
	return nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	var row productRow

	query := `
		SELECT p.id, p.uid, p.name, p.product_type_id, pt.uid as product_type_uid,
		       p.created_at, p.created_by, p.updated_at, p.updated_by, p.deleted_at, p.deleted_by
		FROM products p
		JOIN product_types pt ON p.product_type_id = pt.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, apperrors.NewInternalError("failed to find product", err)
	}

	return productToDomain(&row), nil
}

func (r *ProductRepository) FindByUID(ctx context.Context, uid string) (*domain.Product, error) {
	var row productRow

	query := `
		SELECT p.id, p.uid, p.name, p.product_type_id, pt.uid as product_type_uid,
		       p.created_at, p.created_by, p.updated_at, p.updated_by, p.deleted_at, p.deleted_by
		FROM products p
		JOIN product_types pt ON p.product_type_id = pt.id
		WHERE p.uid = $1 AND p.deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.GetContext(ctx, &row, query, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, apperrors.NewInternalError("failed to find product", err)
	}

	return productToDomain(&row), nil
}

func (r *ProductRepository) FindAll(ctx context.Context) ([]*domain.Product, error) {
	var rows []productRow

	query := `
		SELECT p.id, p.uid, p.name, p.product_type_id, pt.uid as product_type_uid,
		       p.created_at, p.created_by, p.updated_at, p.updated_by, p.deleted_at, p.deleted_by
		FROM products p
		JOIN product_types pt ON p.product_type_id = pt.id
		WHERE p.deleted_at IS NULL
		ORDER BY p.id ASC
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.SelectContext(ctx, &rows, query)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to find products", err)
	}

	result := make([]*domain.Product, len(rows))
	for i, row := range rows {
		result[i] = productToDomain(&row)
	}

	return result, nil
}

func (r *ProductRepository) FindAllPaginated(ctx context.Context, offset, limit int) ([]*domain.Product, int64, error) {
	var rows []productRow

	query := `
		SELECT p.id, p.uid, p.name, p.product_type_id, pt.uid as product_type_uid,
		       p.created_at, p.created_by, p.updated_at, p.updated_by, p.deleted_at, p.deleted_by
		FROM products p
		JOIN product_types pt ON p.product_type_id = pt.id
		WHERE p.deleted_at IS NULL
		ORDER BY p.id ASC
		LIMIT $1 OFFSET $2
	`

	q := database.GetQuerier(ctx, r.db)
	err := q.SelectContext(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, 0, apperrors.NewInternalError("failed to find products", err)
	}

	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM products p
		JOIN product_types pt ON p.product_type_id = pt.id
		WHERE p.deleted_at IS NULL
	`
	err = q.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, apperrors.NewInternalError("failed to count products", err)
	}

	result := make([]*domain.Product, len(rows))
	for i, row := range rows {
		result[i] = productToDomain(&row)
	}

	return result, total, nil
}

func (r *ProductRepository) FindPaginatedList(ctx context.Context, criteria domain.ProductSearchCriteria) ([]*domain.Product, error) {
	var rows []productRow

	// Build the query with optional filters
	query := `
		SELECT p.id, p.uid, p.name, p.product_type_id, pt.uid as product_type_uid,
		       p.created_at, p.created_by, p.updated_at, p.updated_by, p.deleted_at, p.deleted_by
		FROM products p
		JOIN product_types pt ON p.product_type_id = pt.id
		WHERE p.deleted_at IS NULL
	`
	args := []interface{}{}
	argCount := 0

	// Add search filter if provided
	if criteria.Search != "" {
		argCount++
		query += ` AND p.name ILIKE $` + strconv.Itoa(argCount)
		args = append(args, "%"+criteria.Search+"%")
	}

	// Add product type filter if provided
	if criteria.ProductTypeUID != "" {
		argCount++
		query += ` AND pt.uid = $` + strconv.Itoa(argCount)
		args = append(args, criteria.ProductTypeUID)
	}

	query += ` ORDER BY p.id ASC LIMIT $` + strconv.Itoa(argCount+1) + ` OFFSET $` + strconv.Itoa(argCount+2)
	args = append(args, criteria.Limit, criteria.Offset)

	q := database.GetQuerier(ctx, r.db)
	err := q.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to find products", err)
	}

	result := make([]*domain.Product, len(rows))
	for i, row := range rows {
		result[i] = productToDomain(&row)
	}

	return result, nil
}

func (r *ProductRepository) FindPaginatedCount(ctx context.Context, criteria domain.ProductSearchCriteria) (int64, error) {
	// Build count query with same filters
	countQuery := `
		SELECT COUNT(*)
		FROM products p
		JOIN product_types pt ON p.product_type_id = pt.id
		WHERE p.deleted_at IS NULL
	`
	countArgs := []interface{}{}

	if criteria.Search != "" {
		countQuery += ` AND p.name ILIKE $1`
		countArgs = append(countArgs, "%"+criteria.Search+"%")
	}

	if criteria.ProductTypeUID != "" {
		placeholder := "$1"
		if criteria.Search != "" {
			placeholder = "$2"
		}
		countQuery += ` AND pt.uid = ` + placeholder
		countArgs = append(countArgs, criteria.ProductTypeUID)
	}

	q := database.GetQuerier(ctx, r.db)
	var total int64
	err := q.GetContext(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return 0, apperrors.NewInternalError("failed to count products", err)
	}

	return total, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) error {
	query := `
		UPDATE products
		SET name = $2, product_type_id = $3, updated_at = $4, updated_by = $5, deleted_at = $6, deleted_by = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query,
		p.ID(),
		p.Name(),
		p.ProductTypeID(),
		p.UpdatedAt(),
		p.UpdatedBy(),
		p.DeletedAt(),
		p.DeletedBy(),
	)

	if err != nil {
		return apperrors.NewInternalError("failed to update product", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewInternalError("failed to get rows affected", err)
	}

	if rows == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.NewInternalError("failed to delete product", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewInternalError("failed to get rows affected", err)
	}

	if rows == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}

// Ensure ProductRepository implements domain.ProductRepository
var _ domain.ProductRepository = (*ProductRepository)(nil)
