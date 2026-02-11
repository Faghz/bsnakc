package postgres

import (
	"github.com/elzestia/go-boilerplate/internal/products/domain"
	"github.com/elzestia/go-boilerplate/internal/shared/types"
)

type productRow struct {
	ID             int64            `db:"id"`
	UID            string           `db:"uid"`
	Name           string           `db:"name"`
	ProductTypeID  int64            `db:"product_type_id"`
	ProductTypeUID string           `db:"product_type_uid"`
	CreatedAt      types.NullTime   `db:"created_at"`
	CreatedBy      string           `db:"created_by"`
	UpdatedAt      types.NullTime   `db:"updated_at"`
	UpdatedBy      string           `db:"updated_by"`
	DeletedAt      types.NullTime   `db:"deleted_at"`
	DeletedBy      types.NullString `db:"deleted_by"`
}

func productToDomain(row *productRow) *domain.Product {
	return domain.ReconstructProduct(
		row.ID,
		row.UID,
		row.Name,
		row.ProductTypeID,
		row.ProductTypeUID,
		row.CreatedAt.Time,
		row.CreatedBy,
		row.UpdatedAt.Time,
		row.UpdatedBy,
		row.DeletedAt,
		row.DeletedBy,
	)
}

type productTypeRow struct {
	ID        int64            `db:"id"`
	UID       string           `db:"uid"`
	Name      string           `db:"name"`
	CreatedAt types.NullTime   `db:"created_at"`
	CreatedBy string           `db:"created_by"`
	UpdatedAt types.NullTime   `db:"updated_at"`
	UpdatedBy string           `db:"updated_by"`
	DeletedAt types.NullTime   `db:"deleted_at"`
	DeletedBy types.NullString `db:"deleted_by"`
}

func productTypeToDomain(row *productTypeRow) *domain.ProductType {
	return domain.ReconstructProductType(
		row.ID,
		row.UID,
		row.Name,
		row.CreatedAt.Time,
		row.CreatedBy,
		row.UpdatedAt.Time,
		row.UpdatedBy,
		row.DeletedAt,
		row.DeletedBy,
	)
}
