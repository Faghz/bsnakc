package domain

import (
	"context"
)

// ProductSearchCriteria contains criteria for product search and filtering at repository level
type ProductSearchCriteria struct {
	Offset         int
	Limit          int
	Search         string
	ProductTypeUID string
}

// ProductTypeRepository defines the interface for product type data operations
type ProductTypeRepository interface {
	Create(ctx context.Context, pt *ProductType) error
	FindByID(ctx context.Context, id int64) (*ProductType, error)
	FindByUID(ctx context.Context, uid string) (*ProductType, error)
	FindByName(ctx context.Context, name string) (*ProductType, error)
	FindAll(ctx context.Context) ([]*ProductType, error)
	FindAllPaginated(ctx context.Context, offset, limit int) ([]*ProductType, int64, error)
	Update(ctx context.Context, pt *ProductType) error
	Delete(ctx context.Context, id int64) error
}

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	FindByID(ctx context.Context, id int64) (*Product, error)
	FindByUID(ctx context.Context, uid string) (*Product, error)
	FindAll(ctx context.Context) ([]*Product, error)
	FindAllPaginated(ctx context.Context, offset, limit int) ([]*Product, int64, error)
	FindPaginatedList(ctx context.Context, criteria ProductSearchCriteria) ([]*Product, error)
	FindPaginatedCount(ctx context.Context, criteria ProductSearchCriteria) (int64, error)
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id int64) error
}
